package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageHandler struct {
	dbName         string
	collectionName string
	dbClient       *mongo.Client
	redisClient    *redis.Client
}

// NewMessageHandler creates a new MessageHandler instance.
func NewMessageHandler(dbName, collectionName string, dbClient *mongo.Client, redisClient *redis.Client) *MessageHandler {
	return &MessageHandler{
		dbName:         dbName,
		collectionName: collectionName,
		dbClient:       dbClient,
		redisClient:    redisClient,
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
		// Adding message to cache
		cachingKey := fmt.Sprintf("chat:history:recent:%s", msg.ConversationID)
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("redis marshal error: %w", err)
		}
		// Add to cache with expiration and Trim the list to last 50 messages
		// TODO: make expiration and cache size configurable
		// TODO make the operation atomic
		pipe := h.redisClient.Pipeline()
		pipe.LPush(ctx, cachingKey, jsonMsg)
		pipe.Expire(ctx, cachingKey, time.Hour)
		pipe.LTrim(ctx, cachingKey, 0, 49)
		_, err = pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("redis pipeline error: %w", err)
		}
	}
	return nil
}
