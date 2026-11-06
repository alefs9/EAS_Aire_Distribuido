package main

import (
	"math"
	"reflect"
	"testing"
)

func TestInitializeCentroidsReproducible(t *testing.T) {
	data := NewDataset([][]float64{
		{1, 1},
		{2, 2},
		{10, 10},
		{11, 11},
	})

	first, err := initializeCentroids(data, 2, 42)
	if err != nil {
		t.Fatalf("error al inicializar: %v", err)
	}

	second, err := initializeCentroids(data, 2, 42)
	if err != nil {
		t.Fatalf("error al inicializar: %v", err)
	}

	if !reflect.DeepEqual(first, second) {
		t.Errorf("la inicialización no fue reproducible")
	}
}

func TestInitializeCentroidsInvalidK(t *testing.T) {
	data := NewDataset([][]float64{
		{1, 1},
		{2, 2},
	})

	invalidValues := []int{0, -1, 3}

	for _, k := range invalidValues {
		_, err := initializeCentroids(data, k, 42)
		if err == nil {
			t.Errorf("se esperaba un error para K=%d", k)
		}
	}
}

func TestAssignClusters(t *testing.T) {
	data := NewDataset([][]float64{
		{1, 1},
		{2, 2},
		{10, 10},
		{11, 11},
	})

	centroids := [][]float64{
		{1, 1},
		{10, 10},
	}

	labels := assignClusters(data, centroids)

	expected := []int{0, 0, 1, 1}

	if !reflect.DeepEqual(labels, expected) {
		t.Errorf("etiquetas incorrectas: se obtuvo %v, se esperaba %v", labels, expected)
	}
}

func TestUpdateCentroids(t *testing.T) {
	data := NewDataset([][]float64{
		{1, 1},
		{3, 3},
		{10, 10},
		{12, 12},
	})

	labels := []int{0, 0, 1, 1}

	centroids := [][]float64{
		{1, 1},
		{10, 10},
	}

	updated := updateCentroids(data, labels, centroids)

	expected := [][]float64{
		{2, 2},
		{11, 11},
	}

	if !reflect.DeepEqual(updated, expected) {
		t.Errorf("centroides incorrectos: se obtuvo %v", updated)
	}
}

func TestUpdateCentroidsEmptyCluster(t *testing.T) {
	data := NewDataset([][]float64{
		{1, 1},
		{2, 2},
	})

	labels := []int{0, 0}

	centroids := [][]float64{
		{1, 1},
		{10, 10},
	}

	updated := updateCentroids(data, labels, centroids)

	if !reflect.DeepEqual(updated[1], centroids[1]) {
		t.Errorf("el centroide del clúster vacío debería conservarse")
	}
}

func TestRunKMeans(t *testing.T) {
	data := NewDataset([][]float64{
		{1, 1},
		{2, 2},
		{10, 10},
		{11, 11},
	})

	result, err := RunKMeans(data, 2, 100, 0.001, 42)
	if err != nil {
		t.Fatalf("no se esperaba un error: %v", err)
	}

	if len(result.Centroids) != 2 {
		t.Errorf("se esperaban 2 centroides")
	}

	if len(result.Labels) != data.N {
		t.Errorf("se esperaba una etiqueta por punto")
	}

	if !result.Converged {
		t.Errorf("se esperaba que el algoritmo convergiera")
	}

	if result.Iterations <= 0 || result.Iterations > 100 {
		t.Errorf("cantidad de iteraciones inválida: %d", result.Iterations)
	}

	if math.IsNaN(result.Inertia) || math.IsInf(result.Inertia, 0) || result.Inertia < 0 {
		t.Errorf("inercia inválida: %f", result.Inertia)
	}

	if math.Abs(result.Inertia-2.0) > 0.000001 {
		t.Errorf("inercia incorrecta: se obtuvo %f, se esperaba 2", result.Inertia)
	}
}

func TestRunKMeansInvalidParameters(t *testing.T) {
	data := NewDataset([][]float64{
		{1, 1},
		{2, 2},
	})

	_, err := RunKMeans(data, 0, 100, 0.001, 42)
	if err == nil {
		t.Error("se esperaba un error para K=0")
	}

	_, err = RunKMeans(data, 2, 0, 0.001, 42)
	if err == nil {
		t.Error("se esperaba un error para maxIter=0")
	}

	_, err = RunKMeans(data, 2, 100, -1, 42)
	if err == nil {
		t.Error("se esperaba un error para una tolerancia negativa")
	}
}

func TestRunKMeansInvalidDataset(t *testing.T) {
	data := NewDataset([][]float64{
		{1, 1},
		{math.NaN(), 2},
	})

	_, err := RunKMeans(data, 2, 100, 0.001, 42)
	if err == nil {
		t.Error("se esperaba un error para un dataset con NaN")
	}
}
