package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(2, time.Hour)
	assert.True(t, rl.Allow("client-1"))
	assert.True(t, rl.Allow("client-1"))
	assert.False(t, rl.Allow("client-1"))
	assert.True(t, rl.Allow("client-2"))
}

func TestRateLimiter_Allow_WindowExpiration(t *testing.T) {
	rl := NewRateLimiter(1, 10*time.Millisecond)
	assert.True(t, rl.Allow("client"))
	assert.False(t, rl.Allow("client"))
	time.Sleep(15 * time.Millisecond)
	assert.True(t, rl.Allow("client"))
}

func TestRateLimiter_cleanupOldEntries(t *testing.T) {
	rl := NewRateLimiter(10, time.Hour)
	rl.requests["old-client"] = []time.Time{time.Now().Add(-20 * time.Hour)}
	rl.cleanupOldEntries()
	_, exists := rl.requests["old-client"]
	assert.False(t, exists)
}

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(RateLimitMiddleware(1, time.Hour))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest("GET", "/test", nil))
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest("GET", "/test", nil))
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimitByEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(RateLimitByEndpoint(map[string]int{"/test": 1}, 10, time.Hour))
	r.GET("/test", func(c *gin.Context) { c.Status(200) })
	r.GET("/other", func(c *gin.Context) { c.Status(200) })

	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest("GET", "/test", nil))
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest("GET", "/test", nil))
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, httptest.NewRequest("GET", "/other", nil))
	assert.Equal(t, http.StatusOK, w3.Code)
}
