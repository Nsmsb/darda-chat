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
	// Subscribing to Redis messages channel
	redisPubSub := service.client.Subscribe(ctx, fmt.Sprintf("user:%s", channel))
	redisChan := redisPubSub.Channel()

	// Channel to return to caller
	msgCh := make(chan string)

	go func() {
		defer close(msgCh)
		defer redisPubSub.Close() // ensure Redis subscription is closed

		for {
			select {
			case <-ctx.Done():
				// Context canceled, stop goroutine
				fmt.Println("Subscription canceled for channel", channel)
				return
			case msg, ok := <-redisChan:
				if !ok {
					// Redis channel closed
					return
				}
				select {
				case msgCh <- msg.Payload:
					// message sent to caller
				case <-ctx.Done():
					// Context canceled while sending
					return
				}
			}
		}
	}()

	return msgCh, nil
}

func (service *RedisMessageService) Close() error {
	return service.client.Close()
}
