package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/teradatakeshishou/os_2522/backend/internal/errors"
	"github.com/teradatakeshishou/os_2522/backend/internal/models"
	"github.com/teradatakeshishou/os_2522/backend/internal/service"
	"gorm.io/gorm"
)

// PersonHandler handles person-related requests
type PersonHandler struct {
	personService *service.PersonService
}

// NewPersonHandler creates a new PersonHandler
func NewPersonHandler(personService *service.PersonService) *PersonHandler {
	return &PersonHandler{
		personService: personService,
	}
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

	var cursor *string
	if c := c.Query("cursor"); c != "" {
		cursor = &c
	}

	var q *string
	if qParam := c.Query("q"); qParam != "" {
		q = &qParam
	}

	response, err := h.personService.ListPersons(limit, cursor, q)
	if err != nil {
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
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

	person, err := h.personService.CreatePerson(&req)
	if err != nil {
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	location := "/v1/persons/" + person.PersonID
	c.Header("Location", location)
	c.JSON(http.StatusCreated, person)
}

// GetPerson handles GET /persons/{person_id}
func (h *PersonHandler) GetPerson(c *gin.Context) {
	personID := c.Param("person_id")

	if personID == "" {
		errors.RespondWithError(c, errors.BadRequest("person_id is required"))
		return
	}

	person, err := h.personService.GetPerson(personID)
	if err != nil {
		if err == gorm.ErrRecordNotFound || err.Error() == "person not found" {
			errors.RespondWithError(c, errors.NotFound("Person not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
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

	person, err := h.personService.UpdatePerson(personID, &req)
	if err != nil {
		if err.Error() == "person not found" {
			errors.RespondWithError(c, errors.NotFound("Person not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, person)
}

// DeletePerson handles DELETE /persons/{person_id}
func (h *PersonHandler) DeletePerson(c *gin.Context) {
	personID := c.Param("person_id")

	err := h.personService.DeletePerson(personID)
	if err != nil {
		if err.Error() == "person not found" {
			errors.RespondWithError(c, errors.NotFound("Person not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
