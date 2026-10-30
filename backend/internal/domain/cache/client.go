package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss indicates that the requested key is not present in the cache.
var ErrCacheMiss = errors.New("cache miss")

// CacheClient is a domain-level port for key/value cache operations.
// Infrastructure adapters (e.g. Redis) implement this interface.
type CacheClient interface {
	// Get returns the cached value for key. If the key is missing, it must return ErrCacheMiss.
	Get(ctx context.Context, key string) (string, error)

	// Set stores value under key with the given TTL.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Del removes one or more keys.
	Del(ctx context.Context, keys ...string) error

	// Exists returns the number of matching keys that exist.
	Exists(ctx context.Context, key string) (int64, error)

	// Incr increments the integer value of key by one.
	Incr(ctx context.Context, key string) (int64, error)

	// Expire sets a TTL on a key.
	Expire(ctx context.Context, key string, ttl time.Duration) error

	// Scan returns a batch of keys matching the pattern and the next cursor.
	// Callers should loop until the returned cursor is 0.
	Scan(ctx context.Context, cursor uint64, match string, count int64) (keys []string, nextCursor uint64, err error)
}
