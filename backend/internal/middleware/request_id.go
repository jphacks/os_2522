package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
	TraceIDKey      = "trace_id"
)

// RequestID generates or extracts a request ID for tracking
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get request ID from header first
		requestID := c.GetHeader(RequestIDHeader)

		// If not present, generate a new one
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set it in context for use in error responses
		c.Set(TraceIDKey, requestID)

		// Set response header
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}
