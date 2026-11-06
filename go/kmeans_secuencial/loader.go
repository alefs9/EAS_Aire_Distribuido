package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

// LoadCSV lee un archivo CSV con las columnas PM10 y PM2.5.
func LoadCSV(path string) (Dataset, error) {
	file, err := os.Open(path)
	if err != nil {
		return Dataset{}, fmt.Errorf("no se pudo abrir el archivo: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return Dataset{}, fmt.Errorf("no se pudo leer el encabezado: %w", err)
	}

	if len(header) != 2 || header[0] != "PM10" || header[1] != "PM2.5" {
		return Dataset{}, fmt.Errorf("formato de CSV incorrecto: se esperaban PM10 y PM2.5")
	}

	points := make([][]float64, 0)

	for row := 2; ; row++ {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			return Dataset{}, fmt.Errorf("error al leer la fila %d: %w", row, err)
		}

		if len(record) != 2 {
			return Dataset{}, fmt.Errorf("la fila %d debe tener exactamente 2 columnas", row)
		}

		pm10, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			return Dataset{}, fmt.Errorf("valor PM10 inválido en la fila %d: %w", row, err)
		}

		pm25, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return Dataset{}, fmt.Errorf("valor PM2.5 inválido en la fila %d: %w", row, err)
		}

		if math.IsNaN(pm10) || math.IsInf(pm10, 0) ||
			math.IsNaN(pm25) || math.IsInf(pm25, 0) {
			return Dataset{}, fmt.Errorf("se encontró un valor no finito en la fila %d", row)
		}

		points = append(points, []float64{pm10, pm25})
	}

	if len(points) == 0 {
		return Dataset{}, fmt.Errorf("el CSV no contiene datos")
	}

	return NewDataset(points), nil
}
