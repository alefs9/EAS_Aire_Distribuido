package main

// Dimensions representa la cantidad de características
// que utilizaremos para cada observación: PM10 y PM2.5.
const Dimensions = 2

// Dataset almacena los vectores que utilizaremos en K-Means.
type Dataset struct {
	Points [][]float64
	N      int
	D      int
}

// NewDataset crea un conjunto de datos a partir de una lista de puntos.
func NewDataset(points [][]float64) Dataset {
	return Dataset{
		Points: points,
		N:      len(points),
		D:      Dimensions,
	}
}
