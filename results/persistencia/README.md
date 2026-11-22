# Evidencias de persistencia y caché

## Demostración histórica de Colab

Ejecución observada el 20 de noviembre de 2026. Go 1.22.2, MongoDB 8.0.38, Redis 7.0.15. Las ocho pruebas pasaron tanto en ejecución normal como con `-race`.

| Periodo | Vectores | Iteraciones de referencia | Persistencia y caché |
|---|---:|---:|---|
| 1 hora | 445347 | 96 | PASS |
| 24 horas | 18297 | 44 | PASS |

Se comprobaron estados pendientes, primera consulta en MongoDB, lectura en Redis, expiración e invalidación. Los centroides, conteos, inercia e iteraciones recuperados coinciden con la referencia secuencial PC3. Los tiempos de esa referencia no son mediciones del clúster PC4.

[Cuaderno original con salidas guardadas](https://colab.research.google.com/drive/1buYE0xcebFx30bH_K65zK4uSz4u_VuUb). La carpeta [colab-20261120](colab-20261120/) conserva las salidas recuperadas, los documentos de MongoDB tal como se imprimieron y los datos de procedencia. El cuaderno conserva los resultados de las pruebas, pero no registra sus códigos numéricos de salida. Los archivos independientes de la máquina temporal de Colab no se localizaron.

## Reproducción en Docker Desktop del 22 de noviembre de 2026

En [docker-20261122-pc4](docker-20261122-pc4/) se guardan los metadatos de cada CSV, los reportes de la referencia Go, los documentos MongoDB, los valores Redis, los TTL y las verificaciones con sus logs. Se usaron Go 1.27.4 en Linux, MongoDB 8.3.17 y Redis 7.4.12. Los SHA-256 corresponden al manifiesto de vectores.

| Periodo | Vectores | Iteraciones de la referencia | Consultas y estados |
|---|---:|---:|---|
| Una hora | 445 347 | 96, convergió | [JSON](docker-20261122-pc4/evidencia-1_hora.json) |
| 24 horas | 18 297 | 44, convergió | [JSON](docker-20261122-pc4/evidencia-24_horas.json) |

La demostración usa K=4, semilla 42, tolerancia 0.001, máximo 100 y TTL de 3 s. Ambos periodos comprobaron `queued` y `running` sin caché, la secuencia de lecturas MongoDB → Redis → MongoDB tras expirar → MongoDB tras invalidar, y el documento de un trabajo fallido. La [integración por API y TCP](../api/postman/docker-20261122-pc4/RESUMEN.md) ejecutó otros trabajos, con tolerancia 0.0001 y su propio TTL.

Desde la raíz, con Docker Desktop en modo Linux y Node.js con `npx.cmd`:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\api\cerrar-evidencias.ps1
```

El [script](../../scripts/api/cerrar-evidencias.ps1) registra las salidas por comando y crea una red y una base aisladas. Detiene sus contenedores al terminar y conserva los datos. En la reproducción archivada, el sondeo del reinicio de MongoDB interrumpió el script tras guardar los reportes de Newman. Se corrigió el sondeo y se completaron las pruebas con `-RunId docker-20261122-pc4 -SoloPruebas`, en una base de pruebas separada. El modo de reanudación se aplica a una ejecución cuyos logs de pruebas aún falten.

Las pruebas normales, con `-race` y `go vet` terminaron con salida 0 en `go/plataforma` y `go/kmeans_secuencial`. El primer pase secuencial omitió la prueba optativa de los dos CSV; sus logs se conservan por separado del pase con `KMEANS_REAL_DATA=1`. Las pruebas de persistencia con `PC4_INTEGRATION=1` usaron MongoDB y Redis activos.
