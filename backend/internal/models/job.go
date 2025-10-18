package models

import "time"

// JobStatus represents the status of an async job
type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusSucceeded JobStatus = "succeeded"
	JobStatusFailed    JobStatus = "failed"
)

// TranscriptionResult represents the result of transcription
type TranscriptionResult struct {
	PersonID    *string `json:"person_id,omitempty"`
	Transcript  string  `json:"transcript"`
	Summary     string  `json:"summary"`
	Language    string  `json:"language"`
	DurationSec float64 `json:"duration_sec"`
}

// Job represents an async job
type Job struct {
	JobID      string               `json:"job_id"`
	Status     JobStatus            `json:"status"`
	CreatedAt  time.Time            `json:"created_at"`
	FinishedAt *time.Time           `json:"finished_at,omitempty"`
	Result     *TranscriptionResult `json:"result,omitempty"`
	Error      *Problem             `json:"error,omitempty"`
}
