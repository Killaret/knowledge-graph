package user

import (
	"context"
	"testing"
	"time"

	"knowledge-graph/internal/domain/cache"
	"knowledge-graph/internal/domain/cache/cachetest"
	userDomain "knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSettingsService_SetString(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)
	userID := uuid.New()
	ctx := context.Background()

	repo.On("Upsert", ctx, mock.AnythingOfType("user.UserSetting")).Return(nil).Once()

	err := service.SetString(ctx, userID, userDomain.SettingKeyPreferredLanguage, "en")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestSettingsService_DeleteSetting_Error(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)
	userID := uuid.New()
	ctx := context.Background()

	repo.On("Delete", ctx, userID, userDomain.SettingKeyGalacticMode).Return(assert.AnError).Once()

	err := service.DeleteSetting(ctx, userID, userDomain.SettingKeyGalacticMode)
	assert.Error(t, err)
}

func TestSettingsService_GetSettingValue_FromCache(t *testing.T) {
	repo := new(MockRepository)
	cacheClient := cachetest.NewFakeCacheClient()
	service := NewSettingsService(repo, cacheClient)
	userID := uuid.New()
	ctx := context.Background()

	cacheKey := "setting:" + userID.String() + ":" + userDomain.SettingKeyPreferredLanguage.String()
	_ = cacheClient.Set(ctx, cacheKey, `{"value":"cached"}`, time.Hour)

	value, err := service.GetSettingValue(ctx, userID, userDomain.SettingKeyPreferredLanguage)
	assert.NoError(t, err)
	assert.Equal(t, "cached", value["value"])
}

func TestSettingsService_GetSettingValue_InvalidCacheJSON(t *testing.T) {
	repo := new(MockRepository)
	cacheClient := cachetest.NewFakeCacheClient()
	service := NewSettingsService(repo, cacheClient)
	userID := uuid.New()
	ctx := context.Background()

	cacheKey := "setting:" + userID.String() + ":" + userDomain.SettingKeyPreferredLanguage.String()
	_ = cacheClient.Set(ctx, cacheKey, `not-json`, time.Hour)

	setting, _ := userDomain.NewUserSetting(userID, userDomain.SettingKeyPreferredLanguage, userDomain.SettingValue{Value: "en"})
	repo.On("FindByUserIDAndKey", ctx, userID, userDomain.SettingKeyPreferredLanguage).Return(setting, nil).Once()

	value, err := service.GetSettingValue(ctx, userID, userDomain.SettingKeyPreferredLanguage)
	assert.NoError(t, err)
	assert.Equal(t, "en", value["value"])
}

func TestSettingsService_InvalidateCache(t *testing.T) {
	repo := new(MockRepository)
	cacheClient := cachetest.NewFakeCacheClient()
	service := NewSettingsService(repo, cacheClient)
	userID := uuid.New()
	ctx := context.Background()

	_ = cacheClient.Set(ctx, "setting:"+userID.String()+":lang", "en", time.Hour)
	_ = cacheClient.Set(ctx, "setting:"+userID.String()+":theme", "dark", time.Hour)

	err := service.InvalidateCache(ctx, userID)
	assert.NoError(t, err)

	_, err = cacheClient.Get(ctx, "setting:"+userID.String()+":lang")
	assert.ErrorIs(t, err, cache.ErrCacheMiss)
}

func TestSettingsService_InvalidateCache_NoRedis(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)

	err := service.InvalidateCache(context.Background(), uuid.New())
	assert.NoError(t, err)
}
