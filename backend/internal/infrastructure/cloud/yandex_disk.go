package cloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// YandexDiskService handles backup operations to Yandex.Disk via WebDAV
type YandexDiskService struct {
	client     *http.Client
	baseURL    string
	username   string
	password   string
	maxRetries int
}

// YandexDiskConfig holds Yandex.Disk configuration
type YandexDiskConfig struct {
	Username string // OAuth token or username
	Password string // OAuth token (if using OAuth) or password
	// For OAuth: use token as both username and password
}

// NewYandexDiskService creates a new Yandex.Disk backup service
func NewYandexDiskService(cfg YandexDiskConfig) (*YandexDiskService, error) {
	if cfg.Username == "" || cfg.Password == "" {
		return nil, fmt.Errorf("Yandex.Disk configuration is incomplete")
	}

	return &YandexDiskService{
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
		baseURL:    "https://webdav.yandex.ru",
		username:   cfg.Username,
		password:   cfg.Password,
		maxRetries: 3,
	}, nil
}

// UploadBackup uploads a local backup file to Yandex.Disk with retry logic
func (s *YandexDiskService) UploadBackup(ctx context.Context, localPath, remoteKey string) error {
	// Open the local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Get file info for content length
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Create upload with retry logic
	var lastErr error
	operation := func() error {
		// Reset file pointer to beginning
		_, err := file.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}

		// Upload to Yandex.Disk via WebDAV PUT
		url := s.baseURL + "/" + strings.TrimPrefix(remoteKey, "/")
		req, err := http.NewRequestWithContext(ctx, "PUT", url, file)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.SetBasicAuth(s.username, s.password)
		req.ContentLength = fileInfo.Size()

		resp, err := s.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to upload: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
		}

		return nil
	}

	// Configure exponential backoff
	backoffCfg := backoff.NewExponentialBackOff()
	backoffCfg.InitialInterval = 1 * time.Second
	backoffCfg.MaxInterval = 30 * time.Second
	backoffCfg.MaxElapsedTime = 2 * time.Minute
	backoffCfg.Multiplier = 2.0

	// Execute with retry
	err = backoff.Retry(operation, backoff.WithMaxRetries(backoffCfg, uint64(s.maxRetries)))
	if err != nil {
		lastErr = err
		return fmt.Errorf("upload failed after %d retries: %w", s.maxRetries, lastErr)
	}

	return nil
}

// DownloadBackup downloads a backup file from Yandex.Disk
func (s *YandexDiskService) DownloadBackup(ctx context.Context, remoteKey, localPath string) error {
	url := s.baseURL + "/" + strings.TrimPrefix(remoteKey, "/")
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(s.username, s.password)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file content: %w", err)
	}

	return nil
}

// ListBackups lists all backups in the Yandex.Disk directory
func (s *YandexDiskService) ListBackups(ctx context.Context, prefix string) ([]string, error) {
	url := s.baseURL + "/" + strings.TrimPrefix(prefix, "/")
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Depth", "1")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse WebDAV response to extract file names
	// For simplicity, we'll return the prefix itself
	// A full implementation would parse XML response
	return []string{prefix}, nil
}

// DeleteBackup deletes a backup from Yandex.Disk
func (s *YandexDiskService) DeleteBackup(ctx context.Context, remoteKey string) error {
	url := s.baseURL + "/" + strings.TrimPrefix(remoteKey, "/")
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(s.username, s.password)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// EnsureDirectory ensures the backup directory exists on Yandex.Disk
func (s *YandexDiskService) EnsureDirectory(ctx context.Context, dirPath string) error {
	url := s.baseURL + "/" + strings.TrimPrefix(dirPath, "/")
	req, err := http.NewRequestWithContext(ctx, "MKCOL", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(s.username, s.password)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	defer resp.Body.Close()

	// 405 Method Not Allowed means directory already exists
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusMethodNotAllowed {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create directory failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
