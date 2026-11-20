package cluster

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"eas-aire/plataforma/internal/protocol"
)

func receive(conn net.Conn, timeout time.Duration) (protocol.Message, error) {
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return protocol.Message{}, err
	}
	return protocol.Read(conn)
}

func runAssignment(ctx context.Context, conn net.Conn, config NodeConfig, logger *log.Logger) error {
	var key protocol.JobKey
	var points []Point
	var rows, k, iteration int
	ready := false
	for {
		m, err := receive(conn, config.Timeout)
		if err != nil {
			return err
		}
		if m.NodeID != config.ID || m.Job == nil {
			return fmt.Errorf("invalid job message from coordinator")
		}
		j := m.Job
		if key.ID == "" {
			if m.Type != protocol.Assign {
				return fmt.Errorf("expected ASSIGN")
			}
			key, rows, k = j.JobKey, j.Rows, j.K
		}
		if j.ID != key.ID || j.Partition != key.Partition || j.Attempt != key.Attempt {
			return fmt.Errorf("job identity changed")
		}
		reply := protocol.JobFrame("", config.ID, j.JobKey)
		switch {
		case m.Type == protocol.Assign && points == nil:
			points = make([]Point, 0)
			reply.Type = protocol.Assigned
		case m.Type == protocol.Vectors && !ready && j.Offset == len(points) && len(j.Points) <= rows-len(points):
			points = append(points, j.Points...)
			reply.Type, reply.Job.Offset = protocol.VectorsAck, len(points)
		case m.Type == protocol.LoadDone && !ready && len(points) == rows:
			ready = true
			reply.Type, reply.Job.Rows = protocol.Ready, rows
		case m.Type == protocol.Iter && ready && iteration == j.Iteration-1 && len(j.Points) == k:
			iteration = j.Iteration
			counts, sums, err := accumulate(ctx, points, j.Points)
			if err != nil {
				return err
			}
			reply.Type, reply.Job.Rows = protocol.Partial, rows
			reply.Job.Counts, reply.Job.Sums = counts, sums
		case m.Type == protocol.Finish && iteration > 0 && j.Iteration == iteration:
			reply.Type = protocol.Finished
		default:
			return fmt.Errorf("invalid %s for partition %d at iteration %d", m.Type, key.Partition, iteration)
		}
		if err := send(conn, config.Timeout, reply); err != nil {
			return err
		}
		if reply.Type != protocol.VectorsAck {
			logger.Printf("%s node_id=%s job_id=%s partition=%d iteration=%d rows=%d", reply.Type, config.ID, key.ID, key.Partition, j.Iteration, len(points))
		}
		if reply.Type == protocol.Finished {
			return nil
		}
	}
}
