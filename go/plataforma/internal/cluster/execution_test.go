package cluster

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"eas-aire/plataforma/internal/protocol"
)

func TestPartitionCoverage(t *testing.T) {
	for n := 1; n <= 40; n++ {
		for workers := 1; workers <= n; workers++ {
			next := 0
			for _, p := range partition(n, workers) {
				if p.start != next || p.end <= p.start || p.end > n {
					t.Fatalf("n=%d workers=%d invalid span %+v after %d", n, workers, p, next)
				}
				next = p.end
			}
			if next != n {
				t.Fatalf("covered %d of %d", next, n)
			}
		}
	}
}

func TestAssignAndReduce(t *testing.T) {
	centers := []Point{{0, 0}, {10, 0}, {100, 100}}
	points := []Point{{0, 0}, {5, 0}, {9, 2}, {11, 2}}
	parts := make([]*protocol.JobMessage, 2)
	for i, p := range partition(len(points), len(parts)) {
		counts, sums, err := accumulate(context.Background(), points[p.start:p.end], centers)
		if err != nil {
			t.Fatal(err)
		}
		parts[i] = &protocol.JobMessage{Rows: p.end - p.start, Counts: counts, Sums: sums}
	}
	if !slices.Equal(parts[0].Counts, []int{2, 0, 0}) {
		t.Fatalf("tie should select the first cluster: %v", parts[0].Counts)
	}
	got, counts, _, err := reduce(centers, parts)
	if err != nil || !slices.Equal(got, []Point{{2.5, 0}, {10, 2}, {100, 100}}) ||
		!slices.Equal(counts, []int{2, 2, 0}) {
		t.Fatalf("got centers=%v counts=%v error=%v", got, counts, err)
	}
}

func TestPartialIdentityAndCounts(t *testing.T) {
	key := protocol.JobKey{ID: "execution", Iteration: 2, Partition: 1, Attempt: 1}
	req := request{message: protocol.JobFrame(protocol.Iter, "a", key), expected: protocol.Partial}
	valid := protocol.JobFrame(protocol.Partial, "a", key)
	valid.Job.Rows, valid.Job.Counts, valid.Job.Sums = 4, []int{3, 1}, []Point{{3, 9}, {2, 4}}
	p, err := matchResponse(req, valid)
	if err != nil {
		t.Fatal(err)
	}
	if err := validatePartial(p, 4, 2); err != nil {
		t.Fatal(err)
	}
	next := key
	next.Iteration++
	nextReq := request{message: protocol.JobFrame(protocol.Iter, "a", next), expected: protocol.Partial}
	if _, err := matchResponse(nextReq, valid); err == nil {
		t.Fatal("duplicate partial must not satisfy the following iteration")
	}
	for _, bad := range []protocol.JobKey{
		{ID: "other", Iteration: 2, Partition: 1, Attempt: 1},
		{ID: "execution", Iteration: 3, Partition: 1, Attempt: 1},
		{ID: "execution", Iteration: 2, Partition: 0, Attempt: 1},
		{ID: "execution", Iteration: 2, Partition: 1, Attempt: 2},
	} {
		if _, err := matchResponse(req, protocol.JobFrame(protocol.Partial, "a", bad)); err == nil {
			t.Fatalf("accepted wrong key %+v", bad)
		}
	}
	p.Counts[0]--
	if err := validatePartial(p, 4, 2); err == nil {
		t.Fatal("accepted incomplete counts")
	}
}

func TestDistributedTCP(t *testing.T) {
	path := tempCSV(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	coord := NewCoordinator(2*time.Second, quiet)
	served := make(chan error, 1)
	go func() { served <- coord.Serve(ctx, listener) }()
	t.Cleanup(func() {
		cancel()
		if err := <-served; err != nil {
			t.Error(err)
		}
	})
	nodes := make(chan error, 2)
	for _, id := range []string{"b", "a"} {
		go func(id string) {
			nodes <- RunNode(ctx, NodeConfig{Address: listener.Addr().String(), ID: id, Capacity: 1, Interval: time.Millisecond, Timeout: 2 * time.Second}, quiet)
		}(id)
	}
	result, err := coord.Run(ctx, ExecutionConfig{File: path, Nodes: 2, K: 2, Seed: 42, MaxIterations: 20, Tolerance: 1e-8})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || result.Reason != "converged" || result.Rows != 8200 ||
		!slices.Equal(result.Nodes, []string{"a", "b"}) || !slices.Equal(result.Counts, []int{4100, 4100}) {
		t.Fatalf("unexpected result %+v", result)
	}
	lo, hi := result.Centroids[0], result.Centroids[1]
	if lo[0] > hi[0] {
		lo, hi = hi, lo
	}
	if lo != (Point{1, 1}) || hi != (Point{11, 11}) {
		t.Fatalf("unexpected centroids %v", result.Centroids)
	}
	for range result.Nodes {
		if err := <-nodes; err != nil {
			t.Fatal(err)
		}
	}
}

func TestExecutionTimeout(t *testing.T) {
	path := tempCSV(t)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	result, err := NewCoordinator(time.Second, quiet).Run(ctx, ExecutionConfig{
		File: path, Nodes: 2, K: 2, Seed: 42, Tolerance: 1e-8, MaxIterations: 4,
	})
	if err == nil || result.Status != "failed" || len(result.Centroids) != 0 || len(result.Counts) != 0 {
		t.Fatalf("got result %+v and error %v", result, err)
	}
}

func tempCSV(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.csv")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("PM10,PM2.5\n")
	for i := 0; i < 2050; i++ {
		data = append(data, "0,0\n2,2\n10,10\n12,12\n"...)
	}
	_, err = file.Write(data)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatal(fmt.Errorf("prepare fixture: %w", err))
	}
	return path
}
