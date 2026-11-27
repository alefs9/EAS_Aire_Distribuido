//go:build ignore

// Synthetic HTTP fixtures for full-SPA regression tests; never part of cmd/ui.
package main

import (
	"eas-aire/plataforma/internal/ui"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"
)

func main() {
	handler, err := ui.NewHandler("http://127.0.0.1:1")
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	jsonResponse := func(w http.ResponseWriter, status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
	result := func(id string) any {
		centroids := [][]float64{{10, 5}, {20, 8}}
		if id == "fixture-second" {
			centroids = [][]float64{{30, 15}, {40, 18}}
		}
		return map[string]any{"job_id": id, "dataset_id": "fixture-hour-data", "parametros": map[string]any{"k": 2, "semilla": 42, "max_iter": 100, "tolerancia": 0.0001}, "resultado": map[string]any{"centroides": centroids, "conteos": []int{3, 5}, "iteraciones": 4, "inercia": 12, "convergio": true, "origen": "fixture-sintetico", "segundos_ml": 0.25}, "fuente": "fixture"}
	}
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, map[string]any{"estado": "ok", "dependencias": map[string]string{"memoria": "ok"}})
	})
	mux.HandleFunc("GET /api/datasets", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, 200, map[string]any{"datasets": []any{map[string]any{"id": "fixture-hour-data", "period": "1_hora", "vector_count": 8, "features": []string{"PM10", "PM2.5"}, "unit": "µg/m³", "sha256": strings.Repeat("a", 64)}}})
	})
	mux.HandleFunc("POST /api/jobs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/jobs/fixture-created")
		jsonResponse(w, 202, map[string]string{"job_id": "fixture-created", "estado": "completed"})
	})
	mux.HandleFunc("GET /api/jobs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "fixture-missing" {
			jsonResponse(w, 404, map[string]string{"codigo": "no_encontrado"})
			return
		}
		state := "completed"
		if id == "fixture-pending" {
			state = "running"
		}
		if id == "fixture-failed" {
			state = "failed"
		}
		jsonResponse(w, 200, map[string]any{"trabajo": map[string]any{"job_id": id, "estado": state, "dataset_id": "fixture-hour-data", "actualizado": "2026-11-27T12:00:00Z"}, "fuente": "fixture"})
	})
	mux.HandleFunc("GET /api/jobs/{id}/results", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		switch id {
		case "fixture-missing":
			jsonResponse(w, 404, map[string]string{"codigo": "no_encontrado"})
			return
		case "fixture-pending":
			jsonResponse(w, 409, map[string]string{"job_id": id, "estado": "running", "codigo": "no_disponible_aun"})
			return
		case "fixture-failed":
			jsonResponse(w, 422, map[string]string{"job_id": id, "error": "Fallo sintético de prueba", "codigo": "trabajo_fallido"})
			return
		case "fixture-slow":
			select {
			case <-r.Context().Done():
				return
			case <-time.After(3 * time.Second):
			}
		}
		jsonResponse(w, 200, result(id))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") || r.URL.Path == "/ui-config.json" {
			handler.ServeHTTP(w, r)
			return
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, r)
		for k, v := range response.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(response.Code)
		page := strings.Replace(response.Body.String(), "<body>", `<body><aside class="message" role="note">PRUEBAS DE REGRESIÓN: respuestas sintéticas, sin servicios reales.</aside>`, 1)
		_, _ = w.Write([]byte(page))
	})
	log.Print("Fixtures de integración: http://127.0.0.1:8098")
	addr := os.Getenv("TB2_FIXTURE_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8098"
	}
	log.Fatal(http.ListenAndServe(addr, mux))
}
