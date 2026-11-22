param(
    [ValidatePattern('^[a-z0-9][a-z0-9-]{0,45}$')]
    [string]$RunId = ('docker-' + [DateTimeOffset]::UtcNow.ToOffset([TimeSpan]::FromHours(-5)).ToString('yyyyMMdd-HHmmss')),
    [switch]$SoloPruebas
)
$ErrorActionPreference = 'Stop'
$root = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$persist = Join-Path $root "results\persistencia\$RunId"
$post = Join-Path $root "results\api\postman\$RunId"
if ($SoloPruebas) {
    if (-not (Test-Path $persist -PathType Container)) { throw "No existe la ejecucion $RunId" }
    if (Get-ChildItem -LiteralPath $persist -Filter 'pruebas-*.log') { throw "Ya existen logs de pruebas para $RunId" }
} else {
    if ((Test-Path $persist) -or (Test-Path $post)) { throw "Ya existen evidencias para $RunId" }
    $null = New-Item -ItemType Directory -Path $persist,$post -Force
}
$utf8 = [Text.UTF8Encoding]::new($false)
[Console]::OutputEncoding = $utf8
$prefix = "pc4-$RunId"
$mongo = "$prefix-mongo"
$redis = "$prefix-redis"
$api = "$prefix-api"
$net = "$prefix-net"
$bin = "$prefix-bin"
$cache = "$prefix-go"
$db = $prefix.Replace('-', '_')
$go = 'golang:1.27-bookworm'
$repo = "type=bind,source=$root,target=/repo,readonly"
$out = "type=bind,source=$persist,target=/evidence"
$build = "type=volume,source=$bin,target=/binout"
$goCache = "type=volume,source=$cache,target=/go"
$envCommon = @('-e','MONGODB_URI=mongodb://mongo:27017','-e','REDIS_ADDR=redis:6379','-e','PC4_INTEGRATION=1')
$envArgs = $envCommon + @('-e',"MONGODB_DATABASE=$db")
$runner = @('run','--rm','-t','--network',$net,'--mount',$repo,'--mount',$out,'--mount',$build,'--mount',$goCache)
$script:failed = @()
$started = @()

function Save($file, $lines) {
    [IO.File]::WriteAllLines($file,[string[]]$lines,$utf8)
}

function Run($file, $program, [string[]]$arguments, [switch]$KeepFailure) {
    $old = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        $lines = & $program @arguments 2>&1
        $code = $LASTEXITCODE
    } finally { $ErrorActionPreference = $old }
    $quoted = ($arguments | ForEach-Object { '"' + ($_ -replace '"','\"') + '"' }) -join ' '
    Save $file (@("command=$program $quoted", "exit=$code") + @($lines))
    if ($code -ne 0) {
        $script:failed += $file
        if (-not $KeepFailure) { $lines | Out-Host; throw "Fallo ($code): $file" }
    }
    $lines | ForEach-Object { "$_" }
}

function Docker-Log($label, [string[]]$arguments, [switch]$KeepFailure) {
    Run (Join-Path $persist "$label.log") docker.exe $arguments -KeepFailure:$KeepFailure
}

function Ready($container, [string[]]$probe) {
    $old = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    try {
        for ($i=0; $i -lt 60; $i++) {
            $null = & docker.exe exec $container @probe 2>&1
            if ($LASTEXITCODE -eq 0) { return }
            Start-Sleep -Seconds 1
        }
    } finally { $ErrorActionPreference = $old }
    throw "Servicio sin respuesta: $container"
}

function Test-GoModules {
    foreach ($module in 'plataforma','kmeans_secuencial') {
        $testEnv = $envCommon + @('-e',"MONGODB_DATABASE=${db}_tests")
        if ($module -eq 'kmeans_secuencial') { $testEnv += @('-e','KMEANS_REAL_DATA=1') }
        foreach ($check in 'normal','race','vet') {
            $command = switch ($check) {
                normal { @('test','-v','-count=1','./...') }
                race { @('test','-race','-v','-count=1','./...') }
                vet { @('vet','./...') }
            }
            Docker-Log "pruebas-$module-$check" ($runner + $testEnv +
                @('-w',"/repo/go/$module",$go,'go') + $command) -KeepFailure | Out-Host
        }
    }
}

if ($SoloPruebas) {
    try {
        Docker-Log pruebas-servicios-iniciar @('start',$mongo,$redis) | Out-Null
        Ready $mongo @('mongosh','--quiet','--eval','db.adminCommand({ping:1})')
        Ready $redis @('redis-cli','ping')
        Test-GoModules
    } finally { Docker-Log pruebas-servicios-detener @('stop',$mongo,$redis) -KeepFailure | Out-Null }
    if ($script:failed.Count) { throw ('Fallaron: ' + ($script:failed -join ', ')) }
    return
}

