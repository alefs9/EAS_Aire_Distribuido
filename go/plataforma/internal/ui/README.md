# Interfaz web e integración TB2

SPA de HTML, CSS y módulos JavaScript nativos, incluida mediante `go:embed` y servida por Go. No necesita Node ni compilación de frontend para ejecutarse. Esta elección mantiene el servidor existente y los componentes compartidos, sin otro backend ni CDN. Node y Playwright se usan solo para pruebas; los recursos de prueba quedan fuera de `web` para no incorporarlos al ejecutable.

## Inicio

Desde la raíz, usar Docker según el README principal. Para ejecutar únicamente la UI contra una API ya disponible:

```powershell
cd go/plataforma
$env:UI_API_URL='http://127.0.0.1:8080'
$env:UI_ADDR='127.0.0.1:8090'
go run ./cmd/ui
```

Abrir http://127.0.0.1:8090/. El inicio funciona sin API. `UI_API_URL` admite HTTP(S), sin credenciales, consulta ni fragmento, y puede contener un prefijo. `/ui-config.json` publica solo `/api`; el proxy Go de mismo origen conserva el contrato del servicio sin relajar CORS. La URL nunca debe apuntar al propio servidor UI. El puerto debe estar libre; no iniciar otra UI sobre la instancia de Compose.

## Rutas y contratos existentes

- `/`: propósito, fuente OEFA y recorrido.
- `/datos`: `GET /datasets`, conjuntos de una hora y 24 horas, metadatos presentes y selección verificada. No descarga CSV masivos.
- `/ejecuciones`: vuelve a validar el conjunto, crea mediante `POST /jobs` y consulta `GET /jobs/{id}`. Seguimiento con backoff, hasta cinco errores consecutivos y diez minutos; navegación cancela las lecturas, no el entrenamiento remoto.
- `/resultados?job_id=...`: consulta `GET /jobs/{id}/results`, muestra detalle y gráficos existentes, o estados pendiente, fallido e inexistente. Admite consulta manual, entrada directa, recarga y cambio de identificador en SPA. Cambiar el campo elimina la visualización anterior. Parámetros duplicados o inválidos se rechazan sin reutilizar otro trabajo.
- `/panel`: consulta salud, catálogo y referencias conocidas de la sesión. `session-jobs.js` conserva solo IDs válidos de respuestas POST 202, máximo 50 sin duplicados. No persiste estados ni resultados. Cada actualización consulta hasta 10 referencias con concurrencia tres; las ausentes o fallidas no cuentan como confirmadas. Si el navegador impide guardar sesión, utiliza memoria de esta página. No hay listado global ni actualización push: usar la actualización de la vista.
- `/analisis-ambiental`: consulta los trabajos introducidos y conserva la visualización ambiental existente; no predice concentraciones futuras.

El panel configura sus proveedores reales en `app.js`, sin endpoints adicionales. `segundos_ml` es entrenamiento y no equivale al tiempo HTTP. PM10 y PM2.5 se expresan en µg/m³ cuando la procedencia está verificada; no se mezclan periodos ni se infieren unidades de un dataset desconocido.

`app.js` importa y registra las vistas. `routes.js` resuelve rutas y `view-host.js` contiene errores y cancela el contexto anterior; cada vista recibe `{path,title,search,signal}`. Enlaces internos conservan query, Atrás y Adelante. Un fallo de renderizado permite volver a otras rutas. Un error de sintaxis/importación del módulo principal sigue siendo un error de arranque y debe detectarse con la prueba integrada.

`states.js`, `styles.css`, `menu.js` y los componentes existentes comparten estados, formularios y navegación. Conservar etiquetas asociadas, título enfocable, roles, foco visible y `AbortSignal`. La regla `[hidden]` evita que los estilos de componentes revelen controles todavía no disponibles. No se afirma una auditoría WCAG completa.

## Pruebas reproducibles

Desde `go/plataforma`:

