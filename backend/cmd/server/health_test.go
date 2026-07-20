package main

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph/internal/config"
	nlp "knowledge-graph/internal/infrastructure/nlp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, *sql.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	// gorm.Open may call Ping during initialization
	mock.ExpectPing()

	dialector := postgres.New(postgres.Config{Conn: sqlDB})
	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	return db, sqlDB, mock, func() { sqlDB.Close() }
}

func TestHealthHandler_Healthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, sqlDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectPing()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	redisPinger := &redisPingAdapter{client: rdb}

	nlpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer nlpServer.Close()

	cfg := &config.Config{
		NLPServiceURL:          nlpServer.URL,
		RecommendationCacheTTL: time.Hour,
	}
	nlpClient := nlp.NewNLPClient(cfg.NLPServiceURL, rdb, cfg.RecommendationCacheTTL)

	handler := newHealthHandler(sqlDB, redisPinger, nlpClient)
	r := gin.New()
	r.GET("/health", handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestHealthHandler_DatabaseDown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, sqlDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectPing().WillReturnError(errors.New("db down"))

	cfg := &config.Config{}
	nlpClient := nlp.NewNLPClient(cfg.NLPServiceURL, nil, cfg.RecommendationCacheTTL)
	handler := newHealthHandler(sqlDB, nil, nlpClient)
	r := gin.New()
	r.GET("/health", handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "db down")
}

func TestHealthHandler_RedisDown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, sqlDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectPing()

	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	redisPinger := &redisPingAdapter{client: rdb}

	cfg := &config.Config{}
	nlpClient := nlp.NewNLPClient(cfg.NLPServiceURL, nil, cfg.RecommendationCacheTTL)
	handler := newHealthHandler(sqlDB, redisPinger, nlpClient)
	r := gin.New()
	r.GET("/health", handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHealthHandler_NLPUnhealthyButOptional(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, sqlDB, mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectPing()

	nlpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer nlpServer.Close()

	cfg := &config.Config{
		NLPServiceURL:          nlpServer.URL,
		RecommendationCacheTTL: time.Hour,
	}
	nlpClient := nlp.NewNLPClient(cfg.NLPServiceURL, nil, cfg.RecommendationCacheTTL)
	handler := newHealthHandler(sqlDB, nil, nlpClient)
	r := gin.New()
	r.GET("/health", handler)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "unhealthy")
}
