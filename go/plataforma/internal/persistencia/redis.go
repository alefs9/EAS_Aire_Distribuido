package persistencia

import (
	"github.com/redis/go-redis/v9"
	"time"
)

type Servicio struct {
	Repo    *Repositorio
	Cache   *redis.Client
	TTL     time.Duration
	Prefijo string
}

type Lectura struct {
	Trabajo Trabajo `json:"trabajo"`
	Fuente  string  `json:"fuente"`
	Aviso   string  `json:"aviso,omitempty"`
}

func NuevaCache(addr, password string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr, Password: password, DB: 0, MaxRetries: -1, DialTimeout: 300 * time.Millisecond, ReadTimeout: 300 * time.Millisecond, WriteTimeout: 300 * time.Millisecond, ContextTimeoutEnabled: true})
}
