//go:build ignore

// Test-only browser fixture. Not compiled into cmd/ui or embedded production assets.
package main

import (
	"eas-aire/plataforma/internal/ui"
	_ "embed"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

//go:embed fixture-loader.js
var loader string

func main() {
	realUI, err := ui.NewHandler("http://127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /fixture-loader.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(loader))
	})
	mux.Handle("GET /assets/", realUI)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		response := httptest.NewRecorder()
		realUI.ServeHTTP(response, r)
		if response.Code != 200 {
			http.NotFound(w, r)
			return
		}
		for key, values := range response.Header() {
			w.Header()[key] = values
		}
		page := strings.Replace(response.Body.String(), `<script type="module"`, `<script type="module" src="/fixture-loader.js"></script><script type="module"`, 1)
		page = strings.Replace(page, `<body>`, `<body><aside class="message" role="note">ENTORNO DE PRUEBA: catálogo sintético, no datos reales del servicio.</aside>`, 1)
		_, _ = w.Write([]byte(page))
	})
	log.Print("Fixtures de navegador: http://127.0.0.1:8092/datos?fixture=two")
	log.Fatal(http.ListenAndServe("127.0.0.1:8092", mux))
}
