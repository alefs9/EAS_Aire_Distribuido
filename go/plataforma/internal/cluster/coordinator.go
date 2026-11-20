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
	var workers sync.WaitGroup
	var mu sync.Mutex
	active := make(map[string]struct{})
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

			reserve := func(id string) bool {
				mu.Lock()
				defer mu.Unlock()
				if _, exists := active[id]; exists {
					return false
				}
				active[id] = struct{}{}
				return true
			}
			release := func(id string) {
				mu.Lock()
				delete(active, id)
				mu.Unlock()
			}
			serveNode(conn, timeout, logger, reserve, release)
		}(conn)
	}
}

func serveNode(conn net.Conn, timeout time.Duration, logger *log.Logger, reserve func(string) bool, release func(string)) {
	var id string
	reject := func(code, detail string) {
		logger.Printf("ERROR peer=%s node_id=%q code=%s", conn.RemoteAddr(), id, code)
		reply := protocol.New(protocol.Error, id)
		reply.Code, reply.Detail = code, detail
		_ = send(conn, timeout, reply)
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
			if err := send(conn, timeout, reply); err != nil {
				logger.Printf("HEARTBEAT_ACK failed node_id=%s: %v", id, err)
				return
			}
			logger.Printf("HEARTBEAT_ACK node_id=%s ts=%d", id, m.TS)
		case protocol.Bye:
			release(id)
			reply := protocol.New(protocol.ByeAck, id)
			id = ""
			if err := send(conn, timeout, reply); err != nil {
				logger.Printf("BYE_ACK failed node_id=%s: %v", reply.NodeID, err)
				return
			}
			logger.Printf("BYE_ACK node_id=%s", reply.NodeID)
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
