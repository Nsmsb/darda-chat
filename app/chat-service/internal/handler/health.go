package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/nsmsb/darda-chat/app/chat-service/internal/config"
	"github.com/nsmsb/darda-chat/app/chat-service/pkg/rabbitmq"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	client *redis.Client
}

func NewHealthHandler() *HealthHandler {
	config, _ := config.Get()
	return &HealthHandler{
		client: redis.NewClient(&redis.Options{
			Addr:     config.RedisAddr,
			Password: config.RedisPass,
			DB:       config.RedisDB,
		}),
	}
}

// Liveness checks if the service is alive
func (handler *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "alive",
	})
}

// Readiness checks if the service is ready to accept requests
func (handler *HealthHandler) Readiness(c *gin.Context) {
	// Pinging Redis to check readiness
	_, err := handler.client.Ping(c).Result()
	if err != nil {
		c.JSON(500, gin.H{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	// Checking RabbitMQ connection
	_, err = rabbitmq.Conn().Channel()
	if err != nil {
		c.JSON(500, gin.H{
			"status": "not ready",
			"error":  err.Error(),
		})
		return
	}

	// If Redis and RabbitMQ are reachable, return ready status
	c.JSON(200, gin.H{
		"status": "ready",
	})
}
