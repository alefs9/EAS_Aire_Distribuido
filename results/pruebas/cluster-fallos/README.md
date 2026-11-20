# Validación del clúster distribuido

## Configuración y criterios

Las pruebas usan Docker Desktop en modo Linux. [entorno.log](entorno.log) registra las versiones observadas y el digest de la imagen de Go. La referencia es el ejecutable existente en `go/kmeans_secuencial`, en modo `sequential`. Su JSON ya incluye centroides iniciales, centroides finales, conteos y hash de asignaciones. Cada comparación usa el mismo CSV en su orden original, con coordenadas `[PM10, PM2.5]` y escala original. Se ejecuta el clúster con uno y dos nodos TCP en contenedores separados.

Antes de las comparaciones se fijaron `atol = 1e-8 µg/m³` y `rtol = 1e-10`. Se exige `|distribuido − referencia| ≤ atol + rtol × |referencia|` por coordenada y en orden de clúster. El término absoluto cubre redondeo cerca de cero. El relativo admite diferencias de suma en distinto orden, crecientes con la magnitud. Ambos son mucho menores que las concentraciones registradas y que el umbral de movimiento de `0.0001` usado para converger. Los centroides iniciales, los conteos, las asignaciones independientes por fila, la terminación y las iteraciones se comprueban por igualdad exacta.

## Comparaciones

En todos los casos se usaron semilla 42, tolerancia de convergencia 0.0001 y máximo de 100 actualizaciones. Se contrasta cada JSON del clúster con el de la referencia y con una asignación Python independiente a los centroides finales. La comprobación exige rangos contiguos `[inicio, fin)`, que las filas cargadas y la suma de conteos sean N y que las asignaciones calculadas desde ambos juegos de centroides coincidan por fila. El hash de la asignación independiente a los centroides de referencia también debe coincidir con el que exporta el ejecutable secuencial.

| Entrada | K | N | Terminación y actualizaciones de referencia | Observado con 1 nodo | Observado con 2 nodos | Mayor diferencia absoluta con 1 / 2 nodos (µg/m³) | Evaluación |
| --- | ---: | ---: | --- | --- | --- | --- | --- |
| Sintético de cuatro puntos | 2 | 4 | Convergencia, 3 | Convergencia, 3 | Convergencia, 3 | 0 / 0 | Pasó |
| `features_1_hora.csv` | 4 | 445 347 | Máximo de iteraciones, 100 | Máximo de iteraciones, 100 | Máximo de iteraciones, 100 | 0 / 7.95807864051e-13 | Pasó |
| `features_24_horas.csv` | 4 | 18 297 | Convergencia, 44 | Convergencia, 44 | Convergencia, 44 | 0 / 2.55795384874e-13 | Pasó |

En el sintético se obtuvieron los centroides conocidos `[1,1]` y `[11,11]`, con dos puntos cada uno. Todas las comparaciones tuvieron cero discrepancias de asignación por fila. Los hashes de entrada, los rangos, los conteos y las diferencias absolutas y relativas por coordenada están en [comparaciones.json](comparaciones.json). Los JSON individuales contienen la configuración completa. Los archivos `*.log` de cada caso identifican al coordinador y a cada nodo e indican su código de salida.

## Fallos controlados

Las pruebas usan un CSV sintético de cuatro filas, dos nodos, K=2, semilla 42, tolerancia 0, máximo de cuatro actualizaciones y plazo TCP del coordinador de 1 s. El nodo **a** inicia una sesión TCP de prueba que espera a recibir la trama pertinente antes de inyectar la condición. El nodo **b** ejecuta el cliente real. Cada escenario exige `failed`, un motivo comprobado, ausencia de centroides y conteos finales, salida de ambos participantes y cero registros participantes activos al retornar. Con el servicio activo se arranca después otra ejecución con dos sockets nuevos y los mismos IDs.

| Escenario | Inyección | Resultado esperado y observado |
| --- | --- | --- |
| Duplicado | Reenvío del parcial ya aceptado de la primera actualización | Pasó. Rechazado en la segunda solicitud, con una sola actualización contabilizada. |
| Iteración anterior | Parcial distinto con número de iteración anterior | Pasó. Rechazado, con una sola actualización contabilizada. |
| Iteración futura | Incremento del número antes de responder | Pasó. Rechazado sin actualizar. |
| Ejecución incorrecta | Sustitución del ID de ejecución | Pasó. Rechazado sin actualizar. |
| Partición incorrecta | Sustitución del ID de partición | Pasó. Rechazado sin actualizar. |
| Intento incorrecto | Cambio de 1 a 2 en la trama TCP | Pasó. Trama inválida sin actualizar. |
| Conteos incompletos | Parcial cuyo conteo suma menos que las filas asignadas | Pasó. Fallo de conservación, sin actualizar. |
| Dimensiones incompatibles | Un clúster en el parcial cuando K=2 | Pasó. Fallo de dimensiones, sin actualizar. |
| Desconexión en transferencia | Cierre tras recibir un lote y antes de confirmarlo | Pasó. Fallo de distribución. |
| Desconexión en cálculo | Cierre al recibir `ITER` | Pasó. Fallo de iteración. |
| Nodo silencioso | Socket abierto sin respuesta a `ITER` | Pasó. Venció el plazo de lectura. |
| Cancelación | Cancelación del contexto del trabajo tras recibir `ITER` | Pasó. Trabajo cancelado y sesiones cerradas. |
| Interrupción del coordinador | Cancelación del servicio con un `ITER` pendiente | Pasó. Falló el trabajo y terminó el servicio. |
| Manejador entre órdenes | Se detiene el servicio tras `ASSIGNED`, con el contexto del trabajo aún activo | Pasó tras la corrección. Terminaron el manejador y el servicio. |

La última prueba reprodujo un bloqueo antes del cambio: un manejador sin una orden pendiente solo esperaba al contexto del trabajo. Ahora también observa el contexto del servicio. Al fallar un trabajo, el coordinador espera a que terminen los manejadores participantes antes de liberar su capacidad para iniciar otro. [Registro detallado por escenario y participante](plataforma-fallos.log)

## Reproducción y capturas

Desde la raíz del repositorio, con Docker Desktop iniciado en modo Linux y Python 3 disponible mediante `py`:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\results\pruebas\cluster-fallos\ejecutar.ps1
```

El [script](ejecutar.ps1) contiene los comandos exactos. Ejecuta `go test -count=1 ./...`, `go test -race -count=1 ./...` y `go vet ./...` en `go/plataforma`. También ejecuta las pruebas de fallo en modo detallado y repite diez veces con `-race` las pruebas de fallos, cierre y ejecución TCP. Los logs registran los códigos de salida. Usa `-SoloPruebas` o `-SoloComparaciones` para reproducir solo una parte. El script escribe únicamente en esta carpeta de evidencias.

Para capturar la comparación y una selección compacta de los fallos en PowerShell:

```powershell
$e = '.\results\pruebas\cluster-fallos'
Get-Content "$e\comparaciones.log"
Select-String -Path "$e\plataforma-fallos.log" -Pattern '=== RUN|resultado:|nueva ejecucion:|--- PASS|^exit='
Get-Content "$e\plataforma-estabilidad.log"
```

Las interrupciones probadas son controladas en el nivel de servicio. La política termina el trabajo y requiere iniciar otra ejecución con nuevas sesiones. Las pruebas cubren corrección y fallos TCP en un solo equipo.
