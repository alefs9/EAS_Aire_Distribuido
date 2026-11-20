package persistencia

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
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

func (r *Repositorio) Preparar(ctx context.Context) error {
	// Validación en servidor además de las validaciones del código Go.
	definiciones := map[string]bson.M{
		"datasets": {"bsonType": "object", "required": bson.A{"_id", "archivo", "fuente", "sha256", "periodo", "vectores"}, "properties": bson.M{
			"_id": bson.M{"bsonType": "string", "minLength": 1}, "archivo": bson.M{"bsonType": "string", "minLength": 1}, "fuente": bson.M{"bsonType": "string", "minLength": 1},
			"sha256": bson.M{"bsonType": "string", "pattern": "^[a-fA-F0-9]{64}$"}, "periodo": bson.M{"enum": bson.A{"1_hora", "24_horas"}}, "vectores": bson.M{"bsonType": bson.A{"int", "long"}, "minimum": 1},
		}},
		"jobs": {"bsonType": "object", "required": bson.A{"_id", "dataset_id", "parametros", "estado", "creado", "actualizado"}, "properties": bson.M{
			"_id": bson.M{"bsonType": "string", "minLength": 1}, "dataset_id": bson.M{"bsonType": "string", "minLength": 1},
			"estado": bson.M{"enum": bson.A{"queued", "running", "completed", "failed"}}, "creado": bson.M{"bsonType": "date"}, "actualizado": bson.M{"bsonType": "date"},
			"parametros": bson.M{"bsonType": "object", "required": bson.A{"k", "semilla", "tolerancia", "max_iter"}}, "resultado": bson.M{"bsonType": "object"}, "error": bson.M{"bsonType": "string"},
		}},
	}
	for nombre, schema := range definiciones {
		validator := bson.M{"$jsonSchema": schema}
		if nombre == "jobs" {
			validator["$or"] = bson.A{bson.M{"estado": "completed", "resultado": bson.M{"$exists": true}}, bson.M{"estado": bson.M{"$ne": "completed"}, "resultado": bson.M{"$exists": false}}}
		}
		err := r.DB.CreateCollection(ctx, nombre, options.CreateCollection().SetValidator(validator))
		var ce mongo.CommandError
		if err != nil && !(errors.As(err, &ce) && ce.Code == 48) {
			return err
		}
	}
	_, err := r.DB.Collection("datasets").Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "sha256", Value: 1}, {Key: "periodo", Value: 1}}, Options: options.Index().SetUnique(true).SetName("dataset_hash_periodo")})
	if err != nil {
		return err
	}
	_, err = r.DB.Collection("jobs").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "dataset_id", Value: 1}, {Key: "creado", Value: -1}}, Options: options.Index().SetName("jobs_dataset_fecha")},
		{Keys: bson.D{{Key: "estado", Value: 1}, {Key: "creado", Value: -1}}, Options: options.Index().SetName("jobs_estado_fecha")},
	})
	return err
}
