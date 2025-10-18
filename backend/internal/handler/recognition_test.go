package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"mime/multipart"
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

// MockFaceExtractionService is a mock implementation of FaceExtractionService
type MockFaceExtractionService struct {
	mock.Mock
}

func (m *MockFaceExtractionService) ExtractEmbedding(file *multipart.FileHeader) ([]float32, error) {
	args := m.Called(file)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]float32), args.Error(1)
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
			mockRecogService := new(MockRecognitionService)
			mockExtractService := new(MockFaceExtractionService) // New
			tt.mockSetup(mockRecogService)

			router := gin.New()
			handler := NewRecognitionHandler(mockRecogService, mockExtractService) // Updated
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
			mockRecogService.AssertExpectations(t)
		})
	}
}

func TestRecognitionHandler_PostRecognizeImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	createTestEmbedding := func() []float32 {
		embedding := make([]float32, 512)
		for i := range embedding {
			embedding[i] = 0.123
		}
		return embedding
	}

	tests := []struct {
		name           string
		formData       map[string]string
		fileName       string
		fileContent    string
		mockSetup      func(*MockRecognitionService, *MockFaceExtractionService)
		expectedStatus int
	}{
		{
			name:        "successful recognition",
			formData:    map[string]string{"top_k": "5", "min_score": "0.7"},
			fileName:    "test.jpg",
			fileContent: "fake-image-data",
			mockSetup: func(mrs *MockRecognitionService, mfes *MockFaceExtractionService) {
				mfes.On("ExtractEmbedding", mock.AnythingOfType("*multipart.FileHeader")).Return(createTestEmbedding(), nil)
				mrs.On("Recognize", mock.MatchedBy(func(req *models.RecognitionRequest) bool {
					return req.TopK == 5 && req.MinScore == 0.7
				})).Return(&models.RecognitionResponse{Status: models.RecognitionStatusKnown}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing image file",
			formData:       map[string]string{},
			fileName:       "", // No file
			fileContent:    "",
			mockSetup:      func(mrs *MockRecognitionService, mfes *MockFaceExtractionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "embedding extraction error",
			formData:    map[string]string{},
			fileName:    "test.jpg",
			fileContent: "fake-image-data",
			mockSetup: func(mrs *MockRecognitionService, mfes *MockFaceExtractionService) {
				mfes.On("ExtractEmbedding", mock.AnythingOfType("*multipart.FileHeader")).Return(nil, errors.New("face not found"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid top_k",
			formData:    map[string]string{"top_k": "invalid"},
			fileName:    "test.jpg",
			fileContent: "fake-image-data",
			mockSetup:   func(mrs *MockRecognitionService, mfes *MockFaceExtractionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "recognition service error",
			formData:    map[string]string{},
			fileName:    "test.jpg",
			fileContent: "fake-image-data",
			mockSetup: func(mrs *MockRecognitionService, mfes *MockFaceExtractionService) {
				mfes.On("ExtractEmbedding", mock.AnythingOfType("*multipart.FileHeader")).Return(createTestEmbedding(), nil)
				mrs.On("Recognize", mock.AnythingOfType("*models.RecognitionRequest")).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRecogService := new(MockRecognitionService)
			mockExtractService := new(MockFaceExtractionService)
			tt.mockSetup(mockRecogService, mockExtractService)

			router := gin.New()
			handler := NewRecognitionHandler(mockRecogService, mockExtractService)
			router.POST("/recognize-image", handler.PostRecognizeImage)

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			if tt.fileName != "" {
				part, _ := writer.CreateFormFile("image", tt.fileName)
				_, _ = part.Write([]byte(tt.fileContent))
			}

			for key, val := range tt.formData {
				_ = writer.WriteField(key, val)
			}
			writer.Close()

			req, _ := http.NewRequest(http.MethodPost, "/recognize-image", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRecogService.AssertExpectations(t)
			mockExtractService.AssertExpectations(t)
		})
	}
}