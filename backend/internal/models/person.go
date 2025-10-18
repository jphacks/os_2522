package models

import "time"

// Person represents a person entity
type Person struct {
	PersonID    string    `json:"person_id"`
	Name        string    `json:"name"`
	LastSummary *string   `json:"last_summary,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	FacesCount  int       `json:"faces_count"`
}

// PersonCreate represents the request body for creating a person
type PersonCreate struct {
	Name            string  `json:"name" binding:"required,max=100"`
	FaceImageBase64 *string `json:"face_image_base64,omitempty"`
	Note            *string `json:"note,omitempty" binding:"omitempty,max=2000"`
}

// PersonUpdate represents the request body for updating a person
type PersonUpdate struct {
	Name *string `json:"name,omitempty" binding:"omitempty,max=100"`
	Note *string `json:"note,omitempty" binding:"omitempty,max=2000"`
}

// PersonList represents a paginated list of persons
type PersonList struct {
	Items      []Person `json:"items"`
	NextCursor *string  `json:"next_cursor,omitempty"`
}
