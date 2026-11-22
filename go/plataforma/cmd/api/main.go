// Comando api: API REST que también actúa como coordinador del clúster K-Means.
//
// Variables de entorno (valores por defecto entre paréntesis):
//
//	API_ADDR (:8080)              dirección HTTP
//	COORDINATOR_LISTEN (:9000)    dirección TCP donde se registran los nodos
//	DATA_DIR (../../data/vectors) carpeta con features_1_hora.csv y features_24_horas.csv
//	CLUSTER_NODES (2)             nodos por trabajo si la solicitud no lo indica
//	QUEUE_SIZE (8)                capacidad de la cola de trabajos
//	JOB_TIMEOUT (3m)              plazo total de un trabajo, incluida la espera de nodos
//	STORE (mongo)                 "mongo" (MongoDB + Redis) o "memoria" (desarrollo)
//	MONGODB_URI, MONGODB_DATABASE (eas_pc4), REDIS_ADDR, REDIS_PASSWORD, CACHE_TTL_SECONDS (60)
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"eas-aire/plataforma"
	"eas-aire/plataforma/internal/api"
	"eas-aire/plataforma/internal/persistencia"
)

func env(clave, defecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return defecto
}

func envEntero(clave string, defecto int) int {
	v, err := strconv.Atoi(os.Getenv(clave))
	if err != nil || v < 1 {
		return defecto
	}
	return v
}

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	timeout, err := time.ParseDuration(env("JOB_TIMEOUT", "3m"))
	if err != nil || timeout <= 0 {
		logger.Fatal("JOB_TIMEOUT inválido")
	}

	// 1. Datasets: se cargan una vez para validar filas y calcular la inercia.
	dir := env("DATA_DIR", "../../data/vectors")
	datasets := map[string]api.Dataset{}
	for periodo := range api.Periodos {
		d, err := api.CargarDataset(dir, periodo)
		if err != nil {
			logger.Printf("periodo %s no disponible: %v", periodo, err)
			continue
		}
		datasets[periodo] = d
		logger.Printf("dataset %s: %d vectores sha256=%s", periodo, len(d.Puntos), d.SHA256[:12])
	}
	if len(datasets) == 0 {
		logger.Fatalf("no se encontraron CSV de vectores en %s", dir)
	}

	// 2. Almacén: MongoDB + Redis (persistencia de Valentina) o memoria.
	var almacen api.Almacen
	if env("STORE", "mongo") == "memoria" {
		almacen = api.NuevaMemoria()
		logger.Print("almacén en memoria: los trabajos se pierden al reiniciar")
	} else {
		c, cancel := context.WithTimeout(ctx, 10*time.Second)
		repo, err := persistencia.ConectarMongo(c, env("MONGODB_URI", "mongodb://127.0.0.1:27017"), env("MONGODB_DATABASE", "eas_pc4"))
		if err != nil {
			logger.Fatal(err)
		}
		defer repo.Cliente.Disconnect(context.Background())
		if err := repo.Preparar(c); err != nil {
			logger.Fatalf("preparar MongoDB: %v", err)
		}
		for _, d := range datasets {
			if err := api.RegistrarDataset(c, repo, d); err != nil {
				logger.Fatalf("registrar dataset %s: %v", d.Periodo, err)
			}
		}
		cancel()
		cache := persistencia.NuevaCache(env("REDIS_ADDR", "127.0.0.1:6379"), os.Getenv("REDIS_PASSWORD"))
		defer cache.Close()
		ttl := time.Duration(envEntero("CACHE_TTL_SECONDS", 60)) * time.Second
		almacen = api.AlmacenPersistente{S: &persistencia.Servicio{Repo: repo, Cache: cache, TTL: ttl, Prefijo: env("MONGODB_DATABASE", "eas_pc4")}}
		logger.Printf("almacén MongoDB + Redis (TTL %s)", ttl)
	}

	// 3. Coordinador del clúster (Alexandra): los nodos se registran por TCP.
	ln, err := net.Listen("tcp", env("COORDINATOR_LISTEN", ":9000"))
	if err != nil {
		logger.Fatal(err)
	}
	coordinador := plataforma.NewCoordinator(20*time.Second, logger)
	coordinadorListo := make(chan error, 1)
	go func() { coordinadorListo <- coordinador.Serve(ctx, ln) }()
	logger.Printf("coordinador TCP escuchando en %s", ln.Addr())

	// 4. Servicio y servidor HTTP.
	servicio := api.NuevoServicio(almacen, coordinador, datasets, api.ConfigServicio{
		TamanoCola: envEntero("QUEUE_SIZE", 8), Timeout: timeout, NodosDefecto: envEntero("CLUSTER_NODES", 2),
	}, logger)
	srv := &http.Server{
		Addr:              env("API_ADDR", ":8080"),
		Handler:           api.ConRegistro(api.NuevoRouter(servicio, logger), logger),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	go func() {
		logger.Printf("API escuchando en %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(err)
		}
	}()

	<-ctx.Done()
	logger.Print("cerrando...")
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(c)
	servicio.Cerrar()
	<-coordinadorListo
}
