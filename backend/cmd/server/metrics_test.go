package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsHandler_ReturnsPoolStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/metrics/database", newMetricsHandler(nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics/database", nil)
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Contains(t, body, "timestamp")
	assert.Contains(t, body, "database_pool")
	pool, ok := body["database_pool"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, pool["error"], "nil")
}
