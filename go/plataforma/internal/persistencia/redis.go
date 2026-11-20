package persistencia

import (
	"context"
	"encoding/json"
	"errors"
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

func (s *Servicio) Fallar(ctx context.Context, id, motivo string) error {
	if err := s.Repo.Fallar(ctx, id, motivo); err != nil {
		return err
	}
	_ = s.Invalidar(ctx, id)
	return nil
}

func (s *Servicio) Completar(ctx context.Context, id string, r Resultado) error {
	if err := s.Repo.Completar(ctx, id, r); err != nil {
		return err
	}
	_ = s.Invalidar(ctx, id)
	return nil
}

func (s *Servicio) Resultado(ctx context.Context, id string) (Lectura, error) {
	if s.TTL <= 0 {
		return Lectura{}, errors.New("TTL debe ser positivo")
	}
	var j Trabajo
	raw, cacheErr := s.Cache.Get(ctx, s.Clave(id)).Bytes()
	if cacheErr == nil && json.Unmarshal(raw, &j) == nil && j.ID == id && j.Estado == "completed" && j.Resultado != nil {
		return Lectura{Trabajo: j, Fuente: "Redis"}, nil
	}
	j, err := s.Repo.Consultar(ctx, id)
	if err != nil {
		return Lectura{}, err
	}
	if j.Estado != "completed" || j.Resultado == nil {
		return Lectura{Trabajo: j, Fuente: "MongoDB"}, ErrPendiente
	}
	out := Lectura{Trabajo: j, Fuente: "MongoDB"}
	raw, err = json.Marshal(j)
	if err != nil {
		return Lectura{}, err
	}
	if err = s.Cache.Set(ctx, s.Clave(id), raw, s.TTL).Err(); err != nil || (cacheErr != nil && !errors.Is(cacheErr, redis.Nil)) {
		out.Aviso = "Caché no disponible; resultado recuperado de MongoDB"
	}
	return out, nil
}
