package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"knowledge-graph/internal/application/common"
	"knowledge-graph/internal/domain/cache"
	"knowledge-graph/internal/domain/cache/cachetest"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeNoteRepo is a minimal in-memory note.Repository for unit tests.
type fakeNoteRepo struct {
	mu    sync.RWMutex
	notes map[uuid.UUID]*note.Note
}

func newFakeNoteRepo() *fakeNoteRepo {
	return &fakeNoteRepo{notes: make(map[uuid.UUID]*note.Note)}
}

func (r *fakeNoteRepo) Save(ctx context.Context, n *note.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notes[n.ID()] = n
	return nil
}

func (r *fakeNoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.notes[id], nil
}

func (r *fakeNoteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.notes, id)
	return nil
}

func (r *fakeNoteRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range ids {
		delete(r.notes, id)
	}
	return nil
}

func (r *fakeNoteRepo) Restore(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *fakeNoteRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*note.Note
	for _, n := range r.notes {
		all = append(all, n)
	}

	total := int64(len(all))
	if offset >= len(all) {
		return nil, total, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(all) || end < offset {
		end = len(all)
	}
	return all[offset:end], total, nil
}

func (r *fakeNoteRepo) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	return r.List(ctx, userID, limit, offset)
}

func (r *fakeNoteRepo) FindAll(ctx context.Context) ([]*note.Note, error) {
	notes, _, err := r.List(ctx, uuid.Nil, 0, 0)
	return notes, err
}

func (r *fakeNoteRepo) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return r.List(ctx, userID, limit, offset)
}

// fakeTaskQueue records EnqueueImportBookmarks calls.
type fakeTaskQueue struct {
	called bool
	userID uuid.UUID
	taskID string
	items  []byte
}

func (q *fakeTaskQueue) EnqueueBackupToCloud(ctx context.Context, localPath, remoteKey, backupDate string) error {
	return nil
}

func (q *fakeTaskQueue) EnqueueRefreshRecommendations(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	return nil
}

func (q *fakeTaskQueue) EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error {
	return nil
}

func (q *fakeTaskQueue) EnqueueComputeEmbedding(ctx context.Context, noteID string) error {
	return nil
}

func (q *fakeTaskQueue) EnqueueRecalculateLinkWeights(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	return nil
}

func (q *fakeTaskQueue) EnqueueNotification(ctx context.Context, payload []byte) error {
	return nil
}

func (q *fakeTaskQueue) EnqueueImportBookmarks(ctx context.Context, userID uuid.UUID, taskID string, items []byte) error {
	q.called = true
	q.userID = userID
	q.taskID = taskID
	q.items = items
	return nil
}

// Compile-time interface checks.
var (
	_ note.Repository  = (*fakeNoteRepo)(nil)
	_ common.TaskQueue = (*fakeTaskQueue)(nil)
)

func TestBuildContent(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		url      string
		text     string
		expected string
	}{
		{
			name:     "simple",
			title:    "Example",
			url:      "https://example.com",
			text:     "some text",
			expected: "## [Example](https://example.com)\n\nsome text",
		},
		{
			name:     "truncates text",
			title:    "T",
			url:      "https://example.com",
			text:     strings.Repeat("a", maxContentLen),
			expected: "## [T](https://example.com)\n\n" + strings.Repeat("a", maxContentLen-len("## [T](https://example.com)\n\n")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildContent(tt.title, tt.url, tt.text)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestIsAllowedURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		allowed bool
	}{
		{"public https", "https://example.com/path", true},
		{"public http", "http://example.com", true},
		{"uppercase scheme", "HTTPS://Example.COM", true},
		{"localhost", "http://localhost:3000", false},
		{"127.0.0.1", "http://127.0.0.1", false},
		{"10.0.0.0/8", "http://10.0.0.1", false},
		{"172.16.0.0/12", "http://172.16.0.1", false},
		{"192.168.0.0/16", "http://192.168.1.1", false},
		{"file scheme", "file:///etc/passwd", false},
		{"no scheme", "example.com", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.allowed, IsAllowedURL(tt.url))
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"lower cases", "HTTPS://Example.COM/Path", "https://example.com/Path", false},
		{"strips http default port", "http://example.com:80/", "http://example.com/", false},
		{"strips https default port", "https://example.com:443/", "https://example.com/", false},
		{"keeps non-default port", "https://example.com:8080/", "https://example.com:8080/", false},
		{"removes fragment", "https://example.com/page#section", "https://example.com/page", false},
		{"rejects private", "http://192.168.1.1", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeURL(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, got)
		})
	}
}

func TestParsePlainList(t *testing.T) {
	input := `https://example.com
https://go.dev/doc Title for Go
# comment
not-a-url

https://site.org/page Another title here`

	items, err := ParsePlainList(input)
	require.NoError(t, err)
	require.Len(t, items, 3)

	require.Equal(t, "https://example.com", items[0].URL)
	require.Equal(t, "https://example.com", items[0].Title)

	require.Equal(t, "https://go.dev/doc", items[1].URL)
	require.Equal(t, "Title for Go", items[1].Title)

	require.Equal(t, "https://site.org/page", items[2].URL)
	require.Equal(t, "Another title here", items[2].Title)
}

func TestParseBookmarksHTML(t *testing.T) {
	html := `<!DOCTYPE netscape-bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Bookmarks</TITLE>
<H1>Bookmarks</H1>
<DL><p>
    <DT><A HREF="https://go.dev/" ADD_DATE="123456">Go</A>
    <DD>The Go programming language
    <DT><A HREF="https://pkg.go.dev/">Go Packages</A>
    <DT><A HREF="http://localhost:8080/">private</A>
</DL><p>`

	items, err := ParseBookmarksHTML(bytes.NewReader([]byte(html)))
	require.NoError(t, err)
	require.Len(t, items, 2)

	require.Equal(t, "https://go.dev/", items[0].URL)
	require.Equal(t, "Go", items[0].Title)

	require.Equal(t, "https://pkg.go.dev/", items[1].URL)
	require.Equal(t, "Go Packages", items[1].Title)
}

