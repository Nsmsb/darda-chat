package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/config"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/handler"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/middleware"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/service"
	"github.com/nsmsb/darda-chat/app/chat-service/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

	// Preparing Message Service
	messageService := service.NewRedisMessageService(redisClient)
	// Closing Connection Gracefully on exit
	defer func() {
		if err := messageService.Close(); err != nil {
			logger.Error("Error during closing Message Service", zap.Error(err))
		}
	}()

	// Preparing handlers
	messageHandler := handler.NewMessageHandler(messageService)

	// Router with no middlewares
	r := gin.New()

	// Adding Middlewares
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.ZapLogger(logger))
	// TODO: Add Error middleware

	// Adding Health Handler
	healthHandler := handler.NewHealthHandler()
	r.GET("/healthz", healthHandler.Liveness)
	r.GET("/readyz", healthHandler.Readiness)

	// Grouping endpoint in /api/v1
	api := r.Group("/api/v1")

	// Adding connections handler
	api.GET("/ws", messageHandler.HandleConnections)

	// Running Server
	addr := fmt.Sprintf("0.0.0.0:%s", config.Port)
	logger.Info(fmt.Sprintf("WebSocket server started on %s\n", addr))
	r.Run(addr)
}
