package models

import "time"

// Encounter represents an encounter event
type Encounter struct {
	EncounterID  string    `json:"encounter_id"`
	PersonID     string    `json:"person_id"`
	RecognizedAt time.Time `json:"recognized_at"`
	Score        float64   `json:"score"`
	Summary      *string   `json:"summary,omitempty"`
}

// EncounterList represents a paginated list of encounters
type EncounterList struct {
	Items      []Encounter `json:"items"`
	NextCursor *string     `json:"next_cursor,omitempty"`
}
