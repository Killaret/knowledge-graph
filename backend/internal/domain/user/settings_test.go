package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingKey_Validate(t *testing.T) {
	tests := []struct {
		name    string
		key     SettingKey
		wantErr bool
	}{
		{
			name:    "valid galactic_mode",
			key:     SettingKeyGalacticMode,
			wantErr: false,
		},
		{
			name:    "valid show_achievement_notifications",
			key:     SettingKeyShowAchievementNotifications,
			wantErr: false,
		},
		{
			name:    "valid preferred_language",
			key:     SettingKeyPreferredLanguage,
			wantErr: false,
		},
		{
			name:    "invalid key",
			key:     SettingKey("invalid_key"),
			wantErr: true,
		},
		{
			name:    "empty key",
			key:     SettingKey(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.key.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSettingKey_String(t *testing.T) {
	key := SettingKeyGalacticMode
	assert.Equal(t, "galactic_mode", key.String())
}

func TestNewUserSetting(t *testing.T) {
	tests := []struct {
		name    string
		userID  uuid.UUID
		key     SettingKey
		value   interface{}
		wantErr bool
	}{
		{
			name:    "valid boolean setting",
			userID:  uuid.New(),
			key:     SettingKeyGalacticMode,
			value:   true,
			wantErr: false,
		},
		{
			name:    "valid string setting",
			userID:  uuid.New(),
			key:     SettingKeyPreferredLanguage,
			value:   "en",
			wantErr: false,
		},
		{
			name:    "valid object setting",
			userID:  uuid.New(),
			key:     SettingKeyGalacticMode,
			value:   map[string]interface{}{"enabled": true, "level": 5},
			wantErr: false,
		},
		{
			name:    "invalid key",
			userID:  uuid.New(),
			key:     SettingKey("invalid"),
			value:   "test",
			wantErr: true,
		},
		{
			name:    "nil value",
			userID:  uuid.New(),
			key:     SettingKeyGalacticMode,
			value:   nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting, err := NewUserSetting(tt.userID, tt.key, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, setting)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, setting)
				assert.NotEqual(t, uuid.Nil, setting.ID())
				assert.Equal(t, tt.userID, setting.UserID())
				assert.Equal(t, tt.key, setting.Key())
				assert.False(t, setting.CreatedAt().IsZero())
				assert.False(t, setting.UpdatedAt().IsZero())
			}
		})
	}
}

func TestReconstructUserSetting(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	value := []byte(`{"value": true}`)
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	setting, err := ReconstructUserSetting(id, userID, "galactic_mode", value, createdAt, updatedAt)

	require.NoError(t, err)
	assert.Equal(t, id, setting.ID())
	assert.Equal(t, userID, setting.UserID())
	assert.Equal(t, SettingKeyGalacticMode, setting.Key())
	assert.JSONEq(t, string(value), string(setting.Value()))
	assert.Equal(t, createdAt, setting.CreatedAt())
	assert.Equal(t, updatedAt, setting.UpdatedAt())
}

func TestUserSetting_Getters(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	key := SettingKeyGalacticMode
	value := []byte(`{"value": true}`)

	setting := &UserSetting{
		id:        id,
		userID:    userID,
		key:       key,
		value:     value,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}

	assert.Equal(t, id, setting.ID())
	assert.Equal(t, userID, setting.UserID())
	assert.Equal(t, key, setting.Key())
	assert.JSONEq(t, string(value), string(setting.Value()))
}

func TestUserSetting_GetValue(t *testing.T) {
	tests := []struct {
		name    string
		value   []byte
		wantErr bool
	}{
		{
			name:    "valid object value",
			value:   []byte(`{"value": true, "level": 5}`),
			wantErr: false,
		},
		{
			name:    "empty object",
			value:   []byte(`{}`),
			wantErr: false,
		},
		{
			name:    "invalid json",
			value:   []byte(`{invalid}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &UserSetting{value: tt.value}

			result, err := setting.GetValue()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestUserSetting_GetBoolValue(t *testing.T) {
	tests := []struct {
		name     string
		value    []byte
		expected bool
		wantErr  bool
	}{
		{
			name:     "true value",
			value:    []byte(`{"value": true}`),
			expected: true,
			wantErr:  false,
		},
		{
			name:     "false value",
			value:    []byte(`{"value": false}`),
			expected: false,
			wantErr:  false,
		},
		{
			name:    "invalid json",
			value:   []byte(`{invalid}`),
			wantErr: true,
		},
		{
			name:    "not a boolean",
			value:   []byte(`{"value": "string"}`),
			wantErr: true,
		},
		{
			name:    "missing value key",
			value:   []byte(`{"other": true}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &UserSetting{value: tt.value}

			result, err := setting.GetBoolValue()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestUserSetting_GetStringValue(t *testing.T) {
	tests := []struct {
		name     string
		value    []byte
		expected string
		wantErr  bool
	}{
		{
			name:     "valid string",
			value:    []byte(`{"value": "hello"}`),
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "empty string",
			value:    []byte(`{"value": ""}`),
			expected: "",
			wantErr:  false,
		},
		{
			name:    "invalid json",
			value:   []byte(`{invalid}`),
			wantErr: true,
		},
		{
			name:    "not a string",
			value:   []byte(`{"value": true}`),
			wantErr: true,
		},
		{
			name:    "missing value key",
			value:   []byte(`{"other": "test"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &UserSetting{value: tt.value}

			result, err := setting.GetStringValue()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSettingValue(t *testing.T) {
	value := SettingValue{Value: "test"}
	assert.Equal(t, "test", value.Value)
}
