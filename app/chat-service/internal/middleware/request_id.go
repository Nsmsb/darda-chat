package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}

		// Set it in response so clients can trace it
		c.Writer.Header().Set("X-Request-ID", reqID)

		// Store in Gin context
		c.Set(RequestIDKey, reqID)

		c.Next()
	}
}
