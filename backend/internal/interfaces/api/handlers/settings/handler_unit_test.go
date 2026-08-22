//go:build !integration
// +build !integration

package settings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/application/user"
	"knowledge-graph/internal/domain/cache/cachetest"
	userDomain "knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSettingsRepo struct {
	mock.Mock
}

func (m *mockSettingsRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]userDomain.UserSetting, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]userDomain.UserSetting), args.Error(1)
}

func (m *mockSettingsRepo) FindByUserIDAndKey(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) (*userDomain.UserSetting, error) {
	args := m.Called(ctx, userID, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDomain.UserSetting), args.Error(1)
}

func (m *mockSettingsRepo) Upsert(ctx context.Context, setting userDomain.UserSetting) error {
	return m.Called(ctx, setting).Error(0)
}

func (m *mockSettingsRepo) Delete(ctx context.Context, userID uuid.UUID, key userDomain.SettingKey) error {
	return m.Called(ctx, userID, key).Error(0)
}

func setupSettingsUnitHandler(t *testing.T) (*Handler, *mockSettingsRepo) {
	gin.SetMode(gin.TestMode)
	repo := new(mockSettingsRepo)
	cacheClient := cachetest.NewFakeCacheClient()
	service := user.NewSettingsService(repo, cacheClient)
	h := NewHandler(service)
	return h, repo
}

func testUserID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000001")
}

func newSettingsContext(t *testing.T, method, path string, body []byte, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, bytes.NewBuffer(body))
	if len(body) > 0 {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	c.Set(middleware.ContextUserIDKey, testUserID())
	c.Params = params
	return c, w
}

func TestGetSetting_Success(t *testing.T) {
	h, repo := setupSettingsUnitHandler(t)
	uid := testUserID()

	setting, err := userDomain.NewUserSetting(uid, userDomain.SettingKeyPreferredLanguage, userDomain.SettingValue{Value: "en"})
	require.NoError(t, err)

	repo.On("FindByUserIDAndKey", mock.Anything, uid, userDomain.SettingKeyPreferredLanguage).Return(setting, nil)

	c, w := newSettingsContext(t, http.MethodGet, "/settings/preferred_language", nil, gin.Params{{Key: "key", Value: "preferred_language"}})
	h.GetSetting(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "en")
}

func TestGetSetting_NotFound(t *testing.T) {
	h, repo := setupSettingsUnitHandler(t)
	uid := testUserID()

	repo.On("FindByUserIDAndKey", mock.Anything, uid, userDomain.SettingKeyPreferredLanguage).Return(nil, nil)

	c, w := newSettingsContext(t, http.MethodGet, "/settings/preferred_language", nil, gin.Params{{Key: "key", Value: "preferred_language"}})
	h.GetSetting(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ru")
}

func TestUpdateSetting_Success(t *testing.T) {
	h, repo := setupSettingsUnitHandler(t)

	repo.On("Upsert", mock.Anything, mock.Anything).Return(nil)

	body, _ := json.Marshal(UpdateSettingRequest{Key: "preferred_language", Value: "en"})
	c, w := newSettingsContext(t, http.MethodPost, "/settings", body, nil)
	h.UpdateSetting(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "setting updated")
}

func TestUpdateSetting_ValidationError(t *testing.T) {
	h, _ := setupSettingsUnitHandler(t)

	body, _ := json.Marshal(map[string]interface{}{"key": "invalid_key", "value": "en"})
	c, w := newSettingsContext(t, http.MethodPost, "/settings", body, nil)
	h.UpdateSetting(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSetting_RepositoryError(t *testing.T) {
	h, repo := setupSettingsUnitHandler(t)

	repo.On("Upsert", mock.Anything, mock.Anything).Return(errors.New("db error"))

	body, _ := json.Marshal(UpdateSettingRequest{Key: "preferred_language", Value: "en"})
	c, w := newSettingsContext(t, http.MethodPost, "/settings", body, nil)
	h.UpdateSetting(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetMySettings_Success(t *testing.T) {
	h, repo := setupSettingsUnitHandler(t)
	uid := testUserID()

	setting, err := userDomain.NewUserSetting(uid, userDomain.SettingKeyPreferredLanguage, userDomain.SettingValue{Value: "en"})
	require.NoError(t, err)

	repo.On("FindByUserID", mock.Anything, uid).Return([]userDomain.UserSetting{*setting}, nil)

	c, w := newSettingsContext(t, http.MethodGet, "/settings", nil, nil)
	h.GetMySettings(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "preferred_language")
}

func TestGetMySettings_Error(t *testing.T) {
	h, repo := setupSettingsUnitHandler(t)
	uid := testUserID()

	repo.On("FindByUserID", mock.Anything, uid).Return(nil, errors.New("db error"))

	c, w := newSettingsContext(t, http.MethodGet, "/settings", nil, nil)
	h.GetMySettings(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteSetting_Success(t *testing.T) {
	h, repo := setupSettingsUnitHandler(t)
	uid := testUserID()

	repo.On("Delete", mock.Anything, uid, userDomain.SettingKeyPreferredLanguage).Return(nil)

	c, w := newSettingsContext(t, http.MethodDelete, "/settings/preferred_language", nil, gin.Params{{Key: "key", Value: "preferred_language"}})
	h.DeleteSetting(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteSetting_InvalidKey(t *testing.T) {
	h, _ := setupSettingsUnitHandler(t)

	c, w := newSettingsContext(t, http.MethodDelete, "/settings/invalid_key", nil, gin.Params{{Key: "key", Value: "invalid_key"}})
	h.DeleteSetting(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetGalacticMode_Default(t *testing.T) {
	h, repo := setupSettingsUnitHandler(t)
	uid := testUserID()

	repo.On("FindByUserIDAndKey", mock.Anything, uid, userDomain.SettingKeyGalacticMode).Return(nil, nil)

	c, w := newSettingsContext(t, http.MethodGet, "/settings/galactic-mode", nil, nil)
	h.GetGalacticMode(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "false")
}

func TestToggleGalacticMode(t *testing.T) {
	h, repo := setupSettingsUnitHandler(t)
	uid := testUserID()

	repo.On("FindByUserIDAndKey", mock.Anything, uid, userDomain.SettingKeyGalacticMode).Return(nil, nil)
	repo.On("Upsert", mock.Anything, mock.Anything).Return(nil)

	c, w := newSettingsContext(t, http.MethodPost, "/settings/galactic-mode/toggle", nil, nil)
	h.ToggleGalacticMode(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "enabled")
}

func TestGetSetting_InvalidKey_Unit(t *testing.T) {
	h, _ := setupSettingsUnitHandler(t)

	c, w := newSettingsContext(t, http.MethodGet, "/settings/invalid_key", nil, gin.Params{{Key: "key", Value: "invalid_key"}})
	h.GetSetting(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
