package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/models"
)

// FaceHandler handles face management requests
type FaceHandler struct {
	faceService           FaceServiceInterface
	faceExtractionService FaceExtractionServiceInterface
}

// NewFaceHandler creates a new FaceHandler
func NewFaceHandler(faceService FaceServiceInterface, faceExtractionService FaceExtractionServiceInterface) *FaceHandler {
	return &FaceHandler{
		faceService:           faceService,
		faceExtractionService: faceExtractionService,
	}
}

// AddFace handles POST /persons/{person_id}/faces with embedding-based face addition
func (h *FaceHandler) AddFace(c *gin.Context) {
	personID := c.Param("person_id")

	var req models.FaceEmbeddingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondWithError(c, errors.BadRequest("Invalid request body: "+err.Error()))
		return
	}

	// Validate embedding dimension
	if len(req.Embedding) != req.EmbeddingDim {
		errors.RespondWithError(c, errors.UnprocessableEntity("Embedding length does not match embedding_dim"))
		return
	}

	face, err := h.faceService.AddFace(personID, &req)
	if err != nil {
		if err.Error() == "person not found" {
			errors.RespondWithError(c, errors.NotFound("Person not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, face)
}

// AddFaceImage handles POST /persons/{person_id}/faces-image
func (h *FaceHandler) AddFaceImage(c *gin.Context) {
	personID := c.Param("person_id")

	// Get image file
	imageFile, err := c.FormFile("image")
	if err != nil {
		errors.RespondWithError(c, errors.BadRequest("Image file is required"))
		return
	}

	// Get optional note
	note := c.PostForm("note")
	var notePtr *string
	if note != "" {
		notePtr = &note
	}

	// Extract embedding from image
	embedding, err := h.faceExtractionService.ExtractEmbedding(imageFile)
	if err != nil {
		errors.RespondWithError(c, errors.BadRequest(fmt.Sprintf("Failed to process image: %v", err)))
		return
	}

	// Create a request for the existing AddFace service method
	req := models.FaceEmbeddingRequest{
		Embedding:    embedding,
		EmbeddingDim: 512,                       // Assuming 512, should be a constant
		ModelVersion: "facenet-tflite-v1", // This should probably come from the extraction service
		Note:         notePtr,
	}

	// Call the existing service method to add the face
	face, err := h.faceService.AddFace(personID, &req)
	if err != nil {
		if err.Error() == "person not found" {
			errors.RespondWithError(c, errors.NotFound("Person not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusCreated, face)
}

// ListFaces handles GET /persons/{person_id}/faces with optional embedding inclusion
func (h *FaceHandler) ListFaces(c *gin.Context) {
	personID := c.Param("person_id")
	includeEmbedding := c.DefaultQuery("include_embedding", "false") == "true"

	faceList, err := h.faceService.ListFaces(personID, includeEmbedding)
	if err != nil {
		if err.Error() == "person not found" {
			errors.RespondWithError(c, errors.NotFound("Person not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, faceList)
}

// DeleteFace handles DELETE /persons/{person_id}/faces/{face_id}
func (h *FaceHandler) DeleteFace(c *gin.Context) {
	personID := c.Param("person_id")
	faceID := c.Param("face_id")

	err := h.faceService.DeleteFace(personID, faceID)
	if err != nil {
		if err.Error() == "face not found" || err.Error() == "face does not belong to this person" {
			errors.RespondWithError(c, errors.NotFound("Face not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
