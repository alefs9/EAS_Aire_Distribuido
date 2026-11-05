# Mediciones limpias de PM10 y PM2.5

`PM25_PM10_limpio.parquet` contiene 1 062 897 mediciones individuales en 27 columnas. `auditoria_exclusiones.parquet` conserva 517 733 mediciones PM excluidas con su motivo. Ambos archivos son copias exactas de los resultados de la ejecución `20261105T003520_811897Z`.

`ultima_ejecucion.json` adapta las rutas a la raíz del repositorio y añade los hashes de los Parquet. El manifiesto original, con rutas de Google Colab, permanece en `results/limpieza/procedencia/manifest_original.json`.

## Obtener el CSV

Desde la raíz del repositorio, con Python y las dependencias instaladas:

```powershell
python -m pip install -r requirements.txt
python scripts/exportar_csv_limpio.py
```

El script genera `data/processed/PM25_PM10_limpio.csv`. El CSV queda excluido de Git debido a su tamaño; el Parquet completo sí se versiona. También está disponible el [CSV original de la ejecución](https://drive.google.com/file/d/122iB_Jax-Rmqn_l1Qk_Qtf96viNc0mDU/view).

La regeneración cambia el formato de almacenamiento, sin aplicar filtros adicionales ni construir vectores. No se garantiza que sus bytes u orden físico coincidan con el CSV original; contiene las mismas mediciones y columnas del Parquet. El script comprueba su hash y el número de filas antes de exportar.

## Reproducción en Colab

La preparación independiente se ejecuta en `notebooks/Limpieza.ipynb` desde el CSV original. Sus salidas se guardan en `DATA_DIR/Limpieza_PM/<RUN_ID>/` y actualizan `DATA_DIR/Limpieza_PM/ultima_ejecucion.json`.

Para continuar con el cuaderno de vectores en Drive, utilizar el manifiesto que genera la ejecución de limpieza allí. El manifiesto portátil del repositorio utiliza rutas relativas y no sustituye directamente las rutas montadas de Colab.

Los 109 288 registros con datum pendiente de revisión se conservan señalados. Los 1 025 663 finales desconocidos permanecen sin inferir. Cada fila sigue siendo una medición individual, no un par PM10-PM2.5.
