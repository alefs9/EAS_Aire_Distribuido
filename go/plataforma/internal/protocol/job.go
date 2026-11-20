package protocol

import (
	"encoding/json"
	"fmt"
)

const (
	Assign     = "ASSIGN"
	Assigned   = "ASSIGNED"
	Vectors    = "VECTORS"
	VectorsAck = "VECTORS_ACK"
	LoadDone   = "LOAD_DONE"
	Ready      = "READY"
	Iter       = "ITER"
	Partial    = "PARTIAL"
	Finish     = "FINISH"
	Finished   = "FINISHED"

	BatchSize   = 4096
	MaxClusters = 8192
)

type Point [2]float64

func (p *Point) UnmarshalJSON(data []byte) error {
	var coordinates []*float64
	if err := json.Unmarshal(data, &coordinates); err != nil {
		return err
	}
	if len(coordinates) != 2 || coordinates[0] == nil || coordinates[1] == nil {
		return fmt.Errorf("expected a point with two numeric coordinates")
	}
	*p = Point{*coordinates[0], *coordinates[1]}
	return nil
}

type JobKey struct {
	ID        string `json:"id"`
	Iteration int    `json:"iteration"`
	Partition int    `json:"partition"`
	Attempt   int    `json:"attempt"`
}

type JobMessage struct {
	JobKey
	Rows   int     `json:"rows,omitempty"`
	K      int     `json:"k,omitempty"`
	Offset int     `json:"offset,omitempty"`
	Points []Point `json:"points,omitempty"`
	Counts []int   `json:"counts,omitempty"`
	Sums   []Point `json:"sums,omitempty"`
}

func (m Message) validateJob() error {
	j := m.Job
	if j == nil || !ValidID(j.ID) || j.Partition < 0 || j.Iteration < 0 || j.Attempt != 1 ||
		m.TS != 0 || m.Capacity != 0 || m.Work {
		return ErrMessage
	}
	switch m.Type {
	case Assign:
		if j.Iteration == 0 && j.Rows > 0 && j.K > 0 && j.K <= MaxClusters {
			return nil
		}
	case Assigned, LoadDone, Ready, VectorsAck:
		if j.Iteration == 0 {
			return nil
		}
	case Vectors:
		if j.Iteration == 0 && j.Offset >= 0 && len(j.Points) > 0 && len(j.Points) <= BatchSize {
			return nil
		}
	case Iter:
		if j.Iteration > 0 && len(j.Points) > 0 && len(j.Points) <= MaxClusters {
			return nil
		}
	case Partial:
		if j.Iteration > 0 && j.Rows > 0 && len(j.Counts) > 0 && len(j.Counts) == len(j.Sums) && len(j.Counts) <= MaxClusters {
			return nil
		}
	case Finish, Finished:
		if j.Iteration > 0 {
			return nil
		}
	}
	return ErrMessage
}

func JobFrame(kind, node string, key JobKey) Message {
	m := New(kind, node)
	m.Job = &JobMessage{JobKey: key}
	return m
}
