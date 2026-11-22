# Persistencia en MongoDB

Responsable: Valentina Montenegro (A2).

`ConectarMongo` recibe URI y base. `Preparar` crea las colecciones e índices; `GuardarDataset` registra la entrada; `Crear`, `Iniciar`, `Completar` y `Fallar` controlan el trabajo. `Consultar` recupera el historial.

Estados: `queued → running → completed`, o `queued/running → failed`. Los resultados completados no se modifican. Una nueva ejecución utiliza otro identificador.

La API configura el repositorio con `MONGODB_URI` y `MONGODB_DATABASE`. MongoDB conserva el `job_id` externo en `jobs._id`. La API guarda el ID interno devuelto por el coordinador como `cluster_job` en `resultado.origen`. Véase el [recorrido de un trabajo](api.md#recorrido-de-un-trabajo).

Las pruebas con `PC4_INTEGRATION=1` comprueban conexión, índices, estados y escritura atómica con MongoDB real. La [guía de persistencia](../../results/persistencia/README.md) contiene la reproducción con servicios activos y los documentos recuperados.
