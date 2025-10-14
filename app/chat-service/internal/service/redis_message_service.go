package service

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisMessageService struct {
	client *redis.Client
}

func NewRedisMessageService(options *redis.Options) *RedisMessageService {
	return &RedisMessageService{
		client: redis.NewClient(options),
	}
}

func (service *RedisMessageService) SendMessage(ctx context.Context, destination string, msg string) error {
	return service.client.Publish(ctx, fmt.Sprintf("user:%s", destination), msg).Err()
}

func (service *RedisMessageService) SubscribeToMessages(ctx context.Context, channel string) (<-chan string, error) {
	// Subscribing to messages channel
	redisChan := service.client.Subscribe(ctx, fmt.Sprintf("user:%s", channel)).Channel()
	// Creating the channel to return
	msgCh := make(chan string)

	// Go routing to convert redis.Message channel to string channel
	go func(ctx context.Context) {
		defer close(msgCh)
		for msg := range redisChan {
			msgCh <- msg.Payload
		}
	}(ctx)

	return msgCh, nil
}

func (service *RedisMessageService) Close() error {
	return service.client.Close()
}
