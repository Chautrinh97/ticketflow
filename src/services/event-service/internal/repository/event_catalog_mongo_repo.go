package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ticketflow/services/event-service/internal/model"
)

type EventCatalogMongoRepo struct {
	coll *mongo.Collection
}

func NewEventCatalogMongoRepo(db *mongo.Database) *EventCatalogMongoRepo {
	return &EventCatalogMongoRepo{coll: db.Collection("event_catalog")}
}

// EnsureIndexes per docs/03-data/mongodb-schema.md#index — call once at startup.
func (r *EventCatalogMongoRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "event_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "tags", Value: 1}}},
	})
	return err
}

func (r *EventCatalogMongoRepo) Upsert(ctx context.Context, doc *model.EventCatalog) error {
	_, err := r.coll.ReplaceOne(ctx,
		bson.M{"event_id": doc.EventID},
		doc,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (r *EventCatalogMongoRepo) GetByEventID(ctx context.Context, eventID string) (*model.EventCatalog, error) {
	var doc model.EventCatalog
	if err := r.coll.FindOne(ctx, bson.M{"event_id": eventID}).Decode(&doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
