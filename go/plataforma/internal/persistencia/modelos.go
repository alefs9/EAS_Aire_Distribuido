package persistencia

import (
	"errors"
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
