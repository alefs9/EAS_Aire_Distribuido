package dataset

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// CSVReader mantiene la conexión con el archivo CSV,
// el lector encargado de interpretar sus registros
// y un mapa que relaciona los nombres de las columnas con sus posiciones.
type CSVReader struct {
	file    *os.File
	reader  *csv.Reader
	Columns map[string]int
}

// normalize estandariza los nombres de las columnas para facilitar
// su comparación, eliminando espacios y convirtiéndolos a minúsculas.
func normalize(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "")
	return name
}

// OpenCSV abre el archivo CSV, lee su encabezado y comprueba
// que existan las columnas necesarias para procesar las mediciones.
func OpenCSV(path string) (*CSVReader, error) {

	// Abre el archivo indicado en la ruta recibida.
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir el CSV: %w", err)
	}

	// Crea un lector CSV que interpreta correctamente
	// campos entre comillas y comas dentro de los textos.
	reader := csv.NewReader(file)

	// Permite que los registros tengan distinta cantidad de campos.
	// La validación de las columnas necesarias se realiza posteriormente.
	reader.FieldsPerRecord = -1

	// Elimina espacios iniciales en los campos.
	reader.TrimLeadingSpace = true

	// Lee la primera fila del archivo, que contiene los encabezados.
	header, err := reader.Read()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("no se pudo leer el encabezado: %w", err)
	}

	// Construye un mapa que relaciona cada nombre de columna
	// con su posición dentro de los registros.
	columns := make(map[string]int)

	for i, name := range header {
		key := normalize(name)

		// Evita que existan columnas con nombres duplicados.
		if _, exists := columns[key]; exists {
			file.Close()
			return nil, fmt.Errorf("columna duplicada: %s", name)
		}

		columns[key] = i
	}

	// Comprueba que exista la columna PM10.
	if _, ok := columns["pm10"]; !ok {
		file.Close()
		return nil, fmt.Errorf("falta la columna PM10")
	}

	// Busca la columna PM2.5.
	// Se aceptan distintas formas de escribir su encabezado.
	pm25Found := false

	for _, alias := range []string{"pm2.5", "pm25", "pm2_5"} {
		if _, ok := columns[alias]; ok {
			// Registra la posición encontrada usando un nombre estándar.
			columns["pm2.5"] = columns[alias]
			pm25Found = true
			break
		}
	}

	// Si no encuentra ninguna variante, el archivo no cumple
	// con la estructura mínima necesaria.
	if !pm25Found {
		file.Close()
		return nil, fmt.Errorf("falta la columna PM2.5")
	}

	// Devuelve el lector preparado para comenzar a procesar
	// los registros del archivo.
	return &CSVReader{
		file:    file,
		reader:  reader,
		Columns: columns,
	}, nil
}

// Read obtiene el siguiente registro del CSV.
// Devuelve los campos encontrados o un error si ocurre algún problema.
func (r *CSVReader) Read() ([]string, error) {
	return r.reader.Read()
}

// Close cierra el archivo CSV y libera el recurso utilizado.
func (r *CSVReader) Close() error {
	return r.file.Close()
}

// ReadAllRecords lee todos los registros del CSV y los almacena
// en una lista de registros.
func (r *CSVReader) ReadAllRecords() ([][]string, error) {
	var records [][]string

	for {
		// Lee el siguiente registro.
		record, err := r.Read()

		// io.EOF indica que se llegó al final del archivo.
		if err == io.EOF {
			break
		}

		// Si ocurre otro error, devuelve los registros
		// que alcanzó a leer y el error encontrado.
		if err != nil {
			return records, err
		}

		// Agrega el registro leído a la lista.
		records = append(records, record)
	}

	return records, nil
}
