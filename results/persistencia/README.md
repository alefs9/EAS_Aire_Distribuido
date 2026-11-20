# Resultados verificados en Colab

Ejecución observada el 20 de noviembre de 2026. Go 1.22.2, MongoDB 8.0.38, Redis 7.0.15. Las ocho pruebas pasaron tanto en ejecución normal como con `-race`.

| Periodo | Vectores | Iteraciones de referencia | Persistencia y caché |
|---|---:|---:|---|
| 1 hora | 445347 | 96 | PASS |
| 24 horas | 18297 | 44 | PASS |

Se comprobaron estados pendientes, primera consulta en MongoDB, lectura en Redis, expiración e invalidación. Los centroides, conteos, inercia e iteraciones recuperados coinciden con la referencia secuencial PC3. Los tiempos de esa referencia no son mediciones del clúster PC4.

[Salidas originales del cuaderno](https://colab.research.google.com/drive/1buYE0xcebFx30bH_K65zK4uSz4u_VuUb). Esta tabla resume lo observado; no sustituye un registro de integración con la API.
