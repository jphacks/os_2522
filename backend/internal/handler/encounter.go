package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/teradatakeshishou/os_2522/backend/internal/errors"
	"github.com/teradatakeshishou/os_2522/backend/internal/service"
)

// EncounterHandler handles encounter log requests
type EncounterHandler struct {
	encounterService *service.EncounterService
}

// NewEncounterHandler creates a new EncounterHandler
func NewEncounterHandler(encounterService *service.EncounterService) *EncounterHandler {
	return &EncounterHandler{encounterService: encounterService}
}

// ListEncounters handles GET /persons/{person_id}/encounters
func (h *EncounterHandler) ListEncounters(c *gin.Context) {
	personID := c.Param("person_id")

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		errors.RespondWithError(c, errors.BadRequest("Invalid limit parameter"))
		return
	}

	var cursor *string
	if c := c.Query("cursor"); c != "" {
		cursor = &c
	}

	encounterList, err := h.encounterService.ListEncounters(personID, limit, cursor)
	if err != nil {
		if err.Error() == "person not found" {
			errors.RespondWithError(c, errors.NotFound("Person not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, encounterList)
}
