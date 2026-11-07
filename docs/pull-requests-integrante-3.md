# Tres pull requests preparados para revisión

Destino previsto: `alefs9/EAS_Aire_Distribuido`, rama `develop`. Responsable actual del aporte: Valentina. No publicados. Los textos siguientes pueden usarse como cuerpo de cada PR; no incluyen resultados inexistentes.

## N1 — Implementar K-Means concurrente con workers configurables

Rama: `feature/kmeans-concurrente`.

La referencia solo disponía de ejecución secuencial. Se agrega `RunKMeansConcurrent` en el mismo módulo y con el mismo contrato, inicialización y política de clúster vacío. Se distribuyen rangos contiguos entre workers configurables, se acumulan sumas y conteos privados y se actualizan los centroides después de una barrera `WaitGroup`, con reducción en orden fijo.

Se comprobaron cobertura única de particiones, bloques desiguales, un worker, varios, más workers que observaciones, empates, vacíos, entradas inválidas, no mutación y reproducibilidad mediante `go test -count=1 -timeout=60s ./...` en `go/kmeans_secuencial`. La documentación explica límites numéricos, memoria y terminación. La integración y el rendimiento se presentan en N2 y N3.

## N2 — Integrar carga y modelos y verificar resultados

Rama: `test/integracion-kmeans`. Depende de N1.

Se añade selección explícita de periodo y modo a un comando común, salida JSON con inicialización, resultados, conteos, hash de etiquetas y tiempos separados. Se conserva el modo histórico sin argumentos. Un script coordina los módulos existentes y comprueba los conteos del cargador completo; el emparejamiento de mediciones sigue siendo el cuaderno de vectores publicado.

Validación: pruebas normales y `go test -race` en ambos módulos, incluidos 445 347 vectores horarios y 18 297 diarios, todos en PASS. Se compararon directamente las asignaciones con 1, 2, 4 y 8 workers, centroides e inercia con tolerancia 1e-9 absoluta más 1e-9 relativa, y coincidencia de iteraciones. Ambos cargadores aceptaron 1 062 897 mediciones y rechazaron cero. Se conservan logs completos, procedencia, commit y fallos de preparación corregidos. El detector respalda solo las rutas ensayadas.

## N3 — Evaluar rendimiento secuencial y concurrente de K-Means

Rama: `perf/comparacion-kmeans`. Depende de N2.

Se incorpora una campaña reproducible con cinco repeticiones medidas y un calentamiento por configuración, para ambos periodos y 1/2/4/8 workers. Se mantienen datos, inicialización, K=4, semilla=42, tolerancia=0.001 y GOMAXPROCS=4. Se conservan 50 mediciones, 10 calentamientos, entorno, hashes, resultados, tablas y figuras.

Con cuatro workers, la mediana de entrenamiento horario cambia de 0,811010 s a 0,301061 s (2,694×); en diario el speedup es 1,919×. La variante diaria con un worker es más lenta que la referencia (0,900×). Se separan carga, entrenamiento y total, y se advierte que la variante concurrente fusiona recorridos: la comparación no aísla causalmente el paralelismo. Se incluye el texto para 3.3–3.5 y aportes a conclusiones y recomendaciones.

Validación adicional: pruebas y `go vet ./...` en ambos módulos desde un clon limpio. Las 60 ejecuciones de la campaña verificaron inicialización, asignaciones, conteos, convergencia, iteraciones y equivalencia numérica antes de generar el resumen.
