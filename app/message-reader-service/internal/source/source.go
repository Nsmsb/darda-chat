package source

import "context"

type EventEnvelope[T any] struct {
	ID          string
	Payload     *T
	DeliveryTag uint64
}

type Source[T any] interface {
	DeclareQueue(ctx context.Context) error
	Events(ctx context.Context) <-chan EventEnvelope[T]
	Ack(deliveryTag uint64) error
	Nack(deliveryTag uint64, requeue bool) error
}
