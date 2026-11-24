//go:build ignore

// Browser fixtures only; never built or embedded into the production UI.
package main

import (
	"eas-aire/plataforma/internal/ui"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

//go:embed panel-fixture-loader.js
var panelFixture string

func main() {
	handler, err := ui.NewHandler("http://127.0.0.1:8087")
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/api/", handler)
	mux.Handle("GET /assets/", handler)
	mux.HandleFunc("GET /panel-fixture-loader.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(panelFixture))
	})
	mux.HandleFunc("GET /api/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		state := strings.TrimPrefix(id, "fixture-")
		if state != "queued" && state != "running" && state != "completed" && state != "failed" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"trabajo": map[string]any{"job_id": id, "estado": state, "actualizado": "2026-11-23T12:00:00Z"}, "fuente": "fixture"})
	})
	mux.HandleFunc("GET /api/jobs/{id}/results", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("id") != "fixture-completed" {
			http.Error(w, "Fixture sin resultado", 409)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"job_id": "fixture-completed", "resultado": map[string]any{"segundos_ml": 0.25}})
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ui-config.json" {
			handler.ServeHTTP(w, r)
			return
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, r)
		if response.Code != 200 {
			http.NotFound(w, r)
			return
		}
		for key, values := range response.Header() {
			w.Header()[key] = values
		}
		page := strings.Replace(response.Body.String(), `<script type="module"`, `<script type="module" src="/panel-fixture-loader.js"></script><script type="module"`, 1)
		page = strings.Replace(page, `<body>`, `<body><aside class="message" role="note">ENTORNO DE PRUEBA: catálogo y trabajos sintéticos; no son resultados científicos.</aside>`, 1)
		_, _ = w.Write([]byte(page))
	})
	log.Print("Panel con fixtures: http://127.0.0.1:8095/panel?fixture=full")
	log.Fatal(http.ListenAndServe("127.0.0.1:8095", mux))
}
