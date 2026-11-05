# Preparación de vectores de concentración

Los resultados corresponden a la ejecución `20261105T005132_156935Z` del [cuaderno de vectores](../../notebooks/Vectores.ipynb), utilizando la limpieza `20261105T003520_811897Z`.

## Construcción de observaciones

Cada observación conserva la clave de procedencia, punto, coordenadas UTM, zona, datum, inicio, periodo y final declarado. Se comprueba que exista como máximo una medición por contaminante en cada clave. Solo después se construyen las columnas PM10 y PM2.5; las claves que no tienen ambos contaminantes permanecen separadas, sin promediar ni imputar.

| Resultado | Cantidad |
| --- | ---: |
| Mediciones limpias de entrada | 1 062 897 |
| Mediciones utilizadas en pares | 927 288 |
| Mediciones sin el otro contaminante | 135 609 |
| Vectores completos | 463 644 |
| Vectores de una hora | 445 347 |
| Vectores de 24 horas | 18 297 |

El balance es `1 062 897 = 2 × 463 644 + 135 609`. El millón corresponde a mediciones limpias procesadas, no al número de vectores. La cobertura de los pares es 2019-01-25 a 2024-12-31. La ausencia de pares en 2018 resulta de la disponibilidad de ambos componentes en una misma clave, no de un filtro por antigüedad.

## Tablas completas

| Archivo | Filas de datos |
| --- | ---: |
| [`01_balance_vectores.csv`](tablas/01_balance_vectores.csv) | 5 |
| [`02_vectores_por_periodo.csv`](tablas/02_vectores_por_periodo.csv) | 2 |
| [`03_vectores_por_anio.csv`](tablas/03_vectores_por_anio.csv) | 12 |
| [`04_estadisticas_features.csv`](tablas/04_estadisticas_features.csv) | 4 |

Los cuatro CSV son los archivos completos exportados por el cuaderno. `04_estadisticas_features.csv` describe cada dimensión por periodo en sus unidades físicas. Las estadísticas no equivalen a parámetros de estandarización ni a una evaluación del modelo.

## Figuras

| Archivo | Interpretación |
| --- | --- |
| `figuras/01_balance_mediciones.png` | Destino de las mediciones individuales, distinguiendo las usadas en pares de las que no tienen contraparte. |
| `figuras/02_cobertura_vectores.png` | Disponibilidad anual de vectores, separada por periodo. |

Ambas figuras se generaron para la publicación a partir de los resultados verificados y no se presentan como capturas de Colab. El notebook conserva sus salidas ejecutadas originales.

## Representación y trazabilidad

El orden de las características es `[PM10, PM2.5]`, en µg/m³ y sin escalado. Los periodos se exportan por separado. `vector_id` permite seguir cada par en el Parquet completo; los archivos de características y sus metadatos se ordenan por ese identificador. No se replica información para aumentar la carga y no se entrena el modelo en esta etapa.

Persisten 54 644 vectores con datum pendiente de revisión. La marca se conserva en los metadatos. Las referencias desconocidas no deben utilizarse para inferir distancias geográficas sin validación.

## Comprobaciones de publicación

`verificacion_publicacion.json` documenta el contraste bidireccional de los pares y las observaciones incompletas con la agrupación del Parquet limpio. Se comprobaron el balance, la unicidad de vector_id, las concentraciones, la cobertura, el orden de las dos características y la correspondencia fila a fila de los CSV con el Parquet. El hash del manifiesto original de limpieza coincide con el registrado por el cuaderno.

Estas comprobaciones utilizan archivos ya procesados; no representan una nueva ejecución desde el CSV original ni resultados de K-Means o de Go. `resumen_vectores.json`, `RESUMEN_VECTORES.md` y `procedencia/manifest_original.json` conservan los resultados originales. `inventario.csv` registra filas, tamaños y SHA-256 de los archivos publicados.
