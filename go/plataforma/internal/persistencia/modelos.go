package persistencia

import (
	"encoding/hex"
	"errors"
	"math"
	"time"
)

var (
	ErrNoEncontrado = errors.New("identificador inexistente")
	ErrEstado       = errors.New("transición de estado no permitida")
	ErrPendiente    = errors.New("la ejecución todavía no tiene un resultado definitivo")
)

// Dataset conserva procedencia; los CSV permanecen en sus archivos de origen.

type Dataset struct {
	ID       string `bson:"_id" json:"dataset_id"`
	Archivo  string `bson:"archivo" json:"archivo"`
	Fuente   string `bson:"fuente" json:"fuente"`
	SHA256   string `bson:"sha256" json:"sha256"`
	Periodo  string `bson:"periodo" json:"periodo"`
	Vectores int    `bson:"vectores" json:"vectores"`
}

type Parametros struct {
	K          int     `bson:"k" json:"k"`
	Semilla    int64   `bson:"semilla" json:"semilla"`
	Tolerancia float64 `bson:"tolerancia" json:"tolerancia"`
	MaxIter    int     `bson:"max_iter" json:"max_iter"`
}

type Resultado struct {
	Centroides    [][]float64 `bson:"centroides" json:"centroides"`
	Conteos       []int       `bson:"conteos" json:"conteos"`
	Iteraciones   int         `bson:"iteraciones" json:"iteraciones"`
	Inercia       float64     `bson:"inercia" json:"inercia"`
	Convergio     bool        `bson:"convergio" json:"convergio"`
	SegundosCarga float64     `bson:"segundos_carga" json:"segundos_carga"`
	SegundosML    float64     `bson:"segundos_ml" json:"segundos_ml"`
	Origen        string      `bson:"origen" json:"origen"`
}

// El resultado resumido se incrusta en jobs: estado y resultado se guardan
// en una única escritura atómica. No se incrustan las N asignaciones.

type Trabajo struct {
	ID          string     `bson:"_id" json:"job_id"`
	DatasetID   string     `bson:"dataset_id" json:"dataset_id"`
	Parametros  Parametros `bson:"parametros" json:"parametros"`
	Estado      string     `bson:"estado" json:"estado"`
	Creado      time.Time  `bson:"creado" json:"creado"`
	Actualizado time.Time  `bson:"actualizado" json:"actualizado"`
	Resultado   *Resultado `bson:"resultado,omitempty" json:"resultado,omitempty"`
	Error       string     `bson:"error,omitempty" json:"error,omitempty"`
}

func finito(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

func validarDataset(d Dataset) error {
	b, err := hex.DecodeString(d.SHA256)
	if d.ID == "" || d.Archivo == "" || d.Fuente == "" || err != nil || len(b) != 32 || d.Vectores < 1 || (d.Periodo != "1_hora" && d.Periodo != "24_horas") {
		return errors.New("dataset inválido: revise procedencia, hash, periodo y vectores")
	}
	return nil
}

func validarParametros(p Parametros, n int) error {
	if p.K < 1 || p.K > n || p.MaxIter < 1 || !finito(p.Tolerancia) || p.Tolerancia < 0 {
		return errors.New("parámetros inválidos")
	}
	return nil
}

func validarResultado(r Resultado, p Parametros, n int) error {
	if len(r.Centroides) != p.K || len(r.Conteos) != p.K || r.Iteraciones < 1 || r.Iteraciones > p.MaxIter || r.Origen == "" {
		return errors.New("dimensiones o metadatos del resultado inválidos")
	}
	if !finito(r.Inercia) || r.Inercia < 0 || !finito(r.SegundosCarga) || r.SegundosCarga < 0 || !finito(r.SegundosML) || r.SegundosML < 0 {
		return errors.New("métricas inválidas")
	}
	suma := 0
	for i, c := range r.Centroides {
		if len(c) != 2 || !finito(c[0]) || !finito(c[1]) || r.Conteos[i] < 0 || r.Conteos[i] > n-suma {
			return errors.New("centroides o conteos inválidos")
		}
		suma += r.Conteos[i]
	}
	if suma != n {
		return errors.New("la suma de conteos no coincide con el dataset")
	}
	return nil
}
