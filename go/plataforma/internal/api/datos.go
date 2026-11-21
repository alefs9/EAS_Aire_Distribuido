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