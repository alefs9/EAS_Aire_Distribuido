package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"eas-aire/plataforma/internal/persistencia"
)

// NuevoRouter registra las rutas con los patrones de net/http de Go 1.22.
func NuevoRouter(s *Servicio, logger *log.Logger) *http.ServeMux {
	h := &handlers{s: s}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.salud)
	mux.HandleFunc("POST /jobs", h.crear)
	mux.HandleFunc("GET /jobs/{id}", h.estado)
	mux.HandleFunc("GET /jobs/{id}/results", h.resultado)
	mux.HandleFunc("GET /datasets", h.datasets)
	registrarDocumentacion(mux)
	return mux
}

type handlers struct{ s *Servicio }

func responderJSON(w http.ResponseWriter, codigo int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	_ = json.NewEncoder(w).Encode(v)
}

func responderError(w http.ResponseWriter, codigo int, clave, mensaje string) {
	responderJSON(w, codigo, map[string]string{"codigo": clave, "error": mensaje})
}

// contexto limita cada consulta a las bases de datos para responder a tiempo.
func contexto(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 2*time.Second)
}

func (h *handlers) salud(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()
	deps := h.s.Salud(ctx)
	estado, codigo := "ok", http.StatusOK
	if deps["mongodb"] == "no disponible" {
		estado, codigo = "no disponible", http.StatusServiceUnavailable
	} else if deps["redis"] == "no disponible" {
		estado = "degradado" // funciona leyendo de MongoDB
	}
	responderJSON(w, codigo, map[string]any{"estado": estado, "dependencias": deps})
}

func (h *handlers) crear(w http.ResponseWriter, r *http.Request) {
	var sol Solicitud
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<14))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&sol); err != nil {
		responderError(w, http.StatusBadRequest, "json_invalido", err.Error())
		return
	}
	ctx, cancel := contexto(r)
	defer cancel()
	t, err := h.s.Enviar(ctx, sol)
	var ve ErrorValidacion
	switch {
	case errors.As(err, &ve):
		responderError(w, http.StatusBadRequest, "parametros_invalidos", ve.Mensaje)
	case errors.Is(err, ErrPeriodoNoActivo):
		responderError(w, http.StatusBadRequest, "periodo_no_disponible", err.Error())
	case errors.Is(err, ErrColaLlena):
		responderError(w, http.StatusServiceUnavailable, "cola_llena", err.Error())
	case err != nil:
		responderError(w, http.StatusServiceUnavailable, "almacen_no_disponible", err.Error())
	default:
		w.Header().Set("Location", "/jobs/"+t.ID)
		responderJSON(w, http.StatusAccepted, t)
	}
}

// leer resuelve los errores comunes; devuelve false si ya respondió.
func (h *handlers) leer(w http.ResponseWriter, r *http.Request) (persistencia.Lectura, bool, bool) {
	ctx, cancel := contexto(r)
	defer cancel()
	l, err := h.s.Leer(ctx, r.PathValue("id"))
	switch {
	case errors.Is(err, persistencia.ErrNoEncontrado):
		responderError(w, http.StatusNotFound, "no_encontrado", "no existe un trabajo con ese identificador")
		return l, false, false
	case errors.Is(err, persistencia.ErrPendiente):
		return l, false, true
	case err != nil:
		responderError(w, http.StatusServiceUnavailable, "almacen_no_disponible", err.Error())
		return l, false, false
	}
	return l, true, true
}

func (h *handlers) estado(w http.ResponseWriter, r *http.Request) {
	l, _, ok := h.leer(w, r)
	if !ok {
		return
	}
	t := l.Trabajo
	t.Resultado = nil // el resultado se consulta en /results
	responderJSON(w, http.StatusOK, map[string]any{"trabajo": t, "fuente": l.Fuente})
}

func (h *handlers) resultado(w http.ResponseWriter, r *http.Request) {
	l, completo, ok := h.leer(w, r)
	if !ok {
		return
	}
	switch {
	case completo:
		cuerpo := map[string]any{"job_id": l.Trabajo.ID, "dataset_id": l.Trabajo.DatasetID,
			"parametros": l.Trabajo.Parametros, "resultado": l.Trabajo.Resultado, "fuente": l.Fuente}
		if l.Aviso != "" {
			cuerpo["aviso"] = l.Aviso
		}
		responderJSON(w, http.StatusOK, cuerpo)
	case l.Trabajo.Estado == "failed":
		responderJSON(w, http.StatusUnprocessableEntity, map[string]string{"codigo": "trabajo_fallido", "job_id": l.Trabajo.ID, "error": l.Trabajo.Error})
	default: // queued o running
		responderJSON(w, http.StatusConflict, map[string]string{"codigo": "no_disponible_aun", "job_id": l.Trabajo.ID, "estado": l.Trabajo.Estado})
	}
}

func (h *handlers) datasets(w http.ResponseWriter, _ *http.Request) {
	if h.s.datasets == nil { // nil = catálogo no disponible; mapa vacío no nil = catálogo confirmado vacío.
		responderError(w, http.StatusServiceUnavailable, "catalogo_no_disponible", "El catálogo no está disponible temporalmente.")
		return
	}
	periodos := []string{"1_hora", "24_horas"}
	items := make([]map[string]any, 0, len(h.s.datasets))
	for _, periodo := range periodos {
		d, ok := h.s.datasets[periodo]
		if !ok {
			continue
		}
		item := map[string]any{
			"id": d.ID, "period": d.Periodo,
			"vector_count":      len(d.Puntos),
			"measurement_count": len(d.Puntos) * 2,
		}
		if d.SHA256 != "" {
			item["sha256"] = d.SHA256
		}
		items = append(items, item)
	}
	w.Header().Set("Cache-Control", "no-store")
	responderJSON(w, http.StatusOK, map[string]any{"datasets": items})
}

// ConRegistro escribe método, ruta, código y duración de cada solicitud.
func ConRegistro(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		sw := &escritorEstado{ResponseWriter: w, codigo: http.StatusOK}
		next.ServeHTTP(sw, r)
		logger.Printf("%s %s %d %.2fms", r.Method, r.URL.Path, sw.codigo, float64(time.Since(inicio).Microseconds())/1000)
	})
}

type escritorEstado struct {
	http.ResponseWriter
	codigo int
}

func (e *escritorEstado) WriteHeader(c int) { e.codigo = c; e.ResponseWriter.WriteHeader(c) }
