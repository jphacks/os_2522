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
	"gorm.io/gorm"
)

// MockPersonService is a mock implementation of PersonService
type MockPersonService struct {
	mock.Mock
}

func (m *MockPersonService) ListPersons(limit int, cursor *string, q *string) (*models.PersonList, error) {
	args := m.Called(limit, cursor, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PersonList), args.Error(1)
}

func (m *MockPersonService) GetPerson(personID string) (*models.Person, error) {
	args := m.Called(personID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Person), args.Error(1)
}

func (m *MockPersonService) CreatePerson(req *models.PersonCreate) (*models.Person, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Person), args.Error(1)
}

func (m *MockPersonService) UpdatePerson(personID string, req *models.PersonUpdate) (*models.Person, error) {
	args := m.Called(personID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Person), args.Error(1)
}

func (m *MockPersonService) DeletePerson(personID string) error {
	args := m.Called(personID)
	return args.Error(0)
}

func TestPersonHandler_ListPersons(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockPersonService)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:        "successful list with defaults",
			queryParams: "",
			mockSetup: func(m *MockPersonService) {
				m.On("ListPersons", 20, (*string)(nil), (*string)(nil)).Return(&models.PersonList{
					Items:      []models.Person{{PersonID: "p-123", Name: "Test User"}},
					NextCursor: nil,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful list with limit",
			queryParams: "?limit=10",
			mockSetup: func(m *MockPersonService) {
				m.On("ListPersons", 10, (*string)(nil), (*string)(nil)).Return(&models.PersonList{
					Items:      []models.Person{{PersonID: "p-123", Name: "Test User"}},
					NextCursor: nil,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid limit parameter",
			queryParams:    "?limit=invalid",
			mockSetup:      func(m *MockPersonService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "limit too high",
			queryParams:    "?limit=200",
			mockSetup:      func(m *MockPersonService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "service error",
			queryParams: "",
			mockSetup: func(m *MockPersonService) {
				m.On("ListPersons", 20, (*string)(nil), (*string)(nil)).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPersonService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewPersonHandler(mockService)
			router.GET("/persons", handler.ListPersons)

			req, _ := http.NewRequest(http.MethodGet, "/persons"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestPersonHandler_CreatePerson(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockPersonService)
		expectedStatus int
	}{
		{
			name: "successful creation",
			requestBody: models.PersonCreate{
				Name: "Test User",
			},
			mockSetup: func(m *MockPersonService) {
				m.On("CreatePerson", mock.AnythingOfType("*models.PersonCreate")).Return(&models.Person{
					PersonID:   "p-123",
					Name:       "Test User",
					FacesCount: 0,
				}, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing required field",
			requestBody:    map[string]interface{}{},
			mockSetup:      func(m *MockPersonService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "name too long",
			requestBody: models.PersonCreate{
				Name: string(make([]byte, 101)), // 101 characters
			},
			mockSetup:      func(m *MockPersonService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			requestBody: models.PersonCreate{
				Name: "Test User",
			},
			mockSetup: func(m *MockPersonService) {
				m.On("CreatePerson", mock.AnythingOfType("*models.PersonCreate")).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPersonService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewPersonHandler(mockService)
			router.POST("/persons", handler.CreatePerson)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/persons", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusCreated {
				assert.Contains(t, w.Header().Get("Location"), "/v1/persons/")
			}
			mockService.AssertExpectations(t)
		})
	}
}

func TestPersonHandler_GetPerson(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		personID       string
		mockSetup      func(*MockPersonService)
		expectedStatus int
	}{
		{
			name:     "successful retrieval",
			personID: "p-123",
			mockSetup: func(m *MockPersonService) {
				m.On("GetPerson", "p-123").Return(&models.Person{
					PersonID:   "p-123",
					Name:       "Test User",
					FacesCount: 0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "person not found",
			personID: "p-999",
			mockSetup: func(m *MockPersonService) {
				m.On("GetPerson", "p-999").Return(nil, gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "service error",
			personID: "p-123",
			mockSetup: func(m *MockPersonService) {
				m.On("GetPerson", "p-123").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPersonService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewPersonHandler(mockService)
			router.GET("/persons/:person_id", handler.GetPerson)

			req, _ := http.NewRequest(http.MethodGet, "/persons/"+tt.personID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestPersonHandler_UpdatePerson(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		personID       string
		requestBody    interface{}
		mockSetup      func(*MockPersonService)
		expectedStatus int
	}{
		{
			name:     "successful update",
			personID: "p-123",
			requestBody: models.PersonUpdate{
				Name: stringPtr("Updated Name"),
			},
			mockSetup: func(m *MockPersonService) {
				m.On("UpdatePerson", "p-123", mock.AnythingOfType("*models.PersonUpdate")).Return(&models.Person{
					PersonID:   "p-123",
					Name:       "Updated Name",
					FacesCount: 0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "person not found",
			personID: "p-999",
			requestBody: models.PersonUpdate{
				Name: stringPtr("Updated Name"),
			},
			mockSetup: func(m *MockPersonService) {
				m.On("UpdatePerson", "p-999", mock.AnythingOfType("*models.PersonUpdate")).Return(nil, errors.New("person not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "validation error - name too long",
			personID: "p-123",
			requestBody: models.PersonUpdate{
				Name: stringPtr(string(make([]byte, 101))),
			},
			mockSetup:      func(m *MockPersonService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "service error",
			personID: "p-123",
			requestBody: models.PersonUpdate{
				Name: stringPtr("Updated Name"),
			},
			mockSetup: func(m *MockPersonService) {
				m.On("UpdatePerson", "p-123", mock.AnythingOfType("*models.PersonUpdate")).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPersonService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewPersonHandler(mockService)
			router.PATCH("/persons/:person_id", handler.UpdatePerson)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPatch, "/persons/"+tt.personID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestPersonHandler_DeletePerson(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		personID       string
		mockSetup      func(*MockPersonService)
		expectedStatus int
	}{
		{
			name:     "successful deletion",
			personID: "p-123",
			mockSetup: func(m *MockPersonService) {
				m.On("DeletePerson", "p-123").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:     "person not found",
			personID: "p-999",
			mockSetup: func(m *MockPersonService) {
				m.On("DeletePerson", "p-999").Return(errors.New("person not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "service error",
			personID: "p-123",
			mockSetup: func(m *MockPersonService) {
				m.On("DeletePerson", "p-123").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockPersonService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewPersonHandler(mockService)
			router.DELETE("/persons/:person_id", handler.DeletePerson)

			req, _ := http.NewRequest(http.MethodDelete, "/persons/"+tt.personID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
