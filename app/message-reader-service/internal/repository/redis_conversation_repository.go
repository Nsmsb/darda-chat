package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/config"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/model"
	"github.com/redis/go-redis/v9"
)

type RedisConversationCacheRepository struct {
	client   *redis.Client
	cacheTTL time.Duration
}

func NewRedisConversationCacheRepository(client *redis.Client) *RedisConversationCacheRepository {
	return &RedisConversationCacheRepository{
		client:   client,
		cacheTTL: config.Get().CacheTTL,
	}
}

func (r *RedisConversationCacheRepository) SetConversationMessages(conversationKey string, messages []*model.Message) error {
	ctx := context.Background()

	// Create a redis pipeline
	pipe := r.client.Pipeline()

	// Adding messages to pipeline
	for _, msg := range messages {
		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("message json marshal error: %w", err)
		}
		// Using RPUSH to keep a natural order of messages
		pipe.RPush(ctx, conversationKey, jsonMsg)
	}

	// Setting expiration for the conversation messages
	pipe.Expire(ctx, conversationKey, r.cacheTTL)

	// Executing the pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis pipeline exec error: %w", err)
	}
	return nil
}

func (r *RedisConversationCacheRepository) GetConversationMessages(conversationKey string) ([]*model.Message, error) {
	ctx := context.Background()

	// Retrieving messages from Redis
	jsonMessages, err := r.client.LRange(ctx, conversationKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lrange error: %w", err)
	}

	// If no messages found, return nil
	if len(jsonMessages) == 0 {
		return nil, nil
	}

	// Unmarshal JSON messages into model.Message structs
	var messages []*model.Message
	for _, jsonMsg := range jsonMessages {
		var msg model.Message
		if err := json.Unmarshal([]byte(jsonMsg), &msg); err != nil {
			return nil, fmt.Errorf("message json unmarshal error: %w", err)
		}
		messages = append(messages, &msg)
	}

	return messages, nil
}
