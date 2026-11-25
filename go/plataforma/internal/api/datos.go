package api

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"eas-aire/plataforma"
)

// Dataset es un CSV de vectores cargado en memoria al arrancar la API.
type Dataset struct {
	ID       string
	Periodo  string
	Archivo  string
	Ruta     string
	SHA256   string
	Puntos   []plataforma.Point
	Features []string
	Unit     string
	Coverage string
}

// CargarDataset lee el CSV PM10,PM2.5 y calcula su SHA-256.
func CargarDataset(dir, periodo string) (Dataset, error) {
	nombre, ok := Periodos[periodo]
	if !ok {
		return Dataset{}, fmt.Errorf("periodo desconocido %q", periodo)
	}
	ruta := filepath.Join(dir, nombre)
	f, err := os.Open(ruta)
	if err != nil {
		return Dataset{}, err
	}
	defer f.Close()
	hash := sha256.New()
	r := csv.NewReader(io.TeeReader(f, hash))
	cab, err := r.Read()
	if err != nil || len(cab) != 2 || cab[0] != "PM10" || cab[1] != "PM2.5" {
		return Dataset{}, fmt.Errorf("%s: se esperaba el encabezado PM10,PM2.5", nombre)
	}
	var puntos []plataforma.Point
	for fila := 2; ; fila++ {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Dataset{}, fmt.Errorf("%s fila %d: %w", nombre, fila, err)
		}
		var p plataforma.Point
		for i := range p {
			v, err := strconv.ParseFloat(rec[i], 64)
			if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
				return Dataset{}, fmt.Errorf("%s fila %d: valor inválido", nombre, fila)
			}
			p[i] = v
		}
		puntos = append(puntos, p)
	}
	if len(puntos) == 0 {
		return Dataset{}, fmt.Errorf("%s no contiene vectores", nombre)
	}
	suma := hex.EncodeToString(hash.Sum(nil))

	features, unit, coverage := metadatosVerificados(dir, periodo, len(puntos), suma)

	return Dataset{ID: suma + ":" + periodo, Periodo: periodo, Archivo: nombre,
		Ruta: ruta, SHA256: suma, Puntos: puntos,
		Features: features, Unit: unit, Coverage: coverage}, nil
}

// Inercia suma la distancia al cuadrado de cada punto a su centroide más cercano.
// El clúster no la devuelve, así que la API la calcula en paralelo: cada goroutine
// procesa un bloque contiguo y escribe su suma en una posición propia del arreglo.
func Inercia(puntos, centroides []plataforma.Point, workers int) float64 {
	if workers < 1 {
		workers = 1
	}
	if workers > len(puntos) {
		workers = len(puntos)
	}
	parciales := make([]float64, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		inicio, fin := w*len(puntos)/workers, (w+1)*len(puntos)/workers
		wg.Add(1)
		go func(w, inicio, fin int) {
			defer wg.Done()
			var s float64
			for _, p := range puntos[inicio:fin] {
				mejor := math.Inf(1)
				for _, c := range centroides {
					dx, dy := p[0]-c[0], p[1]-c[1]
					if d := dx*dx + dy*dy; d < mejor {
						mejor = d
					}
				}
				s += mejor
			}
			parciales[w] = s
		}(w, inicio, fin)
	}
	wg.Wait()
	var total float64
	for _, s := range parciales { // orden fijo: resultado reproducible
		total += s
	}
	return total
}

type manifestVectores struct {
	OrdenFeatures []string `json:"orden_features"`
	Unidad        string   `json:"unidad"`
	Cobertura     string   `json:"cobertura"`
	PorPeriodo    []struct {
		Periodo string `json:"periodo"`
		Filas   int    `json:"filas"`
		SHA256  string `json:"sha256_features"`
	} `json:"por_periodo"`
}

func metadatosVerificados(dir, periodo string, filas int, hash string) ([]string, string, string) {
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, "", ""
	}
	var m manifestVectores
	if json.Unmarshal(b, &m) != nil || len(m.OrdenFeatures) != 2 ||
		m.OrdenFeatures[0] != "PM10" || m.OrdenFeatures[1] != "PM2.5" {
		return nil, "", ""
	}
	wanted := map[string]string{"1_hora": "1 hora", "24_horas": "24 horas"}[periodo]
	for _, p := range m.PorPeriodo {
		if p.Periodo == wanted && p.Filas == filas && strings.EqualFold(p.SHA256, hash) {
			features := append([]string(nil), m.OrdenFeatures...)
			unit, coverage := strings.TrimSpace(m.Unidad), strings.TrimSpace(m.Cobertura)
			return features, unit, coverage
		}
	}
	return nil, "", ""
}
