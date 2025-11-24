package repository

import (
	"fmt"
	"time"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoOutboxMessageRepository struct {
	client         *mongo.Client
	dbName         string
	collectionName string
}

func NewMongoOutboxMessageRepository(client *mongo.Client, dbName string, collectionName string) *MongoOutboxMessageRepository {
	return &MongoOutboxMessageRepository{
		client:         client,
		dbName:         dbName,
		collectionName: collectionName,
	}
}

func (r *MongoOutboxMessageRepository) WriteOutboxMessage(ctx mongo.SessionContext, message model.Message) error {
	collection := r.client.Database(r.dbName).Collection(r.collectionName)
	outboxMessage := model.OutboxMessage{
		ID:          message.ID,
		Payload:     message,
		CreatedAt:   time.Now().UTC(),
		ProcessedAt: time.Time{},
	}
	_, err := collection.InsertOne(ctx, outboxMessage)
	if err != nil {
		return fmt.Errorf("insert outbox event error: %w", err)
	}
	return nil
}
