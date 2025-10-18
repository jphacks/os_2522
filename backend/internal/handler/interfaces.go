package handler

import (
	"mime/multipart"

	"github.com/jphacks/os_2522/backend/internal/models"
)

// PersonServiceInterface defines the interface for PersonService
type PersonServiceInterface interface {
	ListPersons(limit int, cursor *string, q *string) (*models.PersonList, error)
	GetPerson(personID string) (*models.Person, error)
	CreatePerson(req *models.PersonCreate) (*models.Person, error)
	UpdatePerson(personID string, req *models.PersonUpdate) (*models.Person, error)
	DeletePerson(personID string) error
}

// FaceServiceInterface defines the interface for FaceService
type FaceServiceInterface interface {
	AddFace(personID string, file *multipart.FileHeader, note *string) (*models.Face, error)
	ListFaces(personID string) (*models.FaceList, error)
	DeleteFace(personID, faceID string) error
}

// JobServiceInterface defines the interface for JobService
type JobServiceInterface interface {
	CreateTranscriptionJob(personID *string, file *multipart.FileHeader, webhookURL *string) (*models.Job, error)
	GetJob(jobID string) (*models.Job, error)
}

// RecognitionServiceInterface defines the interface for RecognitionService
type RecognitionServiceInterface interface {
	Recognize(file *multipart.FileHeader, topK int, minScore float64) (*models.RecognitionResponse, error)
}

// EncounterServiceInterface defines the interface for EncounterService
type EncounterServiceInterface interface {
	ListEncounters(personID string, limit int, cursor *string) (*models.EncounterList, error)
}
