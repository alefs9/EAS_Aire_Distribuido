// Command ui serves the SPA and forwards API traffic to the existing Go API.
package main

import (
	"context"
	"eas-aire/plataforma/internal/ui"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	upstream := os.Getenv("UI_API_URL")
	if upstream == "" {
		upstream = "http://127.0.0.1:8080"
	}
	addr := os.Getenv("UI_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8090"
	}
	handler, err := ui.NewHandler(upstream)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Printf("Interfaz disponible en http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(closeCtx); err != nil {
		log.Print(err)
	}
}
