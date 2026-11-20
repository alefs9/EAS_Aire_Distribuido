package persistencia

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Repositorio struct {
	Cliente *mongo.Client
	DB      *mongo.Database
}

func ConectarMongo(ctx context.Context, uri, nombre string) (*Repositorio, error) {
	c, err := mongo.Connect(options.Client().ApplyURI(uri).SetServerSelectionTimeout(2 * time.Second).SetConnectTimeout(2 * time.Second))
	if err != nil {
		return nil, errors.New("configuración de MongoDB inválida")
	}
	if err = c.Ping(ctx, nil); err != nil {
		_ = c.Disconnect(context.Background())
		return nil, errors.New("MongoDB no disponible")
	}
	return &Repositorio{c, c.Database(nombre)}, nil
}
