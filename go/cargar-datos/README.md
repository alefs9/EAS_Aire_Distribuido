# Cargador y validador de datos en Go

## 1. Descripción

Este módulo permite leer y validar el dataset limpio de mediciones de PM10 y PM2.5 mediante Go.

El procesamiento se realiza sobre un archivo CSV que contiene 1 062 897 mediciones.

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
│       └── validation.go
└── go.mod
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

Desde la raíz del repositorio:

```bash
cd go/cargar-datos
go run ./cmd/cargar-datos ../../data/processed/PM25_PM10_limpio.csv
```

Para comprobar que el código compila y ejecutar las pruebas disponibles:

```bash
go test ./...
```

## 6. Validaciones realizadas

El cargador comprueba:

- Presencia de las columnas `contaminante`, `concentracion` y `unidad`.
- Contaminantes permitidos: `PM10` y `PM2.5`.
- Unidad esperada: `µg/m³`.
- Concentraciones numéricas, finitas y no negativas.
- Registros con campos necesarios incompletos.

El cargador trabaja sobre el dataset previamente limpio. No vuelve a ejecutar los filtros de limpieza aplicados durante la preparación del dataset.

## 7. Resultado de la ejecución secuencial inicial

**Dataset:** `PM25_PM10_limpio.csv`  
**Registros procesados:** 1 062 897

| Resultado | Cantidad |
|---|---:|
| Registros leídos | 1 062 897 |
| Registros aceptados | 1 062 897 |
| Registros rechazados | 0 |
| Mediciones PM10 | 598 502 |
| Mediciones PM2.5 | 464 395 |

Los conteos se encuentran reconciliados:

```text
Registros leídos = Registros aceptados + Registros rechazados
```

Esta ejecución corresponde a la versión secuencial inicial del cargador. Se utilizará como referencia para comparar posteriormente el procesamiento concurrente.