# Swagger y Postman

Responsable: Nicole Bances (N2).

## Swagger

La especificación OpenAPI 3.0 está en `go/plataforma/internal/api/openapi.yaml`. Se embebe en el binario con `go:embed`, así que siempre corresponde a la versión del código que se ejecuta.

- `GET /openapi.yaml`: especificación.
- `GET /docs`: Swagger UI. Carga los estilos y scripts de `unpkg.com`, así que requiere conexión a internet en el navegador.

La especificación documenta las cuatro rutas, sus parámetros, esquemas, ejemplos reales y todos los códigos de respuesta. La prueba `documentacion_test.go` comprueba que se publica y que contiene todas las rutas del router.

## Postman

Archivos en `docs/pc4/postman/`:

- `EAS_API_PC4.postman_collection.json`: colección.
- `EAS_local.postman_environment.json`: entorno con `base_url`, sin contraseñas.

Importar ambos en Postman y seleccionar el entorno **EAS local**. Antes de ejecutar deben estar activos MongoDB, Redis, la API y los dos nodos (ver `docs/pc4/api.md`).

| Carpeta | Casos | Resultado esperado |
|---|---|---|
| 1. Funcionales | Salud | 200 y MongoDB disponible |
| | Crear ejecución horaria | 202, `queued`, cabecera `Location` |
| | Resultado pendiente | 409 con estado `queued` o `running` |
| | Estado de la ejecución | 200 |
| | Esperar resultado final (horario y 24 horas) | Repite cada segundo mientras responda 409; termina con 200 |
| | Resultado con caché | 200 con `fuente: Redis` |
| | Parámetros inválidos / campo desconocido | 400 |
| | Identificador inexistente | 404 |
| 2. Dependencia no disponible | Salud y creación con MongoDB detenido (`docker stop eas-mongo`) | 503 |
| 3. Latencia | `/health`, `/jobs/{id}`, `/jobs/{id}/results` | Para el Collection Runner con 30 iteraciones |

Las respuestas verificadas por las carpetas funcional y de latencia incluyen el umbral estricto de 100 ms. Las consultas intermedias de espera se registran como solicitudes y el escenario con MongoDB detenido comprueba los 503 esperados sin imponer ese umbral.

Para ejecutar la carpeta 1 completa: **Run collection** → seleccionar solo "1. Funcionales" → **Run**. Ejecutarla en orden, porque las solicitudes comparten `job_horario` y `job_id`.

La ejecución del 22 de noviembre usó Newman 6.2.3. Hay un JSON por carpeta, con códigos HTTP, tiempos y aserciones, y un log con el código de salida. Consulte [el resumen y el procedimiento](../../results/api/postman/docker-20261122-pc4/RESUMEN.md).
