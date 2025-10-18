package repository

import (
	"encoding/base64"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PersonRepository handles person data access
type PersonRepository struct {
	db *gorm.DB
}

// NewPersonRepository creates a new PersonRepository
func NewPersonRepository(db *gorm.DB) *PersonRepository {
	return &PersonRepository{db: db}
}

// FindAll retrieves persons with pagination and optional search
func (r *PersonRepository) FindAll(limit int, cursor *string, q *string) ([]PersonEntity, *string, error) {
	var persons []PersonEntity
	query := r.db.Model(&PersonEntity{})

	// Apply search filter if provided
	if q != nil && *q != "" {
		query = query.Where("name LIKE ?", "%"+*q+"%")
	}

	// Apply cursor pagination
	if cursor != nil && *cursor != "" {
		decodedCursor, err := decodeCursor(*cursor)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid cursor: %w", err)
		}
		query = query.Where("created_at < ?", decodedCursor)
	}

	// Fetch limit + 1 to check if there's a next page
	query = query.Order("created_at DESC").Limit(limit + 1)

	if err := query.Find(&persons).Error; err != nil {
		return nil, nil, err
	}

	// Check if there's a next page
	var nextCursor *string
	if len(persons) > limit {
		lastPerson := persons[limit-1]
		encoded := encodeCursor(lastPerson.CreatedAt)
		nextCursor = &encoded
		persons = persons[:limit]
	}

	return persons, nextCursor, nil
}

// FindByID retrieves a person by ID
func (r *PersonRepository) FindByID(personID string) (*PersonEntity, error) {
	var person PersonEntity
	if err := r.db.First(&person, "person_id = ?", personID).Error; err != nil {
		return nil, err
	}
	return &person, nil
}

// Create creates a new person
func (r *PersonRepository) Create(person *PersonEntity) error {
	return r.db.Create(person).Error
}

// Update updates a person
func (r *PersonRepository) Update(person *PersonEntity) error {
	return r.db.Save(person).Error
}

// Delete soft deletes a person
func (r *PersonRepository) Delete(personID string) error {
	return r.db.Delete(&PersonEntity{}, "person_id = ?", personID).Error
}

// CountFaces counts the number of faces for a person
func (r *PersonRepository) CountFaces(personID string) (int64, error) {
	var count int64
	err := r.db.Model(&FaceEntity{}).Where("person_id = ?", personID).Count(&count).Error
	return count, err
}

// Helper functions for cursor encoding/decoding
func encodeCursor(t time.Time) string {
	return base64.StdEncoding.EncodeToString([]byte(t.Format(time.RFC3339Nano)))
}

func decodeCursor(cursor string) (time.Time, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339Nano, string(decoded))
}
