package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	persistencia "eas-aire/plataforma/internal/persistencia"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func nuevoID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func hashTexto(s string) string { b := sha256.Sum256([]byte(s)); return hex.EncodeToString(b[:]) }
func requerido(nombre string) string {
	v := os.Getenv(nombre)
	if v == "" {
		panic("Falta variable " + nombre)
	}
	return v
}
func guardar(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0600)
}

type ReportePC3 struct {
	Period        string      `json:"period"`
	Vectors       int         `json:"vectors"`
	K             int         `json:"k"`
	Seed          int64       `json:"seed"`
	Tolerance     float64     `json:"tolerance"`
	MaxIterations int         `json:"max_iterations"`
	Centroids     [][]float64 `json:"centroids"`
	Counts        []int       `json:"counts"`
	Iterations    int         `json:"iterations"`
	Inertia       float64     `json:"inertia"`
	Converged     bool        `json:"converged"`
	LoadSeconds   float64     `json:"load_seconds"`
	TrainSeconds  float64     `json:"train_seconds"`
}
type Paso struct {
	Caso      string `json:"caso"`
	Estado    string `json:"estado"`
	Fuente    string `json:"fuente"`
	TTLms     int64  `json:"ttl_ms"`
	Resultado string `json:"resultado"`
}

func ejecutar() error {
	entrada := flag.String("reporte", "", "JSON producido por la referencia PC3")
	metadata := flag.String("dataset", "", "Metadatos del CSV real")
	salida := flag.String("salida", "evidencia.json", "Archivo de evidencia")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	r, err := persistencia.ConectarMongo(ctx, requerido("MONGODB_URI"), requerido("MONGODB_DATABASE"))
	if err != nil {
		return err
	}
	defer r.Cliente.Disconnect(context.Background())
	if err = r.Preparar(ctx); err != nil {
		return err
	}
	c := persistencia.NuevaCache(requerido("REDIS_ADDR"), os.Getenv("REDIS_PASSWORD"))
	defer c.Close()
	if err = c.Ping(ctx).Err(); err != nil {
		return errors.New("Redis no disponible para la demostración")
	}
	ttl, err := strconv.Atoi(requerido("CACHE_TTL_SECONDS"))
	if err != nil || ttl < 1 || ttl > 30 {
		return errors.New("use TTL de 1 a 30 segundos en la demostración")
	}
	s := &persistencia.Servicio{Repo: r, Cache: c, TTL: time.Duration(ttl) * time.Second, Prefijo: os.Getenv("MONGODB_DATABASE")}
	var d persistencia.Dataset
	var a ReportePC3
	b, err := os.ReadFile(*metadata)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &d); err != nil {
		return err
	}
	b, err = os.ReadFile(*entrada)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, &a); err != nil {
		return err
	}
	if d.Periodo != a.Period || d.Vectores != a.Vectors {
		return errors.New("reporte y dataset no corresponden")
	}
	// El identificador del dataset depende del contenido y periodo.
	d.ID = d.SHA256 + ":" + d.Periodo
	if previo, e := r.Dataset(ctx, d.ID); errors.Is(e, persistencia.ErrNoEncontrado) {
		err = r.GuardarDataset(ctx, d)
	} else if e != nil {
		return e
	} else if previo != d {
		return errors.New("metadatos distintos para un mismo dataset")
	}
	if err != nil {
		return err
	}
	p := persistencia.Parametros{K: a.K, Semilla: a.Seed, Tolerancia: a.Tolerance, MaxIter: a.MaxIterations}
	result := persistencia.Resultado{Centroides: a.Centroids, Conteos: a.Counts, Iteraciones: a.Iterations, Inercia: a.Inertia, Convergio: a.Converged, SegundosCarga: a.LoadSeconds, SegundosML: a.TrainSeconds, Origen: "referencia secuencial PC3 ejecutada en este cuaderno; no es clúster PC4"}
	id := "pc4-" + nuevoID()
	if err = r.Crear(ctx, id, d.ID, p); err != nil {
		return err
	}
	pasos := []Paso{}
	pendiente := func(estado string) error {
		v, e := s.Resultado(ctx, id)
		if !errors.Is(e, persistencia.ErrPendiente) || v.Trabajo.Estado != estado {
			return errors.New("estado pendiente incorrecto")
		}
		if n, _ := c.Exists(ctx, s.Clave(id)).Result(); n != 0 {
			return errors.New("se cacheó un trabajo pendiente")
		}
		pasos = append(pasos, Paso{"Consulta de estado", estado, "MongoDB", -2, "PASS: sin resultado ni caché"})
		return nil
	}
	if err = pendiente("queued"); err != nil {
		return err
	}
	if err = s.Iniciar(ctx, id); err != nil {
		return err
	}
	if err = pendiente("running"); err != nil {
		return err
	}
	if err = s.Completar(ctx, id, result); err != nil {
		return err
	}
	leer := func(caso, fuente string) error {
		v, e := s.Resultado(ctx, id)
		if e != nil {
			return e
		}
		if v.Fuente != fuente {
			return fmt.Errorf("%s: fuente %s, se esperaba %s", caso, v.Fuente, fuente)
		}
		dur, e := c.PTTL(ctx, s.Clave(id)).Result()
		if e != nil {
			return e
		}
		pasos = append(pasos, Paso{caso, v.Trabajo.Estado, v.Fuente, dur.Milliseconds(), "PASS"})
		return nil
	}
	if err = leer("Primera consulta sin caché", "MongoDB"); err != nil {
		return err
	}
	if err = leer("Segunda consulta con caché", "Redis"); err != nil {
		return err
	}
	time.Sleep(s.TTL + 100*time.Millisecond)
	if err = leer("Consulta después de expirar", "MongoDB"); err != nil {
		return err
	}
	if err = s.Invalidar(ctx, id); err != nil {
		return err
	}
	if n, e := c.Exists(ctx, s.Clave(id)).Result(); e != nil || n != 0 {
		return errors.New("clave no invalidada")
	}
	if err = leer("Consulta después de invalidar", "MongoDB"); err != nil {
		return err
	}
	job, err := r.Consultar(ctx, id)
	if err != nil {
		return err
	}
	cacheValue, err := c.Get(ctx, s.Clave(id)).Result()
	if err != nil {
		return err
	}
	indices := map[string][]bson.M{}
	for _, name := range []string{"datasets", "jobs"} {
		cur, e := r.DB.Collection(name).Indexes().List(ctx)
		if e != nil {
			return e
		}
		var list []bson.M
		if e = cur.All(ctx, &list); e != nil {
			return e
		}
		indices[name] = list
	}
	fallido := "pc4-fallo-" + nuevoID()
	if err = r.Crear(ctx, fallido, d.ID, p); err != nil {
		return err
	}
	if err = s.Fallar(ctx, fallido, "fallo de procesamiento simulado para probar almacenamiento"); err != nil {
		return err
	}
	failed, err := r.Consultar(ctx, fallido)
	if err != nil {
		return err
	}
	if err = guardar(*salida, map[string]any{"fecha_utc": time.Now().UTC(), "dataset": d, "job_id": id, "documento_mongodb": job, "ejecucion_fallida": failed, "indices": indices, "clave_redis": s.Clave(id), "valor_redis": json.RawMessage(cacheValue), "pruebas": pasos, "alcance": "A1-A3: persistencia y caché; API y clúster no ejecutados"}); err != nil {
		return err
	}
	for _, paso := range pasos {
		fmt.Printf("%s | %s | %s | %s\n", paso.Caso, paso.Estado, paso.Fuente, paso.Resultado)
	}
	fmt.Printf("Evidencia: %s\njob_id: %s\n", *salida, id)
	return nil
}
func main() {
	if err := ejecutar(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
