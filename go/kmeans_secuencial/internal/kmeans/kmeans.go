package main

import (
	"fmt"
	"math"
	"math/rand"
)

// initializeCentroids selecciona K puntos aleatorios del dataset
// para utilizarlos como centroides iniciales.
func initializeCentroids(data Dataset, k int, seed int64) ([][]float64, error) {
	if err := validateDataset(data); err != nil {
		return nil, err
	}

	if k <= 0 || k > data.N {
		return nil, fmt.Errorf("K debe estar entre 1 y N")
	}

	rng := rand.New(rand.NewSource(seed))
	indices := rng.Perm(data.N)[:k]

	centroids := make([][]float64, k)

	for i, index := range indices {
		centroids[i] = make([]float64, data.D)
		copy(centroids[i], data.Points[index])
	}

	return centroids, nil
}

// validateDataset comprueba que el conjunto de datos tenga
// una estructura válida para trabajar con K-Means.
func validateDataset(data Dataset) error {
	if data.N <= 0 || len(data.Points) == 0 {
		return fmt.Errorf("el conjunto de datos está vacío")
	}

	if data.N != len(data.Points) {
		return fmt.Errorf("la cantidad de puntos no coincide con N")
	}

	if data.D != 2 {
		return fmt.Errorf("el dataset debe tener exactamente 2 dimensiones")
	}

	for i, point := range data.Points {
		if len(point) != data.D {
			return fmt.Errorf("el punto %d tiene dimensiones inconsistentes", i)
		}

		for _, value := range point {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("el punto %d contiene un valor no finito", i)
			}
		}
	}

	return nil
}

// squaredDistance calcula la distancia euclidiana al cuadrado
// entre dos puntos.
func squaredDistance(a, b []float64) float64 {
	var distance float64

	for i := range a {
		diff := a[i] - b[i]
		distance += diff * diff
	}

	return distance
}

// assignClusters asigna cada punto al clúster cuyo centroide
// se encuentra más cerca.
func assignClusters(data Dataset, centroids [][]float64) []int {
	labels := make([]int, data.N)

	for i, point := range data.Points {
		bestCluster := 0
		minDistance := squaredDistance(point, centroids[0])

		for j := 1; j < len(centroids); j++ {
			distance := squaredDistance(point, centroids[j])

			if distance < minDistance {
				minDistance = distance
				bestCluster = j
			}
		}

		labels[i] = bestCluster
	}

	return labels
}

// updateCentroids actualiza los centroides calculando el promedio
// de los puntos asignados a cada clúster.
// Si un clúster queda vacío, conserva su centroide anterior.
func updateCentroids(data Dataset, labels []int, centroids [][]float64) [][]float64 {
	k := len(centroids)

	newCentroids := make([][]float64, k)
	counts := make([]int, k)

	for i := range newCentroids {
		newCentroids[i] = make([]float64, data.D)
	}

	// Acumulamos los valores de cada dimensión por clúster.
	for i, point := range data.Points {
		cluster := labels[i]
		counts[cluster]++

		for d := 0; d < data.D; d++ {
			newCentroids[cluster][d] += point[d]
		}
	}

	// Calculamos el promedio de cada clúster.
	for cluster := 0; cluster < k; cluster++ {
		if counts[cluster] > 0 {
			for d := 0; d < data.D; d++ {
				newCentroids[cluster][d] /= float64(counts[cluster])
			}
		} else {
			copy(newCentroids[cluster], centroids[cluster])
		}
	}

	return newCentroids
}
