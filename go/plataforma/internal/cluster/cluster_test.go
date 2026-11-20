package cluster

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"testing"
	"time"

	"eas-aire/plataforma/internal/protocol"
)

var quiet = log.New(io.Discard, "", 0)

func start(t *testing.T, timeout time.Duration) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := Serve(ctx, listener, timeout, quiet); err != nil {
			t.Error(err)
		}
	}()
	stop := func() {
		cancel()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			t.Fatal("coordinator did not stop")
		}
	}
	t.Cleanup(stop)
	return listener.Addr().String(), stop
}

func dial(t *testing.T, address string) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	return conn
}

func register(id string) protocol.Message {
	msg := protocol.New(protocol.Register, id)
	msg.Capacity = 1
	return msg
}

func reply(t *testing.T, c net.Conn, request, want protocol.Message) {
	t.Helper()
	if err := protocol.Write(c, request); err != nil {
		t.Fatal(err)
	}
	got, err := protocol.Read(c)
	if err != nil || got != want {
		t.Fatalf("got %+v (%v), want %+v", got, err, want)
	}
}

func expectError(t *testing.T, conn net.Conn, code string) {
	t.Helper()
	msg, err := protocol.Read(conn)
	if err != nil || msg.Type != protocol.Error || msg.Code != code {
		t.Fatalf("got %+v (%v), want ERROR %s", msg, err, code)
	}
	if _, err := protocol.Read(conn); !errors.Is(err, io.EOF) {
		t.Fatalf("want EOF, got %v", err)
	}
}

func TestTwoNodesDuplicateAndGracefulReuse(t *testing.T) {
	address, _ := start(t, 3*time.Second)
	a, b, duplicate := dial(t, address), dial(t, address), dial(t, address)
	reply(t, a, register("a"), protocol.New(protocol.Registered, "a"))
	reply(t, b, register("b"), protocol.New(protocol.Registered, "b"))
	if err := protocol.Write(duplicate, register("a")); err != nil {
		t.Fatal(err)
	}
	expectError(t, duplicate, "DUPLICATE_NODE")

	for id, conn := range map[string]net.Conn{"a": a, "b": b} {
		request := protocol.New(protocol.Heartbeat, id)
		request.TS = 1234
		ack := protocol.New(protocol.HeartbeatAck, id)
		ack.TS = request.TS
		reply(t, conn, request, ack)
	}
	reply(t, a, protocol.New(protocol.Bye, "a"), protocol.New(protocol.ByeAck, "a"))
	reply(t, dial(t, address), register("a"), protocol.New(protocol.Registered, "a"))
}

func TestRejectsInvalidSessions(t *testing.T) {
	address, _ := start(t, 3*time.Second)
	for _, test := range []struct {
		code    string
		before  func(net.Conn)
		message protocol.Message
	}{
		{"EXPECTED_REGISTER", nil, protocol.New(protocol.Bye, "a")},
		{"NODE_MISMATCH", func(c net.Conn) { reply(t, c, register("a"), protocol.New(protocol.Registered, "a")) }, protocol.New(protocol.Bye, "b")},
		{"UNEXPECTED_MESSAGE", func(c net.Conn) { reply(t, c, register("b"), protocol.New(protocol.Registered, "b")) }, register("b")},
	} {
		t.Run(test.code, func(t *testing.T) {
			conn := dial(t, address)
			if test.before != nil {
				test.before(conn)
			}
			if err := protocol.Write(conn, test.message); err != nil {
				t.Fatal(err)
			}
			expectError(t, conn, test.code)
		})
	}
	conn := dial(t, address)
	if _, err := conn.Write([]byte{0, 0, 0, 1, '{'}); err != nil {
		t.Fatal(err)
	}
	expectError(t, conn, "INVALID_MESSAGE")
}

func TestAbruptDisconnectAndCoordinatorShutdown(t *testing.T) {
	address, stop := start(t, 3*time.Second)
	conn := dial(t, address)
	reply(t, conn, register("a"), protocol.New(protocol.Registered, "a"))
	conn.Close()

	deadline := time.Now().Add(time.Second)
	var reconnected net.Conn
	for {
		reconnected = dial(t, address)
		if err := protocol.Write(reconnected, register("a")); err != nil {
			t.Fatal(err)
		}
		got, err := protocol.Read(reconnected)
		if err != nil {
			t.Fatal(err)
		}
		if got.Type == protocol.Registered {
			break
		}
		reconnected.Close()
		if got.Code != "DUPLICATE_NODE" || time.Now().After(deadline) {
			t.Fatalf("could not reuse ID: %+v", got)
		}
		time.Sleep(time.Millisecond)
	}

	pending := dial(t, address)
	stop()
	for _, peer := range []net.Conn{reconnected, pending} {
		if _, err := protocol.Read(peer); !errors.Is(err, io.EOF) {
			t.Fatalf("want EOF after coordinator shutdown, got %v", err)
		}
	}
}

func TestTimeout(t *testing.T) {
	address, _ := start(t, 100*time.Millisecond)
	expectError(t, dial(t, address), "TIMEOUT")
}

func TestRunNode(t *testing.T) {
	address, _ := start(t, 3*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg := NodeConfig{Address: address, ID: "node", Capacity: 1, Interval: time.Millisecond, Timeout: time.Second, Checks: 2}
	if err := RunNode(ctx, cfg, quiet); err != nil {
		t.Fatal(err)
	}
	conn := dial(t, address)
	reply(t, conn, register("node"), protocol.New(protocol.Registered, "node"))
	if err := RunNode(ctx, cfg, quiet); err == nil || !strings.Contains(err.Error(), "DUPLICATE_NODE") {
		t.Fatalf("want duplicate error, got %v", err)
	}
}
