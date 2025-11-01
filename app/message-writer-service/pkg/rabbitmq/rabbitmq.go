package rabbitmq

import (
	"fmt"
	"log"
	"sync"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

var conn *amqp.Connection
var once sync.Once

// Conn returns a singleton RabbitMQ connection
func Conn() *amqp.Connection {
	once.Do(func() {
		var err error
		config := config.Get()
		AMQPUser := config.AMQPUser
		AMQPPass := config.AMQPPass
		AMQPHost := config.AMQPHost
		conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", AMQPUser, AMQPPass, AMQPHost))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
	})
	return conn
}
