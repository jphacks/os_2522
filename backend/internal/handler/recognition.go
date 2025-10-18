package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/models"
)

// RecognitionHandler handles face recognition requests
type RecognitionHandler struct {
	recognitionService    RecognitionServiceInterface
	faceExtractionService FaceExtractionServiceInterface
}

// NewRecognitionHandler creates a new RecognitionHandler
func NewRecognitionHandler(recognitionService RecognitionServiceInterface, faceExtractionService FaceExtractionServiceInterface) *RecognitionHandler {
	return &RecognitionHandler{
		recognitionService:    recognitionService,
		faceExtractionService: faceExtractionService,
	}
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

// PostRecognizeImage handles POST /recognize-image
func (h *RecognitionHandler) PostRecognizeImage(c *gin.Context) {
	// Parse parameters from form
	topKStr := c.DefaultPostForm("top_k", "3")
	topK, err := strconv.Atoi(topKStr)
	if err != nil || topK < 1 || topK > 10 {
		errors.RespondWithError(c, errors.BadRequest("Invalid top_k parameter"))
		return
	}

	minScoreStr := c.DefaultPostForm("min_score", "0.6")
	minScore, err := strconv.ParseFloat(minScoreStr, 64)
	if err != nil || minScore < 0 || minScore > 1 {
		errors.RespondWithError(c, errors.BadRequest("Invalid min_score parameter"))
		return
	}

	// Get image file
	imageFile, err := c.FormFile("image")
	if err != nil {
		errors.RespondWithError(c, errors.BadRequest("Image file is required"))
		return
	}

	// Extract embedding from image
	embedding, err := h.faceExtractionService.ExtractEmbedding(imageFile)
	if err != nil {
		// This could be because no face was found, or the image was invalid.
		errors.RespondWithError(c, errors.BadRequest(fmt.Sprintf("Failed to process image: %v", err)))
		return
	}

	// Create a recognition request similar to the embedding-based endpoint
	req := &models.RecognitionRequest{
		Embedding:    embedding,
		EmbeddingDim: 512, // Assuming 512, should be a constant
		ModelVersion: "facenet-tflite-v1", // This should probably come from the extraction service
		TopK:         topK,
		MinScore:     minScore,
	}

	// Perform recognition
	response, err := h.recognitionService.Recognize(req)
	if err != nil {
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}
