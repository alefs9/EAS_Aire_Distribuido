package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"eas-aire/cargar-datos/internal/dataset"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run ./cmd/cargar-datos <archivo.csv>")
		os.Exit(1)
	}

	path := os.Args[1]

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
	fmt.Println("Validando registros...")

	var read, accepted, rejected int64
	var pm10, pm25 int64
	var errorCount int

	for {
		record, err := csvReader.Read()

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			read++
			rejected++

			if errorCount < 10 {
				fmt.Printf("Registro %d: error de lectura: %v\n", read, err)
				errorCount++
			}
			continue
		}

		read++

		if err := dataset.ValidateRecord(record, csvReader.Columns); err != nil {
			rejected++

			if errorCount < 10 {
				fmt.Printf("Registro %d: %v\n", read, err)
				errorCount++
			}
			continue
		}

		accepted++

		switch record[csvReader.Columns["contaminante"]] {
		case "PM10":
			pm10++
		case "PM2.5":
			pm25++
		}
	}

	fmt.Println("\n--- RESULTADOS ---")
	fmt.Printf("Registros leídos: %d\n", read)
	fmt.Printf("Registros aceptados: %d\n", accepted)
	fmt.Printf("Registros rechazados: %d\n", rejected)
	fmt.Printf("Mediciones PM10 aceptadas: %d\n", pm10)
	fmt.Printf("Mediciones PM2.5 aceptadas: %d\n", pm25)
	fmt.Printf("Conteos reconciliados: %v\n", read == accepted+rejected)
}
