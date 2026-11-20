# Contrato TCP del clúster

El coordinador escucha y cada nodo inicia una conexión TCP independiente. La topología es en estrella, con un registro de nodos activos en memoria. El formato sigue el encuadre del informe (sección 3.2.11.1). Las mismas conexiones transportan los mensajes de sesión y de cálculo.

## Encuadre

Cada mensaje empieza con 4 bytes que representan la longitud **en bytes del cuerpo**, como entero sin signo en orden big-endian. Le sigue exactamente esa cantidad de bytes con un objeto JSON en UTF-8. Se admiten cuerpos de 1 a 1 048 576 bytes. TCP puede fragmentar o agrupar mensajes. El receptor lee exactamente el prefijo y luego el cuerpo, sin usar saltos de línea como separadores.

## Mensajes y estados

Todos los mensajes incluyen `proto: 1` y `type`. Se rechazan campos desconocidos. `node_id` distingue mayúsculas y minúsculas y admite de 1 a 64 caracteres ASCII (letras, dígitos, guion y guion bajo). Debe coincidir con el identificador de la sesión tras el registro.

| Tipo | Dirección | Campos adicionales |
| --- | --- | --- |
| `REGISTER` | Nodo → coordinador | `node_id`, `capacity` (entero positivo informativo) |
| `REGISTERED` | Coordinador → nodo | `node_id` confirmado |
| `HEARTBEAT` | Nodo → coordinador | `node_id`, `ts` (entero positivo en nanosegundos desde Unix epoch) |
| `HEARTBEAT_ACK` | Coordinador → nodo | Mismos `node_id` y `ts` de la solicitud |
| `BYE` | Nodo → coordinador | `node_id` |
| `BYE_ACK` | Coordinador → nodo | `node_id` liberado |
| `ERROR` | Coordinador → nodo | `code`, `message`, `node_id` si ya se registró la sesión |

Ejemplo de cuerpo para el registro:

```json
{"proto":1,"type":"REGISTER","node_id":"nodo-1","capacity":1}
```

1. El primer mensaje debe ser `REGISTER`. El coordinador reserva el ID de forma atómica antes de enviar `REGISTERED`. Si ya está ocupado, rechaza solo la nueva conexión.
2. El nodo envía un latido periódico y comprueba su confirmación antes de continuar. `ts` correlaciona la respuesta con la solicitud sin exigir relojes sincronizados.
3. Al recibir una señal de interrupción, el nodo completa el intercambio en curso dentro de sus plazos y envía `BYE`. El coordinador libera el ID antes de confirmar con `BYE_ACK` y cierra la conexión. Un EOF, error de transporte o vencimiento también libera el ID. Un nodo puede abrir otra conexión y registrarse de nuevo.
4. Al detenerse, el coordinador cierra el listener y las conexiones y espera a que terminen sus manejadores. El nodo detecta la pérdida de la conexión en el siguiente intercambio y sale con error. La reconexión se inicia manualmente.

## Errores

Tras intentar enviar `ERROR`, el coordinador cierra la conexión. Frente a EOF limpio o un fallo de transporte, la sesión termina directamente. La entrega de un error depende de que el socket aún permita escribir.

| Código | Condición |
| --- | --- |
| `INVALID_MESSAGE` | Longitud fuera de rango, trama truncada, JSON inválido o campos incorrectos |
| `UNSUPPORTED_VERSION` | Versión distinta de 1 |
| `EXPECTED_REGISTER` | Primer mensaje válido distinto de `REGISTER` |
| `DUPLICATE_NODE` | Identificador ocupado por otra sesión |
| `NODE_MISMATCH` | Identificador distinto del registrado |
| `UNEXPECTED_MESSAGE` | Mensaje válido fuera de la secuencia permitida |
| `TIMEOUT` | Se agotó el plazo para leer una trama completa |

## Configuración

- Coordinador: `-listen` (por defecto `127.0.0.1:9000`), `-timeout` (15 s por lectura o escritura).
- Nodo: `-coordinator` (por defecto `127.0.0.1:9000`), `-id` (obligatorio), `-capacity` (1), `-interval` (3 s entre latidos), `-timeout` (5 s por conexión, lectura o escritura) y `-checks` (0 para ejecución continua, o un número positivo de latidos confirmados antes de cerrar).
- El intervalo de latidos debe dejar margen dentro del plazo de lectura del coordinador. Para procesos en otros equipos, el coordinador debe escuchar en una interfaz accesible. El clúster usa TCP sin TLS ni autenticación en una red de confianza.

## Ejecución distribuida

