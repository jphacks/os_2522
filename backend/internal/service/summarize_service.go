package service

import (
	"context"

	"github.com/jphacks/os_2522/backend/internal/utils"
)

// SummarizeService handles text summarization
type SummarizeService struct {
	geminiClient *utils.GeminiClient
}

// NewSummarizeService creates a new SummarizeService
func NewSummarizeService(geminiClient *utils.GeminiClient) *SummarizeService {
	return &SummarizeService{
		geminiClient: geminiClient,
	}
}

// Summarize generates a summary of the given text
func (s *SummarizeService) Summarize(ctx context.Context, text string) (string, error) {
	return s.geminiClient.Summarize(ctx, text)
}
