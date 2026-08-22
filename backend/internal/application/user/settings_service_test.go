package user

import (
	"context"
	"testing"

	userDomain "knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository is a mock implementation of UserSettingsRepository
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]userDomain.UserSetting, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]userDomain.UserSetting), args.Error(1)
}

func (m *MockRepository) FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) (*userDomain.UserSetting, error) {
	args := m.Called(ctx, userID, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.UserSetting), args.Error(1)
}

func (m *MockRepository) Upsert(ctx context.Context, setting userDomain.UserSetting) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) error {
	args := m.Called(ctx, userID, key)
	return args.Error(0)
}

func TestSettingsService_GetBool(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)
	userID := uuid.New()
	ctx := context.Background()

	t.Run("returns value from database", func(t *testing.T) {
		setting, _ := userDomain.NewUserSetting(userID, userDomain.SettingKeyGalacticMode, userDomain.SettingValue{Value: true})
		repo.On("FindByUserIDAndKey", ctx, userID, userDomain.SettingKeyGalacticMode).Return(setting, nil).Once()

		value, err := service.GetBool(ctx, userID, userDomain.SettingKeyGalacticMode)

		assert.NoError(t, err)
		assert.True(t, value)
	})

	t.Run("returns default when not found", func(t *testing.T) {
		repo.On("FindByUserIDAndKey", ctx, userID, userDomain.SettingKeyGalacticMode).Return(nil, nil).Once()

		value, err := service.GetBool(ctx, userID, userDomain.SettingKeyGalacticMode)

		assert.NoError(t, err)
		assert.False(t, value) // Default is false
	})
}

func TestSettingsService_SetBool(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)
	userID := uuid.New()
	ctx := context.Background()

	t.Run("saves boolean value", func(t *testing.T) {
		repo.On("Upsert", ctx, mock.AnythingOfType("user.UserSetting")).Return(nil).Once()

		err := service.SetBool(ctx, userID, userDomain.SettingKeyGalacticMode, true)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestSettingsService_GetString(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)
	userID := uuid.New()
	ctx := context.Background()

	t.Run("returns string value from database", func(t *testing.T) {
		setting, _ := userDomain.NewUserSetting(userID, userDomain.SettingKeyPreferredLanguage, userDomain.SettingValue{Value: "en"})
		repo.On("FindByUserIDAndKey", ctx, userID, userDomain.SettingKeyPreferredLanguage).Return(setting, nil).Once()

		value, err := service.GetString(ctx, userID, userDomain.SettingKeyPreferredLanguage)

		assert.NoError(t, err)
		assert.Equal(t, "en", value)
	})

	t.Run("returns default when not found", func(t *testing.T) {
		repo.On("FindByUserIDAndKey", ctx, userID, userDomain.SettingKeyPreferredLanguage).Return(nil, nil).Once()

		value, err := service.GetString(ctx, userID, userDomain.SettingKeyPreferredLanguage)

		assert.NoError(t, err)
		assert.Equal(t, "ru", value) // Default is "ru"
	})
}

func TestSettingsService_DeleteSetting(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)
	userID := uuid.New()
	ctx := context.Background()

	t.Run("deletes setting", func(t *testing.T) {
		repo.On("Delete", ctx, userID, userDomain.SettingKeyGalacticMode).Return(nil).Once()

		err := service.DeleteSetting(ctx, userID, userDomain.SettingKeyGalacticMode)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestSettingsService_GetSettings(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)
	userID := uuid.New()
	ctx := context.Background()

	t.Run("returns all settings", func(t *testing.T) {
		setting1, _ := userDomain.NewUserSetting(userID, userDomain.SettingKeyGalacticMode, userDomain.SettingValue{Value: true})
		setting2, _ := userDomain.NewUserSetting(userID, userDomain.SettingKeyPreferredLanguage, userDomain.SettingValue{Value: "en"})

		repo.On("FindByUserID", ctx, userID).Return([]userDomain.UserSetting{*setting1, *setting2}, nil).Once()

		settings, err := service.GetSettings(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, settings, 2)
	})
}

func TestSettingsService_SetValue(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)
	userID := uuid.New()
	ctx := context.Background()

	t.Run("rejects invalid key", func(t *testing.T) {
		err := service.SetValue(ctx, userID, "invalid_key", "value")

		assert.Error(t, err)
	})

	t.Run("saves valid setting", func(t *testing.T) {
		repo.On("Upsert", ctx, mock.AnythingOfType("user.UserSetting")).Return(nil).Once()

		err := service.SetValue(ctx, userID, userDomain.SettingKeyGalacticMode, true)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})
}
