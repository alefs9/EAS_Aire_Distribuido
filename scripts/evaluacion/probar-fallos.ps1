param(
    [string]$Base = 'http://127.0.0.1:8090/api',
    [string]$Compose = '',
    [string]$Project = 'eas-eval-nicole-20261126'
)

$ErrorActionPreference = 'Stop'

if ([string]::IsNullOrWhiteSpace($Compose)) {
    $Compose = Join-Path $PSScriptRoot '..\..\go\plataforma\compose.yml'
}
$Compose = (Resolve-Path -LiteralPath $Compose).Path

function Get-StatusCode {
    param(
        [Parameter(Mandatory)] [string]$Method,
        [Parameter(Mandatory)] [string]$Uri,
        [string]$Body
    )

    try {
        if ($PSBoundParameters.ContainsKey('Body')) {
            $response = Invoke-WebRequest `
                -UseBasicParsing `
                -Method $Method `
                -Uri $Uri `
                -ContentType 'application/json' `
                -Body $Body `
                -TimeoutSec 15
        } else {
            $response = Invoke-WebRequest `
                -UseBasicParsing `
                -Method $Method `
                -Uri $Uri `
                -TimeoutSec 15
        }

        return [int]$response.StatusCode
    } catch {
        $response = $_.Exception.Response
        if ($null -ne $response -and $null -ne $response.StatusCode) {
            return [int]$response.StatusCode
        }
        throw
    }
}

function Expect-Status {
    param(
        [int]$Actual,
        [int]$Expected,
        [string]$Case
    )

    if ($Actual -ne $Expected) {
        throw "${Case}: se esperaba HTTP $Expected y llegó HTTP $Actual."
    }

    Write-Host "$Case -> HTTP $Actual"
}

function Invoke-ComposeQuiet {
    param(
        [Parameter(Mandatory)] [string[]]$ComposeArgs
    )

    $previousPreference = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'

    try {
        & docker compose -p $Project -f $Compose @ComposeArgs 2>&1 | Out-Null
        if ($LASTEXITCODE -ne 0) {
            throw "Falló docker compose $($ComposeArgs -join ' ')."
        }
    } finally {
        $ErrorActionPreference = $previousPreference
    }
}

# Validaciones: cada cuerpo inválido debe responder 400.
$invalidos = @(
    '{"k":99}',
    '{"tolerancia":-1}',
    '{"max_iter":-1}',
    '{"max_iter":1001}',
    '{"nodos":-1}',
    '{"nodos":9}',
    '{"periodo":"2_horas"}',
    '{"desconocido":1}',
    '{no es json'
)

foreach ($body in $invalidos) {
    $code = Get-StatusCode -Method POST -Uri "$Base/jobs" -Body $body
    Expect-Status -Actual $code -Expected 400 -Case "Entrada inválida $body"
}

# IDs inexistentes.
Expect-Status `
    -Actual (Get-StatusCode -Method GET -Uri "$Base/jobs/pc4-no-existe") `
    -Expected 404 `
    -Case 'Trabajo inexistente'

Expect-Status `
    -Actual (Get-StatusCode -Method GET -Uri "$Base/jobs/pc4-no-existe/results") `
    -Expected 404 `
    -Case 'Resultado inexistente'

