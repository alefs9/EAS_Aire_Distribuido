// Package api expone el clúster K-Means mediante una API REST asíncrona.
// La API actúa como coordinador: registra el trabajo, lo encola, lo ejecuta en
// los nodos TCP y guarda el resultado mediante el paquete persistencia.
package api

import (
	"errors"
	"math"
)

// Periodos admitidos; cada uno corresponde a un CSV de vectores [PM10, PM2.5].
var Periodos = map[string]string{
	"1_hora":   "features_1_hora.csv",
	"24_horas": "features_24_horas.csv",
}

// Solicitud es el cuerpo de POST /jobs. Los campos omitidos toman valores por defecto.
type Solicitud struct {
	Periodo    string   `json:"periodo"`
	K          int      `json:"k"`
	Semilla    *int64   `json:"semilla"`
	Tolerancia *float64 `json:"tolerancia"`
	MaxIter    int      `json:"max_iter"`
	Nodos      int      `json:"nodos"`
}

// Parametros es la solicitud ya validada y completada.
type Parametros struct {
	Periodo    string
	K          int
	Semilla    int64
	Tolerancia float64
	MaxIter    int
	Nodos      int
}

// ErrorValidacion indica parámetros inválidos y se responde con HTTP 400.
type ErrorValidacion struct{ Mensaje string }

func (e ErrorValidacion) Error() string { return e.Mensaje }

var (
	ErrColaLlena       = errors.New("la cola de trabajos está llena")
	ErrPeriodoNoActivo = errors.New("el periodo no está registrado en este servidor")
)

// Validar completa los valores por defecto y comprueba los rangos.
// nodosPorDefecto proviene de la configuración del servidor (CLUSTER_NODES).
func (s Solicitud) Validar(nodosPorDefecto int) (Parametros, error) {
	p := Parametros{Periodo: s.Periodo, K: s.K, MaxIter: s.MaxIter, Nodos: s.Nodos, Semilla: 42, Tolerancia: 1e-4}
	if p.Periodo == "" {
		p.Periodo = "1_hora"
	}
	if p.K == 0 {
		p.K = 4
	}
	if p.MaxIter == 0 {
		p.MaxIter = 100
	}
	if p.Nodos == 0 {
		p.Nodos = nodosPorDefecto
	}
	if s.Semilla != nil {
		p.Semilla = *s.Semilla
	}
	if s.Tolerancia != nil {
		p.Tolerancia = *s.Tolerancia
	}
	switch {
	case Periodos[p.Periodo] == "":
		return p, ErrorValidacion{"periodo debe ser '1_hora' o '24_horas'"}
	case p.K < 1 || p.K > 20:
		return p, ErrorValidacion{"k debe estar entre 1 y 20"}
	case p.MaxIter < 1 || p.MaxIter > 1000:
		return p, ErrorValidacion{"max_iter debe estar entre 1 y 1000"}
	case p.Nodos < 1 || p.Nodos > 8:
		return p, ErrorValidacion{"nodos debe estar entre 1 y 8"}
	case p.Tolerancia < 0 || math.IsNaN(p.Tolerancia) || math.IsInf(p.Tolerancia, 0):
		return p, ErrorValidacion{"tolerancia debe ser un número finito no negativo"}
	}
	return p, nil
}
