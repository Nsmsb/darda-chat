package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/nsmsb/darda-chat/app/message-writer-service/pkg/logger"
	"go.uber.org/zap"
)

// Config holds the configuration values for the application.
type Config struct {
	Port             string
	AMQPUser         string
	AMQPPass         string
	AMQPHost         string
	MsgQueue         string
	MongoAddr        string
	MongoUser        string
	MongoPass        string
	MongoTimeout     string
	ConsumerPoolSize int
}

var (
	instance *Config
	once     sync.Once
)

// Get returns the singleton instance of Config, it reads the configs only once.
func Get() *Config {
	var consumerPoolSize int = 10 // default value
	consumerPoolSizeEnv := getEnv("CONSUMER_POOL_SIZE", "10")
	if consumerPoolSizeEnv != "" {
		if val, err := strconv.Atoi(consumerPoolSizeEnv); err == nil {
			consumerPoolSize = val
		} else {
			logger.Get().Error("Invalid CONSUMER_POOL_SIZE, using default", zap.String("value", consumerPoolSizeEnv), zap.Error(err))
		}
	}

	once.Do(func() {
		instance = &Config{
			Port:             getEnv("PORT", "8080"),
			AMQPUser:         getEnv("AMQP_USER", ""),
			AMQPPass:         getEnv("AMQP_PASS", ""),
			AMQPHost:         getEnv("AMQP_HOST", ""),
			MsgQueue:         getEnv("MSG_QUEUE", "messages"),
			MongoAddr:        getEnv("MONGO_ADDR", "mongodb://localhost:27017"),
			MongoTimeout:     getEnv("MONGO_TIMEOUT", "10s"),
			MongoUser:        getEnv("MONGO_USER", "root"),
			MongoPass:        getEnv("MONGO_PASS", ""),
			ConsumerPoolSize: consumerPoolSize,
		}
	})
	return instance
}

// getEnv retrieves the value of the environment variable named by the key.
// If the variable is not present in the environment, then return the defaultValue.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
