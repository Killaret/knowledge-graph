package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name    string
		handler gin.HandlerFunc
		want    int
	}{
		{
			name: "normal request passes through",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			},
			want: http.StatusOK,
		},
		{
			name: "panic is recovered",
			handler: func(c *gin.Context) {
				panic("test panic")
			},
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(RecoveryMiddleware())
			router.GET("/test", tt.handler)

			w := performRequest(router, "GET", "/test", "")
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

func TestRecoveryMiddlewareWithLogger(t *testing.T) {
	var logged string
	logger := func(format string, v ...interface{}) {
		logged = fmt.Sprintf(format, v...)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RecoveryMiddlewareWithLogger(logger))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := performRequest(router, "GET", "/panic", "")
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, logged, "PANIC RECOVERED")
}

func TestSafeHandler(t *testing.T) {
	tests := []struct {
		name    string
		handler gin.HandlerFunc
		want    int
	}{
		{
			name: "normal handler works",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			},
			want: http.StatusOK,
		},
		{
			name: "panic in handler is recovered",
			handler: func(c *gin.Context) {
				panic("handler panic")
			},
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.GET("/test", SafeHandler(tt.handler))

			w := performRequest(router, "GET", "/test", "")
			assert.Equal(t, tt.want, w.Code)
		})
	}
}

// Helper function to perform HTTP requests in tests
func performRequest(router *gin.Engine, method, path string, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	router.ServeHTTP(w, req)
	return w
}
