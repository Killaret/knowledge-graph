package importer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	dcache "knowledge-graph/internal/domain/cache"

	"github.com/google/uuid"
)

// SMART-CREATE-1 (decision 106): Go-side of the source-text-handler contract
// (TZ-Java-source-text-handler-2026-08-30.md, §3 and §4). Tasks and results are
// plain JSON in Redis lists — no asynq envelope, no import_tasks table
// (decision 6: status lives in the cache like the mass import does).
const (
	// DocumentQueuePending is the inbound list the Java listener reads (§3).
	DocumentQueuePending = "import:document:pending"
	// DocumentQueueResponses is the outbound list the worker drains (§4).
	DocumentQueueResponses = "import:responses:pending"
	// DocumentListenerKey is written by the Java listener every loop turn
	// (see the TZ addition in docs/tasks/SMART-CREATE-1).
	DocumentListenerKey = "source-text-handler:alive"

	documentKind     = "document"
	reasonTimeout    = "timeout"
	documentSource   = "web"
	statusCompleted  = "COMPLETED"
	statusPartial    = "PARTIAL"
	statusFailedResp = "FAILED"
)

// ErrResultTimeout signals that PopResult returned empty before its deadline;
// callers use it to drive periodic work (timeout sweep) between polls.
var ErrResultTimeout = errors.New("document result pop timed out")

// DocumentQueue is the port to the two Redis lists plus the liveness key.
// Implemented by infrastructure/queue on top of *redis.Client.
type DocumentQueue interface {
	IsListenerAlive(ctx context.Context) (bool, error)
	PushTask(ctx context.Context, payload []byte) error
	PopResult(ctx context.Context, timeout time.Duration) ([]byte, error)
}

// DocumentInput is the validated payload of POST /import/document.
type DocumentInput struct {
	Type        string // FILE, URL or TEXT (uppercase)
	Content     string // plain text or URL for TEXT/URL
	FileContent []byte // FILE only
	ContentType string // FILE only
	Filename    string // FILE only
}

// SmartWeights mirrors import_options.smart_weights in §3 of the contract.
type SmartWeights struct {
	HasHeading              int `json:"has_heading"`
	SectionBoundary         int `json:"section_boundary"`
	IntraChunkCohesion      int `json:"intra_chunk_cohesion"`
	PunctuationCompleteness int `json:"punctuation_completeness"`
	ChunkEndingWeight       int `json:"chunk_ending_weight"`
	LengthOK                int `json:"length_ok"`
	KeywordTrendStable      int `json:"keyword_trend_stable"`
	PunctuationBalance      int `json:"punctuation_balance"`
	MetadataPresent         int `json:"metadata_present"`
}

// DocumentImportOptions mirrors import_options in §3 of the contract.
type DocumentImportOptions struct {
	ChunkSize      int          `json:"chunk_size"`
	Overlap        int          `json:"overlap"`
	MinChunkLength int          `json:"min_chunk_length"`
	MaxChunkLength int          `json:"max_chunk_length"`
	MaxRetries     int          `json:"max_retries"`
	CreateLinks    bool         `json:"create_links"`
	SmartThreshold int          `json:"smart_threshold"`
	SmartWeights   SmartWeights `json:"smart_weights"`
}

// DefaultDocumentImportOptions returns the values from the §3 example — the
// contract fixes them until a future task makes them configurable.
func DefaultDocumentImportOptions() DocumentImportOptions {
	return DocumentImportOptions{
		ChunkSize:      500,
		Overlap:        50,
		MinChunkLength: 100,
		MaxChunkLength: 8000,
		MaxRetries:     3,
		CreateLinks:    false,
		SmartThreshold: 70,
		SmartWeights: SmartWeights{
			HasHeading:              25,
			SectionBoundary:         10,
			IntraChunkCohesion:      20,
			PunctuationCompleteness: 15,
			ChunkEndingWeight:       15,
			LengthOK:                10,
			KeywordTrendStable:      5,
			PunctuationBalance:      5,
			MetadataPresent:         5,
		},
	}
}

