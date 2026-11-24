# Interfaz web: base, inicio y navegación

SPA en HTML, CSS y módulos JavaScript nativos, servida por `cmd/ui` dentro del módulo Go existente. Los recursos se incluyen con `go:embed`: no requiere Node para ejecutar, compilación de frontend, CDN ni otro backend. Node solo ejecuta las pruebas de navegador sin dependencias.

La revisión de `develop` no encontró una UI previa ni un enunciado TB2 adicional con restricciones de framework. Esta elección mantiene Go, evita dependencias nuevas y permite compartir el menú, las vistas y los estilos. No cambia API, TCP, MongoDB, Redis ni las pantallas de otras integrantes.

## Ejecutar

Desde la raíz del repositorio, PowerShell:

```powershell
cd go/plataforma
$env:UI_API_URL = "http://127.0.0.1:8080"
$env:UI_ADDR = "127.0.0.1:8090"
go run ./cmd/ui
```

Abrir `http://127.0.0.1:8090/`. El inicio funciona sin bases de datos, nodos ni API. Para utilizar servicios reales, iniciar en otra terminal la API existente según [su documentación](../../../../docs/pc4/api.md), sin cambiar sus variables ni sus credenciales. También puede ejecutarse con `go build -o ui.exe ./cmd/ui` y `./ui.exe`; el ejecutable local está ignorado por Git.

`UI_ADDR` vale `127.0.0.1:8090` por defecto. `UI_API_URL` vale `http://127.0.0.1:8080` y solo admite HTTP(S), sin credenciales, query ni fragmento. Si el despliegue utiliza un prefijo, puede incluirse en esa URL.

El navegador obtiene `{ "apiBaseURL": "/api" }` de `/ui-config.json`. El servidor Go reenvía `/api/*` a la API configurada, conservando método, cuerpo, consulta y respuesta. Es un proxy de mismo origen para el servicio existente; no añade operaciones de negocio ni requiere relajar CORS. No configurar el destino hacia el propio servidor UI. La configuración no contiene secretos. Este arranque local no sustituye autenticación, TLS ni el despliegue de la API.

## Rutas y componentes compartidos

- `/`: inicio, propósito, fuente OEFA y recorrido de tres pasos.
- `/datos`, `/ejecuciones`, `/resultados`, `/analisis-ambiental`: estados vacíos explícitos mientras se incorporan las pantallas correspondientes. No se muestran métricas ni resultados ficticios.
- `web/assets/routes.js`: catálogo de rutas y reglas de navegación interna. History API y `popstate` mantienen Atrás/Adelante; el servidor resuelve las rutas directas y sus actualizaciones.
- `web/assets/app.js`: shell, menú adaptable y registro de vistas. Para incorporar una pantalla, exportar su función desde `assets/views/`, importarla en `app.js` y llamar a `registerView('/ruta', funcion)`. La función devuelve un nodo DOM. Conservar el título `id="titulo-vista" tabindex="-1"` para gestionar el foco. Si se añade una ruta nueva o parametrizada, extender también el catálogo y el handler Go, junto con sus pruebas.
- `states.js`: `loadingState`, `emptyState`, `errorState`, con roles accesibles y reintento opcional. Los mensajes remotos se insertan con `textContent`.
- `api.js`: configuración compartida y `apiRequest('/health')`, que devuelve la respuesta HTTP sin inventar esquemas. Cada pantalla debe manejar los errores, validar su contrato y pasar `AbortSignal` al desmontarse si mantiene consultas pendientes.
- `styles.css`: tokens y componentes `.button`, `.card`, `.form-field`, `.field-help`, `.table-wrap`, `.data-table`, `.message`. Formularios: asociar siempre `label for` con `input id`; errores con `aria-invalid` y `aria-describedby`. Tablas: usar `caption`, encabezados con `scope` y un contenedor con nombre accesible si requiere desplazamiento.

