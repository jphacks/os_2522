package service

import (
	"fmt"

	"github.com/teradatakeshishou/os_2522/backend/internal/models"
	"github.com/teradatakeshishou/os_2522/backend/internal/repository"
	"gorm.io/gorm"
)

// EncounterService handles encounter business logic
type EncounterService struct {
	encounterRepo *repository.EncounterRepository
	personRepo    *repository.PersonRepository
}

// NewEncounterService creates a new EncounterService
func NewEncounterService(encounterRepo *repository.EncounterRepository, personRepo *repository.PersonRepository) *EncounterService {
	return &EncounterService{
		encounterRepo: encounterRepo,
		personRepo:    personRepo,
	}
}

// ListEncounters retrieves encounters for a person with pagination
func (s *EncounterService) ListEncounters(personID string, limit int, cursor *string) (*models.EncounterList, error) {
	// Verify person exists
	_, err := s.personRepo.FindByID(personID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("person not found")
		}
		return nil, err
	}

	entities, nextCursor, err := s.encounterRepo.FindByPersonID(personID, limit, cursor)
	if err != nil {
		return nil, err
	}

	encounters := make([]models.Encounter, len(entities))
	for i, entity := range entities {
		encounters[i] = models.Encounter{
			EncounterID:  entity.EncounterID,
			PersonID:     entity.PersonID,
			RecognizedAt: entity.RecognizedAt,
			Score:        entity.Score,
			Summary:      entity.Summary,
		}
	}

	return &models.EncounterList{
		Items:      encounters,
		NextCursor: nextCursor,
	}, nil
}
