package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/errors"
)

// TranscribeHandler handles transcription requests
type TranscribeHandler struct {
	jobService JobServiceInterface
}

// NewTranscribeHandler creates a new TranscribeHandler
func NewTranscribeHandler(jobService JobServiceInterface) *TranscribeHandler {
	return &TranscribeHandler{jobService: jobService}
}

// PostTranscribe handles POST /transcribe
func (h *TranscribeHandler) PostTranscribe(c *gin.Context) {
	var personID *string
	if pid := c.PostForm("person_id"); pid != "" {
		personID = &pid
	}

	file, err := c.FormFile("audio")
	if err != nil {
		errors.RespondWithError(c, errors.BadRequest("Audio file is required"))
		return
	}

	var webhookURL *string
	if wh := c.PostForm("webhook_url"); wh != "" {
		webhookURL = &wh
	}

	job, err := h.jobService.CreateTranscriptionJob(personID, file, webhookURL)
	if err != nil {
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusAccepted, job)
}

// GetJob handles GET /jobs/{job_id}
func (h *TranscribeHandler) GetJob(c *gin.Context) {
	jobID := c.Param("job_id")

	job, err := h.jobService.GetJob(jobID)
	if err != nil {
		if err.Error() == "job not found" {
			errors.RespondWithError(c, errors.NotFound("Job not found"))
			return
		}
		errors.RespondWithError(c, errors.InternalServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, job)
}
