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

func TestCLIRejectsInvalidContract(t *testing.T) {
	for _, args := range [][]string{{}, {"-input", "missing", "-period", "mixed"}, {"-input", "missing", "-period", "1_hora", "-mode", "invalid"}, {"-input", "missing", "-period", "1_hora", "-workers", "0"}, {"-input", "missing", "-period", "1_hora"}} {
		if err := runCLI(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	for _, csv := range []string{"PM2.5,PM10\n1,2\n", "PM10,PM2.5\nNaN,1\n", "PM10,PM2.5\n1\n", "PM10,PM2.5\n"} {
		p := filepath.Join(t.TempDir(), "invalid.csv")
		if err := os.WriteFile(p, []byte(csv), 0600); err != nil {
			t.Fatal(err)
		}
		if err := runCLI([]string{"-input", p, "-period", "24_horas"}, &bytes.Buffer{}); err == nil {
			t.Fatal("accepted invalid CSV")
		}
	}
}

func TestRealPeriodsEquivalence(t *testing.T) {
	if os.Getenv("KMEANS_REAL_DATA") != "1" {
		t.Skip("KMEANS_REAL_DATA=1 activa los dos datasets reales")
	}
	for _, tc := range []struct {
		period string
		n      int
	}{{"1_hora", 445347}, {"24_horas", 18297}} {
		t.Run(tc.period, func(t *testing.T) {
			data, err := LoadCSV("../../data/vectors/features_" + tc.period + ".csv")
			if err != nil {
				t.Fatal(err)
			}
			if data.N != tc.n {
				t.Fatalf("N=%d", data.N)
			}
			a, err := RunKMeans(data, 4, 100, .001, 42)
			if err != nil {
				t.Fatal(err)
			}
			for _, workers := range []int{1, 2, 4, 8} {
				b, err := RunKMeansConcurrent(data, 4, 100, .001, 42, workers)
				if err != nil {
					t.Fatal(err)
				}
				assertEquivalent(t, a, b)
				t.Logf("period=%s N=%d workers=%d iterations=%d inertia=%.15g labels=equal", tc.period, data.N, workers, b.Iterations, b.Inertia)
			}
		})
	}
}
