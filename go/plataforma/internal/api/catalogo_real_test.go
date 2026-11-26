package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func TestCatalogoRealCoincideConManifiesto(t *testing.T) {
	_, archivo, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("no se pudo localizar el archivo de prueba")
	}

	dir := filepath.Clean(filepath.Join(
		filepath.Dir(archivo), "..", "..", "..", "..", "data", "vectors",
	))

	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}

	var manifest manifestVectores
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}

	periodos := []struct {
		api        string
		manifiesto string
	}{
		{api: "1_hora", manifiesto: "1 hora"},
		{api: "24_horas", manifiesto: "24 horas"},
	}

	datasets := make(map[string]Dataset, len(periodos))

	for _, periodo := range periodos {
		t.Run(periodo.api, func(t *testing.T) {
			var esperado *struct {
				Periodo string `json:"periodo"`
				Filas   int    `json:"filas"`
				SHA256  string `json:"sha256_features"`
			}

			for i := range manifest.PorPeriodo {
				if manifest.PorPeriodo[i].Periodo == periodo.manifiesto {
					esperado = &manifest.PorPeriodo[i]
					break
				}
			}

			if esperado == nil {
				t.Fatalf("no existe %q en manifest.json", periodo.manifiesto)
			}

			dataset, err := CargarDataset(dir, periodo.api)
			if err != nil {
				t.Fatal(err)
			}

			if len(dataset.Puntos) != esperado.Filas {
				t.Fatalf("vectores: API=%d manifiesto=%d", len(dataset.Puntos), esperado.Filas)
			}
			if dataset.SHA256 != esperado.SHA256 {
				t.Fatalf("SHA: API=%s manifiesto=%s", dataset.SHA256, esperado.SHA256)
			}
			if !reflect.DeepEqual(dataset.Features, manifest.OrdenFeatures) {
				t.Fatalf("features: API=%v manifiesto=%v", dataset.Features, manifest.OrdenFeatures)
			}
			if dataset.Unit != manifest.Unidad {
				t.Fatalf("unit: API=%q manifiesto=%q", dataset.Unit, manifest.Unidad)
			}
			if dataset.Coverage != manifest.Cobertura {
				t.Fatalf("coverage: API=%q manifiesto=%q", dataset.Coverage, manifest.Cobertura)
			}

			datasets[periodo.api] = dataset
		})
	}

	status, body := pedirCatalogoDirecto(t, datasets)
	if status != 200 {
		t.Fatalf("GET /datasets: esperaba 200; obtuvo %d", status)
	}

	items, ok := body["datasets"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("esperaba dos conjuntos; obtuvo %#v", body["datasets"])
	}

	for _, value := range items {
		item := value.(map[string]any)
		periodo := item["period"].(string)
		dataset := datasets[periodo]

		if item["id"] != dataset.ID {
			t.Errorf("%s: ID incorrecto: %v", periodo, item["id"])
		}
		if item["sha256"] != dataset.SHA256 {
			t.Errorf("%s: SHA incorrecto: %v", periodo, item["sha256"])
		}
		if item["vector_count"] != float64(len(dataset.Puntos)) {
			t.Errorf("%s: vector_count incorrecto: %v", periodo, item["vector_count"])
		}
		if item["measurement_count"] != float64(len(dataset.Puntos)*2) {
			t.Errorf("%s: measurement_count incorrecto: %v", periodo, item["measurement_count"])
		}
	}
}
