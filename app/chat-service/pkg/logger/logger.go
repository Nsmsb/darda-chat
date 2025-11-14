package logger

import (
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	log  *zap.Logger
	once sync.Once
)

// GetLogger initializes (once) and returns a global zap.Logger instance
func GetLogger() *zap.Logger {
	once.Do(func() {
		var err error
		// If in production, use production config
		if env := os.Getenv("ENV"); env == "production" {
			// Use production config to display logs in json format for tracing tools
			log, err = zap.NewProduction()
		} else {
			// Otherwise, use development config
			log, err = zap.NewDevelopment()
		}
		if err != nil {
			panic("Failed to initialize zap logger: " + err.Error())
		}
	})
	return log
}

// GetFromContext retrieves a zap.Logger from the Gin context, or returns the global logger if not found
func GetFromContext(c *gin.Context) *zap.Logger {
	if logger, exists := c.Get("logger"); exists {
		if zapLogger, ok := logger.(*zap.Logger); ok {
			return zapLogger
		}
	}
	// Fallback to global logger if not found in context
	return GetLogger()
}
