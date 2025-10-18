package handler

import (
	"bytes"
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

func (m *MockRecognitionService) Recognize(file *multipart.FileHeader, topK int, minScore float64) (*models.RecognitionResponse, error) {
	args := m.Called(file, topK, minScore)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RecognitionResponse), args.Error(1)
}

func TestRecognitionHandler_PostRecognize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		setupRequest   func() (*bytes.Buffer, string)
		mockSetup      func(*MockRecognitionService)
		expectedStatus int
	}{
		{
			name:        "successful recognition - known person",
			queryParams: "",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{`form-data; name="image"; filename="test.jpg"`}
				h["Content-Type"] = []string{"image/jpeg"}
				part, _ := writer.CreatePart(h)
				part.Write([]byte("fake image data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup: func(m *MockRecognitionService) {
				m.On("Recognize", mock.AnythingOfType("*multipart.FileHeader"), 3, 0.6).Return(&models.RecognitionResponse{
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
		},
		{
			name:        "successful recognition - unknown person",
			queryParams: "",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{`form-data; name="image"; filename="test.jpg"`}
				h["Content-Type"] = []string{"image/jpeg"}
				part, _ := writer.CreatePart(h)
				part.Write([]byte("fake image data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup: func(m *MockRecognitionService) {
				m.On("Recognize", mock.AnythingOfType("*multipart.FileHeader"), 3, 0.6).Return(&models.RecognitionResponse{
					Status:     models.RecognitionStatusUnknown,
					BestMatch:  nil,
					Candidates: []models.RecognitionCandidate{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "custom top_k and min_score",
			queryParams: "?top_k=5&min_score=0.8",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{`form-data; name="image"; filename="test.jpg"`}
				h["Content-Type"] = []string{"image/jpeg"}
				part, _ := writer.CreatePart(h)
				part.Write([]byte("fake image data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup: func(m *MockRecognitionService) {
				m.On("Recognize", mock.AnythingOfType("*multipart.FileHeader"), 5, 0.8).Return(&models.RecognitionResponse{
					Status:     models.RecognitionStatusUnknown,
					BestMatch:  nil,
					Candidates: []models.RecognitionCandidate{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "missing image file",
			queryParams: "",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid top_k - too low",
			queryParams: "?top_k=0",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("image", "test.jpg")
				part.Write([]byte("fake image data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid top_k - too high",
			queryParams: "?top_k=11",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("image", "test.jpg")
				part.Write([]byte("fake image data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid min_score - negative",
			queryParams: "?min_score=-0.1",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("image", "test.jpg")
				part.Write([]byte("fake image data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid min_score - too high",
			queryParams: "?min_score=1.1",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("image", "test.jpg")
				part.Write([]byte("fake image data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "unsupported media type",
			queryParams: "",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{`form-data; name="image"; filename="test.gif"`}
				h["Content-Type"] = []string{"image/gif"}
				part, _ := writer.CreatePart(h)
				part.Write([]byte("fake image data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup:      func(m *MockRecognitionService) {},
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:        "service error",
			queryParams: "",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				h := make(map[string][]string)
				h["Content-Disposition"] = []string{`form-data; name="image"; filename="test.jpg"`}
				h["Content-Type"] = []string{"image/jpeg"}
				part, _ := writer.CreatePart(h)
				part.Write([]byte("fake image data"))
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup: func(m *MockRecognitionService) {
				m.On("Recognize", mock.AnythingOfType("*multipart.FileHeader"), 3, 0.6).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRecognitionService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewRecognitionHandler(mockService)
			router.POST("/recognize", handler.PostRecognize)

			body, contentType := tt.setupRequest()
			req, _ := http.NewRequest(http.MethodPost, "/recognize"+tt.queryParams, body)
			req.Header.Set("Content-Type", contentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
