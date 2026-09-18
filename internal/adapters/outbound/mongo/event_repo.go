package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/duborgis/llm-mock-api/internal/openmeter/domain"
)

// EventRepo implements ports.EventRepository against a real MongoDB collection, so events
// LiteLLM sends to the OpenMeter mock can be inspected afterwards with any Mongo client
// (mongosh, Compass, ...) rather than only through this service's own List endpoint.
type EventRepo struct {
	collection *mongo.Collection
}

func NewEventRepo(ctx context.Context, uri, database, collection string) (*EventRepo, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, err
	}
	return &EventRepo{collection: client.Database(database).Collection(collection)}, nil
}

func (r *EventRepo) Save(ctx context.Context, e domain.Event) error {
	_, err := r.collection.InsertOne(ctx, e)
	return err
}

func (r *EventRepo) List(ctx context.Context, limit int) ([]domain.Event, error) {
	opts := options.Find().SetSort(bson.D{{Key: "received_at", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []domain.Event
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}
