package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jphacks/os_2522/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEncounterService is a mock implementation of EncounterService
type MockEncounterService struct {
	mock.Mock
}

func (m *MockEncounterService) ListEncounters(personID string, limit int, cursor *string) (*models.EncounterList, error) {
	args := m.Called(personID, limit, cursor)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EncounterList), args.Error(1)
}

func TestEncounterHandler_ListEncounters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		personID       string
		queryParams    string
		mockSetup      func(*MockEncounterService)
		expectedStatus int
	}{
		{
			name:        "successful list with defaults",
			personID:    "p-123",
			queryParams: "",
			mockSetup: func(m *MockEncounterService) {
				m.On("ListEncounters", "p-123", 20, (*string)(nil)).Return(&models.EncounterList{
					Items: []models.Encounter{
						{
							EncounterID:  "e-1",
							PersonID:     "p-123",
							RecognizedAt: time.Now(),
							Score:        0.95,
						},
						{
							EncounterID:  "e-2",
							PersonID:     "p-123",
							RecognizedAt: time.Now().Add(-1 * time.Hour),
							Score:        0.88,
						},
					},
					NextCursor: nil,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful list with custom limit",
			personID:    "p-123",
			queryParams: "?limit=10",
			mockSetup: func(m *MockEncounterService) {
				m.On("ListEncounters", "p-123", 10, (*string)(nil)).Return(&models.EncounterList{
					Items:      []models.Encounter{},
					NextCursor: nil,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful list with cursor",
			personID:    "p-123",
			queryParams: "?cursor=abc123",
			mockSetup: func(m *MockEncounterService) {
				cursor := "abc123"
				m.On("ListEncounters", "p-123", 20, &cursor).Return(&models.EncounterList{
					Items:      []models.Encounter{},
					NextCursor: nil,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid limit parameter - non-numeric",
			personID:       "p-123",
			queryParams:    "?limit=invalid",
			mockSetup:      func(m *MockEncounterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid limit parameter - too low",
			personID:       "p-123",
			queryParams:    "?limit=0",
			mockSetup:      func(m *MockEncounterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid limit parameter - too high",
			personID:       "p-123",
			queryParams:    "?limit=101",
			mockSetup:      func(m *MockEncounterService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "person not found",
			personID:    "p-999",
			queryParams: "",
			mockSetup: func(m *MockEncounterService) {
				m.On("ListEncounters", "p-999", 20, (*string)(nil)).Return(nil, errors.New("person not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "service error",
			personID:    "p-123",
			queryParams: "",
			mockSetup: func(m *MockEncounterService) {
				m.On("ListEncounters", "p-123", 20, (*string)(nil)).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockEncounterService)
			tt.mockSetup(mockService)

			router := gin.New()
			handler := NewEncounterHandler(mockService)
			router.GET("/persons/:person_id/encounters", handler.ListEncounters)

			req, _ := http.NewRequest(http.MethodGet, "/persons/"+tt.personID+"/encounters"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}
