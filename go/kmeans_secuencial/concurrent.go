package main

import "sync"

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
