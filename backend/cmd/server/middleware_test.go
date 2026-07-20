package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCorsMiddleware_AllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("CORS_ALLOWED_ORIGINS", "http://example.com")

	middleware := corsMiddleware()
	r := gin.New()
	r.Use(middleware)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCorsMiddleware_Options(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := corsMiddleware()
	r := gin.New()
	r.Use(middleware)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCorsMiddleware_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := corsMiddleware()
	r := gin.New()
	r.Use(middleware)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestNewWriteLimiter_Disabled(t *testing.T) {
	cfg := &config.Config{ServerRateLimitEnabled: false}
	limiter := newWriteLimiter(cfg)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(limiter)
	r.GET("/test", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewWriteLimiter_Enabled(t *testing.T) {
	cfg := &config.Config{
		ServerRateLimitEnabled:       true,
		ServerRateLimitRequests:      100,
		ServerRateLimitWindowSeconds: 1,
		ServerRateLimitEndpoints: map[string]int{
			"notes_create":  10,
			"links_create":  10,
			"notes_update":  10,
		},
	}
	limiter := newWriteLimiter(cfg)
	assert.NotNil(t, limiter)
}
