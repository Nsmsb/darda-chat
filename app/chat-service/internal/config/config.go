package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
)

// Config holds the configuration values for the application.
type Config struct {
	Port                     string
	RedisAddr                string
	RedisPass                string
	RedisDB                  int
	SubsChanBufferSize       int // Buffer size for subscription channels
	AMQPUser                 string
	AMQPPass                 string
	AMQPHost                 string
	MsgQueue                 string
	MessageReaderServiceAddr string
	Env                      string
	CORSConfig               cors.Config
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
			CORSConfig:               setupCORS("CORS_ALLOWED_ORIGINS"),
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

// setupCORS configures CORS settings based on environment variables.
func setupCORS(corsEnvVar string) cors.Config {

	allowedOrigins := os.Getenv(corsEnvVar) // e.g. "https://myapp.com,https://admin.myapp.com"
	return cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowOrigins:     strings.Split(allowedOrigins, ","),
	}
}