func TestPreview_DedupAndValidation(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newFakeNoteRepo()
	cache := cachetest.NewFakeCacheClient()
	svc := NewService(repo, cache, nil)

	// Seed an existing note with source_url.
	title, _ := note.NewTitle("Existing")
	content, _ := note.NewContent("content")
	meta, _ := note.NewMetadata(map[string]interface{}{"source_url": "https://example.com/existing"})
	existing := note.NewNoteWithCreator(title, content, "asteroid", meta, userID)
	require.NoError(t, repo.Save(ctx, existing))

	items := []Item{
		{Title: "Existing", URL: "https://example.com/existing", Type: "asteroid"},
		{Title: "New", URL: "https://example.com/new", Type: "asteroid"},
		{Title: "Bad", URL: "http://localhost", Type: "asteroid"},
		{Title: strings.Repeat("a", 201), URL: "https://example.com/long-title", Type: "asteroid"},
	}

	preview, err := svc.Preview(ctx, userID, items)
	require.NoError(t, err)
	require.Len(t, preview, 4)

	// Existing.
	require.False(t, preview[0].IsNew)
	require.Equal(t, existing.ID().String(), preview[0].ExistingNoteID)

	// New.
	require.True(t, preview[1].IsNew)
	require.Empty(t, preview[1].ExistingNoteID)
	require.Equal(t, "https://example.com/new", preview[1].URL)

	// Disallowed URL.
	require.True(t, preview[2].IsNew)
	require.NotEmpty(t, preview[2].Error)

	// Invalid title.
	require.NotEmpty(t, preview[3].Error)
}

func TestStartImport(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newFakeNoteRepo()
	cache := cachetest.NewFakeCacheClient()
	queue := &fakeTaskQueue{}
	svc := NewService(repo, cache, queue)

	items := []Item{
		{Title: "One", URL: "https://example.com/one", Type: "asteroid"},
		{Title: "Two", URL: "https://example.com/two", Type: "asteroid"},
	}

	taskID, err := svc.StartImport(ctx, userID, items)
	require.NoError(t, err)
	require.NotEmpty(t, taskID)

	// Status stored as pending.
	status, err := svc.GetTaskStatus(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, taskID, status.TaskID)
	require.Equal(t, "pending", status.Status)
	require.Equal(t, 2, status.Progress.Total)

	// Task enqueued.
	require.True(t, queue.called)
	require.Equal(t, userID, queue.userID)
	require.Equal(t, taskID, queue.taskID)
	var enqueuedItems []Item
	require.NoError(t, json.Unmarshal(queue.items, &enqueuedItems))
	require.Len(t, enqueuedItems, 2)
}

func TestStartImport_ExceedsBatchSize(t *testing.T) {
	svc := NewService(nil, cachetest.NewFakeCacheClient(), nil)
	items := make([]Item, MaxBatchSize+1)
	_, err := svc.StartImport(context.Background(), uuid.New(), items)
	require.Error(t, err)
}

func TestProcessImportTask(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newFakeNoteRepo()
	cache := cachetest.NewFakeCacheClient()
	queue := &fakeTaskQueue{}
	svc := NewService(repo, cache, queue)

	items := []Item{
		{Title: "First", URL: "https://example.com/first", Type: "asteroid"},
		{Title: "Second", URL: "https://example.com/second", Type: "planet"},
		{Title: "Duplicate", URL: "https://example.com/first", Type: "asteroid"},
		{Title: "Invalid URL", URL: "http://127.0.0.1", Type: "asteroid"},
	}

	taskID := uuid.New().String()
	err := svc.ProcessImportTask(ctx, userID, taskID, items)
	require.NoError(t, err)

	// Check final status.
	status, err := svc.GetTaskStatus(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, taskID, status.TaskID)
	require.Equal(t, "done", status.Status)
	require.Equal(t, 4, status.Progress.Total)
	require.Equal(t, 4, status.Progress.Processed)
	require.Equal(t, 2, status.Progress.Created)
	require.Equal(t, 1, status.Progress.Skipped)
	require.Equal(t, 1, status.Progress.Failed)

	// Notes saved.
	require.Len(t, repo.notes, 2)
}

func TestProcessImportTask_AllFail(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := newFakeNoteRepo()
	cache := cachetest.NewFakeCacheClient()
	svc := NewService(repo, cache, nil)

	items := []Item{
		{Title: "Bad", URL: "http://localhost", Type: "asteroid"},
	}

	taskID := uuid.New().String()
	err := svc.ProcessImportTask(ctx, userID, taskID, items)
	require.NoError(t, err)

	status, err := svc.GetTaskStatus(ctx, taskID)
	require.NoError(t, err)
	require.Equal(t, "failed", status.Status)
}

func TestGetTaskStatus_Missing(t *testing.T) {
	svc := NewService(nil, cachetest.NewFakeCacheClient(), nil)
	_, err := svc.GetTaskStatus(context.Background(), uuid.New().String())
	require.ErrorIs(t, err, cache.ErrCacheMiss)
}

func TestPreview_ExceedsBatchSize(t *testing.T) {
	svc := NewService(newFakeNoteRepo(), cachetest.NewFakeCacheClient(), nil)
	items := make([]Item, MaxBatchSize+1)
	_, err := svc.Preview(context.Background(), uuid.New(), items)
	require.Error(t, err)
}
