package cache

import (
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// TestNewRedisCache tests the constructor
func TestNewRedisCache(t *testing.T) {
	// Test that nil client is handled gracefully
	var client *redis.Client
	cache := NewRedisCache(client)
	assert.NotNil(t, cache)
}

// TestCacheKey tests the key generation
func TestCacheKey(t *testing.T) {
	var client *redis.Client
	cache := NewRedisCache(client)
	
	// Test basic key generation
	key := cache.cacheKey("note", "123", "depth-2")
	assert.Equal(t, "graph-service:note:123:depth-2", key)
	
	// Test single part
	key = cache.cacheKey("full")
	assert.Equal(t, "graph-service:full", key)
	
	// Test multiple parts
	key = cache.cacheKey("delta", "user1", "hash123")
	assert.Equal(t, "graph-service:delta:user1:hash123", key)
}