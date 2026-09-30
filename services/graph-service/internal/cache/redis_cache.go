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
	snapshotTTL   time.Duration
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
		snapshotTTL:   cfg.SnapshotTTL,
	}
}

func (c *RedisCache) cacheKey(parts ...string) string {
	return fmt.Sprintf("graph-service:%s", strings.Join(parts, ":"))
}

func (c *RedisCache) LoadNoteLayout(ctx context.Context, userID, noteID string, depth int) (*engine.LayoutResponse, string, error) {
	key := c.cacheKey("note", userID, noteID, fmt.Sprintf("depth-%d", depth))
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

func (c *RedisCache) SaveNoteLayout(ctx context.Context, userID, noteID string, depth int, layout *engine.LayoutResponse, hash string) error {
	key := c.cacheKey("note", userID, noteID, fmt.Sprintf("depth-%d", depth))
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

// Snapshot is an immutable served-layout record keyed by the data hash the
// client holds. Snapshots survive event-based cache invalidation: only the
// "full" current pointer is cleared, so a delta can always be computed against
// the version the client actually has — or answered with resync once the
// snapshot expires.
type Snapshot struct {
	Layout     *engine.LayoutResponse `json:"layout"`
	LayoutKind string                 `json:"layout_kind"`
}

func (c *RedisCache) snapshotKey(userID, hash string) string {
	return c.cacheKey("snapshot", userID, hash)
}

// LoadSnapshot returns the layout previously served under hash, or nil when no
// snapshot exists (expired or never stored). A nil snapshot means the caller
// must answer resync — never fall back to the current full cache.
func (c *RedisCache) LoadSnapshot(ctx context.Context, userID, hash string) (*Snapshot, error) {
	raw, err := c.client.Get(ctx, c.snapshotKey(userID, hash)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var snap Snapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

func (c *RedisCache) SaveSnapshot(ctx context.Context, userID, hash, layoutKind string, layout *engine.LayoutResponse) error {
	payload, err := json.Marshal(Snapshot{Layout: layout, LayoutKind: layoutKind})
	if err != nil {
		return err
	}
	ttl := c.snapshotTTL
	if ttl == 0 {
		// Several poll intervals by default; clients holding this hash keep
		// getting real deltas instead of a silent "everything added".
		ttl = 15 * time.Minute
	}
	return c.client.Set(ctx, c.snapshotKey(userID, hash), payload, ttl).Err()
}

// InvalidateAll clears every cache key except snapshots — historical
// snapshots are the baseline client deltas are computed against and must
// outlive cache clears until their own TTL expires.
func (c *RedisCache) InvalidateAll(ctx context.Context) error {
	pattern := c.cacheKey("*")
	snapshotPrefix := c.cacheKey("snapshot") + ":"
	iter := c.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		if strings.HasPrefix(key, snapshotPrefix) {
			continue
		}
		if err := c.client.Del(ctx, key).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}
