package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCSVToBothModels(t *testing.T) {
	path := filepath.Join(t.TempDir(), "features.csv")
	if err := os.WriteFile(path, []byte("PM10,PM2.5\n1,1\n1,2\n9,9\n9,10\n10,10\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var reports []execution
	for _, mode := range []string{"sequential", "concurrent"} {
		var out bytes.Buffer
		if err := runCLI([]string{"-input", path, "-period", "1_hora", "-mode", mode, "-k", "2", "-workers", "8", "-labels"}, &out); err != nil {
			t.Fatal(err)
		}
		var report execution
		if err := json.Unmarshal(out.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if report.Vectors != 5 || len(report.Labels) != 5 || report.Period != "1_hora" {
			t.Fatalf("bad report: %+v", report)
		}
		reports = append(reports, report)
	}
	a, b := reports[0], reports[1]
	assertEquivalent(t, KMeansResult{a.Centroids, a.Labels, a.Inertia, a.Iterations, a.Converged}, KMeansResult{b.Centroids, b.Labels, b.Inertia, b.Iterations, b.Converged})
	if a.LabelsSHA256 != b.LabelsSHA256 || b.WorkersEffective != 5 {
		t.Fatal("hash or worker count")
	}
}
