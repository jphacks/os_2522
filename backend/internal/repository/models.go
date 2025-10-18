package repository

import (
	"time"

	"gorm.io/gorm"
)

// PersonEntity represents a person in the database
type PersonEntity struct {
	PersonID    string         `gorm:"primaryKey;type:varchar(50)"`
	Name        string         `gorm:"type:varchar(100);not null;index"`
	Note        *string        `gorm:"type:text"`
	LastSummary *string        `gorm:"type:text"`
	CreatedAt   time.Time      `gorm:"not null;index"`
	UpdatedAt   time.Time      `gorm:"not null"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`

	// Relations
	Faces      []FaceEntity      `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
	Encounters []EncounterEntity `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
}

// TableName specifies the table name for PersonEntity
func (PersonEntity) TableName() string {
	return "persons"
}

// FaceEntity represents a face image in the database
type FaceEntity struct {
	FaceID       string         `gorm:"primaryKey;type:varchar(50)"`
	PersonID     string         `gorm:"type:varchar(50);not null;index"`
	ImagePath    *string        `gorm:"type:varchar(500)"`
	Embedding    []byte         `gorm:"type:bytea"` // Store as binary for flexibility
	EmbeddingDim int            `gorm:"not null;default:512"`
	Note         *string        `gorm:"type:text"`
	CreatedAt    time.Time      `gorm:"not null;index"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	// Relations
	Person PersonEntity `gorm:"foreignKey:PersonID"`
}

// TableName specifies the table name for FaceEntity
func (FaceEntity) TableName() string {
	return "faces"
}

// EncounterEntity represents an encounter log in the database
type EncounterEntity struct {
	EncounterID  string    `gorm:"primaryKey;type:varchar(50)"`
	PersonID     string    `gorm:"type:varchar(50);not null;index"`
	RecognizedAt time.Time `gorm:"not null;index"`
	Score        float64   `gorm:"type:double precision;not null"`
	Summary      *string   `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"not null;index"`

	// Relations
	Person PersonEntity `gorm:"foreignKey:PersonID"`
}

// TableName specifies the table name for EncounterEntity
func (EncounterEntity) TableName() string {
	return "encounters"
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusSucceeded JobStatus = "succeeded"
	JobStatusFailed    JobStatus = "failed"
)

// JobEntity represents an async job in the database
type JobEntity struct {
	JobID        string     `gorm:"primaryKey;type:varchar(50)"`
	PersonID     *string    `gorm:"type:varchar(50);index"`
	Status       JobStatus  `gorm:"type:varchar(20);not null;index;default:'queued'"`
	AudioPath    *string    `gorm:"type:varchar(500)"`
	WebhookURL   *string    `gorm:"type:varchar(500)"`
	Transcript   *string    `gorm:"type:text"`
	Summary      *string    `gorm:"type:text"`
	Language     *string    `gorm:"type:varchar(10)"`
	DurationSec  *float64   `gorm:"type:double precision"`
	ErrorMessage *string    `gorm:"type:text"`
	CreatedAt    time.Time  `gorm:"not null;index"`
	FinishedAt   *time.Time `gorm:"index"`
}

// TableName specifies the table name for JobEntity
func (JobEntity) TableName() string {
	return "jobs"
}
