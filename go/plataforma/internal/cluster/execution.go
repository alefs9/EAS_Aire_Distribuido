package cluster

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"time"

	"eas-aire/plataforma/internal/protocol"
)

type ExecutionConfig struct {
	File          string  `json:"file"`
	Nodes         int     `json:"nodes"`
	K             int     `json:"k"`
	Seed          int64   `json:"seed"`
	Tolerance     float64 `json:"tolerance"`
	MaxIterations int     `json:"max_iterations"`
}

type Timings struct {
	Load         time.Duration `json:"load_ns"`
	WaitForNodes time.Duration `json:"wait_for_nodes_ns"`
	Distribution time.Duration `json:"distribution_ns"`
	Iterations   time.Duration `json:"iterations_ns"`
	Total        time.Duration `json:"total_ns"`
}

type Result struct {
	ID               string          `json:"job_id"`
	Config           ExecutionConfig `json:"config"`
	Rows             int             `json:"rows"`
	Nodes            []string        `json:"participants,omitempty"`
	Iterations       int             `json:"iterations"`
	Status           string          `json:"status"`
	Reason           string          `json:"reason"`
	Centroids        []Point         `json:"centroids,omitempty"`
	Counts           []int           `json:"counts,omitempty"`
	InitialCentroids []Point         `json:"initial_centroids,omitempty"`
	Partitions       []Partition     `json:"partitions,omitempty"`
	Timings          Timings         `json:"timings"`
}

