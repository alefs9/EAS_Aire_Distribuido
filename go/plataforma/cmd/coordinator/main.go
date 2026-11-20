package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"eas-aire/plataforma"
)

func main() {
	addr := flag.String("listen", "127.0.0.1:9000", "TCP listen address (host:port)")
	timeout := flag.Duration("timeout", 15*time.Second, "maximum duration per read or write")
	jobTimeout := flag.Duration("job-timeout", 5*time.Minute, "maximum duration of a job including node registration")
	output := flag.String("output", "", "result JSON path; defaults to stdout")
	var config plataforma.ExecutionConfig
	flag.StringVar(&config.File, "file", "", "PM10,PM2.5 CSV; empty runs TCP registration only")
	flag.IntVar(&config.Nodes, "nodes", 2, "number of participating nodes")
	flag.IntVar(&config.K, "k", 4, "number of clusters")
	flag.Int64Var(&config.Seed, "seed", 42, "initialization seed")
	flag.Float64Var(&config.Tolerance, "tol", 1e-4, "convergence tolerance")
	flag.IntVar(&config.MaxIterations, "max-iter", 100, "maximum number of centroid updates")
	flag.Parse()
	if *timeout <= 0 || *jobTimeout <= 0 {
		log.Fatal("timeouts must be positive")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listening addr=%s pid=%d", listener.Addr(), os.Getpid())
	coordinator := plataforma.NewCoordinator(*timeout, log.Default())
	if config.File == "" {
		if err := coordinator.Serve(ctx, listener); err != nil {
			log.Fatal(err)
		}
		log.Print("coordinator stopped")
		return
	}
	serveCtx, stopServing := context.WithCancel(ctx)
	jobCtx, stopJob := context.WithTimeout(serveCtx, *jobTimeout)
	served := make(chan error, 1)
	go func() {
		served <- coordinator.Serve(serveCtx, listener)
		stopJob()
	}()
	result, err := coordinator.Run(jobCtx, config)
	stopJob()
	stopServing()
	serveErr := <-served
	log.Print("coordinator stopped")
	if writeErr := writeResult(*output, result); writeErr != nil {
		log.Printf("write result: %v", writeErr)
		if err == nil {
			err = writeErr
		}
	}
	if serveErr != nil {
		log.Printf("serve: %v", serveErr)
		if err == nil {
			err = serveErr
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func writeResult(path string, result plataforma.Result) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if path == "" {
		_, err := os.Stdout.Write(data)
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".kmeans-result-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(data)
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("write temporary result: %w", writeErr)
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}
