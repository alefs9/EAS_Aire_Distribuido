# Resultados del análisis exploratorio

Los archivos proceden de la ejecución `20261104T235909_455000Z` del [cuaderno EDA](../../notebooks/EDA.ipynb). Los CSV fueron recuperados completos de la carpeta de resultados; no se reconstruyeron a partir de las primeras 50 filas visibles en Colab. Se conservan los nombres originales para permitir su correspondencia con las celdas.

## Lectura de resultados

- **Fuente completa:** `00`, `01` y `13` a `18` describen estructura, calidad, parámetros y cobertura del archivo original. Las estadísticas se separan por parámetro, unidad y periodo.
- **Diagnóstico de PM antes de transformar:** `02` a `05` y `19` a `21` examinan selección de contaminantes, unidades, valores no cuantificables y referencia espacial.
- **Subconjunto exploratorio:** `06` a `12` y `22` a `27` describen las 1 137 870 mediciones que cumplen las reglas exploratorias. Estos resultados no representan la limpieza definitiva. Los extremos IQR/P99 son señales de revisión, no exclusiones automáticas.

La fuente tiene 11 407 394 registros. El subconjunto seleccionado de PM contiene 1 580 630 mediciones; el subconjunto exploratorio contiene 645 346 mediciones de PM10 y 492 524 de PM2.5. Se conservan todos los años disponibles.

## Tablas completas

| Archivo | Filas de datos |
| --- | ---: |
| [`00_esquema_original.csv`](tablas/00_esquema_original.csv) | 23 |
| [`01_conteos_originales.csv`](tablas/01_conteos_originales.csv) | 1 |
| [`02_parametros_y_unidades.csv`](tablas/02_parametros_y_unidades.csv) | 14 |
| [`03_calidad_por_contaminante_periodo.csv`](tablas/03_calidad_por_contaminante_periodo.csv) | 4 |
| [`04_unidades_espaciales.csv`](tablas/04_unidades_espaciales.csv) | 10 |
| [`05_nombres_de_punto_ambiguos.csv`](tablas/05_nombres_de_punto_ambiguos.csv) | 19 |
| [`06_resumen_limpieza.csv`](tablas/06_resumen_limpieza.csv) | 10 |
| [`07_cobertura_anual.csv`](tablas/07_cobertura_anual.csv) | 26 |
| [`08_concentraciones.csv`](tablas/08_concentraciones.csv) | 2 |
| [`09_composicion_del_dataset.csv`](tablas/09_composicion_del_dataset.csv) | 3 |
| [`10_PM_por_contaminante.csv`](tablas/10_PM_por_contaminante.csv) | 4 |
| [`11_PM_por_anio_etapa.csv`](tablas/11_PM_por_anio_etapa.csv) | 26 |
| [`12_extremos_por_contaminante.csv`](tablas/12_extremos_por_contaminante.csv) | 80 |
| [`13_perfil_columnas_fuente.csv`](tablas/13_perfil_columnas_fuente.csv) | 23 |
| [`14_parametros_unidades_periodos_fuente.csv`](tablas/14_parametros_unidades_periodos_fuente.csv) | 324 |
| [`15_valores_no_cuantificables_fuente.csv`](tablas/15_valores_no_cuantificables_fuente.csv) | 48 |
| [`16_cobertura_anual_fuente.csv`](tablas/16_cobertura_anual_fuente.csv) | 7 |
| [`17_formatos_temporales.csv`](tablas/17_formatos_temporales.csv) | 7 |
| [`18_estadisticas_fuente_por_parametro.csv`](tablas/18_estadisticas_fuente_por_parametro.csv) | 295 |
| [`19_codigos_no_cuantificables_PM.csv`](tablas/19_codigos_no_cuantificables_PM.csv) | 22 |
| [`20_estadisticas_PM_entrada.csv`](tablas/20_estadisticas_PM_entrada.csv) | 8 |
| [`21_cobertura_PM_entrada.csv`](tablas/21_cobertura_PM_entrada.csv) | 26 |
| [`22_cobertura_mensual.csv`](tablas/22_cobertura_mensual.csv) | 309 |
| [`23_concentraciones_por_periodo.csv`](tablas/23_concentraciones_por_periodo.csv) | 4 |
| [`24_estadisticas_por_punto.csv`](tablas/24_estadisticas_por_punto.csv) | 223 |
| [`25_extremos_IQR.csv`](tablas/25_extremos_IQR.csv) | 4 |
| [`26_puntos_sobre_p99.csv`](tablas/26_puntos_sobre_p99.csv) | 85 |
| [`27_temporalidad_y_ubicacion_limpias.csv`](tablas/27_temporalidad_y_ubicacion_limpias.csv) | 4 |

## Figuras y evidencia

| Archivo | Contenido e interpretación |
| --- | --- |
| `figuras/01_cobertura_fuente.png` | Cobertura anual de la fuente completa. Las diferencias reflejan disponibilidad de registros. |
| `figuras/02_cobertura_PM.png` | Cobertura por año y contaminante en el subconjunto exploratorio. |
| `figuras/03_distribucion_periodos.png` | Mediana, cuartiles y rango mínimo a P99 por contaminante y periodo. |
| `figuras/04_perfil_columnas_captura.png` | Captura del perfil de las 23 columnas, obtenida de la salida del EDA incorporada al informe. |
| `figuras/05_valores_no_cuantificables_captura.png` | Captura de los valores más frecuentes. El listado completo de 48 categorías está en `15_valores_no_cuantificables_fuente.csv`. |

Los tres gráficos son los archivos originales exportados por el cuaderno. Las dos capturas complementarias muestran las salidas incorporadas por el equipo al informe y no sustituyen los CSV. Los códigos de la fuente completa no deben atribuirse automáticamente al subconjunto PM.

## Procedencia y límites

`ejecucion.json` y `resumen_exploratorio.json` conservan los datos de ejecución originales. `RESUMEN_EXPLORATORIO.md` es la síntesis generada en esa ejecución. Sus referencias a limpieza corresponden a la transformación exploratoria del propio EDA.

El nombre de un punto no garantiza una estación física única. La ubicación, el datum, el periodo y la información temporal deben preservarse para las etapas posteriores. La existencia de valores cero no demuestra ausencia de contaminación; los valores extremos no prueban errores ni excedencias sanitarias por sí solos.

`inventario.csv` permite comprobar las filas, tamaños y SHA-256 de los archivos conservados. El inventario se generó al preparar la publicación y no constituye una nueva ejecución del análisis.
