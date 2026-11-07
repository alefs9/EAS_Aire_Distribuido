package main

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestPartitionsCoverExactlyOnce(t *testing.T) {
	for n := 1; n <= 31; n++ {
		for workers := 1; workers <= 40; workers++ {
			seen := make([]int, n)
			ranges := partitionRows(n, workers)
			minSize, maxSize := n, 0
			for _, r := range ranges {
				if r.start < 0 || r.end > n || r.start >= r.end {
					t.Fatalf("invalid range: %+v", r)
				}
				size := r.end - r.start
				if size < minSize {
					minSize = size
				}
				if size > maxSize {
					maxSize = size
				}
				for i := r.start; i < r.end; i++ {
					seen[i]++
				}
			}
			if maxSize-minSize > 1 {
				t.Fatal("unbalanced")
			}
			for _, count := range seen {
				if count != 1 {
					t.Fatalf("n=%d workers=%d coverage=%v", n, workers, seen)
				}
			}
		}
	}
}

func assertEquivalent(t *testing.T, a, b KMeansResult) {
	t.Helper()
	if a.Iterations != b.Iterations || a.Converged != b.Converged || !reflect.DeepEqual(a.Labels, b.Labels) {
		t.Fatalf("different iterations, convergence or labels: %v / %v", a, b)
	}
	for c := range a.Centroids {
		for d := range a.Centroids[c] {
			if math.Abs(a.Centroids[c][d]-b.Centroids[c][d]) > 1e-9+1e-9*math.Abs(a.Centroids[c][d]) {
				t.Fatal("centroids differ")
			}
		}
	}
	if math.Abs(a.Inertia-b.Inertia) > 1e-9+1e-9*math.Abs(a.Inertia) {
		t.Fatal("inertia differs")
	}
}

func TestConcurrentEquivalence(t *testing.T) {
	cases := [][][]float64{
		{{1, 1}, {1, 2}, {9, 9}, {9, 10}, {10, 10}},
		{{0, 0}, {0, 0}, {0, 0}}, // empate y clúster vacío
		{{2, 3}},
	}
	for i, points := range cases {
		for _, workers := range []int{1, 2, 3, 4, 100} {
			t.Run(fmt.Sprintf("case%d_workers%d", i, workers), func(t *testing.T) {
				data := NewDataset(points)
				k := 2
				if data.N < k {
					k = data.N
				}
				for _, tol := range []float64{0, 0.001} {
					a, err := RunKMeans(data, k, 10, tol, 42)
					if err != nil {
						t.Fatal(err)
					}
					b, err := RunKMeansConcurrent(data, k, 10, tol, 42, workers)
					if err != nil {
						t.Fatal(err)
					}
					assertEquivalent(t, a, b)
				}
			})
		}
	}
}

func TestConcurrentRejectsInvalidInput(t *testing.T) {
	valid := NewDataset([][]float64{{1, 2}, {3, 4}})
	for _, data := range []Dataset{{}, {Points: valid.Points, N: 3, D: 2}, NewDataset([][]float64{{1}}), NewDataset([][]float64{{math.NaN(), 2}})} {
		if _, err := RunKMeansConcurrent(data, 1, 10, 0.001, 42, 2); err == nil {
			t.Fatal("accepted invalid data")
		}
	}
	for _, c := range []struct {
		k, it, w int
		tol      float64
	}{{0, 10, 2, .001}, {3, 10, 2, .001}, {1, 0, 2, .001}, {1, 10, 0, .001}, {1, 10, -1, .001}, {1, 10, 2, -1}, {1, 10, 2, math.NaN()}, {1, 10, 2, math.Inf(1)}} {
		if _, err := RunKMeansConcurrent(valid, c.k, c.it, c.tol, 42, c.w); err == nil {
			t.Fatal("accepted invalid config")
		}
	}
}
