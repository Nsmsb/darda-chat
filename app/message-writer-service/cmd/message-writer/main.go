package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/config"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/consumer"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/db"
	"github.com/nsmsb/darda-chat/app/message-writer-service/internal/handler"
	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger"
	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/rabbitmq"
	"github.com/redis/go-redis/v9"
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

	// Connection to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPass,
		DB:       config.RedisDB,
	})
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("Error closing Redis connection", zap.Error(err))
		}
	}()
	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Error("Error connecting to Redis", zap.Error(err))
		os.Exit(1)
	}

	// Initializing message consumer
	handler := handler.NewMessageHandler(config.MongoDBName, config.MongoCollectionName, dbClient, redisClient)
	logger.Info("Initializing message consumer")
	consumer := consumer.NewMessageConsumer(config.MsgQueue, handler, conn, config.ConsumerPoolSize)

	// Declaring the message queue
	logger.Info("Declaring message queue", zap.String("queue", config.MsgQueue))
	err := consumer.DeclareQueue(config.MsgQueue)
	if err != nil {
		logger.Error("Failed to declare queue", zap.Error(err))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Starting the message consumer with context
	go func() {
		logger.Info("Starting message consumer", zap.String("queue", config.MsgQueue))
		err = consumer.Start(ctx)
		if err != nil {
			logger.Error("Failed to start message consumer", zap.Error(err))
			cancel()
		}
	}()

	// Graceful shutdown: Listen for interrupt or termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-quit
	logger.Info("Received shutdown signal, waiting for workers to finish...")

	// Gracefully stop the consumer by cancelling the context
	cancel()

	// Wait for all workers and close the consumer
	consumer.Close()

	logger.Info("Gracefully shutting down the writer service")
}
