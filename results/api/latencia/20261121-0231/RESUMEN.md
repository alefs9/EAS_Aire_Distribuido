# Latencia de la API

Fecha: 2026-11-21 02:32:30 -05 · Base: http://localhost:8080 · Solicitudes por escenario: 30 · Cliente: Go go1.27.1 windows/amd64

| Escenario | Ruta | n | Bytes (mediana) | Primera (ms) | Mediana (ms) | p95 (ms) | Máx (ms) | Errores | Fuente | < 100 ms |
|---|---|---:|---:|---:|---:|---:|---:|---:|---|---|
| salud | GET /health | 30 | 61 | 42.82 | 3.42 | 4.42 | 42.82 | 0 |  | Sí |
| aceptacion | POST /jobs | 30 | 303 | 9.54 | 6.16 | 8.60 | 9.54 | 0 |  | Sí |
| resultado_sin_cache | GET /jobs/{id}/results | 30 | 642 | 5.03 | 1.79 | 4.23 | 5.03 | 0 | Redis (30) | Sí |
| resultado_con_cache | GET /jobs/{id}/results | 30 | 642 | 1.61 | 1.09 | 1.66 | 1.80 | 0 | Redis (30) | Sí |
| estado | GET /jobs/{id} | 30 | 335 | 1.07 | 1.09 | 1.67 | 1.70 | 0 | Redis (30) | Sí |
| no_encontrado | GET /jobs/{id} | 30 | 80 | 2.12 | 2.22 | 2.77 | 2.94 | 0 |  | Sí |

La latencia es el tiempo de ida y vuelta medido por el cliente. No incluye el entrenamiento,
que se ejecuta de forma asíncrona y se informa en `resultado.segundos_ml`.
