// Package importer implements the mass bookmark import use case.
// The directory name is "import" (the API route noun), but the package name
// cannot be the Go keyword "import", so it is named "importer".
package importer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"knowledge-graph/internal/application/common"
	dcache "knowledge-graph/internal/domain/cache"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"golang.org/x/net/html"
)

const (
	// MaxBatchSize limits the number of bookmarks accepted in one request.
	MaxBatchSize = 50

	maxContentLen = 10000

	cacheKeyPrefix = "import:task:"
	cacheTTL       = time.Hour

	statusPending    = "pending"
	statusProcessing = "processing"
	statusDone       = "done"
	statusFailed     = "failed"
)

// Item represents a single captured web page to import.
type Item struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Text  string `json:"text"`
	Type  string `json:"type"`
}

// PreviewItem extends Item with deduplication and per-item validation info.
type PreviewItem struct {
	Title          string `json:"title"`
	URL            string `json:"url"`
	Text           string `json:"text"`
	Type           string `json:"type"`
	IsNew          bool   `json:"is_new"`
	ExistingNoteID string `json:"existing_note_id,omitempty"`
	Error          string `json:"error,omitempty"`
}

// TaskStatusProgress holds the progress counters for an import task.
type TaskStatusProgress struct {
	Total     int `json:"total"`
	Processed int `json:"processed"`
	Created   int `json:"created"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

// TaskStatus represents the current state of an async import task.
type TaskStatus struct {
	TaskID   string             `json:"task_id"`
	Status   string             `json:"status"`
	Progress TaskStatusProgress `json:"progress"`
}

// Service orchestrates bookmark import: parsing, preview, deduplication,
// async task creation, and background processing.
type Service struct {
	repo      note.Repository
	cache     dcache.CacheClient
	taskQueue common.TaskQueue
}

// NewService creates a new ImportService.
func NewService(repo note.Repository, cache dcache.CacheClient, taskQueue common.TaskQueue) *Service {
	return &Service{
		repo:      repo,
		cache:     cache,
		taskQueue: taskQueue,
	}
}

// BuildContent creates Markdown body with title, URL and selected text,
// matching the bookmarklet format used by NoteHandler.
func BuildContent(title, urlStr, text string) string {
	prefix := fmt.Sprintf("## [%s](%s)\n\n", title, urlStr)
	remaining := maxContentLen - len(prefix)
	if remaining < 0 {
		remaining = 0
	}
	if len(text) > remaining {
		text = text[:remaining]
	}
	return prefix + text
}

// IsAllowedURL returns true for public http(s) URLs that are safe to fetch.
// It rejects file://, localhost, and private IPv4 ranges (10/8, 172.16/12,
// 192.168/16) and loopback addresses including 127.0.0.1.
func IsAllowedURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme == "" {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false
	}
	if host == "localhost" {
		return false
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return false
		}
	}

	return true
}

// NormalizeURL returns a canonical form of an allowed URL. It lower-cases the
// scheme and host, strips default ports and fragments, and ensures a path.
func NormalizeURL(raw string) (string, error) {
	if !IsAllowedURL(raw) {
		return "", errors.New("URL is not allowed or uses an unsupported scheme")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	u.Scheme = strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	port := u.Port()

	if (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		port = ""
	}

	if port != "" {
		u.Host = net.JoinHostPort(host, port)
	} else {
		u.Host = host
	}

	u.Fragment = ""
	u.User = nil
	return u.String(), nil
}

// ParsePlainList parses a plain-text list of bookmarks, one per line.
// Each line may contain a URL, or a URL followed by a title.
func ParsePlainList(input string) ([]Item, error) {
	var items []Item
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		if len(fields) == 1 {
			urlStr := fields[0]
			if !IsAllowedURL(urlStr) {
				continue
			}
			normalized, err := NormalizeURL(urlStr)
			if err != nil {
				continue
			}
			items = append(items, Item{Title: urlStr, URL: normalized})
			continue
		}

		urlStr := fields[0]
		if !IsAllowedURL(urlStr) {
			continue
		}
		normalized, err := NormalizeURL(urlStr)
		if err != nil {
			continue
		}
		title := strings.Join(fields[1:], " ")
		items = append(items, Item{Title: title, URL: normalized})
	}

	return items, nil
}

// ParseBookmarksHTML parses a Netscape-style bookmarks HTML file and extracts
// <a href="..."> links with their titles.
func ParseBookmarksHTML(r io.Reader) ([]Item, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var items []Item
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := ""
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = attr.Val
					break
				}
			}
			if href == "" || !IsAllowedURL(href) {
				return
			}
			normalized, err := NormalizeURL(href)
			if err != nil {
				return
			}
			title := strings.TrimSpace(extractText(n))
			if title == "" {
				title = href
			}
			items = append(items, Item{Title: title, URL: normalized})
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return items, nil
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		b.WriteString(extractText(c))
	}
	return b.String()
}

// Preview validates, normalizes and deduplicates a list of bookmarks against
// the user's existing notes. It does not persist anything.
func (s *Service) Preview(ctx context.Context, userID uuid.UUID, items []Item) ([]PreviewItem, error) {
	if len(items) > MaxBatchSize {
		return nil, fmt.Errorf("too many items: max %d", MaxBatchSize)
	}

	existing, err := s.loadExistingSourceMap(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load existing notes: %w", err)
	}

	out := make([]PreviewItem, 0, len(items))
	for _, it := range items {
		pi := PreviewItem{
			Title: it.Title,
			Text:  it.Text,
			Type:  it.Type,
			IsNew: true,
		}

		normalized, err := NormalizeURL(it.URL)
		if err != nil {
			pi.URL = it.URL
			pi.Error = err.Error()
			out = append(out, pi)
			continue
		}
		pi.URL = normalized

		if _, err := note.NewTitle(it.Title); err != nil {
			pi.Error = err.Error()
			out = append(out, pi)
			continue
		}

		contentStr := BuildContent(it.Title, normalized, it.Text)
		if _, err := note.NewContent(contentStr); err != nil {
			pi.Error = err.Error()
			out = append(out, pi)
			continue
		}

		if noteID, ok := existing[normalized]; ok {
			pi.IsNew = false
			pi.ExistingNoteID = noteID
		} else {
			pi.IsNew = true
		}

		out = append(out, pi)
	}

	return out, nil
}

// StartImport validates a batch of bookmarks, creates a task, stores its
// initial status in cache, and enqueues the background import task.
func (s *Service) StartImport(ctx context.Context, userID uuid.UUID, items []Item) (string, error) {
	if len(items) > MaxBatchSize {
		return "", fmt.Errorf("too many items: max %d", MaxBatchSize)
	}
	if s.cache == nil {
		return "", errors.New("cache is not configured")
	}

	taskID := uuid.New().String()
	status := TaskStatus{
		TaskID: taskID,
		Status: statusPending,
		Progress: TaskStatusProgress{
			Total: len(items),
		},
	}
	if err := s.storeStatus(ctx, status); err != nil {
		return "", fmt.Errorf("failed to store task status: %w", err)
	}

	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return "", fmt.Errorf("failed to marshal items: %w", err)
	}

	if s.taskQueue == nil {
		return "", errors.New("task queue is not configured")
	}

	if err := s.taskQueue.EnqueueImportBookmarks(ctx, userID, taskID, itemsJSON); err != nil {
		return "", fmt.Errorf("failed to enqueue import task: %w", err)
	}

	return taskID, nil
}

// ProcessImportTask processes a batch of bookmarks, creating notes and updating
// the task status as it progresses.
func (s *Service) ProcessImportTask(ctx context.Context, userID uuid.UUID, taskID string, items []Item) error {
	if s.cache == nil {
		return errors.New("cache is not configured")
	}
	if s.repo == nil {
		return errors.New("note repository is not configured")
	}

	status := TaskStatus{
		TaskID: taskID,
		Status: statusProcessing,
		Progress: TaskStatusProgress{
			Total: len(items),
		},
	}
	if err := s.storeStatus(ctx, status); err != nil {
		log.Printf("[ImportService] failed to store processing status: %v", err)
	}

	existing, err := s.loadExistingSourceMap(ctx, userID)
	if err != nil {
		log.Printf("[ImportService] failed to load existing notes: %v", err)
	}

	seen := make(map[string]bool)
	for _, it := range items {
		status.Progress.Processed++

		noteType := it.Type
		if noteType == "" {
			noteType = "asteroid"
		}

		normalized, err := NormalizeURL(it.URL)
		if err != nil {
			status.Progress.Failed++
			_ = s.storeStatus(ctx, status)
			continue
		}

		if seen[normalized] {
			status.Progress.Skipped++
			_ = s.storeStatus(ctx, status)
			continue
		}
		seen[normalized] = true

		if _, dup := existing[normalized]; dup {
			status.Progress.Skipped++
			_ = s.storeStatus(ctx, status)
			continue
		}

		title, err := note.NewTitle(it.Title)
		if err != nil {
			status.Progress.Failed++
			_ = s.storeStatus(ctx, status)
			continue
		}

		contentStr := BuildContent(it.Title, normalized, it.Text)
		content, err := note.NewContent(contentStr)
		if err != nil {
			status.Progress.Failed++
			_ = s.storeStatus(ctx, status)
			continue
		}

		metadata, err := note.NewMetadata(map[string]interface{}{
			"source_url": it.URL,
			"type":       noteType,
		})
		if err != nil {
			status.Progress.Failed++
			_ = s.storeStatus(ctx, status)
			continue
		}

		newNote := note.NewNoteWithCreator(title, content, noteType, metadata, userID)
		if err := s.repo.Save(ctx, newNote); err != nil {
			status.Progress.Failed++
			_ = s.storeStatus(ctx, status)
			continue
		}

		status.Progress.Created++
		existing[normalized] = newNote.ID().String()
		if err := s.storeStatus(ctx, status); err != nil {
			log.Printf("[ImportService] failed to update status for task %s: %v", taskID, err)
		}

		if s.taskQueue != nil {
			if err := s.taskQueue.EnqueueExtractKeywords(ctx, newNote.ID().String(), 10); err != nil {
				log.Printf("[ImportService] failed to enqueue extract keywords for %s: %v", newNote.ID(), err)
			}
			if err := s.taskQueue.EnqueueComputeEmbedding(ctx, newNote.ID().String()); err != nil {
				log.Printf("[ImportService] failed to enqueue compute embedding for %s: %v", newNote.ID(), err)
			}
			if err := s.taskQueue.EnqueueRecalculateLinkWeights(ctx, newNote.ID(), 0); err != nil {
				log.Printf("[ImportService] failed to enqueue link weight recalculation for %s: %v", newNote.ID(), err)
			}
		}
	}

	if status.Progress.Failed == status.Progress.Total && status.Progress.Total > 0 {
		status.Status = statusFailed
	} else {
		status.Status = statusDone
	}

	if err := s.storeStatus(ctx, status); err != nil {
		log.Printf("[ImportService] failed to store final status for task %s: %v", taskID, err)
		return err
	}

	return nil
}

// GetTaskStatus returns the current status of an import task from cache.
func (s *Service) GetTaskStatus(ctx context.Context, taskID string) (TaskStatus, error) {
	if s.cache == nil {
		return TaskStatus{}, errors.New("cache is not configured")
	}
	return s.loadStatus(ctx, taskID)
}

func (s *Service) loadExistingSourceMap(ctx context.Context, userID uuid.UUID) (map[string]string, error) {
	notes, _, err := s.repo.List(ctx, userID, 0, 0)
	if err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, n := range notes {
		meta := n.Metadata().Value()
		if meta == nil {
			continue
		}
		raw, ok := meta["source_url"].(string)
		if !ok || raw == "" {
			continue
		}
		normalized, err := NormalizeURL(raw)
		if err != nil {
			continue
		}
		m[normalized] = n.ID().String()
	}

	return m, nil
}

func (s *Service) storeStatus(ctx context.Context, status TaskStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return s.cache.Set(ctx, cacheKeyPrefix+status.TaskID, string(data), cacheTTL)
}

func (s *Service) loadStatus(ctx context.Context, taskID string) (TaskStatus, error) {
	data, err := s.cache.Get(ctx, cacheKeyPrefix+taskID)
	if err != nil {
		return TaskStatus{}, err
	}
	var status TaskStatus
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return TaskStatus{}, err
	}
	return status, nil
}
