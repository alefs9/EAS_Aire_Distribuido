# Medición de latencia de la API

Responsable: Nicole Bances (N3).

## Qué se mide

La latencia es el tiempo de ida y vuelta de una solicitud HTTP medido por el cliente. **No es el tiempo de entrenamiento:** `POST /jobs` responde 202 al registrar el trabajo, y el cálculo en el clúster se informa por separado en `resultado.segundos_ml`.

| Escenario | Solicitud | Condición |
|---|---|---|
| `salud` | `GET /health` | Incluye ping a MongoDB y Redis |
| `aceptacion` | `POST /jobs` | Registro en MongoDB y encolado; cada trabajo se ejecuta después en el clúster |
| `resultado_sin_cache` | `GET /jobs/{id}/results` | Primera lectura de 30 trabajos distintos (`fuente: MongoDB`) |
| `resultado_con_cache` | `GET /jobs/{id}/results` | 30 lecturas del mismo trabajo (`fuente: Redis`) |
| `estado` | `GET /jobs/{id}` | 30 trabajos distintos |
| `no_encontrado` | `GET /jobs/{id}` | Identificador inexistente (404 esperado) |

Para cada escenario se informan n, P50 por rango más cercano del tamaño de respuesta, primera solicitud, P50 por rango más cercano y p95 de los tiempos, máximo, errores y fuente. P50 y p95 ordenan las observaciones y seleccionan la posición `techo(p × n)`, contando desde 1. Con 30 observaciones, el P50 toma la posición 15 sin promediarla con la 16, y el p95 toma la posición 29. Un escenario cumple si su máximo es menor a 100 ms y no hay errores. Los tiempos individuales quedan en `mediciones.csv`.

## Cómo repetir la campaña

Con MongoDB, Redis, la API y los dos nodos en marcha (ver `docs/pc4/api.md`), desde la raíz del repositorio:

```powershell
.\scripts\api\medir-latencia.ps1 -N 30 -Periodo 24_horas
```

Los resultados se guardan en `results/api/latencia/<fecha>/`: `RESUMEN.md`, `mediciones.csv` y `entorno.txt`. Los 30 trabajos de `POST /jobs` se ejecutan de verdad en el clúster; el script espera a que terminen antes de medir las lecturas.

La carpeta «3. Latencia» de la colección realiza otras 30 iteraciones de tres rutas con un trabajo ya completado. Su [reporte de Newman](../../results/api/postman/docker-20261122-pc4/latencia.json) pertenece a una ejecución independiente.

## Resultados

En la [campaña publicada del cliente Go](../../results/api/latencia/20261121-1230/RESUMEN.md), las 180 solicitudes de los seis escenarios cumplieron el umbral estricto de 100 ms. El máximo observado fue 47.95 ms. La prueba con MongoDB detenido se ejecutó aparte: las dos solicitudes devolvieron HTTP 503 y la creación tardó aproximadamente 2 segundos. [Reporte de indisponibilidad](../../results/api/postman/docker-20261122-pc4/sin-mongodb.json)
