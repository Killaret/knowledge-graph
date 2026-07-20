package user

import (
	"context"
	"testing"
	"time"

	userDomain "knowledge-graph/internal/domain/user"
	infracache "knowledge-graph/internal/infrastructure/cache"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	service := NewSettingsService(repo, infracache.NewRedisCacheClient(rdb))
	userID := uuid.New()
	ctx := context.Background()

	cacheKey := "setting:" + userID.String() + ":" + userDomain.SettingKeyPreferredLanguage.String()
	_ = rdb.Set(ctx, cacheKey, `{"value":"cached"}`, time.Hour)

	value, err := service.GetSettingValue(ctx, userID, userDomain.SettingKeyPreferredLanguage)
	assert.NoError(t, err)
	assert.Equal(t, "cached", value["value"])
}

func TestSettingsService_GetSettingValue_InvalidCacheJSON(t *testing.T) {
	repo := new(MockRepository)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	service := NewSettingsService(repo, infracache.NewRedisCacheClient(rdb))
	userID := uuid.New()
	ctx := context.Background()

	cacheKey := "setting:" + userID.String() + ":" + userDomain.SettingKeyPreferredLanguage.String()
	_ = rdb.Set(ctx, cacheKey, `not-json`, time.Hour)

	setting, _ := userDomain.NewUserSetting(userID, userDomain.SettingKeyPreferredLanguage, userDomain.SettingValue{Value: "en"})
	repo.On("FindByUserIDAndKey", ctx, userID, userDomain.SettingKeyPreferredLanguage).Return(setting, nil).Once()

	value, err := service.GetSettingValue(ctx, userID, userDomain.SettingKeyPreferredLanguage)
	assert.NoError(t, err)
	assert.Equal(t, "en", value["value"])
}

func TestSettingsService_InvalidateCache(t *testing.T) {
	repo := new(MockRepository)
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	service := NewSettingsService(repo, infracache.NewRedisCacheClient(rdb))
	userID := uuid.New()
	ctx := context.Background()

	_ = rdb.Set(ctx, "setting:"+userID.String()+":lang", "en", time.Hour)
	_ = rdb.Set(ctx, "setting:"+userID.String()+":theme", "dark", time.Hour)

	err := service.InvalidateCache(ctx, userID)
	assert.NoError(t, err)

	_, err = rdb.Get(ctx, "setting:"+userID.String()+":lang").Result()
	assert.Error(t, err)
}

func TestSettingsService_InvalidateCache_NoRedis(t *testing.T) {
	repo := new(MockRepository)
	service := NewSettingsService(repo, nil)

	err := service.InvalidateCache(context.Background(), uuid.New())
	assert.NoError(t, err)
}
