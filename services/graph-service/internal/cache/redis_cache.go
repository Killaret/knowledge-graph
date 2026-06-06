package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"knowledge-graph-graph-service/internal/config"
	"knowledge-graph-graph-service/internal/engine"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client        *redis.Client
	noteLayoutTTL time.Duration
	fullLayoutTTL time.Duration
	deltaTTL      time.Duration
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

// NewRedisCacheWithConfig creates a cache with TTLs loaded from config.
func NewRedisCacheWithConfig(client *redis.Client, cfg *config.Config) *RedisCache {
	return &RedisCache{
		client:        client,
		noteLayoutTTL: cfg.NoteLayoutTTL,
		fullLayoutTTL: cfg.FullLayoutTTL,
		deltaTTL:      cfg.DeltaTTL,
	}
}

func (c *RedisCache) cacheKey(parts ...string) string {
	return fmt.Sprintf("graph-service:%s", strings.Join(parts, ":"))
}

func (c *RedisCache) LoadNoteLayout(ctx context.Context, noteID string, depth int) (*engine.LayoutResponse, string, error) {
	key := c.cacheKey("note", noteID, fmt.Sprintf("depth-%d", depth))
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, "", err
	}
	var stored struct {
		Layout *engine.LayoutResponse `json:"layout"`
		Hash   string                 `json:"hash"`
	}
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nil, "", err
	}
	return stored.Layout, stored.Hash, nil
}

func (c *RedisCache) SaveNoteLayout(ctx context.Context, noteID string, depth int, layout *engine.LayoutResponse, hash string) error {
	key := c.cacheKey("note", noteID, fmt.Sprintf("depth-%d", depth))
	payload, err := json.Marshal(map[string]interface{}{
		"layout": layout,
		"hash":   hash,
	})
	if err != nil {
		return err
	}
	ttl := c.noteLayoutTTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return c.client.Set(ctx, key, payload, ttl).Err()
}

func (c *RedisCache) LoadFullLayout(ctx context.Context, userID string) (*engine.LayoutResponse, string, error) {
	key := c.cacheKey("full", userID)
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, "", err
	}
	var stored struct {
		Layout *engine.LayoutResponse `json:"layout"`
		Hash   string                 `json:"hash"`
	}
	if err := json.Unmarshal(raw, &stored); err != nil {
		return nil, "", err
	}
	return stored.Layout, stored.Hash, nil
}

func (c *RedisCache) SaveFullLayout(ctx context.Context, userID string, layout *engine.LayoutResponse, hash string) error {
	key := c.cacheKey("full", userID)
	payload, err := json.Marshal(map[string]interface{}{
		"layout": layout,
		"hash":   hash,
	})
	if err != nil {
		return err
	}
	ttl := c.fullLayoutTTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return c.client.Set(ctx, key, payload, ttl).Err()
}

func (c *RedisCache) LoadDelta(ctx context.Context, userID, lastHash string) (*engine.DeltaResponse, error) {
	key := c.cacheKey("delta", userID, lastHash)
	raw, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var delta engine.DeltaResponse
	if err := json.Unmarshal(raw, &delta); err != nil {
		return nil, err
	}
	return &delta, nil
}

func (c *RedisCache) SaveDelta(ctx context.Context, userID, lastHash string, delta *engine.DeltaResponse) error {
	key := c.cacheKey("delta", userID, lastHash)
	payload, err := json.Marshal(delta)
	if err != nil {
		return err
	}
	ttl := c.deltaTTL
	if ttl == 0 {
		ttl = 1 * time.Minute
	}
	return c.client.Set(ctx, key, payload, ttl).Err()
}

func (c *RedisCache) InvalidateAll(ctx context.Context) error {
	pattern := c.cacheKey("*")
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
