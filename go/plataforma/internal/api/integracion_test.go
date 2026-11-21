package api

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"eas-aire/plataforma"
	"eas-aire/plataforma/internal/cluster"
)

// TestIntegracionClusterTCP levanta el coordinador real y dos nodos reales por TCP
// en este proceso, crea un trabajo por HTTP y comprueba el resultado.
func TestIntegracionClusterTCP(t *testing.T) {
	dir := t.TempDir()
	csv := "PM10,PM2.5\n0,0\n0,2\n2,0\n2,2\n10,10\n10,12\n12,10\n12,12\n"
	if err := os.WriteFile(filepath.Join(dir, "features_1_hora.csv"), []byte(csv), 0o600); err != nil {
		t.Fatal(err)
	}
	d, err := CargarDataset(dir, "1_hora")
	if err != nil {
		t.Fatal(err)
	}

	logger := log.New(io.Discard, "", 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	coord := plataforma.NewCoordinator(5*time.Second, logger)
	servido := make(chan error, 1)
	go func() { servido <- coord.Serve(ctx, ln) }()
	nodos := make(chan error, 2)
	for _, id := range []string{"nodo-1", "nodo-2"} {
		go func(id string) {
			nodos <- cluster.RunNode(ctx, cluster.NodeConfig{Address: ln.Addr().String(), ID: id, Capacity: 1,
				Interval: 20 * time.Millisecond, Timeout: 5 * time.Second}, logger)
		}(id)
	}

	s := NuevoServicio(NuevaMemoria(), coord, map[string]Dataset{"1_hora": d},
		ConfigServicio{TamanoCola: 2, Timeout: 10 * time.Second, NodosDefecto: 2}, logger)
	ts := httptest.NewServer(NuevoRouter(s, logger))
	defer ts.Close()

	codigo, tr := pedir(t, "POST", ts.URL+"/jobs", `{"k":2,"semilla":1,"max_iter":20}`)
	if codigo != http.StatusAccepted {
		t.Fatalf("POST /jobs: %d %v", codigo, tr)
	}
	id := tr["job_id"].(string)
	esperarEstado(t, ts.URL, id, "completed")
	_, m := pedir(t, "GET", ts.URL+"/jobs/"+id+"/results", "")
	r := m["resultado"].(map[string]any)
	conteos := r["conteos"].([]any)
	if len(conteos) != 2 || conteos[0].(float64)+conteos[1].(float64) != 8 {
		t.Fatalf("los conteos no conservan las 8 filas: %v", r)
	}
	if r["inercia"].(float64) != 16 { // 8 puntos a distancia² 2 de su centroide
		t.Fatalf("inercia inesperada: %v", r["inercia"])
	}
	for i := 0; i < 2; i++ { // los nodos terminan tras FINISH y BYE
		if err := <-nodos; err != nil {
			t.Fatalf("nodo: %v", err)
		}
	}
	s.Cerrar()
	cancel()
	<-servido
}