# Mongo detenido: health y creación deben responder 503.
try {
    Invoke-ComposeQuiet -ComposeArgs @('stop', 'mongo')

    $mongoDown = $false
    $deadline = (Get-Date).AddSeconds(30)
    while ((Get-Date) -lt $deadline) {
        if ((Get-StatusCode -Method GET -Uri "$Base/health") -eq 503) {
            $mongoDown = $true
            break
        }
        Start-Sleep -Seconds 1
    }

    if (-not $mongoDown) {
        throw 'La API no informó Mongo como no disponible.'
    }

    Expect-Status `
        -Actual (Get-StatusCode -Method POST -Uri "$Base/jobs" -Body '{"periodo":"1_hora","k":4}') `
        -Expected 503 `
        -Case 'Crear trabajo con Mongo detenido'
} finally {
    Invoke-ComposeQuiet -ComposeArgs @('start', 'mongo')
}

# Esperar a que Mongo vuelva.
$mongoReady = $false
$deadline = (Get-Date).AddSeconds(45)

while ((Get-Date) -lt $deadline) {
    if ((Get-StatusCode -Method GET -Uri "$Base/health") -eq 200) {
        $mongoReady = $true
        break
    }
    Start-Sleep -Seconds 1
}

if (-not $mongoReady) {
    throw 'Mongo no volvió a estar disponible.'
}
Write-Host 'Mongo volvió a estar disponible.'

# Crear un trabajo largo y consultar resultados inmediatamente: debe responder 409.
$pending = Invoke-RestMethod `
    -Method POST `
    -Uri "$Base/jobs" `
    -ContentType 'application/json' `
    -Body '{"periodo":"1_hora","k":4,"nodos":2}' `
    -TimeoutSec 20

$pendingId = [string]$pending.job_id
if ([string]::IsNullOrWhiteSpace($pendingId)) {
    throw 'El POST de preparación no devolvió job_id.'
}

Expect-Status `
    -Actual (Get-StatusCode -Method GET -Uri "$Base/jobs/$pendingId/results") `
    -Expected 409 `
    -Case 'Resultado solicitado antes de completar el trabajo'

# Esperar a que el trabajo termine para reutilizarlo en la prueba de Redis.
$state = ''
$deadline = (Get-Date).AddMinutes(3)

while ((Get-Date) -lt $deadline) {
    Start-Sleep -Milliseconds 500
    try {
        $job = Invoke-RestMethod -Uri "$Base/jobs/$pendingId" -TimeoutSec 10
        $state = [string]$job.trabajo.estado
    } catch {
        $state = ''
    }

    if ($state -in @('completed', 'failed')) {
        break
    }
}

if ($state -ne 'completed') {
    throw "El trabajo de preparación terminó en '$state', no en 'completed'."
}

# Redis detenido: health degradado y resultado disponible desde Mongo.
try {
    Invoke-ComposeQuiet -ComposeArgs @('stop', 'redis')

    $health = Invoke-RestMethod -Uri "$Base/health" -TimeoutSec 15
    if ($health.estado -ne 'degradado') {
        throw "Se esperaba estado degradado con Redis detenido; llegó '$($health.estado)'."
    }

    Expect-Status `
        -Actual (Get-StatusCode -Method GET -Uri "$Base/jobs/$pendingId/results") `
        -Expected 200 `
        -Case 'Resultado desde Mongo con Redis detenido'
} finally {
    Invoke-ComposeQuiet -ComposeArgs @('start', 'redis')
}

# Nodos detenidos: el trabajo debe fallar y /results debe responder 422.
try {
    Invoke-ComposeQuiet -ComposeArgs @('stop', 'nodo-1', 'nodo-2')

    $failedJob = Invoke-RestMethod `
        -Method POST `
        -Uri "$Base/jobs" `
        -ContentType 'application/json' `
        -Body '{"periodo":"24_horas","k":4,"nodos":2}' `
        -TimeoutSec 20

    $failedId = [string]$failedJob.job_id
    if ([string]::IsNullOrWhiteSpace($failedId)) {
        throw 'El POST sin nodos no devolvió job_id.'
    }

    $state = ''
    $deadline = (Get-Date).AddSeconds(60)

    while ((Get-Date) -lt $deadline) {
        Start-Sleep -Milliseconds 500
        try {
            $job = Invoke-RestMethod -Uri "$Base/jobs/$failedId" -TimeoutSec 10
            $state = [string]$job.trabajo.estado
        } catch {
            $state = ''
        }

        if ($state -in @('completed', 'failed')) {
            break
        }
    }

    if ($state -ne 'failed') {
        throw "El trabajo sin nodos terminó en '$state'; se esperaba 'failed'."
    }

    Expect-Status `
        -Actual (Get-StatusCode -Method GET -Uri "$Base/jobs/$failedId/results") `
        -Expected 422 `
        -Case 'Resultado de trabajo fallido'
} finally {
    Invoke-ComposeQuiet -ComposeArgs @('start', 'nodo-1', 'nodo-2')
}

Write-Host 'Pruebas 400, 404, 409, 422, Mongo y Redis completadas.'