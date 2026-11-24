package main

import (
	"bytes"
	"context"
	"log"
	"net"
	"strings"
	"testing"
)

func TestBusyPortDoesNotAnnounceAvailability(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	var output bytes.Buffer
	err = run(context.Background(), listener.Addr().String(), "http://127.0.0.1:8080", log.New(&output, "", 0))
	if err == nil || !strings.Contains(err.Error(), "puerto está ocupado") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(output.String(), "Interfaz disponible") {
		t.Fatal("announced an unavailable server")
	}
}
func TestShutdownOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := run(ctx, "127.0.0.1:0", "http://127.0.0.1:8080", log.New(&bytes.Buffer{}, "", 0)); err != nil {
		t.Fatal(err)
	}
}
