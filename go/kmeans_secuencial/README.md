# K-Means secuencial en Go

## Modos integrados: secuencial y concurrente

El módulo también incluye `RunKMeansConcurrent`, pruebas de integración y salida JSON. Se conserva la ejecución histórica sin argumentos. Para seleccionar explícitamente un periodo:

```powershell
go run . -input ../../data/vectors/features_1_hora.csv -period 1_hora -mode concurrent -workers 4
go run . -input ../../data/vectors/features_1_hora.csv -period 1_hora -mode sequential
```

La [guía de integración](../../docs/pruebas-integracion.md) describe parámetros, pruebas y recorrido de datos. La [entrega del integrante 3](../../docs/ENTREGA_INTEGRANTE_3.md) reúne informe, evidencias y rendimiento. Para incluir datos reales en las pruebas, establecer `KMEANS_REAL_DATA=1` antes de `go test ./...`.

## Descripción

Implementación secuencial del algoritmo K-Means en Go, utilizada como línea base para comparar posteriormente su rendimiento con la implementación paralela.

El algoritmo trabaja con dos dimensiones:

- PM10
- PM2.5

Se utiliza `k = 4` para formar cuatro clústeres.

## Estructura del proyecto

```text
kmeans_secuencial/
├── go.mod
├── README.md
├── main.go
├── dataset.go
├── loader.go
├── kmeans.go
├── kmeans_test.go
└── loader_test.go
```

## Requisitos

- Go 1.22 o superior.
- Archivos CSV con las columnas `PM10` y `PM2.5`.
- Los archivos `features_1_hora.csv` y `features_24_horas.csv` deben estar ubicados en `data/vectors/` dentro del repositorio principal.

## Ejecución

El programa procesa automáticamente los dos conjuntos de datos: una hora y 24 horas.

Todos los comandos deben ejecutarse desde la carpeta `go/kmeans_secuencial`, ubicada dentro del repositorio principal.

Desde la raíz del repositorio, ingresar a la carpeta:

```powershell
cd go/kmeans_secuencial
```

**Ejecutar K-Means secuencial para ambos datasets:**

```powershell
go run .
```

El programa muestra para cada conjunto de datos:

- Cantidad de vectores y dimensiones.
- Estadísticas descriptivas de PM10 y PM2.5.
- Percentiles P50, P90, P95 y P99.
- Número de iteraciones y estado de convergencia.
- Inercia y tiempos de ejecución.
- Centroides finales y cantidad de puntos por clúster.

## Pruebas

Los archivos de pruebas permiten validar la carga de datos y el funcionamiento del algoritmo K-Means.

Todos los comandos siguientes se ejecutan desde `go/kmeans_secuencial`.

**1. Pruebas del lector de CSV (`loader_test.go`):**

```powershell
go test -v loader_test.go loader.go dataset.go
```

**2. Pruebas del algoritmo K-Means (`kmeans_test.go`):**

```powershell
go test -v kmeans_test.go kmeans.go dataset.go
```

**3. Ejecutar todas las pruebas:**

```powershell
go test -v ./...
```

## Parámetros de K-Means

| Parámetro | Valor |
|---|---:|
| Número de clústeres (k) | 4 |
| Máximo de iteraciones | 100 |
| Tolerancia | 0.001 |
| Semilla aleatoria | 42 |
