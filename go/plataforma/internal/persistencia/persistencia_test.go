package persistencia

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func nuevoID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}

func hashTexto(s string) string { b := sha256.Sum256([]byte(s)); return hex.EncodeToString(b[:]) }

func servicioPrueba(t *testing.T) (*Servicio, context.Context) {
	t.Helper()
	if os.Getenv("PC4_INTEGRATION") != "1" {
		t.Skip("PC4_INTEGRATION=1 requiere MongoDB y Redis reales")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	r, err := ConectarMongo(ctx, os.Getenv("MONGODB_URI"), os.Getenv("MONGODB_DATABASE"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Cliente.Disconnect(context.Background()) })
	if err = r.Preparar(ctx); err != nil {
		t.Fatal(err)
	}
	c := NuevaCache(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"))
	t.Cleanup(func() { _ = c.Close() })
	if err = c.Ping(ctx).Err(); err != nil {
		t.Fatal(err)
	}
	s := &Servicio{r, c, 2 * time.Second, os.Getenv("MONGODB_DATABASE")}
	return s, ctx
}

func crearFixture(t *testing.T, s *Servicio, ctx context.Context) (string, Resultado) {
	t.Helper()
	d, p, r := ejemplo()
	id := nuevoID()
	d.ID = id
	d.SHA256 = hashTexto(id)
	if err := s.Repo.GuardarDataset(ctx, d); err != nil {
		t.Fatal(err)
	}
	if err := s.Repo.Crear(ctx, id, d.ID, p); err != nil {
		t.Fatal(err)
	}
	return id, r
}

func TestMongoEscrituraConsultaIndices(t *testing.T) {
	s, ctx := servicioPrueba(t)
	id, _ := crearFixture(t, s, ctx)
	j, err := s.Repo.Consultar(ctx, id)
	if err != nil || j.Estado != "queued" {
		t.Fatal(j, err)
	}
	if err = s.Repo.Crear(ctx, id, j.DatasetID, j.Parametros); !mongo.IsDuplicateKeyError(err) {
		t.Fatal("no se rechazó job_id duplicado", err)
	}
	d, err := s.Repo.Dataset(ctx, j.DatasetID)
	if err != nil {
		t.Fatal(err)
	}
	d.ID = nuevoID()
	if err = s.Repo.GuardarDataset(ctx, d); !mongo.IsDuplicateKeyError(err) {
		t.Fatal("índice hash/periodo no rechazó duplicado", err)
	}
	_, err = s.Repo.DB.Collection("jobs").InsertOne(ctx, bson.M{"_id": nuevoID(), "estado": "inventado"})
	if err == nil {
		t.Fatal("MongoDB aceptó documento fuera del esquema")
	}
	_, err = s.Repo.Consultar(ctx, "no-existe")
	if !errors.Is(err, ErrNoEncontrado) {
		t.Fatal(err)
	}
}

func TestEstadosResultadoAtomico(t *testing.T) {
	s, ctx := servicioPrueba(t)
	id, result := crearFixture(t, s, ctx)
	if err := s.Completar(ctx, id, result); !errors.Is(err, ErrEstado) {
		t.Fatal("queued no debe completarse directamente", err)
	}
	if err := s.Iniciar(ctx, id); err != nil {
		t.Fatal(err)
	}
	bad := result
	bad.Conteos = []int{1, 2}
	if s.Completar(ctx, id, bad) == nil {
		t.Fatal("resultado incompleto aceptado")
	}
	j, _ := s.Repo.Consultar(ctx, id)
	if j.Estado != "running" || j.Resultado != nil {
		t.Fatal("se publicó un resultado parcial")
	}
	if err := s.Completar(ctx, id, result); err != nil {
		t.Fatal(err)
	}
	j, err := s.Repo.Consultar(ctx, id)
	if err != nil || j.Estado != "completed" || !reflect.DeepEqual(*j.Resultado, result) {
		t.Fatal(j, err)
	}
	if err = s.Fallar(ctx, id, "fallo tardío"); !errors.Is(err, ErrEstado) {
		t.Fatal("estado terminal mutable", err)
	}
	if err = s.Completar(ctx, id, result); !errors.Is(err, ErrEstado) {
		t.Fatal("resultado terminal mutable", err)
	}
}

func TestEjecucionFallida(t *testing.T) {
	s, ctx := servicioPrueba(t)
	id, _ := crearFixture(t, s, ctx)
	if err := s.Fallar(ctx, id, "nodo desconectado: caso de prueba de almacenamiento"); err != nil {
		t.Fatal(err)
	}
	j, err := s.Repo.Consultar(ctx, id)
	if err != nil || j.Estado != "failed" || j.Resultado != nil || j.Error == "" {
		t.Fatal(j, err)
	}
	if _, err = s.Resultado(ctx, id); !errors.Is(err, ErrPendiente) {
		t.Fatal("se publicó resultado de ejecución fallida")
	}
}

func TestTransicionConcurrente(t *testing.T) {
	s, ctx := servicioPrueba(t)
	id, _ := crearFixture(t, s, ctx)
	var ganadores atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.Iniciar(ctx, id)
			if err == nil {
				ganadores.Add(1)
			} else if !errors.Is(err, ErrEstado) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if ganadores.Load() != 1 {
		t.Fatalf("transiciones aceptadas: %d", ganadores.Load())
	}
}