try {
    Save (Join-Path $persist 'entorno.txt') @("fecha_lima=$([DateTimeOffset]::UtcNow.ToOffset([TimeSpan]::FromHours(-5)).ToString('o'))",
        "run_id=$RunId","base=$db","ttl_demo_segundos=3","ttl_api_segundos=60",
        "windows=$([Environment]::OSVersion.Version)","commit=$(git -C $root rev-parse HEAD)",
        "node=$(& node.exe --version)")
    Docker-Log docker-version @('version','--format','{{.Client.Version}} / {{.Server.Version}}') | Out-Null
    Docker-Log docker-entorno @('info','--format','{{.OSType}} {{.Architecture}}') | Out-Null
    Docker-Log red @('network','create',$net) | Out-Null
    Docker-Log volumen-binarios @('volume','create',$bin) | Out-Null
    Docker-Log volumen-go @('volume','create',$cache) | Out-Null
    Docker-Log mongo-inicio @('run','-d','--name',$mongo,'--network',$net,'--network-alias','mongo','mongo:8') | Out-Null
    $started += $mongo
    Docker-Log redis-inicio @('run','-d','--name',$redis,'--network',$net,'--network-alias','redis','redis:7') | Out-Null
    $started += $redis
    Ready $mongo @('mongosh','--quiet','--eval','db.adminCommand({ping:1})')
    Ready $redis @('redis-cli','ping')
    Docker-Log mongo-version @('exec',$mongo,'mongosh','--quiet','--eval','db.version()') | Out-Null
    Docker-Log redis-version @('exec',$redis,'redis-server','--version') | Out-Null
    Docker-Log imagenes @('image','inspect','--format','{{index .RepoDigests 0}}',$go,'mongo:8','redis:7') | Out-Null
    Docker-Log go-version @('run','--rm',$go,'go','version') | Out-Null
    Docker-Log compilar ($runner + @($go,'sh','-ec','cd /repo/go/kmeans_secuencial; go build -o /binout/secuencial .; cd /repo/go/plataforma; go build -o /binout/persistencia ./cmd/persistencia; go build -o /binout/api ./cmd/api; go build -o /binout/node ./cmd/node')) | Out-Host

    foreach ($period in '1_hora','24_horas') {
        $csvRel = "data/vectors/features_$period.csv"
        $csv = Join-Path $root $csvRel
        $sha = (Get-FileHash -LiteralPath $csv -Algorithm SHA256).Hash.ToLowerInvariant()
        $n = 0
        foreach ($line in [IO.File]::ReadLines($csv)) { if ($line.Length) { $n++ } }
        $n--
        Save (Join-Path $persist "dataset-$period.json") (@{
            dataset_id=${sha}+':'+$period; archivo=$csvRel; fuente='CSV versionado del repositorio'
            sha256=$sha; periodo=$period; vectores=$n
        } | ConvertTo-Json)
        Docker-Log "referencia-$period" ($runner + @($go,'sh','-ec',
            "/binout/secuencial -input /repo/$csvRel -period $period -mode sequential -k 4 -seed 42 -tol 0.001 -max-iter 100 > /evidence/reporte-$period.json")) | Out-Host
        Docker-Log "persistencia-$period" ($runner + $envArgs + @('-e','CACHE_TTL_SECONDS=3',$go,'/binout/persistencia','-dataset',"/evidence/dataset-$period.json",'-reporte',"/evidence/reporte-$period.json",'-salida',"/evidence/evidencia-$period.json")) | Out-Host
        $ref = Get-Content (Join-Path $persist "reporte-$period.json") -Raw -Encoding UTF8 | ConvertFrom-Json
        $ev = Get-Content (Join-Path $persist "evidencia-$period.json") -Raw -Encoding UTF8 | ConvertFrom-Json
        $stored = $ev.documento_mongodb.resultado
        $cached = $ev.valor_redis.resultado
        $sources = $ev.pruebas.fuente -join ','
        if ($ref.vectors -ne $n -or $ev.dataset.sha256 -cne $sha -or
            $ref.tolerance -ne 0.001 -or $ref.k -ne 4 -or $ref.seed -ne 42 -or $ref.max_iterations -ne 100 -or
            ($ref.counts | Measure-Object -Sum).Sum -ne $n -or $ev.documento_mongodb.parametros.tolerancia -ne 0.001 -or
            ($stored.centroides | ConvertTo-Json -Compress -Depth 8) -cne ($ref.centroids | ConvertTo-Json -Compress -Depth 8) -or
            ($stored.conteos -join ',') -cne ($ref.counts -join ',') -or $stored.inercia -ne $ref.inertia -or
            $stored.iteraciones -ne $ref.iterations -or $stored.convergio -ne $ref.converged -or
            ($stored.centroides | ConvertTo-Json -Compress -Depth 8) -cne ($cached.centroides | ConvertTo-Json -Compress -Depth 8) -or
            $ev.ejecucion_fallida.estado -cne 'failed' -or $ev.ejecucion_fallida.resultado -or
            $sources -cne 'MongoDB,MongoDB,MongoDB,Redis,MongoDB,MongoDB' -or
            ($ev.pruebas | Where-Object { $_.resultado -notlike 'PASS*' })) { throw "Evidencia inconsistente: $period" }
        Save (Join-Path $persist "verificacion-$period.log") @("exit=0","vectores=$n","sha256=$sha","iteraciones=$($stored.iteraciones)","convergio=$($stored.convergio)","fuentes=$sources","mongodb_redis_referencia=verificados")
    }

    Docker-Log api-inicio (@('run','-dt','--name',$api,'--network',$net,'--network-alias','api','-p','127.0.0.1::8080','--mount',$repo,'--mount',"$build,readonly") +
        $envArgs + @('-e','DATA_DIR=/repo/data/vectors','-e','CACHE_TTL_SECONDS=60',$go,'/binout/api')) | Out-Null
    $started += $api
    $address = Docker-Log api-puerto @('port',$api,'8080/tcp') | Select-Object -First 1
    $baseUrl = "http://$address"
    for ($i=0; $i -lt 60; $i++) {
        try { if ((Invoke-WebRequest -UseBasicParsing "$baseUrl/health" -TimeoutSec 2).StatusCode -eq 200) { break } } catch {}
        Start-Sleep -Seconds 1
    }
    if ($i -eq 60) { throw "API sin respuesta: $baseUrl" }
    foreach ($id in 'nodo-1','nodo-2') {
        $name = "$prefix-$id"
        Docker-Log "$id-inicio" @('run','-dt','--name',$name,'--network',$net,'--mount',"$build,readonly",$go,'sh','-c',
            "while true; do /binout/node -coordinator api:9000 -id $id -interval 200ms -timeout 20s; sleep 1; done") | Out-Null
        $started += $name
    }
    Copy-Item (Join-Path $persist 'entorno.txt') (Join-Path $post 'entorno.txt')
    Add-Content (Join-Path $post 'entorno.txt') "base_url=$baseUrl"
    Run (Join-Path $post 'newman-version.log') npx.cmd @('--yes','newman@6.2.3','--version') | Out-Host
    $collection = Join-Path $root 'docs\pc4\postman\EAS_API_PC4.postman_collection.json'
    $environment = Join-Path $root 'docs\pc4\postman\EAS_local.postman_environment.json'
    function Newman($label,$folder,$envPath,$iterations,[string[]]$extra) {
        $argsN = @('--yes','newman@6.2.3','run',$collection,'-e',$envPath,'--env-var',"base_url=$baseUrl",
            '--folder',$folder,'-n',"$iterations",'-r','cli,json','--reporter-json-export',(Join-Path $post "$label.json"))
        Run (Join-Path $post "$label.log") npx.cmd ($argsN + $extra) -KeepFailure | Out-Host
    }
    $export = Join-Path $post 'entorno-final.json'
    Newman funcionales '1. Funcionales' $environment 1 @('--export-environment',$export)
    if (Test-Path $export) {
        Newman latencia '3. Latencia (Runner, 30 iteraciones)' $export 30 @()
    } else { Save (Join-Path $post 'latencia.omitida.txt') 'Newman no exporto el entorno funcional.' }
    Run (Join-Path $post 'mongo-detener.log') docker.exe @('stop',$mongo) | Out-Null
    try { Newman sin-mongodb '2. Dependencia no disponible (manual)' $environment 1 @() }
    finally {
        Run (Join-Path $post 'mongo-reiniciar.log') docker.exe @('start',$mongo) | Out-Null
        Ready $mongo @('mongosh','--quiet','--eval','db.adminCommand({ping:1})')
    }

    Test-GoModules
} finally {
    foreach ($container in $started) {
        $dir = if ($container -eq $api -or $container -match '-nodo-[12]$') { $post } else { $persist }
        Run (Join-Path $dir "$container.log") docker.exe @('logs',$container) -KeepFailure | Out-Null
    }
    for ($i=$started.Count-1; $i -ge 0; $i--) {
        Run (Join-Path $persist "$($started[$i])-stop.log") docker.exe @('stop',$started[$i]) -KeepFailure | Out-Null
    }
}
Write-Host "Persistencia: $persist"
Write-Host "Postman: $post"
if ($script:failed.Count) { throw ('Ver los pasos con salida distinta de 0: ' + ($script:failed -join ', ')) }
