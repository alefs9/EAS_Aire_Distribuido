// Package plataforma exposes the distributed K-Means coordinator to Go callers.
package plataforma

import (
	"log"
	"time"

	"eas-aire/plataforma/internal/cluster"
)

type Coordinator = cluster.Coordinator
type ExecutionConfig = cluster.ExecutionConfig
type Result = cluster.Result
type Timings = cluster.Timings
type Point = cluster.Point

func NewCoordinator(timeout time.Duration, logger *log.Logger) *Coordinator {
	return cluster.NewCoordinator(timeout, logger)
}
