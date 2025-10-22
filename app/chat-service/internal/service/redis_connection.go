package service

import (
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

type RedisConnection struct {
	Conn        *redis.PubSub
	Count       int
	Subscribers map[<-chan string]chan string
	m           sync.Mutex
	startOnce   sync.Once
}

func NewRedisConnection(pubsub *redis.PubSub) *RedisConnection {
	return &RedisConnection{
		Conn:        pubsub,
		Count:       0,
		Subscribers: make(map[<-chan string]chan string),
	}
}

// StartReading starts a goroutine to read messages from Redis and distribute them to subscribers
func (conn *RedisConnection) StartReading() {
	// Ensure the reading goroutine is started only once per user
	conn.startOnce.Do(func() {
		go func() {

			redisChan := conn.Conn.Channel()
			for msg := range redisChan {
				conn.m.Lock()
				for ch := range conn.Subscribers {
					biCh := conn.Subscribers[ch]
					biCh <- msg.Payload
				}
				conn.m.Unlock()
			}
		}()
	})
}

// AddSubscriber adds a subscriber channel to the connections
func (conn *RedisConnection) NewSubscriber() <-chan string {
	// Creating a new channel for the subscriber
	ch := make(chan string)

	conn.m.Lock()
	defer conn.m.Unlock()
	if conn.Subscribers == nil {
		conn.Subscribers = make(map[<-chan string]chan string)
	}
	conn.Subscribers[ch] = ch
	conn.Count++

	return ch
}

// RemoveSubscriber removes a subscriber channel from the connections
func (conn *RedisConnection) RemoveSubscriber(ch <-chan string) error {
	conn.m.Lock()
	defer conn.m.Unlock()
	biCh, exists := conn.Subscribers[ch]
	if !exists {
		return fmt.Errorf("subscriber channel not found")
	}
	// Updating subscriber count and closing the channel
	conn.Count--
	close(biCh)
	delete(conn.Subscribers, ch)

	return nil
}

// SubscriberCount returns the number of active subscribers
func (conn *RedisConnection) SubscriberCount() int {
	conn.m.Lock()
	defer conn.m.Unlock()
	return conn.Count
}

// Close closes the Redis PubSub connection and all subscriber channels
func (conn *RedisConnection) Close() error {
	conn.m.Lock()
	defer conn.m.Unlock()
	// Close the Redis PubSub connection
	err := conn.Conn.Close()
	if err != nil {
		return err
	}
	// Close all subscriber channels
	for _, ch := range conn.Subscribers {
		close(ch)
		conn.Count--
	}
	return nil
}
