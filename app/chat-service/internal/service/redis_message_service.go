package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/nsmsb/darda-chat/app/chat-service/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisMessageService struct {
	client      *redis.Client
	publisher   Publisher
	connections map[string]Connection
	m           sync.Mutex
}

func NewRedisMessageService(client *redis.Client, publisher Publisher) *RedisMessageService {
	return &RedisMessageService{
		client:      client,
		publisher:   publisher,
		connections: make(map[string]Connection),
	}
}

func (service *RedisMessageService) SendMessage(ctx context.Context, destination string, msg string) error {
	config, _ := config.Get()
	// Publishing message to message queue
	err := service.publisher.Publish(ctx, msg, config.MsgQueue)
	if err != nil {
		return err
	}
	// Publishing message to Redis channel for real-time delivery
	return service.client.Publish(ctx, fmt.Sprintf("user:%s", destination), msg).Err()
}

func (service *RedisMessageService) SubscribeToMessages(ctx context.Context, channel string) (<-chan string, error) {

	// Subscribing to Redis messages channel
	service.m.Lock()
	conn, exists := service.connections[channel]
	// Creating the channel connection if it doesn't exist
	if !exists {
		redisPubSub := service.client.Subscribe(ctx, fmt.Sprintf("user:%s", channel))
		conn = NewRedisConnection(redisPubSub)
		service.connections[channel] = conn
	}
	service.m.Unlock()
	// Start reading messages from Redis since it's the first connection
	conn.StartReading()

	return conn.NewSubscriber(), nil
}

func (service *RedisMessageService) UnsubscribeFromMessages(channel string, msgCh <-chan string) error {
	service.m.Lock()
	defer service.m.Unlock()
	conn, exists := service.connections[channel]
	if !exists {
		return fmt.Errorf("no connection found for channel %s", channel)
	}
	// Remove the subscriber channel
	err := conn.RemoveSubscriber(msgCh)
	if err != nil {
		return err
	}

	// If there are no more subscribers, remove the connection
	if conn.SubscriberCount() == 0 {
		fmt.Println("No more subscribers, closing connection for channel", channel)
		err := conn.Close()
		if err != nil {
			return err
		}
		delete(service.connections, channel)
	}

	return err
}

func (service *RedisMessageService) Close() error {
	service.m.Lock()
	defer service.m.Unlock()
	// Close all connections
	for _, conn := range service.connections {
		conn.Close()
	}
	return service.client.Close()
}
