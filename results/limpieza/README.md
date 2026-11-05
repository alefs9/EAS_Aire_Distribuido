# Resultados de la limpieza definitiva

Los archivos proceden de la ejecución `20261105T003520_811897Z` de [Limpieza.ipynb](../../notebooks/Limpieza.ipynb). Se procesó el archivo original, con los criterios justificados por el EDA y sin recorte por antigüedad.

## Resultado y balance

| Concepto | Mediciones |
| --- | ---: |
| Fuente completa | 11 407 394 |
| Subconjunto PM10 y PM2.5 | 1 580 630 |
| Exclusiones dentro de PM | 517 733 |
| Resultado limpio | 1 062 897 |
| PM10 conservado | 598 502 |
| PM2.5 conservado | 464 395 |

El balance es `1 580 630 = 517 733 + 1 062 897`. La selección de otros parámetros fuera del tema no se contabiliza como error de datos. Cada medición PM excluida recibe el primer motivo incumplido para evitar doble conteo. No se eliminaron duplicados ni claves ambiguas en esta ejecución.

## Criterios aplicados

Se conservaron concentraciones numéricas finitas no negativas, unidades compatibles, procedencia y punto informados, inicio interpretable, periodo reconocido, zona UTM 17, 18 o 19 y coordenadas dentro del control estructural. Los finales completos anteriores al inicio se excluyeron. Los finales ausentes permanecen desconocidos. Se normalizaron variantes de periodo y datum; las referencias no reconocidas permanecen señaladas.

Los extremos y los ceros se conservan como resultados observados. No se imputan valores para completar pares PM10-PM2.5. La preparación de vectores corresponde a otra etapa.

## Tablas originales

| Archivo | Filas de datos |
| --- | ---: |
| [`01_entrada.csv`](tablas/01_entrada.csv) | 2 |
| [`02_exclusiones_por_motivo.csv`](tablas/02_exclusiones_por_motivo.csv) | 9 |
| [`03_balance_limpieza.csv`](tablas/03_balance_limpieza.csv) | 15 |
| [`04_resumen_claves_ambiguas.csv`](tablas/04_resumen_claves_ambiguas.csv) | 0 |
| [`05_composicion_final.csv`](tablas/05_composicion_final.csv) | 4 |
| [`06_cobertura_anual.csv`](tablas/06_cobertura_anual.csv) | 26 |
| [`07_estadisticas_concentracion.csv`](tablas/07_estadisticas_concentracion.csv) | 4 |
| [`08_referencia_espacial.csv`](tablas/08_referencia_espacial.csv) | 4 |
| [`09_disponibilidad_por_clave.csv`](tablas/09_disponibilidad_por_clave.csv) | 2 |

`04_resumen_claves_ambiguas.csv` tiene solo encabezado: no se encontraron claves ambiguas. Los nueve CSV son los archivos completos exportados, no recortes de las primeras filas del notebook.

## Evidencias

| Archivo | Procedencia |
| --- | --- |
| `figuras/01_composicion_final_captura.png` | Captura aportada por el equipo de la salida `05_composicion_final`. |
| `figuras/02_exclusiones_captura.png` | Captura de la salida `02_exclusiones_por_motivo` incorporada al informe. |
| `figuras/03_exclusiones_por_motivo.png` | Gráfico generado desde el CSV original de exclusiones, sumando los dos contaminantes por motivo. |

El gráfico complementario se preparó para la publicación y no se presenta como una captura del notebook. Las cifras de ambas capturas coinciden con las tablas y los Parquet.

## Procedencia y comprobaciones

`ejecucion.json`, `resumen_limpieza.json` y `RESUMEN_LIMPIEZA.md` conservan la información de la ejecución. `procedencia/manifest_original.json` conserva las rutas originales de Drive. `data/processed/ultima_ejecucion.json` adapta esas rutas para el repositorio.

`verificacion_publicacion.json` registra las comprobaciones realizadas sobre los Parquet descargados: filas, balance, distribución por contaminante, cumplimiento de criterios comprobados, ausencia de claves ambiguas y concordancia con las tablas. Estas comprobaciones no implican volver a limpiar los 11 407 394 registros originales. `inventario.csv` documenta tamaños, filas y SHA-256 de los archivos del resultado y los Parquet.

Persisten 109 288 mediciones con datum pendiente de revisión y 1 025 663 sin final interpretable. No se inventó información para completarlas. El dataset conserva 27 columnas y cubre 2018-01-01 a 2024-12-31.
