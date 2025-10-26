package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/config"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/handler"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/service"
	"github.com/redis/go-redis/v9"
)

func main() {

	// Loading configs
	config, err := config.Get()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Preparing dependencies for Handlers

	// Connection to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPass,
		DB:       config.RedisDB,
	})
	// Testing Redis connection
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	// Preparing Message Service
	messageService := service.NewRedisMessageService(redisClient)
	// Closing Connection Gracefully on exit
	defer func() {
		if err := messageService.Close(); err != nil {
			fmt.Printf("Error closing Redis client: %v\n", err)
		}
	}()

	// Preparing handlers
	handler := handler.NewMessageHandler(messageService)

	// Router with Logger registered by default
	r := gin.Default()

	// TODO: Add Error middleware

	// Grouping endpoint in /api/v1
	api := r.Group("/api/v1")

	// Adding connections handler
	api.GET("/ws", handler.HandleConnections)

	// Running Server
	addr := fmt.Sprintf("0.0.0.0:%s", config.Port)
	fmt.Printf("WebSocket server started on %s\n", addr)
	r.Run(addr)
}
