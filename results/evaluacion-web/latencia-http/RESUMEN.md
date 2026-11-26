# Latencia de la API

Fecha: 2026-11-26 13:11:26 -05 · Base: http://127.0.0.1:8090/api · Solicitudes por escenario: 30 · Cliente: Go go1.27.1 windows/amd64

| Escenario | Ruta | n | Bytes (P50 por rango más cercano) | Primera (ms) | P50 por rango más cercano (ms) | p95 (ms) | Máx (ms) | Errores | Fuente | < 100 ms |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|---|
| salud | GET /health | 30 | 61 | 5.33 | 1.69 | 5.33 | 5.99 | 0 |  | Sí |
| catalogo | GET /datasets | 930 | 511 | 1.71 | 1.20 | 4.80 | 23.16 | 0 |  | Sí |
| aceptacion | POST /jobs | 30 | 303 | 3.25 | 3.02 | 6.05 | 8.17 | 0 |  | Sí |
| resultado_sin_cache | GET /jobs/{id}/results | 30 | 572 | 15.04 | 2.80 | 8.28 | 15.04 | 0 | MongoDB (30) | Sí |
| resultado_con_cache | GET /jobs/{id}/results | 30 | 570 | 1.69 | 1.52 | 2.22 | 2.25 | 0 | Redis (30) | Sí |
| estado | GET /jobs/{id} | 30 | 335 | 1.11 | 1.65 | 3.25 | 4.48 | 0 | Redis (30) | Sí |
| no_encontrado | GET /jobs/{id} | 30 | 80 | 1.66 | 1.66 | 2.36 | 2.86 | 0 |  | Sí |

P50 y p95 seleccionan la posición `techo(p × n)` de la muestra ordenada, contando desde 1.
Con 30 observaciones corresponden a las posiciones 15 y 29.

La latencia es el tiempo de ida y vuelta medido por el cliente. No incluye el entrenamiento,
que se ejecuta de forma asíncrona y se informa en `resultado.segundos_ml`.
