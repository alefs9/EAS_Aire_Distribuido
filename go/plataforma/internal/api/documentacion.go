package api

import (
	_ "embed"
	"net/http"
)

// especificacion es el contrato OpenAPI, embebido en el binario.
//
//go:embed openapi.yaml
var especificacion []byte

const paginaSwagger = `<!doctype html>
<html lang="es"><head><meta charset="utf-8"><title>EAS Aire Distribuido - API</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head>
<body><div id="swagger"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({ url: "/openapi.yaml", dom_id: "#swagger" });</script>
</body></html>`

// registrarDocumentacion publica la especificación y la interfaz Swagger UI.
func registrarDocumentacion(mux *http.ServeMux) {
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		_, _ = w.Write(especificacion)
	})
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(paginaSwagger))
	})
}
