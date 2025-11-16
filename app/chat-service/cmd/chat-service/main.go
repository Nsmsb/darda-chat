package main

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	pb "github.com/nsmsb/darda-chat/app/chat-service/internal/api/message/gen"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/config"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/handler"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/middleware"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/service"
	"github.com/nsmsb/darda-chat/app/chat-service/pkg/logger"
	"github.com/nsmsb/darda-chat/app/chat-service/pkg/rabbitmq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	// Loading configs
	config, err := config.Get()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// setting Gin to Release mode in production
	if config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Preparing logger
	logger := logger.GetLogger()
	defer logger.Sync()

	// Preparing dependencies for Handlers

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

	// Preparing AMQP Publisher

	// Declaring the queue during initialization
	ch, err := rabbitmq.Conn().Channel()
	_, err = ch.QueueDeclare(
		config.MsgQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		logger.Fatal("Failed to declare queue", zap.Error(err))
	}
	// Creating the publisher
	publisher, err := service.NewAMQPPublisher()
	if err != nil {
		logger.Fatal("Failed to create AMQP publisher", zap.Error(err))
	}

	// Preparing Message Service
	messageService := service.NewRedisMessageService(redisClient, publisher)
	// Closing Connection Gracefully on exit
	defer func() {
		if err := messageService.Close(); err != nil {
			logger.Error("Error during closing Message Service", zap.Error(err))
		}
	}()

	// Preparing MessageReaderService
	grpcConn, err := grpc.NewClient(config.MessageReaderServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to messageReaderService: %v", err)
	}
	defer grpcConn.Close()
	messageReaderClient := pb.NewMessageServiceClient(grpcConn)

	messageReaderService := service.NewMessageReaderService(messageReaderClient)
	if err != nil {
		logger.Fatal("Failed to create Message Reader Service client", zap.Error(err))
	}

	// Preparing handlers
	messageHandler := handler.NewMessageHandler(messageService, messageReaderService)

	// Router with no middlewares
	r := gin.New()
	r.Use(gin.Recovery())
	// CORS Middleware, temporary allow all origins
	r.Use(cors.Default())

	// Adding Health Handler
	healthHandler := handler.NewHealthHandler()
	r.GET("/healthz", healthHandler.Liveness)
	r.GET("/readyz", healthHandler.Readiness)

	// Grouping endpoint in /api/v1
	api := r.Group("/api/v1")

	// Adding Middlewares
	api.Use(middleware.RequestIDMiddleware())
	api.Use(middleware.ZapLogger(logger))
	// TODO: Add Error middleware

	// Adding connections handler
	api.GET("/ws", messageHandler.HandleConnections)
	api.GET("/messages/:user", messageHandler.GetMessages)

	// Running Server
	addr := fmt.Sprintf("0.0.0.0:%s", config.Port)
	logger.Info(fmt.Sprintf("WebSocket server started on %s\n", addr))
	r.Run(addr)
}
