package cluster

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"eas-aire/plataforma/internal/protocol"
)

type NodeConfig struct {
	Address  string
	ID       string
	Capacity int
	Interval time.Duration
	Timeout  time.Duration
	Checks   int
}

func RunNode(ctx context.Context, config NodeConfig, logger *log.Logger) error {
	if !protocol.ValidID(config.ID) || config.Capacity < 1 || config.Interval <= 0 || config.Timeout <= 0 || config.Checks < 0 {
		return fmt.Errorf("invalid node configuration")
	}
	dialer := net.Dialer{Timeout: config.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", config.Address)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close()

	var work bool
	exchange := func(m protocol.Message, expected string) error {
		if err := send(conn, config.Timeout, m); err != nil {
			return fmt.Errorf("%s: %w", m.Type, err)
		}
		if err := conn.SetReadDeadline(time.Now().Add(config.Timeout)); err != nil {
			return err
		}
		reply, err := protocol.Read(conn)
		if err != nil {
			return fmt.Errorf("waiting for %s: %w", expected, err)
		}
		if reply.Type == protocol.Error {
			return fmt.Errorf("coordinator %s: %s", reply.Code, reply.Detail)
		}
		if reply.Type != expected || reply.NodeID != m.NodeID || reply.TS != m.TS {
			return fmt.Errorf("unexpected reply to %s: %+v", m.Type, reply)
		}
		work = reply.Work
		logger.Printf("%s node_id=%s ts=%d", reply.Type, reply.NodeID, reply.TS)
		return nil
	}
	register := protocol.New(protocol.Register, config.ID)
	register.Capacity = config.Capacity
	if err := exchange(register, protocol.Registered); err != nil {
		return err
	}

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()
	for checks := 0; ctx.Err() == nil && (config.Checks == 0 || checks < config.Checks); checks++ {
		select {
		case <-ctx.Done():
			continue
		case <-ticker.C:
			heartbeat := protocol.New(protocol.Heartbeat, config.ID)
			heartbeat.TS = time.Now().UnixNano()
			if err := exchange(heartbeat, protocol.HeartbeatAck); err != nil {
				return err
			}
			if work {
				if err := runAssignment(ctx, conn, config, logger); err != nil {
					return fmt.Errorf("job failed: %w", err)
				}
				return exchange(protocol.New(protocol.Bye, config.ID), protocol.ByeAck)
			}
		}
	}
	return exchange(protocol.New(protocol.Bye, config.ID), protocol.ByeAck)
}
