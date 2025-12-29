package repository

import (
	"fmt"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoMessageRepository struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	collection     *mongo.Collection
}

func NewMongoMessageRepository(client *mongo.Client, dbName string, collectionName string) *MongoMessageRepository {
	return &MongoMessageRepository{
		client:         client,
		dbName:         dbName,
		collectionName: collectionName,
		collection:     client.Database(dbName).Collection(collectionName),
	}
}

func (r *MongoMessageRepository) Client() *mongo.Client {
	return r.client
}

func (r *MongoMessageRepository) WriteMessage(ctx mongo.SessionContext, message model.Message) error {
	_, err := r.collection.InsertOne(ctx, message)
	if err != nil {
		return fmt.Errorf("insert message error: %w", err)
	}
	return nil
}
