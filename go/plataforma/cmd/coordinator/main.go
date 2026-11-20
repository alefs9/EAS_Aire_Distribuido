package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eas-aire/plataforma/internal/cluster"
)

func main() {
	addr := flag.String("listen", "127.0.0.1:9000", "TCP listen address (host:port)")
	timeout := flag.Duration("timeout", 15*time.Second, "maximum duration per read or write")
	flag.Parse()
	if *timeout <= 0 {
		log.Fatal("timeout must be positive")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening addr=%s pid=%d", listener.Addr(), os.Getpid())
	if err := cluster.Serve(ctx, listener, *timeout, log.Default()); err != nil {
		log.Fatal(err)
	}
	log.Print("coordinator stopped")
}
