param([switch]$SoloPruebas, [switch]$SoloComparaciones)
$ErrorActionPreference = 'Stop'
$root = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
$image = 'golang:1.27-bookworm'
$prefix = "eas-validacion-$PID"
$network = "$prefix-net"
$volume = "$prefix-bin"
$repo = "type=bind,source=$root,target=/repo,readonly"
$evidence = "type=bind,source=$PSScriptRoot,target=/evidence"
$bin = "type=volume,source=$volume,target=/binout"
$utf8 = [System.Text.UTF8Encoding]::new($false)

function Docker {
    & docker.exe @args
    if ($LASTEXITCODE -ne 0) { throw "docker ($LASTEXITCODE): $args" }
}

function Save($file, $lines) {
    [System.IO.File]::WriteAllLines((Join-Path $PSScriptRoot $file), [string[]]$lines, $utf8)
}

function Test-Module($module, $label, $command) {
    $output = & docker.exe run --rm -t --mount $repo -w "/repo/go/$module" $image sh -c $command
    $code = $LASTEXITCODE
    Save "$module-$label.log" (@("command=$command", "exit=$code") + @($output))
    $output | Out-Host
    if ($code -ne 0) { throw "$module $label fallo ($code)" }
}

Save 'entorno.log' @(
    "base=5424512"
    "windows=$([System.Environment]::OSVersion.Version)"
    (Docker version --format 'docker_client={{.Client.Version}} docker_server={{.Server.Version}}')
    (Docker info --format 'docker_os={{.OSType}} arch={{.Architecture}}')
    (Docker run --rm $image go version)
    (Docker image inspect --format '{{index .RepoDigests 0}}' $image)
)

if (-not $SoloComparaciones) {
    Test-Module plataforma test 'go test -count=1 ./...'
    Test-Module plataforma race 'go test -race -count=1 ./...'
    Test-Module plataforma vet 'go vet ./...'
    Test-Module plataforma fallos "go test -v -count=1 -run 'TestExecutionFaults|TestServeStopsIdleAssignment' ./internal/cluster"
    Test-Module plataforma estabilidad "go test -race -count=10 -run 'TestExecutionFaults|TestServeStopsIdleAssignment|TestDistributedTCP' ./internal/cluster"
}
if ($SoloPruebas) { return }

function Run-Case($case, $period, $inputFile, $k) {
    $common = @('-k', "$k", '-seed', '42', '-tol', '0.0001', '-max-iter', '100')
    $ref = & docker.exe run --rm -t --mount $repo --mount "$bin,readonly" $image /binout/secuencial -input $inputFile -period $period -mode sequential @common
    $refExit = $LASTEXITCODE
    Save "$case-secuencial.json" $ref
    Save "$case-secuencial-exit.log" "exit=$refExit"
    if ($refExit -ne 0) { throw "reference $case failed: $refExit" }

    foreach ($n in 1, 2) {
        $tag = "$case-${n}n"
        $coord = "$prefix-$tag"
        $containers = @()
        try {
            Docker run -dt --name $coord --network $network --mount $repo --mount $evidence --mount "$bin,readonly" $image /binout/coordinator -listen 0.0.0.0:9000 -timeout 20s -job-timeout 2m -file $inputFile -nodes $n @common -output "/evidence/$tag.json" | Out-Null
            $containers += $coord
            foreach ($i in 1..$n) {
                $node = "$coord-node-$i"
                Docker run -dt --name $node --network $network --mount "$bin,readonly" $image /binout/node -coordinator "${coord}:9000" -id "nodo-$i" -interval 200ms -timeout 20s | Out-Null
                $containers += $node
            }
            $codes = @(Docker wait @containers)
            for ($i = 0; $i -lt $containers.Count; $i++) {
                $role = if ($i -eq 0) { 'coordinator' } else { "nodo-$i" }
                Save "$tag-$role.log" (@("exit=$($codes[$i])") + @(Docker logs $containers[$i]))
            }
            if ($codes.Count -ne $containers.Count -or ($codes | Where-Object { $_ -ne '0' })) {
                throw "$tag failed: $codes"
            }
        } finally {
            if ($containers.Count) { Docker rm -f @containers | Out-Null }
        }
    }
}

Docker network create $network | Out-Null
try {
    Docker volume create $volume | Out-Null
    try {
        Docker run --rm --mount $repo --mount $bin $image sh -ec 'cd /repo/go/plataforma; go build -o /binout/coordinator ./cmd/coordinator; go build -o /binout/node ./cmd/node; cd /repo/go/kmeans_secuencial; go build -o /binout/secuencial .'
        Run-Case sintetico 1_hora /repo/results/pruebas/cluster-fallos/sintetico.csv 2
        Run-Case 1_hora 1_hora /repo/data/vectors/features_1_hora.csv 4
        Run-Case 24_horas 24_horas /repo/data/vectors/features_24_horas.csv 4
        $summary = & py -3 (Join-Path $PSScriptRoot 'comparar.py') $PSScriptRoot (Join-Path $root 'data\vectors')
        $code = $LASTEXITCODE
        Save 'comparaciones.log' (@($summary) + @("exit=$code"))
        $summary | Out-Host
        if ($code -ne 0) { throw "Equivalence failed: $code" }
    } finally {
        Docker volume rm $volume | Out-Null
    }
} finally {
    Docker network rm $network | Out-Null
}
