package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
)

const (
	APIKeyHeader = "X-API-Key"
)

// APIKeyAuth validates the API key from request header
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get expected API key from environment
		expectedKey := os.Getenv("API_KEY")
		if expectedKey == "" {
			// For development, allow bypassing if no API key is set
			c.Next()
			return
		}

		// Get API key from header
		apiKey := c.GetHeader(APIKeyHeader)
		apiKey = strings.TrimSpace(apiKey)

		if apiKey == "" {
			err := errors.Unauthorized("API key is missing")
			errors.RespondWithError(c, err)
			c.Abort()
			return
		}

		if apiKey != expectedKey {
			err := errors.Unauthorized("Invalid API key")
			errors.RespondWithError(c, err)
			c.Abort()
			return
		}

		c.Next()
	}
}
