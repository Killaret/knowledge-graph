package settings

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/application/user"
	userDomain "knowledge-graph/internal/domain/user"
	infracache "knowledge-graph/internal/infrastructure/cache"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserSettingsRepo struct {
	mock.Mock
}

func (m *mockUserSettingsRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]userDomain.UserSetting, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]userDomain.UserSetting), args.Error(1)
}

func (m *mockUserSettingsRepo) FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) (*userDomain.UserSetting, error) {
	args := m.Called(ctx, userID, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.UserSetting), args.Error(1)
}

func (m *mockUserSettingsRepo) Upsert(ctx context.Context, setting userDomain.UserSetting) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

func (m *mockUserSettingsRepo) Delete(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) error {
	args := m.Called(ctx, userID, key)
	return args.Error(0)
}

func setupSettingsHandler() (*gin.Engine, *mockUserSettingsRepo) {
	gin.SetMode(gin.TestMode)
	repo := new(mockUserSettingsRepo)
	rdb := redis.NewClient(&redis.Options{Addr: miniredis.RunT(&testing.T{}).Addr()})
	service := user.NewSettingsService(repo, infracache.NewRedisCacheClient(rdb))
	h := NewHandler(service)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
		c.Next()
	})
	r.GET("/settings", h.GetMySettings)
	r.GET("/settings/:key", h.GetSetting)
	r.POST("/settings", h.UpdateSetting)
	r.DELETE("/settings/:key", h.DeleteSetting)
	r.GET("/settings/galactic", h.GetGalacticMode)
	r.POST("/settings/galactic/toggle", h.ToggleGalacticMode)
	return r, repo
}

func TestGetMySettings(t *testing.T) {
	r, repo := setupSettingsHandler()
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	settingValue := userDomain.SettingValue{Value: "ru"}
	setting, err := userDomain.NewUserSetting(uid, userDomain.SettingKeyPreferredLanguage, settingValue)
	require.NoError(t, err)

	repo.On("FindByUserID", mock.Anything, uid).Return([]userDomain.UserSetting{*setting}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSetting(t *testing.T) {
	r, repo := setupSettingsHandler()
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	repo.On("FindByUserIDAndKey", mock.Anything, uid, userDomain.SettingKeyPreferredLanguage).Return(nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings/preferred_language", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetSetting_InvalidKey(t *testing.T) {
	r, _ := setupSettingsHandler()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings/invalid_key", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSetting(t *testing.T) {
	r, repo := setupSettingsHandler()

	repo.On("Upsert", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(map[string]interface{}{"key": "preferred_language", "value": "en"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateSetting_InvalidKey(t *testing.T) {
	r, _ := setupSettingsHandler()

	body, _ := json.Marshal(map[string]interface{}{"key": "invalid_key", "value": "en"})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteSetting(t *testing.T) {
	r, repo := setupSettingsHandler()
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	repo.On("Delete", mock.Anything, uid, userDomain.SettingKeyPreferredLanguage).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/settings/preferred_language", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGalacticModeEndpoints(t *testing.T) {
	r, repo := setupSettingsHandler()
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	repo.On("FindByUserIDAndKey", mock.Anything, uid, userDomain.SettingKeyGalacticMode).Return(nil, nil)
	repo.On("Upsert", mock.Anything, mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/settings/galactic", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/settings/galactic/toggle", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
