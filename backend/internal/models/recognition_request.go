package models

// RecognitionRequest represents a face recognition request with pre-computed embedding
type RecognitionRequest struct {
	Embedding    []float32 `json:"embedding" binding:"required,len=512"`
	EmbeddingDim int       `json:"embedding_dim" binding:"required,eq=512"`
	ModelVersion string    `json:"model_version" binding:"required"`
	TopK         int       `json:"top_k" binding:"omitempty,min=1,max=10"`
	MinScore     float64   `json:"min_score" binding:"omitempty,min=0,max=1"`
}
