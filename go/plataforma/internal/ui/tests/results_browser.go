//go:build ignore

// Ejecutar desde go/plataforma. Usa CSV y cluster TCP reales, almacenamiento en memoria.
package main

import (
	"context"
	"eas-aire/plataforma"
	"eas-aire/plataforma/internal/api"
	"eas-aire/plataforma/internal/cluster"
	"eas-aire/plataforma/internal/ui"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

func main() {
	logger := log.New(io.Discard, "", 0)
	ctx := context.Background()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	coordinator := plataforma.NewCoordinator(10*time.Second, logger)
	go coordinator.Serve(ctx, ln)
	for _, id := range []string{"ui-1", "ui-2"} {
		go func(id string) {
			for {
				cluster.RunNode(ctx, cluster.NodeConfig{Address: ln.Addr().String(), ID: id, Capacity: 1, Interval: 100 * time.Millisecond, Timeout: 10 * time.Second}, logger)
				time.Sleep(250 * time.Millisecond)
			}
		}(id)
	}
	datasets := map[string]api.Dataset{}
	for _, period := range []string{"1_hora", "24_horas"} {
		d, err := api.CargarDataset("../../data/vectors", period)
		if err != nil {
			log.Fatal(err)
		}
		datasets[period] = d
	}
	service := api.NuevoServicio(api.NuevaMemoria(), coordinator, datasets, api.ConfigServicio{TamanoCola: 2, Timeout: 2 * time.Minute, NodosDefecto: 2}, logger)
	go func() { log.Fatal(http.ListenAndServe("127.0.0.1:8088", api.NuevoRouter(service, logger))) }()
	handler, err := ui.NewHandler("http://127.0.0.1:8088")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("UI de comprobacion: http://127.0.0.1:8096/resultados (memoria, sin MongoDB/Redis)")
	log.Fatal(http.ListenAndServe("127.0.0.1:8096", handler))
}
