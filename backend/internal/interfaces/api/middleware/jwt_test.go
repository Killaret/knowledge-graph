package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
	}

	tests := []struct {
		name      string
		setupReq  func(*http.Request)
		wantToken string
	}{
		{
			name: "valid bearer token",
			setupReq: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer test-token")
			},
			wantToken: "test-token",
		},
		{
			name: "token without bearer prefix",
			setupReq: func(req *http.Request) {
				req.Header.Set("Authorization", "test-token")
			},
			wantToken: "test-token",
		},
		{
			name: "token from query parameter",
			setupReq: func(req *http.Request) {
				req.URL.RawQuery = "token=query-token"
			},
			wantToken: "query-token",
		},
		{
			name:      "no token",
			setupReq:  func(req *http.Request) {},
			wantToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			tt.setupReq(req)

			c := &gin.Context{Request: req}
			token, _ := extractToken(c, config)

			assert.Equal(t, tt.wantToken, token)
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
			wantStatus: http.StatusOK,
		},
		{
			name: "token from query parameter",
			path: "/api/v1/notes",
			setupToken: func(req *http.Request) {
				testID := uuid.New()
				token, _ := jwtManager.GenerateTokenPair(testID, "testuser", "user")
				req.URL.RawQuery = "token=" + token.AccessToken
			},
			wantStatus: http.StatusOK,
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
