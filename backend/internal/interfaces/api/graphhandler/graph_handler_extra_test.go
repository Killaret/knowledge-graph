package graphhandler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/application/cache"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCalculateDelta(t *testing.T) {
	cached := GraphData{
		Nodes: []GraphNode{{ID: "1", Title: "A", Type: "star"}, {ID: "2", Title: "B", Type: "planet"}},
		Links: []GraphLink{{Source: "1", Target: "2", Weight: 0.5, LinkType: "reference"}},
	}
	fresh := GraphData{
		Nodes: []GraphNode{{ID: "1", Title: "A", Type: "star"}, {ID: "3", Title: "C", Type: "moon"}},
		Links: []GraphLink{{Source: "1", Target: "3", Weight: 0.8, LinkType: "reference"}},
	}

	delta := calculateDelta(cached, fresh)
	require.NotNil(t, delta)
	assert.Len(t, delta.AddedNodes, 1)
	assert.Equal(t, "3", delta.AddedNodes[0].ID)
	assert.Len(t, delta.RemovedNodes, 1)
	assert.Equal(t, "2", delta.RemovedNodes[0])
	assert.Len(t, delta.AddedLinks, 1)
	assert.Len(t, delta.RemovedLinks, 1)

	same := calculateDelta(cached, cached)
	assert.Nil(t, same)
}

func TestConvertCacheGraphData(t *testing.T) {
	handlerData := GraphData{
		Nodes: []GraphNode{{ID: "1", Title: "A", Type: "star", X: 1.0, Y: 2.0}},
		Links: []GraphLink{{Source: "1", Target: "2", Weight: 0.5, LinkType: "reference"}},
	}

	cacheData := convertToCacheGraphData(handlerData)
	assert.Len(t, cacheData.Nodes, 1)
	assert.Equal(t, 1.0, cacheData.Nodes[0].X)

	roundTrip := convertFromCacheGraphData(cacheData)
	assert.Len(t, roundTrip.Nodes, 1)
	assert.Equal(t, "A", roundTrip.Nodes[0].Title)
}

func TestPreserveCachedPositions(t *testing.T) {
	handler := New(nil, nil, &config.Config{}, nil)
	fresh := GraphData{
		Nodes: []GraphNode{{ID: "1", Title: "A", Type: "star", X: 0, Y: 0}},
		Links: []GraphLink{},
	}
	cached := GraphData{
		Nodes: []GraphNode{{ID: "1", Title: "A", Type: "star", X: 10.0, Y: 20.0}},
		Links: []GraphLink{},
	}

	preserved := handler.preserveCachedPositions(fresh, cached)
	assert.Equal(t, 10.0, preserved.Nodes[0].X)
	assert.Equal(t, 20.0, preserved.Nodes[0].Y)
}

func setupGraphRouterWithCache() (*gin.Engine, *cache.GraphCache, *mockNoteRepo, *mockLinkRepo) {
	gin.SetMode(gin.TestMode)
	graphCache := cache.NewGraphCache(cachetest.NewFakeCacheClient())

	noteRepo := new(mockNoteRepo)
	linkRepo := new(mockLinkRepo)
	cfg := &config.Config{
		GraphLoadDepth:        3,
		GraphDefaultLimit:     100,
		GraphMaxLimit:         1000,
		GraphLinkDefaultLimit: 500,
		GraphLinkMaxLimit:     5000,
	}
	handler := New(noteRepo, linkRepo, cfg, graphCache)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
		c.Next()
	})
	r.GET("/graph/cached", handler.GetCachedGraph)
	r.GET("/graph/fresh", handler.GetFreshGraph)
	return r, graphCache, noteRepo, linkRepo
}

func TestGetCachedGraph_NotFound(t *testing.T) {
	r, _, _, _ := setupGraphRouterWithCache()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/graph/cached", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGetCachedGraph_Found(t *testing.T) {
	r, graphCache, _, _ := setupGraphRouterWithCache()

	data := cache.GraphData{
		Nodes: []cache.GraphNode{{ID: "1", Title: "A", Type: "star"}},
		Links: []cache.GraphLink{},
	}
	err := graphCache.CacheUserGraph(context.Background(), "00000000-0000-0000-0000-000000000001", data)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/graph/cached", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	dataObj := resp["data"].(map[string]interface{})
	assert.Len(t, dataObj["nodes"], 1)
}

func TestGetFreshGraph(t *testing.T) {
	r, _, noteRepo, linkRepo := setupGraphRouterWithCache()

	noteRepo.On("FindAllPaginated", mock.Anything, 100, 0).Return([]*note.Note{}, int64(0), nil)
	linkRepo.On("FindAllPaginated", mock.Anything, 500, 0).Return([]*link.Link{}, int64(0), nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/graph/fresh", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetCachedGraph_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := New(nil, nil, &config.Config{}, cache.NewGraphCache(nil))
	r := gin.Default()
	r.GET("/graph/cached", handler.GetCachedGraph)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/graph/cached", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCachedGraph_NoCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := New(nil, nil, &config.Config{}, nil)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
		c.Next()
	})
	r.GET("/graph/cached", handler.GetCachedGraph)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/graph/cached", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
