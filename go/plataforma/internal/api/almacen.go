package api

import (
	"context"
	"sync"
	"time"

	"eas-aire/plataforma/internal/persistencia"
)

// Almacen es lo que la API necesita de la persistencia. Lo cumplen
// AlmacenPersistente (MongoDB + Redis) y Memoria (desarrollo y pruebas).
type Almacen interface {
	Crear(ctx context.Context, id, datasetID string, p persistencia.Parametros) error
	Iniciar(ctx context.Context, id string) error
	Completar(ctx context.Context, id string, r persistencia.Resultado) error
	Fallar(ctx context.Context, id, motivo string) error
	// Leer devuelve persistencia.ErrNoEncontrado si no existe y
	// persistencia.ErrPendiente si todavía no está completado.
	Leer(ctx context.Context, id string) (persistencia.Lectura, error)
	Salud(ctx context.Context) map[string]string // dependencia -> "ok" o el error
}

// Memoria guarda los trabajos en un mapa protegido por un RWMutex.
// Reproduce los estados y errores de persistencia, sin MongoDB ni Redis.
type Memoria struct {
	mu       sync.RWMutex
	trabajos map[string]persistencia.Trabajo
}

func NuevaMemoria() *Memoria { return &Memoria{trabajos: make(map[string]persistencia.Trabajo)} }

func (m *Memoria) Crear(_ context.Context, id, datasetID string, p persistencia.Parametros) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ahora := time.Now().UTC()
	m.trabajos[id] = persistencia.Trabajo{ID: id, DatasetID: datasetID, Parametros: p, Estado: "queued", Creado: ahora, Actualizado: ahora}
	return nil
}

func (m *Memoria) cambiar(id string, previos []string, f func(*persistencia.Trabajo)) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.trabajos[id]
	if !ok {
		return persistencia.ErrNoEncontrado
	}
	permitido := false
	for _, e := range previos {
		permitido = permitido || t.Estado == e
	}
	if !permitido {
		return persistencia.ErrEstado
	}
	f(&t)
	t.Actualizado = time.Now().UTC()
	m.trabajos[id] = t
	return nil
}

func (m *Memoria) Iniciar(_ context.Context, id string) error {
	return m.cambiar(id, []string{"queued"}, func(t *persistencia.Trabajo) { t.Estado = "running" })
}

func (m *Memoria) Completar(_ context.Context, id string, r persistencia.Resultado) error {
	return m.cambiar(id, []string{"running"}, func(t *persistencia.Trabajo) { t.Estado = "completed"; t.Resultado = &r })
}

func (m *Memoria) Fallar(_ context.Context, id, motivo string) error {
	return m.cambiar(id, []string{"queued", "running"}, func(t *persistencia.Trabajo) { t.Estado = "failed"; t.Error = motivo })
}

func (m *Memoria) Leer(_ context.Context, id string) (persistencia.Lectura, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.trabajos[id]
	if !ok {
		return persistencia.Lectura{}, persistencia.ErrNoEncontrado
	}
	l := persistencia.Lectura{Trabajo: t, Fuente: "memoria"}
	if t.Estado != "completed" {
		return l, persistencia.ErrPendiente
	}
	return l, nil
}

func (m *Memoria) Salud(context.Context) map[string]string { return map[string]string{"memoria": "ok"} }
