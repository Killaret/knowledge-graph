package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDefaultJWTConfig(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := DefaultJWTConfig(jwtManager, nil)

	assert.NotNil(t, config)
	assert.Equal(t, jwtManager, config.JWTManager)
	assert.Nil(t, config.TokenStore)
	assert.Equal(t, "Authorization", config.HeaderName)
	assert.Equal(t, "header", config.TokenLookup)
	assert.NotEmpty(t, config.SkipPaths)
	assert.Contains(t, config.SkipPaths, "/api/v1/auth/login")
	assert.Contains(t, config.SkipPaths, "/health")
}

func TestExtractToken(t *testing.T) {
	config := &JWTConfig{
		HeaderName:  "Authorization",
		TokenLookup: "header",
		CookieName:  "access_token",
	}

	tests := []struct {
		name      string
		setupReq  func(*http.Request)
		wantToken string
		wantErr   bool
	}{
		{
			name: "valid bearer token",
			setupReq: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer test-token")
			},
			wantToken: "test-token",
			wantErr:   false,
		},
		{
			name: "token without bearer prefix",
			setupReq: func(req *http.Request) {
				req.Header.Set("Authorization", "test-token")
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "empty bearer token",
			setupReq: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer ")
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "token from query parameter",
			setupReq: func(req *http.Request) {
				req.URL.RawQuery = "token=query-token"
			},
			wantToken: "",
			wantErr:   true,
		},
		{
			name: "token from cookie",
			setupReq: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
			},
			wantToken: "cookie-token",
			wantErr:   false,
		},
		{
			name:      "no token",
			setupReq:  func(req *http.Request) {},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tt.setupReq(req)

			c := &gin.Context{Request: req}
			token, err := extractToken(c, config)

			assert.Equal(t, tt.wantToken, token)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(*gin.Context)
		wantExists bool
	}{
		{
			name: "user id exists in context",
			setupCtx: func(c *gin.Context) {
				testID := uuid.New()
				c.Set(ContextUserIDKey, testID)
			},
			wantExists: true,
		},
		{
			name:       "user id not in context",
			setupCtx:   func(c *gin.Context) {},
			wantExists: false,
		},
		{
			name: "invalid type in context",
			setupCtx: func(c *gin.Context) {
				c.Set(ContextUserIDKey, "invalid")
			},
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gin.Context{}
			tt.setupCtx(c)

			_, exists := GetUserID(c)
			assert.Equal(t, tt.wantExists, exists)
		})
	}
}

func TestGetUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(*gin.Context)
		wantRole   string
		wantExists bool
	}{
		{
			name: "role exists in context",
			setupCtx: func(c *gin.Context) {
				c.Set(ContextRoleKey, "admin")
			},
			wantRole:   "admin",
			wantExists: true,
		},
		{
			name:       "role not in context",
			setupCtx:   func(c *gin.Context) {},
			wantRole:   "",
			wantExists: false,
		},
		{
			name: "invalid type in context",
			setupCtx: func(c *gin.Context) {
				c.Set(ContextRoleKey, 123)
			},
			wantRole:   "",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gin.Context{}
			tt.setupCtx(c)

			role, exists := GetUserRole(c)

			assert.Equal(t, tt.wantExists, exists)
			assert.Equal(t, tt.wantRole, role)
		})
	}
}

func TestRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(*gin.Context)
		wantStatus int
	}{
		{
			name: "authenticated user",
			setupCtx: func(c *gin.Context) {
				c.Set(ContextUserIDKey, uuid.New())
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthenticated user",
			setupCtx:   func(c *gin.Context) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				tt.setupCtx(c)
				c.Next()
			})
			router.Use(RequireAuth())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestJWTAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := DefaultJWTConfig(jwtManager, nil)

	tests := []struct {
		name       string
		path       string
		setupToken func(*http.Request)
		wantStatus int
	}{
		{
			name: "valid token",
			path: "/api/v1/notes",
			setupToken: func(req *http.Request) {
				testID := uuid.New()
				token, _ := jwtManager.GenerateTokenPair(testID, "testuser", "user")
				req.Header.Set("Authorization", "Bearer "+token.AccessToken)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "skip path - login",
			path:       "/api/v1/auth/login",
			setupToken: func(req *http.Request) {},
			wantStatus: http.StatusOK,
		},
		{
			name:       "skip path - health",
			path:       "/health",
			setupToken: func(req *http.Request) {},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing token",
			path:       "/api/v1/notes",
			setupToken: func(req *http.Request) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			path: "/api/v1/notes",
			setupToken: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid-token")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "token without bearer prefix",
			path: "/api/v1/notes",
			setupToken: func(req *http.Request) {
				testID := uuid.New()
				token, _ := jwtManager.GenerateTokenPair(testID, "testuser", "user")
				req.Header.Set("Authorization", token.AccessToken)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "token from query parameter",
			path: "/api/v1/notes",
			setupToken: func(req *http.Request) {
				testID := uuid.New()
				token, _ := jwtManager.GenerateTokenPair(testID, "testuser", "user")
				req.URL.RawQuery = "token=" + token.AccessToken
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(JWTAuth(config))
			router.GET(tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			tt.setupToken(req)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestJWTAuthSwaggerPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := DefaultJWTConfig(jwtManager, nil)

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/swagger/index.html", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "swagger"})
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestContextKeys(t *testing.T) {
	assert.Equal(t, "user_id", ContextUserIDKey)
	assert.Equal(t, "user_role", ContextRoleKey)
	assert.Equal(t, "user_login", ContextLoginKey)
}

func TestJWTAuthEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := DefaultJWTConfig(jwtManager, nil)

	tests := []struct {
		name       string
		path       string
		setupToken func(*http.Request)
		wantStatus int
	}{
		{
			name: "malformed token",
			path: "/api/v1/notes",
			setupToken: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer malformed.token.here")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "wrong secret",
			path: "/api/v1/notes",
			setupToken: func(req *http.Request) {
				wrongJWTManager := auth.NewJWTManager("wrong-secret", 24*3600, 7*24*3600)
				testID := uuid.New()
				token, _ := wrongJWTManager.GenerateTokenPair(testID, "testuser", "user")
				req.Header.Set("Authorization", "Bearer "+token.AccessToken)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "empty token",
			path: "/api/v1/notes",
			setupToken: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer ")
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(JWTAuth(config))
			router.GET(tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			tt.setupToken(req)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestJWTAuthWithTokenStore(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := DefaultJWTConfig(jwtManager, nil)

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	testID := uuid.New()
	token, _ := jwtManager.GenerateTokenPair(testID, "testuser", "user")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should pass since token is not blacklisted
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthAlreadyAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := DefaultJWTConfig(jwtManager, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate SkipAuth setting user in context
		testID := uuid.New()
		c.Set(ContextUserIDKey, testID)
		c.Set(ContextRoleKey, "admin")
		c.Next()
	})
	router.Use(JWTAuth(config))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		role, _ := GetUserRole(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"role":    role,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestJWTAuthCustomSkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := &JWTConfig{
		JWTManager:  jwtManager,
		TokenStore:  nil,
		SkipPaths:   []string{"/custom/skip"},
		HeaderName:  "Authorization",
		TokenLookup: "header",
	}

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/custom/skip", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "skipped"})
	})
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected"})
	})

	// Test skip path
	req1 := httptest.NewRequest(http.MethodGet, "/custom/skip", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Test protected path without token
	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusUnauthorized, w2.Code)
}

func TestJWTAuthCustomHeaderName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := &JWTConfig{
		JWTManager:  jwtManager,
		TokenStore:  nil,
		SkipPaths:   []string{},
		HeaderName:  "X-Custom-Auth",
		TokenLookup: "header",
	}

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	testID := uuid.New()
	token, _ := jwtManager.GenerateTokenPair(testID, "testuser", "user")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom-Auth", "Bearer "+token.AccessToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(*gin.Context)
		wantLogin  string
		wantExists bool
	}{
		{
			name: "login exists in context",
			setupCtx: func(c *gin.Context) {
				c.Set(ContextLoginKey, "testuser")
			},
			wantLogin:  "testuser",
			wantExists: true,
		},
		{
			name:       "login not in context",
			setupCtx:   func(c *gin.Context) {},
			wantLogin:  "",
			wantExists: false,
		},
		{
			name: "invalid type in context",
			setupCtx: func(c *gin.Context) {
				c.Set(ContextLoginKey, 123)
			},
			wantLogin:  "",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gin.Context{}
			tt.setupCtx(c)

			login, exists := GetLogin(c)

			assert.Equal(t, tt.wantExists, exists)
			assert.Equal(t, tt.wantLogin, login)
		})
	}
}

func TestExtractTokenBearerCaseInsensitive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &JWTConfig{
		HeaderName:  "Authorization",
		TokenLookup: "header",
	}

	tests := []struct {
		name       string
		authHeader string
		wantToken  string
		wantErr    bool
	}{
		{
			name:       "Bearer lowercase",
			authHeader: "bearer token123",
			wantToken:  "token123",
			wantErr:    false,
		},
		{
			name:       "Bearer uppercase",
			authHeader: "BEARER token123",
			wantToken:  "token123",
			wantErr:    false,
		},
		{
			name:       "Bearer mixed case",
			authHeader: "BeArEr token123",
			wantToken:  "token123",
			wantErr:    false,
		},
		{
			name:       "no Bearer prefix",
			authHeader: "token123",
			wantToken:  "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tt.authHeader)

			c := &gin.Context{Request: req}
			token, err := extractToken(c, config)

			assert.Equal(t, tt.wantToken, token)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestJWTAuthTokenLookupQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := &JWTConfig{
		JWTManager:  jwtManager,
		TokenStore:  nil,
		SkipPaths:   []string{},
		HeaderName:  "Authorization",
		TokenLookup: "query",
	}

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	testID := uuid.New()
	token, _ := jwtManager.GenerateTokenPair(testID, "testuser", "user")
	req := httptest.NewRequest(http.MethodGet, "/test?token="+token.AccessToken, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Query token transport has been removed.
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthTokenLookupMixed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := &JWTConfig{
		JWTManager:  jwtManager,
		TokenStore:  nil,
		SkipPaths:   []string{},
		HeaderName:  "Authorization",
		TokenLookup: "header",
	}

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	testID := uuid.New()
	token, _ := jwtManager.GenerateTokenPair(testID, "testuser", "user")
	// Try query parameter as fallback
	req := httptest.NewRequest(http.MethodGet, "/test?token="+token.AccessToken, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Query token transport has been removed.
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := &JWTConfig{
		JWTManager:  jwtManager,
		TokenStore:  nil,
		SkipPaths:   []string{},
		HeaderName:  "Authorization",
		TokenLookup: "header",
	}

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthMalformedToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := &JWTConfig{
		JWTManager:  jwtManager,
		TokenStore:  nil,
		SkipPaths:   []string{},
		HeaderName:  "Authorization",
		TokenLookup: "header",
	}

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer malformed-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthWrongSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager1 := auth.NewJWTManager("secret1", 24*3600, 7*24*3600)
	jwtManager2 := auth.NewJWTManager("secret2", 24*3600, 7*24*3600)
	config := &JWTConfig{
		JWTManager:  jwtManager2, // Different secret
		TokenStore:  nil,
		SkipPaths:   []string{},
		HeaderName:  "Authorization",
		TokenLookup: "header",
	}

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	testID := uuid.New()
	token, _ := jwtManager1.GenerateTokenPair(testID, "testuser", "user")

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuthSwaggerWildcard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := &JWTConfig{
		JWTManager:  jwtManager,
		TokenStore:  nil,
		SkipPaths:   []string{"/swagger/*any"},
		HeaderName:  "Authorization",
		TokenLookup: "header",
	}

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/swagger/index.html", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "swagger"})
	})

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// panicTokenStore fails any call to IsTokenBlacklisted, proving that a missing
// token does not reach the store.
type panicTokenStore struct{}

func (p *panicTokenStore) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	return nil
}
func (p *panicTokenStore) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	panic("IsTokenBlacklisted should not be called when token is missing")
}
func (p *panicTokenStore) StoreRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error {
	return nil
}
func (p *panicTokenStore) ValidateRefreshToken(ctx context.Context, token string) (string, error) {
	return "", nil
}
func (p *panicTokenStore) RevokeRefreshToken(ctx context.Context, token string, ttl time.Duration) error {
	return nil
}
func (p *panicTokenStore) StorePasswordResetToken(ctx context.Context, userID string, token string, ttl time.Duration) error {
	return nil
}
func (p *panicTokenStore) ValidatePasswordResetToken(ctx context.Context, token string) (string, error) {
	return "", nil
}
func (p *panicTokenStore) DeletePasswordResetToken(ctx context.Context, token string) error {
	return nil
}
func (p *panicTokenStore) StorePKCE(ctx context.Context, state string, pkce *auth.PKCE, ttl time.Duration) error {
	return nil
}
func (p *panicTokenStore) GetPKCE(ctx context.Context, state string) (*auth.PKCE, error) {
	return nil, nil
}
func (p *panicTokenStore) StoreState(ctx context.Context, state string, ttl time.Duration) error {
	return nil
}
func (p *panicTokenStore) GetState(ctx context.Context, state string) (string, error) { return "", nil }
func (p *panicTokenStore) CachePermission(ctx context.Context, userID, resource, action string, allowed bool, ttl time.Duration) error {
	return nil
}
func (p *panicTokenStore) CheckCachedPermission(ctx context.Context, userID, resource, action string) (bool, bool, error) {
	return false, false, nil
}
func (p *panicTokenStore) InvalidatePermissionCache(ctx context.Context, userID string) error {
	return nil
}

func TestJWTAuthNoTokenSkipsBlacklist(t *testing.T) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", 24*3600, 7*24*3600)
	config := &JWTConfig{
		JWTManager:  jwtManager,
		TokenStore:  &panicTokenStore{},
		SkipPaths:   []string{},
		HeaderName:  "Authorization",
		TokenLookup: "header",
	}

	router := gin.New()
	router.Use(JWTAuth(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
