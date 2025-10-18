package handler

import (
	"bytes"
	"encoding/json"
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

// MockFaceService is a mock implementation of FaceService
type MockFaceService struct {
	mock.Mock
}

func (m *MockFaceService) AddFace(personID string, req *models.FaceEmbeddingRequest) (*models.Face, error) {
	args := m.Called(personID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Face), args.Error(1)
}

func (m *MockFaceService) ListFaces(personID string, includeEmbedding bool) (*models.FaceList, error) {
	args := m.Called(personID, includeEmbedding)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.FaceList), args.Error(1)
}

func (m *MockFaceService) DeleteFace(personID, faceID string) error {
	args := m.Called(personID, faceID)
	return args.Error(0)
}

func TestFaceHandler_AddFace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Helper function to create test embedding
	createTestEmbedding := func() []float32 {
		embedding := make([]float32, 512)
		for i := range embedding {
			embedding[i] = 0.5
		}
		return embedding
	}

	tests := []struct {
		name           string
		personID       string
		requestBody    interface{}
		mockSetup      func(*MockFaceService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:     "successful face addition",
			personID: "p-123",
			requestBody: models.FaceEmbeddingRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
			},
			mockSetup: func(m *MockFaceService) {
				m.On("AddFace", "p-123", mock.AnythingOfType("*models.FaceEmbeddingRequest")).Return(&models.Face{
					FaceID:       "f-123",
					PersonID:     "p-123",
					EmbeddingDim: 512,
					ModelVersion: stringPtr("facenet-tflite-v1"),
					CreatedAt:    time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.Face
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "f-123", response.FaceID)
				assert.Equal(t, "p-123", response.PersonID)
				assert.Equal(t, 512, response.EmbeddingDim)
			},
		},
		{
			name:     "successful face addition with note",
			personID: "p-123",
			requestBody: models.FaceEmbeddingRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
				Note:         stringPtr("Front facing photo"),
			},
			mockSetup: func(m *MockFaceService) {
				m.On("AddFace", "p-123", mock.MatchedBy(func(req *models.FaceEmbeddingRequest) bool {
					return req.Note != nil && *req.Note == "Front facing photo"
				})).Return(&models.Face{
					FaceID:       "f-123",
					PersonID:     "p-123",
					EmbeddingDim: 512,
					Note:         stringPtr("Front facing photo"),
					CreatedAt:    time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "successful face addition with source image hash",
			personID: "p-123",
			requestBody: models.FaceEmbeddingRequest{
				Embedding:       createTestEmbedding(),
				EmbeddingDim:    512,
				ModelVersion:    "facenet-tflite-v1",
				SourceImageHash: stringPtr("sha256:abc123"),
			},
			mockSetup: func(m *MockFaceService) {
				m.On("AddFace", "p-123", mock.AnythingOfType("*models.FaceEmbeddingRequest")).Return(&models.Face{
					FaceID:       "f-123",
					PersonID:     "p-123",
					EmbeddingDim: 512,
					CreatedAt:    time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "missing embedding field",
			personID: "p-123",
			requestBody: map[string]interface{}{
				"embedding_dim": 512,
				"model_version": "facenet-tflite-v1",
			},
			mockSetup:      func(m *MockFaceService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "invalid embedding dimension - wrong count",
			personID: "p-123",
			requestBody: models.FaceEmbeddingRequest{
				Embedding:    make([]float32, 256), // Wrong size
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
			},
			mockSetup:      func(m *MockFaceService) {},
			expectedStatus: http.StatusBadRequest, // Gin validation returns 400
		},
		{
			name:     "invalid embedding dimension - mismatch",
			personID: "p-123",
			requestBody: models.FaceEmbeddingRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 256, // Doesn't match actual size
				ModelVersion: "facenet-tflite-v1",
			},
			mockSetup:      func(m *MockFaceService) {},
			expectedStatus: http.StatusBadRequest, // Gin validation returns 400
		},
		{
			name:     "missing model version",
			personID: "p-123",
			requestBody: map[string]interface{}{
				"embedding":     createTestEmbedding(),
				"embedding_dim": 512,
			},
			mockSetup:      func(m *MockFaceService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "person not found",
			personID: "p-999",
			requestBody: models.FaceEmbeddingRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
			},
			mockSetup: func(m *MockFaceService) {
				m.On("AddFace", "p-999", mock.AnythingOfType("*models.FaceEmbeddingRequest")).Return(nil, errors.New("person not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "service error",
			personID: "p-123",
			requestBody: models.FaceEmbeddingRequest{
				Embedding:    createTestEmbedding(),
				EmbeddingDim: 512,
				ModelVersion: "facenet-tflite-v1",
			},
			mockSetup: func(m *MockFaceService) {
				m.On("AddFace", "p-123", mock.AnythingOfType("*models.FaceEmbeddingRequest")).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid JSON",
			personID:       "p-123",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockFaceService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFaceService)
			mockExtractionService := new(MockFaceExtractionService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewFaceHandler(mockService, mockExtractionService)
			router.POST("/persons/:person_id/faces", handler.AddFace)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, _ := http.NewRequest(http.MethodPost, "/persons/"+tt.personID+"/faces", bytes.NewBuffer(body))
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

func TestFaceHandler_ListFaces(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		personID       string
		queryParams    string
		mockSetup      func(*MockFaceService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful list without embeddings",
			personID:    "p-123",
			queryParams: "",
			mockSetup: func(m *MockFaceService) {
				m.On("ListFaces", "p-123", false).Return(&models.FaceList{
					Items: []models.Face{
						{FaceID: "f-1", PersonID: "p-123", EmbeddingDim: 512},
						{FaceID: "f-2", PersonID: "p-123", EmbeddingDim: 512},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.FaceList
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Items, 2)
				// Embeddings should not be included
				assert.Nil(t, response.Items[0].Embedding)
			},
		},
		{
			name:        "successful list with embeddings",
			personID:    "p-123",
			queryParams: "?include_embedding=true",
			mockSetup: func(m *MockFaceService) {
				testEmbedding := make([]float32, 512)
				for i := range testEmbedding {
					testEmbedding[i] = 0.5
				}
				m.On("ListFaces", "p-123", true).Return(&models.FaceList{
					Items: []models.Face{
						{FaceID: "f-1", PersonID: "p-123", Embedding: testEmbedding, EmbeddingDim: 512},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.FaceList
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Items, 1)
				// Embeddings should be included
				assert.NotNil(t, response.Items[0].Embedding)
				assert.Equal(t, 512, len(response.Items[0].Embedding))
			},
		},
		{
			name:        "empty list",
			personID:    "p-123",
			queryParams: "",
			mockSetup: func(m *MockFaceService) {
				m.On("ListFaces", "p-123", false).Return(&models.FaceList{
					Items: []models.Face{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response models.FaceList
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Empty(t, response.Items)
			},
		},
		{
			name:        "person not found",
			personID:    "p-999",
			queryParams: "",
			mockSetup: func(m *MockFaceService) {
				m.On("ListFaces", "p-999", false).Return(nil, errors.New("person not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "service error",
			personID:    "p-123",
			queryParams: "",
			mockSetup: func(m *MockFaceService) {
				m.On("ListFaces", "p-123", false).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFaceService)
			mockExtractionService := new(MockFaceExtractionService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewFaceHandler(mockService, mockExtractionService)
			router.GET("/persons/:person_id/faces", handler.ListFaces)

			req, _ := http.NewRequest(http.MethodGet, "/persons/"+tt.personID+"/faces"+tt.queryParams, nil)
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

func TestFaceHandler_DeleteFace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		personID       string
		faceID         string
		mockSetup      func(*MockFaceService)
		expectedStatus int
	}{
		{
			name:     "successful deletion",
			personID: "p-123",
			faceID:   "f-123",
			mockSetup: func(m *MockFaceService) {
				m.On("DeleteFace", "p-123", "f-123").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:     "face not found",
			personID: "p-123",
			faceID:   "f-999",
			mockSetup: func(m *MockFaceService) {
				m.On("DeleteFace", "p-123", "f-999").Return(errors.New("face not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "face does not belong to person",
			personID: "p-123",
			faceID:   "f-456",
			mockSetup: func(m *MockFaceService) {
				m.On("DeleteFace", "p-123", "f-456").Return(errors.New("face does not belong to this person"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "service error",
			personID: "p-123",
			faceID:   "f-123",
			mockSetup: func(m *MockFaceService) {
				m.On("DeleteFace", "p-123", "f-123").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFaceService)
			mockExtractionService := new(MockFaceExtractionService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewFaceHandler(mockService, mockExtractionService)
			router.DELETE("/persons/:person_id/faces/:face_id", handler.DeleteFace)

			req, _ := http.NewRequest(http.MethodDelete, "/persons/"+tt.personID+"/faces/"+tt.faceID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestFaceHandler_AddFaceImage(t *testing.T) {
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
		personID       string
		formData       map[string]string
		fileName       string
		fileContent    string
		mockSetup      func(*MockFaceService, *MockFaceExtractionService)
		expectedStatus int
	}{
		{
			name:        "successful face image addition",
			personID:    "p-123",
			formData:    map[string]string{"note": "test note"},
			fileName:    "test.jpg",
			fileContent: "fake-image-data",
			mockSetup: func(mfs *MockFaceService, mfes *MockFaceExtractionService) {
				mfes.On("ExtractEmbedding", mock.AnythingOfType("*multipart.FileHeader")).Return(createTestEmbedding(), nil)
				mfs.On("AddFace", "p-123", mock.MatchedBy(func(req *models.FaceEmbeddingRequest) bool {
					return req.Note != nil && *req.Note == "test note"
				})).Return(&models.Face{FaceID: "f-new", PersonID: "p-123"}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing image file",
			personID:       "p-123",
			formData:       map[string]string{},
			fileName:       "",
			fileContent:    "",
			mockSetup:      func(mfs *MockFaceService, mfes *MockFaceExtractionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "person not found",
			personID:    "p-999",
			formData:    map[string]string{},
			fileName:    "test.jpg",
			fileContent: "fake-image-data",
			mockSetup: func(mfs *MockFaceService, mfes *MockFaceExtractionService) {
				mfes.On("ExtractEmbedding", mock.AnythingOfType("*multipart.FileHeader")).Return(createTestEmbedding(), nil)
				mfs.On("AddFace", "p-999", mock.AnythingOfType("*models.FaceEmbeddingRequest")).Return(nil, errors.New("person not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "embedding extraction error",
			personID:    "p-123",
			formData:    map[string]string{},
			fileName:    "test.jpg",
			fileContent: "fake-image-data",
			mockSetup: func(mfs *MockFaceService, mfes *MockFaceExtractionService) {
				mfes.On("ExtractEmbedding", mock.AnythingOfType("*multipart.FileHeader")).Return(nil, errors.New("extraction failed"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFaceService := new(MockFaceService)
			mockExtractionService := new(MockFaceExtractionService)
			tt.mockSetup(mockFaceService, mockExtractionService)

			router := gin.New()
			handler := NewFaceHandler(mockFaceService, mockExtractionService)
			router.POST("/persons/:person_id/faces-image", handler.AddFaceImage)

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

			req, _ := http.NewRequest(http.MethodPost, "/persons/"+tt.personID+"/faces-image", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockFaceService.AssertExpectations(t)
			mockExtractionService.AssertExpectations(t)
		})
	}
}
