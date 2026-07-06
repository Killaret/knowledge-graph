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

func TestDefaultSkipAuthConfig(t *testing.T) {
	config := DefaultSkipAuthConfig(true)

	assert.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Equal(t, uuid.MustParse("00000000-0000-0000-0000-000000000000"), config.DefaultUserID)
}

func TestSkipAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		enabled    bool
		wantUserID uuid.UUID
		wantRole   string
	}{
		{
			name:       "skip auth enabled",
			enabled:    true,
			wantUserID: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			wantRole:   "test",
		},
		{
			name:       "skip auth disabled",
			enabled:    false,
			wantUserID: uuid.Nil,
			wantRole:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSkipAuthConfig(tt.enabled)
			config.Enabled = tt.enabled

			router := gin.New()
			router.Use(SkipAuth(config))
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

			if tt.enabled {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUserID.String(), response["user_id"])
				assert.Equal(t, tt.wantRole, response["role"])
			}
		})
	}
}
