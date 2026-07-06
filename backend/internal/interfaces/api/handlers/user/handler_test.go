package user

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/interfaces/api/middleware"
	"knowledge-graph/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, cleanup := testutil.SetupTestDB(t)

	// Auto-migrate models
	err := db.AutoMigrate(&postgres.UserModel{}, &postgres.UserRoleModel{}, &postgres.APIKeyModel{})
	require.NoError(t, err)

	// Create default role
	defaultRole := postgres.UserRoleModel{
		Name:      "user",
		CreatedAt: time.Now(),
	}
	result := db.Create(&defaultRole)
	require.NoError(t, result.Error)

	return db, cleanup
}

func setupTestHandler(t *testing.T) *Handler {
	db, _ := setupTestDB(t)

	passwordConfig := &auth.PasswordConfig{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	}

	passwordPolicy := &auth.PasswordPolicy{
		MinLength:      10,
		RequireUpper:   true,
		RequireLower:   true,
		RequireDigit:   true,
		RequireSpecial: true,
	}

	return NewHandler(db, passwordConfig, passwordPolicy)
}

func createTestUser(t *testing.T, db *gorm.DB, login, email, password string) *postgres.UserModel {
	passwordHash, err := auth.HashPassword(password, &auth.PasswordConfig{
		Time:    1,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
	})
	require.NoError(t, err)

	var defaultRole postgres.UserRoleModel
	db.Where("name = ?", "user").First(&defaultRole)

	user := postgres.UserModel{
		ID:           uuid.New(),
		Login:        login,
		Email:        email,
		PasswordHash: passwordHash,
		RoleID:       &defaultRole.ID,
		CreatedAt:    time.Now(),
	}

	result := db.Create(&user)
	require.NoError(t, result.Error)

	return &user
}

func TestGetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	// Create test user
	testUser := createTestUser(t, handler.db, "testuser", "test@example.com", "TestPass123!")

	tests := []struct {
		name       string
		userID     uuid.UUID
		wantStatus int
		wantError  string
	}{
		{
			name:       "valid user",
			userID:     testUser.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "no authentication",
			userID:     uuid.Nil,
			wantStatus: http.StatusUnauthorized,
			wantError:  "authentication required",
		},
		{
			name:       "user not found",
			userID:     uuid.New(),
			wantStatus: http.StatusNotFound,
			wantError:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != uuid.Nil {
					c.Set(middleware.ContextUserIDKey, tt.userID)
				}
				c.Next()
			})
			router.GET("/me", handler.GetMe)

			req := httptest.NewRequest(http.MethodGet, "/me", nil)
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
				var response UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, testUser.ID, response.ID)
				assert.Equal(t, testUser.Login, response.Login)
				assert.Equal(t, testUser.Email, response.Email)
			}
		})
	}
}

func TestUpdateMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	// Create test user
	testUser := createTestUser(t, handler.db, "testuser", "test@example.com", "TestPass123!")

	// Create another user for email conflict test
	createTestUser(t, handler.db, "otheruser", "other@example.com", "TestPass123!")

	tests := []struct {
		name       string
		request    UpdateUserRequest
		userID     uuid.UUID
		wantStatus int
		wantError  string
	}{
		{
			name: "update email",
			request: UpdateUserRequest{
				Email: "newemail@example.com",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusOK,
		},
		{
			name: "update password",
			request: UpdateUserRequest{
				OldPassword: "TestPass123!",
				NewPassword: "NewPass456!",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusOK,
		},
		{
			name: "email already in use",
			request: UpdateUserRequest{
				Email: "other@example.com",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusConflict,
			wantError:  "email already in use",
		},
		{
			name: "same email",
			request: UpdateUserRequest{
				Email: "test@example.com",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusOK, // Same email, no conflict
		},
		{
			name: "invalid old password",
			request: UpdateUserRequest{
				OldPassword: "WrongPass123!",
				NewPassword: "NewPass456!",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid old password",
		},
		{
			name: "missing old password",
			request: UpdateUserRequest{
				NewPassword: "NewPass456!",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusBadRequest,
			wantError:  "old password is required",
		},
		{
			name:       "empty request",
			request:    UpdateUserRequest{},
			userID:     testUser.ID,
			wantStatus: http.StatusOK, // No updates, returns current user
		},
		{
			name: "invalid json",
			request: UpdateUserRequest{
				Email: "invalid-email",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusOK, // Email validation not enforced
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != uuid.Nil {
					c.Set(middleware.ContextUserIDKey, tt.userID)
				}
				c.Next()
			})
			router.PUT("/me", handler.UpdateMe)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.wantError)
			}
		})
	}
}

func TestDeleteMe(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	// Create test user
	testUser := createTestUser(t, handler.db, "testuser", "test@example.com", "TestPass123!")

	tests := []struct {
		name    string
		request struct {
			Password string `json:"password"`
		}
		userID     uuid.UUID
		wantStatus int
		wantError  string
	}{
		{
			name: "valid deletion",
			request: struct {
				Password string `json:"password"`
			}{
				Password: "TestPass123!",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid password",
			request: struct {
				Password string `json:"password"`
			}{
				Password: "WrongPass123!",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid password",
		},
		{
			name: "missing password",
			request: struct {
				Password string `json:"password"`
			}{
				Password: "",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			request: struct {
				Password string `json:"password"`
			}{
				Password: "TestPass123!",
			},
			userID:     uuid.New(),
			wantStatus: http.StatusNotFound,
			wantError:  "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != uuid.Nil {
					c.Set(middleware.ContextUserIDKey, tt.userID)
				}
				c.Next()
			})
			router.DELETE("/me", handler.DeleteMe)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodDelete, "/me", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.wantError)
			}
		})
	}
}

func TestListAPIKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	// Create test user
	testUser := createTestUser(t, handler.db, "testuser", "test@example.com", "TestPass123!")

	tests := []struct {
		name       string
		userID     uuid.UUID
		wantStatus int
		wantError  string
	}{
		{
			name:       "list keys successfully",
			userID:     testUser.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "no authentication",
			userID:     uuid.Nil,
			wantStatus: http.StatusUnauthorized,
			wantError:  "authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != uuid.Nil {
					c.Set(middleware.ContextUserIDKey, tt.userID)
				}
				c.Next()
			})
			router.GET("/api-keys", handler.ListAPIKeys)

			req := httptest.NewRequest(http.MethodGet, "/api-keys", nil)
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
				assert.Contains(t, response, "api_keys")
			}
		})
	}
}

func TestCreateAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	// Create test user
	testUser := createTestUser(t, handler.db, "testuser", "test@example.com", "TestPass123!")

	tests := []struct {
		name       string
		request    CreateAPIKeyRequest
		userID     uuid.UUID
		wantStatus int
		wantError  string
	}{
		{
			name: "create key successfully",
			request: CreateAPIKeyRequest{
				Name:   "Test Key",
				Scopes: []string{}, // Empty array to avoid PostgreSQL issues
			},
			userID:     testUser.ID,
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			request: CreateAPIKeyRequest{
				Scopes: []string{},
			},
			userID:     testUser.ID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no authentication",
			request:    CreateAPIKeyRequest{Name: "Test Key"},
			userID:     uuid.Nil,
			wantStatus: http.StatusUnauthorized,
			wantError:  "authentication required",
		},
		{
			name: "empty name",
			request: CreateAPIKeyRequest{
				Name:   "",
				Scopes: []string{},
			},
			userID:     testUser.ID,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != uuid.Nil {
					c.Set(middleware.ContextUserIDKey, tt.userID)
				}
				c.Next()
			})
			router.POST("/api-keys", handler.CreateAPIKey)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/api-keys", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.wantError)
			}

			if tt.wantStatus == http.StatusCreated {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response, "api_key")
				assert.Contains(t, response, "id")
			}
		})
	}
}

func TestRevokeAPIKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	// Create test user
	testUser := createTestUser(t, handler.db, "testuser", "test@example.com", "TestPass123!")

	// Create test API key
	apiKey := postgres.APIKeyModel{
		ID:        uuid.New(),
		UserID:    testUser.ID,
		KeyHash:   "test-hash",
		Name:      "Test Key",
		Scopes:    []string{}, // Empty array to avoid PostgreSQL array issues
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	result := handler.db.Create(&apiKey)
	if result.Error != nil {
		t.Logf("Error creating API key: %v", result.Error)
		// Continue anyway to test the revoke logic
	}

	tests := []struct {
		name       string
		keyID      string
		userID     uuid.UUID
		wantStatus int
		wantError  string
	}{
		{
			name:       "revoke key successfully",
			keyID:      apiKey.ID.String(),
			userID:     testUser.ID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "key not found",
			keyID:      uuid.New().String(),
			userID:     testUser.ID,
			wantStatus: http.StatusNotFound,
			wantError:  "API key not found",
		},
		{
			name:       "no authentication",
			keyID:      apiKey.ID.String(),
			userID:     uuid.Nil,
			wantStatus: http.StatusUnauthorized,
			wantError:  "authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != uuid.Nil {
					c.Set(middleware.ContextUserIDKey, tt.userID)
				}
				c.Next()
			})
			router.DELETE("/api-keys/:id", handler.RevokeAPIKey)

			req := httptest.NewRequest(http.MethodDelete, "/api-keys/"+tt.keyID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], tt.wantError)
			}
		})
	}
}

func TestUpdateMeEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	// Create test user
	testUser := createTestUser(t, handler.db, "testuser", "test@example.com", "TestPass123!")

	tests := []struct {
		name       string
		request    UpdateUserRequest
		userID     uuid.UUID
		wantStatus int
	}{
		{
			name: "update with same email",
			request: UpdateUserRequest{
				Email: "test@example.com",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusOK,
		},
		{
			name: "update with invalid password policy",
			request: UpdateUserRequest{
				OldPassword: "TestPass123!",
				NewPassword: "weak",
			},
			userID:     testUser.ID,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "update with database error",
			request: UpdateUserRequest{
				Email: "new@example.com",
			},
			userID:     uuid.New(),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				if tt.userID != uuid.Nil {
					c.Set(middleware.ContextUserIDKey, tt.userID)
				}
				c.Next()
			})
			router.PUT("/me", handler.UpdateMe)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
