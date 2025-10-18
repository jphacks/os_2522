package handler

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockJobService is a mock implementation of JobService
type MockJobService struct {
	mock.Mock
}

func (m *MockJobService) CreateTranscriptionJob(personID *string, file *multipart.FileHeader, webhookURL *string) (*models.Job, error) {
	args := m.Called(personID, file, webhookURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *MockJobService) GetJob(jobID string) (*models.Job, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Job), args.Error(1)
}

func TestTranscribeHandler_PostTranscribe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupRequest   func() (*bytes.Buffer, string)
		mockSetup      func(*MockJobService)
		expectedStatus int
	}{
		{
			name: "successful transcription job creation",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("audio", "test.mp3")
				part.Write([]byte("fake audio data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup: func(m *MockJobService) {
				m.On("CreateTranscriptionJob", (*string)(nil), mock.AnythingOfType("*multipart.FileHeader"), (*string)(nil)).Return(&models.Job{
					JobID:     "job-123",
					Status:    models.JobStatusQueued,
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "with person_id",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("audio", "test.mp3")
				part.Write([]byte("fake audio data"))
				writer.WriteField("person_id", "p-123")
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup: func(m *MockJobService) {
				m.On("CreateTranscriptionJob", mock.AnythingOfType("*string"), mock.AnythingOfType("*multipart.FileHeader"), (*string)(nil)).Return(&models.Job{
					JobID:     "job-123",
					Status:    models.JobStatusQueued,
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "missing audio file",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup:      func(m *MockJobService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("audio", "test.mp3")
				part.Write([]byte("fake audio data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup: func(m *MockJobService) {
				m.On("CreateTranscriptionJob", (*string)(nil), mock.AnythingOfType("*multipart.FileHeader"), (*string)(nil)).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockJobService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewTranscribeHandler(mockService)
			router.POST("/transcribe", handler.PostTranscribe)

			body, contentType := tt.setupRequest()
			req, _ := http.NewRequest(http.MethodPost, "/transcribe", body)
			req.Header.Set("Content-Type", contentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestTranscribeHandler_GetJob(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		jobID          string
		mockSetup      func(*MockJobService)
		expectedStatus int
	}{
		{
			name:  "successful job retrieval - queued",
			jobID: "job-123",
			mockSetup: func(m *MockJobService) {
				m.On("GetJob", "job-123").Return(&models.Job{
					JobID:     "job-123",
					Status:    models.JobStatusQueued,
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "successful job retrieval - succeeded",
			jobID: "job-456",
			mockSetup: func(m *MockJobService) {
				finishedAt := time.Now()
				m.On("GetJob", "job-456").Return(&models.Job{
					JobID:      "job-456",
					Status:     models.JobStatusSucceeded,
					CreatedAt:  time.Now().Add(-5 * time.Minute),
					FinishedAt: &finishedAt,
					Result: &models.TranscriptionResult{
						Transcript:  "Test transcript",
						Summary:     "Test summary",
						Language:    "ja",
						DurationSec: 120.5,
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "job not found",
			jobID: "job-999",
			mockSetup: func(m *MockJobService) {
				m.On("GetJob", "job-999").Return(nil, errors.New("job not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "service error",
			jobID: "job-123",
			mockSetup: func(m *MockJobService) {
				m.On("GetJob", "job-123").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockJobService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewTranscribeHandler(mockService)
			router.GET("/jobs/:job_id", handler.GetJob)

			req, _ := http.NewRequest(http.MethodGet, "/jobs/"+tt.jobID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
