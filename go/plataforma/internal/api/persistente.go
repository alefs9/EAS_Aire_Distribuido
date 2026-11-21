package api

import (
	"context"
	"errors"
	"fmt"

	"eas-aire/plataforma/internal/persistencia"
)

// AlmacenPersistente adapta persistencia.Servicio (MongoDB + caché Redis) a Almacen.
type AlmacenPersistente struct{ S *persistencia.Servicio }

func (a AlmacenPersistente) Crear(ctx context.Context, id, datasetID string, p persistencia.Parametros) error {
	return a.S.Repo.Crear(ctx, id, datasetID, p)
}
func (a AlmacenPersistente) Iniciar(ctx context.Context, id string) error {
	return a.S.Iniciar(ctx, id)
}
func (a AlmacenPersistente) Completar(ctx context.Context, id string, r persistencia.Resultado) error {
	return a.S.Completar(ctx, id, r)
}
func (a AlmacenPersistente) Fallar(ctx context.Context, id, motivo string) error {
	return a.S.Fallar(ctx, id, motivo)
}
func (a AlmacenPersistente) Leer(ctx context.Context, id string) (persistencia.Lectura, error) {
	return a.S.Resultado(ctx, id)
}

// Salud comprueba MongoDB (obligatorio) y Redis (opcional: sin Redis se lee de MongoDB).
func (a AlmacenPersistente) Salud(ctx context.Context) map[string]string {
	estado := map[string]string{"mongodb": "ok", "redis": "ok"}
	if err := a.S.Repo.Cliente.Ping(ctx, nil); err != nil {
		estado["mongodb"] = "no disponible"
	}
	if err := a.S.Cache.Ping(ctx).Err(); err != nil {
		estado["redis"] = "no disponible"
	}
	return estado
}

// RegistrarDataset guarda la procedencia del CSV si todavía no existe.
// Usa la misma convención de identificador que cmd/persistencia: sha256:periodo.
func RegistrarDataset(ctx context.Context, repo *persistencia.Repositorio, d Dataset) error {
	doc := persistencia.Dataset{
		ID: d.ID, Archivo: d.Archivo, Fuente: "OEFA - Evaluaciones ambientales de seguimiento (datosabiertos.gob.pe)",
		SHA256: d.SHA256, Periodo: d.Periodo, Vectores: len(d.Puntos),
	}
	previo, err := repo.Dataset(ctx, d.ID)
	switch {
	case errors.Is(err, persistencia.ErrNoEncontrado):
		return repo.GuardarDataset(ctx, doc)
	case err != nil:
		return err
	case previo.Vectores != doc.Vectores:
		return fmt.Errorf("el dataset %s ya existe con otra cantidad de vectores", d.ID)
	}
	return nil
}
