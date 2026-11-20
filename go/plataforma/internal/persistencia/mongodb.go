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

func (r *Repositorio) GuardarDataset(ctx context.Context, d Dataset) error {
	if err := validarDataset(d); err != nil {
		return err
	}
	_, err := r.DB.Collection("datasets").InsertOne(ctx, d)
	return err
}

func (r *Repositorio) Dataset(ctx context.Context, id string) (Dataset, error) {
	var d Dataset
	err := r.DB.Collection("datasets").FindOne(ctx, bson.M{"_id": id}).Decode(&d)
	if errors.Is(err, mongo.ErrNoDocuments) {
		err = ErrNoEncontrado
	}
	return d, err
}

func (r *Repositorio) Crear(ctx context.Context, id, dataset string, p Parametros) error {
	if id == "" {
		return errors.New("job_id vacío")
	}
	d, err := r.Dataset(ctx, dataset)
	if err != nil {
		return err
	}
	if err = validarParametros(p, d.Vectores); err != nil {
		return err
	}
	now := time.Now().UTC()
	_, err = r.DB.Collection("jobs").InsertOne(ctx, Trabajo{ID: id, DatasetID: dataset, Parametros: p, Estado: "queued", Creado: now, Actualizado: now})
	return err
}

func (r *Repositorio) Consultar(ctx context.Context, id string) (Trabajo, error) {
	var j Trabajo
	err := r.DB.Collection("jobs").FindOne(ctx, bson.M{"_id": id}).Decode(&j)
	if errors.Is(err, mongo.ErrNoDocuments) {
		err = ErrNoEncontrado
	}
	return j, err
}

func (r *Repositorio) cambiar(ctx context.Context, id string, previos bson.A, campos bson.M) error {
	campos["actualizado"] = time.Now().UTC()
	u, err := r.DB.Collection("jobs").UpdateOne(ctx, bson.M{"_id": id, "estado": bson.M{"$in": previos}}, bson.M{"$set": campos})
	if err != nil {
		return err
	}
	if u.MatchedCount != 1 {
		return ErrEstado
	}
	return nil
}

func (r *Repositorio) Iniciar(ctx context.Context, id string) error {
	return r.cambiar(ctx, id, bson.A{"queued"}, bson.M{"estado": "running"})
}

func (r *Repositorio) Fallar(ctx context.Context, id, motivo string) error {
	if motivo == "" {
		return errors.New("indique el motivo del fallo")
	}
	return r.cambiar(ctx, id, bson.A{"queued", "running"}, bson.M{"estado": "failed", "error": motivo})
}
