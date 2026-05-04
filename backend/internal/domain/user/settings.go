// Package user provides domain types for user settings
package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SettingKey defines valid setting keys
 type SettingKey string

// Valid setting keys
const (
	SettingKeyGalacticMode               SettingKey = "galactic_mode"
	SettingKeyShowAchievementNotifications SettingKey = "show_achievement_notifications"
	SettingKeyPreferredLanguage          SettingKey = "preferred_language"
)

// ValidSettingKeys contains all valid setting keys
var ValidSettingKeys = map[SettingKey]bool{
	SettingKeyGalacticMode:               true,
	SettingKeyShowAchievementNotifications: true,
	SettingKeyPreferredLanguage:          true,
}

// Validate checks if the setting key is valid
func (k SettingKey) Validate() error {
	if !ValidSettingKeys[k] {
		return fmt.Errorf("invalid setting key: %s", k)
	}
	return nil
}

// String returns the string representation
func (k SettingKey) String() string {
	return string(k)
}

// UserSetting represents a single user setting
type UserSetting struct {
	id        uuid.UUID
	userID    uuid.UUID
	key       SettingKey
	value     json.RawMessage
	createdAt time.Time
	updatedAt time.Time
}

// NewUserSetting creates a new user setting
func NewUserSetting(userID uuid.UUID, key SettingKey, value interface{}) (*UserSetting, error) {
	if err := key.Validate(); err != nil {
		return nil, err
	}

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}

	now := time.Now()
	return &UserSetting{
		id:        uuid.New(),
		userID:    userID,
		key:       key,
		value:     valueBytes,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstructUserSetting reconstructs a setting from stored data
func ReconstructUserSetting(id, userID uuid.UUID, key string, value []byte, createdAt, updatedAt time.Time) (*UserSetting, error) {
	return &UserSetting{
		id:        id,
		userID:    userID,
		key:       SettingKey(key),
		value:     value,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

// ID returns the setting ID
func (s *UserSetting) ID() uuid.UUID {
	return s.id
}

// UserID returns the user ID
func (s *UserSetting) UserID() uuid.UUID {
	return s.userID
}

// Key returns the setting key
func (s *UserSetting) Key() SettingKey {
	return s.key
}

// Value returns the raw JSON value
func (s *UserSetting) Value() json.RawMessage {
	return s.value
}

// GetValue returns the value as a map
func (s *UserSetting) GetValue() (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(s.value, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}
	return result, nil
}

// GetBoolValue extracts a boolean value from the setting
func (s *UserSetting) GetBoolValue() (bool, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(s.value, &data); err != nil {
		return false, fmt.Errorf("failed to unmarshal value: %w", err)
	}
	val, ok := data["value"].(bool)
	if !ok {
		return false, fmt.Errorf("value is not a boolean")
	}
	return val, nil
}

// GetStringValue extracts a string value from the setting
func (s *UserSetting) GetStringValue() (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(s.value, &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal value: %w", err)
	}
	val, ok := data["value"].(string)
	if !ok {
		return "", fmt.Errorf("value is not a string")
	}
	return val, nil
}

// CreatedAt returns creation time
func (s *UserSetting) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns last update time
func (s *UserSetting) UpdatedAt() time.Time {
	return s.updatedAt
}

// SettingValue wraps a setting value for storage
type SettingValue struct {
	Value interface{} `json:"value"`
}

// UserSettingsRepository defines the interface for user settings storage
type UserSettingsRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]UserSetting, error)
	FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key SettingKey) (*UserSetting, error)
	Upsert(ctx context.Context, setting UserSetting) error
	Delete(ctx context.Context, userID uuid.UUID, key SettingKey) error
}
