package main

import (
	"fmt"
	"math"
	"sync"
)

type rowRange struct{ start, end int }

// partitionRows reparte el resto entre los primeros rangos, sin bloques vacíos.
func partitionRows(n, workers int) []rowRange {
	if n <= 0 || workers <= 0 {
		return nil
	}
	if workers > n {
		workers = n
	}
	ranges := make([]rowRange, workers)
	start := 0
	for w := range ranges {
		size := n / workers
		if w < n%workers {
			size++
		}
		ranges[w] = rowRange{start, start + size}
		start += size
	}
	return ranges
}

type partial struct {
	sums   [][Dimensions]float64
	counts []int
}

// concurrentAssignment usa rangos estáticos, no una cola de trabajos.
// Cada goroutine escribe solo su segmento de labels y su acumulador privado.
func concurrentAssignment(data Dataset, centroids [][]float64, ranges []rowRange) ([]int, []partial) {
	labels := make([]int, data.N)
	parts := make([]partial, len(ranges))
	var wg sync.WaitGroup
	for w, rows := range ranges {
		wg.Add(1)
		go func(w int, rows rowRange) {
			defer wg.Done()
			p := partial{make([][Dimensions]float64, len(centroids)), make([]int, len(centroids))}
			for i := rows.start; i < rows.end; i++ {
				point := data.Points[i]
				best := 0
				distance := squaredDistance(point, centroids[0])
				for c := 1; c < len(centroids); c++ {
					if d := squaredDistance(point, centroids[c]); d < distance {
						best, distance = c, d
					}
				}
				labels[i] = best
				p.counts[best]++
				for d := 0; d < Dimensions; d++ {
					p.sums[best][d] += point[d]
				}
			}
			parts[w] = p
		}(w, rows)
	}
	wg.Wait()
	return labels, parts
}

func reduceCentroids(parts []partial, previous [][]float64) [][]float64 {
	centroids := make([][]float64, len(previous))
	for c := range previous {
		centroids[c] = make([]float64, Dimensions)
		count := 0
		// Orden fijo de worker: no depende del orden de finalización.
		for _, p := range parts {
			count += p.counts[c]
			for d := 0; d < Dimensions; d++ {
				centroids[c][d] += p.sums[c][d]
			}
		}
		if count == 0 {
			copy(centroids[c], previous[c])
		} else {
			for d := 0; d < Dimensions; d++ {
				centroids[c][d] /= float64(count)
			}
		}
	}
	return centroids
}

// RunKMeansConcurrent conserva el contrato y la inicialización de RunKMeans.
// El llamador no debe mutar data.Points durante la ejecución.
func RunKMeansConcurrent(data Dataset, k, maxIter int, tol float64, seed int64, workers int) (KMeansResult, error) {
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
	if workers <= 0 {
		return KMeansResult{}, fmt.Errorf("workers debe ser positivo")
	}
	centroids, err := initializeCentroids(data, k, seed)
	if err != nil {
		return KMeansResult{}, err
	}
	ranges := partitionRows(data.N, workers)
	result := KMeansResult{}
	for iteration := 0; iteration < maxIter; iteration++ {
		_, parts := concurrentAssignment(data, centroids, ranges)
		next := reduceCentroids(parts, centroids)
		var movement float64
		for c := range centroids {
			movement += squaredDistance(centroids[c], next[c])
		}
		centroids = next
		result.Iterations = iteration + 1
		if math.Sqrt(movement) < tol {
			result.Converged = true
			break
		}
	}
	result.Centroids = centroids
	result.Labels, _ = concurrentAssignment(data, centroids, ranges)
	// Misma suma en orden de observaciones que la referencia.
	result.Inertia = calculateInertia(data, centroids, result.Labels)
	return result, nil
}
