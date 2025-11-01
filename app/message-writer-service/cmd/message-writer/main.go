package main

import (
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/config"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/consumer"
	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Loading configuration
	config := config.Get()

	// Initializing logger
	logger := logger.Get()
	defer logger.Sync()

	// Connecting to RabbitMQ

	// Initializing message consumer
	consumer := consumer.NewMessageConsumer(config.MsgQueue, config.ConsumerPoolSize)

	// Declaring the message queue
	err := consumer.DeclareQueue(config.MsgQueue)
	if err != nil {
		logger.Error("Failed to declare queue", zap.Error(err))
		return
	}

	// Starting the message consumer
	err = consumer.Start()
	if err != nil {
		logger.Error("Failed to start message consumer", zap.Error(err))
	}
	defer consumer.Close()
}
