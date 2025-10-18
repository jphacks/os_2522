package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/service"
)

// RecognitionHandler handles face recognition requests
type RecognitionHandler struct {
	recognitionService *service.RecognitionService
}

// NewRecognitionHandler creates a new RecognitionHandler
func NewRecognitionHandler(recognitionService *service.RecognitionService) *RecognitionHandler {
	return &RecognitionHandler{recognitionService: recognitionService}
}

// PostRecognize handles POST /recognize
func (h *RecognitionHandler) PostRecognize(c *gin.Context) {
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

	file, err := c.FormFile("image")
	if err != nil {
		errors.RespondWithError(c, errors.BadRequest("Image file is required"))
		return
	}

	contentType := file.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		errors.RespondWithError(c, errors.UnsupportedMediaType("Only JPEG and PNG images are supported"))
		return
	}

	response, err := h.recognitionService.Recognize(file, topK, minScore)
	if err != nil {
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response)
}
