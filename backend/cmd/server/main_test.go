package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knowledge-graph/internal/config"
)

func TestRun_NilConfig(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	_, _, err := run(context.Background(), nil, db, nil, nil, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config is nil")
}

func TestRun_NilDatabase(t *testing.T) {
	cfg := &config.Config{ServerPort: "9999"}
	_, _, err := run(context.Background(), cfg, nil, nil, nil, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database is nil")
}

func TestRun_Success(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	cfg := &config.Config{
		ServerPort:                     "9999",
		ServerRateLimitEnabled:         false,
		SkipAuth:                       true,
		APIKeyEnabled:                  false,
		RecommendationAlpha:            1.0,
		RecommendationBeta:             0.0,
		RecommendationDepth:            2,
		RecommendationDecay:            0.5,
		RecommendationCacheTTL:         time.Minute,
		EmbeddingSimilarityLimit:       10,
		GraphLoadDepth:                 2,
		BFSAggregation:                 "max",
		BFSNormalize:                   false,
		RecommendationTaskDelaySeconds: 0,
	}

	srv, cleanupRun, err := run(context.Background(), cfg, db, nil, nil, nil, "")
	require.NoError(t, err)
	require.NotNil(t, srv)
	assert.Equal(t, ":9999", srv.Addr)
	assert.NotNil(t, srv.Handler)

	if cleanupRun != nil {
		cleanupRun()
	}
}

func TestConnectDatabaseWithRetry(t *testing.T) {
	cfg := &config.Config{DatabaseURL: "invalid://", DatabaseRetryDelaySeconds: 0}
	_, err := connectDatabaseWithRetry(context.Background(), cfg)
	assert.Error(t, err)
}

func TestNewRedisClient_EmptyURL(t *testing.T) {
	cfg := &config.Config{RedisURL: ""}
	client := newRedisClient(cfg)
	assert.Nil(t, client)
}

func TestNewMongoClient_EmptyURL(t *testing.T) {
	cfg := &config.Config{MongoDBURL: ""}
	client := newMongoClient(context.Background(), cfg)
	assert.Nil(t, client)
}

func TestNewAsynqClient_EmptyURL(t *testing.T) {
	cfg := &config.Config{RedisURL: ""}
	queue := newAsynqClient(cfg)
	assert.Nil(t, queue)
}
