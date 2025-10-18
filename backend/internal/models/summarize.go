package models

// SummarizeRequest represents a request to summarize text
type SummarizeRequest struct {
	Text string `json:"text" binding:"required"`
}

// SummarizeResponse represents a response from the summarize endpoint
type SummarizeResponse struct {
	Summary string `json:"summary"`
}
