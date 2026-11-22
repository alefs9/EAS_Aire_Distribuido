# Ejecuta la campaña de latencia contra la API en marcha y guarda los resultados
# en results/api/latencia/<fecha>. Requiere la API, MongoDB, Redis y los dos nodos activos.
# Uso: .\scripts\api\medir-latencia.ps1 [-N 30] [-Periodo 24_horas]
param(
    [string]$Base = "http://localhost:8080",
    [int]$N = 30,
    [string]$Periodo = "24_horas"
)
$ErrorActionPreference = "Stop"
$raiz = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$salida = Join-Path $raiz ("results\api\latencia\" + (Get-Date -Format "yyyyMMdd-HHmm"))
try { Invoke-RestMethod "$Base/health" | Out-Null }
catch { throw "La API no responde correctamente en $Base/health. Revise la API y MongoDB." }
Push-Location (Join-Path $raiz "go\plataforma")
try {
    New-Item -ItemType Directory -Force $salida | Out-Null
    go version | Out-File -Encoding utf8 (Join-Path $salida "entorno.txt")
    go run ./cmd/latencia -base $Base -n $N -periodo $Periodo -salida $salida
}
finally { Pop-Location }
Write-Host "`nResultados guardados en $salida"