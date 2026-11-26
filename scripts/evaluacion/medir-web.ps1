param(
    [string]$BaseUrl = 'http://127.0.0.1:8090/api',
    [string]$Output = 'results/evaluacion-web/run-local',
    [int]$Repetitions = 20,
    [int]$PollMs = 100,
    [string[]]$Periods = @('1_hora', '24_horas'),
    [int[]]$NodeCounts = @(1, 2)
)

$ErrorActionPreference = 'Stop'
New-Item -ItemType Directory -Force -Path $Output | Out-Null

$catalog = Invoke-RestMethod -Uri "$BaseUrl/datasets" -TimeoutSec 15
$catalogByPeriod = @{}
foreach ($dataset in $catalog.datasets) {
    $catalogByPeriod[[string]$dataset.period] = $dataset
}

$commit = (git rev-parse HEAD 2>$null)
if (-not $commit) { $commit = 'no-disponible' }

$dataHostDir = if ($env:DATA_HOST_DIR) { $env:DATA_HOST_DIR } else { '../../data/vectors' }
$cpu = (Get-CimInstance Win32_Processor | ForEach-Object { $_.Name } | Select-Object -Unique) -join '; '
$ramGiB = [math]::Round(((Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory / 1GB), 2)

$environment = [ordered]@{
    captured_at           = (Get-Date).ToString('o')
    commit                = $commit
    os                    = (Get-CimInstance Win32_OperatingSystem).Caption
    cpu                   = $cpu
    ram_gib               = $ramGiB
    go                    = (go version)
    docker_compose        = (docker compose version)
    data_host_dir         = $dataHostDir
    job_posts_per_minute  = $env:JOB_POSTS_PER_MINUTE
    queue_size            = $env:QUEUE_SIZE
    job_timeout           = $env:JOB_TIMEOUT
}

$imageIds = @(docker compose -f go/plataforma/compose.yml images -q | Sort-Object -Unique)
$environment['compose_images'] = @(
    foreach ($imageId in $imageIds) {
        docker image inspect $imageId --format '{{.Id}} | {{json .RepoTags}} | {{json .RepoDigests}}'
    }
)

$environment |
    ConvertTo-Json -Depth 5 |
    Set-Content -Encoding utf8 (Join-Path $Output 'entorno.json')

$rows = [System.Collections.Generic.List[object]]::new()

foreach ($period in $Periods) {
    foreach ($nodes in $NodeCounts) {
        # Repetición 0: calentamiento, no se agrega a las mediciones.
        for ($rep = 0; $rep -le $Repetitions; $rep++) {
            $body = @{
                periodo   = $period
                k         = 4
                semilla   = 42
                tolerancia = 0.0001
                max_iter  = 100
                nodos     = $nodes
            } | ConvertTo-Json -Compress

            $total = [Diagnostics.Stopwatch]::StartNew()
            $accept = [Diagnostics.Stopwatch]::StartNew()

            try {
                $created = Invoke-RestMethod `
                    -Method Post `
                    -Uri "$BaseUrl/jobs" `
                    -ContentType 'application/json' `
                    -Body $body `
                    -TimeoutSec 20

                $accept.Stop()

                if (-not $created.job_id) {
                    throw 'POST /jobs no devolvió job_id.'
                }

                $id = [string]$created.job_id
                $state = ''
                $errorText = $null
                $result = $null
                $limit = [DateTime]::UtcNow.AddMinutes(10)

                while ([DateTime]::UtcNow -lt $limit) {
                    try {
                        $stateBody = Invoke-RestMethod -Uri "$BaseUrl/jobs/$id" -TimeoutSec 15
                        $state = [string]$stateBody.trabajo.estado
                        if ($state -in @('completed', 'failed')) { break }
                    } catch {
                        Start-Sleep -Milliseconds $PollMs
                        continue
                    }

                    Start-Sleep -Milliseconds $PollMs
                }

                if ($state -eq 'completed') {
                    $result = Invoke-RestMethod -Uri "$BaseUrl/jobs/$id/results" -TimeoutSec 20
                } elseif ($state -eq 'failed') {
                    $errorText = 'El trabajo terminó fallido; GET /results debe devolver 422.'
                } else {
                    $state = 'timeout'
                    $errorText = 'El trabajo no llegó a estado terminal en 10 minutos.'
                }

                $total.Stop()
                $timings = $result.resultado

                $row = [pscustomobject]@{
                    timestamp       = (Get-Date).ToString('o')
                    commit          = $commit
                    period          = $period
                    nodes           = $nodes
                    repetition      = $rep
                    k               = 4
                    seed            = 42
                    tolerance       = 0.0001
                    max_iter        = 100
                    job_id          = $id
                    dataset_id      = $created.dataset_id
                    vector_count    = $catalogByPeriod[$period].vector_count
                    sha256          = $catalogByPeriod[$period].sha256
                    http_status     = 202
                    acceptance_ms   = [math]::Round($accept.Elapsed.TotalMilliseconds, 3)
                    load_ms         = if ($timings) { [math]::Round(1000 * $timings.segundos_carga, 3) } else { $null }
                    training_ms     = if ($timings) { [math]::Round(1000 * $timings.segundos_ml, 3) } else { $null }
                    total_ms        = [math]::Round($total.Elapsed.TotalMilliseconds, 3)
                    poll_interval_ms = $PollMs
                    state           = $state
                    cluster_origin  = if ($timings) { $timings.origen } else { $null }
                    error           = $errorText
                }
            } catch {
                $accept.Stop()
                $total.Stop()

                $row = [pscustomobject]@{
                    timestamp        = (Get-Date).ToString('o')
                    commit           = $commit
                    period           = $period
                    nodes            = $nodes
                    repetition       = $rep
                    k                = 4
                    seed             = 42
                    tolerance        = 0.0001
                    max_iter         = 100
                    job_id           = $null
                    dataset_id       = $null
                    vector_count     = $catalogByPeriod[$period].vector_count
                    sha256           = $catalogByPeriod[$period].sha256
                    http_status      = $null
                    acceptance_ms    = $null
                    load_ms          = $null
                    training_ms      = $null
                    total_ms         = [math]::Round($total.Elapsed.TotalMilliseconds, 3)
                    poll_interval_ms = $PollMs
                    state            = 'request_error'
                    cluster_origin   = $null
                    error            = $_.Exception.Message
                }
            }

            if ($rep -eq 0) { continue }

            $rows.Add($row)
            $safe = "$period-$nodes-n$rep"

            $row |
                ConvertTo-Json -Depth 8 |
                Set-Content -Encoding utf8 (Join-Path $Output "$safe.json")

            $row | Format-Table -AutoSize
        }
    }
}

$rows | Export-Csv -NoTypeInformation -Encoding utf8 (Join-Path $Output 'mediciones.csv')

$summary = foreach ($group in ($rows | Where-Object { $_.state -eq 'completed' } | Group-Object period, nodes)) {
    $measure = $group.Group

    foreach ($name in @('acceptance_ms', 'load_ms', 'training_ms', 'total_ms')) {
        $values = @(
            $measure |
                ForEach-Object { $_.$name } |
                Where-Object { $null -ne $_ } |
                Sort-Object { [double]$_ }
        )

        if ($values.Count -eq 0) { continue }

        $p50Index = [math]::Ceiling(0.50 * $values.Count) - 1
        $p95Index = [math]::Ceiling(0.95 * $values.Count) - 1

        [pscustomobject]@{
            group  = $group.Name
            metric = $name
            n      = $values.Count
            p50    = [double]$values[$p50Index]
            p95    = [double]$values[$p95Index]
            method = 'nearest-rank: ceil(p*n), base 1'
        }
    }
}

$summary |
    ConvertTo-Json -Depth 5 |
    Set-Content -Encoding utf8 (Join-Path $Output 'resumen.json')

$summary | Export-Csv -NoTypeInformation -Encoding utf8 (Join-Path $Output 'resumen.csv')