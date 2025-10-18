package models

import "time"

// Face represents a face entity
type Face struct {
	FaceID       string    `json:"face_id"`
	PersonID     string    `json:"person_id"`
	ImageURL     *string   `json:"image_url,omitempty"`
	EmbeddingDim int       `json:"embedding_dim"`
	Note         *string   `json:"note,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// FaceList represents a list of faces
type FaceList struct {
	Items []Face `json:"items"`
}
