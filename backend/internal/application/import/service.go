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
	"sync"
	"time"
	"unicode/utf8"

	"knowledge-graph/internal/application/common"
	dcache "knowledge-graph/internal/domain/cache"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/shared/textutil"

	"github.com/google/uuid"
	"golang.org/x/net/html"
)

const (
	// MaxBatchSize limits the number of bookmarks accepted in one request.
	MaxBatchSize = 50

	// maxContentLen is the domain Content limit in runes
	// (note.NewContent rejects anything above 50 000). The URL-HEADING-1
	// stage-A body budget (IMPORT_CONTENT_MAX_RUNES, default 20 000) is
	// applied earlier, per-section, in the extractor — this cap only guards
	// the assembled note body.
	maxContentLen = 50000

	cacheKeyPrefix = "import:task:"
	cacheTTL       = time.Hour

	statusPending    = "pending"
	statusProcessing = "processing"
	statusDone       = "done"
	statusFailed     = "failed"
)

// Item represents a single captured web page to import.
type Item struct {
	Title          string `json:"title"`
	URL            string `json:"url"`
	Text           string `json:"text"`
	Type           string `json:"type"`
	ExtractContent bool   `json:"extract_content,omitempty"`
}

// PreviewItem extends Item with deduplication, per-item validation info and
// the URL-HEADING-1 stage-A extraction data: title candidates, outline,
// noise counter and title_source (see docs/tasks/URL-HEADING-1, A6).
type PreviewItem struct {
	Title           string         `json:"title"`
	URL             string         `json:"url"`
	Text            string         `json:"text"`
	Type            string         `json:"type"`
	IsNew           bool           `json:"is_new"`
	ExistingNoteID  string         `json:"existing_note_id,omitempty"`
	Error           string         `json:"error,omitempty"`
	TitleCandidates []string       `json:"title_candidates,omitempty"`
	TitleSource     string         `json:"title_source,omitempty"`
	Outline         []OutlineEntry `json:"outline,omitempty"`
	NoiseDropped    int            `json:"noise_dropped,omitempty"`
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
	extractor ContentExtractor
}

// NewService creates a new ImportService.
func NewService(repo note.Repository, cache dcache.CacheClient, taskQueue common.TaskQueue, extractor ContentExtractor) *Service {
	return &Service{
		repo:      repo,
		cache:     cache,
		taskQueue: taskQueue,
		extractor: extractor,
	}
}

// BuildContent creates Markdown body with title, URL and extracted Markdown
// sections, matching the bookmarklet format used by NoteHandler.
// It keeps the result under the domain Content limit (runes) and never splits
// a multi-byte UTF-8 rune.
func BuildContent(title, urlStr, text string) string {
	prefix := fmt.Sprintf("## [%s](%s)\n\n", title, urlStr)
	remaining := maxContentLen - utf8.RuneCountInString(prefix)
	if remaining <= 0 {
		// Extremely long title+URL combination; keep a valid, truncated prefix.
		return textutil.TruncateToMaxRunes(prefix, maxContentLen)
	}
	return prefix + textutil.TruncateToMaxRunes(text, remaining)
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

// maybeExtract fetches the page when the caller explicitly asked for content
// extraction, or when the title is missing. URL policy (scheme, host, IP) is
// enforced here, before passing the normalized URL to the infrastructure
// ContentExtractor. The returned *ExtractedPage carries the stage-A contract
// (title candidates, outline, related links, truncation counters); it is nil
// when no extraction happened.
func (s *Service) maybeExtract(ctx context.Context, it Item) (Item, *ExtractedPage, error) {
	if s.extractor == nil {
		return it, nil, nil
	}

	needsExtraction := it.ExtractContent || it.Title == ""
	if !needsExtraction {
		return it, nil, nil
	}

	if !IsAllowedURL(it.URL) {
		return it, nil, fmt.Errorf("URL is not allowed: %s", it.URL)
	}
	normalized, err := NormalizeURL(it.URL)
	if err != nil {
		return it, nil, err
	}
	it.URL = normalized

	page, err := s.extractor.Extract(ctx, it.URL)
	if err != nil {
		return it, nil, err
	}

	if it.ExtractContent {
		it.Title = page.Title
		it.Text = page.Text
	} else if it.Title == "" {
		it.Title = page.Title
		if it.Text == "" {
			it.Text = page.Text
		}
	}
	return it, page, nil
}

// previewConcurrency bounds parallel page fetches during preview so a 50-item
// batch finishes within the client's request timeout instead of running
// sequentially at up to the per-fetch timeout each.
const previewConcurrency = 10

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

	out := make([]PreviewItem, len(items))
	sem := make(chan struct{}, previewConcurrency)
	var wg sync.WaitGroup
	for i, it := range items {
		wg.Add(1)
		go func(i int, it Item) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			out[i] = s.previewItem(ctx, it, existing)
		}(i, it)
	}
	wg.Wait()

	return out, nil
}

