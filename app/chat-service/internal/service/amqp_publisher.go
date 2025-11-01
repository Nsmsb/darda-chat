package service

import (
	"context"

	"github.com/nsmsb/darda-chat/app/chat-service/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPPublisher struct {
	ch *amqp.Channel
}

func NewAMQPPublisher() (*AMQPPublisher, error) {
	ch, err := rabbitmq.Conn().Channel()
	if err != nil {
		return nil, err
	}
	return &AMQPPublisher{
		ch: ch,
	}, nil
}

func (p *AMQPPublisher) Publish(ctx context.Context, msg string, queue string) error {

	// Publishing the message to the queue
	// The queue has been declared during the initialization of the service
	err := p.ch.PublishWithContext(ctx,
		"",    // exchange
		queue, // routing key (queue name)
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(msg),
		})
	return err
}

func (p *AMQPPublisher) Close() error {
	return p.ch.Close()
}
