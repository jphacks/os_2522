package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/models"
)

// FaceHandler handles face management requests
type FaceHandler struct {
	// Add service dependencies here when implemented
}

// NewFaceHandler creates a new FaceHandler
func NewFaceHandler() *FaceHandler {
	return &FaceHandler{}
}

// AddFace handles POST /persons/{person_id}/faces
func (h *FaceHandler) AddFace(c *gin.Context) {
	personID := c.Param("person_id")

	// Get multipart form file
	file, err := c.FormFile("image")
	if err != nil {
		errors.RespondWithError(c, errors.BadRequest("Image file is required"))
		return
	}

	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		errors.RespondWithError(c, errors.UnsupportedMediaType("Only JPEG and PNG images are supported"))
		return
	}

	note := c.PostForm("note")

	// TODO: Implement actual face embedding generation and storage
	_ = personID
	_ = note

	// Mock response
	faceID := fmt.Sprintf("f-%s", uuid.New().String()[:8])
	var notePtr *string
	if note != "" {
		notePtr = &note
	}

	face := models.Face{
		FaceID:       faceID,
		PersonID:     personID,
		EmbeddingDim: 512,
		Note:         notePtr,
		CreatedAt:    time.Now(),
	}

	c.JSON(http.StatusCreated, face)
}

// ListFaces handles GET /persons/{person_id}/faces
func (h *FaceHandler) ListFaces(c *gin.Context) {
	personID := c.Param("person_id")

	// TODO: Implement actual repository query
	_ = personID

	// Mock response
	faces := []models.Face{
		{
			FaceID:       "f-12345",
			PersonID:     personID,
			EmbeddingDim: 512,
			CreatedAt:    time.Now(),
		},
	}

	response := models.FaceList{
		Items: faces,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteFace handles DELETE /persons/{person_id}/faces/{face_id}
func (h *FaceHandler) DeleteFace(c *gin.Context) {
	personID := c.Param("person_id")
	faceID := c.Param("face_id")

	// TODO: Implement actual deletion
	_ = personID
	_ = faceID

	c.Status(http.StatusNoContent)
}
