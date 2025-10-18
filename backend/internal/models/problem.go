package models

// Problem represents an RFC 7807 Problem Details structure
type Problem struct {
	Type     string  `json:"type"`
	Title    string  `json:"title"`
	Status   int     `json:"status"`
	Detail   *string `json:"detail,omitempty"`
	Instance *string `json:"instance,omitempty"`
	TraceID  *string `json:"trace_id,omitempty"`
}
