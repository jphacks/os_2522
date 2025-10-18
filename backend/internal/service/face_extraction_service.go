package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
)

// FaceExtractionServiceInterface defines the interface for a service that extracts face embeddings from images.
type FaceExtractionServiceInterface interface {
	ExtractEmbedding(file *multipart.FileHeader) ([]float32, error)
}

// FaceExtractionService runs a Python script to extract embeddings.
type FaceExtractionService struct {
	PythonPath string
	ScriptPath string
}

// NewFaceExtractionService creates a new FaceExtractionService.
func NewFaceExtractionService() *FaceExtractionService {
	return &FaceExtractionService{
		PythonPath: "backend/ml/.venv/bin/python",
		ScriptPath: "backend/ml/extract_embedding.py",
	}
}

// ExtractEmbedding saves the uploaded image to a temporary file and runs the Python script to get the embedding.
func (s *FaceExtractionService) ExtractEmbedding(fileHeader *multipart.FileHeader) ([]float32, error) {
	if fileHeader == nil {
		return nil, fmt.Errorf("image file is nil")
	}

	// 1. Save the uploaded file to a temporary file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	tempFile, err := os.CreateTemp("", "upload-*.jpg") // Assume jpg for simplicity, though script handles others
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name()) // Clean up the temp file

	_, err = io.Copy(tempFile, src)
	if err != nil {
		return nil, fmt.Errorf("failed to save to temp file: %w", err)
	}
	tempFile.Close() // Close the file so the python script can open it

	// 2. Execute the Python script
	cmd := exec.Command(s.PythonPath, s.ScriptPath, tempFile.Name())
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("python script execution failed: %w - Stderr: %s", err, stderr.String())
	}

	if stderr.Len() > 0 {
		return nil, fmt.Errorf("python script error: %s", stderr.String())
	}

	// 3. Parse the JSON output from the script
	var embedding []float32
	if err := json.Unmarshal(stdout.Bytes(), &embedding); err != nil {
		return nil, fmt.Errorf("failed to parse embedding from python script: %w", err)
	}

	return embedding, nil
}