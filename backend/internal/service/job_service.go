package service

import (
	"fmt"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/teradatakeshishou/os_2522/backend/internal/models"
	"github.com/teradatakeshishou/os_2522/backend/internal/repository"
	"gorm.io/gorm"
)

// JobService handles job business logic
type JobService struct {
	jobRepo *repository.JobRepository
}

// NewJobService creates a new JobService
func NewJobService(jobRepo *repository.JobRepository) *JobService {
	return &JobService{
		jobRepo: jobRepo,
	}
}

// CreateTranscriptionJob creates a new transcription job
func (s *JobService) CreateTranscriptionJob(personID *string, file *multipart.FileHeader, webhookURL *string) (*models.Job, error) {
	// Generate job ID
	jobID := fmt.Sprintf("j-%s", uuid.New().String()[:8])

	// TODO: Save audio file to storage and get path
	audioPath := fmt.Sprintf("/uploads/audio/%s", file.Filename)

	entity := &repository.JobEntity{
		JobID:      jobID,
		PersonID:   personID,
		Status:     repository.JobStatusQueued,
		AudioPath:  &audioPath,
		WebhookURL: webhookURL,
		CreatedAt:  time.Now(),
	}

	if err := s.jobRepo.Create(entity); err != nil {
		return nil, err
	}

	// TODO: Enqueue job for async processing

	return &models.Job{
		JobID:     entity.JobID,
		Status:    models.JobStatus(entity.Status),
		CreatedAt: entity.CreatedAt,
	}, nil
}

// GetJob retrieves a job by ID
func (s *JobService) GetJob(jobID string) (*models.Job, error) {
	entity, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("job not found")
		}
		return nil, err
	}

	job := &models.Job{
		JobID:      entity.JobID,
		Status:     models.JobStatus(entity.Status),
		CreatedAt:  entity.CreatedAt,
		FinishedAt: entity.FinishedAt,
	}

	// Add result if succeeded
	if entity.Status == repository.JobStatusSucceeded && entity.Transcript != nil {
		job.Result = &models.TranscriptionResult{
			PersonID:    entity.PersonID,
			Transcript:  *entity.Transcript,
			Summary:     *entity.Summary,
			Language:    *entity.Language,
			DurationSec: *entity.DurationSec,
		}
	}

	// Add error if failed
	if entity.Status == repository.JobStatusFailed && entity.ErrorMessage != nil {
		job.Error = &models.Problem{
			Type:   "https://api.example.com/problems/job-failed",
			Title:  "Job Failed",
			Status: 500,
			Detail: entity.ErrorMessage,
		}
	}

	return job, nil
}

// UpdateJobStatus updates the status of a job
func (s *JobService) UpdateJobStatus(jobID string, status models.JobStatus) error {
	now := time.Now()
	repoStatus := repository.JobStatus(status)

	var finishedAt *time.Time
	if status == models.JobStatusSucceeded || status == models.JobStatusFailed {
		finishedAt = &now
	}

	return s.jobRepo.UpdateStatus(jobID, repoStatus, finishedAt)
}
