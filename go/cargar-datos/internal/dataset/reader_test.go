package dataset

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// createCSV crea un archivo CSV temporal para las pruebas.
func createCSV(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.csv")

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("no se pudo crear el CSV temporal: %v", err)
	}

	return path
}

func TestOpenCSVValid(t *testing.T) {
	content := "informe,contaminante,concentracion,unidad\n" +
		"Informe 1,PM10,25.5,µg/m³\n"

	path := createCSV(t, content)

	reader, err := OpenCSV(path)
	if err != nil {
		t.Fatalf("se esperaba abrir el CSV correctamente: %v", err)
	}
	defer reader.Close()

	expected := map[string]int{
		"informe":       0,
		"contaminante":  1,
		"concentracion": 2,
		"unidad":        3,
	}

	if !reflect.DeepEqual(reader.Columns, expected) {
		t.Errorf("columnas incorrectas: se obtuvo %v, se esperaba %v",
			reader.Columns, expected)
	}
}

func TestOpenCSVMissingRequiredColumn(t *testing.T) {
	content := "informe,contaminante,unidad\n" +
		"Informe 1,PM10,µg/m³\n"

	path := createCSV(t, content)

	reader, err := OpenCSV(path)

	if err == nil {
		if reader != nil {
			reader.Close()
		}
		t.Fatal("se esperaba un error por falta de la columna concentracion")
	}

	if !strings.Contains(err.Error(), "concentracion") {
		t.Errorf("mensaje de error inesperado: %v", err)
	}
}

func TestOpenCSVDuplicateColumn(t *testing.T) {
	content := "contaminante,concentracion,unidad,unidad\n"

	path := createCSV(t, content)

	_, err := OpenCSV(path)

	if err == nil {
		t.Fatal("se esperaba un error por columna duplicada")
	}

	if !strings.Contains(err.Error(), "duplicada") {
		t.Errorf("mensaje de error inesperado: %v", err)
	}
}

func TestOpenCSVEmptyHeader(t *testing.T) {
	content := "contaminante,concentracion,unidad,\n"

	path := createCSV(t, content)

	_, err := OpenCSV(path)

	if err == nil {
		t.Fatal("se esperaba un error por encabezado vacío")
	}
}

func TestReadCSVWithCommasAndQuotes(t *testing.T) {
	content := "informe,contaminante,concentracion,unidad\n" +
		"\"Informe, zona \"\"Norte\"\"\",PM10,25.5,µg/m³\n"

	path := createCSV(t, content)

	reader, err := OpenCSV(path)
	if err != nil {
		t.Fatalf("no se pudo abrir el CSV: %v", err)
	}
	defer reader.Close()

	record, err := reader.Read()
	if err != nil {
		t.Fatalf("no se pudo leer el registro: %v", err)
	}

	expected := "Informe, zona \"Norte\""

	if record[reader.Columns["informe"]] != expected {
		t.Errorf("texto incorrecto: se obtuvo %q, se esperaba %q",
			record[reader.Columns["informe"]], expected)
	}
}
