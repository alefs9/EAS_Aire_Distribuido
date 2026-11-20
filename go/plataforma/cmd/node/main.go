package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eas-aire/plataforma/internal/cluster"
)

func main() {
	var cfg cluster.NodeConfig
	flag.StringVar(&cfg.Address, "coordinator", "127.0.0.1:9000", "coordinator TCP address (host:port)")
	flag.StringVar(&cfg.ID, "id", "", "unique node identifier (required)")
	flag.IntVar(&cfg.Capacity, "capacity", 1, "declared node capacity")
	flag.DurationVar(&cfg.Interval, "interval", 3*time.Second, "heartbeat interval")
	flag.DurationVar(&cfg.Timeout, "timeout", 5*time.Second, "maximum duration per dial, read or write")
	flag.IntVar(&cfg.Checks, "checks", 0, "stop after this many acknowledged heartbeats (0 runs until interrupted)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	log.Printf("starting node_id=%q coordinator=%s pid=%d", cfg.ID, cfg.Address, os.Getpid())
	if err := cluster.RunNode(ctx, cfg, log.Default()); err != nil {
		log.Fatal(err)
	}
	log.Printf("node stopped node_id=%s", cfg.ID)
}
