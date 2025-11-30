package config

import (
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/nsmsb/darda-chat/app/message-reader-service/pkg/logger"
	"go.uber.org/zap"
)

// Config holds the configuration values for the application.
type Config struct {
	Port                string
	CacheTTL            time.Duration
	MessagePageSize     int
	MongoDBName         string
	MongoCollectionName string
	MongoAddr           string
	MongoUser           string
	MongoPass           string
	MongoTimeout        string
	RedisAddr           string
	RedisPass           string
	RedisDB             int
	AMQPUser            string
	AMQPPass            string
	AMQPHost            string
	MsgQueue            string
	MsgExchange         string
	WorkerPoolSize      int
}

var (
	instance *Config
	once     sync.Once
)

// Get returns the singleton instance of Config, it reads the configs only once.
func Get() *Config {
	var messagesPageSize, redisDB, workerPoolSize int = 20, 0, 10 // default value
	var err error
	redisDB, err = strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		logger.Get().Error("Invalid REDIS_DB, using default", zap.Int("value", redisDB), zap.Error(err))
	}
	messagesPageSize, err = strconv.Atoi(getEnv("MESSAGE_PAGE_SIZE", "20"))
	if err != nil {
		logger.Get().Error("Invalid MESSAGE_PAGE_SIZE, using default", zap.Int("value", messagesPageSize), zap.Error(err))
	}
	cacheTTL, err := time.ParseDuration(getEnv("CACHE_TTL_HOURS", "6h"))
	if err != nil {
		logger.Get().Error("Invalid CACHE_TTL_HOURS, using default", zap.Duration("value", cacheTTL), zap.Error(err))
	}
	workerPoolSizeEnv := getEnv("WORKER_POOL_SIZE", "")
	if workerPoolSizeEnv != "" {
		if val, err := strconv.Atoi(workerPoolSizeEnv); err == nil {
			workerPoolSize = val
		} else {
			logger.Get().Error("Invalid CONSUMER_POOL_SIZE, using default", zap.String("value", workerPoolSizeEnv), zap.Error(err))
		}
	}

	once.Do(func() {
		instance = &Config{
			Port:                getEnv("PORT", "50051"),
			CacheTTL:            cacheTTL,
			MessagePageSize:     messagesPageSize,
			MongoDBName:         getEnv("MONGO_DB_NAME", "darda_chat"),
			MongoCollectionName: getEnv("MONGO_COLLECTION_NAME", "messages"),
			MongoAddr:           getEnv("MONGO_ADDR", "mongodb://localhost:27017"),
			MongoTimeout:        getEnv("MONGO_TIMEOUT", "10s"),
			MongoUser:           getEnv("MONGO_USER", "root"),
			MongoPass:           getEnv("MONGO_PASS", ""),
			RedisAddr:           getEnv("REDIS_ADDR", "localhost:6379"),
			RedisPass:           getEnv("REDIS_PASS", ""),
			RedisDB:             redisDB,
			WorkerPoolSize:      workerPoolSize,
			AMQPUser:            getEnv("AMQP_USER", ""),
			AMQPPass:            getEnv("AMQP_PASS", ""),
			AMQPHost:            getEnv("AMQP_HOST", ""),
			MsgQueue:            getEnv("MSG_QUEUE", "conversation.cache"),
			MsgExchange:         getEnv("MSG_EXCHANGE", "message.dispatched"),
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
