package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// YandexDiskService handles backup operations to Yandex.Disk via REST API
type YandexDiskService struct {
	client     *http.Client
	baseURL    string
	oauthToken string
	maxRetries int
}

// YandexDiskConfig holds Yandex.Disk configuration
type YandexDiskConfig struct {
	OAuthToken string // OAuth token for REST API
}

// UploadURLResponse represents the response from upload URL request
type UploadURLResponse struct {
	Href      string `json:"href"`
	Method    string `json:"method"`
	Templated bool   `json:"templated"`
}

// ResourceResponse represents the response from resource operations
type ResourceResponse struct {
	Path     string `json:"path"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
	Embedded struct {
		Items []ResourceItem `json:"items"`
	} `json:"_embedded"`
}

// ResourceItem represents an item in the resource list
type ResourceItem struct {
	Path     string `json:"path"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
	Size     int64  `json:"size"`
}

// ErrorResponse represents an error response from Yandex API
type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description"`
	Message     string `json:"message"`
}

// NewYandexDiskService creates a new Yandex.Disk backup service using REST API
func NewYandexDiskService(cfg YandexDiskConfig) (*YandexDiskService, error) {
	if cfg.OAuthToken == "" {
		return nil, fmt.Errorf("Yandex.Disk OAuth token is required")
	}

	return &YandexDiskService{
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
		baseURL:    "https://cloud-api.yandex.net/v1/disk",
		oauthToken: cfg.OAuthToken,
		maxRetries: 3,
	}, nil
}

// UploadBackup uploads a local backup file to Yandex.Disk via REST API
func (s *YandexDiskService) UploadBackup(ctx context.Context, localPath, remoteKey string) error {
	// Ensure directory exists
	lastSlash := strings.LastIndex(remoteKey, "/")
	var dirPath string
	if lastSlash > 0 {
		dirPath = remoteKey[:lastSlash]
	}
	if dirPath != "" && dirPath != "." {
		if err := s.EnsureDirectory(ctx, dirPath); err != nil {
			return fmt.Errorf("failed to ensure directory: %w", err)
		}
	}

	// Get upload URL from Yandex API
	uploadURL, err := s.getUploadURL(ctx, remoteKey)
	if err != nil {
		return fmt.Errorf("failed to get upload URL: %w", err)
	}

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

	// Upload with retry logic
	var lastErr error
	operation := func() error {
		// Reset file pointer to beginning
		_, err := file.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}

		// Upload to the upload URL
		req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL, file)
		if err != nil {
			return fmt.Errorf("failed to create upload request: %w", err)
		}

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

// getUploadURL gets the upload URL for a file from Yandex API
func (s *YandexDiskService) getUploadURL(ctx context.Context, remoteKey string) (string, error) {
	url := fmt.Sprintf("%s/resources/upload?path=%s&overwrite=true", s.baseURL, remoteKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+s.oauthToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get upload URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return "", fmt.Errorf("failed to get upload URL: %s - %s", errResp.Error, errResp.Description)
		}
		return "", fmt.Errorf("failed to get upload URL with status %d: %s", resp.StatusCode, string(body))
	}

	var uploadResp UploadURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", fmt.Errorf("failed to decode upload URL response: %w", err)
	}

	return uploadResp.Href, nil
}

// DownloadBackup downloads a backup file from Yandex.Disk via REST API
func (s *YandexDiskService) DownloadBackup(ctx context.Context, remoteKey, localPath string) error {
	// Get download URL from Yandex API
	downloadURL, err := s.getDownloadURL(ctx, remoteKey)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	// Download from the download URL
	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

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

// getDownloadURL gets the download URL for a file from Yandex API
func (s *YandexDiskService) getDownloadURL(ctx context.Context, remoteKey string) (string, error) {
	url := fmt.Sprintf("%s/resources/download?path=%s", s.baseURL, remoteKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+s.oauthToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return "", fmt.Errorf("failed to get download URL: %s - %s", errResp.Error, errResp.Description)
		}
		return "", fmt.Errorf("failed to get download URL with status %d: %s", resp.StatusCode, string(body))
	}

	var downloadResp UploadURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&downloadResp); err != nil {
		return "", fmt.Errorf("failed to decode download URL response: %w", err)
	}

	return downloadResp.Href, nil
}

// ListBackups lists all backups in the Yandex.Disk directory via REST API
func (s *YandexDiskService) ListBackups(ctx context.Context, prefix string) ([]string, error) {
	url := fmt.Sprintf("%s/resources?path=%s&limit=100", s.baseURL, prefix)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+s.oauthToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return nil, fmt.Errorf("failed to list: %s - %s", errResp.Error, errResp.Description)
		}
		return nil, fmt.Errorf("list failed with status %d: %s", resp.StatusCode, string(body))
	}

	var resourceResp ResourceResponse
	if err := json.NewDecoder(resp.Body).Decode(&resourceResp); err != nil {
		return nil, fmt.Errorf("failed to decode list response: %w", err)
	}

	var files []string
	for _, item := range resourceResp.Embedded.Items {
		if item.Type == "file" {
			files = append(files, item.Name)
		}
	}

	return files, nil
}

// DeleteBackup deletes a backup from Yandex.Disk via REST API
func (s *YandexDiskService) DeleteBackup(ctx context.Context, remoteKey string) error {
	url := fmt.Sprintf("%s/resources?path=%s&permanently=true", s.baseURL, remoteKey)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+s.oauthToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return fmt.Errorf("failed to delete: %s - %s", errResp.Error, errResp.Description)
		}
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// EnsureDirectory ensures the backup directory exists on Yandex.Disk via REST API
func (s *YandexDiskService) EnsureDirectory(ctx context.Context, dirPath string) error {
	url := fmt.Sprintf("%s/resources?path=%s", s.baseURL, dirPath)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "OAuth "+s.oauthToken)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	defer resp.Body.Close()

	// 409 Conflict means directory already exists
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil {
			return fmt.Errorf("failed to create directory: %s - %s", errResp.Error, errResp.Description)
		}
		return fmt.Errorf("create directory failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
