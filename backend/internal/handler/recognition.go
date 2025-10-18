package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/models"
)

// RecognitionHandler handles face recognition requests
type RecognitionHandler struct {
	// Add service dependencies here when implemented
}

// NewRecognitionHandler creates a new RecognitionHandler
func NewRecognitionHandler() *RecognitionHandler {
	return &RecognitionHandler{}
}

// PostRecognize handles POST /recognize
func (h *RecognitionHandler) PostRecognize(c *gin.Context) {
	// Parse query parameters
	topKStr := c.DefaultQuery("top_k", "3")
	topK, err := strconv.Atoi(topKStr)
	if err != nil || topK < 1 || topK > 10 {
		errors.RespondWithError(c, errors.BadRequest("Invalid top_k parameter"))
		return
	}

	minScoreStr := c.DefaultQuery("min_score", "0.6")
	minScore, err := strconv.ParseFloat(minScoreStr, 64)
	if err != nil || minScore < 0 || minScore > 1 {
		errors.RespondWithError(c, errors.BadRequest("Invalid min_score parameter"))
		return
	}

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

	// TODO: Implement actual face recognition logic
	_ = topK
	_ = minScore

	// Mock response
	summary := "前回の会話で技術的な議論をしました"
	response := models.RecognitionResponse{
		Status: models.RecognitionStatusKnown,
		BestMatch: &models.RecognitionCandidate{
			PersonID:    "p-12345",
			Name:        "山田 太郎",
			Score:       0.92,
			LastSummary: &summary,
		},
		Candidates: []models.RecognitionCandidate{
			{
				PersonID:    "p-12345",
				Name:        "山田 太郎",
				Score:       0.92,
				LastSummary: &summary,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}
