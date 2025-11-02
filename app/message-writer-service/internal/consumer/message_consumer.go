package consumer

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/handler"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type MessageConsumer struct {
	handler handler.Handler
	queue   string
	conn    *amqp.Connection
	workers chan struct{}
	wg      sync.WaitGroup
}

// NewMessageConsumer creates a new MessageConsumer instance.
func NewMessageConsumer(queue string, handler handler.Handler, conn *amqp.Connection, poolSize int) *MessageConsumer {
	return &MessageConsumer{
		handler: handler,
		queue:   queue,
		conn:    conn,
		workers: make(chan struct{}, poolSize),
	}
}

// DeclareQueue declares the queue to consume messages from.
func (c *MessageConsumer) DeclareQueue(queueName string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		c.queue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	return err
}

// Start starts consuming messages from the queue.
// Messages are processed in parallel based on the pool size.
func (c *MessageConsumer) Start(ctx context.Context) error {
	log := logger.Get()

	ch, err := c.conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel", zap.Error(err))
		return err
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		c.queue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("Message consumer is shutting down")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				log.Info("Message channel closed")
				return nil
			}
			c.workers <- struct{}{}
			c.wg.Add(1)

			go func(m amqp.Delivery) {
				defer c.wg.Done()
				defer func() {
					if r := recover(); r != nil {
						log.Error("Recovered in message processing", zap.Any("error", r))
						_ = m.Nack(false, true)
					}
					// Release the worker slot by reading from the channel
					<-c.workers
				}()

				// creating message object from message body
				var msg model.Message

				if err := json.Unmarshal(m.Body, &msg); err != nil {
					log.Error("Invalid message type", zap.Error(err))
					_ = m.Nack(false, false) // discard the message
					return
				}

				// Process the message
				if err := c.handler.Handle(msg); err != nil {
					log.Error("Failed to process message", zap.Error(err))
					_ = m.Nack(false, true) // requeue the message
					return
				}

				log.Info("Processed a message", zap.ByteString("body", m.Body))

				// Acknowledge the message after processing
				err := m.Ack(false)
				if err != nil {
					log.Error("Failed to acknowledge message", zap.Error(err))
				}

			}(msg)
		}
	}
}

// Close gracefully shuts down the consumer, waiting for all workers to finish.
func (c *MessageConsumer) Close() error {
	c.wg.Wait()
	return c.conn.Close()
}
