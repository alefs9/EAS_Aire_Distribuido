package main

import (
	"os"
	"path/filepath"
	"testing"
)

func createCSV(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.csv")

	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatalf("no se pudo crear el CSV: %v", err)
	}

	return path
}

func TestLoadCSVSuccess(t *testing.T) {
	path := createCSV(t, "PM10,PM2.5\n10.5,5.2\n20.0,8.5\n")

	data, err := LoadCSV(path)
	if err != nil {
		t.Fatalf("no se esperaba un error: %v", err)
	}

	if data.N != 2 {
		t.Errorf("se esperaban 2 puntos, se obtuvieron %d", data.N)
	}

	if data.D != 2 {
		t.Errorf("se esperaban 2 dimensiones, se obtuvieron %d", data.D)
	}

	if data.Points[0][0] != 10.5 || data.Points[0][1] != 5.2 {
		t.Errorf("valores incorrectos en el primer punto: %v", data.Points[0])
	}
}

func TestLoadCSVEmpty(t *testing.T) {
	path := createCSV(t, "PM10,PM2.5\n")

	_, err := LoadCSV(path)
	if err == nil {
		t.Fatal("se esperaba un error para un CSV sin registros")
	}
}

func TestLoadCSVInvalidHeader(t *testing.T) {
	path := createCSV(t, "PM10,NO2\n10,5\n")

	_, err := LoadCSV(path)
	if err == nil {
		t.Fatal("se esperaba un error por encabezado incorrecto")
	}
}

func TestLoadCSVInvalidValue(t *testing.T) {
	path := createCSV(t, "PM10,PM2.5\nabc,5\n")

	_, err := LoadCSV(path)
	if err == nil {
		t.Fatal("se esperaba un error por valor no numérico")
	}
}

func TestLoadCSVInvalidDimensions(t *testing.T) {
	path := createCSV(t, "PM10,PM2.5\n10\n")

	_, err := LoadCSV(path)
	if err == nil {
		t.Fatal("se esperaba un error por cantidad incorrecta de columnas")
	}
}
