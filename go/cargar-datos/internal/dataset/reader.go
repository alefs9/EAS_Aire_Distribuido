package dataset

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// CSVReader mantiene el archivo abierto, el lector CSV
// y las posiciones de las columnas.
type CSVReader struct {
	file    *os.File
	reader  *csv.Reader
	Columns map[string]int
}

// normalize estandariza los nombres de las columnas.
func normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// OpenCSV abre el CSV y comprueba que tenga las columnas necesarias.
func OpenCSV(path string) (*CSVReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el CSV: %w", err)
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("no se pudo leer el encabezado: %w", err)
	}

	columns := make(map[string]int)

	for i, name := range header {
		key := normalize(name)

		if key == "" {
			file.Close()
			return nil, fmt.Errorf("se encontró un encabezado vacío")
		}

		if _, exists := columns[key]; exists {
			file.Close()
			return nil, fmt.Errorf("columna duplicada: %s", name)
		}

		columns[key] = i
	}

	required := []string{
		"contaminante",
		"concentracion",
		"unidad",
	}

	for _, field := range required {
		if _, ok := columns[field]; !ok {
			file.Close()
			return nil, fmt.Errorf("falta la columna necesaria: %s", field)
		}
	}

	return &CSVReader{
		file:    file,
		reader:  reader,
		Columns: columns,
	}, nil
}

// Read obtiene el siguiente registro del CSV.
func (r *CSVReader) Read() ([]string, error) {
	return r.reader.Read()
}

// Close cierra el archivo CSV.
func (r *CSVReader) Close() error {
	return r.file.Close()
}
