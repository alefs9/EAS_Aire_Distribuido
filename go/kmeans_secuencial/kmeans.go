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

// KMeansResult almacena los resultados obtenidos al ejecutar K-Means.
type KMeansResult struct {
	Centroids  [][]float64
	Labels     []int
	Inertia    float64
	Iterations int
	Converged  bool
}

// calculateInertia calcula la suma de las distancias al cuadrado
// entre cada punto y el centroide de su clúster.
func calculateInertia(data Dataset, centroids [][]float64, labels []int) float64 {
	var inertia float64

	for i, point := range data.Points {
		inertia += squaredDistance(point, centroids[labels[i]])
	}

	return inertia
}

// RunKMeans ejecuta el algoritmo K-Means secuencial.
func RunKMeans(data Dataset, k int, maxIter int, tol float64, seed int64) (KMeansResult, error) {
	if err := validateDataset(data); err != nil {
		return KMeansResult{}, err
	}

	if k <= 0 || k > data.N {
		return KMeansResult{}, fmt.Errorf("K debe estar entre 1 y N")
	}

	if maxIter <= 0 {
		return KMeansResult{}, fmt.Errorf("el máximo de iteraciones debe ser positivo")
	}

	if tol < 0 || math.IsNaN(tol) || math.IsInf(tol, 0) {
		return KMeansResult{}, fmt.Errorf("la tolerancia debe ser finita y no negativa")
	}

	centroids, err := initializeCentroids(data, k, seed)
	if err != nil {
		return KMeansResult{}, err
	}

	result := KMeansResult{}

	for iteration := 0; iteration < maxIter; iteration++ {
		labels := assignClusters(data, centroids)
		newCentroids := updateCentroids(data, labels, centroids)

		var movement float64

		for cluster := 0; cluster < k; cluster++ {
			movement += squaredDistance(centroids[cluster], newCentroids[cluster])
		}

		centroids = newCentroids
		result.Iterations = iteration + 1

		if math.Sqrt(movement) < tol {
			result.Converged = true
			break
		}
	}

	// Reasignamos los puntos usando los centroides finales.
	result.Centroids = centroids
	result.Labels = assignClusters(data, centroids)
	result.Inertia = calculateInertia(data, centroids, result.Labels)

	return result, nil
}
