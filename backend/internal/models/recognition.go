package models

// RecognitionStatus represents the status of face recognition
type RecognitionStatus string

const (
	RecognitionStatusKnown   RecognitionStatus = "known"
	RecognitionStatusUnknown RecognitionStatus = "unknown"
)

// RecognitionCandidate represents a potential match candidate
type RecognitionCandidate struct {
	PersonID    string  `json:"person_id"`
	Name        string  `json:"name"`
	Score       float64 `json:"score"`
	LastSummary *string `json:"last_summary,omitempty"`
}

// RecognitionResponse represents the response for face recognition
type RecognitionResponse struct {
	Status           RecognitionStatus      `json:"status"`
	BestMatch        *RecognitionCandidate  `json:"best_match,omitempty"`
	Candidates       []RecognitionCandidate `json:"candidates"`
	CreatedEncounter *Encounter             `json:"created_encounter,omitempty"`
}
