package source

import (
	"context"
	"encoding/json"

	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQSource[T any] struct {
	channel *amqp.Channel
	queue   string
}

func NewRabbitMQSource[T any](channel *amqp.Channel, queue string) *RabbitMQSource[T] {
	return &RabbitMQSource[T]{
		channel: channel,
		queue:   queue,
	}
}

func (r *RabbitMQSource[T]) DeclareQueue(ctx context.Context) error {
	_, err := r.channel.QueueDeclare(
		r.queue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	return err
}

func (r *RabbitMQSource[T]) Events(ctx context.Context) (<-chan EventEnvelope[T], error) {
	log := logger.FromContext(ctx)

	// Create output channel
	out := make(chan EventEnvelope[T])

	msgs, err := r.channel.Consume(
		r.queue,
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, err
	}

	go func() {

		for {
			select {
			case <-ctx.Done():
				log.Info("Shutting down RabbitMQ source event listener")
				close(out)
				return
			case d := <-msgs:
				var payload T
				err := json.Unmarshal(d.Body, &payload)
				if err != nil {
					// failed to unmarshal, nack the message and continue
					_ = r.Nack(d.DeliveryTag, false)
					continue
				}

				envelope := EventEnvelope[T]{
					ID:          d.MessageId,
					Payload:     &payload,
					DeliveryTag: d.DeliveryTag,
				}
				out <- envelope
			}
		}
	}()

	return out, nil
}

func (r *RabbitMQSource[T]) Ack(deliveryTag uint64) error {
	return r.channel.Ack(deliveryTag, false)
}

func (r *RabbitMQSource[T]) Nack(deliveryTag uint64, requeue bool) error {
	return r.channel.Nack(deliveryTag, false, requeue)
}
