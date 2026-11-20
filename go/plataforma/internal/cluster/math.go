package cluster

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"strconv"

	"eas-aire/plataforma/internal/protocol"
)

type Point = protocol.Point

func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }

func loadCSV(ctx context.Context, path string) ([]Point, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := csv.NewReader(file)
	header, err := r.Read()
	if err != nil || len(header) != 2 || header[0] != "PM10" || header[1] != "PM2.5" {
		return nil, fmt.Errorf("expected CSV header PM10,PM2.5")
	}
	var points []Point
	for row := 2; ; row++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		record, err := r.Read()
		if err == io.EOF {
			return points, nil
		}
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", row, err)
		}
		var point Point
		for i, value := range record {
			x, err := strconv.ParseFloat(value, 64)
			if err != nil || !finite(x) {
				return nil, fmt.Errorf("row %d: invalid %s", row, header[i])
			}
			point[i] = x
		}
		points = append(points, point)
	}
}

type span struct{ start, end int }

func partition(n, workers int) []span {
	ranges := make([]span, workers)
	start := 0
	for i := range ranges {
		size := n / workers
		if i < n%workers {
			size++
		}
		ranges[i] = span{start, start + size}
		start += size
	}
	return ranges
}

func initialize(points []Point, k int, seed int64) []Point {
	indices := rand.New(rand.NewSource(seed)).Perm(len(points))
	centers := make([]Point, k)
	for i := range centers {
		centers[i] = points[indices[i]]
	}
	return centers
}

func distance(a, b Point) float64 {
	x, y := a[0]-b[0], a[1]-b[1]
	return x*x + y*y
}

func accumulate(ctx context.Context, points, centers []Point) ([]int, []Point, error) {
	counts, sums := make([]int, len(centers)), make([]Point, len(centers))
	for i, point := range points {
		if i%256 == 0 && ctx.Err() != nil {
			return nil, nil, ctx.Err()
		}
		best, minDistance := 0, distance(point, centers[0])
		for c := 1; c < len(centers); c++ {
			if d := distance(point, centers[c]); d < minDistance {
				best, minDistance = c, d
			}
		}
		if !finite(minDistance) {
			return nil, nil, fmt.Errorf("distance overflow")
		}
		counts[best]++
		sums[best][0] += point[0]
		sums[best][1] += point[1]
	}
	return counts, sums, nil
}

func reduce(centers []Point, parts []*protocol.JobMessage) ([]Point, []int, float64, error) {
	counts, sums := make([]int, len(centers)), make([]Point, len(centers))
	for _, part := range parts {
		for c, count := range part.Counts {
			counts[c] += count
			sums[c][0] += part.Sums[c][0]
			sums[c][1] += part.Sums[c][1]
		}
	}
	next := append([]Point(nil), centers...)
	var movement float64
	for c, count := range counts {
		if count > 0 {
			next[c] = Point{sums[c][0] / float64(count), sums[c][1] / float64(count)}
			movement += distance(next[c], centers[c])
		}
	}
	if !finite(movement) {
		return nil, nil, 0, fmt.Errorf("non-finite centroid movement")
	}
	return next, counts, math.Sqrt(movement), nil
}
