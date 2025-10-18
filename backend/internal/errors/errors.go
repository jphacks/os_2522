package errors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/teradatakeshishou/os_2522/backend/internal/models"
)

// AppError represents an application error
type AppError struct {
	Type       string
	Title      string
	StatusCode int
	Detail     string
	Instance   string
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Detail)
}

// NewAppError creates a new AppError
func NewAppError(statusCode int, title, detail string) *AppError {
	return &AppError{
		Type:       fmt.Sprintf("https://api.example.com/problems/%d", statusCode),
		Title:      title,
		StatusCode: statusCode,
		Detail:     detail,
	}
}

// Common error constructors

func BadRequest(detail string) *AppError {
	return NewAppError(http.StatusBadRequest, "Bad Request", detail)
}

func Unauthorized(detail string) *AppError {
	return NewAppError(http.StatusUnauthorized, "Unauthorized", detail)
}

func NotFound(detail string) *AppError {
	return NewAppError(http.StatusNotFound, "Not Found", detail)
}

func Conflict(detail string) *AppError {
	return NewAppError(http.StatusConflict, "Conflict", detail)
}

func PayloadTooLarge(detail string) *AppError {
	return NewAppError(http.StatusRequestEntityTooLarge, "Payload Too Large", detail)
}

func UnsupportedMediaType(detail string) *AppError {
	return NewAppError(http.StatusUnsupportedMediaType, "Unsupported Media Type", detail)
}

func InternalServerError(detail string) *AppError {
	return NewAppError(http.StatusInternalServerError, "Internal Server Error", detail)
}

// RespondWithError sends a RFC 7807 compliant error response
func RespondWithError(c *gin.Context, err *AppError) {
	traceID := c.GetString("trace_id")
	instance := c.Request.URL.Path

	problem := models.Problem{
		Type:     err.Type,
		Title:    err.Title,
		Status:   err.StatusCode,
		Detail:   &err.Detail,
		Instance: &instance,
		TraceID:  &traceID,
	}

	c.JSON(err.StatusCode, problem)
}

// HandleValidationErrors converts validation errors to Problem Details
func HandleValidationErrors(c *gin.Context, validationErr error) {
	detail := fmt.Sprintf("Validation failed: %v", validationErr)
	err := BadRequest(detail)
	RespondWithError(c, err)
}
