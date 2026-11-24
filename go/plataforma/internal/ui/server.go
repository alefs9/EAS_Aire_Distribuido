// Package ui serves the browser interface without changing the API contract.
package ui

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

//go:embed web
var content embed.FS

// NewHandler embeds the SPA and forwards /api/* to the configured existing API.
// The upstream is server configuration, never a browser-supplied destination.
func NewHandler(upstream string) (http.Handler, error) {
	target, err := url.Parse(upstream)
	if err != nil || target == nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") || target.User != nil || target.RawQuery != "" || target.Fragment != "" {
		return nil, fmt.Errorf("UI_API_URL debe ser una URL HTTP(S) sin credenciales, query ni fragmento")
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = &http.Transport{Proxy: http.ProxyFromEnvironment, DialContext: (&net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}).DialContext, ResponseHeaderTimeout: 15 * time.Second, IdleConnTimeout: 60 * time.Second}
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, _ error) {
		http.Error(w, "El servicio no está disponible. Intenta nuevamente más tarde.", http.StatusBadGateway)
	}
	assets, _ := fs.Sub(content, "web/assets")
	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", proxy))
	mux.HandleFunc("GET /ui-config.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{"apiBaseURL": "/api"})
	})
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}
		path := strings.TrimSuffix(r.URL.Path, "/")
		switch path {
		case "", "/datos", "/ejecuciones", "/resultados", "/analisis-ambiental":
			page, _ := content.ReadFile("web/index.html")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-cache")
			_, _ = w.Write(page)
		default:
			http.NotFound(w, r)
		}
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'")
		mux.ServeHTTP(w, r)
	}), nil
}
