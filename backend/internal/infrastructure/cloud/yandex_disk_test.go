package cloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func diskHandler(t *testing.T, content []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		href := "http://" + r.Host
		switch r.URL.Path {
		case "/v1/disk/resources":
			if r.Method == http.MethodPut {
				w.WriteHeader(http.StatusCreated)
				return
			}
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"type":"dir","name":"backups","_embedded":{"items":[{"path":"disk:/backups/backup.tar.gz","type":"file","name":"backup.tar.gz","size":42,"created":"2026-07-21T00:00:00Z","modified":"2026-07-21T00:00:00Z"}]}}`)
				return
			}
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		case "/v1/disk/resources/upload":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"href":"%s/upload","method":"PUT"}`, href)
		case "/upload":
			if r.Method != http.MethodPut {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			io.Copy(io.Discard, r.Body)
			w.Header().Set("ETag", "abc123")
			w.WriteHeader(http.StatusCreated)
		case "/v1/disk/resources/download":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"href":"%s/download","method":"GET"}`, href)
		case "/download":
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(content)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func newTestDiskService(t *testing.T, handler http.HandlerFunc) (*YandexDiskService, *httptest.Server) {
	server := httptest.NewServer(handler)
	svc, err := NewYandexDiskService(YandexDiskConfig{OAuthToken: "token"})
	require.NoError(t, err)
	svc.baseURL = server.URL + "/v1/disk"
	svc.client = server.Client()
	svc.client.Timeout = 30 * time.Second
	return svc, server
}

func TestNewYandexDiskService_MissingToken(t *testing.T) {
	_, err := NewYandexDiskService(YandexDiskConfig{})
	assert.Error(t, err)
}

func TestYandexDiskService_UploadBackup(t *testing.T) {
	content := []byte("backup data")
	dir := t.TempDir()
	localPath := filepath.Join(dir, "backup.tar.gz")
	require.NoError(t, os.WriteFile(localPath, content, 0o644))

	svc, server := newTestDiskService(t, diskHandler(t, content))
	defer server.Close()

	ctx := context.Background()
	err := svc.UploadBackup(ctx, localPath, "/backups/backup.tar.gz")
	assert.NoError(t, err)
}

func TestYandexDiskService_DownloadBackup(t *testing.T) {
	content := []byte("downloaded backup")
	dir := t.TempDir()
	localPath := filepath.Join(dir, "backup.tar.gz")

	svc, server := newTestDiskService(t, diskHandler(t, content))
	defer server.Close()

	ctx := context.Background()
	err := svc.DownloadBackup(ctx, "/backups/backup.tar.gz", localPath)
	assert.NoError(t, err)

	downloaded, err := os.ReadFile(localPath)
	require.NoError(t, err)
	assert.Equal(t, content, downloaded)
}

func TestYandexDiskService_ListBackups(t *testing.T) {
	svc, server := newTestDiskService(t, diskHandler(t, nil))
	defer server.Close()

	ctx := context.Background()
	backups, err := svc.ListBackups(ctx, "/backups/")
	require.NoError(t, err)
	assert.Equal(t, []string{"backup.tar.gz"}, backups)
}

func TestYandexDiskService_DeleteBackup(t *testing.T) {
	svc, server := newTestDiskService(t, diskHandler(t, nil))
	defer server.Close()

	ctx := context.Background()
	err := svc.DeleteBackup(ctx, "/backups/backup.tar.gz")
	assert.NoError(t, err)
}

func TestYandexDiskService_EnsureDirectory(t *testing.T) {
	svc, server := newTestDiskService(t, diskHandler(t, nil))
	defer server.Close()

	ctx := context.Background()
	err := svc.EnsureDirectory(ctx, "/backups")
	assert.NoError(t, err)
}
