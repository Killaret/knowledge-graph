package cache

import (
	"context"
	"testing"
	"time"

	"knowledge-graph/internal/domain/cache/cachetest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCacheClient(t *testing.T) *cachetest.FakeCacheClient {
	return cachetest.NewFakeCacheClient()
}

func TestNewGraphCache(t *testing.T) {
	cache := NewGraphCache(newTestCacheClient(t))
	assert.NotNil(t, cache)
	assert.Equal(t, 5*time.Minute, cache.ttl)
}

func TestCacheUserGraph(t *testing.T) {
	ctx := context.Background()
	cacheClient := newTestCacheClient(t)
	graphCache := NewGraphCache(cacheClient)

	userID := "test-user-123"
	testData := GraphData{
		Nodes: []GraphNode{
			{ID: "node1", Title: "Test Node 1", Type: "star"},
			{ID: "node2", Title: "Test Node 2", Type: "planet"},
		},
		Links: []GraphLink{
			{Source: "node1", Target: "node2", Weight: 0.5, LinkType: "reference"},
		},
	}

	// Cache the graph
	err := graphCache.CacheUserGraph(ctx, userID, testData)
	require.NoError(t, err)

	// Verify the key exists
	key := graphCache.key(userID)
	exists, err := cacheClient.Exists(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Verify TTL is set (should be approximately 5 minutes)
	ttl, err := cacheClient.TTL(ctx, key)
	require.NoError(t, err)
	assert.True(t, ttl > 4*time.Minute && ttl <= 5*time.Minute)
}

func TestGetCachedUserGraph(t *testing.T) {
	ctx := context.Background()
	cacheClient := newTestCacheClient(t)
	graphCache := NewGraphCache(cacheClient)

	userID := "test-user-456"
	testData := GraphData{
		Nodes: []GraphNode{
			{ID: "node1", Title: "Test Node 1", Type: "star"},
		},
		Links: []GraphLink{},
	}

	// Test cache miss
	data, found, err := graphCache.GetCachedUserGraph(ctx, userID)
	require.NoError(t, err)
	assert.False(t, found)
	assert.Equal(t, GraphData{}, data)

	// Cache the graph
	err = graphCache.CacheUserGraph(ctx, userID, testData)
	require.NoError(t, err)

	// Test cache hit
	data, found, err = graphCache.GetCachedUserGraph(ctx, userID)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, testData.Nodes[0].ID, data.Nodes[0].ID)
	assert.Equal(t, testData.Nodes[0].Title, data.Nodes[0].Title)
	assert.Equal(t, testData.Nodes[0].Type, data.Nodes[0].Type)

	// Verify hit count increased
	stats := graphCache.GetStats()
	assert.Equal(t, int64(1), stats.HitCount)
	assert.Equal(t, int64(1), stats.MissCount)
	assert.Equal(t, 0.5, stats.HitRate)
}

func TestInvalidateUserGraph(t *testing.T) {
	ctx := context.Background()
	cacheClient := newTestCacheClient(t)
	graphCache := NewGraphCache(cacheClient)

	userID := "test-user-789"
	testData := GraphData{
		Nodes: []GraphNode{
			{ID: "node1", Title: "Test Node 1", Type: "star"},
		},
		Links: []GraphLink{},
	}

	// Cache the graph
	err := graphCache.CacheUserGraph(ctx, userID, testData)
	require.NoError(t, err)

	// Verify it exists
	_, found, err := graphCache.GetCachedUserGraph(ctx, userID)
	require.NoError(t, err)
	assert.True(t, found)

	// Invalidate the cache
	err = graphCache.InvalidateUserGraph(ctx, userID)
	require.NoError(t, err)

	// Verify it's gone
	_, found, err = graphCache.GetCachedUserGraph(ctx, userID)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestGetStats(t *testing.T) {
	ctx := context.Background()
	cacheClient := newTestCacheClient(t)
	graphCache := NewGraphCache(cacheClient)

	// Initial stats should be zero
	stats := graphCache.GetStats()
	assert.Equal(t, int64(0), stats.HitCount)
	assert.Equal(t, int64(0), stats.MissCount)
	assert.Equal(t, 0.0, stats.HitRate)

	userID := "test-user-stats"
	testData := GraphData{
		Nodes: []GraphNode{{ID: "node1", Title: "Test", Type: "star"}},
		Links: []GraphLink{},
	}

	// Generate a miss
	_, _, err := graphCache.GetCachedUserGraph(ctx, userID)
	require.NoError(t, err)

	stats = graphCache.GetStats()
	assert.Equal(t, int64(0), stats.HitCount)
	assert.Equal(t, int64(1), stats.MissCount)
	assert.Equal(t, 0.0, stats.HitRate)

	// Cache and generate a hit
	err = graphCache.CacheUserGraph(ctx, userID, testData)
	require.NoError(t, err)

	_, _, err = graphCache.GetCachedUserGraph(ctx, userID)
	require.NoError(t, err)

	stats = graphCache.GetStats()
	assert.Equal(t, int64(1), stats.HitCount)
	assert.Equal(t, int64(1), stats.MissCount)
	assert.Equal(t, 0.5, stats.HitRate)
}

func TestCacheKeyGeneration(t *testing.T) {
	cacheClient := newTestCacheClient(t)
	graphCache := NewGraphCache(cacheClient)

	userID := "user-123"
	expectedKey := "graph:user-123"
	actualKey := graphCache.key(userID)
	assert.Equal(t, expectedKey, actualKey)
}

func TestCacheWithMultipleUsers(t *testing.T) {
	ctx := context.Background()
	cacheClient := newTestCacheClient(t)
	graphCache := NewGraphCache(cacheClient)

	user1 := "user-1"
	user2 := "user-2"

	data1 := GraphData{
		Nodes: []GraphNode{{ID: "node1", Title: "User 1 Node", Type: "star"}},
		Links: []GraphLink{},
	}

	data2 := GraphData{
		Nodes: []GraphNode{{ID: "node2", Title: "User 2 Node", Type: "planet"}},
		Links: []GraphLink{},
	}

	// Cache different data for different users
	err := graphCache.CacheUserGraph(ctx, user1, data1)
	require.NoError(t, err)

	err = graphCache.CacheUserGraph(ctx, user2, data2)
	require.NoError(t, err)

	// Verify each user gets their own data
	retrieved1, found, err := graphCache.GetCachedUserGraph(ctx, user1)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "User 1 Node", retrieved1.Nodes[0].Title)

	retrieved2, found, err := graphCache.GetCachedUserGraph(ctx, user2)
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "User 2 Node", retrieved2.Nodes[0].Title)

	// Invalidate one user's cache
	err = graphCache.InvalidateUserGraph(ctx, user1)
	require.NoError(t, err)

	// Verify user1's cache is gone but user2's remains
	_, found, err = graphCache.GetCachedUserGraph(ctx, user1)
	require.NoError(t, err)
	assert.False(t, found)

	_, found, err = graphCache.GetCachedUserGraph(ctx, user2)
	require.NoError(t, err)
	assert.True(t, found)
}
