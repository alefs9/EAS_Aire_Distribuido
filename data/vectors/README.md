# Vectores PM10 y PM2.5

Cada vector contiene dos concentraciones en el orden fijo `[PM10, PM2.5]`, expresadas en µg/m³. Se conserva la identidad de la observación, con procedencia, ubicación, inicio, periodo y final declarado. No se aplicó escalado, imputación, réplica ni entrenamiento.

| Archivo incluido | Filas | Contenido |
| --- | ---: | --- |
| `vectores_PM10_PM25.parquet` | 463 644 | Pares completos, vector_id y metadatos |
| `observaciones_incompletas.parquet` | 135 609 | Claves con un contaminante ausente |
| `features_1_hora.csv` | 445 347 | PM10 y PM2.5, en ese orden |
| `features_24_horas.csv` | 18 297 | PM10 y PM2.5, en ese orden |

Los dos CSV de características tienen encabezado y solo dos columnas. La fila de cada archivo corresponde al registro de sus metadatos del mismo periodo, ordenado por `vector_id`. No se deben mezclar lecturas de una hora y de 24 horas como si fueran intervalos equivalentes. El escalado que se elija después deberá documentarse y aplicarse de forma consistente entre implementaciones.

## CSV completo y metadatos

El Parquet completo contiene los metadatos. Para obtener sus CSV desde la raíz del repositorio:

```powershell
python -m pip install -r requirements.txt
python scripts/exportar_csv_vectores.py
```

El script genera `vectores_PM10_PM25.csv`, `metadatos_1_hora.csv` y `metadatos_24_horas.csv`, manteniendo el orden por `vector_id`. Estos CSV quedan fuera de Git. Los archivos originales también están en [la carpeta de esta ejecución](https://drive.google.com/drive/folders/10zHBdB7wt5R1vKcbou0rLVMoyNsiLkvW).

Los metadatos contienen textos con comas entre comillas. Debe utilizarse un lector CSV que respete el entrecomillado. En DuckDB, especificar `delim=','`, `quote='"'` y `escape='"'` evita que una muestra inicial sin campos entrecomillados produzca una detección incorrecta.

## Manifiesto y procedencia

`manifest.json` utiliza rutas relativas a la raíz del repositorio y añade hashes de los archivos publicados. El manifiesto original está en `results/vectores/procedencia/manifest_original.json`. El hash `origen_manifest_sha256` corresponde al manifiesto original de limpieza, conservado en `results/limpieza/procedencia/manifest_original.json`; no al manifiesto portátil adaptado.

El cuaderno original de Colab sigue leyendo `DATA_DIR/Limpieza_PM/ultima_ejecucion.json` desde Drive. El manifiesto portátil no sustituye directamente las rutas montadas de Colab. Para reproducir el flujo en Drive, ejecutar Limpieza.ipynb y después Vectores.ipynb con el mismo DATA_DIR.

El volumen de entrada es 1 062 897 mediciones limpias. El número de vectores completos es 463 644 y se informa separadamente. Los 54 644 vectores con datum pendiente de revisión conservan esa marca; no deben utilizarse para inferir distancias geográficas sin validar su referencia espacial.
