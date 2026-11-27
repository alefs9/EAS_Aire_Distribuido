# Evidencias de corrección de integración TB2 — 27/11/2026

Rama: `fix/integracion-web-tb2`, creada desde develop `37fd4f4a32c39e72edf0e43137331d68cd462af5`. Código de aplicación de las capturas: correcciones hasta `9870298`; revisión completa capturada `7167f31` (el JSON conserva el hash completo). Las imágenes se obtuvieron sin modificar la aplicación normal para simular resultados. El Docker UI se reconstruyó después de la última corrección de sus recursos.

## Entorno y resultados

Windows 11, Go host 1.27.2, Node 24.21.0 y Edge headless instalado. Docker Desktop con motor 29.8.2 y Compose 5.6.0. Las imágenes exactas y digests se conservan en `imagenes.txt`; las pruebas Go se ejecutaron en `golang:1.27-bookworm`. La publicación de puertos se registra en `compose-ps-final.txt`: solo UI 127.0.0.1:8090; API, coordinador y bases permanecen en red interna. Compose conserva límites y políticas existentes, sin autenticación ni TLS públicos.

- `node-unit.log`: 60 pruebas JavaScript aprobadas, sin omisiones.
- `playwright-fixtures.json` y `.log`: 12 pruebas integradas en Edge aprobadas con respuestas **sintéticas**. Comprueban entrada real del módulo, las seis rutas, navegación, consulta manual y por URL, recarga, cambio de query, cancelación, panel, estados de error y ocultación del enlace de resultados mientras se ejecuta un trabajo. No son evidencia de integración con bases ni nodos reales. No se publican capturas de fixtures como resultados científicos.
- `go-test.log`: intento nativo de Windows **fallido** por bloqueo de ejecutables temporales; se conserva como limitación, no como suite aprobada. El host tiene CGO deshabilitado y carece de compilador C.
- `go-linux-test-vet-race.log`: ejecución con salida final 0 de `go test -count=1 ./...`, `go vet ./...` y `go test -race -count=1 ./...` dentro de Docker Linux, todos satisfactorios. No se modificaron políticas de Windows.
- `validacion-entrega/recorrido-real.json`: recorrido con API, MongoDB, Redis y dos nodos reales, `success: true`, sin errores JavaScript. No usa fixtures ni interceptación HTTP.

Datos: catálogo real con 445 347 vectores de una hora y 18 297 de 24 horas; hashes completos en el JSON. Se seleccionaron 24 horas, k=4, semilla=42, tolerancia=0.0001, máximo 100 iteraciones y dos nodos. Trabajo `pc4-86357b4c37f392421dd8a01d00956c15`: convergencia en 44 iteraciones; conteos [10075, 992, 4848, 2382]; inercia 6283274.552233539; tiempo de entrenamiento 0.018809535 s, una observación funcional, no una mediana de rendimiento ni latencia HTTP. El origen registrado identifica nodo-1 y nodo-2. La comprobación no cambió los CSV originales.

## Reproducir

Desde la raíz del repositorio, PowerShell, con Docker Desktop iniciado:

```powershell
docker compose -p eas-tb2-integracion -f go/plataforma/compose.yml config --quiet
docker compose -p eas-tb2-integracion -f go/plataforma/compose.yml build
docker compose -p eas-tb2-integracion -f go/plataforma/compose.yml up -d --wait
npm ci --prefix go/plataforma/internal/ui/tests/browser
$env:PLAYWRIGHT_CHANNEL='msedge'
node go/plataforma/internal/ui/tests/browser/real-compose.mjs results/evaluacion-web/nueva-campana-integracion
```

La carpeta de salida debe ser nueva; el script rechaza sobrescribir campañas. Comprobar `recorrido-real.json` y `capturas/`. URL http://127.0.0.1:8090; escritorio 1280 × 900, móvil 390 × 844. El script espera a los estados requeridos y las imágenes b desplazan la página hasta gráficos o filas, conservando texto de contexto. Reproduce selección horaria y de 24 horas, transferencia a ejecución, creación y seguimiento, resultados del mismo ID, recarga, panel y perfiles ambientales. Los jobs reales se conservan en MongoDB y los volúmenes no se eliminaron.

Para las 12 pruebas sintéticas y las pruebas Go/race, los comandos exactos, incluido el servidor de fixtures en Docker, están en `go/plataforma/internal/ui/README.md`. Los logs se produjeron con `Tee-Object`; `compose-build*.log`, `compose-up*.log` y `servicios-reales.log` conservan arranque, reconstrucción y actividad real. Los logs de texto solo normalizan espacios al final de línea; el registro de servicios omite heartbeats y sondeos repetitivos de salud. El registro completo se conserva con los diagnósticos fuera de Git. Los registros no incluyen secretos. Los IDs de trabajos no son credenciales.

Parada de la aplicación conservando datos:

```powershell
docker compose -p eas-tb2-integracion -f go/plataforma/compose.yml down
```

## Imágenes para el informe

Las 12 imágenes de `validacion-entrega/capturas/` son capturas de la aplicación normal; ninguna presenta fixtures. Para títulos y notas por imagen consultar `docs/correcciones-informe-tb2.md`, tabla de capturas. Secuencia de reproducción: 01 Inicio; 02 radio Una hora; 03 radio 24 horas; 04 Continuar a ejecuciones; 05 Iniciar ejecución y esperar resultado; 06 abrir enlace; 06b desplazar a gráficos; 07 navegar al Panel; 07b desplazar al trabajo y tiempo; 08 abrir Análisis ambiental, introducir job_id y seleccionar 24 horas; 08b desplazar a barras; 09 volver a Inicio con viewport móvil. Resultados y Análisis ambiental conservan las implementaciones del equipo y no se atribuyen como trabajo exclusivo de Alexandra.

## Diagnóstico previo y límites

El primer recorrido detectó que `.button` anulaba visualmente el atributo `hidden` del enlace a Resultados y permitía abrirlo antes de terminar. Se corrigió con `[hidden]` y se cubrió con una prueba integrada. Otro intento del comprobador buscaba un SVG en la vista ambiental, cuyo gráfico existente son barras HTML; se corrigió el selector del script, sin reemplazar la vista. Los intentos intermedios y capturas repetidas se conservaron fuera de Git en la carpeta de trabajo, no como evidencia final aprobada.

No se ejecutó una nueva comparación de rendimiento, una campaña de carga, una auditoría WCAG completa ni un lector de pantalla. El recorrido real creó un trabajo de 24 horas; los errores pendiente/fallido/inexistente y cancelación se verificaron con fixtures. No se indujeron fallos de MongoDB, Redis o nodos en el despliegue real de esta campaña. El panel muestra referencias de sesión, no todos los trabajos existentes. La predicción temporal y el enlace del video siguen pendientes; no se modificó Google Docs.
