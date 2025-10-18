package repository

import (
	"fmt"

	"gorm.io/gorm"
)

// EncounterRepository handles encounter data access
type EncounterRepository struct {
	db *gorm.DB
}

// NewEncounterRepository creates a new EncounterRepository
func NewEncounterRepository(db *gorm.DB) *EncounterRepository {
	return &EncounterRepository{db: db}
}

// Create creates a new encounter
func (r *EncounterRepository) Create(encounter *EncounterEntity) error {
	return r.db.Create(encounter).Error
}

// FindByPersonID retrieves encounters for a person with pagination
func (r *EncounterRepository) FindByPersonID(personID string, limit int, cursor *string) ([]EncounterEntity, *string, error) {
	var encounters []EncounterEntity
	query := r.db.Where("person_id = ?", personID)

	// Apply cursor pagination
	if cursor != nil && *cursor != "" {
		decodedCursor, err := decodeCursor(*cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}
		query = query.Where("recognized_at < ?", decodedCursor)
	}

	// Fetch limit + 1 to check if there's a next page
	query = query.Order("recognized_at DESC").Limit(limit + 1)

	if err := query.Find(&encounters).Error; err != nil {
		return nil, nil, err
	}

	// Check if there's a next page
	var nextCursor *string
	if len(encounters) > limit {
		lastEncounter := encounters[limit-1]
		encoded := encodeCursor(lastEncounter.RecognizedAt)
		nextCursor = &encoded
		encounters = encounters[:limit]
	}

	return encounters, nextCursor, nil
}

// FindByID retrieves an encounter by ID
func (r *EncounterRepository) FindByID(encounterID string) (*EncounterEntity, error) {
	var encounter EncounterEntity
	if err := r.db.First(&encounter, "encounter_id = ?", encounterID).Error; err != nil {
		return nil, err
	}
	return &encounter, nil
}

// UpdateLastSummaryForPerson updates the last_summary field of a person based on the latest encounter
func (r *EncounterRepository) UpdateLastSummaryForPerson(personID string, summary *string) error {
	return r.db.Model(&PersonEntity{}).
		Where("person_id = ?", personID).
		Update("last_summary", summary).Error
}
