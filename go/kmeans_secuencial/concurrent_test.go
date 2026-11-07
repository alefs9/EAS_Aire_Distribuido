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
