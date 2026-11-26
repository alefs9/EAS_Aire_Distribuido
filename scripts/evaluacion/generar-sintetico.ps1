param([string]$Output = "data\evaluacion-sintetica")

$ErrorActionPreference = 'Stop'
New-Item -ItemType Directory -Force -Path $Output | Out-Null
$resolvedOutput = (Resolve-Path $Output).Path
$utf8 = [System.Text.UTF8Encoding]::new($false)

foreach ($period in @('1_hora', '24_horas')) {
    $lines = [System.Collections.Generic.List[string]]::new()
    $lines.Add('PM10,PM2.5')

    for ($i = 0; $i -lt 200; $i++) {
        $group = [math]::Floor($i / 50)
        $x = 10 + 80 * $group + ($i % 10) * 0.25
        $y = 4 + 20 * $group + ($i % 7) * 0.2
        $line = [string]::Format(
            [Globalization.CultureInfo]::InvariantCulture,
            '{0:F2},{1:F2}',
            $x,
            $y
        )
        $lines.Add($line)
    }

    $csvPath = Join-Path $resolvedOutput "features_$period.csv"
    [IO.File]::WriteAllText($csvPath, ($lines -join "`n") + "`n", $utf8)
}

$inventory = foreach ($period in @('1_hora', '24_horas')) {
    $csvPath = Join-Path $resolvedOutput "features_$period.csv"
    $sha = (Get-FileHash -Algorithm SHA256 -LiteralPath $csvPath).Hash.ToLowerInvariant()

    [pscustomobject]@{
        period       = $period
        vector_count = 200
        sha256       = $sha
    }
}

$periodManifest = foreach ($item in $inventory) {
    $periodName = if ($item.period -eq '1_hora') { '1 hora' } else { '24 horas' }

    [ordered]@{
        periodo         = $periodName
        filas           = $item.vector_count
        sha256_features = $item.sha256
    }
}

$manifest = [ordered]@{
    orden_features = @('PM10', 'PM2.5')
    unidad         = ''
    cobertura      = ''
    por_periodo    = @($periodManifest)
}

$manifestJson = $manifest | ConvertTo-Json -Depth 5
[IO.File]::WriteAllText(
    (Join-Path $resolvedOutput 'manifest.json'),
    $manifestJson + "`n",
    $utf8
)

$inventory | Format-Table

$firstRows = Get-Content (Join-Path $resolvedOutput 'features_1_hora.csv') -TotalCount 2
if ($firstRows[0] -ne 'PM10,PM2.5') {
    throw 'Encabezado CSV sintético incorrecto.'
}
if ($firstRows[1] -notmatch '^\d+(\.\d+)?,-?\d+(\.\d+)?$') {
    throw 'Formato CSV sintético incorrecto.'
}