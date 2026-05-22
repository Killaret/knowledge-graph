package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, client
}

func TestNewGraphCache(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	cache := NewGraphCache(client)
	assert.NotNil(t, cache)
	assert.Equal(t, 5*time.Minute, cache.ttl)
}

func TestCacheUserGraph(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	graphCache := NewGraphCache(client)

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

	// Verify the key exists in Redis
	key := graphCache.key(userID)
	exists, err := client.Exists(ctx, key).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), exists)

	// Verify TTL is set (should be approximately 5 minutes)
	ttl, err := client.TTL(ctx, key).Result()
	require.NoError(t, err)
	assert.True(t, ttl > 4*time.Minute && ttl <= 5*time.Minute)
}

func TestGetCachedUserGraph(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	graphCache := NewGraphCache(client)

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
	mr, client := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	graphCache := NewGraphCache(client)

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
	mr, client := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	graphCache := NewGraphCache(client)

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
	mr, client := setupTestRedis(t)
	defer mr.Close()

	graphCache := NewGraphCache(client)

	userID := "user-123"
	expectedKey := "graph:user-123"
	actualKey := graphCache.key(userID)
	assert.Equal(t, expectedKey, actualKey)
}

func TestCacheWithMultipleUsers(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	graphCache := NewGraphCache(client)

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
