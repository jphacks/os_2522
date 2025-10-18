package models

import "time"

// Face represents a face entity
type Face struct {
	FaceID            string    `json:"face_id"`
	PersonID          string    `json:"person_id"`
	ImageURL          *string   `json:"image_url,omitempty"`
	Embedding         []float32 `json:"embedding,omitempty"` // Only included when include_embedding=true
	EmbeddingDim      int       `json:"embedding_dim"`
	ModelVersion      *string   `json:"model_version,omitempty"`
	EmbeddingChecksum *string   `json:"embedding_checksum,omitempty"`
	Note              *string   `json:"note,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// FaceList represents a list of faces
type FaceList struct {
	Items []Face `json:"items"`
}
