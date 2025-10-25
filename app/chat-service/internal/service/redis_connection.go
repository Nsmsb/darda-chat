package service

import (
	"fmt"
	"sync"
	"time"

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
			defer fmt.Println("Stopped reading from Redis for a connection")
			redisChan := conn.Conn.Channel()
			for msg := range redisChan {
				conn.m.Lock()
				for ch := range conn.Subscribers {
					select {
					case conn.Subscribers[ch] <- msg.Payload:
						// Message sent successfully
					case <-time.After(time.Millisecond * 100):
						// Timeout ended, subscriber is not receiving messages or so slow
						fmt.Println("Subscriber is not receiving messages, removing subscriber")
						close(conn.Subscribers[ch])
						delete(conn.Subscribers, ch)
						conn.Count--
						if conn.Count <= 0 {
							// No more subscribers, exit reading loop
							conn.m.Unlock()
							fmt.Println("No more subscribers, exiting reading loop")
							return
						}
					}
				}
				conn.m.Unlock()
			}
		}()
	})
}

// AddSubscriber adds a subscriber channel to the connections
func (conn *RedisConnection) NewSubscriber() <-chan string {
	// Creating a new buffered channel for the subscriber
	// So the sender won't be blocked if the receiver is slow
	ch := make(chan string, 30)

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
