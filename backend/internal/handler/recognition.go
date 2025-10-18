package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/models"
)

// RecognitionHandler handles face recognition requests
type RecognitionHandler struct {
	recognitionService RecognitionServiceInterface
}

// NewRecognitionHandler creates a new RecognitionHandler
func NewRecognitionHandler(recognitionService RecognitionServiceInterface) *RecognitionHandler {
	return &RecognitionHandler{recognitionService: recognitionService}
}

// PostRecognize handles POST /recognize with embedding-based recognition
func (h *RecognitionHandler) PostRecognize(c *gin.Context) {
	var req models.RecognitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondWithError(c, errors.BadRequest("Invalid request body: "+err.Error()))
		return
	}

	// Validate embedding dimension
	if len(req.Embedding) != req.EmbeddingDim {
		errors.RespondWithError(c, errors.UnprocessableEntity("Embedding length does not match embedding_dim"))
		return
	}

	response, err := h.recognitionService.Recognize(&req)
	if err != nil {
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}
