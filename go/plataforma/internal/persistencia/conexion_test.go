package persistencia

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMongoNoDisponible(t *testing.T) {
	if os.Getenv("PC4_INTEGRATION") != "1" {
		t.Skip("prueba de integración")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if _, err := ConectarMongo(ctx, "mongodb://127.0.0.1:1", "pc4_prueba"); err == nil {
		t.Fatal("se esperaba error de conexión")
	}
}
