# Verificación de resultados

Se consultó la API local con los dos CSV del repositorio, k=4, semilla=42,
tolerancia=0.0001, máximo de 100 iteraciones y dos nodos TCP reales.
El almacenamiento de esta prueba es en memoria: no verifica MongoDB ni Redis.

| Periodo | Observaciones | Iteraciones | Finalización |
| --- | ---: | ---: | --- |
| Una hora | 445347 | 100 | Límite de iteraciones sin convergencia |
| 24 horas | 18297 | 44 | Convergencia |

Los archivos `1_hora.json` y `24_horas.json` conservan las respuestas completas.
Las capturas muestran esos mismos identificadores, parámetros, centroides y conteos.
Se comprobó la consulta de un identificador inexistente y el diseño móvil de 390 px,
sin desbordamiento horizontal de la página. La tabla tiene desplazamiento propio.
Los estados pendientes, fallidos, las respuestas inconsistentes y la cancelación
se verifican en `tests/results.test.mjs`.

Los CSV de Windows tienen CRLF; al convertir únicamente sus saltos de línea a LF,
sus hashes coinciden con `data/vectors/manifest.json`. La pantalla reconoce ambas
variantes verificadas, conserva el identificador original de la API y comprueba
el total de observaciones. Para otros conjuntos no atribuye unidades ni escala.

## Reproducir

Desde `go/plataforma`:

```powershell
go run internal/ui/tests/results_browser.go
```

En otra terminal, crear un trabajo y consultar su estado hasta que termine:

```powershell
$body = @{periodo='1_hora'; k=4; semilla=42; tolerancia=0.0001; max_iter=100; nodos=2} | ConvertTo-Json
$job = Invoke-RestMethod http://127.0.0.1:8088/jobs -Method Post -ContentType application/json -Body $body
Invoke-RestMethod "http://127.0.0.1:8088/jobs/$($job.job_id)"
```

Abrir `http://127.0.0.1:8096/resultados`, escribir el identificador y pulsar Consultar.
Repetir con `periodo='24_horas'`. Cada ejecución genera identificadores nuevos.
Detener el servidor con Ctrl+C al terminar.

Pruebas ejecutadas: `node --test internal/ui/tests/*.test.mjs` y `go test ./...`.
Los resultados son concentraciones originales en µg/m³, no datos normalizados.
Los grupos no equivalen a categorías sanitarias ni a población expuesta.
