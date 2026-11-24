package ui

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSPARoutes(t *testing.T) {
	h, err := NewHandler("http://127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/", "/panel", "/panel/", "/datos", "/datos/", "/ejecuciones", "/resultados", "/analisis-ambiental"} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
			if w.Code != 200 || !strings.Contains(w.Body.String(), "<html lang=\"es\">") {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
	for _, path := range []string{"/inexistente", "/assets/missing.js", "/.git/config"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 404 {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/datos", nil))
	if w.Code != 405 {
		t.Fatal(w.Code)
	}
}
func TestConfigAndProxy(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/base/jobs/abc" || r.URL.RawQuery != "detail=1" || r.Method != "POST" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "payload" {
			t.Error("body lost")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(409)
		_, _ = w.Write([]byte(`{"estado":"queued"}`))
	}))
	defer upstream.Close()
	h, err := NewHandler(upstream.URL + "/base")
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/ui-config.json", nil))
	var config map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &config); err != nil || config["apiBaseURL"] != "/api" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal(w.Body.String())
	}
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("POST", "/api/jobs/abc?detail=1", strings.NewReader("payload")))
	if w.Code != 409 || w.Body.String() != `{"estado":"queued"}` {
		t.Fatal(w.Code, w.Body.String())
	}
	upstream.Close()
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/api/health", nil))
	if w.Code != 502 {
		t.Fatal(w.Code)
	}
}
func TestRejectInvalidUpstream(t *testing.T) {
	for _, target := range []string{"", "file:///tmp/a", "http://user:secret@localhost:8080", "http://localhost:8080?x=1"} {
		if _, err := NewHandler(target); err == nil {
			t.Errorf("accepted %q", target)
		}
	}
}

func TestAssetsRevalidateAndKeepContentTypes(t *testing.T) {
	h, err := NewHandler("http://127.0.0.1:8080")
	if err != nil {
		t.Fatal(err)
	}
	for path, mime := range map[string]string{"/assets/app.js": "javascript", "/assets/styles.css": "text/css"} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 || w.Header().Get("Cache-Control") != "no-cache" || !strings.Contains(w.Header().Get("Content-Type"), mime) {
			t.Fatalf("%s: %d %v", path, w.Code, w.Header())
		}
	}
}
