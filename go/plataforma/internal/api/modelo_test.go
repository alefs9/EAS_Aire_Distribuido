package api

import (
	"errors"
	"testing"
)

func TestValidarValoresPorDefecto(t *testing.T) {
	p, err := Solicitud{}.Validar(2)
	if err != nil {
		t.Fatal(err)
	}
	want := Parametros{Periodo: "1_hora", K: 4, Semilla: 42, Tolerancia: 1e-4, MaxIter: 100, Nodos: 2}
	if p != want {
		t.Fatalf("obtuve %+v, esperaba %+v", p, want)
	}
}

func TestValidarSemillaYToleranciaCero(t *testing.T) {
	cero, tol := int64(0), 0.0
	p, err := Solicitud{Semilla: &cero, Tolerancia: &tol}.Validar(1)
	if err != nil || p.Semilla != 0 || p.Tolerancia != 0 {
		t.Fatalf("semilla y tolerancia explícitas en cero no se respetaron: %+v %v", p, err)
	}
}

func TestValidarRechazaParametros(t *testing.T) {
	neg := -1.0
	casos := map[string]Solicitud{
		"periodo":    {Periodo: "semanal"},
		"k alto":     {K: 21},
		"k negativo": {K: -1},
		"max_iter":   {MaxIter: 1001},
		"nodos":      {Nodos: 9},
		"tolerancia": {Tolerancia: &neg},
	}
	for nombre, s := range casos {
		_, err := s.Validar(2)
		var ve ErrorValidacion
		if !errors.As(err, &ve) {
			t.Errorf("%s: esperaba ErrorValidacion, obtuve %v", nombre, err)
		}
	}
}
