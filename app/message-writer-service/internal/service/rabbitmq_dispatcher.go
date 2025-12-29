package service

import (
	"encoding/json"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQDispatcher struct {
	channel  *amqp.Channel
	exchange string
}

func NewRabbitMQDispatcher(channel *amqp.Channel, exchange string) *RabbitMQDispatcher {

	return &RabbitMQDispatcher{
		channel:  channel,
		exchange: exchange,
	}
}

func (d *RabbitMQDispatcher) DeclareExchange() error {
	return d.channel.ExchangeDeclare(
		d.exchange, // name
		"fanout",   // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
}

func (d *RabbitMQDispatcher) Dispatch(message model.Message) error {
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}
	// publish the message to the RabbitMQ queue
	err = d.channel.Publish(
		d.exchange, // exchange
		"",         // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(jsonMessage),
		},
	)
	return err
}
