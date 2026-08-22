package achievementhandler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	achievementApp "knowledge-graph/internal/application/achievement"
	achievementDomain "knowledge-graph/internal/domain/achievement"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockService is a mock for achievementApp.Service
type MockService struct {
	mock.Mock
}

func (m *MockService) GetAllAchievements(ctx context.Context) ([]achievementDomain.Achievement, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]achievementDomain.Achievement), args.Error(1)
}

func (m *MockService) GetUserAchievementsWithStatus(ctx context.Context, userID uuid.UUID) ([]achievementApp.UserAchievementWithStatus, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]achievementApp.UserAchievementWithStatus), args.Error(1)
}

func (m *MockService) MarkNotificationSeen(ctx context.Context, userID uuid.UUID, achievementID string) error {
	args := m.Called(ctx, userID, achievementID)
	return args.Error(0)
}

func TestListAchievements(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockService)
	handler := NewHandler(&achievementApp.Service{})
	handler.SetService(mockService)
	router := gin.New()
	router.GET("/achievements", handler.ListAchievements)

	tests := []struct {
		name       string
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name: "successful list empty",
			setupMock: func() {
				mockService.On("GetAllAchievements", mock.Anything).Return([]achievementDomain.Achievement{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "successful list with data",
			setupMock: func() {
				// Create mock achievement using domain constructors
				condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 1}
				achievement, _ := achievementDomain.NewAchievement("test-code", "Test Achievement", "Test Description", "🏆", condition, 100, false)
				mockService.On("GetAllAchievements", mock.Anything).Return([]achievementDomain.Achievement{*achievement}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "service error",
			setupMock: func() {
				mockService.On("GetAllAchievements", mock.Anything).Return(nil, assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "failed to fetch achievements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil // Reset mock
			tt.setupMock()

			req := httptest.NewRequest(http.MethodGet, "/achievements", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.wantError)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "achievements")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetUserAchievements(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockService)
	handler := NewHandler(&achievementApp.Service{})
	handler.SetService(mockService)
	router := gin.New()
	router.GET("/user/achievements", handler.GetUserAchievements)

	testUserID := uuid.New()

	tests := []struct {
		name       string
		userID     uuid.UUID
		setupMock  func()
		wantStatus int
		wantError  string
	}{
		{
			name:   "successful get",
			userID: testUserID,
			setupMock: func() {
				mockService.On("GetUserAchievementsWithStatus", mock.Anything, testUserID).Return([]achievementApp.UserAchievementWithStatus{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized",
			userID:     uuid.Nil,
			setupMock:  func() {},
			wantStatus: http.StatusUnauthorized,
			wantError:  "unauthorized",
		},
		{
			name:   "service error",
			userID: testUserID,
			setupMock: func() {
				mockService.On("GetUserAchievementsWithStatus", mock.Anything, testUserID).Return(nil, assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "failed to fetch user achievements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil // Reset mock
			tt.setupMock()

			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != uuid.Nil {
					c.Set(middleware.ContextUserIDKey, tt.userID)
				}
				c.Next()
			})
			router.GET("/user/achievements", handler.GetUserAchievements)

			req := httptest.NewRequest(http.MethodGet, "/user/achievements", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.wantError)
			}

			if tt.wantStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "achievements")
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestMarkSeen(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockService)
	handler := NewHandler(&achievementApp.Service{})
	handler.SetService(mockService)
	router := gin.New()
	router.POST("/achievements/:id/seen", handler.MarkSeen)

	testUserID := uuid.New()
	testAchievementID := uuid.New().String()

	tests := []struct {
		name          string
		userID        uuid.UUID
		achievementID string
		setupMock     func()
		wantStatus    int
		wantError     string
	}{
		{
			name:          "successful mark",
			userID:        testUserID,
			achievementID: testAchievementID,
			setupMock: func() {
				mockService.On("MarkNotificationSeen", mock.Anything, testUserID, testAchievementID).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:          "unauthorized",
			userID:        uuid.Nil,
			achievementID: testAchievementID,
			setupMock:     func() {},
			wantStatus:    http.StatusUnauthorized,
			wantError:     "unauthorized",
		},
		{
			name:          "missing achievement id",
			userID:        testUserID,
			achievementID: "",
			setupMock:     func() {},
			wantStatus:    http.StatusBadRequest,
			wantError:     "missing achievement id",
		},
		{
			name:          "service error",
			userID:        testUserID,
			achievementID: testAchievementID,
			setupMock: func() {
				mockService.On("MarkNotificationSeen", mock.Anything, testUserID, testAchievementID).Return(assert.AnError)
			},
			wantStatus: http.StatusInternalServerError,
			wantError:  "failed to mark notification as seen",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil // Reset mock
			tt.setupMock()

			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != uuid.Nil {
					c.Set(middleware.ContextUserIDKey, tt.userID)
				}
				c.Next()
			})
			router.POST("/achievements/:id/seen", handler.MarkSeen)

			req := httptest.NewRequest(http.MethodPost, "/achievements/"+tt.achievementID+"/seen", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.wantError)
			}

			mockService.AssertExpectations(t)
		})
	}
}
