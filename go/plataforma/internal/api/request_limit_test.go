package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJobLimiterDevuelve429YRetryAfter(t *testing.T) {
	llamadas := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		llamadas++
		w.WriteHeader(http.StatusAccepted)
	})

	handler := limitJobCreation(next, 1)

	primera := httptest.NewRequest(http.MethodPost, "/jobs", nil)
	primera.RemoteAddr = "127.0.0.1:1234"
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, primera)

	if rec1.Code != http.StatusAccepted {
		t.Fatalf("primera solicitud: esperaba 202; obtuvo %d", rec1.Code)
	}

	segunda := httptest.NewRequest(http.MethodPost, "/jobs", nil)
	segunda.RemoteAddr = "127.0.0.1:1234"
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, segunda)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("segunda solicitud: esperaba 429; obtuvo %d", rec2.Code)
	}
	if rec2.Header().Get("Retry-After") != "60" {
		t.Fatalf("Retry-After incorrecto: %q", rec2.Header().Get("Retry-After"))
	}
	if llamadas != 1 {
		t.Fatalf("el handler pasó %d solicitudes; esperaba 1", llamadas)
	}
}

func TestJobLimiterCeroDesactivaElLimite(t *testing.T) {
	llamadas := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		llamadas++
		w.WriteHeader(http.StatusAccepted)
	})

	handler := limitJobCreation(next, 0)

	for i := 0; i < 12; i++ {
		req := httptest.NewRequest(http.MethodPost, "/jobs", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("solicitud %d: esperaba 202; obtuvo %d", i+1, rec.Code)
		}
	}

	if llamadas != 12 {
		t.Fatalf("se procesaron %d solicitudes; esperaba 12", llamadas)
	}
}

func TestConfiguredJobLimit(t *testing.T) {
	t.Setenv("JOB_POSTS_PER_MINUTE", "0")
	if got := configuredJobLimit(); got != 0 {
		t.Fatalf("con 0: esperaba 0; obtuvo %d", got)
	}

	t.Setenv("JOB_POSTS_PER_MINUTE", "7")
	if got := configuredJobLimit(); got != 7 {
		t.Fatalf("con 7: esperaba 7; obtuvo %d", got)
	}

	t.Setenv("JOB_POSTS_PER_MINUTE", "-1")
	if got := configuredJobLimit(); got != 10 {
		t.Fatalf("valor inválido: esperaba default 10; obtuvo %d", got)
	}
}
