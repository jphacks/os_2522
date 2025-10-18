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

// MockFaceService is a mock implementation of FaceService
type MockFaceService struct {
	mock.Mock
}

func (m *MockFaceService) AddFace(personID string, file *multipart.FileHeader, note *string) (*models.Face, error) {
	args := m.Called(personID, file, note)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Face), args.Error(1)
}

func (m *MockFaceService) ListFaces(personID string) (*models.FaceList, error) {
	args := m.Called(personID)
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

	tests := []struct {
		name           string
		personID       string
		setupRequest   func() (*bytes.Buffer, string)
		mockSetup      func(*MockFaceService)
		expectedStatus int
	}{
		{
			name:     "successful face addition",
			personID: "p-123",
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
			mockSetup: func(m *MockFaceService) {
				m.On("AddFace", "p-123", mock.AnythingOfType("*multipart.FileHeader"), (*string)(nil)).Return(&models.Face{
					FaceID:       "f-123",
					PersonID:     "p-123",
					EmbeddingDim: 512,
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:     "missing image file",
			personID: "p-123",
			setupRequest: func() (*bytes.Buffer, string) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				return body, writer.FormDataContentType()
			},
			mockSetup:      func(m *MockFaceService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "person not found",
			personID: "p-999",
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
			mockSetup: func(m *MockFaceService) {
				m.On("AddFace", "p-999", mock.AnythingOfType("*multipart.FileHeader"), (*string)(nil)).Return(nil, errors.New("person not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "service error",
			personID: "p-123",
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
			mockSetup: func(m *MockFaceService) {
				m.On("AddFace", "p-123", mock.AnythingOfType("*multipart.FileHeader"), (*string)(nil)).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFaceService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewFaceHandler(mockService)
			router.POST("/persons/:person_id/faces", handler.AddFace)

			body, contentType := tt.setupRequest()
			req, _ := http.NewRequest(http.MethodPost, "/persons/"+tt.personID+"/faces", body)
			req.Header.Set("Content-Type", contentType)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestFaceHandler_ListFaces(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		personID       string
		mockSetup      func(*MockFaceService)
		expectedStatus int
	}{
		{
			name:     "successful list",
			personID: "p-123",
			mockSetup: func(m *MockFaceService) {
				m.On("ListFaces", "p-123").Return(&models.FaceList{
					Items: []models.Face{
						{FaceID: "f-1", PersonID: "p-123"},
						{FaceID: "f-2", PersonID: "p-123"},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "person not found",
			personID: "p-999",
			mockSetup: func(m *MockFaceService) {
				m.On("ListFaces", "p-999").Return(nil, errors.New("person not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "service error",
			personID: "p-123",
			mockSetup: func(m *MockFaceService) {
				m.On("ListFaces", "p-123").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockFaceService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewFaceHandler(mockService)
			router.GET("/persons/:person_id/faces", handler.ListFaces)

			req, _ := http.NewRequest(http.MethodGet, "/persons/"+tt.personID+"/faces", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
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
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewFaceHandler(mockService)
			router.DELETE("/persons/:person_id/faces/:face_id", handler.DeleteFace)

			req, _ := http.NewRequest(http.MethodDelete, "/persons/"+tt.personID+"/faces/"+tt.faceID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
