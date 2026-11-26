# Latencia de la API

Fecha: 2026-11-26 14:30:05 -05 · Base: http://127.0.0.1:8090/api · Solicitudes por escenario: 30 · Cliente: Go go1.27.1 windows/amd64

| Escenario | Ruta | n | Bytes (P50 por rango más cercano) | Primera (ms) | P50 por rango más cercano (ms) | p95 (ms) | Máx (ms) | Errores | Fuente | < 100 ms |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|---|
| salud | GET /health | 30 | 61 | 13.45 | 3.89 | 5.38 | 13.45 | 0 |  | Sí |
| catalogo | GET /datasets | 30 | 667 | 2.96 | 2.57 | 3.54 | 4.58 | 0 |  | Sí |
| aceptacion | POST /jobs | 30 | 303 | 6.31 | 4.97 | 8.54 | 15.79 | 0 |  | Sí |
| resultado_sin_cache | GET /jobs/{id}/results | 30 | 648 | 13.92 | 2.94 | 4.18 | 13.92 | 0 | MongoDB (30) | Sí |
| resultado_con_cache | GET /jobs/{id}/results | 30 | 646 | 1.61 | 1.62 | 2.20 | 2.54 | 0 | Redis (30) | Sí |
| estado | GET /jobs/{id} | 30 | 335 | 1.61 | 1.62 | 2.20 | 2.28 | 0 | Redis (30) | Sí |
| no_encontrado | GET /jobs/{id} | 30 | 80 | 2.18 | 2.13 | 2.69 | 2.69 | 0 |  | Sí |

P50 y p95 seleccionan la posición `techo(p × n)` de la muestra ordenada, contando desde 1.
Con 30 observaciones corresponden a las posiciones 15 y 29.

La latencia es el tiempo de ida y vuelta medido por el cliente. No incluye el entrenamiento,
que se ejecuta de forma asíncrona y se informa en `resultado.segundos_ml`.
