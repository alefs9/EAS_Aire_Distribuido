# API coordinadora en Go

Responsable: Nicole Bances (N1).

La API es un servidor HTTP escrito con `net/http` dentro del módulo `go/plataforma`. También actúa como **coordinador del clúster**: en el mismo proceso inicia `plataforma.NewCoordinator(...).Serve(...)`, los nodos se registran por TCP y cada trabajo se ejecuta con `Run(...)`. Los trabajos se guardan con el paquete `persistencia` (MongoDB y caché Redis).

## Recorrido de un trabajo

1. `POST /jobs` valida el cuerpo, registra el trabajo en MongoDB como `queued` y lo coloca en una cola acotada (canal con capacidad `QUEUE_SIZE`). Responde **202** sin esperar el entrenamiento.
2. Un único worker toma el trabajo, porque el coordinador ejecuta un trabajo a la vez. Lo marca como `running` y llama a `Run` con el CSV del periodo, K, semilla, tolerancia, máximo de iteraciones y número de nodos.
3. El clúster devuelve centroides, conteos, iteraciones, motivo y tiempos. La API calcula la **inercia** en paralelo con goroutines, porque el clúster no la devuelve, y guarda el resultado con `Completar`. Si el clúster falla o vence `JOB_TIMEOUT`, el trabajo pasa a `failed` con el motivo.
4. `GET /jobs/{id}` y `GET /jobs/{id}/results` usan `Servicio.Resultado`. Un resultado completado se sirve desde Redis mientras dura el TTL; si no está en Redis o Redis no responde, se lee de MongoDB. Un trabajo pendiente nunca se lee de la caché.

El `job_id` externo se mantiene en las rutas HTTP, en `jobs._id` de MongoDB y en la clave de Redis. El coordinador genera otro identificador para el cálculo. Al persistir el resultado, la API registra en `resultado.origen` el `cluster_job` interno y los IDs de los nodos. La [ejecución de Newman con servicios reales](../../results/api/postman/docker-20261122-pc4/RESUMEN.md) conserva las respuestas de ambos periodos y la segunda lectura del resultado de 24 horas desde Redis.

## Rutas

| Método | Ruta | Respuesta correcta | Otros códigos |
|---|---|---|---|
| GET | `/health` | 200 (`ok` o `degradado` si falta Redis) | 503 si MongoDB no responde |
| GET | `/datasets` | 200 con metadatos; `datasets: []` confirma catálogo vacío | 503 si el catálogo no está disponible |
| POST | `/jobs` | 202 + `Location: /jobs/{id}` | 400 parámetros o JSON inválido, 429 límite alcanzado, 503 cola llena o MongoDB caído |
| GET | `/jobs/{id}` | 200 con `trabajo` y `fuente` | 404, 503 |
| GET | `/jobs/{id}/results` | 200 con resultado resumido | 404, 409 pendiente, 422 fallido, 503 |

El cuerpo de `POST /jobs` admite `periodo` (`1_hora` o `24_horas`), `k` (1-20), `semilla`, `tolerancia` (≥ 0), `max_iter` (1-1000) y `nodos` (1-8). Todos son opcionales; los valores por defecto son `1_hora`, 4, 42, 0.0001, 100 y `CLUSTER_NODES`. Se rechazan campos desconocidos.

## Datasets

Al arrancar, la API lee `features_1_hora.csv` y `features_24_horas.csv` de `DATA_DIR`, calcula su SHA-256 y los registra en MongoDB con el identificador `sha256:periodo`, la misma convención de `cmd/persistencia`. Los CSV no se cargan en MongoDB.

## Configuración

| Variable | Por defecto | Uso |
|---|---|---|
| `API_ADDR` | `127.0.0.1:8080` | Dirección HTTP |
| `COORDINATOR_LISTEN` | `127.0.0.1:9000` | Dirección TCP donde se registran los nodos |
| `DATA_DIR` | `../../data/vectors` | Carpeta de los CSV de vectores |
| `CLUSTER_NODES` | `2` | Nodos por trabajo si la solicitud no los indica |
| `QUEUE_SIZE` | `8` | Capacidad de la cola |
| `JOB_TIMEOUT` | `3m` | Plazo total de un trabajo, incluida la espera de nodos |
| `STORE` | `mongo` | `mongo` (MongoDB + Redis) o `memoria` (desarrollo, sin bases de datos) |
| `MONGODB_URI` | `mongodb://127.0.0.1:27017` | Conexión a MongoDB |
| `MONGODB_DATABASE` | `eas_pc4` | Base de datos y prefijo de las claves Redis |
| `REDIS_ADDR` | `127.0.0.1:6379` | Conexión a Redis |
| `REDIS_PASSWORD` | vacío | No se guarda en Git |
| `CACHE_TTL_SECONDS` | `60` | Duración de un resultado en Redis |
| `JOB_POSTS_PER_MINUTE` | `10` | Máximo de solicitudes `POST /jobs` por IP y minuto; `0` desactiva el límite solo para pruebas locales aisladas |

