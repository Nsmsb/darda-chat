package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/config"
	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/server"
	"github.com/nsmsb/darda-chat/app/message-reader-service/pkg/logger"
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

	// Create listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", config.Port))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// Create gRPC server with already registered handlers
	s := server.NewMessageGRPCServer()

	go func() {

		logger.Info("server listening on port", zap.String("port", config.Port))
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	// graceful shutdown
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	s.GracefulStop()
	fmt.Println("server stopped")
}
