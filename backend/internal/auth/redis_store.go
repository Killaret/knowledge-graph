// Package auth provides Redis-backed token storage
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisTokenStore provides storage for tokens and blacklisting
type RedisTokenStore struct {
	client *redis.Client
	prefix string
}

// NewRedisTokenStore creates a new Redis token store
func NewRedisTokenStore(client *redis.Client) *RedisTokenStore {
	return &RedisTokenStore{
		client: client,
		prefix: "auth:",
	}
}

// key generates a prefixed Redis key
func (s *RedisTokenStore) key(suffix string) string {
	return s.prefix + suffix
}

// hashToken creates a hash of a token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// BlacklistToken adds a token to the blacklist with a TTL
func (s *RedisTokenStore) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	hash := hashToken(token)
	key := s.key("blacklist:" + hash)
	return s.client.Set(ctx, key, "1", ttl).Err()
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *RedisTokenStore) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	hash := hashToken(token)
	key := s.key("blacklist:" + hash)
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// StoreRefreshToken stores a refresh token with metadata
func (s *RedisTokenStore) StoreRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error {
	hash := hashToken(token)
	key := s.key("refresh:" + hash)
	
	data := map[string]interface{}{
		"user_id": userID,
		"created": time.Now().Unix(),
	}
	
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return fmt.Errorf("token already expired")
	}
	
	return s.client.HSet(ctx, key, data).Err()
}

// ValidateRefreshToken checks if a refresh token is valid and returns the user ID
func (s *RedisTokenStore) ValidateRefreshToken(ctx context.Context, token string) (string, error) {
	hash := hashToken(token)
	key := s.key("refresh:" + hash)
	
	// Check if blacklisted first
	blacklisted, err := s.IsTokenBlacklisted(ctx, token)
	if err != nil {
		return "", err
	}
	if blacklisted {
		return "", fmt.Errorf("token has been revoked")
	}
	
	// Get token data
	userID, err := s.client.HGet(ctx, key, "user_id").Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token not found")
	}
	if err != nil {
		return "", err
	}
	
	return userID, nil
}

// RevokeRefreshToken blacklists a refresh token
func (s *RedisTokenStore) RevokeRefreshToken(ctx context.Context, token string, ttl time.Duration) error {
	hash := hashToken(token)
	key := s.key("refresh:" + hash)
	
	// Delete from active tokens
	if err := s.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	
	// Add to blacklist
	return s.BlacklistToken(ctx, token, ttl)
}

// StorePasswordResetToken stores a password reset token
func (s *RedisTokenStore) StorePasswordResetToken(ctx context.Context, userID string, token string, ttl time.Duration) error {
	hash := hashToken(token)
	key := s.key("password_reset:" + hash)
	
	return s.client.Set(ctx, key, userID, ttl).Err()
}

// ValidatePasswordResetToken validates a password reset token and returns the user ID
func (s *RedisTokenStore) ValidatePasswordResetToken(ctx context.Context, token string) (string, error) {
	hash := hashToken(token)
	key := s.key("password_reset:" + hash)
	
	userID, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("invalid or expired token")
	}
	if err != nil {
		return "", err
	}
	
	return userID, nil
}

// DeletePasswordResetToken removes a password reset token
func (s *RedisTokenStore) DeletePasswordResetToken(ctx context.Context, token string) error {
	hash := hashToken(token)
	key := s.key("password_reset:" + hash)
	return s.client.Del(ctx, key).Err()
}

// StorePKCE stores PKCE parameters for OAuth flow
func (s *RedisTokenStore) StorePKCE(ctx context.Context, state string, pkce *PKCE, ttl time.Duration) error {
	key := s.key("pkce:" + state)
	data := map[string]interface{}{
		"code_challenge":        pkce.CodeChallenge,
		"code_challenge_method": pkce.CodeChallengeMethod,
		"code_verifier":         pkce.CodeVerifier,
	}
	return s.client.HSet(ctx, key, data).Err()
}

// GetPKCE retrieves PKCE parameters by state
func (s *RedisTokenStore) GetPKCE(ctx context.Context, state string) (*PKCE, error) {
	key := s.key("pkce:" + state)
	
	data, err := s.client.HGetAll(ctx, key).Result()
	if err == redis.Nil || len(data) == 0 {
		return nil, fmt.Errorf("pkce data not found")
	}
	if err != nil {
		return nil, err
	}
	
	// Delete after retrieval (one-time use)
	defer s.client.Del(ctx, key)
	
	return &PKCE{
		CodeChallenge:       data["code_challenge"],
		CodeChallengeMethod: data["code_challenge_method"],
		CodeVerifier:        data["code_verifier"],
	}, nil
}

// CachePermission caches user permissions in Redis
func (s *RedisTokenStore) CachePermission(ctx context.Context, userID, resource, action string, allowed bool, ttl time.Duration) error {
	key := s.key(fmt.Sprintf("perm:%s:%s:%s", userID, resource, action))
	value := "0"
	if allowed {
		value = "1"
	}
	return s.client.Set(ctx, key, value, ttl).Err()
}

// CheckCachedPermission checks cached user permissions
func (s *RedisTokenStore) CheckCachedPermission(ctx context.Context, userID, resource, action string) (bool, bool, error) {
	key := s.key(fmt.Sprintf("perm:%s:%s:%s", userID, resource, action))
	
	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, false, nil // Not cached
	}
	if err != nil {
		return false, false, err
	}
	
	return val == "1", true, nil
}

// InvalidatePermissionCache clears permission cache for a user
func (s *RedisTokenStore) InvalidatePermissionCache(ctx context.Context, userID string) error {
	pattern := s.key("perm:" + userID + ":*")
	
	// Use scan to find and delete keys
	var cursor uint64
	for {
		keys, nextCursor, err := s.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		
		if len(keys) > 0 {
			if err := s.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	
	return nil
}
