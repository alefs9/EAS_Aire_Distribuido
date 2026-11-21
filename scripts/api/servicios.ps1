# Inicia, detiene o muestra MongoDB y Redis en Docker para ejecutar la API localmente.
# Uso: .\scripts\api\servicios.ps1 iniciar | detener | estado
param([ValidateSet("iniciar", "detener", "estado")][string]$Accion = "iniciar")
$servicios = @(
    @{ Nombre = "eas-mongo"; Imagen = "mongo:8"; Puerto = "27017:27017" },
    @{ Nombre = "eas-redis"; Imagen = "redis:7"; Puerto = "6379:6379" }
)
foreach ($s in $servicios) {
    $existe = docker ps -a --filter "name=^$($s.Nombre)$" --format "{{.Names}}"
    switch ($Accion) {
        "iniciar" {
            if ($existe) { docker start $s.Nombre | Out-Null }
            else { docker run -d --name $s.Nombre -p $s.Puerto $s.Imagen | Out-Null }
            Write-Host "$($s.Nombre) iniciado ($($s.Puerto))"
        }
        "detener" { if ($existe) { docker stop $s.Nombre | Out-Null; Write-Host "$($s.Nombre) detenido" } }
        "estado"  { docker ps -a --filter "name=^$($s.Nombre)$" --format "{{.Names}}: {{.Status}}" }
    }
}