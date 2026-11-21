# Latencia de la API

Fecha: 2026-11-21 12:32:36 -05 · Base: http://localhost:8080 · Solicitudes por escenario: 30 · Cliente: Go go1.27.1 windows/amd64

| Escenario | Ruta | n | Bytes (mediana) | Primera (ms) | Mediana (ms) | p95 (ms) | Máx (ms) | Errores | Fuente | < 100 ms |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|---|
| salud | GET /health | 30 | 61 | 47.95 | 3.13 | 5.26 | 47.95 | 0 |  | Sí |
| aceptacion | POST /jobs | 30 | 303 | 6.25 | 6.25 | 11.52 | 12.15 | 0 |  | Sí |
| resultado_sin_cache | GET /jobs/{id}/results | 30 | 644 | 12.89 | 9.94 | 15.13 | 19.03 | 0 | MongoDB (30) | Sí |
| resultado_con_cache | GET /jobs/{id}/results | 30 | 642 | 1.12 | 1.09 | 2.24 | 5.67 | 0 | Redis (30) | Sí |
| estado | GET /jobs/{id} | 30 | 335 | 1.06 | 1.06 | 1.10 | 1.30 | 0 | Redis (30) | Sí |
| no_encontrado | GET /jobs/{id} | 30 | 80 | 1.59 | 1.60 | 2.17 | 2.96 | 0 |  | Sí |

La latencia es el tiempo de ida y vuelta medido por el cliente. No incluye el entrenamiento,
que se ejecuta de forma asíncrona y se informa en `resultado.segundos_ml`.