En reposo, el nodo sigue iniciando cada intercambio con `HEARTBEAT`. Cuando el coordinador tiene trabajo para esa sesión, añade `"work":true` a `HEARTBEAT_ACK`; conserva el identificador y la marca temporal. Desde ese instante, el coordinador inicia cada intercambio del trabajo y espera una respuesta antes de enviar el siguiente mensaje a ese socket. Un manejador por conexión en el coordinador y un bucle en el nodo son los únicos lectores y escritores de sus respectivos extremos. Durante el trabajo, los plazos de sus intercambios comprueban la conexión. Ambos ejecutables deben admitir los mensajes de cálculo para participar en el trabajo.

Los mensajes de trabajo incluyen `proto:1`, `type`, `node_id` y un objeto `job` con `id` (identificador de ejecución), `partition` (índice desde cero), `iteration` (cero durante la carga) y `attempt` (siempre 1). La respuesta debe repetir los cuatro identificadores y corresponder al tipo esperado en esa conexión.

| Solicitud → confirmación | Campos de la solicitud | Campos de la confirmación |
| --- | --- | --- |
| `ASSIGN` → `ASSIGNED` | `rows`, cantidad de filas de la partición, y `k` | Identificadores |
| `VECTORS` → `VECTORS_ACK` | `offset` desde cero dentro de la partición, y `points`, hasta 4096 pares `[PM10, PM2.5]` | `offset`, siguiente posición esperada |
| `LOAD_DONE` → `READY` | Identificadores | `rows`, filas cargadas |
| `ITER` → `PARTIAL` | `points`, los K centroides en orden | `rows` procesadas, `counts` y `sums` en orden de clúster |
| `FINISH` → `FINISHED` | Identificadores de la última asignación | Identificadores y, en un mensaje separado, `BYE` |

`offset` omitido equivale a cero. Las coordenadas son `float64`. El nodo valida que los lotes sean contiguos y que el total recibido coincida con `rows` antes de confirmar `READY`. El límite TCP de 1 MiB sigue vigente y K admite hasta 8192. El coordinador comprueba filas, dimensiones, sumas finitas y conservación de conteos en cada parcial. Una respuesta duplicada, fuera de orden o con otra clave provoca que falle el trabajo sin contabilizar ese parcial. También falla ante desconexión, cancelación o plazo vencido. Ante un fallo, el coordinador cancela el trabajo, cierra las conexiones participantes y espera a sus manejadores antes de devolver `failed` con el motivo. El resultado fallido omite centroides y conteos finales. Si el servicio sigue activo, acepta nuevas conexiones para otra ejecución. Cada fallo termina el trabajo en curso. El intento permanece en 1 y el coordinador no reasigna la partición. En un cierre exitoso, cada participante sale tras recibir `BYE_ACK`.

### Cálculo y resultado

El coordinador carga un CSV de dos columnas `PM10,PM2.5` por ejecución. Cuenta las filas en la carga y conserva sus valores y orden. Selecciona los nodos registrados en orden de identificador y reparte rangos contiguos de tamaño equilibrado. Inicializa los centroides una vez con la permutación pseudoaleatoria de la referencia y la semilla configurada. Cada nodo asigna solo su partición al centroide más cercano por distancia euclidiana al cuadrado, con desempate por el menor índice.

El coordinador espera todos los parciales de cada iteración y reduce por índice de partición. Promedia las coordenadas por clúster y conserva el centroide anterior si el clúster queda vacío. Termina cuando la raíz de la suma de desplazamientos al cuadrado es estrictamente menor que la tolerancia o cuando alcanza el máximo de actualizaciones. Una última asignación distribuida contra esos centroides devuelve los conteos finales; esa pasada no incrementa el número de actualizaciones.

Desde Go, se puede iniciar `plataforma.NewCoordinator(timeout, logger).Serve(ctx, listener)` en una goroutine y llamar a `Run(ctx, ExecutionConfig)` en la misma instancia. El contexto de `Run` puede imponer un plazo total. La instancia ejecuta un trabajo a la vez y admite nuevos registros tras cada trabajo.

El JSON de resultado contiene configuración, filas leídas, participantes, número de actualizaciones, estado, motivo, centroides iniciales, centroides y conteos finales. `partitions` identifica los nodos y los límites de cada rango `[start,end)` en el CSV. Los campos `timings` están en nanosegundos: carga CSV, espera de nodos, distribución hasta `READY`, cálculo con inicialización y asignación final, y tiempo total hasta el cierre de la ejecución. El tiempo total incluye el cierre de las sesiones. Véanse los [comandos y resultados de las pruebas](../results/pruebas/kmeans-distribuido/comandos.md).

## Validación

La equivalencia con la referencia secuencial se comprueba con uno y dos nodos por entrada. El control de centroides admite un error absoluto de 1e-8 µg/m³ más un error relativo de 1e-10 por coordenada. El umbral de convergencia de las comparaciones es 0.0001. Se exige igualdad exacta de inicialización, asignaciones independientes, conteos, iteraciones y motivo. La campaña usa interrupciones controladas en un solo equipo. Los pasos, diferencias y escenarios de cierre están en la [guía de validación](../results/pruebas/cluster-fallos/README.md).
