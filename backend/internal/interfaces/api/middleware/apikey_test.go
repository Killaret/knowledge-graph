package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDefaultAPIKeyConfig(t *testing.T) {
	config := &APIKeyConfig{
		HeaderName:   "X-API-Key",
		Enabled:      true,
		StaticAPIKey: "static-key",
		SkipPaths: []string{
			"/api/v1/auth/*",
			"/health",
		},
	}

	assert.NotNil(t, config)
	assert.Equal(t, "X-API-Key", config.HeaderName)
	assert.True(t, config.Enabled)
	assert.Equal(t, "static-key", config.StaticAPIKey)
	assert.NotEmpty(t, config.SkipPaths)
	assert.Contains(t, config.SkipPaths, "/api/v1/auth/*")
	assert.Contains(t, config.SkipPaths, "/health")
}

func TestHashAPIKey(t *testing.T) {
	key := "test-api-key"
	hash := hashAPIKey(key)

	assert.NotEmpty(t, hash)
	assert.NotEqual(t, key, hash)
	assert.Len(t, hash, 64) // SHA256 hex length
}

func TestGetAPIKeyID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupCtx   func(*gin.Context)
		wantExists bool
	}{
		{
			name: "api key id exists in context",
			setupCtx: func(c *gin.Context) {
				testID := uuid.New()
				c.Set("api_key_id", testID)
			},
			wantExists: true,
		},
		{
			name:       "api key id not in context",
			setupCtx:   func(c *gin.Context) {},
			wantExists: false,
		},
		{
			name: "invalid type in context",
			setupCtx: func(c *gin.Context) {
				c.Set("api_key_id", "invalid")
			},
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gin.Context{}
			tt.setupCtx(c)

			_, exists := GetAPIKeyID(c)
			assert.Equal(t, tt.wantExists, exists)
		})
	}
}

func TestAPIKeyDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled: false,
	}

	router := gin.New()
	router.Use(APIKey(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeySkipPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled: true,
		SkipPaths: []string{
			"/api/v1/auth/login",
			"/health",
		},
	}

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "skip auth path",
			path:       "/api/v1/auth/login",
			wantStatus: http.StatusOK,
		},
		{
			name:       "skip health path",
			path:       "/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "protected path without key",
			path:       "/api/v1/notes",
			wantStatus: http.StatusOK, // Continues to next auth method
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(APIKey(config))
			router.GET(tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAPIKeyStaticKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled:      true,
		StaticAPIKey: "test-static-key",
		SkipPaths:    []string{},
	}

	router := gin.New()
	router.Use(APIKey(config))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		role, _ := GetUserRole(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"role":    role,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "test-static-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyJWTAlreadyAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled:   true,
		SkipPaths: []string{},
	}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Simulate JWT authentication
		testID := uuid.New()
		c.Set(ContextUserIDKey, testID)
		c.Set(ContextRoleKey, "user")
		c.Next()
	})
	router.Use(APIKey(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyNoKeyProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled:   true,
		SkipPaths: []string{},
	}

	router := gin.New()
	router.Use(APIKey(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // Continues to next auth method
}

func TestAPIKeyModelTableName(t *testing.T) {
	model := APIKeyModel{}
	assert.Equal(t, "api_keys", model.TableName())
}

func TestAPIKeyInvalidStaticKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled:      true,
		StaticAPIKey: "test-static-key",
		SkipPaths:    []string{},
	}

	router := gin.New()
	router.Use(APIKey(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "wrong-static-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // Continues to next auth method
}

func TestAPIKeyCustomHeaderName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled:      true,
		StaticAPIKey: "test-static-key",
		HeaderName:   "X-Custom-API-Key",
		SkipPaths:    []string{},
	}

	router := gin.New()
	router.Use(APIKey(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Custom-API-Key", "test-static-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyEmptyStaticKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled:      true,
		StaticAPIKey: "",
		SkipPaths:    []string{},
	}

	router := gin.New()
	router.Use(APIKey(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // Continues to next auth method
}

func TestAPIKeySkipPathWildcard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled:   true,
		SkipPaths: []string{"/api/v1/auth/*"},
	}

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "skip wildcard path",
			path:       "/api/v1/auth/login",
			wantStatus: http.StatusOK,
		},
		{
			name:       "skip wildcard path deep",
			path:       "/api/v1/auth/register",
			wantStatus: http.StatusOK,
		},
		{
			name:       "protected path",
			path:       "/api/v1/notes",
			wantStatus: http.StatusOK, // Continues to next auth method
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(APIKey(config))
			router.GET(tt.path, func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAPIKeyStaticKeyWithWhitespace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled:      true,
		StaticAPIKey: "test-key",
		SkipPaths:    []string{},
	}

	router := gin.New()
	router.Use(APIKey(config))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		role, _ := GetUserRole(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"role":    role,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", " test-key ") // With whitespace
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyHashConsistency(t *testing.T) {
	key := "test-api-key"
	hash1 := hashAPIKey(key)
	hash2 := hashAPIKey(key)

	assert.Equal(t, hash1, hash2)
	assert.Len(t, hash1, 64) // SHA256 hex length
}

func TestAPIKeyDifferentKeysDifferentHashes(t *testing.T) {
	key1 := "test-api-key-1"
	key2 := "test-api-key-2"

	hash1 := hashAPIKey(key1)
	hash2 := hashAPIKey(key2)

	assert.NotEqual(t, hash1, hash2)
}

func TestAPIKeyEmptyKeyHash(t *testing.T) {
	key := ""
	hash := hashAPIKey(key)

	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
}

func TestAPIKeySpecialCharacters(t *testing.T) {
	key := "test-@#$%^&*()_+-=[]{}|;':,.<>?/~`"
	hash := hashAPIKey(key)

	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)
}

func TestAPIKeyStaticKeySetsAdminRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := &APIKeyConfig{
		Enabled:      true,
		StaticAPIKey: "test-key",
		HeaderName:   "X-API-Key",
	}

	router := gin.New()
	router.Use(APIKey(config))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		role, _ := GetUserRole(c)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"role":    role,
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "test-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "admin", response["role"])
}
