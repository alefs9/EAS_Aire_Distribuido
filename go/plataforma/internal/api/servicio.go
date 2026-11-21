package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"eas-aire/plataforma"
	"eas-aire/plataforma/internal/persistencia"
)

// Ejecutor es lo que la API necesita del clúster. *plataforma.Coordinator lo cumple.
type Ejecutor interface {
	Run(ctx context.Context, config plataforma.ExecutionConfig) (plataforma.Result, error)
}

// Servicio recibe trabajos y los ejecuta en orden con una cola acotada.
// Hay un único worker porque el coordinador ejecuta un trabajo a la vez.
type Servicio struct {
	almacen      Almacen
	ejecutor     Ejecutor
	datasets     map[string]Dataset // por periodo
	cola         chan tarea
	timeout      time.Duration
	nodosDefecto int
	log          *log.Logger
	wg           sync.WaitGroup
}

// tarea viaja por la cola. Los nodos no forman parte de persistencia.Parametros.
type tarea struct {
	id    string
	nodos int
}

type ConfigServicio struct {
	TamanoCola   int
	Timeout      time.Duration // plazo total de un trabajo, incluida la espera de nodos
	NodosDefecto int
}

func NuevoServicio(a Almacen, e Ejecutor, datasets map[string]Dataset, cfg ConfigServicio, logger *log.Logger) *Servicio {
	s := &Servicio{almacen: a, ejecutor: e, datasets: datasets, cola: make(chan tarea, cfg.TamanoCola),
		timeout: cfg.Timeout, nodosDefecto: cfg.NodosDefecto, log: logger}
	s.wg.Add(1)
	go s.worker()
	return s
}

func nuevoID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return "pc4-" + hex.EncodeToString(b[:])
}

// Enviar valida, registra el trabajo como queued y lo encola. No espera el entrenamiento.
func (s *Servicio) Enviar(ctx context.Context, sol Solicitud) (persistencia.Trabajo, error) {
	p, err := sol.Validar(s.nodosDefecto)
	if err != nil {
		return persistencia.Trabajo{}, err
	}
	d, ok := s.datasets[p.Periodo]
	if !ok {
		return persistencia.Trabajo{}, ErrPeriodoNoActivo
	}
	id := nuevoID()
	params := persistencia.Parametros{K: p.K, Semilla: p.Semilla, Tolerancia: p.Tolerancia, MaxIter: p.MaxIter}
	if err := s.almacen.Crear(ctx, id, d.ID, params); err != nil {
		return persistencia.Trabajo{}, err
	}
	select {
	case s.cola <- tarea{id, p.Nodos}:
	default:
		_ = s.almacen.Fallar(ctx, id, ErrColaLlena.Error())
		return persistencia.Trabajo{}, ErrColaLlena
	}
	if l, _ := s.almacen.Leer(ctx, id); l.Trabajo.ID != "" {
		return l.Trabajo, nil
	}
	return persistencia.Trabajo{ID: id, DatasetID: d.ID, Parametros: params, Estado: "queued"}, nil
}

func (s *Servicio) Leer(ctx context.Context, id string) (persistencia.Lectura, error) {
	return s.almacen.Leer(ctx, id)
}

func (s *Servicio) Salud(ctx context.Context) map[string]string { return s.almacen.Salud(ctx) }

func (s *Servicio) worker() {
	defer s.wg.Done()
	for t := range s.cola {
		s.procesar(t.id, t.nodos)
	}
}

func (s *Servicio) procesar(id string, nodos int) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	l, err := s.almacen.Leer(ctx, id)
	if l.Trabajo.ID == "" {
		s.log.Printf("job %s: no se pudo leer: %v", id, err)
		return
	}
	t := l.Trabajo
	var d Dataset
	for _, ds := range s.datasets {
		if ds.ID == t.DatasetID {
			d = ds
		}
	}
	if err := s.almacen.Iniciar(ctx, id); err != nil {
		s.log.Printf("job %s: no se pudo iniciar: %v", id, err)
		return
	}
	fallar := func(motivo string) {
		// Contexto propio: el del trabajo puede haber vencido.
		c, cl := context.WithTimeout(context.Background(), 5*time.Second)
		defer cl()
		if err := s.almacen.Fallar(c, id, motivo); err != nil {
			s.log.Printf("job %s: no se pudo registrar el fallo: %v", id, err)
		}
		s.log.Printf("job %s failed: %s", id, motivo)
	}
	s.log.Printf("job %s running periodo=%s nodos=%d k=%d", id, d.Periodo, nodos, t.Parametros.K)
	res, err := s.ejecutor.Run(ctx, plataforma.ExecutionConfig{
		File: d.Ruta, Nodes: nodos, K: t.Parametros.K, Seed: t.Parametros.Semilla,
		Tolerance: t.Parametros.Tolerancia, MaxIterations: t.Parametros.MaxIter,
	})
	if err != nil {
		fallar(err.Error())
		return
	}
	if res.Rows != len(d.Puntos) {
		fallar(fmt.Sprintf("el clúster leyó %d filas y la API %d", res.Rows, len(d.Puntos)))
		return
	}
	centroides := make([][]float64, len(res.Centroids))
	for i, c := range res.Centroids {
		centroides[i] = []float64{c[0], c[1]}
	}
	r := persistencia.Resultado{
		Centroides:    centroides,
		Conteos:       res.Counts,
		Iteraciones:   res.Iterations,
		Inercia:       Inercia(d.Puntos, res.Centroids, runtime.GOMAXPROCS(0)),
		Convergio:     res.Reason == "converged",
		SegundosCarga: res.Timings.Load.Seconds(),
		SegundosML:    res.Timings.Iterations.Seconds(),
		Origen:        fmt.Sprintf("clúster PC4 cluster_job=%s nodos=%s", res.ID, strings.Join(res.Nodes, ",")),
	}
	if err := s.almacen.Completar(ctx, id, r); err != nil {
		fallar("resultado no guardado: " + err.Error())
		return
	}
	s.log.Printf("job %s completed iteraciones=%d convergio=%v ml=%.3fs", id, r.Iteraciones, r.Convergio, r.SegundosML)
}

// Cerrar deja de aceptar trabajos y espera a que termine el que está en curso.
func (s *Servicio) Cerrar() { close(s.cola); s.wg.Wait() }
