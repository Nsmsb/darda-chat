package config

import (
	"os"
	"strconv"
	"sync"
)

// Config holds the configuration values for the application.
type Config struct {
	Port                     string
	RedisAddr                string
	RedisPass                string
	RedisDB                  int
	SubsChanBufferSize       int
	AMQPUser                 string
	AMQPPass                 string
	AMQPHost                 string
	MsgQueue                 string
	MessageReaderServiceAddr string
	Env                      string
}

var (
	instance *Config
	once     sync.Once
)

// Get returns the singleton instance of Config, it reads the configs only once.
func Get() (*Config, error) {
	var err error
	once.Do(func() {
		var redisDB, subsChanBufferSize int
		redisDB, err = strconv.Atoi(getEnv("REDIS_DB", "0"))
		if err != nil {
			return
		}
		subsChanBufferSize, err = strconv.Atoi(getEnv("SUBS_CHAN_BUFFER_SIZE", "30"))
		if err != nil {
			return
		}
		instance = &Config{
			Port:                     getEnv("PORT", "8080"),
			RedisAddr:                getEnv("REDIS_ADDR", "localhost:6379"),
			RedisPass:                getEnv("REDIS_PASS", ""),
			RedisDB:                  redisDB,
			SubsChanBufferSize:       subsChanBufferSize,
			AMQPUser:                 getEnv("AMQP_USER", ""),
			AMQPPass:                 getEnv("AMQP_PASS", ""),
			AMQPHost:                 getEnv("AMQP_HOST", ""),
			MsgQueue:                 getEnv("MSG_QUEUE", "messages"),
			MessageReaderServiceAddr: getEnv("MESSAGE_READER_SERVICE_ADDR", "localhost:50051"),
			Env:                      getEnv("ENV", "development"),
		}
	})
	return instance, err
}

// getEnv retrieves the value of the environment variable named by the key.
// If the variable is not present in the environment, then return the defaultValue.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
