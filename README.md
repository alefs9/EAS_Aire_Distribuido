# EAS Aire Distribuido

Análisis de mediciones ambientales de PM10 y PM2.5 para estudiar patrones espaciales y temporales de calidad del aire. Esta versión incorpora el análisis exploratorio de la fuente, sus diagnósticos de calidad y sus evidencias reproducibles.

## Fuente y cobertura

El archivo `EAS_Aire_0.csv` corresponde al componente aire del conjunto [Evaluaciones ambientales de seguimiento (EAS)](https://www.datosabiertos.gob.pe/dataset/evaluaciones-ambientales-de-seguimiento-eas), publicado por el Organismo de Evaluación y Fiscalización Ambiental (OEFA).

| Característica | Resultado de la ejecución |
| --- | ---: |
| Registros originales | 11 407 394 |
| Variables originales | 23 |
| Fechas de inicio | 2018-01-01 a 2024-12-31 |
| Mediciones seleccionadas de PM10 y PM2.5 | 1 580 630 |
| Mediciones del subconjunto exploratorio | 1 137 870 |

El total original incluye otros parámetros y variables meteorológicas. El subconjunto exploratorio se utiliza para describir concentraciones; no equivale al resultado de la limpieza definitiva ni a una cantidad de vectores del modelo. No se replicaron filas para aumentar el volumen.

## Contenido

| Ruta | Contenido |
| --- | --- |
| `notebooks/EDA.ipynb` | Cuaderno ejecutado, código, explicaciones y salidas |
| `results/eda/tablas/` | 28 CSV completos exportados por la ejecución |
| `results/eda/figuras/` | Tres gráficos y dos capturas de las salidas |
| `results/eda/README.md` | Guía e interpretación de resultados |
| `results/eda/ejecucion.json` | Entorno y procedencia de la ejecución |
| `results/eda/resumen_exploratorio.json` | Conteos y cobertura del subconjunto exploratorio |
| `results/eda/RESUMEN_EXPLORATORIO.md` | Síntesis generada durante la ejecución |
| `results/eda/inventario.csv` | Tamaños, filas y hashes de los archivos publicados |
| `data/raw/README.md` | Fuente y acceso al archivo original |
| `requirements.txt` | Dependencias de análisis |

## Reproducción en Google Colab

1. Abrir `notebooks/EDA.ipynb` en Google Colab. También está disponible el [cuaderno de origen](https://colab.research.google.com/drive/1ILtge8uqvyCZ4Isuqcm-kiV079eUNEQh).
2. Tener el archivo original en Google Drive. Consultar [las instrucciones de entrada](data/raw/README.md).
3. Ejecutar las celdas en orden. La celda de configuración utiliza `DATA_DIR = Path('/content/drive/MyDrive/Concurrente')`; ajustar únicamente esa ruta si la carpeta tiene otra ubicación.
4. Autorizar el montaje de Drive en la propia sesión. El cuaderno instala DuckDB; las demás bibliotecas se incluyen en Colab. La ejecución registrada empleó DuckDB 1.3.2 y Python 3.13.18. Para igualar DuckDB, cambiar la celda de instalación a `%pip -q install duckdb==1.3.2` antes de una nueva ejecución.
5. El cuaderno comprueba al menos 15 GiB libres en `/content`, limita DuckDB a 3 GB de memoria y utiliza hasta cuatro hilos. La entrada tiene 5 153 626 505 bytes; la lectura y las consultas completas requieren tiempo y almacenamiento temporal.
6. Las nuevas salidas se guardan en `DATA_DIR/Procesado_PM/<RUN_ID>/`. Para actualizar los resultados versionados, copiar los CSV y las figuras correspondientes a una misma ejecución a `results/eda/` y actualizar su procedencia.

El cuaderno está adaptado a Colab; no se ejecuta directamente como un script de Python local sin adaptar el montaje y las rutas. Las salidas incluidas pertenecen a la ejecución conservada; preparar este repositorio no implicó volver a procesar el archivo masivo.

## Alcance de la interpretación

El EDA examina tipos de lectura, parámetros, unidades, periodos, cobertura temporal, información desconocida, identidades espaciales y distribuciones. Los años con más registros indican mayor disponibilidad en la fuente, no necesariamente mayor contaminación. Los valores extremos se señalan para revisión y no se eliminan automáticamente. Las coordenadas UTM no se interpretan como latitud y longitud; las variantes de zona y datum requieren validación.

El notebook contiene una transformación exploratoria y conserva nombres históricos como `06_resumen_limpieza.csv` y `mediciones_PM_limpias`. En este directorio esos nombres describen únicamente el subconjunto exploratorio de 1 137 870 mediciones. El dataset de limpieza definitiva y la preparación de vectores se incorporan en sus etapas correspondientes.

Los CSV masivos y las bases temporales no forman parte del repositorio. No se incluyen implementaciones ni resultados de rendimiento de Go en esta versión.
