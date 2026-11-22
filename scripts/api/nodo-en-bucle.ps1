# Mantiene activo un nodo del clúster. El nodo termina después de cada trabajo
# (FINISH y BYE, según el contrato TCP) y este bucle lo vuelve a registrar.
# Uso, en una terminal por nodo:
#   .\scripts\api\nodo-en-bucle.ps1 -Id nodo-1
#   .\scripts\api\nodo-en-bucle.ps1 -Id nodo-2
param(
    [Parameter(Mandatory = $true)][string]$Id,
    [string]$Coordinador = "127.0.0.1:9000"
)
$ErrorActionPreference = "Stop"
$modulo = Join-Path $PSScriptRoot "..\..\go\plataforma"
$binario = Join-Path $env:TEMP "eas-nodo-$Id.exe"
Push-Location $modulo
try { go build -o $binario ./cmd/node } finally { Pop-Location }
Write-Host "Nodo $Id conectándose a $Coordinador (Ctrl+C para detener)"
while ($true) {
    & $binario -coordinator $Coordinador -id $Id -interval 200ms -timeout 20s
    Start-Sleep -Seconds 1
}