package cluster

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"eas-aire/plataforma/internal/protocol"
)

// Serve owns listener and all accepted connections until it returns.
func Serve(ctx context.Context, listener net.Listener, timeout time.Duration, logger *log.Logger) error {
	return NewCoordinator(timeout, logger).Serve(ctx, listener)
}

type Coordinator struct {
	timeout time.Duration
	logger  *log.Logger
	mu      sync.Mutex
	active  map[string]*session
	busy    bool
}

func NewCoordinator(timeout time.Duration, logger *log.Logger) *Coordinator {
	return &Coordinator{timeout: timeout, logger: logger, active: make(map[string]*session)}
}

// Serve accepts connections for this coordinator. Call it once per coordinator.
func (c *Coordinator) Serve(ctx context.Context, listener net.Listener) error {
	var workers sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		listener.Close()
		workers.Wait()
	}()
	stop := context.AfterFunc(ctx, func() { listener.Close() })
	defer stop()

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		workers.Add(1)
		go func(conn net.Conn) {
			defer workers.Done()
			defer conn.Close()
			stop := context.AfterFunc(ctx, func() { conn.Close() })
			defer stop()

			peer := &session{conn: conn, done: make(chan struct{}), jobs: make(chan *assignment, 1)}
			defer close(peer.done)
			reserve := func(id string) bool {
				c.mu.Lock()
				defer c.mu.Unlock()
				if _, exists := c.active[id]; exists {
					return false
				}
				c.active[id] = peer
				return true
			}
			release := func(id string) {
				c.mu.Lock()
				delete(c.active, id)
				c.mu.Unlock()
			}
			serveNode(peer, c.timeout, c.logger, reserve, release)
		}(conn)
	}
}

func serveNode(peer *session, timeout time.Duration, logger *log.Logger, reserve func(string) bool, release func(string)) {
	conn := peer.conn
	var id string
	reject := func(code, detail string) {
		logger.Printf("ERROR peer=%s node_id=%q code=%s", conn.RemoteAddr(), id, code)
		reply := protocol.New(protocol.Error, id)
		reply.Code, reply.Detail = code, detail
		_ = send(conn, timeout, reply)
	}
	closeSession := func() error {
		release(id)
		reply := protocol.New(protocol.ByeAck, id)
		id = ""
		if err := send(conn, timeout, reply); err != nil {
			logger.Printf("BYE_ACK failed node_id=%s: %v", reply.NodeID, err)
			return err
		}
		logger.Printf("BYE_ACK node_id=%s", reply.NodeID)
		return nil
	}
	for {
		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return
		}
		m, err := protocol.Read(conn)
		if err != nil {
			var netErr net.Error
			switch {
			case errors.Is(err, io.EOF), errors.Is(err, net.ErrClosed):
				return
			case errors.As(err, &netErr) && netErr.Timeout():
				reject("TIMEOUT", "read deadline exceeded")
			case errors.Is(err, protocol.ErrVersion):
				reject("UNSUPPORTED_VERSION", "expected proto 1")
			case errors.Is(err, protocol.ErrFrame), errors.Is(err, protocol.ErrJSON),
				errors.Is(err, protocol.ErrMessage), errors.Is(err, io.ErrUnexpectedEOF):
				reject("INVALID_MESSAGE", "invalid frame or message")
			default:
				logger.Printf("read failed peer=%s: %v", conn.RemoteAddr(), err)
			}
			return
		}

		if id == "" {
			if m.Type != protocol.Register {
				reject("EXPECTED_REGISTER", "first message must be REGISTER")
				return
			}
			if !reserve(m.NodeID) {
				reject("DUPLICATE_NODE", "node_id already registered")
				return
			}
			id = m.NodeID
			// A graceful BYE releases the ID before its acknowledgment.
			defer func(registeredID string) {
				if id != "" {
					release(registeredID)
				}
				logger.Printf("disconnected node_id=%s", registeredID)
			}(id)
			if err := send(conn, timeout, protocol.New(protocol.Registered, id)); err != nil {
				logger.Printf("REGISTERED failed node_id=%s: %v", id, err)
				return
			}
			logger.Printf("REGISTERED node_id=%s peer=%s capacity=%d", id, conn.RemoteAddr(), m.Capacity)
			continue
		}

		if m.NodeID != id {
			reject("NODE_MISMATCH", "node_id differs from registered ID")
			return
		}
		switch m.Type {
		case protocol.Heartbeat:
			reply := protocol.New(protocol.HeartbeatAck, id)
			reply.TS = m.TS
			var work *assignment
			select {
			case work = <-peer.jobs:
				reply.Work = true
			default:
			}
			if err := send(conn, timeout, reply); err != nil {
				logger.Printf("HEARTBEAT_ACK failed node_id=%s: %v", id, err)
				return
			}
			logger.Printf("HEARTBEAT_ACK node_id=%s ts=%d", id, m.TS)
			if work != nil {
				if err := work.serve(conn, timeout, closeSession); err != nil {
					logger.Printf("job session failed node_id=%s: %v", id, err)
				}
				return
			}
		case protocol.Bye:
			_ = closeSession()
			return
		default:
			reject("UNEXPECTED_MESSAGE", "expected HEARTBEAT or BYE")
			return
		}
	}
}

func send(conn net.Conn, timeout time.Duration, m protocol.Message) error {
	if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	return protocol.Write(conn, m)
}
