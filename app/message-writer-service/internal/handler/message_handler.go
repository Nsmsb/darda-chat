package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageHandler struct {
	dbName         string
	collectionName string
	dbClient       *mongo.Client
}

// NewMessageHandler creates a new MessageHandler instance.
func NewMessageHandler(dbName, collectionName string, dbClient *mongo.Client, redisClient *redis.Client) *MessageHandler {
	return &MessageHandler{
		dbName:         dbName,
		collectionName: collectionName,
		dbClient:       dbClient,
	}
}

// Handle handles the message event and writes it to the database
func (h *MessageHandler) Handle(ctx context.Context, event model.Event) error {
	if event.Type == model.EventTypeMessage {
		// Adding message to DB
		// TODO: idempotency check: ignore if message with same ID exists
		var msg model.Message
		if err := json.Unmarshal(event.Content, &msg); err != nil {
			return fmt.Errorf("unmarshal event content error: %w", err)
		}
		collection := h.dbClient.Database(h.dbName).Collection(h.collectionName)
		_, err := collection.InsertOne(ctx, msg)
		if err != nil {
			return err
		}
		// TODO: send MessagePersisted event using outbox pattern
	}
	return nil
}
