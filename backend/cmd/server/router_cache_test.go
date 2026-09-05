package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCacheControlMiddlewareHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name      string
		maxAge    int
		wantCC    string
		wantVary  string
		wantPublic bool
	}{
		{
			name:       "cached endpoint uses private and Vary",
			maxAge:     60,
			wantCC:     "private, max-age=60",
			wantVary:   "Authorization, Cookie",
			wantPublic: false,
		},
		{
			name:       "uncached endpoint uses private and Vary",
			maxAge:     0,
			wantCC:     "private, max-age=0",
			wantVary:   "Authorization, Cookie",
			wantPublic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("GET", "/test", nil)

			handler := cacheControlMiddleware(tt.maxAge)
			handler(c)

			cc := w.Header().Get("Cache-Control")
			vary := w.Header().Get("Vary")

			assert.Equal(t, tt.wantCC, cc)
			assert.Equal(t, tt.wantVary, vary)
			assert.NotContains(t, cc, "public", "Cache-Control must not contain 'public'")
		})
	}
}
