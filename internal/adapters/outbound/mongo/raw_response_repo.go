package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/duborgis/llm-mock-api/internal/openmeter/domain"
)

// RawResponseRepo implements ports.RawResponseRepository against a real MongoDB collection.
// Upserted by CallID so retried/duplicate callback deliveries don't pile up duplicate docs.
type RawResponseRepo struct {
	collection *mongo.Collection
}

func NewRawResponseRepo(ctx context.Context, uri, database, collection string) (*RawResponseRepo, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, err
	}
	return &RawResponseRepo{collection: client.Database(database).Collection(collection)}, nil
}

func (r *RawResponseRepo) Save(ctx context.Context, resp domain.RawResponse) error {
	_, err := r.collection.ReplaceOne(ctx, bson.D{{Key: "_id", Value: resp.CallID}}, resp, options.Replace().SetUpsert(true))
	return err
}

func (r *RawResponseRepo) List(ctx context.Context, limit int) ([]domain.RawResponse, error) {
	opts := options.Find().SetSort(bson.D{{Key: "received_at", Value: -1}}).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var responses []domain.RawResponse
	if err := cursor.All(ctx, &responses); err != nil {
		return nil, err
	}
	return responses, nil
}
