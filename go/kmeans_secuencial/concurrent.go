package main

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
