package persistencia

import (
	"context"
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

func (s *Servicio) Clave(id string) string { return s.Prefijo + ":resultado:v1:" + id }

func (s *Servicio) Invalidar(ctx context.Context, id string) error {
	return s.Cache.Del(ctx, s.Clave(id)).Err()
}

func (s *Servicio) Iniciar(ctx context.Context, id string) error {
	if err := s.Repo.Iniciar(ctx, id); err != nil {
		return err
	}
	_ = s.Invalidar(ctx, id)
	return nil
}
