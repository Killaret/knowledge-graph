// Package user provides application services for user settings
package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	dcache "knowledge-graph/internal/domain/cache"
	userDomain "knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
)

// SettingsService provides user settings management with caching
type SettingsService struct {
	repo        userDomain.UserSettingsRepository
	cacheClient dcache.CacheClient
	cacheTTL    time.Duration
}

// NewSettingsService creates a new settings service
func NewSettingsService(repo userDomain.UserSettingsRepository, cacheClient dcache.CacheClient) *SettingsService {
	return &SettingsService{
		repo:        repo,
		cacheClient: cacheClient,
		cacheTTL:    5 * time.Minute,
	}
}

// cacheKey generates a cache key for a setting
func (s *SettingsService) cacheKey(userID uuid.UUID, key string) string {
	return fmt.Sprintf("setting:%s:%s", userID.String(), key)
}

// GetSettings returns all settings for a user
func (s *SettingsService) GetSettings(ctx context.Context, userID uuid.UUID) ([]userDomain.UserSetting, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// GetSettingValue retrieves a setting value with caching
func (s *SettingsService) GetSettingValue(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) (map[string]interface{}, error) {
	cacheKey := s.cacheKey(userID, key.String())

	// Try cache first
	if s.cacheClient != nil {
		cached, err := s.cacheClient.Get(ctx, cacheKey)
		if err == nil {
			var value map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &value); err == nil {
				return value, nil
			}
		}
	}

	// Fetch from database
	setting, err := s.repo.FindByUserIDAndKey(ctx, userID, key)
	if err != nil {
		return nil, err
	}

	if setting == nil {
		// Return default value
		return s.getDefaultValue(key), nil
	}

	value, err := setting.GetValue()
	if err != nil {
		return nil, err
	}

	// Cache the value
	if s.cacheClient != nil {
		data, _ := json.Marshal(value)
		_ = s.cacheClient.Set(ctx, cacheKey, string(data), s.cacheTTL)
	}

	return value, nil
}

// getDefaultValue returns the default value for a setting
func (s *SettingsService) getDefaultValue(key userDomain.SettingKey) map[string]interface{} {
	switch key {
	case userDomain.SettingKeyGalacticMode:
		return map[string]interface{}{"value": false}
	case userDomain.SettingKeyShowAchievementNotifications:
		return map[string]interface{}{"value": true}
	case userDomain.SettingKeyPreferredLanguage:
		return map[string]interface{}{"value": "ru"}
	default:
		return map[string]interface{}{"value": nil}
	}
}

// GetBool retrieves a boolean setting value
func (s *SettingsService) GetBool(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) (bool, error) {
	value, err := s.GetSettingValue(ctx, userID, key)
	if err != nil {
		return false, err
	}

	val, ok := value["value"].(bool)
	if !ok {
		// Return default
		defaultVal := s.getDefaultValue(key)
		defBool, _ := defaultVal["value"].(bool)
		return defBool, nil
	}
	return val, nil
}

// GetString retrieves a string setting value
func (s *SettingsService) GetString(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) (string, error) {
	value, err := s.GetSettingValue(ctx, userID, key)
	if err != nil {
		return "", err
	}

	val, ok := value["value"].(string)
	if !ok {
		// Return default
		defaultVal := s.getDefaultValue(key)
		defStr, _ := defaultVal["value"].(string)
		return defStr, nil
	}
	return val, nil
}

// SetValue sets a setting value with any type
func (s *SettingsService) SetValue(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey, value interface{}) error {
	if err := key.Validate(); err != nil {
		return err
	}

	settingValue := userDomain.SettingValue{Value: value}
	setting, err := userDomain.NewUserSetting(userID, key, settingValue)
	if err != nil {
		return err
	}

	// Save to database
	if err := s.repo.Upsert(ctx, *setting); err != nil {
		return fmt.Errorf("failed to save setting: %w", err)
	}

	// Invalidate cache
	if s.cacheClient != nil {
		cacheKey := s.cacheKey(userID, key.String())
		_ = s.cacheClient.Del(ctx, cacheKey)
	}

	return nil
}

// SetBool sets a boolean setting value
func (s *SettingsService) SetBool(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey, val bool) error {
	return s.SetValue(ctx, userID, key, val)
}

// SetString sets a string setting value
func (s *SettingsService) SetString(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey, val string) error {
	return s.SetValue(ctx, userID, key, val)
}

// DeleteSetting deletes a setting
func (s *SettingsService) DeleteSetting(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) error {
	if err := s.repo.Delete(ctx, userID, key); err != nil {
		return err
	}

	// Invalidate cache
	if s.cacheClient != nil {
		cacheKey := s.cacheKey(userID, key.String())
		_ = s.cacheClient.Del(ctx, cacheKey)
	}

	return nil
}

// InvalidateCache clears all settings cache for a user
func (s *SettingsService) InvalidateCache(ctx context.Context, userID uuid.UUID) error {
	if s.cacheClient == nil {
		return nil
	}

	pattern := fmt.Sprintf("setting:%s:*", userID.String())

	// Use scan to find and delete keys
	var cursor uint64
	for {
		keys, nextCursor, err := s.cacheClient.Scan(ctx, cursor, pattern, 100)
		if err != nil {
			return err
		}

		if len(keys) > 0 {
			if err := s.cacheClient.Del(ctx, keys...); err != nil {
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
