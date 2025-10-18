package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Healthz handles GET /healthz
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.Status(http.StatusOK)
}
