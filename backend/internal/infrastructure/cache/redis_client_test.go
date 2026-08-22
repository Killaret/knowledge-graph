package cache

import (
	"context"
	"testing"
	"time"

	dcache "knowledge-graph/internal/domain/cache"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) dcache.CacheClient {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return NewRedisCacheClient(rdb)
}

func TestRedisCacheClient_GetSetDel(t *testing.T) {
	client := setupTestRedis(t)
	ctx := context.Background()

	_, err := client.Get(ctx, "missing")
	assert.ErrorIs(t, err, dcache.ErrCacheMiss)

	require.NoError(t, client.Set(ctx, "key", "value", 0))

	val, err := client.Get(ctx, "key")
	require.NoError(t, err)
	assert.Equal(t, "value", val)

	require.NoError(t, client.Del(ctx, "key"))

	_, err = client.Get(ctx, "key")
	assert.ErrorIs(t, err, dcache.ErrCacheMiss)
}

func TestRedisCacheClient_ExistsIncrExpire(t *testing.T) {
	client := setupTestRedis(t)
	ctx := context.Background()

	ok, err := client.Exists(ctx, "counter")
	require.NoError(t, err)
	assert.Equal(t, int64(0), ok)

	n, err := client.Incr(ctx, "counter")
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)

	ok, err = client.Exists(ctx, "counter")
	require.NoError(t, err)
	assert.Equal(t, int64(1), ok)

	require.NoError(t, client.Expire(ctx, "counter", time.Hour))
}
