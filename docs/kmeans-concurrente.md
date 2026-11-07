# N1 — K-Means concurrente (Valentina, integrante 3)

El intercambio de responsabilidades asigna a Valentina los entregables N1, N2 y N3 que el plan rotula Nicole. Se conserva el módulo `go/kmeans_secuencial` y sus tipos `Dataset` y `KMeansResult`; no se duplica el módulo ni la referencia.

## Implementación

`RunKMeansConcurrent` recibe los mismos parámetros de `RunKMeans`, más `workers > 0`. Divide N filas contiguas entre P=min(workers,N) goroutines. Las primeras N mod P reciben una fila adicional. Los rangos semiabiertos cubren [0,N) exactamente una vez; nunca se lanzan bloques vacíos. Para 445 347 filas y cuatro workers: 111 337, 111 337, 111 337 y 111 336 filas.

Se implementa partición estática por iteración, no un Worker Pool persistente. Cada goroutine consulta datos y centroides de solo lectura, escribe únicamente su segmento de etiquetas y calcula K conteos y K×2 sumas privadas. `sync.WaitGroup` establece la barrera. Después de `Wait`, el coordinador reduce en orden creciente de worker, actualiza las medias y evalúa la convergencia. No se usan canales ni mutex en este algoritmo; el cargador de mediciones sí tiene su propio pipeline de canales.

Se reutiliza `initializeCentroids`: muestreo sin reemplazo con la misma semilla; distancia euclidiana al cuadrado, empate a favor del primer centroide, clúster vacío conserva su centro anterior y convergencia estricta `sqrt(sum(distancia²)) < tol`. Con tolerancia cero se agota el máximo de iteraciones, igual que la referencia. La asignación final usa los centroides definitivos y la inercia se suma en el orden original de observaciones.

## Memoria, terminación y límites

El entrenamiento tiene costo aproximado O(I×N×K×2), más reducción O(I×P×K×2); memoria adicional O(N+P×K×2). Las goroutines se crean por iteración y todas terminan antes de la siguiente; no hay envíos ni recepciones que puedan bloquear un canal. Los errores de entrada se detectan antes de lanzarlas. Se asume que el llamador no modifica el dataset simultáneamente. Se mantienen los límites numéricos de la referencia: no se promete estabilidad para magnitudes arbitrarias que desborden float64.

La reducción fija hace reproducible una configuración, pero cambiar el número de workers cambia la agrupación de las sumas. Se comparan valores con tolerancia absoluta 1e-9 más relativa 1e-9, etiquetas exactamente iguales e iteraciones/convergencia iguales para los casos ensayados. Esto no garantiza iguales trayectorias para cualquier dataset cercano a una frontera de asignación. El orden de inicialización fija la identidad de los clústeres, por lo que no se permite ocultar diferencias mediante permutación en estos ensayos.

La creación de goroutines, asignaciones de memoria, barreras y reducción tienen costos; su contribución individual no se ha perfilado. Concurrencia en un proceso no demuestra distribución entre computadores. GOMAXPROCS limita la ejecución paralela de Go; workers no equivale a núcleos.

## Pruebas

Desde `go/kmeans_secuencial`: `go test -count=1 -timeout=60s ./...`.
Las pruebas agregadas recorren 1 240 combinaciones de N y workers para cobertura de particiones, bloques desiguales, un worker, varios, más workers que puntos, datos idénticos, empates, clúster vacío, tolerancia cero, entradas inválidas, reproducibilidad y no mutación. El timeout del ejecutor acota el fallo si la implementación no termina. Los registros finales se consolidan en N2.