## Ejecución local (PowerShell, desde la raíz del repositorio)

```powershell
.\scripts\api\servicios.ps1 iniciar        # MongoDB y Redis en Docker
cd go\plataforma
go run ./cmd/api                           # terminal 1: API + coordinador
```

```powershell
.\scripts\api\nodo-en-bucle.ps1 -Id nodo-1 # terminal 2
.\scripts\api\nodo-en-bucle.ps1 -Id nodo-2 # terminal 3
```

Cada nodo termina después de un trabajo, según el contrato TCP; el script lo vuelve a registrar para el siguiente. Para probar sin bases de datos: `$env:STORE = "memoria"` antes de `go run ./cmd/api`.

Ejemplo:

```powershell
Invoke-RestMethod -Method Post http://localhost:8080/jobs -ContentType "application/json" -Body '{"periodo":"24_horas","k":4}'
Invoke-RestMethod http://localhost:8080/jobs/<job_id>/results
```

## Pruebas

```powershell
cd go\plataforma
go test -count=1 ./internal/api/
go test -race -count=1 ./internal/api/
```

- `modelo_test.go`: valores por defecto y rechazo de parámetros fuera de rango.
- `http_test.go`: 202, 409 mientras está pendiente, 200 con inercia correcta, 422 si el clúster falla, 400, 404, 503 con la cola llena y salud.
- `integracion_test.go`: levanta el coordinador real y **dos nodos reales por TCP** en el mismo proceso, crea un trabajo por HTTP y verifica que los conteos conservan las filas y que la inercia es correcta.

## Concurrencia

- **Cola acotada:** un canal con buffer y un worker fijo. No se crea una goroutine por solicitud; si la cola se llena, la API responde 503.
- **Memoria:** el almacén en memoria usa `sync.RWMutex`.
- **Inercia:** se divide en bloques contiguos, uno por goroutine, con sumas parciales en posiciones propias del arreglo y una reducción en orden fijo.
- **Plazos:** cada consulta HTTP a las bases de datos tiene un plazo de 2 s, `/health` de 500 ms y cada trabajo de `JOB_TIMEOUT`.
- **Cierre ordenado:** `Shutdown` del servidor HTTP, cierre de la cola, espera del worker y cierre del coordinador.

## Seguridad y acceso web

La interfaz web usa un proxy de mismo origen para enviar las solicitudes bajo `/api/*` a la API. El destino se configura en el servidor mediante `UI_API_URL`; el navegador no puede elegirlo. En este despliegue no se habilita CORS. La UI es el único servicio publicado y la comunicación entre la UI y la API ocurre dentro de la red del servidor.

La API no implementa autenticación ni autorización. Este demo está limitado a `localhost`, loopback o una red de laboratorio y no debe exponerse públicamente. CORS no sustituye la autenticación. Para un despliegue externo se debe agregar autenticación en el servidor y TLS/HTTPS en un proxy frontal. La API actual no implementa WebSocket.

La API limita el cuerpo de `POST /jobs` a 16 KiB, rechaza propiedades no admitidas y aplica límites de tiempo a las solicitudes, consultas y trabajos. Los errores internos devuelven mensajes genéricos al cliente; el detalle técnico se conserva únicamente en los logs del servidor.


`GET /datasets` devuelve `200` con `datasets: []` cuando el catálogo está inicializado pero vacío. Devuelve `503` cuando el catálogo no está disponible porque no se inicializó. `vector_count` cuenta filas/vectores y `measurement_count` cuenta los dos valores escalares PM10 y PM2.5 por fila. El manifiesto no contiene `source`, así que la API no informa ese campo.

### Catálogo de conjuntos: `GET /datasets`

El endpoint devuelve metadatos del catálogo y no envía filas de los archivos CSV. `vector_count` es el número de filas o vectores; `measurement_count` cuenta los dos valores escalares PM10 y PM2.5 de cada vector.

Un `200` con `datasets: []` significa que el catálogo se inicializó correctamente y no contiene conjuntos. Un `503` significa que el catálogo no está disponible porque no se inicializó. El manifiesto no contiene `source`, por lo que la API no informa ese campo.