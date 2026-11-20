# Reproducción de K-Means distribuido

## PowerShell con Docker Desktop

Desde la raíz del repositorio, con Docker Desktop en modo Linux:

```powershell
docker info --format '{{.OSType}}'
powershell -NoProfile -ExecutionPolicy Bypass -File .\results\pruebas\kmeans-distribuido\ejecutar.ps1
```

[ejecutar.ps1](ejecutar.ps1) contiene los comandos Docker de la ejecución del 20 de noviembre de 2026. Usa la imagen `golang:1.27-bookworm` y ejecuta `go test ./...`, `go test -race ./...` y `go vet ./...` en `go/plataforma`. Compila los ejecutables una vez y, por caso, ejecuta tres contenedores Linux en una red privada de Docker Desktop sobre un solo equipo Windows. La imagen y la versión de Go observadas están en [entorno.log](entorno.log).

Cada caso usa dos nodos, semilla 42, tolerancia 0.0001 y un máximo de 100 actualizaciones. Los JSON guardan la configuración y los resultados. Cada archivo `*-verificacion.log` contiene la cuenta independiente del CSV, la suma de conteos, el hash SHA-256 de la entrada y los tres códigos de salida. Los hashes reales coinciden con [el manifiesto de vectores](../../../data/vectors/manifest.json). Volver a ejecutar el script actualiza los archivos de evidencia.

## PowerShell con Go local

Ejemplo con el conjunto horario. Desde la raíz del repositorio, ejecuta una línea por terminal y espera a que el coordinador muestre `listening` antes de iniciar los nodos. El coordinador resuelve las rutas de entrada y salida desde `go/plataforma`.

```powershell
go -C .\go\plataforma run ./cmd/coordinator -listen 127.0.0.1:9000 -file ../../data/vectors/features_1_hora.csv -nodes 2 -k 4 -seed 42 -tol 0.0001 -max-iter 100 -output ../../results/pruebas/kmeans-distribuido/local-horario.json
go -C .\go\plataforma run ./cmd/node -coordinator 127.0.0.1:9000 -id nodo-1
go -C .\go\plataforma run ./cmd/node -coordinator 127.0.0.1:9000 -id nodo-2
```

Para una ejecución separada de 24 horas, usa `features_24_horas.csv` y `local-24-horas.json`, y vuelve a iniciar ambos nodos. Si una política de Windows bloquea los ejecutables locales generados por Go, usa la reproducción Docker.

## Resultados observados con Docker

| Caso | K | Filas del CSV | Suma de conteos | Iteraciones | Motivo de finalización | Resultado |
| --- | ---: | ---: | ---: | ---: | --- | --- |
| Sintético | 2 | 4 | 4 | 3 | Convergencia | [JSON](sintetico-result.json) |
| Real, 1 hora | 4 | 445 347 | 445 347 | 100 | Máximo de iteraciones | [JSON](real-1_hora-result.json) |
| Real, 24 horas | 4 | 18 297 | 18 297 | 44 | Convergencia | [JSON](real-24_horas-result.json) |

El caso sintético obtuvo centroides `[1,1]` y `[11,11]`, con dos observaciones en cada clúster. Los nueve contenedores de aplicación salieron con código 0. Las pruebas normales, el detector de carreras y `go vet` terminaron con código 0. [Registro de pruebas](go-test-vet.log)

Los tiempos de cada fase están en los JSON. Son mediciones de esta reproducción en un solo equipo.
