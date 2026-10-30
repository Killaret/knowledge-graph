package cache

import (
	"context"
	"time"

	dcache "knowledge-graph/internal/domain/cache"

	"github.com/redis/go-redis/v9"
)

type redisCacheClient struct {
	client *redis.Client
}

// NewRedisCacheClient adapts a go-redis client to the domain cache port.
func NewRedisCacheClient(client *redis.Client) dcache.CacheClient {
	if client == nil {
		return nil
	}
	return &redisCacheClient{client: client}
}

func (r *redisCacheClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", dcache.ErrCacheMiss
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *redisCacheClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisCacheClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *redisCacheClient) Exists(ctx context.Context, key string) (int64, error) {
	return r.client.Exists(ctx, key).Result()
}

func (r *redisCacheClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.client.Incr(ctx, key).Result()
}

func (r *redisCacheClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

func (r *redisCacheClient) Scan(ctx context.Context, cursor uint64, match string, count int64) (keys []string, nextCursor uint64, err error) {
	return r.client.Scan(ctx, cursor, match, count).Result()
}
