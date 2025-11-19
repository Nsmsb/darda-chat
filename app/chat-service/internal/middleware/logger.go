package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapLogger is a Gin middleware that logs requests using a zap.Logger
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		clientIP := c.ClientIP()
		requestId := c.GetString(RequestIDKey)
		method := c.Request.Method
		path := c.Request.URL.Path

		// Create a logger with request-specific fields
		reqLogger := logger.With(
			zap.String("method", method),
			zap.String("path", path),
			zap.String("request_id", requestId),
			zap.String("client_ip", clientIP),
		)
		// Store the logger in the context for use in handlers
		c.Set("logger", reqLogger)

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		reqLogger.Info("incoming request",
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
		)
	}
}