```powershell
node --test internal/ui/tests/*.test.mjs
go test -count=1 ./...
go vet ./...
go test -race -count=1 ./...
npm ci --prefix internal/ui/tests/browser
$env:PLAYWRIGHT_CHANNEL='msedge'
npm --prefix internal/ui/tests/browser test
```

Playwright usa Edge instalado; alternativamente instalar Chromium con `npx playwright install chromium` desde `internal/ui/tests/browser` y no configurar `PLAYWRIGHT_CHANNEL`. Node 20 o posterior es necesario para esta dependencia de pruebas. El servidor sintético arranca en 8098 y muestra una advertencia explícita: comprueba los módulos normales, rutas, gráficos, query, cancelación, estados y referencias, pero no demuestra integración con bases ni entrenamiento real.

Si Windows bloquea los ejecutables temporales de Go, utilizar Docker sin desactivar sus políticas. Desde la **raíz del repositorio**:

```powershell
docker run --rm --mount "type=bind,source=$($PWD.Path),target=/src,readonly" --mount type=volume,source=eas-tb2-go-cache,target=/root/.cache/go-build --mount type=volume,source=eas-tb2-go-mod,target=/go/pkg/mod -w /src/go/plataforma golang:1.27-bookworm sh -c 'go test -count=1 ./... && go vet ./... && go test -race -count=1 ./...'
docker run -d --name eas-tb2-fixtures -p 127.0.0.1:8098:8098 -e TB2_FIXTURE_ADDR=:8098 --mount "type=bind,source=$($PWD.Path),target=/src,readonly" --mount type=volume,source=eas-tb2-go-cache,target=/root/.cache/go-build --mount type=volume,source=eas-tb2-go-mod,target=/go/pkg/mod -w /src/go/plataforma golang:1.27-bookworm go run internal/ui/tests/integration_browser.go
$env:TB2_FIXTURE_URL='http://127.0.0.1:8098'
$env:PLAYWRIGHT_CHANNEL='msedge'
npm --prefix go/plataforma/internal/ui/tests/browser test
docker stop eas-tb2-fixtures
docker rm eas-tb2-fixtures
Remove-Item Env:TB2_FIXTURE_URL
```

Esperar a que `docker logs eas-tb2-fixtures` indique que escucha antes de lanzar Playwright. No utilizar un servicio de producción como `TB2_FIXTURE_URL`.

Para el recorrido **real** con Compose ya iniciado y una carpeta nueva de evidencias, desde la raíz:

```powershell
$env:PLAYWRIGHT_CHANNEL='msedge'
node go/plataforma/internal/ui/tests/browser/real-compose.mjs results/evaluacion-web/nueva-campana
```

Este comando crea un trabajo real de 24 horas (k=4, semilla=42, tolerancia=0.0001, máximo 100 iteraciones, dos nodos), guarda respuestas y capturas y exige el mismo ID en Ejecuciones, Resultados y Panel. Incluye verificación de recarga y vista ambiental. Usa una sesión limpia y publica resultados solo si pasan las comprobaciones. Puede configurarse `TB2_UI_URL` para otra URL local. El nombre de carpeta debe ser nuevo para preservar campañas anteriores.

## Evidencias y alcance

La campaña de integración del 27/11/2026 está en `results/evaluacion-web/2026-11-27-integracion-web-tb2-110325`, desde la raíz. Su nota separa pruebas sintéticas y servicios reales e identifica revisiones y límites. Las capturas de Resultados y Análisis ambiental muestran la integración de las vistas existentes; no deben atribuirse como desarrollo exclusivo de Alexandra.

`evidencias/` conserva la revisión histórica de la base del 23/11/2026, realizada sin API. Sus pantallas vacías y conteos de pruebas describen ese momento; no sustituyen el estado actual. Las campañas históricas de datasets, panel y despliegue también se conservan, sin presentar sus mediciones como nuevas. No se certifican despliegue público, autenticación, TLS, auditoría integral de accesibilidad ni predicción temporal.
