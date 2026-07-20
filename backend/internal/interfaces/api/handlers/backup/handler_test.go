package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTaskQueue struct {
	fail bool
}

func (m *mockTaskQueue) Enqueue(ctx context.Context, task *asynq.Task) error {
	if m.fail {
		return errors.New("enqueue failed")
	}
	return nil
}

func (m *mockTaskQueue) EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error {
	return nil
}
func (m *mockTaskQueue) EnqueueComputeEmbedding(ctx context.Context, noteID string) error { return nil }

func TestTriggerCloudBackup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("disabled", func(t *testing.T) {
		h := NewHandler(&config.Config{BackupCloudEnabled: false}, nil, nil)
		r := gin.New()
		r.POST("/backup/cloud", h.TriggerCloudBackup)

		req := httptest.NewRequest(http.MethodPost, "/backup/cloud", bytes.NewBufferString(`{"local_path":"x"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing local path", func(t *testing.T) {
		h := NewHandler(&config.Config{BackupCloudEnabled: true}, nil, nil)
		r := gin.New()
		r.POST("/backup/cloud", h.TriggerCloudBackup)

		req := httptest.NewRequest(http.MethodPost, "/backup/cloud", bytes.NewBufferString(`{}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		h := NewHandler(&config.Config{BackupCloudEnabled: true}, nil, &mockTaskQueue{})
		r := gin.New()
		r.POST("/backup/cloud", h.TriggerCloudBackup)

		body, _ := json.Marshal(map[string]string{"local_path": "/tmp/backup.sql"})
		req := httptest.NewRequest(http.MethodPost, "/backup/cloud", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Contains(t, resp["remote_key"], "backups/backup-personal-")
	})

	t.Run("enqueue failure", func(t *testing.T) {
		h := NewHandler(&config.Config{BackupCloudEnabled: true}, nil, &mockTaskQueue{fail: true})
		r := gin.New()
		r.POST("/backup/cloud", h.TriggerCloudBackup)

		body, _ := json.Marshal(map[string]string{"local_path": "/tmp/backup.sql"})
		req := httptest.NewRequest(http.MethodPost, "/backup/cloud", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("nil task queue", func(t *testing.T) {
		h := NewHandler(&config.Config{BackupCloudEnabled: true}, nil, nil)
		r := gin.New()
		r.POST("/backup/cloud", h.TriggerCloudBackup)

		body, _ := json.Marshal(map[string]string{"local_path": "/tmp/backup.sql"})
		req := httptest.NewRequest(http.MethodPost, "/backup/cloud", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGetBackupStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		BackupCloudEnabled:     true,
		BackupCloudProvider:    "yandex",
		BackupLocalPath:        "/backups",
		BackupSchedule:         "0 2 * * *",
		BackupRetentionDays:    7,
		BackupYandexFolder:     "/kg",
		BackupYandexOAuthToken: "token12345",
		BackupYandexMaxBackups: 10,
	}

	h := NewHandler(cfg, nil, nil)
	r := gin.New()
	r.GET("/backup/status", h.GetBackupStatus)

	req := httptest.NewRequest(http.MethodGet, "/backup/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.True(t, resp["cloud_enabled"].(bool))
}

func TestMaskSensitive(t *testing.T) {
	assert.Equal(t, "***", maskSensitive(""))
	assert.Equal(t, "***", maskSensitive("short"))
	assert.Equal(t, "toke****2345", maskSensitive("token12345"))
}
