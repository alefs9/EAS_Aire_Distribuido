# Indicadores ambientales · Punto 5.3.3

## Procesamiento y trazabilidad

Se utiliza el Parquet de pares completos del proyecto, verificado por SHA-256
contra el manifiesto, con vector_id único y conteos comprobados por periodo.
La fuente es OEFA EAS. Los valores originales se conservan en µg/m³, sin escalar,
imputar, replicar ni eliminar extremos por su magnitud.

Cada indicador separa periodo y contaminante. El denominador es el número de
observaciones completas de ese grupo, no personas. Cada vector aporta una lectura
de PM10 y otra de PM2.5; sumar ambos contaminantes no duplica la cantidad de pares.
Los percentiles usan interpolación continua de DuckDB (`quantile_cont`); la media
pondera cada observación por igual, no cada estación o año por igual.

Se procesaron 463644 pares: 445347 horarios y 18297 de 24 horas. No hubo exclusiones
adicionales en esta ejecución. Las 135609 mediciones sin pareja fueron excluidas
en la preparación previa y no están representadas por estos indicadores.
El código registra por separado periodos no admitidos, concentraciones ausentes,
no finitas y negativas, con una causa prioritaria por fila. Los ceros se conservan
y se cuentan. Una fecha ausente solo se excluye del resumen anual.

Las salidas son:

- `ejecucion/por_periodo.csv`: cuatro filas con volumen, mínimo, cuartiles, mediana,
  media, p95, máximo, ceros y cobertura de fechas.
- `ejecucion/por_anio.csv`: los mismos estadísticos por año; años sin observaciones
  no se rellenan con ceros. Las diferencias anuales no son tendencias de una red fija.
- `ejecucion/por_punto.csv`: 426 filas de indicador, separando contaminante y periodo.
  La clave incluye informe, evaluación, punto, coordenadas, zona y datum. No se
  fusionan nombres iguales de distintos contextos ni se asignan ubicaciones nuevas.
  Estas filas no equivalen a 426 estaciones únicas.
- `ejecucion/procedencia.json`: identidad de entrada, reglas de denominador,
  versión de DuckDB y hashes de scripts y CSV de la ejecución.
- `ejecucion/distribucion_periodos.png`: mediana e intervalo entre p25 y p75.
  Las barras no son intervalos de confianza ni muestran máximos o límites sanitarios.

Se conservan 54644 pares con datum no verificado. Pueden describirse por su clave
de procedencia, pero no se usan para calcular distancias ni atribuir distritos.
El informe y la evaluación forman parte de la identidad: su cambio puede separar
registros del mismo lugar físico. No se afirma una red de estaciones consolidada.

## Hallazgos e interpretación para el informe

| Periodo | Contaminante | Observaciones | Mediana | Media | p95 |
| --- | --- | ---: | ---: | ---: | ---: |
| Una hora | PM10 | 445347 | 23.45 | 42.38 | 149.47 |
| Una hora | PM2.5 | 445347 | 11.31 | 19.81 | 76.81 |
| 24 horas | PM10 | 18297 | 27.10 | 42.89 | 136.26 |
| 24 horas | PM2.5 | 18297 | 12.60 | 20.03 | 76.50 |

Todas las concentraciones de la tabla están en µg/m³; el redondeo es solo de presentación.
En ambos periodos, las medias superan las medianas. Esto muestra la influencia de
valores elevados en la distribución y justifica consultar cuartiles y percentiles,
además del promedio. Los periodos se describen por separado: no son muestras
equivalentes para atribuir diferencias al intervalo de medición.

La cobertura de los pares horarios va del 24 de abril de 2019 al 31 de diciembre
de 2024; la de 24 horas, del 25 de enero de 2019 al 31 de diciembre de 2024.
El Parquet de pares no contiene observaciones de 2018, aunque la fuente original
tenía cobertura más amplia. No se extienden las conclusiones a registros excluidos.

Los 4344 pares horarios de 2019 y los 3168 de 2020 contienen ceros para ambos
contaminantes. Se conservan como valores registrados y deben revisarse contra
la procedencia; no prueban ausencia de contaminación. En todo el conjunto hay
16438 ceros horarios y 617 de 24 horas para cada contaminante.

El resumen por punto permite revisar perfiles junto con su volumen y cobertura.
Por ejemplo, CA-01 del informe 00250-2019-OEFA/DEAM-STEC, evaluación de Gran
Trapecio, presenta cinco observaciones de 24 horas entre el 7 y el 16 de mayo de
2019: medianas de 26.0 para PM10 y 11.0 para PM2.5. Ese resultado describe esa
campaña concreta; no se extrapola al año ni se compara como una estación de cobertura continua.

Estos indicadores apoyan el monitoreo al permitir identificar distribuciones,
revisar valores elevados y detectar limitaciones de cobertura antes de comparar
puntos o años. No estiman población expuesta, efectos causales ni reducción de
enfermedades. Tampoco establecen alertas o categorías sanitarias.

Los indicadores describen concentraciones observadas; los centroides de las
pantallas describen agrupamientos de K-Means. Esta rama no implementa pronóstico
temporal. El objetivo predictivo del proyecto queda pendiente de variable objetivo,
horizonte, separación temporal, modelo de referencia y métricas de error verificables.

## Reproducción y validación

Desde la raíz del repositorio, con las dependencias de `requirements.txt`:

```powershell
python scripts/indicadores/test_indicadores.py
python scripts/indicadores/ejecutar.py --salida results/indicadores/nueva_ejecucion
python scripts/indicadores/graficar.py --carpeta results/indicadores/nueva_ejecucion
```

Usar una carpeta nueva: el ejecutor rechaza sobrescribir evidencias. DuckDB 1.3.2
y un hilo mantienen estable la agregación. Los hashes de scripts normalizan a LF.
El gráfico publicado se generó con Matplotlib 3.8.4; la representación puede variar
entre versiones sin cambiar los indicadores.

Pasaron seis pruebas con valores conocidos: estadísticos y periodos, exclusiones,
ceros, fechas desconocidas, identidad contextual de puntos y protección de salidas.
Una segunda ejecución produjo hashes idénticos en los tres CSV. Los hashes de
salida corresponden a los bytes exportados; un checkout que convierta LF a CRLF
puede cambiarlos sin alterar los datos. Para verificar bytes, comparar ejecuciones
recién generadas con el manifiesto de su propia ejecución.

Este paquete aporta cuadros, gráfico e interpretación para incorporar al informe.
No modifica las pantallas ni publica nuevas consultas en la API; esos contratos
de integración deben acordarse antes de enlazar filtros espaciales o temporales.
