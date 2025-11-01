package consumer

import (
	"context"
	"sync"

	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger"
	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type MessageConsumer struct {
	queue   string
	conn    *amqp.Connection
	workers chan struct{}
	wg      sync.WaitGroup
}

func NewMessageConsumer(queue string, poolSize int) *MessageConsumer {
	return &MessageConsumer{
		queue:   queue,
		conn:    rabbitmq.Conn(),
		workers: make(chan struct{}, poolSize),
	}
}

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
					// Realease the worker slot by reading from the channel
					<-c.workers
				}()

				// Process the message
				// TODO: implement message processing logic here
				log.Info("Processed a message", zap.ByteString("body", m.Body))

				// Acknowledge the message after processing
				err := m.Ack(false)
				if err != nil {
					// TODO: log error
					// If ack fails, the message will be re-delivered
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
