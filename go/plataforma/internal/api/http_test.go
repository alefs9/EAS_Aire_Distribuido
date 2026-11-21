package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"eas-aire/plataforma"
)

// ejecutorFalso responde como el coordinador sin abrir conexiones TCP.
type ejecutorFalso struct {
	espera time.Duration
	err    error
}

func (e ejecutorFalso) Run(ctx context.Context, c plataforma.ExecutionConfig) (plataforma.Result, error) {
	select {
	case <-time.After(e.espera):
	case <-ctx.Done():
		return plataforma.Result{}, ctx.Err()
	}
	if e.err != nil {
		return plataforma.Result{Status: "failed"}, e.err
	}
	return plataforma.Result{ID: "falso", Rows: 4, Iterations: 2, Status: "completed", Reason: "converged",
		Nodes: []string{"nodo-1"}, Centroids: []plataforma.Point{{1, 1}, {11, 11}}, Counts: []int{2, 2}}, nil
}

func datasetPrueba() map[string]Dataset {
	return map[string]Dataset{"1_hora": {ID: "abc:1_hora", Periodo: "1_hora", Ruta: "x.csv",
		Puntos: []plataforma.Point{{0, 0}, {2, 2}, {10, 10}, {12, 12}}}}
}

func servidorPrueba(t *testing.T, e Ejecutor, cola int) *httptest.Server {
	t.Helper()
	logger := log.New(io.Discard, "", 0)
	s := NuevoServicio(NuevaMemoria(), e, datasetPrueba(), ConfigServicio{TamanoCola: cola, Timeout: 2 * time.Second, NodosDefecto: 1}, logger)
	ts := httptest.NewServer(NuevoRouter(s, logger))
	t.Cleanup(func() { ts.Close(); s.Cerrar() })
	return ts
}

func pedir(t *testing.T, metodo, url, cuerpo string) (int, map[string]any) {
	t.Helper()
	req, _ := http.NewRequest(metodo, url, strings.NewReader(cuerpo))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var m map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&m)
	return resp.StatusCode, m
}

func esperarEstado(t *testing.T, url, id, estado string) {
	t.Helper()
	limite := time.Now().Add(3 * time.Second)
	for time.Now().Before(limite) {
		_, m := pedir(t, "GET", url+"/jobs/"+id, "")
		if tr, ok := m["trabajo"].(map[string]any); ok && tr["estado"] == estado {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("el trabajo %s no llegó al estado %s", id, estado)
}

func TestFlujoCompleto(t *testing.T) {
	ts := servidorPrueba(t, ejecutorFalso{espera: 100 * time.Millisecond}, 4)
	codigo, tr := pedir(t, "POST", ts.URL+"/jobs", `{"periodo":"1_hora","k":2}`)
	if codigo != http.StatusAccepted || tr["estado"] != "queued" {
		t.Fatalf("POST /jobs: %d %v", codigo, tr)
	}
	id := tr["job_id"].(string)
	if codigo, _ := pedir(t, "GET", ts.URL+"/jobs/"+id+"/results", ""); codigo != http.StatusConflict {
		t.Fatalf("resultado pendiente: esperaba 409, obtuve %d", codigo)
	}
	esperarEstado(t, ts.URL, id, "completed")
	codigo, m := pedir(t, "GET", ts.URL+"/jobs/"+id+"/results", "")
	if codigo != http.StatusOK {
		t.Fatalf("resultado: esperaba 200, obtuve %d", codigo)
	}
	r := m["resultado"].(map[string]any)
	// Inercia con centroides (1,1) y (11,11): cuatro puntos a distancia² 2 => 8.
	if r["inercia"].(float64) != 8 || r["convergio"] != true {
		t.Fatalf("resultado inesperado: %v", r)
	}
}

func TestTrabajoFallido(t *testing.T) {
	ts := servidorPrueba(t, ejecutorFalso{err: errors.New("waiting for 2 nodes: deadline exceeded")}, 4)
	_, tr := pedir(t, "POST", ts.URL+"/jobs", `{}`)
	id := tr["job_id"].(string)
	esperarEstado(t, ts.URL, id, "failed")
	if codigo, _ := pedir(t, "GET", ts.URL+"/jobs/"+id+"/results", ""); codigo != http.StatusUnprocessableEntity {
		t.Fatalf("trabajo fallido: esperaba 422, obtuve %d", codigo)
	}
}

func TestParametrosInvalidos(t *testing.T) {
	ts := servidorPrueba(t, ejecutorFalso{}, 4)
	for _, cuerpo := range []string{`{"k":99}`, `{"periodo":"semanal"}`, `{"nodos":0.5}`, `{"campo":1}`, `no es json`} {
		if codigo, _ := pedir(t, "POST", ts.URL+"/jobs", cuerpo); codigo != http.StatusBadRequest {
			t.Errorf("%s: esperaba 400, obtuve %d", cuerpo, codigo)
		}
	}
	if codigo, _ := pedir(t, "POST", ts.URL+"/jobs", `{"periodo":"24_horas"}`); codigo != http.StatusBadRequest {
		t.Errorf("periodo sin CSV cargado: esperaba 400, obtuve %d", codigo)
	}
}

func TestNoEncontrado(t *testing.T) {
	ts := servidorPrueba(t, ejecutorFalso{}, 4)
	for _, ruta := range []string{"/jobs/no-existe", "/jobs/no-existe/results"} {
		if codigo, _ := pedir(t, "GET", ts.URL+ruta, ""); codigo != http.StatusNotFound {
			t.Errorf("%s: esperaba 404, obtuve %d", ruta, codigo)
		}
	}
}

func TestColaLlena(t *testing.T) {
	ts := servidorPrueba(t, ejecutorFalso{espera: 500 * time.Millisecond}, 1)
	vistos503 := 0
	for i := 0; i < 5; i++ {
		if codigo, _ := pedir(t, "POST", ts.URL+"/jobs", `{}`); codigo == http.StatusServiceUnavailable {
			vistos503++
		}
	}
	if vistos503 == 0 {
		t.Fatal("esperaba 503 al llenar la cola")
	}
}

func TestSalud(t *testing.T) {
	ts := servidorPrueba(t, ejecutorFalso{}, 1)
	if codigo, m := pedir(t, "GET", ts.URL+"/health", ""); codigo != http.StatusOK || m["estado"] != "ok" {
		t.Fatalf("health: %d %v", codigo, m)
	}
}
