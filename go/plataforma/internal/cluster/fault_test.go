package cluster

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"eas-aire/plataforma/internal/protocol"
)

type jobOutcome struct {
	result Result
	err    error
}

func faultInput(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vectors.csv")
	if err := os.WriteFile(path, []byte("PM10,PM2.5\n0,0\n2,2\n10,10\n12,12\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func within[T any](t *testing.T, done <-chan T) T {
	t.Helper()
	select {
	case result := <-done:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("operation did not finish within its test deadline")
		var zero T
		return zero
	}
}

func readFrame(t *testing.T, conn net.Conn, kind string) protocol.Message {
	t.Helper()
	m, err := protocol.Read(conn)
	if err != nil || m.Type != kind {
		t.Fatalf("expected %s, got %s (%v)", kind, m.Type, err)
	}
	return m
}

func writeFrame(t *testing.T, conn net.Conn, m protocol.Message) {
	t.Helper()
	if err := protocol.Write(conn, m); err != nil {
		t.Fatal(err)
	}
}

func beginAssignment(t *testing.T, conn net.Conn, id string) protocol.Message {
	t.Helper()
	for ts := int64(1); ; ts++ {
		m := protocol.New(protocol.Heartbeat, id)
		m.TS = ts
		writeFrame(t, conn, m)
		ack := readFrame(t, conn, protocol.HeartbeatAck)
		if ack.Work {
			return readFrame(t, conn, protocol.Assign)
		}
	}
}

func ackJob(t *testing.T, conn net.Conn, m protocol.Message, kind string, value int) {
	t.Helper()
	ack := protocol.JobFrame(kind, m.NodeID, m.Job.JobKey)
	switch kind {
	case protocol.VectorsAck:
		ack.Job.Offset = value
	case protocol.Ready:
		ack.Job.Rows = value
	}
	writeFrame(t, conn, ack)
}

func TestExecutionFaults(t *testing.T) {
	cases := []struct {
		name, marker string
		updates      int
	}{
		{"duplicate", "invalid response", 1},
		{"stale_iteration", "invalid response", 1},
		{"future_iteration", "invalid response", 0},
		{"wrong_execution", "invalid response", 0},
		{"wrong_partition", "invalid response", 0},
		{"wrong_attempt", "invalid message fields", 0},
		{"incomplete_counts", "partial counts do not cover partition", 0},
		{"incompatible_dimensions", "partial dimensions", 0},
		{"disconnect_transfer", "distribution:", 0},
		{"disconnect_iteration", "iteration 1:", 0},
		{"silent_node", "i/o timeout", 0},
		{"cancel_context", "context canceled", 0},
		{"coordinator_interruption", "iteration 1:", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			serveCtx, stop := context.WithCancel(ctx)
			defer stop()
			jobCtx, cancelJob := context.WithCancel(ctx)
			defer cancelJob()
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			var coordinatorLog, goodLog bytes.Buffer
			c := NewCoordinator(time.Second, log.New(&coordinatorLog, "", 0))
			served := make(chan error, 1)
			go func() { served <- c.Serve(serveCtx, listener) }()
			serviceStopped := false
			joinService := func() {
				stop()
				if err := within(t, served); err != nil {
					t.Error(err)
				}
				serviceStopped = true
			}
			defer func() {
				if !serviceStopped {
					joinService()
				}
				t.Logf("coordinador:\n%s", coordinatorLog.String())
			}()

			addr := listener.Addr().String()
			bad := dial(t, addr)
			reply(t, bad, register("a"), protocol.New(protocol.Registered, "a"))
			good := startFaultNode(ctx, addr, "b", &goodLog)
			path := faultInput(t)
			cfg := ExecutionConfig{File: path, Nodes: 2, K: 2, Seed: 42, MaxIterations: 4}
			done := make(chan jobOutcome, 1)
			go func() { result, err := c.Run(jobCtx, cfg); done <- jobOutcome{result, err} }()
			assign := beginAssignment(t, bad, "a")
			ackJob(t, bad, assign, protocol.Assigned, 0)
			batch := readFrame(t, bad, protocol.Vectors)
			if tc.name == "disconnect_transfer" {
				t.Log("nodo a: cierra TCP tras recibir VECTORS, antes de confirmarlo")
				bad.Close()
			} else {
				ackJob(t, bad, batch, protocol.VectorsAck, len(batch.Job.Points))
				ackJob(t, bad, readFrame(t, bad, protocol.LoadDone), protocol.Ready, assign.Job.Rows)
				iter := readFrame(t, bad, protocol.Iter)
				part := protocol.JobFrame(protocol.Partial, "a", iter.Job.JobKey)
				part.Job.Rows, part.Job.Counts, part.Job.Sums = 2, []int{2, 0}, []Point{{2, 2}, {}}
				t.Logf("nodo a: inyecta %s tras recibir ITER %d", tc.name, iter.Job.Iteration)
				switch tc.name {
				case "duplicate":
					writeFrame(t, bad, part)
					readFrame(t, bad, protocol.Iter)
					writeFrame(t, bad, part)
				case "stale_iteration":
					writeFrame(t, bad, part)
					next := readFrame(t, bad, protocol.Iter)
					stale := protocol.JobFrame(protocol.Partial, "a", next.Job.JobKey)
					stale.Job.Iteration--
					stale.Job.Rows, stale.Job.Counts, stale.Job.Sums = 2, []int{1, 1}, []Point{{0, 0}, {2, 2}}
					writeFrame(t, bad, stale)
				case "future_iteration":
					part.Job.Iteration++
					writeFrame(t, bad, part)
				case "wrong_execution":
					part.Job.ID = "another-job"
					writeFrame(t, bad, part)
				case "wrong_partition":
					part.Job.Partition++
					writeFrame(t, bad, part)
				case "wrong_attempt":
					// Encode a valid frame and change only the attempt, bypassing sender validation.
					var frame bytes.Buffer
					if err := protocol.Write(&frame, part); err != nil {
						t.Fatal(err)
					}
					wire := bytes.Replace(frame.Bytes(), []byte(`"attempt":1`), []byte(`"attempt":2`), 1)
					if _, err := bad.Write(wire); err != nil {
						t.Fatal(err)
					}
				case "incomplete_counts":
					part.Job.Counts[0]--
					writeFrame(t, bad, part)
				case "incompatible_dimensions":
					part.Job.Counts, part.Job.Sums = part.Job.Counts[:1], part.Job.Sums[:1]
					writeFrame(t, bad, part)
				case "disconnect_iteration":
					bad.Close()
				case "cancel_context":
					cancelJob()
				case "coordinator_interruption":
					joinService()
				case "silent_node":
					// Keep the socket open; the coordinator's read deadline must fire.
				}
			}
			out := within(t, done)
			c.mu.Lock()
			remaining := len(c.active)
			c.mu.Unlock()
			if remaining != 0 {
				t.Fatalf("failed job returned with %d participant registrations still active", remaining)
			}
			t.Logf("resultado: status=%s updates=%d reason=%s", out.result.Status, out.result.Iterations, out.result.Reason)
			if out.err == nil || out.result.Status != "failed" || !strings.Contains(out.result.Reason, tc.marker) ||
				len(out.result.Centroids) != 0 || len(out.result.Counts) != 0 || out.result.Iterations != tc.updates {
				t.Fatalf("unexpected failure result %+v, error %v", out.result, out.err)
			}
			if tc.name != "disconnect_transfer" && tc.name != "disconnect_iteration" {
				if _, err := protocol.Read(bad); !errors.Is(err, io.EOF) {
					t.Fatalf("expected session closure, got %v", err)
				}
			}
			if err := within(t, good); err == nil {
				t.Fatal("other participant unexpectedly succeeded")
			}
			t.Logf("nodo b:\n%s", goodLog.String())

			if !serviceStopped {
				// Same IDs, fresh sockets and the same serving coordinator.
				var nextA, nextB bytes.Buffer
				a := startFaultNode(ctx, addr, "a", &nextA)
				b := startFaultNode(ctx, addr, "b", &nextB)
				next, err := c.Run(ctx, cfg)
				if err != nil || next.Status != "completed" || next.Counts[0]+next.Counts[1] != 4 {
					t.Fatalf("next job failed: %+v (%v)", next, err)
				}
				for _, stream := range []<-chan error{a, b} {
					if err := within(t, stream); err != nil {
						t.Fatal(err)
					}
				}
				t.Logf("nueva ejecucion: status=%s nodos=%v counts=%v", next.Status, next.Nodes, next.Counts)
			}
		})
	}
}

func startFaultNode(ctx context.Context, addr, id string, output *bytes.Buffer) <-chan error {
	done := make(chan error, 1)
	go func() {
		done <- RunNode(ctx, NodeConfig{Address: addr, ID: id, Capacity: 1, Interval: 10 * time.Millisecond, Timeout: 4 * time.Second}, log.New(output, "", 0))
	}()
	return done
}

func TestServeStopsIdleAssignment(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serveCtx, stop := context.WithCancel(context.Background())
	defer stop()
	jobCtx, cancelJob := context.WithCancel(context.Background())
	defer cancelJob()
	c := NewCoordinator(time.Second, quiet)
	served := make(chan error, 1)
	go func() { served <- c.Serve(serveCtx, listener) }()
	conn := dial(t, listener.Addr().String())
	reply(t, conn, register("a"), protocol.New(protocol.Registered, "a"))
	work, err := c.claim(jobCtx, 1)
	if err != nil {
		t.Fatal(err)
	}
	assigned := make(chan error, 1)
	go func() {
		msg := protocol.JobFrame(protocol.Assign, "a", protocol.JobKey{ID: "idle", Attempt: 1})
		msg.Job.Rows, msg.Job.K = 2, 2
		_, err := work[0].call(msg, protocol.Assigned)
		assigned <- err
	}()
	m := beginAssignment(t, conn, "a")
	ackJob(t, conn, m, protocol.Assigned, 0)
	if err := within(t, assigned); err != nil {
		t.Fatal(err)
	}
	// No job request is in flight and the job context is deliberately still live.
	stop()
	select {
	case err := <-served:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		cancelJob()
		within(t, served)
		t.Fatal("closing the listener left an idle assignment handler waiting on the job context")
	}
	select {
	case <-work[0].peer.done:
	default:
		t.Fatal("Serve returned without joining the assignment handler")
	}
}
