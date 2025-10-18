package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRecognitionService is a mock implementation of RecognitionService
type MockRecognitionService struct {
	mock.Mock
}

func (m *MockRecognitionService) Recognize(req *models.RecognitionRequest) (*models.RecognitionResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecognitionResponse), args.Error(1)
}

func TestRecognitionHandler_PostRecognize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Helper function to create a test embedding
	createTestEmbedding := func() []float32 {
		embedding := make([]float32, 512)
		for i := range embedding {
			embedding[i] = 0.5
		}
		return embedding
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockRecognitionService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful recognition - known person",
			requestBody: models.RecognitionRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
				TopK:         3,
				MinScore:     0.6,
			},
			mockSetup: func(m *MockRecognitionService) {
				m.On("Recognize", mock.AnythingOfType("*models.RecognitionRequest")).Return(&models.RecognitionResponse{
					Status: models.RecognitionStatusKnown,
					BestMatch: &models.RecognitionCandidate{
						PersonID: "p-123",
						Name:     "Test User",
						Score:    0.95,
					},
					Candidates: []models.RecognitionCandidate{
						{PersonID: "p-123", Name: "Test User", Score: 0.95},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.RecognitionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.RecognitionStatusKnown, response.Status)
				assert.NotNil(t, response.BestMatch)
				assert.Equal(t, "p-123", response.BestMatch.PersonID)
				assert.Equal(t, 0.95, response.BestMatch.Score)
			},
		},
		{
			name: "successful recognition - unknown person",
			requestBody: models.RecognitionRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
				TopK:         3,
				MinScore:     0.6,
			},
			mockSetup: func(m *MockRecognitionService) {
				m.On("Recognize", mock.AnythingOfType("*models.RecognitionRequest")).Return(&models.RecognitionResponse{
					Status:     models.RecognitionStatusUnknown,
					BestMatch:  nil,
					Candidates: []models.RecognitionCandidate{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.RecognitionResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.RecognitionStatusUnknown, response.Status)
				assert.Nil(t, response.BestMatch)
			},
		},
		{
			name: "custom top_k and min_score",
			requestBody: models.RecognitionRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
				TopK:         5,
				MinScore:     0.8,
			},
			mockSetup: func(m *MockRecognitionService) {
				m.On("Recognize", mock.MatchedBy(func(req *models.RecognitionRequest) bool {
					return req.TopK == 5 && req.MinScore == 0.8
				})).Return(&models.RecognitionResponse{
					Status:     models.RecognitionStatusUnknown,
					BestMatch:  nil,
					Candidates: []models.RecognitionCandidate{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing embedding field",
			requestBody: map[string]interface{}{
				"embedding_dim": 512,
				"model_version": "facenet-tflite-v1",
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid embedding dimension - wrong count",
			requestBody: models.RecognitionRequest{
				Embedding:    make([]float32, 256), // Wrong size
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest, // Gin validation returns 400
		},
		{
			name: "invalid embedding dimension - mismatch",
			requestBody: models.RecognitionRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 256, // Doesn't match actual size
				ModelVersion: "facenet-tflite-v1",
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest, // Gin validation returns 400
		},
		{
			name: "missing model version",
			requestBody: map[string]interface{}{
				"embedding":     createTestEmbedding(),
				"embedding_dim": 512,
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: models.RecognitionRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
			},
			mockSetup: func(m *MockRecognitionService) {
				m.On("Recognize", mock.AnythingOfType("*models.RecognitionRequest")).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "with defaults (TopK and MinScore omitted)",
			requestBody: models.RecognitionRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
				// TopK and MinScore omitted, service should use defaults
			},
			mockSetup: func(m *MockRecognitionService) {
				m.On("Recognize", mock.AnythingOfType("*models.RecognitionRequest")).Return(&models.RecognitionResponse{
					Status:     models.RecognitionStatusUnknown,
					BestMatch:  nil,
					Candidates: []models.RecognitionCandidate{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRecognitionService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewRecognitionHandler(mockService)
			router.POST("/recognize", handler.PostRecognize)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, _ := http.NewRequest(http.MethodPost, "/recognize", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}
			mockService.AssertExpectations(t)
		})
	}
}
