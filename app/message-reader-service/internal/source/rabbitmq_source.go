package source

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQSource[T any] struct {
	channel  *amqp.Channel
	queue    string
	exchange string
}

func NewRabbitMQSource[T any](channel *amqp.Channel, exchange, queue string) *RabbitMQSource[T] {
	return &RabbitMQSource[T]{
		channel:  channel,
		exchange: exchange,
		queue:    queue,
	}
}

func (r *RabbitMQSource[T]) DeclareQueue(ctx context.Context) error {
	// Declare exchange
	err := r.channel.ExchangeDeclare(
		r.exchange,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Declare queue
	q, err := r.channel.QueueDeclare(
		r.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	// Bind queue to exchange
	err = r.channel.QueueBind(
		q.Name,
		"",
		r.exchange,
		false,
		nil,
	)

	return err
}

func (r *RabbitMQSource[T]) Events() <-chan EventEnvelope[T] {
	out := make(chan EventEnvelope[T])

	msgs, _ := r.channel.Consume(
		r.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	go func() {
		for d := range msgs {
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
	}()

	return out
}

func (r *RabbitMQSource[T]) Ack(deliveryTag uint64) error {
	return r.channel.Ack(deliveryTag, false)
}

func (r *RabbitMQSource[T]) Nack(deliveryTag uint64, requeue bool) error {
	return r.channel.Nack(deliveryTag, false, requeue)
}
