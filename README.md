# EAS Aire: analisis de PM10 y PM2.5

Analisis exploratorio de mediciones ambientales de OEFA, limpieza y preparacion de pares reales de PM10 y PM2.5.

## Estructura

- `EDA.ipynb`: revision de calidad, cobertura, limpieza y exportacion de mediciones.
- `Vectores.ipynb`: construccion de pares PM10/PM2.5 a partir de la salida de EDA.
- `Procesado_PM/20261104T214733_533275Z/`: cinco tablas de evidencia conservadas (01, 07, 08, 09 y 11), `README_datos.md` y `resumen_exploratorio.json`.
- `Vectores_PM/20261104T215328_705038Z/resumen_vectores_publico.json`: resumen de vectores sin rutas locales.

Esta seleccion no incluye todas las tablas generadas por EDA. Los notebooks permiten regenerar los resultados completos.

## Ejecucion

Los notebooks estan preparados para Google Colab y montan Google Drive. Ajustar `DATA_DIR` en ambos a la misma carpeta de trabajo. EDA espera el CSV original `EAS_Aire_0.csv`, que debe obtenerse de OEFA y mantenerse fuera de Git. Registrar la URL oficial y fecha de descarga cuando se obtenga la fuente.

Ejecutar primero EDA de principio a fin y luego Vectores. El entorno necesita pandas, DuckDB y soporte de lectura Parquet; EDA incluye la instalacion de DuckDB y comprueba al menos 15 GiB libres para su trabajo temporal.

EDA genera los datos limpios y los archivos locales `manifest.json`, `ejecucion.json` y `Procesado_PM/ultima_ejecucion.json`. Vectores utiliza este ultimo para localizar el Parquet y verificar el numero de filas. Esos archivos se conservan localmente y Git los ignora; el resumen publico no sustituye al manifiesto.

## Archivos publicables

Se conservan los notebooks, la documentacion, las tablas pequenas numeradas del EDA y los resumenes sin rutas privadas. Los datasets CSV completos, Parquet, bases DuckDB, manifiestos con rutas de Drive, checkpoints y temporales quedan excluidos mediante `.gitignore`.

Se anonimizaron las rutas impresas de Drive y se retiraron metadatos de identidad de Colab; el codigo y los resultados cientificos de los notebooks se conservaron. Revisar salidas y metadatos antes de publicar nuevas ejecuciones, porque `.gitignore` no limpia el contenido de un notebook.

## Flujo de trabajo

Crear cada rama de trabajo desde `develop`, realizar alli los commits y abrir un PR hacia `develop`. Integrar en `develop` los cambios revisados y probados. Publicar versiones mediante PR de `develop` hacia `main`.
