// Comando latencia: mide las rutas reales de la API y guarda cada tiempo individual.
// Complementa las capturas de Postman exigidas en el informe (sección 4.2.3).
//
//	go run ./cmd/latencia -base http://localhost:8080 -n 30 -salida ../../results/api/latencia
//
// Escenarios: salud, aceptación de trabajos (POST), estado, resultado sin caché
// (primera lectura de cada trabajo, desde MongoDB), resultado con caché (lecturas
// repetidas, desde Redis) e identificador inexistente. Los trabajos creados se
// ejecutan de verdad en el clúster, así que los nodos deben estar activos.
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type muestra struct {
	escenario, ruta string
	i, codigo       int
	bytes           int
	ms              float64
	fuente          string
	ok              bool
}

type cliente struct {
	base string
	http *http.Client
}

func (c cliente) pedir(metodo, ruta, cuerpo string) (int, []byte, float64, error) {
	var r io.Reader
	if cuerpo != "" {
		r = bytes.NewBufferString(cuerpo)
	}
	req, err := http.NewRequest(metodo, c.base+ruta, r)
	if err != nil {
		return 0, nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	inicio := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, float64(time.Since(inicio).Microseconds()) / 1000, err
	}
	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, b, float64(time.Since(inicio).Microseconds()) / 1000, err
}

func percentil(xs []float64, p float64) float64 {
	s := append([]float64(nil), xs...)
	sort.Float64s(s)
	i := int(math.Ceil(p*float64(len(s)))) - 1 // rango más cercano
	if i < 0 {
		i = 0
	}
	return s[i]
}

func main() {
	base := flag.String("base", "http://localhost:8080", "URL base de la API")
	n := flag.Int("n", 30, "solicitudes por escenario")
	periodo := flag.String("periodo", "24_horas", "periodo de los trabajos creados")
	espera := flag.Duration("espera", 10*time.Minute, "plazo para que terminen los trabajos creados")
	salida := flag.String("salida", "latencia", "carpeta de resultados")
	ttl := flag.Duration("ttl", 61*time.Second, "espera para que venza la caché antes de medir lecturas sin caché (CACHE_TTL_SECONDS + 1 s)")
	flag.Parse()
	if err := os.MkdirAll(*salida, 0o755); err != nil {
		log.Fatal(err)
	}
	c := cliente{base: strings.TrimRight(*base, "/"), http: &http.Client{Timeout: 10 * time.Second}}
	var muestras []muestra
	medir := func(escenario, metodo, ruta, cuerpo string, esperado, i int) []byte {
		codigo, b, ms, err := c.pedir(metodo, ruta, cuerpo)
		var m struct {
			Fuente string `json:"fuente"`
		}
		_ = json.Unmarshal(b, &m)
		muestras = append(muestras, muestra{escenario, metodo + " " + plantilla(ruta), i, codigo, len(b), ms, m.Fuente, err == nil && codigo == esperado})
		return b
	}

	// 1. Salud.
	for i := 1; i <= *n; i++ {
		medir("salud", "GET", "/health", "", 200, i)
	}
	// 2. Aceptación de trabajos: mide solo la respuesta 202, no el entrenamiento.
	var ids []string
	cuerpo := fmt.Sprintf(`{"periodo":%q,"k":4,"semilla":42}`, *periodo)
	for i := 1; i <= *n; i++ {
		b := medir("aceptacion", "POST", "/jobs", cuerpo, 202, i)
		var t struct {
			ID string `json:"job_id"`
		}
		if json.Unmarshal(b, &t) == nil && t.ID != "" {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		log.Fatal("no se creó ningún trabajo; revise la API")
	}
	// 3. Esperar a que el clúster complete los trabajos (no se mide).
	log.Printf("esperando %d trabajos en el clúster...", len(ids))
	limite := time.Now().Add(*espera)
	for _, id := range ids {
		for {
			_, b, _, _ := c.pedir("GET", "/jobs/"+id, "")
			var e struct {
				Trabajo struct{ Estado string } `json:"trabajo"`
			}
			_ = json.Unmarshal(b, &e)
			if e.Trabajo.Estado == "completed" || e.Trabajo.Estado == "failed" {
				break
			}
			if time.Now().After(limite) {
				log.Fatalf("el trabajo %s no terminó a tiempo; ¿están activos los nodos?", id)
			}
			time.Sleep(300 * time.Millisecond)
		}
	}
	// Al consultar el estado durante la espera, la API ya guardó los resultados en Redis.
	// Se espera a que venza el TTL para que la primera lectura de cada trabajo vaya a MongoDB.
	log.Printf("esperando %s a que venza la caché Redis...", *ttl)
	time.Sleep(*ttl)

	// 4. Resultado sin caché: primera lectura de cada trabajo (MongoDB).
	for i, id := range ids {
		medir("resultado_sin_cache", "GET", "/jobs/"+id+"/results", "", 200, i+1)
	}
	// 5. Resultado con caché: lecturas repetidas de un mismo trabajo (Redis).
	for i := 1; i <= *n; i++ {
		medir("resultado_con_cache", "GET", "/jobs/"+ids[0]+"/results", "", 200, i)
	}
	// 6. Estado de distintos trabajos.
	for i, id := range ids {
		medir("estado", "GET", "/jobs/"+id, "", 200, i+1)
	}
	// 7. Identificador inexistente.
	for i := 1; i <= *n; i++ {
		medir("no_encontrado", "GET", "/jobs/pc4-no-existe", "", 404, i)
	}

	guardarCSV(filepath.Join(*salida, "mediciones.csv"), muestras)
	resumen := resumir(muestras, *base, *n)
	if err := os.WriteFile(filepath.Join(*salida, "RESUMEN.md"), []byte(resumen), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Print(resumen)
}

// plantilla reemplaza el identificador para agrupar por ruta.
func plantilla(ruta string) string {
	partes := strings.Split(ruta, "/")
	if len(partes) >= 3 && partes[1] == "jobs" {
		partes[2] = "{id}"
	}
	return strings.Join(partes, "/")
}

func guardarCSV(ruta string, ms []muestra) {
	f, err := os.Create(ruta)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"escenario", "ruta", "i", "codigo", "bytes", "ms", "fuente", "correcto"})
	for _, m := range ms {
		_ = w.Write([]string{m.escenario, m.ruta, strconv.Itoa(m.i), strconv.Itoa(m.codigo), strconv.Itoa(m.bytes),
			strconv.FormatFloat(m.ms, 'f', 3, 64), m.fuente, strconv.FormatBool(m.ok)})
	}
	w.Flush()
}

