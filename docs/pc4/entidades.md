# Entidades e índices

Responsable: Valentina Montenegro (A1, parte intercambiada con Alexandra).

Se reutiliza el código del Colab dentro del módulo existente `go/plataforma`.

`datasets._id` se relaciona con `jobs.dataset_id`. `jobs._id` es el `job_id`.

| Colección | Índices |
|---|---|
| datasets | Identificador único y combinación única SHA256 + periodo |
| jobs | Identificador único, dataset + fecha, estado + fecha |

El resultado se guarda dentro del trabajo para publicar el estado final y sus valores juntos. Los CSV quedan fuera de MongoDB. Go valida dimensiones, números finitos y suma de conteos.
