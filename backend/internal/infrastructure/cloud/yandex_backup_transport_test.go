package cloud

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rewriteTransport struct {
	server *httptest.Server
	base   http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.server.URL, "http://")
	return t.base.RoundTrip(req)
}

func newTestYandexBackupService(t *testing.T, server *httptest.Server) *YandexBackupService {
	svc, err := NewYandexBackupService(YandexConfig{
		OAuthToken:   "test-token",
		BackupFolder: "/test-backups",
		MaxBackups:   2,
	})
	require.NoError(t, err)
	svc.client = &http.Client{Transport: &rewriteTransport{server: server, base: http.DefaultTransport}}
	return svc
}

func TestYandexBackupService_UploadAndList(t *testing.T) {
	var uploaded []string
	var deleted []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		require.True(t, strings.HasPrefix(auth, "OAuth "))

		switch r.Method {
		case "MKCOL":
			w.WriteHeader(http.StatusCreated)
		case "PUT":
			_, _ = io.Copy(io.Discard, r.Body)
			uploaded = append(uploaded, r.URL.Path)
			w.WriteHeader(http.StatusCreated)
		case "PROPFIND":
			body := `<?xml version="1.0"?>
<multistatus xmlns:D="DAV:">
  <D:response><D:href>/test-backups/file1.gz</D:href></D:response>
  <D:response><D:href>/test-backups/file2.gz</D:href></D:response>
  <D:response><D:href>/test-backups/file3.gz</D:href></D:response>
</multistatus>`
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(body))
		case "DELETE":
			deleted = append(deleted, r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	svc := newTestYandexBackupService(t, server)

	tmpFile, err := os.CreateTemp("", "test-backup-*.sql.gz")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString("backup data")
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	ctx := context.Background()
	err = svc.UploadBackup(ctx, tmpFile.Name(), "backup-2026-01-01.sql.gz")
	require.NoError(t, err)

	files, err := svc.ListBackups(ctx, "")
	require.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Subset(t, []string{"file1.gz", "file2.gz", "file3.gz"}, files)

	assert.Len(t, deleted, 1)
}

func TestYandexBackupService_DownloadBackup_Transport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			_, _ = w.Write([]byte("downloaded backup data"))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	svc := newTestYandexBackupService(t, server)

	tmpFile, err := os.CreateTemp("", "downloaded-*.gz")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	ctx := context.Background()
	err = svc.DownloadBackup(ctx, "backup-2026-01-01.sql.gz", tmpFile.Name())
	require.NoError(t, err)

	data, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)
	assert.Equal(t, "downloaded backup data", string(data))
}

func TestYandexBackupService_DeleteBackup_Transport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	svc := newTestYandexBackupService(t, server)

	err := svc.DeleteBackup(context.Background(), "backup-2026-01-01.sql.gz")
	require.NoError(t, err)
}

func TestYandexBackupService_UploadFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "MKCOL" {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	svc := newTestYandexBackupService(t, server)
	svc.maxRetries = 1

	tmpFile, err := os.CreateTemp("", "test-backup-*.sql.gz")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	_, err = tmpFile.WriteString("data")
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	err = svc.UploadBackup(context.Background(), tmpFile.Name(), "backup.gz")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload failed")
}

func TestYandexBackupService_ListBackups_Prefix(t *testing.T) {
	calledPath := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledPath = r.URL.Path
		if r.Method == "PROPFIND" {
			w.WriteHeader(http.StatusMultiStatus)
			_, _ = w.Write([]byte(`<D:multistatus xmlns:D="DAV:"><D:response><D:href>/test-backups/sub/file.gz</D:href></D:response></D:multistatus>`))
		}
	}))
	defer server.Close()

	svc := newTestYandexBackupService(t, server)
	_, _ = svc.ListBackups(context.Background(), "sub")
	assert.Equal(t, "/test-backups/sub", calledPath)
}

func TestYandexBackupService_NewConfigValidation(t *testing.T) {
	_, err := NewYandexBackupService(YandexConfig{})
	assert.Error(t, err)
}
