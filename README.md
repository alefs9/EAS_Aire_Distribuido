# EAS Aire Distribuido

Análisis de PM10 y PM2.5 del OEFA. Incluye EDA, limpieza, preparación de vectores, carga y validación en Go, K-Means secuencial, concurrente y distribuido, API HTTP, coordinador TCP, nodos ML, persistencia en MongoDB, caché Redis e interfaz web servida por Go.

## Datos y estructura

Fuente: componente aire de [Evaluaciones ambientales de seguimiento (EAS)](https://www.datosabiertos.gob.pe/dataset/evaluaciones-ambientales-de-seguimiento-eas), publicado por OEFA.

| Etapa | Resultado |
| --- | ---: |
| Archivo original | 11 407 394 registros, 23 variables, años 2018-2024 |
| PM10 y PM2.5 antes de limpiar | 1 580 630 mediciones |
| Subconjunto exploratorio del EDA | 1 137 870 mediciones |
| Limpieza definitiva | 1 062 897 mediciones y 517 733 exclusiones |
| Vectores horarios | 445 347 |
| Vectores de 24 horas | 18 297 |
| Mediciones sin pareja | 135 609 |

Las mediciones y los vectores son unidades distintas: `1 062 897 = 2 × 463 644 + 135 609`. No se replican filas ni se imputan concentraciones. Los periodos se procesan por separado.

| Ruta | Contenido |
| --- | --- |
| [notebooks](notebooks/) | EDA, limpieza y preparación de vectores, con salidas |
| [data/raw](data/raw/) | Fuente y acceso al CSV original |
| [data/processed](data/processed/) | Parquet limpio, exclusiones y manifiesto |
| [data/vectors](data/vectors/) | Vectores y CSV por periodo |
| [go/cargar-datos](go/cargar-datos/) | Carga y validación secuencial y concurrente |
| [go/kmeans_secuencial](go/kmeans_secuencial/) | Ambas versiones de K-Means, notebooks y pruebas |
| [results/eda](results/eda/), [results/limpieza](results/limpieza/), [results/vectores](results/vectores/) | Tablas, figuras e inventarios |
| [results/pruebas](results/pruebas/) | Registros de carga completa y pruebas |
| [results/rendimiento/ejecucion-local](results/rendimiento/ejecucion-local/) | Entorno y mediciones de Windows |

## Requisitos y preparación

Se requiere **Go 1.27.1 o posterior** para ejecutar los módulos del repositorio. K-Means declara Go 1.22 como mínimo, pero el cargador requiere 1.27.1. Las pruebas con `-race` necesitan una plataforma y un compilador C compatibles. Las herramientas Python se instalan desde la raíz:

```powershell
python -m pip install -r requirements.txt
```

En Colab, ejecutar [EDA.ipynb](notebooks/EDA.ipynb), [Limpieza.ipynb](notebooks/Limpieza.ipynb) y [Vectores.ipynb](notebooks/Vectores.ipynb), en ese orden, ajustando `DATA_DIR` y las rutas de entrada. El EDA caracteriza los datos; la limpieza definitiva aplica y documenta sus filtros; los vectores utilizan el resultado definitivo. Para igualar la ejecución conservada, fijar DuckDB 1.3.2. El EDA requiere al menos 15 GiB libres en `/content` y la limpieza al menos 10 GiB. Ambos limitan DuckDB a 3 GB de memoria.

El CSV original y el CSV limpio masivo se mantienen fuera de Git. El Parquet limpio está publicado. Para regenerar el CSV desde ese Parquet:

```powershell
python scripts/exportar_csv_limpio.py
```

La salida es `data/processed/PM25_PM10_limpio.csv`. Los archivos `features_1_hora.csv` y `features_24_horas.csv` ya están publicados en `data/vectors/`.

## Carga y validación en Go

Desde la raíz del repositorio:

```powershell
cd go/cargar-datos
go run ./cmd/cargar-datos -mode=sequential ../../data/processed/PM25_PM10_limpio.csv
go run ./cmd/cargar-datos -mode=concurrent -workers=4 -batch=1000 ../../data/processed/PM25_PM10_limpio.csv
go test -count=1 ./...
go test -race -count=1 ./...
cd ../..
```

Un lector secuencial distribuye la validación en lotes entre workers mediante canales. Se verifican encabezados, contaminante, unidad y concentraciones numéricas, finitas y no negativas. Este módulo valida una entrada ya limpia; no repite los filtros del notebook ni construye vectores. Las evidencias publicadas aceptan 1 062 897 registros y rechazan cero en ambos modos: 598 502 de PM10 y 464 395 de PM2.5.

## K-Means y pruebas con los vectores reales

Desde la raíz:

```powershell
cd go/kmeans_secuencial
go run . -input ../../data/vectors/features_1_hora.csv -period 1_hora -mode sequential -k 4 -seed 42 -tol 0.001 -max-iter 100
go run . -input ../../data/vectors/features_1_hora.csv -period 1_hora -mode concurrent -workers 4 -k 4 -seed 42 -tol 0.001 -max-iter 100
go run . -input ../../data/vectors/features_24_horas.csv -period 24_horas -mode sequential -k 4 -seed 42 -tol 0.001 -max-iter 100
go run . -input ../../data/vectors/features_24_horas.csv -period 24_horas -mode concurrent -workers 4 -k 4 -seed 42 -tol 0.001 -max-iter 100
$env:KMEANS_REAL_DATA = "1"
go test -count=1 -timeout=300s ./...
go test -race -count=1 -timeout=300s ./...
Remove-Item Env:KMEANS_REAL_DATA
cd ../..
```

La salida JSON registra parámetros, centroides, conteos, inercia y tiempos. La variante concurrente utiliza bloques contiguos, acumuladores privados y `sync.WaitGroup`. Las pruebas reales comparan ambos periodos con 1, 2, 4 y 8 workers y verifican asignaciones, centroides e inercia con las tolerancias del código.

Las ejecuciones de Colab se conservan en los cuadernos [secuencial](go/kmeans_secuencial/kmeans_secuencial.ipynb) y [concurrente](go/kmeans_secuencial/kmeans_concurrente.ipynb).

## Rendimiento y procedencia

Cada speedup utiliza las medianas secuencial y concurrente del mismo periodo y entorno. No se combinan tiempos entre equipos.

| Entorno | Evidencia | Entrenamiento horario con 4 workers |
| --- | --- | --- |
| Colab | [Cuaderno ejecutado](go/kmeans_secuencial/kmeans_concurrente.ipynb), 7 de noviembre de 2026 a las 08:55 UTC | Secuencial: 1,783167 s; concurrente: 1,175477 s; speedup: 1,517 |
| Windows | [Entorno](results/rendimiento/ejecucion-local/entorno.json), [mediciones](results/rendimiento/ejecucion-local/mediciones.csv) y [medianas](results/rendimiento/ejecucion-local/RESUMEN.md), 7 de noviembre de 2026 a las 00:25 UTC | Secuencial: 0,811010 s; concurrente: 0,301061 s; speedup: 2,694 |

En Windows se registraron Go 1.27.2, Windows 11, Intel Core i7-12650H de 10 núcleos y 16 procesadores lógicos y aproximadamente 15,6 GiB de memoria física. En Colab no se conservaron las respuestas completas de versión de Go, sistema operativo, CPU y RAM. Ambas evaluaciones fijaron `GOMAXPROCS=4`, cinco repeticiones medidas y un calentamiento por configuración.

Para medir nuevamente, desde la raíz y usando una carpeta nueva:

```powershell
python scripts/medir_kmeans.py --repetitions 5 --gomaxprocs 4 --output results/rendimiento/nueva-ejecucion
python scripts/graficar_rendimiento.py results/rendimiento/nueva-ejecucion
```

Se excluyen la preparación de datos, la compilación y las pruebas. La mediana del tiempo total se calcula a partir de sus propias mediciones. La versión concurrente fusiona asignación y acumulación: el speedup no aísla exclusivamente el efecto de aumentar workers. Los clústeres no constituyen umbrales sanitarios ni un modelo predictivo.

## Plataforma distribuida e interfaz TB2

`go/plataforma` contiene API, coordinador, nodos y servidor UI. La API carga los vectores locales de cada periodo al arrancar y publica `GET /datasets`, `POST /jobs`, `GET /jobs/{id}` y `GET /jobs/{id}/results`; también expone `/health`, `/docs` y `/openapi.yaml`. MongoDB conserva los trabajos y resultados; Redis funciona como caché, con lectura desde MongoDB cuando no está disponible. El entrenamiento distribuye el trabajo entre nodos mediante TCP. No existe un listado global de trabajos.

La SPA presenta Inicio, Datos, Ejecuciones, Panel general, Resultados y Análisis ambiental. El catálogo usa la API real y conserva la selección durante la navegación. Ejecuciones crea trabajos asíncronos y consulta su estado. Resultados recibe `job_id` en la URL, permite elegir ejecuciones mediante nombres legibles y muestra los gráficos existentes. El panel consulta referencias conocidas de esta sesión: guarda únicamente identificadores válidos, hasta 50, revalida los 10 más recientes y no considera los datos locales como estados confirmados. La consulta de estados tiene concurrencia máxima de tres. La interfaz distingue latencia HTTP y tiempo de entrenamiento (`segundos_ml`).

Con Docker Desktop en ejecución y los CSV en `data/vectors`, desde la raíz:

```powershell
docker compose -p eas-tb2-integracion -f go/plataforma/compose.yml config --quiet
docker compose -p eas-tb2-integracion -f go/plataforma/compose.yml build
docker compose -p eas-tb2-integracion -f go/plataforma/compose.yml up -d --wait
Invoke-RestMethod http://127.0.0.1:8090/api/health
Invoke-RestMethod http://127.0.0.1:8090/api/datasets
docker compose -p eas-tb2-integracion -f go/plataforma/compose.yml ps
# Abrir http://127.0.0.1:8090/
docker compose -p eas-tb2-integracion -f go/plataforma/compose.yml down
```

`down` conserva los volúmenes; no usar `down -v` para una parada habitual. La UI es el único puerto publicado, ligado a `127.0.0.1:8090`; la API, TCP y bases permanecen en la red interna. Si el puerto ya está ocupado, revisa la instancia existente antes de iniciar otra. Este despliegue es de laboratorio: no implementa autenticación ni TLS para acceso público. Mantiene los límites de creación de trabajos, cola, tiempo de ejecución y consultas existentes. No deben publicarse secretos ni datasets masivos. La predicción temporal sigue pendiente respecto de los objetivos: K-Means agrupa mediciones históricas, sin pronósticos, alertas sanitarias ni categorías clínicas.

Las pruebas unitarias e integradas de UI están en `go/plataforma/internal/ui/tests`; su [guía](go/plataforma/internal/ui/README.md) incluye ejecución local, navegador, datos sintéticos y alternativa Docker ante bloqueos de Windows. Las pruebas Go del módulo se ejecutan con `go test -count=1 ./...`, `go vet ./...` y `go test -race -count=1 ./...` desde `go/plataforma`; `-race` requiere CGO y compilador C. Las campañas históricas de rendimiento anteriores no representan una ejecución actual. La campaña de integración del 27 de noviembre de 2026 se conserva en `results/evaluacion-web/2026-11-27-integracion-web-tb2-110325`; allí se identifican datos, entorno, revisión, capturas, resultados y límites.

## Integridad

Los inventarios conservan rutas, tamaños, cantidades de filas cuando corresponden y SHA-256. No incluyen archivos eliminados ni el propio inventario. Después de modificar resultados:

```powershell
python scripts/actualizar_inventarios.py
```

La configuración local del editor, incluida `.idea/workspace.xml`, permanece fuera del versionado.
