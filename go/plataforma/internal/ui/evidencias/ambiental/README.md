# Exploración y comparación ambiental · 5.3.2

La vista `/analisis-ambiental` consulta uno o dos trabajos completados mediante
`GET /jobs/{id}/results`. Seleccionar su periodo antes de consultar. El segundo
identificador es opcional. Los filtros de contaminante y grupo se aplican a los
perfiles ya consultados y no modifican ni recalculan los resultados.

## Evidencias

Se usaron los CSV existentes, dos nodos TCP reales y la API del proyecto, con
almacenamiento temporal en memoria. Estas pruebas no validan MongoDB ni Redis.
Las respuestas completas están en `24-horas-a.json`, `24-horas-b.json` y
`una-hora.json`. Semilla 42, tolerancia 0.0001 y máximo de 100 iteraciones.

| Trabajo | Periodo | k | Observaciones | Iteraciones |
| --- | --- | ---: | ---: | ---: |
| A | 24 horas | 4 | 18297 | 44 |
| B | 24 horas | 3 | 18297 | 22 |
| Horario | Una hora | 4 | 445347 | 100, sin convergencia |

- `filtros-24-horas.jpg`: dos trabajos del mismo periodo, PM2.5 y grupo 2.
- `comparacion-filtrada.jpg`: valor de A, 93.1765084471024 µg/m³ y 992 observaciones.
  B conserva 86.65851210097698 µg/m³ y 1298 observaciones. Los números de grupo
  no establecen equivalencia entre ejecuciones.
- `perfil-una-hora.jpg`: consulta independiente del periodo horario.
- `periodos-incompatibles.jpg`: se rechaza comparar periodos distintos.
- `filtros-movil.jpg`: controles adaptados a una pantalla estrecha; la comprobación
  del DOM no detectó desbordamiento horizontal de la página.

Las barras horizontales muestran la concentración del centroide en µg/m³; las
filas identifican grupo y contaminante. Ambas ejecuciones usan la misma escala,
calculada con los valores visibles. La tabla conserva los valores exactos.
Cambiar filtros puede cambiar el máximo de esa escala, indicado en el texto.

La cobertura corresponde a pares completos del conjunto identificado. Los filtros
no cambian ese denominador. No se mezclan periodos, se imputan ubicaciones ni se
interpretan conteos como personas. Los perfiles no son alertas ni pronósticos.

## Reproducir

Desde `go/plataforma`, iniciar:

```powershell
go run internal/ui/tests/environment_browser.go
```

En otra terminal:

```powershell
$body = @{periodo='24_horas'; k=4; semilla=42; tolerancia=0.0001; max_iter=100; nodos=2} | ConvertTo-Json
$job = Invoke-RestMethod http://127.0.0.1:8088/jobs -Method Post -ContentType application/json -Body $body
Invoke-RestMethod "http://127.0.0.1:8088/jobs/$($job.job_id)"
```

Al completarse, abrir `http://127.0.0.1:8096/analisis-ambiental` e introducir su ID.
Repetir con k=3 para el trabajo B. Para comprobar el periodo horario, crear otro
trabajo con periodo `1_hora` y k=4. Los identificadores cambian en cada ejecución.
Cerrar el servidor con Ctrl+C.

Pruebas: `node --test internal/ui/tests/*.test.mjs` (43 aprobadas) y `go test ./...`.

## Alcance de la consulta y coordinación de API

La API actual entrega centroides y conteos, sin observaciones individuales, fechas
ni claves de monitoreo. Por ello, esta rama permite explorar perfiles de K-Means;
no ofrece filtros espaciales o temporales que la consulta no puede respaldar.
No necesita endpoints nuevos para este alcance.

Para una futura consulta por fecha o punto se deberá acordar con la responsable
de la API el contrato y la recuperación mediante `vector_id` y los metadatos del
Parquet, preservando periodo, unidad y valores desconocidos. No se ha enviado
ningún mensaje ni se da por acordado ese contrato. El procesamiento de indicadores
y la interpretación detallada corresponden a la rama del punto 5.3.3.
