# Diseño del modelo ML distribuido

Este documento justifica el uso de K-Means sobre los vectores `[PM10, PM2.5]` y describe el diseño del cálculo distribuido. Separa lo **implementado** (preparación de vectores y referencia secuencial en Go) de lo **propuesto** (nodos, coordinador, API y bases de datos).

## 1. Justificación de K-Means y representación de entrada

### 1.1 Para qué sirve el agrupamiento en este proyecto

La pregunta del proyecto es cómo se distribuyen las concentraciones observadas de PM10 y PM2.5 entre los puntos y periodos disponibles y qué patrones pueden apoyar el monitoreo ambiental. Los registros de OEFA no traen una etiqueta que indique a qué "perfil de contaminación" pertenece cada observación, por lo que el problema es **no supervisado**. K-Means encaja porque agrupa observaciones con concentraciones semejantes y entrega, por cada grupo, un centroide interpretable: un par `[PM10, PM2.5]` en µg/m³ que resume el perfil típico del grupo.

El algoritmo también es adecuado para el objetivo del curso. Cada iteración se descompone en (a) un cálculo independiente por observación (asignarla al centroide más cercano) y (b) una reducción pequeña (sumas y conteos por clúster). Esa estructura permite repartir las observaciones en rangos, procesarlos en paralelo y combinar resultados parciales.

**Alcance del resultado.** Un clúster expresa similitud entre pares `[PM10, PM2.5]`. No es una categoría sanitaria, no define umbrales de alerta y no es una predicción. La predicción de concentraciones requiere un modelo y una validación temporal propios.

### 1.2 Observación y dimensiones

La observación del modelo es el vector `x = [PM10, PM2.5]`, en µg/m³, con `d = 2`. Proviene de la preparación de `Vectores.ipynb`: 463 644 pares completos, de los cuales 445 347 son de 1 hora y 18 297 de 24 horas. El archivo que consume el modelo contiene solo esas dos columnas (`features_1_hora.csv` y `features_24_horas.csv`). La ubicación, la fecha, la hora y el datum son metadatos: no forman parte del vector ni de la distancia.

En la implementación secuencial, los datos se cargan en `Dataset{Points [][]float64, N int, D int}` con `D = 2`. Los ceros numéricos son valores registrados y se conservan como observaciones válidas; el archivo de 1 hora contiene 16 437 vectores con ambas concentraciones en cero.

### 1.3 Distancia

Se usa la **distancia euclidiana al cuadrado**. Ambas dimensiones están en la misma unidad (µg/m³), por lo que la distancia tiene sentido físico. Además, la media aritmética es el punto que minimiza la suma de distancias euclidianas al cuadrado, de modo que el paso de actualización de centroides es coherente con la métrica. Se omite la raíz porque no cambia cuál centroide es el más cercano y ahorra cálculo en el bucle principal.

### 1.4 Periodo

El periodo del modelo distribuido es **1 hora** (445 347 vectores). Los vectores de 24 horas (18 297) corresponden a otra ventana de medición y **no se mezclan** con los horarios: tratarlos como equivalentes combinaría concentraciones obtenidas bajo intervalos distintos. La referencia secuencial se ejecutó además sobre 24 horas, de forma independiente, solo como comparación.

### 1.5 Escalado

La implementación actual **no escala** los datos. Así los centroides se leen directamente en µg/m³. El riesgo es que PM10, con mayor dispersión que PM2.5, pese más en la distancia; PM2.5 es además una fracción de PM10, por lo que ambas variables están relacionadas. No se ha evaluado todavía una estandarización (z-score). Si se adopta, debe calcularse una sola vez y aplicarse idéntica en la referencia, en la versión concurrente y en los nodos.

### 1.6 Parámetros de la referencia secuencial

