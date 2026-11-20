# Persistencia en MongoDB

Responsable: Valentina Montenegro (A2).

`ConectarMongo` recibe URI y base. `Preparar` crea las colecciones e índices; `GuardarDataset` registra la entrada; `Crear`, `Iniciar`, `Completar` y `Fallar` controlan el trabajo. `Consultar` recupera el historial.

Estados: `queued → running → completed`, o `queued/running → failed`. Los resultados completados no se modifican. Una nueva ejecución utiliza otro identificador.

Configuración: `MONGODB_URI` y `MONGODB_DATABASE`. No se guardan contraseñas en Git. La API y el coordinador pueden importar `eas-aire/plataforma/internal/persistencia` y compartir el mismo job_id; su conexión se realiza en la integración del equipo.

Desde `go/plataforma`: `go test ./...`. La prueba de conexión se habilita con `PC4_INTEGRATION=1`. Las pruebas completas con los servicios se incorporan en A3.
