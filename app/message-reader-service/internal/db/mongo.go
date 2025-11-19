package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nsmsb/darda-chat/app/message-reader-service/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance *mongo.Client
	clientOnce     sync.Once
)

// Client returns a singleton MongoDB client instance
func Client() *mongo.Client {
	clientOnce.Do(func() {
		config := config.Get()

		timeoutStr := config.MongoTimeout
		timeout := 10 * time.Second
		if timeoutStr != "" {
			if parsed, err := time.ParseDuration(timeoutStr); err == nil {
				timeout = parsed
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		mongoURI := fmt.Sprintf("mongodb://%s:%s@%s/", config.MongoUser, config.MongoPass, config.MongoAddr)
		log.Printf("Connecting to MongoDB at %s", config.MongoAddr)

		clientOptions := options.Client().ApplyURI(mongoURI)

		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			log.Fatalf("MongoDB connection error: %v", err)
		}

		// Ping to confirm the connection
		if err := client.Ping(ctx, nil); err != nil {
			log.Fatalf("MongoDB ping error: %v", err)
		}

		fmt.Println("Connected to MongoDB:", config.MongoAddr)
		clientInstance = client
	})

	return clientInstance
}
