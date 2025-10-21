package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/config"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/handler"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/service"
	"github.com/redis/go-redis/v9"
)

func main() {

	// Loading configs
	config, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// Preparing dependencies
	messageService := service.NewRedisMessageService(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPass,
		DB:       config.RedisDB,
	})
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
	addr := fmt.Sprintf(":%s", config.Port)
	fmt.Printf("WebSocket server started on %s\n", addr)
	r.Run(addr)

}
