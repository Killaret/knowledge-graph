package auth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedisStore(t *testing.T) (*RedisTokenStore, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return NewRedisTokenStore(rdb), mr
}

func TestRedisTokenStore_BlacklistAndCheck(t *testing.T) {
	ctx := context.Background()
	store, _ := setupTestRedisStore(t)

	token := "my-token"
	require.NoError(t, store.BlacklistToken(ctx, token, time.Hour))

	blacklisted, err := store.IsTokenBlacklisted(ctx, token)
	require.NoError(t, err)
	assert.True(t, blacklisted)

	blacklisted, err = store.IsTokenBlacklisted(ctx, "other-token")
	require.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestRedisTokenStore_RefreshToken(t *testing.T) {
	ctx := context.Background()
	store, _ := setupTestRedisStore(t)

	token := "refresh-token"
	userID := "user-123"
	expiresAt := time.Now().Add(time.Hour)

	require.NoError(t, store.StoreRefreshToken(ctx, userID, token, expiresAt))

	got, err := store.ValidateRefreshToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, userID, got)

	expired := time.Now().Add(-time.Hour)
	assert.Error(t, store.StoreRefreshToken(ctx, userID, "expired", expired))
}

func TestRedisTokenStore_RevokeRefreshToken(t *testing.T) {
	ctx := context.Background()
	store, _ := setupTestRedisStore(t)

	token := "refresh-token"
	userID := "user-123"
	expiresAt := time.Now().Add(time.Hour)

	require.NoError(t, store.StoreRefreshToken(ctx, userID, token, expiresAt))
	require.NoError(t, store.RevokeRefreshToken(ctx, token, time.Hour))

	_, err := store.ValidateRefreshToken(ctx, token)
	assert.Error(t, err)
}

func TestRedisTokenStore_PasswordResetToken(t *testing.T) {
	ctx := context.Background()
	store, _ := setupTestRedisStore(t)

	token := "reset-token"
	userID := "user-123"

	require.NoError(t, store.StorePasswordResetToken(ctx, userID, token, time.Hour))

	got, err := store.ValidatePasswordResetToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, userID, got)

	require.NoError(t, store.DeletePasswordResetToken(ctx, token))

	_, err = store.ValidatePasswordResetToken(ctx, token)
	assert.Error(t, err)
}

func TestRedisTokenStore_PKCE(t *testing.T) {
	ctx := context.Background()
	store, _ := setupTestRedisStore(t)

	pkce := &PKCE{
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
		CodeVerifier:        "verifier",
	}

	require.NoError(t, store.StorePKCE(ctx, "state", pkce, time.Hour))

	got, err := store.GetPKCE(ctx, "state")
	require.NoError(t, err)
	assert.Equal(t, pkce, got)

	_, err = store.GetPKCE(ctx, "state")
	assert.Error(t, err)
}

func TestRedisTokenStore_PermissionCache(t *testing.T) {
	ctx := context.Background()
	store, _ := setupTestRedisStore(t)

	userID := "user-123"
	require.NoError(t, store.CachePermission(ctx, userID, "notes", "read", true, time.Hour))

	allowed, cached, err := store.CheckCachedPermission(ctx, userID, "notes", "read")
	require.NoError(t, err)
	assert.True(t, cached)
	assert.True(t, allowed)

	_, cached, err = store.CheckCachedPermission(ctx, userID, "notes", "delete")
	require.NoError(t, err)
	assert.False(t, cached)

	require.NoError(t, store.CachePermission(ctx, userID, "notes", "write", false, time.Hour))
	allowed, cached, err = store.CheckCachedPermission(ctx, userID, "notes", "write")
	require.NoError(t, err)
	assert.True(t, cached)
	assert.False(t, allowed)

	require.NoError(t, store.InvalidatePermissionCache(ctx, userID))
	_, cached, err = store.CheckCachedPermission(ctx, userID, "notes", "read")
	require.NoError(t, err)
	assert.False(t, cached)
}
