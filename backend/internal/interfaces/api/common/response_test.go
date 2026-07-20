package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseHelpers(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(*gin.Context)
		wantStatus int
		wantBody   map[string]any
		emptyBody  bool
	}{
		{
			name: "JSON",
			fn:   func(c *gin.Context) { JSON(c, http.StatusOK, gin.H{"key": "value"}) },
			wantStatus: http.StatusOK,
			wantBody:   map[string]any{"data": map[string]any{"key": "value"}},
		},
		{
			name: "JSONWithMessage",
			fn:   func(c *gin.Context) { JSONWithMessage(c, http.StatusCreated, gin.H{"id": 1}, "created") },
			wantStatus: http.StatusCreated,
			wantBody:   map[string]any{"data": map[string]any{"id": float64(1)}, "message": "created"},
		},
		{
			name:      "NoContent",
			fn:        NoContent,
			wantStatus: http.StatusNoContent,
			emptyBody: true,
		},
		{
			name: "BadRequest",
			fn: func(c *gin.Context) {
				BadRequest(c, []FieldError{NewFieldError("name", ReasonRequired, "required")})
			},
			wantStatus: http.StatusBadRequest,
			wantBody: map[string]any{
				"code":    ErrCodeValidationError,
				"message": MsgValidationError,
				"details": []any{map[string]any{"field": "name", "reason": "required", "message": "required"}},
			},
		},
		{
			name:       "BadRequestSimple",
			fn:         func(c *gin.Context) { BadRequestSimple(c, "bad request") },
			wantStatus: http.StatusBadRequest,
			wantBody:   map[string]any{"code": ErrCodeInvalidRequest, "message": "bad request"},
		},
		{
			name:       "NotFound",
			fn:         func(c *gin.Context) { NotFound(c, "Note") },
			wantStatus: http.StatusNotFound,
			wantBody:   map[string]any{"code": ErrCodeNotFound, "message": "Note not found"},
		},
		{
			name: "Conflict",
			fn: func(c *gin.Context) {
				Conflict(c, []FieldError{NewFieldErrorWithValue("email", ReasonAlreadyExists, "exists", "a@b.c")})
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "ConflictSimple",
			fn:         func(c *gin.Context) { ConflictSimple(c, "duplicate") },
			wantStatus: http.StatusConflict,
			wantBody:   map[string]any{"code": ErrCodeConflict, "message": "duplicate"},
		},
		{
			name:       "InternalError",
			fn:         InternalError,
			wantStatus: http.StatusInternalServerError,
			wantBody:   map[string]any{"code": ErrCodeInternalError, "message": MsgInternalError},
		},
		{
			name:       "InternalErrorWithMessage",
			fn:         func(c *gin.Context) { InternalErrorWithMessage(c, "boom") },
			wantStatus: http.StatusInternalServerError,
			wantBody:   map[string]any{"code": ErrCodeInternalError, "message": "boom"},
		},
		{
			name:       "Forbidden",
			fn:         Forbidden,
			wantStatus: http.StatusForbidden,
			wantBody:   map[string]any{"code": ErrCodeForbidden, "message": MsgForbidden},
		},
		{
			name:       "Unauthorized",
			fn:         Unauthorized,
			wantStatus: http.StatusUnauthorized,
			wantBody:   map[string]any{"code": ErrCodeUnauthorized, "message": MsgUnauthorized},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			_, r := gin.CreateTestContext(w)
			r.GET("/test", tt.fn)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.emptyBody {
				assert.Empty(t, w.Body.String())
				return
			}

			var got map[string]any
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
			if tt.wantBody != nil {
				for key, want := range tt.wantBody {
					assert.Equal(t, want, got[key])
				}
			}
		})
	}
}

func TestFieldErrorConstructors(t *testing.T) {
	err1 := NewFieldError("f", "r", "m")
	assert.Equal(t, "f", err1.Field)
	assert.Equal(t, "r", err1.Reason)
	assert.Equal(t, "m", err1.Message)

	err2 := NewFieldErrorWithValue("f", "r", "m", "v")
	assert.Equal(t, "v", err2.Received)

	err3 := NewFieldErrorFull("f", "r", "m", "v", "e")
	assert.Equal(t, "e", err3.Expected)
}
