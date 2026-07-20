package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	dcache "knowledge-graph/internal/domain/cache"
)

// GraphCache provides caching for user graph data
type GraphCache struct {
	client    dcache.CacheClient
	prefix    string
	ttl       time.Duration
	hitCount  int64
	missCount int64
}

// NewGraphCache creates a new graph cache instance
func NewGraphCache(client dcache.CacheClient) *GraphCache {
	return &GraphCache{
		client: client,
		prefix: "graph:",
		ttl:    5 * time.Minute, // 5 minutes TTL as specified
	}
}

// key generates a prefixed cache key for a user's graph
func (c *GraphCache) key(userID string) string {
	return c.prefix + userID
}

// GraphData represents the cached graph structure
type GraphData struct {
	Nodes []GraphNode `json:"nodes"`
	Links []GraphLink `json:"links"`
}

// GraphNode represents a cached node
type GraphNode struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Type  string  `json:"type"`
	X     float64 `json:"x,omitempty"` // Cached position for instant visual stability
	Y     float64 `json:"y,omitempty"`
}

// GraphLink represents a cached link
type GraphLink struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	LinkType string  `json:"link_type"`
}

// CacheUserGraph stores the user's graph data in cache
func (c *GraphCache) CacheUserGraph(ctx context.Context, userID string, data GraphData) error {
	key := c.key(userID)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal graph data: %w", err)
	}

	if err := c.client.Set(ctx, key, string(jsonData), c.ttl); err != nil {
		return fmt.Errorf("failed to cache graph: %w", err)
	}

	return nil
}

// GetCachedUserGraph retrieves the user's cached graph data from cache
func (c *GraphCache) GetCachedUserGraph(ctx context.Context, userID string) (GraphData, bool, error) {
	key := c.key(userID)

	data, err := c.client.Get(ctx, key)
	if err == dcache.ErrCacheMiss {
		c.missCount++
		return GraphData{}, false, nil
	}
	if err != nil {
		return GraphData{}, false, fmt.Errorf("failed to get cached graph: %w", err)
	}

	c.hitCount++

	var graphData GraphData
	if err := json.Unmarshal([]byte(data), &graphData); err != nil {
		return GraphData{}, false, fmt.Errorf("failed to unmarshal graph data: %w", err)
	}

	return graphData, true, nil
}

// InvalidateUserGraph removes the cached graph for a specific user
func (gc *GraphCache) InvalidateUserGraph(ctx context.Context, userID string) error {
	key := gc.key(userID)
	return gc.client.Del(ctx, key)
}

// InvalidateAll removes all cached graph data
func (gc *GraphCache) InvalidateAll(ctx context.Context) error {
	// Find all keys matching the graph cache pattern
	pattern := "graph:*"

	var allKeys []string
	var cursor uint64
	for {
		keys, nextCursor, err := gc.client.Scan(ctx, cursor, pattern, 100)
		if err != nil {
			return err
		}
		allKeys = append(allKeys, keys...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// Delete all found keys
	if len(allKeys) > 0 {
		return gc.client.Del(ctx, allKeys...)
	}

	return nil
}

// GetStats returns cache statistics
func (c *GraphCache) GetStats() CacheStats {
	total := c.hitCount + c.missCount
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hitCount) / float64(total)
	}

	return CacheStats{
		HitCount:  c.hitCount,
		MissCount: c.missCount,
		HitRate:   hitRate,
	}
}

// CacheStats represents cache statistics
type CacheStats struct {
	HitCount  int64   `json:"hit_count"`
	MissCount int64   `json:"miss_count"`
	HitRate   float64 `json:"hit_rate"`
}
