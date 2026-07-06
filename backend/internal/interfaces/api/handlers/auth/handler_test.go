package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, cleanup := testutil.SetupTestDB(t)

	// Auto-migrate models
	err := db.AutoMigrate(&postgres.UserModel{}, &postgres.UserRoleModel{})
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

	jwtManager := auth.NewJWTManager("test-secret-key", 24*time.Hour, 7*24*time.Hour)

	cfg := &config.Config{
		Argon2Time:                   1,
		Argon2Memory:                 64 * 1024,
		Argon2Threads:                4,
		PasswordPolicyMinLength:      10,
		PasswordPolicyRequireUpper:   true,
		PasswordPolicyRequireLower:   true,
		PasswordPolicyRequireDigit:   true,
		PasswordPolicyRequireSpecial: true,
	}

	return NewHandler(db, jwtManager, nil, cfg)
}

func TestRegister(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/register", handler.Register)

	tests := []struct {
		name       string
		request    RegisterRequest
		wantStatus int
		wantError  string
	}{
		{
			name: "valid registration",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid email",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "invalid-email",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "short",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "password policy violation - no uppercase",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "testpass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "password policy violation - no lowercase",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "TESTPASS123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "password policy violation - no digit",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "TestPass!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "password policy violation - no special",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "TestPass123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate login",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "test2@example.com",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusConflict,
			wantError:  "login already exists",
		},
		{
			name: "missing login",
			request: RegisterRequest{
				Login:    "",
				Email:    "test@example.com",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing email",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
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

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/login", handler.Login)

	// Register a user first
	registerReq := RegisterRequest{
		Login:    "testuser",
		Email:    "test@example.com",
		Password: "TestPass123!",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.POST("/register", handler.Register)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	tests := []struct {
		name       string
		request    LoginRequest
		wantStatus int
		wantError  string
	}{
		{
			name: "valid login",
			request: LoginRequest{
				Login:    "testuser",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			request: LoginRequest{
				Login:    "testuser",
				Password: "WrongPass123!",
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid credentials",
		},
		{
			name: "user not found",
			request: LoginRequest{
				Login:    "nonexistent",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid credentials",
		},
		{
			name: "missing login",
			request: LoginRequest{
				Login:    "",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			request: LoginRequest{
				Login:    "testuser",
				Password: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
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

			if tt.wantStatus == http.StatusOK {
				var response TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Equal(t, "testuser", response.User.Login)
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/refresh", handler.Refresh)

	// Register and login to get tokens
	registerReq := RegisterRequest{
		Login:    "testuser",
		Email:    "test@example.com",
		Password: "TestPass123!",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.POST("/register", handler.Register)
	router.ServeHTTP(w, req)

	loginReq := LoginRequest{
		Login:    "testuser",
		Password: "TestPass123!",
	}
	body, _ = json.Marshal(loginReq)
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.POST("/login", handler.Login)
	router.ServeHTTP(w, req)

	var loginResponse TokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)

	tests := []struct {
		name       string
		request    RefreshRequest
		wantStatus int
		wantError  string
	}{
		{
			name: "valid refresh",
			request: RefreshRequest{
				RefreshToken: loginResponse.RefreshToken,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid token",
			request: RefreshRequest{
				RefreshToken: "invalid-token",
			},
			wantStatus: http.StatusUnauthorized,
			wantError:  "invalid refresh token",
		},
		{
			name: "missing token",
			request: RefreshRequest{
				RefreshToken: "",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty json",
			request: RefreshRequest{
				RefreshToken: "",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(body))
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

			if tt.wantStatus == http.StatusOK {
				var response TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
			}
		})
	}
}

func TestRegisterEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/register", handler.Register)

	tests := []struct {
		name       string
		request    RegisterRequest
		wantStatus int
	}{
		{
			name: "empty login",
			request: RegisterRequest{
				Login:    "",
				Email:    "test@example.com",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty email",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty password",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "test@example.com",
				Password: "",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email format",
			request: RegisterRequest{
				Login:    "testuser",
				Email:    "invalid-email",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
