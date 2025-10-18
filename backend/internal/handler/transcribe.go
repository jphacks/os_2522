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

// TranscribeHandler handles transcription requests
type TranscribeHandler struct {
	// Add service dependencies here when implemented
}

// NewTranscribeHandler creates a new TranscribeHandler
func NewTranscribeHandler() *TranscribeHandler {
	return &TranscribeHandler{}
}

// PostTranscribe handles POST /transcribe
func (h *TranscribeHandler) PostTranscribe(c *gin.Context) {
	// Get optional person_id
	personID := c.PostForm("person_id")

	// Get multipart form file
	file, err := c.FormFile("audio")
	if err != nil {
		errors.RespondWithError(c, errors.BadRequest("Audio file is required"))
		return
	}

	// Get optional webhook URL
	webhookURL := c.PostForm("webhook_url")

	// TODO: Implement actual transcription job creation
	_ = personID
	_ = webhookURL
	_ = file

	// Mock response - create a job
	jobID := fmt.Sprintf("j-%s", uuid.New().String()[:8])
	job := models.Job{
		JobID:     jobID,
		Status:    models.JobStatusQueued,
		CreatedAt: time.Now(),
	}

	c.JSON(http.StatusAccepted, job)
}

// GetJob handles GET /jobs/{job_id}
func (h *TranscribeHandler) GetJob(c *gin.Context) {
	jobID := c.Param("job_id")

	// TODO: Implement actual job status query
	_ = jobID

	// Mock response - return a completed job
	personID := "p-12345"
	job := models.Job{
		JobID:      jobID,
		Status:     models.JobStatusSucceeded,
		CreatedAt:  time.Now().Add(-5 * time.Minute),
		FinishedAt: &[]time.Time{time.Now()}[0],
		Result: &models.TranscriptionResult{
			PersonID:    &personID,
			Transcript:  "こんにちは、今日は良い天気ですね。",
			Summary:     "天候について話しました。",
			Language:    "ja",
			DurationSec: 3.5,
		},
	}

	c.JSON(http.StatusOK, job)
}
