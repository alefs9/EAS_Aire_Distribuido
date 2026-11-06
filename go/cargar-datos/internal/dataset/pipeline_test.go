package dataset

import (
	"os"
	"reflect"
	"testing"
)

func TestProcessCSVSequential(t *testing.T) {
	content := "contaminante,concentracion,unidad\n" +
		"PM10,25.5,µg/m³\n" +
		"PM2.5,12.8,µg/m³\n" +
		"CO2,10,µg/m³\n" +
		"PM10,-5,µg/m³\n"

	path := createCSV(t, content)

	reader, err := OpenCSV(path)
	if err != nil {
		t.Fatalf("no se pudo abrir el CSV: %v", err)
	}
	defer reader.Close()

	stats, err := ProcessCSVSequential(reader)
	if err != nil {
		t.Fatalf("error en procesamiento secuencial: %v", err)
	}

	if stats.Read != 4 || stats.Accepted != 2 || stats.Rejected != 2 {
		t.Errorf("conteos inesperados: %+v", stats)
	}

	if stats.PM10 != 1 || stats.PM25 != 1 {
		t.Errorf("conteos por contaminante inesperados: %+v", stats)
	}
}

func TestProcessCSVConcurrent(t *testing.T) {
	content := "contaminante,concentracion,unidad\n" +
		"PM10,25.5,µg/m³\n" +
		"PM2.5,12.8,µg/m³\n" +
		"CO2,10,µg/m³\n" +
		"PM10,-5,µg/m³\n"

	path := createCSV(t, content)

	reader, err := OpenCSV(path)
	if err != nil {
		t.Fatalf("no se pudo abrir el CSV: %v", err)
	}
	defer reader.Close()

	stats, err := ProcessCSVConcurrent(reader, 2, 2)
	if err != nil {
		t.Fatalf("error en procesamiento concurrente: %v", err)
	}

	if stats.Read != 4 || stats.Accepted != 2 || stats.Rejected != 2 {
		t.Errorf("conteos inesperados: %+v", stats)
	}

	if stats.PM10 != 1 || stats.PM25 != 1 {
		t.Errorf("conteos por contaminante inesperados: %+v", stats)
	}
}

func TestSequentialAndConcurrentResultsMatch(t *testing.T) {
	content := "contaminante,concentracion,unidad\n" +
		"PM10,25.5,µg/m³\n" +
		"PM2.5,12.8,µg/m³\n" +
		"CO2,10,µg/m³\n" +
		"PM10,-5,µg/m³\n" +
		"PM2.5,8.2,µg/m³\n"

	path := createCSV(t, content)

	sequentialReader, err := OpenCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	defer sequentialReader.Close()

	concurrentReader, err := OpenCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	defer concurrentReader.Close()

	sequential, err := ProcessCSVSequential(sequentialReader)
	if err != nil {
		t.Fatal(err)
	}

	concurrent, err := ProcessCSVConcurrent(concurrentReader, 4, 2)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(sequential, concurrent) {
		t.Errorf("los resultados no coinciden:\nsecuencial: %+v\nconcurrente: %+v",
			sequential, concurrent)
	}
}

func TestProcessCSVConcurrentInvalidParameters(t *testing.T) {
	content := "contaminante,concentracion,unidad\n" +
		"PM10,25.5,µg/m³\n"

	tests := []struct {
		name      string
		workers   int
		batchSize int
	}{
		{name: "sin workers", workers: 0, batchSize: 10},
		{name: "workers negativos", workers: -1, batchSize: 10},
		{name: "lote cero", workers: 2, batchSize: 0},
		{name: "lote negativo", workers: 2, batchSize: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createCSV(t, content)

			reader, err := OpenCSV(path)
			if err != nil {
				t.Fatal(err)
			}
			defer reader.Close()

			_, err = ProcessCSVConcurrent(reader, tt.workers, tt.batchSize)

			if err == nil {
				t.Fatal("se esperaba un error por parámetros inválidos")
			}
		})
	}
}

func TestProcessCSVConcurrentEmptyDataset(t *testing.T) {
	content := "contaminante,concentracion,unidad\n"
	path := createCSV(t, content)

	reader, err := OpenCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	stats, err := ProcessCSVConcurrent(reader, 2, 10)
	if err != nil {
		t.Fatal(err)
	}

	if stats.Read != 0 || stats.Accepted != 0 || stats.Rejected != 0 {
		t.Errorf("se esperaban conteos en cero: %+v", stats)
	}
}

// Comprueba que los archivos temporales de prueba se creen correctamente.
func TestCreateCSVHelper(t *testing.T) {
	path := createCSV(t, "prueba\n")

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("no se encontró el archivo temporal: %v", err)
	}
}
