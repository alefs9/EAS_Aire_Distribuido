package main

import "testing"

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