func (s *Service) previewItem(ctx context.Context, it Item, existing map[string]string) PreviewItem {
	it, page, err := s.maybeExtract(ctx, it)
	if err != nil {
		return PreviewItem{
			Title: it.Title,
			URL:   it.URL,
			Text:  it.Text,
			Type:  it.Type,
			Error: err.Error(),
		}
	}

	pi := PreviewItem{
		Title: it.Title,
		Text:  it.Text,
		Type:  it.Type,
		IsNew: true,
	}
	if page != nil {
		pi.TitleCandidates = page.TitleCandidates
		pi.TitleSource = page.TitleSource
		pi.Outline = page.Outline
		pi.NoiseDropped = page.NoiseDropped
	}

	normalized, err := NormalizeURL(it.URL)
	if err != nil {
		pi.URL = it.URL
		pi.Error = err.Error()
		return pi
	}
	pi.URL = normalized

	if _, err := note.NewTitle(it.Title); err != nil {
		pi.Error = err.Error()
		return pi
	}

	contentStr := BuildContent(it.Title, normalized, it.Text)
	if _, err := note.NewContent(contentStr); err != nil {
		pi.Error = err.Error()
		return pi
	}

	if noteID, ok := existing[normalized]; ok {
		pi.IsNew = false
		pi.ExistingNoteID = noteID
	} else {
		pi.IsNew = true
	}

	return pi
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

		extracted, page, err := s.maybeExtract(ctx, it)
		if err != nil {
			log.Printf("[ImportService] content extraction failed for %s: %v", it.URL, err)
		} else {
			it = extracted
		}

		noteType, err := note.NewType(it.Type)
		if err != nil {
			noteType = note.MustType("asteroid")
		}

		normalized, err := NormalizeURL(it.URL)
		if err != nil {
			status.Progress.Failed++
			_ = s.storeStatus(ctx, status)
			continue
		}

		if it.Title == "" {
			it.Title = normalized
			if len(it.Title) > 200 {
				it.Title = it.Title[:200]
			}
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

		metaMap := map[string]interface{}{
			"source_url": it.URL,
			"type":       noteType,
		}
		// URL-HEADING-1 stage-A extraction metadata (see task file, A5/A6 and
		// question 3): candidates and related links live in metadata, never in
		// the note body or domain model.
		if page != nil {
			metaMap["title_candidates"] = page.TitleCandidates
			metaMap["title_source"] = page.TitleSource
			if len(page.RelatedLinks) > 0 {
				links := make([]map[string]string, 0, len(page.RelatedLinks))
				for _, l := range page.RelatedLinks {
					links = append(links, map[string]string{"text": l.Text, "url": l.URL})
				}
				metaMap["related_links"] = links
			}
			if page.SectionsDropped > 0 {
				metaMap["import_truncated"] = map[string]interface{}{
					"sections_dropped": page.SectionsDropped,
				}
			}
		}
		metadata, err := note.NewMetadata(metaMap)
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
			if err := s.taskQueue.EnqueueNormalizeNote(ctx, newNote.ID().String()); err != nil {
				log.Printf("[ImportService] failed to enqueue normalize note for %s: %v", newNote.ID(), err)
			}
			if err := s.taskQueue.EnqueueRecalculateLinkWeights(ctx, newNote.ID(), 0); err != nil {
				log.Printf("[ImportService] failed to enqueue link weight recalculation for %s: %v", newNote.ID(), err)
			}
		}
	}

	if s.taskQueue != nil {
		if err := s.taskQueue.EnqueueBackupOnNoteChange(ctx); err != nil {
			log.Printf("[ImportService] failed to enqueue backup on note change: %v", err)
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
