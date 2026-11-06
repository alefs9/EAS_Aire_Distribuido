package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

func procesarDataset(nombre string, ruta string) error {
	fmt.Println("\n========================================")
	fmt.Println("DATASET:", nombre)
	fmt.Println("========================================")

	fmt.Println("\nCargando datos...")
	startLoad := time.Now()

	dataset, err := LoadCSV(ruta)
	if err != nil {
		return fmt.Errorf("error al cargar %s: %w", nombre, err)
	}

	loadTime := time.Since(startLoad)

	fmt.Println("Vectores:", dataset.N)
	fmt.Println("Dimensiones:", dataset.D)

	var sumaPM10, sumaPM25 float64
	var maxPM10, maxPM25 float64
	var cerosAmbos int

	pm10Values := make([]float64, dataset.N)
	pm25Values := make([]float64, dataset.N)

	for i, point := range dataset.Points {
		pm10 := point[0]
		pm25 := point[1]

		sumaPM10 += pm10
		sumaPM25 += pm25

		if i == 0 || pm10 > maxPM10 {
			maxPM10 = pm10
		}

		if i == 0 || pm25 > maxPM25 {
			maxPM25 = pm25
		}

		if pm10 == 0 && pm25 == 0 {
			cerosAmbos++
		}

		pm10Values[i] = pm10
		pm25Values[i] = pm25
	}

	fmt.Println("\n--- ESTADÍSTICAS DE LOS DATOS ---")
	fmt.Printf("Promedio PM10: %.4f\n", sumaPM10/float64(dataset.N))
	fmt.Printf("Promedio PM2.5: %.4f\n", sumaPM25/float64(dataset.N))
	fmt.Printf("Máximo PM10: %.4f\n", maxPM10)
	fmt.Printf("Máximo PM2.5: %.4f\n", maxPM25)
	fmt.Printf("Vectores con ambas concentraciones en cero: %d\n", cerosAmbos)

	fmt.Println("\n--- PERCENTILES ---")
	fmt.Printf("PM10 P50: %.2f\n", calculatePercentile(pm10Values, 0.50))
	fmt.Printf("PM10 P90: %.2f\n", calculatePercentile(pm10Values, 0.90))
	fmt.Printf("PM10 P95: %.2f\n", calculatePercentile(pm10Values, 0.95))
	fmt.Printf("PM10 P99: %.2f\n", calculatePercentile(pm10Values, 0.99))

	fmt.Printf("PM2.5 P50: %.2f\n", calculatePercentile(pm25Values, 0.50))
	fmt.Printf("PM2.5 P90: %.2f\n", calculatePercentile(pm25Values, 0.90))
	fmt.Printf("PM2.5 P95: %.2f\n", calculatePercentile(pm25Values, 0.95))
	fmt.Printf("PM2.5 P99: %.2f\n", calculatePercentile(pm25Values, 0.99))

	fmt.Println("\nEjecutando K-Means secuencial...")

	start := time.Now()

	result, err := RunKMeans(dataset, 4, 100, 0.001, 42)
	if err != nil {
		return fmt.Errorf("error al ejecutar K-Means: %w", err)
	}

	elapsed := time.Since(start)

	fmt.Println("\n--- RESULTADOS DE K-MEANS ---")
	fmt.Println("Iteraciones:", result.Iterations)
	fmt.Println("Convergencia:", result.Converged)
	fmt.Printf("Inercia: %.6f\n", result.Inertia)
	fmt.Println("Tiempo de carga:", loadTime)
	fmt.Println("Tiempo de ejecución:", elapsed)

	fmt.Println("\nCentroides:")
	for i, centroid := range result.Centroids {
		fmt.Printf("Clúster %d: PM10=%.4f, PM2.5=%.4f\n",
			i, centroid[0], centroid[1])
	}

	counts := make([]int, len(result.Centroids))

	for _, label := range result.Labels {
		counts[label]++
	}

	fmt.Println("\nCantidad de puntos por clúster:")
	for i, count := range counts {
		fmt.Printf("Clúster %d: %d\n", i, count)
	}

	return nil
}

// calculatePercentile calcula el percentil indicado
// sobre una copia ordenada de los valores.
func calculatePercentile(values []float64, percentile float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)

	sort.Float64s(sorted)

	index := int(float64(len(sorted)-1) * percentile)
	return sorted[index]
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run . <archivo.csv> [archivo24h.csv]")
		os.Exit(1)
	}

	if err := procesarDataset("DATASET", os.Args[1]); err != nil {
		log.Fatal(err)
	}

	if len(os.Args) >= 3 {
		if err := procesarDataset("SEGUNDO DATASET", os.Args[2]); err != nil {
			log.Fatal(err)
		}
	}
}
