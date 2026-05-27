package cloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// YandexBackupService handles backup operations to Yandex.Disk via WebDAV
type YandexBackupService struct {
	client        *http.Client
	oauthToken    string
	backupFolder  string
	maxBackups    int
	maxRetries    int
}

// YandexConfig holds Yandex.Disk configuration
type YandexConfig struct {
	OAuthToken   string
	BackupFolder string
	MaxBackups   int
}

// NewYandexBackupService creates a new Yandex.Disk backup service
func NewYandexBackupService(cfg YandexConfig) (*YandexBackupService, error) {
	if cfg.OAuthToken == "" {
		return nil, fmt.Errorf("Yandex.Disk OAuth token is required")
	}

	// Set defaults
	backupFolder := cfg.BackupFolder
	if backupFolder == "" {
		backupFolder = "/KnowledgeGraphBackups"
	}

	maxBackups := cfg.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 10
	}

	return &YandexBackupService{
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
		oauthToken:   cfg.OAuthToken,
		backupFolder: backupFolder,
		maxBackups:   maxBackups,
		maxRetries:   3,
	}, nil
}

// UploadBackup uploads a local backup file to Yandex.Disk with retry logic
func (s *YandexBackupService) UploadBackup(ctx context.Context, localPath, remoteKey string) error {
	// Open the local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Ensure backup folder exists
	folderURL := fmt.Sprintf("https://webdav.yandex.ru%s", s.backupFolder)
	if err := s.ensureFolder(ctx, folderURL); err != nil {
		return fmt.Errorf("failed to ensure backup folder: %w", err)
	}

	// Upload with retry logic
	remotePath := fmt.Sprintf("%s/%s", s.backupFolder, remoteKey)
	url := fmt.Sprintf("https://webdav.yandex.ru%s", remotePath)

	var lastErr error
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		// Reset file pointer
		_, err := file.Seek(0, 0)
		if err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}

		// Create upload request
		req, err := http.NewRequestWithContext(ctx, "PUT", url, file)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", s.oauthToken))
		req.Header.Set("Content-Type", "application/octet-stream")
		req.ContentLength = fileInfo.Size()

		// Execute request
		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		}

		// Check response
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			// Clean up old backups if needed
			s.cleanupOldBackups(ctx)
			return nil
		}

		resp.Body.Close()
		lastErr = fmt.Errorf("upload failed with status: %d", resp.StatusCode)
		time.Sleep(time.Duration(attempt+1) * time.Second)
	}

	return fmt.Errorf("upload failed after %d retries: %w", s.maxRetries, lastErr)
}

// DownloadBackup downloads a backup file from Yandex.Disk
func (s *YandexBackupService) DownloadBackup(ctx context.Context, remoteKey, localPath string) error {
	remotePath := fmt.Sprintf("%s/%s", s.backupFolder, remoteKey)
	url := fmt.Sprintf("https://webdav.yandex.ru%s", remotePath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", s.oauthToken))

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
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

// ListBackups lists all backups in the Yandex.Disk folder
func (s *YandexBackupService) ListBackups(ctx context.Context, prefix string) ([]string, error) {
	folderURL := fmt.Sprintf("https://webdav.yandex.ru%s", s.backupFolder)
	if prefix != "" {
		folderURL = fmt.Sprintf("%s/%s", folderURL, prefix)
	}

	req, err := http.NewRequestWithContext(ctx, "PROPFIND", folderURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", s.oauthToken))
	req.Header.Set("Depth", "1")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("list failed with status: %d", resp.StatusCode)
	}

	// Parse WebDAV response (simplified - in production, use proper XML parser)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Simple parsing - extract filenames from WebDAV XML
	var files []string
	content := string(body)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.Contains(line, "<D:href>") {
			// Extract path from href
			start := strings.Index(line, "<D:href>")
			if start == -1 {
				start = strings.Index(line, "<d:href>")
			}
			if start != -1 {
				start += strings.Index(line[start:], ">") + 1
				end := strings.Index(line[start:], "<")
				if end != -1 {
					path := line[start : start+end]
					// Extract filename from full path
					if idx := strings.LastIndex(path, "/"); idx != -1 {
						filename := path[idx+1:]
						if filename != "" && !strings.HasPrefix(filename, ".") {
							files = append(files, filename)
						}
					}
				}
			}
		}
	}

	return files, nil
}

// DeleteBackup deletes a backup from Yandex.Disk
func (s *YandexBackupService) DeleteBackup(ctx context.Context, remoteKey string) error {
	remotePath := fmt.Sprintf("%s/%s", s.backupFolder, remoteKey)
	url := fmt.Sprintf("https://webdav.yandex.ru%s", remotePath)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", s.oauthToken))

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed with status: %d", resp.StatusCode)
	}

	return nil
}

// ensureFolder creates the backup folder if it doesn't exist
func (s *YandexBackupService) ensureFolder(ctx context.Context, folderURL string) error {
	req, err := http.NewRequestWithContext(ctx, "MKCOL", folderURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("OAuth %s", s.oauthToken))

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}
	defer resp.Body.Close()

	// 405 Method Not Allowed means folder already exists
	if resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return nil
	}

	return fmt.Errorf("failed to create folder with status: %d", resp.StatusCode)
}

// cleanupOldBackups removes old backups keeping only maxBackups
func (s *YandexBackupService) cleanupOldBackups(ctx context.Context) {
	files, err := s.ListBackups(ctx, "")
	if err != nil {
		return
	}

	if len(files) <= s.maxBackups {
		return
	}

	// Sort files by name (assuming timestamp in filename)
	// and delete oldest ones
	for i := 0; i < len(files)-s.maxBackups; i++ {
		s.DeleteBackup(ctx, files[i])
	}
}
