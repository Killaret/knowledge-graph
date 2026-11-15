package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/mongo"
	"knowledge-graph/internal/interfaces/api/notehandler"
)

// NOTE-QUALITY-1 criterion 6: assembling the server with
// nlp.quality.enabled=false leaves the endpoints answering disabled and
// never reaches Mongo or the task queue.
func TestInstallQualityEndpoints_FlagOff(t *testing.T) {
	h := notehandler.New(nil, nil, nil, nil, 0, nil, nil, nil, &config.Config{}, nil, nil, nil)
	require.False(t, installQualityEndpoints(context.Background(), h, false, &mongo.Client{}, nil))

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/notes/:id/quality", h.GetQuality)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/notes/"+uuid.NewString()+"/quality", nil))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.JSONEq(t, `{"enabled": false}`, rec.Body.String())
}

func TestInstallQualityEndpoints_NoMongo(t *testing.T) {
	h := notehandler.New(nil, nil, nil, nil, 0, nil, nil, nil, &config.Config{}, nil, nil, nil)
	assert.False(t, installQualityEndpoints(context.Background(), h, true, nil, nil))
}
