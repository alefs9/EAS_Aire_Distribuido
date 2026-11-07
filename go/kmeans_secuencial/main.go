package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

// Procesa un dataset y muestra sus estadísticas y resultados de K-Means.
func procesarDataset(nombre string, ruta string) {
	fmt.Println("\n========================================")
	fmt.Println("DATASET:", nombre)
	fmt.Println("========================================")

	fmt.Println("\nCargando datos...")
	startLoad := time.Now()

	dataset, err := LoadCSV(ruta)
	if err != nil {
		log.Fatal(err)
	}

	loadTime := time.Since(startLoad)

	fmt.Println("Vectores:", dataset.N)
	fmt.Println("Dimensiones:", dataset.D)

	// Estadísticas básicas de los datos.
	var sumaPM10, sumaPM25 float64
	var maxPM10, maxPM25 float64
	var cerosAmbos int

	for _, point := range dataset.Points {
		pm10 := point[0]
		pm25 := point[1]

		sumaPM10 += pm10
		sumaPM25 += pm25

		if pm10 > maxPM10 {
			maxPM10 = pm10
		}

		if pm25 > maxPM25 {
			maxPM25 = pm25
		}

		if pm10 == 0 && pm25 == 0 {
			cerosAmbos++
		}
	}

	fmt.Println("\n--- ESTADÍSTICAS DE LOS DATOS ---")
	fmt.Printf("Promedio PM10: %.4f\n", sumaPM10/float64(dataset.N))
	fmt.Printf("Promedio PM2.5: %.4f\n", sumaPM25/float64(dataset.N))
	fmt.Printf("Máximo PM10: %.4f\n", maxPM10)
	fmt.Printf("Máximo PM2.5: %.4f\n", maxPM25)
	fmt.Printf("Vectores con ambas concentraciones en cero: %d\n", cerosAmbos)

	// Preparamos los valores para calcular los percentiles.
	pm10Values := make([]float64, dataset.N)
	pm25Values := make([]float64, dataset.N)

	for i, point := range dataset.Points {
		pm10Values[i] = point[0]
		pm25Values[i] = point[1]
	}

	fmt.Println("\n--- PERCENTILES ---")
	fmt.Printf("PM10 P50: %.2f\n", calculatePercentile(pm10Values, 0.50))
	fmt.Printf("PM10 P90: %.2f\n", calculatePercentile(pm10Values, 0.90))
	fmt.Printf("PM10 P95: %.2f\n", calculatePercentile(pm10Values, 0.95))
	fmt.Printf("PM10 P99: %.2f\n", calculatePercentile(pm10Values, 0.99))

	fmt.Printf("PM2.5 P50: %.2f\n", calculatePercentile(pm25Values, 0.50))
	fmt.Printf("PM2.5 P90: %.2f\n", calculatePercentile(pm25Values, 0.90))
	fmt.Printf("PM2.5 P95: %.2f\n", calculatePercentile(pm25Values, 0.95))
	fmt.Printf("PM2.5 P99: %.2f\n", calculatePercentile(pm25Values, 0.99))

	// Ejecutamos K-Means secuencial.
	fmt.Println("\nEjecutando K-Means secuencial...")

	start := time.Now()

	result, err := RunKMeans(dataset, 4, 100, 0.001, 42)
	if err != nil {
		log.Fatal(err)
	}

	elapsed := time.Since(start)

	fmt.Println("\n--- RESULTADOS DE K-MEANS ---")
	fmt.Println("Iteraciones:", result.Iterations)
	fmt.Println("Convergencia:", result.Converged)
	fmt.Println("Inercia:", result.Inertia)
	fmt.Println("Tiempo de carga:", loadTime)
	fmt.Println("Tiempo de ejecución:", elapsed)

	fmt.Println("\nCentroides:")
	for i, centroid := range result.Centroids {
		fmt.Printf("Clúster %d: PM10=%.4f, PM2.5=%.4f\n",
			i, centroid[0], centroid[1])
	}

	counts := make([]int, 4)
	for _, label := range result.Labels {
		counts[label]++
	}

	fmt.Println("\nCantidad de puntos por clúster:")
	for i, count := range counts {
		fmt.Printf("Clúster %d: %d\n", i, count)
	}
}

// Calcula el percentil indicado sobre una copia de los valores.
func calculatePercentile(values []float64, percentile float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)

	sort.Float64s(sorted)

	index := int(float64(len(sorted)-1) * percentile)
	return sorted[index]
}

func main() {
	if len(os.Args) > 1 {
		if err := runCLI(os.Args[1:], os.Stdout); err != nil {
			log.Fatal(err)
		}
		return
	}
	ruta1Hora := "../../data/vectors/features_1_hora.csv"
	ruta24Horas := "../../data/vectors/features_24_horas.csv"

	procesarDataset("1 HORA", ruta1Hora)
	procesarDataset("24 HORAS", ruta24Horas)
}