type Partition struct {
	Node  string `json:"node_id"`
	ID    int    `json:"id"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type session struct {
	conn    net.Conn
	done    chan struct{}
	jobs    chan *assignment
	retired bool // protected by Coordinator.mu
}

type assignment struct {
	ctx      context.Context
	id       string
	peer     *session
	requests chan request
}

type request struct {
	message  protocol.Message
	expected string
	reply    chan response
}

type response struct {
	job *protocol.JobMessage
	err error
}

// A session handler is the only reader and writer on its coordinator socket.
func (a *assignment) serve(serveCtx context.Context, conn net.Conn, timeout time.Duration, closeSession func() error) error {
	for {
		select {
		case <-serveCtx.Done():
			return serveCtx.Err()
		case <-a.ctx.Done():
			return a.ctx.Err()
		case req := <-a.requests:
			job, err := exchangeJob(conn, timeout, req)
			if err == nil && req.message.Type == protocol.Finish {
				bye, readErr := receive(conn, timeout)
				switch {
				case readErr != nil:
					err = readErr
				case bye.Type != protocol.Bye || bye.NodeID != a.id:
					err = fmt.Errorf("expected BYE after FINISHED")
				default:
					err = closeSession()
				}
			}
			req.reply <- response{job, err}
			if err != nil || req.message.Type == protocol.Finish {
				return err
			}
		}
	}
}

func exchangeJob(conn net.Conn, timeout time.Duration, req request) (*protocol.JobMessage, error) {
	if err := send(conn, timeout, req.message); err != nil {
		return nil, err
	}
	m, err := receive(conn, timeout)
	if err != nil {
		return nil, err
	}
	return matchResponse(req, m)
}

func matchResponse(req request, m protocol.Message) (*protocol.JobMessage, error) {
	if m.Type != req.expected || m.NodeID != req.message.NodeID || m.Job == nil ||
		m.Job.JobKey != req.message.Job.JobKey {
		return nil, fmt.Errorf("invalid response to %s: expected %s with job key %+v, got %s", req.message.Type, req.expected, req.message.Job.JobKey, m.Type)
	}
	return m.Job, nil
}

func (a *assignment) call(m protocol.Message, expected string) (*protocol.JobMessage, error) {
	req := request{m, expected, make(chan response, 1)}
	select {
	case <-a.ctx.Done():
		return nil, a.ctx.Err()
	case <-a.peer.done:
		return nil, fmt.Errorf("node disconnected")
	case a.requests <- req:
	}
	select {
	case <-a.ctx.Done():
		return nil, a.ctx.Err()
	case <-a.peer.done:
		select {
		case r := <-req.reply:
			return r.job, r.err
		default:
			return nil, fmt.Errorf("node disconnected")
		}
	case r := <-req.reply:
		return r.job, r.err
	}
}

func distribute(a *assignment, key protocol.JobKey, points []Point, k int) error {
	assign := protocol.JobFrame(protocol.Assign, a.id, key)
	assign.Job.Rows, assign.Job.K = len(points), k
	if _, err := a.call(assign, protocol.Assigned); err != nil {
		return err
	}
	for offset := 0; offset < len(points); {
		end := min(offset+protocol.BatchSize, len(points))
		batch := protocol.JobFrame(protocol.Vectors, a.id, key)
		batch.Job.Offset, batch.Job.Points = offset, points[offset:end]
		ack, err := a.call(batch, protocol.VectorsAck)
		if err != nil {
			return err
		}
		if ack.Offset != end {
			return fmt.Errorf("batch acknowledgment offset %d, expected %d", ack.Offset, end)
		}
		offset = end
	}
	ready, err := a.call(protocol.JobFrame(protocol.LoadDone, a.id, key), protocol.Ready)
	if err != nil {
		return err
	}
	if ready.Rows != len(points) {
		return fmt.Errorf("node loaded %d rows, expected %d", ready.Rows, len(points))
	}
	return nil
}

func (c *Coordinator) claim(ctx context.Context, n int) ([]*assignment, error) {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		c.mu.Lock()
		ids := make([]string, 0, len(c.active))
		for id, peer := range c.active {
			if !peer.retired {
				ids = append(ids, id)
			}
		}
		if len(ids) >= n {
			sort.Strings(ids)
			work := make([]*assignment, n)
			for i, id := range ids[:n] {
				peer := c.active[id]
				work[i] = &assignment{ctx, id, peer, make(chan request)}
				peer.jobs <- work[i]
			}
			c.mu.Unlock()
			return work, nil
		}
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("waiting for %d nodes: %w", n, ctx.Err())
		case <-ticker.C:
		}
	}
}

func runPhase(ctx context.Context, work []*assignment, run func(int, *assignment) error) error {
	replies := make(chan error, len(work))
	for i, a := range work {
		go func(i int, a *assignment) {
			err := run(i, a)
			if err != nil {
				err = fmt.Errorf("node %s: %w", a.id, err)
			}
			replies <- err
		}(i, a)
	}
	var first error
	for range work {
		if err := <-replies; err != nil && first == nil {
			first = err
			// Closing all participating sockets wakes peers blocked in this phase.
			for _, a := range work {
				a.peer.conn.Close()
			}
		}
	}
	if first != nil {
		return first
	}
	return ctx.Err()
}

func validatePartial(p *protocol.JobMessage, rows, k int) error {
	if p.Rows != rows || len(p.Counts) != k || len(p.Sums) != k {
		return fmt.Errorf("partial dimensions do not match assignment")
	}
	remaining := rows
	for i, n := range p.Counts {
		sum := p.Sums[i]
		if n < 0 || n > remaining || !finite(sum[0]) || !finite(sum[1]) || (n == 0 && sum != (Point{})) {
			return fmt.Errorf("invalid partial for cluster %d", i)
		}
		remaining -= n
	}
	if remaining != 0 {
		return fmt.Errorf("partial counts do not cover partition")
	}
	return nil
}

// Run executes one job using registered TCP sessions. Use a context deadline to
// limit the entire job, including the wait for the required number of nodes.
func (c *Coordinator) Run(ctx context.Context, config ExecutionConfig) (result Result, err error) {
	start := time.Now()
	result.Config = config
	defer func() {
		result.Timings.Total = time.Since(start)
		if err != nil {
			result.Status, result.Reason = "failed", err.Error()
		}
	}()
	if config.Nodes < 1 || config.K < 1 || config.K > protocol.MaxClusters || config.MaxIterations < 1 ||
		config.MaxIterations == int(^uint(0)>>1) || config.Tolerance < 0 || !finite(config.Tolerance) {
		return result, fmt.Errorf("invalid execution configuration")
	}
	c.mu.Lock()
	if c.busy {
		c.mu.Unlock()
		return result, fmt.Errorf("coordinator already running a job")
	}
	c.busy = true
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		c.busy = false
		c.mu.Unlock()
	}()

	var token [16]byte
	if _, err := rand.Read(token[:]); err != nil {
		return result, err
	}
	jobID := hex.EncodeToString(token[:])
	result.ID = jobID
	loadStart := time.Now()
	points, err := loadCSV(ctx, config.File)
	result.Timings.Load = time.Since(loadStart)
	if err != nil {
		return result, fmt.Errorf("load vectors: %w", err)
	}
	result.Rows = len(points)
	if config.K > len(points) || config.Nodes > len(points) {
		return result, fmt.Errorf("K and nodes must not exceed the %d loaded rows", len(points))
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	waitStart := time.Now()
	work, err := c.claim(ctx, config.Nodes)
	result.Timings.WaitForNodes = time.Since(waitStart)
	if err != nil {
		return result, err
	}
	defer func() {
		if err != nil {
			cancel()
			c.mu.Lock()
			for _, a := range work {
				a.peer.retired = true
			}
			c.mu.Unlock()
			for _, a := range work {
				a.peer.conn.Close()
			}
			for _, a := range work {
				<-a.peer.done
			}
		}
	}()
	for _, a := range work {
		result.Nodes = append(result.Nodes, a.id)
	}
	ranges := partition(len(points), len(work))
	for i, p := range ranges {
		result.Partitions = append(result.Partitions, Partition{Node: work[i].id, ID: i, Start: p.start, End: p.end})
	}
	key := func(part, iteration int) protocol.JobKey {
		return protocol.JobKey{ID: jobID, Iteration: iteration, Partition: part, Attempt: 1}
	}

	distributionStart := time.Now()
	err = runPhase(ctx, work, func(i int, a *assignment) error {
		p := ranges[i]
		return distribute(a, key(i, 0), points[p.start:p.end], config.K)
	})
	result.Timings.Distribution = time.Since(distributionStart)
	if err != nil {
		return result, fmt.Errorf("distribution: %w", err)
	}

	calculationStart := time.Now()
	defer func() {
		if result.Timings.Iterations == 0 {
			result.Timings.Iterations = time.Since(calculationStart)
		}
	}()
	centers := initialize(points, config.K, config.Seed)
	result.InitialCentroids = centers
	parts := make([]*protocol.JobMessage, len(work))
	iterate := func(iteration int) error {
		return runPhase(ctx, work, func(i int, a *assignment) error {
			msg := protocol.JobFrame(protocol.Iter, a.id, key(i, iteration))
			msg.Job.Points = centers
			p, err := a.call(msg, protocol.Partial)
			if err != nil {
				return err
			}
			if err := validatePartial(p, ranges[i].end-ranges[i].start, config.K); err != nil {
				return err
			}
			parts[i] = p
			return nil
		})
	}
	reason := "max_iterations"
	for result.Iterations < config.MaxIterations {
		if err := iterate(result.Iterations + 1); err != nil {
			return result, fmt.Errorf("iteration %d: %w", result.Iterations+1, err)
		}
		next, _, movement, err := reduce(centers, parts)
		if err != nil {
			return result, err
		}
		centers = next
		result.Iterations++
		c.logger.Printf("iteration job_id=%s number=%d movement=%g", jobID, result.Iterations, movement)
		if movement < config.Tolerance {
			reason = "converged"
			break
		}
	}

	// Match the reference by counting assignments against the final centroids.
	final := result.Iterations + 1
	if err := iterate(final); err != nil {
		return result, fmt.Errorf("final assignment: %w", err)
	}
	counts := make([]int, config.K)
	for _, part := range parts {
		for i, count := range part.Counts {
			counts[i] += count
		}
	}
	result.Timings.Iterations = time.Since(calculationStart)
	if err := runPhase(ctx, work, func(i int, a *assignment) error {
		_, err := a.call(protocol.JobFrame(protocol.Finish, a.id, key(i, final)), protocol.Finished)
		return err
	}); err != nil {
		return result, fmt.Errorf("finish: %w", err)
	}
	result.Centroids, result.Counts = centers, counts
	result.Status, result.Reason = "completed", reason
	return result, nil
}
