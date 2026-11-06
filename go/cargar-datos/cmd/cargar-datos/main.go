package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"eas-aire/cargar-datos/internal/dataset"
)

func main() {
	workers := flag.Int("workers", 4, "Número de workers concurrentes")
	batchSize := flag.Int("batch", 1000, "Cantidad de registros por lote")
	mode := flag.String("mode", "concurrent", "Modo de ejecución: sequential o concurrent")

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Uso: go run ./cmd/cargar-datos [opciones] <archivo.csv>")
		os.Exit(1)
	}

	path := flag.Arg(0)

	csvReader, err := dataset.OpenCSV(path)
	if err != nil {
		fmt.Printf("Error de lectura: %v\n", err)
		os.Exit(1)
	}
	defer csvReader.Close()

	fmt.Println("CSV abierto correctamente.")
	fmt.Println("Columnas necesarias encontradas.")
	fmt.Println("Contaminantes esperados: PM10 y PM2.5.")
	fmt.Println("Unidad esperada: µg/m³.")
	fmt.Printf("Modo de ejecución: %s\n", *mode)

	start := time.Now()

	var stats dataset.Stats

	switch *mode {
	case "sequential":
		stats, err = dataset.ProcessCSVSequential(csvReader)

	case "concurrent":
		fmt.Printf("Workers: %d\n", *workers)
		fmt.Printf("Tamaño del lote: %d\n", *batchSize)

		stats, err = dataset.ProcessCSVConcurrent(
			csvReader,
			*workers,
			*batchSize,
		)

	default:
		fmt.Printf("Modo no reconocido: %s\n", *mode)
		os.Exit(1)
	}

	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("Error durante el procesamiento: %v\n", err)
		os.Exit(1)
	}

	for _, message := range stats.Errors {
		fmt.Println(message)
	}

	fmt.Println("\n--- RESULTADOS ---")
	fmt.Printf("Registros leídos: %d\n", stats.Read)
	fmt.Printf("Registros aceptados: %d\n", stats.Accepted)
	fmt.Printf("Registros rechazados: %d\n", stats.Rejected)
	fmt.Printf("Mediciones PM10 aceptadas: %d\n", stats.PM10)
	fmt.Printf("Mediciones PM2.5 aceptadas: %d\n", stats.PM25)
	fmt.Printf("Conteos reconciliados: %v\n",
		stats.Read == stats.Accepted+stats.Rejected)
	fmt.Printf("Tiempo de procesamiento: %v\n", elapsed)
}
