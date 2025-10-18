package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/teradatakeshishou/os_2522/backend/internal/models"
	"github.com/teradatakeshishou/os_2522/backend/internal/repository"
	"gorm.io/gorm"
)

// PersonService handles person business logic
type PersonService struct {
	personRepo *repository.PersonRepository
	faceRepo   *repository.FaceRepository
}

// NewPersonService creates a new PersonService
func NewPersonService(personRepo *repository.PersonRepository, faceRepo *repository.FaceRepository) *PersonService {
	return &PersonService{
		personRepo: personRepo,
		faceRepo:   faceRepo,
	}
}

// ListPersons retrieves persons with pagination and search
func (s *PersonService) ListPersons(limit int, cursor *string, q *string) (*models.PersonList, error) {
	entities, nextCursor, err := s.personRepo.FindAll(limit, cursor, q)
	if err != nil {
		return nil, err
	}

	// Convert entities to models
	persons := make([]models.Person, len(entities))
	for i, entity := range entities {
		// Count faces for each person
		faceCount, _ := s.personRepo.CountFaces(entity.PersonID)

		persons[i] = models.Person{
			PersonID:    entity.PersonID,
			Name:        entity.Name,
			LastSummary: entity.LastSummary,
			CreatedAt:   entity.CreatedAt,
			UpdatedAt:   entity.UpdatedAt,
			FacesCount:  int(faceCount),
		}
	}

	return &models.PersonList{
		Items:      persons,
		NextCursor: nextCursor,
	}, nil
}

// GetPerson retrieves a person by ID
func (s *PersonService) GetPerson(personID string) (*models.Person, error) {
	entity, err := s.personRepo.FindByID(personID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("person not found")
		}
		return nil, err
	}

	// Count faces
	faceCount, _ := s.personRepo.CountFaces(entity.PersonID)

	return &models.Person{
		PersonID:    entity.PersonID,
		Name:        entity.Name,
		LastSummary: entity.LastSummary,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
		FacesCount:  int(faceCount),
	}, nil
}

// CreatePerson creates a new person
func (s *PersonService) CreatePerson(req *models.PersonCreate) (*models.Person, error) {
	// Generate person ID
	personID := fmt.Sprintf("p-%s", uuid.New().String()[:8])

	entity := &repository.PersonEntity{
		PersonID:  personID,
		Name:      req.Name,
		Note:      req.Note,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.personRepo.Create(entity); err != nil {
		return nil, err
	}

	return &models.Person{
		PersonID:   entity.PersonID,
		Name:       entity.Name,
		CreatedAt:  entity.CreatedAt,
		UpdatedAt:  entity.UpdatedAt,
		FacesCount: 0,
	}, nil
}

// UpdatePerson updates a person
func (s *PersonService) UpdatePerson(personID string, req *models.PersonUpdate) (*models.Person, error) {
	entity, err := s.personRepo.FindByID(personID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("person not found")
		}
		return nil, err
	}

	// Update fields if provided
	if req.Name != nil {
		entity.Name = *req.Name
	}
	if req.Note != nil {
		entity.Note = req.Note
	}
	entity.UpdatedAt = time.Now()

	if err := s.personRepo.Update(entity); err != nil {
		return nil, err
	}

	// Count faces
	faceCount, _ := s.personRepo.CountFaces(entity.PersonID)

	return &models.Person{
		PersonID:    entity.PersonID,
		Name:        entity.Name,
		LastSummary: entity.LastSummary,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
		FacesCount:  int(faceCount),
	}, nil
}

// DeletePerson soft deletes a person
func (s *PersonService) DeletePerson(personID string) error {
	// Check if person exists
	_, err := s.personRepo.FindByID(personID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("person not found")
		}
		return err
	}

	return s.personRepo.Delete(personID)
}