| Parámetro | Valor | Observación |
| --- | --- | --- |
| Número de clústeres `K` | 4 | Valor fijado para la referencia. **No se ha seleccionado con un criterio** (por ejemplo, método del codo); queda pendiente. |
| Inicialización | `K` observaciones elegidas al azar sin reposición (`rng.Perm`) | Semilla fija `42`. No es k-means++. |
| Distancia | Euclidiana al cuadrado | Ver 1.3. |
| Escalado | Ninguno | Ver 1.5. |
| Tolerancia | `0.001` | Convergencia cuando `sqrt(Σ_c ‖μ_c' − μ_c‖²) < tol`, con la suma sobre los K centroides. |
| Máximo de iteraciones | `100` | Se registra si se alcanza. |
| Clúster vacío | Conserva su centroide anterior | Regla que debe repetir la versión distribuida. |
| Salida final | Reasignación con los centroides finales | Etiquetas e inercia corresponden a los centroides devueltos. |

### 1.7 Resultados de la referencia

Ejecución única en Google Colab (Go 1.22.2, Linux, Intel Xeon 2.20 GHz, 2 CPU lógicas), con `K = 4`, tolerancia `0.001`, máximo 100 iteraciones y semilla `42`.

| Periodo | Vectores | Iteraciones | Convergió | Inercia | Tiempo de carga | Tiempo de ejecución |
| --- | --- | --- | --- | --- | --- | --- |
| 1 hora | 445 347 | 96 | Sí | 3,898 × 10⁸ | 2,005 s | 2,726 s |
| 24 horas | 18 297 | 44 | Sí | 6,283 × 10⁶ | 0,224 s | 0,048 s |

Centroides y tamaño de cada clúster, periodo de 1 hora:

| Clúster | PM10 (µg/m³) | PM2.5 (µg/m³) | Vectores | % del total | PM2.5 / PM10 |
| --- | --- | --- | --- | --- | --- |
| 1 | 18,48 | 9,64 | 330 559 | 74,2 % | 0,52 |
| 3 | 80,52 | 34,10 | 85 807 | 19,3 % | 0,42 |
| 2 | 185,73 | 92,26 | 27 648 | 6,2 % | 0,50 |
| 0 | 541,69 | 117,25 | 1 333 | 0,3 % | 0,22 |

Centroides y tamaño de cada clúster, periodo de 24 horas:

| Clúster | PM10 (µg/m³) | PM2.5 (µg/m³) | Vectores | % del total |
| --- | --- | --- | --- | --- |
| 0 | 15,53 | 8,37 | 10 075 | 55,1 % |
| 2 | 46,40 | 19,04 | 4 848 | 26,5 % |
| 3 | 97,61 | 40,92 | 2 382 | 13,0 % |
| 1 | 172,25 | 93,18 | 992 | 5,4 % |

**Lectura.** En 1 hora, tres de cada cuatro vectores caen en el grupo de concentraciones bajas y un grupo de apenas 1 333 vectores concentra los valores más altos de PM10, con una relación PM2.5/PM10 menor que los demás. Esto describe la forma de los datos; no establece causas ni excedencias normativas.

### 1.8 Limitaciones de interpretación

- Los clústeres son similitudes de concentración, no categorías sanitarias ni alertas.
- K-Means usa la media, que es sensible a valores extremos. La limpieza conservó los extremos a propósito (máximo de 3 848,78 µg/m³ en PM10 horario), por lo que es esperable que un clúster pequeño agrupe episodios muy altos. Se documenta como hallazgo y no se corrige eliminando observaciones.
- `K = 4` no está justificado todavía con un criterio de selección.
- La inicialización es aleatoria y se probó una sola semilla: no se midió la estabilidad de los grupos frente a otras semillas.
- Cada resultado y tiempo proviene de una sola ejecución en Colab; no es una mediana de varias ejecuciones.
- Un par `[PM10, PM2.5]` no contiene el punto, la fecha ni la hora; esa interpretación requiere cruzar los metadatos.
