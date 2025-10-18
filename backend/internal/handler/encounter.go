package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/models"
)

// EncounterHandler handles encounter log requests
type EncounterHandler struct {
	// Add service dependencies here when implemented
}

// NewEncounterHandler creates a new EncounterHandler
func NewEncounterHandler() *EncounterHandler {
	return &EncounterHandler{}
}

// ListEncounters handles GET /persons/{person_id}/encounters
func (h *EncounterHandler) ListEncounters(c *gin.Context) {
	personID := c.Param("person_id")

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		errors.RespondWithError(c, errors.BadRequest("Invalid limit parameter"))
		return
	}

	cursor := c.Query("cursor")

	// TODO: Implement actual repository query
	_ = personID
	_ = limit
	_ = cursor

	// Mock response
	summary := "技術的な議論について話しました"
	encounters := []models.Encounter{
		{
			EncounterID:  "e-abc123",
			PersonID:     personID,
			RecognizedAt: time.Now(),
			Score:        0.92,
			Summary:      &summary,
		},
	}

	response := models.EncounterList{
		Items:      encounters,
		NextCursor: nil,
	}

	c.JSON(http.StatusOK, response)
}