La navegación no intercepta enlaces externos, anclas, descargas, pestañas nuevas ni clics con modificadores. Hay salto al contenido, foco visible, título y enlace actual por ruta. El menú móvil admite Enter, Escape y navegación con Tab; al cambiar de vista, el foco pasa al título. Se respeta movimiento reducido y colores forzados. Los colores principales de texto, botones y foco utilizan fondos de alto contraste; esto no constituye una auditoría WCAG completa.

## Contrato real y dependencias

Se revisaron `internal/api/http.go`, `modelo.go`, `datos.go`, `openapi.yaml`, `cmd/api/main.go`, los módulos Go y `docs/pc4/api.md`.

Las operaciones de negocio existentes son `GET /health`, `POST /jobs`, `GET /jobs/{id}` y `GET /jobs/{id}/results`. La API también publica `/docs` y `/openapi.yaml`. Actualmente no hay endpoint para catálogo de datasets ni listado de trabajos: las tareas de consulta y panel deberán documentar o resolver esas dependencias, sin inventar datos. Esta entrega no realiza consultas de negocio al abrir el inicio.

Las concentraciones son PM10 y PM2.5 en µg/m³. Los periodos de una hora y 24 horas se tratan por separado. No se descargan CSV masivos, no hay carga de archivos y K-Means no se presenta como pronóstico ni clasificación clínica.

## Verificar

Desde `go/plataforma`:

```powershell
go build ./cmd/ui
go vet ./...
go test -count=1 ./...
go test -race -count=1 ./internal/ui
node --test internal/ui/tests/navigation.test.mjs
```

Las pruebas Go cubren rutas directas, método no permitido, recursos inexistentes, configuración, rechazo de destinos inválidos, conservación del contrato a través del proxy y respuesta 502. El servidor upstream de la prueba es un doble local de HTTP, no evidencia de una ejecución científica ni de bases reales. Las pruebas JavaScript verifican resolución de rutas, preservación de interacciones propias del navegador, reintento y configuración compartida y conservación de errores HTTP.

Comprobación manual: abrir cada enlace, usar Atrás, abrir directamente `/resultados`, actualizar la página, navegar con Tab/Enter, y abrir/cerrar el menú con Enter/Escape a menos de 700 px. La API puede permanecer apagada: el inicio debe funcionar y `/api/health` devolverá 502, sin simular salud correcta.

## Evidencias conservadas

En `evidencias/`:

- `inicio-escritorio.jpg`: captura real del inicio con fuente, periodos y recorrido.
- `inicio-movil.jpg`: captura real de la composición en una columna y menú compacto.
- `datos-movil.jpg`: estado vacío real, sin catálogo inventado.
- `http.json`: códigos, tipos y tamaños observados en el servidor Go real; API upstream no iniciada. No contiene respuestas científicas ni credenciales.

Revisión en navegador realizada el 23 de noviembre de 2026: cinco enlaces, Atrás, actualización de `/datos`, apertura directa y actualización de `/resultados`, foco en el título y enlace activo; menú móvil con Enter/Escape y cierre al navegar. Viewport solicitado 390 × 844; ancho útil observado 375 px, igual al ancho desplazable, sin desbordamiento horizontal. Se revisó visualmente el inicio de escritorio y móvil. Las capturas no prueban datos reales ni ejecución del algoritmo.

Resultado automatizado: `go build ./cmd/ui`, `go vet ./...` y `go test -count=1 ./...` finalizaron correctamente; pasaron las tres pruebas JavaScript. `go test -race -count=1 ./internal/ui` no pudo ejecutarse porque el entorno tiene CGO deshabilitado. Para esa comprobación se necesita CGO y un compilador C compatible; no se considera aprobada.

Las pruebas no incluyen lector de pantalla, auditoría WCAG completa, despliegue público ni una sesión real con MongoDB/Redis. No se ejecutaron cálculos K-Means para esta tarea. Un bloqueo inicial de Windows al binario temporal de pruebas no cambió las políticas del equipo; la ejecución posterior del módulo completo sí pudo completarse.

Esta rama entrega solo la base. No hacer merge automático. Iniciar consulta de datasets únicamente cuando este PR esté integrado en `develop` y se indique continuar; crear entonces `feature/ui-datasets` desde `develop` actualizado.
