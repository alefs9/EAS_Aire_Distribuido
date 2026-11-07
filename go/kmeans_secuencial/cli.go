package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"runtime"
	"time"
)

type execution struct {
	Period           string      `json:"period"`
	Input            string      `json:"input"`
	Mode             string      `json:"mode"`
	Vectors          int         `json:"vectors"`
	K                int         `json:"k"`
	MaxIterations    int         `json:"max_iterations"`
	Tolerance        float64     `json:"tolerance"`
	Seed             int64       `json:"seed"`
	WorkersRequested int         `json:"workers_requested"`
	WorkersEffective int         `json:"workers_effective"`
	GOMAXPROCS       int         `json:"gomaxprocs"`
	GoVersion        string      `json:"go_version"`
	InitialCentroids [][]float64 `json:"initial_centroids"`
	Centroids        [][]float64 `json:"centroids"`
	Counts           []int       `json:"counts"`
	Labels           []int       `json:"labels,omitempty"`
	LabelsSHA256     string      `json:"labels_sha256"`
	Iterations       int         `json:"iterations"`
	Converged        bool        `json:"converged"`
	Inertia          float64     `json:"inertia"`
	LoadSeconds      float64     `json:"load_seconds"`
	TrainSeconds     float64     `json:"train_seconds"`
	TotalSeconds     float64     `json:"total_seconds"`
}

func runCLI(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("kmeans", flag.ContinueOnError)
	input := flags.String("input", "", "CSV de características PM10,PM2.5")
	period := flags.String("period", "", "Periodo explícito: 1_hora o 24_horas")
	mode := flags.String("mode", "sequential", "sequential o concurrent")
	workers := flags.Int("workers", 4, "Goroutines solicitadas")
	k := flags.Int("k", 4, "Número de clústeres")
	maxIter := flags.Int("max-iter", 100, "Máximo de iteraciones")
	tol := flags.Float64("tol", .001, "Tolerancia del movimiento")
	seed := flags.Int64("seed", 42, "Semilla compartida")
	labels := flags.Bool("labels", false, "Incluir todas las asignaciones en JSON")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("no se admiten argumentos posicionales")
	}
	if *input == "" || (*period != "1_hora" && *period != "24_horas") {
		return fmt.Errorf("indique -input y -period=1_hora o 24_horas")
	}
	if *mode != "sequential" && *mode != "concurrent" {
		return fmt.Errorf("modo inválido: %s", *mode)
	}
	if *workers <= 0 {
		return fmt.Errorf("workers debe ser positivo")
	}
	start := time.Now()
	data, err := LoadCSV(*input)
	if err != nil {
		return err
	}
	load := time.Since(start)
	trainStart := time.Now()
	var result KMeansResult
	if *mode == "sequential" {
		result, err = RunKMeans(data, *k, *maxIter, *tol, *seed)
	} else {
		result, err = RunKMeansConcurrent(data, *k, *maxIter, *tol, *seed, *workers)
	}
	train := time.Since(trainStart)
	total := time.Since(start)
	if err != nil {
		return err
	}
	// Metadatos y serialización fuera del alcance medido de carga+modelo.
	initial, err := initializeCentroids(data, *k, *seed)
	if err != nil {
		return err
	}
	counts := make([]int, *k)
	h := sha256.New()
	var buf [8]byte
	for _, label := range result.Labels {
		counts[label]++
		binary.LittleEndian.PutUint64(buf[:], uint64(label))
		_, _ = h.Write(buf[:])
	}
	effective := 1
	if *mode == "concurrent" {
		effective = len(partitionRows(data.N, *workers))
	}
	report := execution{Period: *period, Input: *input, Mode: *mode, Vectors: data.N, K: *k, MaxIterations: *maxIter, Tolerance: *tol, Seed: *seed, WorkersRequested: *workers, WorkersEffective: effective, GOMAXPROCS: runtime.GOMAXPROCS(0), GoVersion: runtime.Version(), InitialCentroids: initial, Centroids: result.Centroids, Counts: counts, LabelsSHA256: hex.EncodeToString(h.Sum(nil)), Iterations: result.Iterations, Converged: result.Converged, Inertia: result.Inertia, LoadSeconds: load.Seconds(), TrainSeconds: train.Seconds(), TotalSeconds: total.Seconds()}
	if *labels {
		report.Labels = result.Labels
	}
	return json.NewEncoder(output).Encode(report)
}
