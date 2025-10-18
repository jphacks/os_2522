package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
	"github.com/jphacks/os_2522/backend/internal/models"
	"github.com/jphacks/os_2522/backend/internal/service"
)

// SummarizeHandler handles summarization requests
type SummarizeHandler struct {
	summarizeService *service.SummarizeService
}

// NewSummarizeHandler creates a new SummarizeHandler
func NewSummarizeHandler(summarizeService *service.SummarizeService) *SummarizeHandler {
	return &SummarizeHandler{
		summarizeService: summarizeService,
	}
}

// PostSummarize handles POST /summarize
func (h *SummarizeHandler) PostSummarize(c *gin.Context) {
	var req models.SummarizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.RespondWithError(c, errors.BadRequest("Invalid request body"))
		return
	}

	if req.Text == "" {
		errors.RespondWithError(c, errors.BadRequest("Text is required"))
		return
	}

	summary, err := h.summarizeService.Summarize(c.Request.Context(), req.Text)
	if err != nil {
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	resp := models.SummarizeResponse{
		Summary: summary,
	}

	c.JSON(http.StatusOK, resp)
}
