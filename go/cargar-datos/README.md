# Cargador y validador de datos en Go

## 1. Descripción

Este módulo permite leer y validar el dataset limpio de mediciones de PM10 y PM2.5 mediante Go.

El procesamiento se realiza sobre un archivo CSV que contiene 1 062 897 mediciones. Se implementaron dos modalidades:

- **Secuencial:** procesa los registros uno por uno y sirve como referencia para comparar resultados.
- **Concurrente:** distribuye los registros en lotes entre varios workers para realizar la validación concurrentemente.

## 2. Requisitos

- Go (versión 1.25.0 o compatible con el módulo).
- Archivo CSV limpio: `data/processed/PM25_PM10_limpio.csv`.

## 3. Estructura del módulo

```text
go/cargar-datos/
├── cmd/
│   └── cargar-datos/
│       └── main.go
├── internal/
│   └── dataset/
│       ├── reader.go
│       ├── validation.go
│       ├── pipeline.go
│       ├── reader_test.go
│       ├── validation_test.go
│       └── pipeline_test.go
├── go.mod
└── README.md
```

## 4. Preparación

El CSV se genera desde el archivo Parquet del proyecto. Desde la raíz del repositorio, ejecutar:

```bash
python scripts/exportar_csv_limpio.py
```

El archivo generado se encuentra en:

```text
data/processed/PM25_PM10_limpio.csv
```

No se requieren dependencias externas de Go para este módulo.

## 5. Ejecución

Los siguientes comandos se ejecutan desde la carpeta `go/cargar-datos`:

```powershell
cd go/cargar-datos
```

### 5.1. Procesamiento secuencial

Permite procesar los registros uno por uno y obtener resultados de referencia:

```powershell
go run ./cmd/cargar-datos -mode=sequential ../../data/processed/PM25_PM10_limpio.csv
```

### 5.2. Procesamiento concurrente

Ejecuta el procesamiento utilizando 4 workers y lotes de 1 000 registros:

```powershell
go run ./cmd/cargar-datos -mode=concurrent -workers=4 -batch=1000 ../../data/processed/PM25_PM10_limpio.csv
```

Los parámetros pueden modificarse:

- `-mode`: modalidad de ejecución (`sequential` o `concurrent`).
- `-workers`: cantidad de workers concurrentes.
- `-batch`: cantidad de registros por lote.

El modo concurrente es el modo predeterminado si no se especifica `-mode`.

**Importante:** las opciones deben colocarse antes de la ruta del archivo CSV.

## 6. Pruebas automatizadas

El módulo incluye tres archivos de pruebas:

- `reader_test.go`: comprueba la apertura del CSV, los encabezados requeridos, los encabezados duplicados o vacíos y la lectura de textos con comas y comillas.
- `validation_test.go`: comprueba la validación de contaminantes, unidades y concentraciones, incluyendo valores negativos, no numéricos y no finitos.
- `pipeline_test.go`: comprueba los modos secuencial y concurrente, la coincidencia de sus resultados, la reconciliación de conteos, los parámetros inválidos y el procesamiento de un dataset vacío.

Para ejecutar todas las pruebas desde `go/cargar-datos`:

```powershell
go test ./...
```

Para visualizar cada prueba individualmente:

```powershell
go test -v ./...
```

Para comprobar el formato del código de pruebas:

```powershell
gofmt -w internal/dataset/reader_test.go
gofmt -w internal/dataset/validation_test.go
gofmt -w internal/dataset/pipeline_test.go
```

Las pruebas utilizan archivos CSV temporales, por lo que no necesitan procesar el dataset completo.

## 7. Validaciones realizadas

El cargador comprueba:

- Presencia de las columnas `contaminante`, `concentracion` y `unidad`.
- Encabezados vacíos o duplicados.
- Contaminantes permitidos: `PM10` y `PM2.5`.
- Unidad esperada: `µg/m³`.
- Concentraciones numéricas, finitas y no negativas.
- Registros con campos necesarios incompletos.

El cargador trabaja sobre el dataset previamente limpio. No vuelve a ejecutar los filtros de limpieza aplicados durante la preparación del dataset.

## 8. Resultados de la ejecución completa

**Dataset:** `PM25_PM10_limpio.csv`  
**Registros procesados:** 1 062 897

### 8.1. Resultados del procesamiento secuencial

| Resultado | Cantidad |
|---|---:|
| Registros leídos | 1 062 897 |
| Registros aceptados | 1 062 897 |
| Registros rechazados | 0 |
| Mediciones PM10 | 598 502 |
| Mediciones PM2.5 | 464 395 |

**Tiempo de procesamiento:** 1.7916722 segundos.

### 8.2. Resultados del procesamiento concurrente

**Configuración:** 4 workers y lotes de 1 000 registros.

| Resultado | Cantidad |
|---|---:|
| Registros leídos | 1 062 897 |
| Registros aceptados | 1 062 897 |
| Registros rechazados | 0 |
| Mediciones PM10 | 598 502 |
| Mediciones PM2.5 | 464 395 |

**Tiempo de procesamiento:** 1.5947679 segundos.

### 8.3. Reconciliación de conteos

En ambas modalidades se cumple:

```text
Registros leídos = Registros aceptados + Registros rechazados
```

El resultado de la ejecución fue:

```text
Conteos reconciliados: true
```

Los dos modos produjeron los mismos conteos. En las ejecuciones realizadas, el modo concurrente presentó un menor tiempo de procesamiento. Estos tiempos corresponden a ejecuciones individuales y no constituyen por sí solos un benchmark definitivo.

## 9. Consideraciones sobre el procesamiento concurrente

La implementación utiliza goroutines para ejecutar los workers concurrentemente y canales para distribuir los lotes de registros.

Cada worker mantiene sus propios conteos y resultados de validación. Al finalizar el procesamiento, los resultados individuales se consolidan para obtener los conteos generales.

La sincronización se realiza mediante `sync.WaitGroup`, y el canal de trabajos se cierra una vez terminada la lectura del CSV.

El procesamiento concurrente permite configurar la cantidad de workers y el tamaño de los lotes desde la línea de comandos.