package api

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// La especificación debe documentar todas las rutas que registra el router.
func TestDocumentacionPublicada(t *testing.T) {
	ts := servidorPrueba(t, ejecutorFalso{}, 1)
	resp, err := http.Get(ts.URL + "/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	cuerpo, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/openapi.yaml: %d", resp.StatusCode)
	}
	for _, ruta := range []string{"/health:", "/jobs:", "/jobs/{id}:", "/jobs/{id}/results:"} {
		if !strings.Contains(string(cuerpo), ruta) {
			t.Errorf("la especificación no documenta %s", ruta)
		}
	}
	if r, err := http.Get(ts.URL + "/docs"); err != nil || r.StatusCode != http.StatusOK {
		t.Fatalf("/docs no disponible: %v", err)
	}
}