// DocumentTask is the plain-JSON job pushed to import:document:pending (§3).
type DocumentTask struct {
	EventID       string                 `json:"event_id"`
	CorrelationID string                 `json:"correlation_id"`
	UserID        string                 `json:"user_id"`
	JWT           string                 `json:"jwt"`
	Type          string                 `json:"type"`
	Content       string                 `json:"content"`
	ContentType   string                 `json:"content_type,omitempty"`
	ImportOptions DocumentImportOptions  `json:"import_options"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// DocumentResult is the plain-JSON answer read from import:responses:pending
// (§4). raw_chunks stay opaque — Go only counts them (decision 7 postpones
// their storage).
type DocumentResult struct {
	CorrelationID string            `json:"correlation_id"`
	EventID       string            `json:"event_id"`
	UserID        string            `json:"user_id"`
	Status        string            `json:"status"`
	NoteIDs       []string          `json:"note_ids"`
	RawChunks     []json.RawMessage `json:"raw_chunks"`
	Errors        []string          `json:"errors"`
}

// SetDocumentQueue wires the Redis adapter into the service. Optional: when
// unset, availability reports false and submissions are rejected.
func (s *Service) SetDocumentQueue(q DocumentQueue) {
	s.docQueue = q
}

// DocumentAvailability reports whether the Java listener is alive — the
// listener key exists in Redis (TZ addition: SET source-text-handler:alive
// every loop turn). Any error is reported as "unavailable".
func (s *Service) DocumentAvailability(ctx context.Context) (bool, error) {
	if s.docQueue == nil {
		return false, errors.New("document queue is not configured")
	}
	return s.docQueue.IsListenerAlive(ctx)
}

// SubmitDocument checks listener liveness, builds the §3 task, pushes it to
// the pending list and stores a pending TaskStatus under task_id =
// correlation_id. Returns ErrDocumentUnavailable when the listener key is
// absent — nothing is queued in that case.
var ErrDocumentUnavailable = errors.New("source_text_handler_unavailable")

func (s *Service) SubmitDocument(ctx context.Context, userID uuid.UUID, jwt string, in DocumentInput) (string, error) {
	if s.docQueue == nil || s.cache == nil {
		return "", errors.New("document queue is not configured")
	}

	alive, err := s.docQueue.IsListenerAlive(ctx)
	if err != nil {
		return "", fmt.Errorf("checking document listener: %w", err)
	}
	if !alive {
		return "", ErrDocumentUnavailable
	}

	task := DocumentTask{
		EventID:       uuid.New().String(),
		CorrelationID: uuid.New().String(),
		UserID:        userID.String(),
		JWT:           jwt,
		Type:          strings.ToUpper(in.Type),
		ImportOptions: DefaultDocumentImportOptions(),
		Metadata: map[string]interface{}{
			"source": documentSource,
		},
	}
	switch task.Type {
	case "FILE":
		task.Content = base64.StdEncoding.EncodeToString(in.FileContent)
		task.ContentType = in.ContentType
		task.Metadata["filename"] = in.Filename
	case "URL", "TEXT":
		task.Content = in.Content
	default:
		return "", fmt.Errorf("unsupported document type %q", in.Type)
	}

	payload, err := json.Marshal(task)
	if err != nil {
		return "", fmt.Errorf("failed to marshal document task: %w", err)
	}

	status := TaskStatus{
		TaskID:    task.CorrelationID,
		Status:    statusPending,
		UserID:    task.UserID,
		Kind:      documentKind,
		CreatedAt: time.Now(),
	}
	if err := s.storeStatus(ctx, status); err != nil {
		return "", fmt.Errorf("failed to store task status: %w", err)
	}
	if err := s.docQueue.PushTask(ctx, payload); err != nil {
		return "", fmt.Errorf("failed to enqueue document task: %w", err)
	}

	return task.CorrelationID, nil
}

// GetTaskStatusForUser returns the status only to its owner. Tasks carry the
// creator id since SMART-CREATE-1 (status responses include note titles, so a
// foreign task answers 404); legacy bookmark tasks without user_id keep the
// old shared behaviour.
func (s *Service) GetTaskStatusForUser(ctx context.Context, taskID string, userID uuid.UUID) (TaskStatus, error) {
	if s.cache == nil {
		return TaskStatus{}, errors.New("cache is not configured")
	}
	status, err := s.loadStatus(ctx, taskID)
	if err != nil {
		return TaskStatus{}, err
	}
	if status.UserID != "" && status.UserID != userID.String() {
		return TaskStatus{}, dcache.ErrCacheMiss
	}
	return status, nil
}

// RunDocumentResultLoop drains import:responses:pending until ctx is done.
// Between polls it fails document tasks whose answer never arrived within
// taskTimeout (TZ is silent about timeouts — the deadline is Go-side only).
func (s *Service) RunDocumentResultLoop(ctx context.Context, popTimeout, taskTimeout time.Duration) {
	if s.docQueue == nil {
		log.Println("[ImportService] document queue not configured — result reader disabled")
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		payload, err := s.docQueue.PopResult(ctx, popTimeout)
		switch {
		case errors.Is(err, ErrResultTimeout):
			s.failTimedOutTasks(ctx, taskTimeout)
		case err != nil:
			if ctx.Err() != nil {
				return
			}
			log.Printf("[ImportService] document result pop failed: %v", err)
			time.Sleep(time.Second)
		default:
			s.ApplyDocumentResult(ctx, payload)
		}
	}
}

// ApplyDocumentResult handles one §4 answer: maps the status onto the cached
// TaskStatus found by correlation_id. Unknown correlations are logged without
// the payload and dropped.
func (s *Service) ApplyDocumentResult(ctx context.Context, payload []byte) {
	var res DocumentResult
	if err := json.Unmarshal(payload, &res); err != nil {
		log.Printf("[ImportService] malformed document result dropped: %v", err)
		return
	}
	if res.CorrelationID == "" {
		log.Println("[ImportService] document result without correlation_id dropped")
		return
	}

	status, err := s.loadStatus(ctx, res.CorrelationID)
	if err != nil {
		log.Printf("[ImportService] document result for unknown task %s dropped", res.CorrelationID)
		return
	}

	switch res.Status {
	case statusCompleted:
		status.Status = statusDone
	case statusPartial:
		status.Status = statusDone
		status.RawChunksCount = len(res.RawChunks)
	case statusFailedResp:
		status.Status = statusFailed
		status.Errors = res.Errors
	default:
		log.Printf("[ImportService] document result for task %s has unknown status %q — dropped",
			res.CorrelationID, res.Status)
		return
	}

	status.Notes = s.resolveNoteRefs(ctx, res.NoteIDs)
	if err := s.storeStatus(ctx, status); err != nil {
		log.Printf("[ImportService] failed to store status for document task %s: %v", res.CorrelationID, err)
	}
}

// resolveNoteRefs turns note_ids into {id, title} pairs so the frontend does
// not have to fetch each note. Unresolvable ids keep an empty title.
func (s *Service) resolveNoteRefs(ctx context.Context, noteIDs []string) []TaskStatusNote {
	if len(noteIDs) == 0 || s.repo == nil {
		return nil
	}
	refs := make([]TaskStatusNote, 0, len(noteIDs))
	for _, raw := range noteIDs {
		ref := TaskStatusNote{ID: raw}
		if id, err := uuid.Parse(raw); err == nil {
			if n, err := s.repo.FindByID(ctx, id); err == nil && n != nil {
				ref.Title = n.Title().String()
			}
		}
		refs = append(refs, ref)
	}
	return refs
}

// failTimedOutTasks marks pending document tasks older than taskTimeout as
// failed/reason=timeout. Only kind=document tasks are swept — bookmark tasks
// own their lifecycle.
func (s *Service) failTimedOutTasks(ctx context.Context, taskTimeout time.Duration) {
	var cursor uint64
	for {
		keys, next, err := s.cache.Scan(ctx, cursor, cacheKeyPrefix+"*", 100)
		if err != nil {
			log.Printf("[ImportService] timeout sweep scan failed: %v", err)
			return
		}
		for _, key := range keys {
			status, err := s.loadStatus(ctx, strings.TrimPrefix(key, cacheKeyPrefix))
			if err != nil {
				continue
			}
			if status.Kind != documentKind || status.Status != statusPending {
				continue
			}
			if status.CreatedAt.IsZero() || time.Since(status.CreatedAt) <= taskTimeout {
				continue
			}
			status.Status = statusFailed
			status.Reason = reasonTimeout
			if err := s.storeStatus(ctx, status); err != nil {
				log.Printf("[ImportService] failed to mark document task %s timed out: %v", status.TaskID, err)
			}
		}
		if next == 0 {
			return
		}
		cursor = next
	}
}
