package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthHandler_Healthz(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		expectedStatus int
	}{
		{
			name:           "health check returns OK",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup router
			router := gin.New()
			handler := NewHealthHandler()
			router.GET("/healthz", handler.Healthz)

			// Create request
			req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
