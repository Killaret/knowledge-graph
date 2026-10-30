//go:build integration
// +build integration

package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/email"
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
		JWTAccessTTL:                 24 * time.Hour,
		JWTRefreshTTL:                7 * 24 * time.Hour,
	}

	userRepo := postgres.NewUserRepository(db, nil)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	return NewHandler(userRepo, refreshTokenRepo, nil, jwtManager, cfg, email.NewConsole(), nil)
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
				Password: "TestPass!!",
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
				Email:    "test@example.com",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "missing login",
			request: RegisterRequest{
				Email:    "test@example.com",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing email",
			request: RegisterRequest{
				Login:    "testuser",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			request: RegisterRequest{
				Login: "testuser",
				Email: "test@example.com",
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

			if tt.wantStatus == http.StatusCreated {
				var response TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Equal(t, tt.request.Login, response.User.Login)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/register", handler.Register)
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
	router.ServeHTTP(w, req)

	tests := []struct {
		name       string
		request    LoginRequest
		wantStatus int
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
		},
		{
			name: "user not found",
			request: LoginRequest{
				Login:    "nonexistent",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing login",
			request: LoginRequest{
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			request: LoginRequest{
				Login: "testuser",
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

			if tt.wantStatus == http.StatusOK {
				var response TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.AccessToken)
				assert.NotEmpty(t, response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)
	router.POST("/refresh", handler.Refresh)

	// Register and login to get refresh token
	registerReq := RegisterRequest{
		Login:    "testuser",
		Email:    "test@example.com",
		Password: "TestPass123!",
	}
	body, _ := json.Marshal(registerReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginReq := LoginRequest{
		Login:    "testuser",
		Password: "TestPass123!",
	}
	body, _ = json.Marshal(loginReq)
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResponse TokenResponse
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

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

func TestLoginEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	tests := []struct {
		name       string
		request    LoginRequest
		wantStatus int
	}{
		{
			name: "empty login",
			request: LoginRequest{
				Login:    "",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "empty password",
			request: LoginRequest{
				Login:    "testuser",
				Password: "",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "login too short",
			request: LoginRequest{
				Login:    "ab",
				Password: "TestPass123!",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "login too long",
			request: LoginRequest{
				Login:    strings.Repeat("a", 51),
				Password: "TestPass123!",
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/login", handler.Login)

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestLogout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "logout without token",
			token:      "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "logout with token",
			token:      "test-token",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/logout", handler.Logout)

			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestForgotPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)

	tests := []struct {
		name       string
		email      string
		wantStatus int
	}{
		{
			name:       "valid email",
			email:      "test@example.com",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid email",
			email:      "invalid-email",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty email",
			email:      "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.POST("/forgot-password", handler.ForgotPassword)

			body := map[string]string{"email": tt.email}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestRegisterDuplicateLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/register", handler.Register)

	// Register first user
	body1 := map[string]string{
		"login":    "testuser",
		"email":    "test1@example.com",
		"password": "TestPass123!",
	}
	jsonBody1, _ := json.Marshal(body1)
	req1 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	// Try to register with same login
	body2 := map[string]string{
		"login":    "testuser",
		"email":    "test2@example.com",
		"password": "TestPass123!",
	}
	jsonBody2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestRegisterDuplicateEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/register", handler.Register)

	// Register first user
	body1 := map[string]string{
		"login":    "testuser1",
		"email":    "test@example.com",
		"password": "TestPass123!",
	}
	jsonBody1, _ := json.Marshal(body1)
	req1 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	// Try to register with same email
	body2 := map[string]string{
		"login":    "testuser2",
		"email":    "test@example.com",
		"password": "TestPass123!",
	}
	jsonBody2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestRegisterPasswordPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/register", handler.Register)

	tests := []struct {
		name       string
		password   string
		wantStatus int
	}{
		{
			name:       "too short",
			password:   "Short1!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no uppercase",
			password:   "testpass123!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no lowercase",
			password:   "TESTPASS123!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no digit",
			password:   "TestPass!!",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no special",
			password:   "TestPass123",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{
				"login":    "testuser" + tt.name,
				"email":    tt.name + "@example.com",
				"password": tt.password,
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestLoginWrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)

	// Register a user
	body1 := map[string]string{
		"login":    "testuser",
		"email":    "test@example.com",
		"password": "TestPass123!",
	}
	jsonBody1, _ := json.Marshal(body1)
	req1 := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	// Try to login with wrong password
	body2 := map[string]string{
		"login":    "testuser",
		"password": "WrongPass123!",
	}
	jsonBody2, _ := json.Marshal(body2)
	req2 := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

func TestLoginNonexistentUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/login", handler.Login)

	body := map[string]string{
		"login":    "nonexistent",
		"password": "TestPass123!",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefreshInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/refresh", handler.Refresh)

	body := map[string]string{
		"refresh_token": "invalid-token",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefreshEmptyToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/refresh", handler.Refresh)

	body := map[string]string{
		"refresh_token": "",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForgotPasswordNonexistentEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/forgot-password", handler.ForgotPassword)

	body := map[string]string{
		"email": "nonexistent@example.com",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return success even for nonexistent email (security best practice)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestForgotPasswordEmptyEmail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.POST("/forgot-password", handler.ForgotPassword)

	body := map[string]string{
		"email": "",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestYandexLoginNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler(t)
	router := gin.New()
	router.GET("/auth/yandex/login", handler.YandexLogin)

	req := httptest.NewRequest(http.MethodGet, "/auth/yandex/login", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return not implemented since Yandex is not configured
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}
