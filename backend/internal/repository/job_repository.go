package repository

import (
	"time"

	"gorm.io/gorm"
)

// JobRepository handles job data access
type JobRepository struct {
	db *gorm.DB
}

// NewJobRepository creates a new JobRepository
func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Create creates a new job
func (r *JobRepository) Create(job *JobEntity) error {
	return r.db.Create(job).Error
}

// FindByID retrieves a job by ID
func (r *JobRepository) FindByID(jobID string) (*JobEntity, error) {
	var job JobEntity
	if err := r.db.First(&job, "job_id = ?", jobID).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// Update updates a job
func (r *JobRepository) Update(job *JobEntity) error {
	return r.db.Save(job).Error
}

// UpdateStatus updates the status of a job
func (r *JobRepository) UpdateStatus(jobID string, status JobStatus, finishedAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if finishedAt != nil {
		updates["finished_at"] = finishedAt
	}
	return r.db.Model(&JobEntity{}).Where("job_id = ?", jobID).Updates(updates).Error
}

// FindPendingJobs retrieves jobs that are queued or running
func (r *JobRepository) FindPendingJobs(limit int) ([]JobEntity, error) {
	var jobs []JobEntity
	err := r.db.Where("status IN ?", []JobStatus{JobStatusQueued, JobStatusRunning}).
		Order("created_at ASC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}
