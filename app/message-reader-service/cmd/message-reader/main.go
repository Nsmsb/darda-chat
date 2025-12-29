package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/config"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/db"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/model"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/processor"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/repository"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/server"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/source"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/worker"
	"github.com/nsmsb/darda-chat/app/message-reader-service/pkg/logger"
	"github.com/nsmsb/darda-chat/app/message-reader-service/pkg/rabbitmq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	// TODO: Add graceful shutdown for MongoDB and Redis connections
	// TODO: Add readiness and liveness probes
	// Get configuration
	config := config.Get()

	// Getting logger
	logger := logger.Get()
	defer logger.Sync()

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
	// TODO: add readiness probe
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Error("Error connecting to Redis", zap.Error(err))
		os.Exit(1)
	}
	logger.Info("Connected to Redis successfully")

	// Connection to RabbitMQ
	amqpClient := rabbitmq.Conn()
	defer func() {
		if err := amqpClient.Close(); err != nil {
			logger.Error("Error closing AMQP connection", zap.Error(err))
		}
	}()

	// Creating amqp channel
	amqpChannel, err := amqpClient.Channel()
	if err != nil {
		logger.Error("Error creating AMQP channel", zap.Error(err))
		os.Exit(1)
	}
	defer func() {
		if err := amqpChannel.Close(); err != nil {
			logger.Error("Error closing AMQP channel", zap.Error(err))
		}
	}()

	// Creating MongoDB client
	mongoClient := db.Client()

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Port))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// Preparing for message service creation
	conversationRepo := repository.NewMongoConversationRepository(mongoClient, config.MongoDBName, config.MongoCollectionName, config.MessagePageSize)
	conversationCacheRepo := repository.NewRedisConversationCacheRepository(redisClient, config.CacheTTL)

	// Preparing cache update worker
	cacheUpdateProcessor := processor.NewCacheUpdateProcessor(conversationCacheRepo)
	amqpSource := source.NewRabbitMQSource[model.Message](amqpChannel, config.MsgExchange, config.MsgQueue)
	cacheUpdateWorkerPool := worker.NewWorkerPool[model.Message](amqpSource, cacheUpdateProcessor, config.WorkerPoolSize)

	// Create gRPC server with already registered handlers
	s := server.NewMessageGRPCServer(conversationRepo, conversationCacheRepo)

	// Start serving
	go func() {
		logger.Info("server listening on port", zap.String("port", config.Port))
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	// Start cache update worker pool
	go func() {
		logger.Info("Starting cache update worker pool", zap.Int("poolSize", config.WorkerPoolSize))
		if err := cacheUpdateWorkerPool.Start(context.Background()); err != nil {
			logger.Error("Error starting cache update worker pool", zap.Error(err))
			panic(err)
		}
	}()

	// graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	s.GracefulStop()
	cacheUpdateWorkerPool.Stop()
	fmt.Println("server stopped")
}
