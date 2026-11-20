$ErrorActionPreference = 'Stop'
$root = (Resolve-Path (Join-Path $PSScriptRoot '..\..\..')).Path
$image = 'golang:1.27-bookworm'
$prefix = "eas-v2-$PID"
$network = "$prefix-net"
$volume = "$prefix-bin"
$moduleMount = "type=bind,source=$root\go\plataforma,target=/app,readonly"
$dataMount = "type=bind,source=$root\data\vectors,target=/vectors,readonly"
$evidenceMount = "type=bind,source=$PSScriptRoot,target=/evidence"
$binMount = "type=volume,source=$volume,target=/binaries"
$utf8 = [System.Text.UTF8Encoding]::new($false)

function Save-Log($name, $lines) {
    [System.IO.File]::WriteAllLines((Join-Path $PSScriptRoot $name), [string[]]$lines, $utf8)
    $lines | Write-Host
}

function Invoke-Docker {
    & docker @args
    if ($LASTEXITCODE -ne 0) {
        throw "Docker failed ($LASTEXITCODE): $args"
    }
}

function Save-ContainerLog($container, $name) {
    $lines = Invoke-Docker logs $container
    Save-Log $name $lines
}

function Invoke-Case($case, $inputPath, $localCSV, $k) {
    $coord = "$prefix-$case-coordinator"
    $one = "$prefix-$case-node1"
    $two = "$prefix-$case-node2"
    $resultName = "$case-result.json"
    $containers = @()
    try {
        Invoke-Docker run -dt --name $coord --network $network --mount "$binMount,readonly" --mount $dataMount --mount $evidenceMount $image /binaries/coordinator -listen 0.0.0.0:9000 -timeout 20s -job-timeout 2m -file $inputPath -nodes 2 -k $k -seed 42 -tol 0.0001 -max-iter 100 -output "/evidence/$resultName" | Out-Host
        $containers += $coord
        Invoke-Docker run -dt --name $one --network $network --mount "$binMount,readonly" $image /binaries/node -coordinator "${coord}:9000" -id nodo-1 -interval 200ms -timeout 20s | Out-Host
        $containers += $one
        Invoke-Docker run -dt --name $two --network $network --mount "$binMount,readonly" $image /binaries/node -coordinator "${coord}:9000" -id nodo-2 -interval 200ms -timeout 20s | Out-Host
        $containers += $two

        $exitCodes = Invoke-Docker wait $coord $one $two
        Save-ContainerLog $coord "$case-coordinator.log"
        Save-ContainerLog $one "$case-nodo-1.log"
        Save-ContainerLog $two "$case-nodo-2.log"
        if (@($exitCodes).Count -ne 3 -or @($exitCodes | Where-Object { $_ -ne '0' }).Count -gt 0) {
            throw "$case failed with container exit codes $exitCodes"
        }
        $result = Get-Content -Raw (Join-Path $PSScriptRoot $resultName) | ConvertFrom-Json
        $rows = 0
        foreach ($line in [System.IO.File]::ReadLines($localCSV)) {
            if ($line.Length -gt 0) { $rows++ }
        }
        $rows-- # CSV header
        $counted = ($result.counts | Measure-Object -Sum).Sum
        if ($result.status -ne 'completed' -or $result.rows -ne $rows -or $result.rows -ne $counted -or $result.participants.Count -ne 2) {
            throw "$case produced an inconsistent result"
        }
        $sha256 = (Get-FileHash -Algorithm SHA256 -LiteralPath $localCSV).Hash.ToLowerInvariant()
        Save-Log "$case-verificacion.log" @("filas_csv=$rows", "filas_resultado=$($result.rows)", "suma_conteos=$counted", "sha256_csv=$sha256", "salidas_coordinador_nodo1_nodo2=$($exitCodes -join ',')")
    } finally {
        if ($containers.Count -gt 0) { Invoke-Docker rm -f @containers | Out-Host }
    }
}

Invoke-Docker network create $network | Out-Host
try {
    Invoke-Docker volume create $volume | Out-Host
    try {
        $version = Invoke-Docker run --rm -t $image go version
        $digest = Invoke-Docker image inspect --format '{{index .RepoDigests 0}}' $image
        Save-Log 'entorno.log' @("imagen=$digest", $version)
        $tests = Invoke-Docker run --rm -t --mount $moduleMount -w /app $image sh -ec 'go test ./... && go test -race ./... && go vet ./...'
        Save-Log 'go-test-vet.log' $tests
        Invoke-Docker run --rm --mount $moduleMount --mount $binMount -w /app $image sh -ec 'go build -o /binaries/coordinator ./cmd/coordinator && go build -o /binaries/node ./cmd/node'

        Invoke-Case sintetico /evidence/sintetico.csv (Join-Path $PSScriptRoot 'sintetico.csv') 2
        foreach ($period in @('1_hora', '24_horas')) {
            $file = "features_$period.csv"
            $localCSV = Join-Path $root "data\vectors\$file"
            if (Test-Path -LiteralPath $localCSV -PathType Leaf) {
                Invoke-Case "real-$period" "/vectors/$file" $localCSV 4
            } else {
                Write-Host "Omitido real-$period. No se encontro $localCSV"
            }
        }
    } finally {
        Invoke-Docker volume rm $volume | Out-Host
    }
} finally {
    Invoke-Docker network rm $network | Out-Host
}
