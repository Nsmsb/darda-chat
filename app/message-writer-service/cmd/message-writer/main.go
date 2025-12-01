package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/config"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/db"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/processor"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/repository"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/service"
	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger"
	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/rabbitmq"
	"go.uber.org/zap"
)

func main() {
	// TODO: Add readiness and liveness probes

	// Loading configuration
	config := config.Get()

	// Initializing logger
	logger := logger.Get()
	defer logger.Sync()

	// Connecting to DB
	dbClient := db.Client()
	defer func() {
		if err := dbClient.Disconnect(context.Background()); err != nil {
			logger.Error("Error disconnecting MongoDB client", zap.Error(err))
		}
	}()

	// Preparing rabbitMQ connection
	conn := rabbitmq.Conn()
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error("Error closing RabbitMQ connection", zap.Error(err))
		}
	}()

	// Preparing repositories
	messageRepository := repository.NewMongoMessageRepository(dbClient, config.MongoDBName, config.MongoCollectionName)
	outboxRepository := repository.NewMongoOutboxMessageRepository(dbClient, config.MongoDBName, fmt.Sprintf("%s_outbox", config.MongoCollectionName))

	// Initializing Message consumer Service
	processor := processor.NewMessageProcessor(messageRepository, outboxRepository, dbClient)
	logger.Info("Initializing message consumer service")
	consumerService := service.NewMessageConsumerService(config.MsgQueue, processor, conn, config.ConsumerPoolSize)

	// Initializing the message dispatcher service
	dispatcherCh, err := conn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel", zap.Error(err))
	}
	defer dispatcherCh.Close()

	dispatcher := service.NewRabbitMQDispatcher(dispatcherCh, "message.dispatched")
	dispatcherService := service.NewMessageDispatcherService(dispatcher, outboxRepository)

	// Creating root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Starting the message consumerService with context
	go func() {
		logger.Info("Starting message consumerService", zap.String("queue", config.MsgQueue))
		err = consumerService.Start(ctx)
		if err != nil {
			logger.Error("Failed to start message consumerService", zap.Error(err))
			cancel()
		}
	}()

	// Starting the message dispatcher service with context
	go func() {
		logger.Info("Starting message dispatcher service")
		err = dispatcherService.Start(ctx)
		if err != nil {
			logger.Error("Failed to start message dispatcher service", zap.Error(err))
			cancel()
		}
	}()

	// Graceful shutdown: Listen for interrupt or termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-quit
	logger.Info("Received shutdown signal, waiting for workers to finish...")

	// Gracefully stop the consumerService by cancelling the context
	cancel()

	// Wait for all workers and stop the consumerService
	consumerService.Stop()

	logger.Info("Gracefully shutting down the writer service")
}
