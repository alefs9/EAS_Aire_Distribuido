# N2 — Integración y pruebas

Responsable actual: Valentina (integrante 3; N2 figuraba como Nicole en el plan). Base: `develop` en `a01883ddf398a704d0bbfe7855cd19e336aa5ccc`.

## Recorrido real y unidades

1. `notebooks/Limpieza.ipynb` produce las 1 062 897 mediciones limpias; `data/processed/PM25_PM10_limpio.parquet` permite regenerar su CSV.
2. `go/cargar-datos` valida ese CSV: contaminante, concentración y unidad; informa aceptados, rechazados y conteos por contaminante. Su API devuelve estadísticas, no vectores ni asociaciones temporales.
3. `notebooks/Vectores.ipynb` realiza el emparejamiento por claves de procedencia, ubicación y tiempo. Los CSV publicados contienen 445 347 vectores horarios y 18 297 diarios. La preparación ya ejecutada está auditada en `results/vectores`; no se repitió aquí el procesamiento del CSV bruto de 4,8 GB.
4. `LoadCSV` de `go/kmeans_secuencial` carga el archivo de características seleccionado y entrega el mismo `Dataset` a `RunKMeans` o `RunKMeansConcurrent`. El comando común serializa resultados y mediciones en JSON.

La integración reutiliza los dos módulos existentes mediante sus comandos. No convierte la salida estadística del cargador en puntos ficticios ni empareja mediciones por posición. El paso de mediciones a vectores sigue siendo el cuaderno publicado; no se afirma que exista un pipeline enteramente en Go para ese emparejamiento.

## Reproducción desde un clon

Requisitos: Go compatible con ambos `go.mod`, Python 3 y DuckDB 1.3.2 para regenerar el CSV. Para `-race` en Windows: CGO habilitado y compilador C compatible. Ejecutar desde la raíz:

```powershell
python -m pip install -r requirements.txt
python scripts/exportar_csv_limpio.py
python scripts/verificar_integracion.py
$env:CGO_ENABLED='1'
$env:CC='clang' # LLVM-MinGW en PATH; también se puede usar un GCC compatible
python scripts/verificar_integracion.py --race
```

En este equipo DuckDB 1.3.2 importó pero terminó con un error nativo (`-1073741819`). Se exportó el mismo Parquet mediante PyArrow, sin filtros, agrupaciones ni cambios de filas:

```powershell
python -c "import pyarrow.parquet as pq, pyarrow.csv as csv; t=pq.read_table('data/processed/PM25_PM10_limpio.parquet'); print(t.num_rows); csv.write_csv(t,'data/processed/PM25_PM10_limpio.csv')"
```

El resultado contiene 1 062 897 filas, y se comprueba con ambos modos del cargador. PyArrow es una alternativa de exportación; no cambia la preparación ni los modelos. El CSV generado permanece excluido de Git.

## Comando común

```powershell
cd go/kmeans_secuencial
go run . -input ../../data/vectors/features_1_hora.csv -period 1_hora -mode concurrent -workers 4
go run . -input ../../data/vectors/features_1_hora.csv -period 1_hora -mode sequential
go run . -input ../../data/vectors/features_24_horas.csv -period 24_horas -mode concurrent -workers 4
```

`-k 4 -max-iter 100 -tol 0.001 -seed 42` son los valores predeterminados. `-labels` incluye todas las asignaciones. Siempre se emiten conteos y SHA-256 de las etiquetas (enteros uint64 little-endian). El periodo es obligatorio con flags; como el CSV tiene solo dos columnas, el programa no puede deducir su procedencia temporal: corresponde al usuario seleccionar el archivo correcto. El modo histórico `go run .` se conserva y procesa separadamente los dos periodos con la referencia.

## Casos y aceptación

| Caso | Criterio | Evidencia |
|---|---|---|
| Preparación y balance | 1 062 897 mediciones; 463 644 pares completos; 135 609 mediciones sin pareja | Auditorías previas de Alexandra en `results/limpieza` y `results/vectores` |
| Lector y validación de mediciones | Encabezados, comillas, unidades, no finitos y conteos | Pruebas existentes en `go/cargar-datos`; ejecutadas nuevamente |
| Referencia secuencial | Casos conocidos, inicialización, vacío e inválidos | Pruebas existentes en `go/kmeans_secuencial`; ejecutadas nuevamente |
| Particiones | Cobertura única, sin huecos, balance máximo de una fila | `TestPartitionsCoverExactlyOnce` |
| Workers y casos límite | 1, 2, 3, 4 y 100 workers; empates, vacíos, máximo de iteraciones | `TestConcurrentEquivalence` |
| Validación concurrente | Rechazar configuración y estructuras inválidas | `TestConcurrentRejectsInvalidInput` |
| No mutación y repetición | Dataset intacto y resultados repetibles | `TestConcurrentRepeatableAndDoesNotMutate` |
| CSV → modelos → JSON | Mismas etiquetas, centroides, conteos e inercia | `TestCSVToBothModels` |
| Contrato de CLI | Rechazar periodo, modo, workers, ruta y CSV inválidos | `TestCLIRejectsInvalidContract` |
| Dos periodos reales | Conteos separados, igualdad de etiquetas e iteraciones; centroides e inercia dentro de 1e-9 abs + 1e-9 rel | `TestRealPeriodsEquivalence`, habilitada con `KMEANS_REAL_DATA=1` |
| Carga masiva | 1 062 897 aceptados, cero rechazados, 598 502 PM10 y 464 395 PM2.5 | `scripts/verificar_integracion.py` |

`go test ./...` debe ejecutarse dentro de **cada módulo**, porque no existe módulo raíz. El script conserva fecha UTC, commit, comandos, entorno y salidas completas en `results/pruebas`. Las pruebas de datos reales comparan 1, 2, 4 y 8 workers, incluyendo asignación final y reducción. El detector de carreras respalda únicamente las rutas recorridas. El primer intento falló porque CGO separó la ruta del compilador por el espacio de `Concurrente 2`; se corrigió incorporando su directorio a PATH y usando `CC=clang`. El intento se conserva por separado.

La reproducción independiente por otra integrante y su revisión siguen siendo tareas del equipo; no se presentan como realizadas por esta ejecución.

Resultado final: las pruebas normales y con detector de carreras finalizaron en PASS en ambos módulos, incluyendo los dos periodos reales. El segundo intento encontró una restricción del enlazador en el sandbox; al ejecutar en el entorno compatible de Windows con LLVM-MinGW se completó la comprobación. Los logs conservan ambos fallos de preparación y la ejecución final. No se detectaron carreras en las rutas ensayadas.
