package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient wraps the Gemini AI client
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-flash")

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

// Summarize generates a summary of the given text
func (g *GeminiClient) Summarize(ctx context.Context, text string) (string, error) {
	prompt := fmt.Sprintf("以下のテキストを簡潔に要約してください:\n\n%s", text)

	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no summary generated")
	}

	summary := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	return summary, nil
}

// Close closes the Gemini client
func (g *GeminiClient) Close() error {
	return g.client.Close()
}
