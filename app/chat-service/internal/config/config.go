package config

import (
	"os"
	"strconv"
)

// Config holds the configuration values for the application.
type Config struct {
	Port       string
	REDIS_ADDR string
	REDIS_PASS string
	REDIS_DB   int
}

// Load reads configuration from environment variables and returns a Config struct.
func Load() (*Config, error) {
	// Read REDIS_DB as an integer
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, err
	}
	return &Config{
		Port:       getEnv("PORT", "8080"),
		REDIS_ADDR: getEnv("REDIS_ADDR", "localhost:6379"),
		REDIS_PASS: getEnv("REDIS_PASS", ""),
		REDIS_DB:   redisDB,
	}, nil
}

// getEnv retrieves the value of the environment variable named by the key.
// If the variable is not present in the environment, then return the defaultValue.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