func resumir(ms []muestra, base string, n int) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Latencia de la API\n\nFecha: %s · Base: %s · Solicitudes por escenario: %d · Cliente: Go %s %s/%s\n\n",
		time.Now().Format("2006-01-02 15:04:05 MST"), base, n, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	sb.WriteString("| Escenario | Ruta | n | Bytes (mediana) | Primera (ms) | Mediana (ms) | p95 (ms) | Máx (ms) | Errores | Fuente | < 100 ms |\n")
	sb.WriteString("|---|---|---:|---:|---:|---:|---:|---:|---:|---|---|\n")
	var orden []string
	grupos := map[string][]muestra{}
	for _, m := range ms {
		if _, ok := grupos[m.escenario]; !ok {
			orden = append(orden, m.escenario)
		}
		grupos[m.escenario] = append(grupos[m.escenario], m)
	}
	for _, esc := range orden {
		g := grupos[esc]
		var t, bs []float64
		errores := 0
		fuentes := map[string]int{}
		for _, m := range g {
			t = append(t, m.ms)
			bs = append(bs, float64(m.bytes))
			if !m.ok {
				errores++
			}
			if m.fuente != "" {
				fuentes[m.fuente]++
			}
		}
		var fs []string
		for f, k := range fuentes {
			fs = append(fs, fmt.Sprintf("%s (%d)", f, k))
		}
		sort.Strings(fs)
		maximo := percentil(t, 1)
		cumple := "Sí"
		if maximo >= 100 || errores > 0 {
			cumple = "No"
		}
		fmt.Fprintf(&sb, "| %s | %s | %d | %.0f | %.2f | %.2f | %.2f | %.2f | %d | %s | %s |\n",
			esc, g[0].ruta, len(g), percentil(bs, 0.5), g[0].ms, percentil(t, 0.5), percentil(t, 0.95), maximo, errores, strings.Join(fs, ", "), cumple)
	}
	sb.WriteString("\nLa latencia es el tiempo de ida y vuelta medido por el cliente. No incluye el entrenamiento,\n" +
		"que se ejecuta de forma asíncrona y se informa en `resultado.segundos_ml`.\n")
	return sb.String()
}
