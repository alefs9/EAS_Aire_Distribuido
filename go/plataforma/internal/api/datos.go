package api

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"eas-aire/plataforma"
)

// Dataset es un CSV de vectores cargado en memoria al arrancar la API.
type Dataset struct {
	ID      string // sha256:periodo, igual que en MongoDB
	Periodo string
	Archivo string
	Ruta    string
	SHA256  string
	Puntos  []plataforma.Point
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

	return Dataset{
		ID:      suma + ":" + periodo,
		Periodo: periodo,
		Archivo: nombre,
		Ruta:    ruta,
		SHA256:  suma,
		Puntos:  puntos,
	}, nil
}
