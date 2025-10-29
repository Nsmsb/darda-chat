package service

import (
	"context"

	"github.com/nsmsb/darda-chat/app/chat-service/pkg/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPPublisher struct {
	conn *amqp.Connection
}

func NewAMQPPublisher() *AMQPPublisher {
	return &AMQPPublisher{
		conn: rabbitmq.Conn(),
	}
}

func (p *AMQPPublisher) Publish(ctx context.Context, msg string, queue string) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declaring the queue to ensure it exists
	// TODO: Move this to initialization phase
	_, err = ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	// Publishing the message to the queue

	err = ch.PublishWithContext(ctx,
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
