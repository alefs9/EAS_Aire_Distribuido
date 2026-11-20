package persistencia

import (
	"math"
	"strings"
	"testing"
)

func ejemplo() (Dataset, Parametros, Resultado) {
	return Dataset{"prueba-4", "fixture_4.csv", "fixture sintético de prueba", strings.Repeat("a", 64), "1_hora", 4},
		Parametros{2, 42, .001, 100}, Resultado{Centroides: [][]float64{{1.5, 1.5}, {10.5, 10.5}}, Conteos: []int{2, 2}, Iteraciones: 2, Inercia: 2, Convergio: true, Origen: "fixture conocido; no es ejecución distribuida"}
}

func TestValidaciones(t *testing.T) {
	d, p, r := ejemplo()
	if validarDataset(d) != nil || validarParametros(p, 4) != nil || validarResultado(r, p, 4) != nil {
		t.Fatal("fixture válido rechazado")
	}
	d.SHA256 = "incorrecto"
	if validarDataset(d) == nil {
		t.Fatal("hash inválido aceptado")
	}
	r.Inercia = math.NaN()
	if validarResultado(r, p, 4) == nil {
		t.Fatal("NaN aceptado")
	}
	_, _, r = ejemplo()
	r.Conteos = []int{1, 2}
	if validarResultado(r, p, 4) == nil {
		t.Fatal("conteos incompletos aceptados")
	}
	p.K = 0
	if validarParametros(p, 4) == nil {
		t.Fatal("K inválido aceptado")
	}
}
