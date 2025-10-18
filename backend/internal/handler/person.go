package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/models"
)

// PersonHandler handles person-related requests
type PersonHandler struct {
	// Add service dependencies here when implemented
}

// NewPersonHandler creates a new PersonHandler
func NewPersonHandler() *PersonHandler {
	return &PersonHandler{}
}

// ListPersons handles GET /persons
func (h *PersonHandler) ListPersons(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		errors.RespondWithError(c, errors.BadRequest("Invalid limit parameter"))
		return
	}

	cursor := c.Query("cursor")
	q := c.Query("q")

	// TODO: Implement actual repository query
	_ = cursor
	_ = q
	_ = limit

	// Mock response for now
	persons := []models.Person{
		{
			PersonID:   "p-12345",
			Name:       "山田 太郎",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			FacesCount: 3,
		},
	}

	response := models.PersonList{
		Items:      persons,
		NextCursor: nil,
	}

	c.JSON(http.StatusOK, response)
}

// CreatePerson handles POST /persons
func (h *PersonHandler) CreatePerson(c *gin.Context) {
	var req models.PersonCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationErrors(c, err)
		return
	}

	// TODO: Implement actual person creation logic
	// For now, return a mock response
	personID := fmt.Sprintf("p-%s", uuid.New().String()[:8])
	person := models.Person{
		PersonID:   personID,
		Name:       req.Name,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		FacesCount: 0,
	}

	location := fmt.Sprintf("/v1/persons/%s", personID)
	c.Header("Location", location)
	c.JSON(http.StatusCreated, person)
}

// GetPerson handles GET /persons/{person_id}
func (h *PersonHandler) GetPerson(c *gin.Context) {
	personID := c.Param("person_id")

	// Validate person_id format
	if personID == "" {
		errors.RespondWithError(c, errors.BadRequest("person_id is required"))
		return
	}

	// TODO: Implement actual repository query
	// Mock response for now
	person := models.Person{
		PersonID:   personID,
		Name:       "山田 太郎",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		FacesCount: 3,
	}

	c.JSON(http.StatusOK, person)
}

// UpdatePerson handles PATCH /persons/{person_id}
func (h *PersonHandler) UpdatePerson(c *gin.Context) {
	personID := c.Param("person_id")

	var req models.PersonUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleValidationErrors(c, err)
		return
	}

	// TODO: Implement actual update logic
	_ = personID

	// Mock response
	person := models.Person{
		PersonID:   personID,
		Name:       *req.Name,
		CreatedAt:  time.Now().Add(-24 * time.Hour),
		UpdatedAt:  time.Now(),
		FacesCount: 3,
	}

	c.JSON(http.StatusOK, person)
}

// DeletePerson handles DELETE /persons/{person_id}
func (h *PersonHandler) DeletePerson(c *gin.Context) {
	personID := c.Param("person_id")

	// TODO: Implement actual deletion (soft delete)
	_ = personID

	c.Status(http.StatusNoContent)
}
