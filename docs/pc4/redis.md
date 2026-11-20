# Redis y pruebas

Responsable: Valentina Montenegro (A3).

Redis guarda resultados completados durante el TTL configurado. Si la copia no existe o vence, se recupera MongoDB. Si Redis no responde, se consulta MongoDB. Los trabajos pendientes y fallidos no se presentan como resultados completados.

## Ejecutar
Desde `go/plataforma`, con MongoDB y Redis disponibles:

```powershell
$env:MONGODB_URI = "mongodb://127.0.0.1:27018"
$env:MONGODB_DATABASE = "eas_pc4_pruebas"
$env:REDIS_ADDR = "127.0.0.1:6380"
$env:CACHE_TTL_SECONDS = "3"
$env:PC4_INTEGRATION = "1"
go test -count=1 ./...
go test -race -count=1 ./...
go run ./cmd/persistencia -dataset dataset_1_hora.json -reporte reporte_1_hora.json -salida evidencia_1_hora.json
```

Los JSON de entrada los genera el Colab. Repetir con `24_horas` para el otro periodo. `-race` requiere compilador C. Sin `PC4_INTEGRATION=1`, las pruebas que requieren servicios se omiten explícitamente.

## Colab
[Cuaderno ejecutado y salidas](https://colab.research.google.com/drive/1buYE0xcebFx30bH_K65zK4uSz4u_VuUb).

La copia en `notebooks/Persistencia_cache.ipynb` permite repetir el proceso. El código del paquete conserva la lógica del Colab; solo se cambia `package main` por `persistencia` y se separa el comando para que el equipo pueda importarlo en el módulo existente.

Los datos de MongoDB sobreviven a eliminar la caché, pero no a eliminar la máquina temporal de Colab. La integración del coordinador y de la API con este paquete queda pendiente; estas pruebas no acreditan latencia HTTP.
