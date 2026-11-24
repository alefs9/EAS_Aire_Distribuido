// Command ui serves the SPA and forwards API traffic to the existing Go API.
package main

import (
	"context"
	"eas-aire/plataforma/internal/ui"
	"errors"
	"fmt"
	"log"
	"net"
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, addr, upstream, log.Default()); err != nil {
		log.Fatal(err)
	}
}

// run only announces availability after the listening socket is acquired.
func run(ctx context.Context, addr, upstream string, logger *log.Logger) error {
	handler, err := ui.NewHandler(upstream)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("no se pudo iniciar la interfaz en %s; comprueba si el puerto está ocupado: %w", addr, err)
	}
	defer listener.Close()
	srv := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	done := make(chan error, 1)
	go func() { done <- srv.Serve(listener) }()
	logger.Printf("Interfaz disponible en http://%s", listener.Addr())
	select {
	case err := <-done:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(closeCtx)
	}
}
