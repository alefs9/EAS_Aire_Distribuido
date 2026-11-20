# Contrato TCP del clúster · V1

El coordinador escucha y cada nodo inicia una conexión TCP independiente. La topología es en estrella, con un registro de nodos activos en memoria. El formato sigue el encuadre del informe (sección 3.2.11.1). Los mensajes de V1 pertenecen a la sesión TCP, por lo que sus campos identifican al nodo y todavía no incluyen una ejecución ni una iteración.

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
- El intervalo de latidos debe dejar margen dentro del plazo de lectura del coordinador. Para procesos en otros equipos, el coordinador debe escuchar en una interfaz accesible. V1 usa TCP sin TLS ni autenticación y está pensado para una red de confianza.
