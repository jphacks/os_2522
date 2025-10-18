package models

// FaceEmbeddingRequest represents a request to add a face embedding
type FaceEmbeddingRequest struct {
	Embedding       []float32 `json:"embedding" binding:"required,len=512"`
	EmbeddingDim    int       `json:"embedding_dim" binding:"required,eq=512"`
	ModelVersion    string    `json:"model_version" binding:"required"`
	Note            *string   `json:"note,omitempty" binding:"omitempty,max=2000"`
	SourceImageHash *string   `json:"source_image_hash,omitempty"`
}
