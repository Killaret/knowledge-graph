package notehandler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"knowledge-graph/internal/application/achievement"
	appcache "knowledge-graph/internal/application/cache"
	"knowledge-graph/internal/application/common"
	appevents "knowledge-graph/internal/application/events"
	importer "knowledge-graph/internal/application/import"
	graphQueries "knowledge-graph/internal/application/queries/graph"
	"knowledge-graph/internal/application/recommendation"
	"knowledge-graph/internal/config"
	dcache "knowledge-graph/internal/domain/cache"
	graphdomain "knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/domain/note"
	apicommon "knowledge-graph/internal/interfaces/api/common"
	"knowledge-graph/internal/interfaces/api/common/validation"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	repo               note.Repository
	taskQueue          common.TaskQueue
	suggestionsHandler *graphQueries.GetSuggestionsHandler
	affectedNotesSvc   *recommendation.AffectedNotesService
	taskDelay          time.Duration
	recRepo            recommendation.Repository
	embeddingRepo      recommendation.EmbeddingRepository
	cacheClient        dcache.CacheClient
	cfg                *config.Config
	graphCache         *appcache.GraphCache
	achievementService *achievement.Service
	importSvc          *importer.Service
	eventPublisher     appevents.Publisher
}

// SuggestionsResponse represents the response for recommendations
type SuggestionsResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	GeneratedAt time.Time    `json:"generated_at,omitempty"`
}

// Suggestion represents a single recommendation
type Suggestion struct {
	NoteID string  `json:"note_id"`
	Title  string  `json:"title"`
	Score  float64 `json:"score"`
}

func getUserIDString(c *gin.Context) string {
	if userID, exists := middleware.GetUserID(c); exists && userID != uuid.Nil {
		return userID.String()
	}
	return ""
}

func New(repo note.Repository, taskQueue common.TaskQueue, suggestionsHandler *graphQueries.GetSuggestionsHandler, affectedNotesSvc *recommendation.AffectedNotesService, taskDelay time.Duration, recRepo recommendation.Repository, embeddingRepo recommendation.EmbeddingRepository, cacheClient dcache.CacheClient, cfg *config.Config, graphCache *appcache.GraphCache, achievementService *achievement.Service, importSvc *importer.Service) *Handler {
	return &Handler{
		repo:               repo,
		taskQueue:          taskQueue,
		suggestionsHandler: suggestionsHandler,
		affectedNotesSvc:   affectedNotesSvc,
		taskDelay:          taskDelay,
		recRepo:            recRepo,
		embeddingRepo:      embeddingRepo,
		cacheClient:        cacheClient,
		cfg:                cfg,
		graphCache:         graphCache,
		achievementService: achievementService,
		importSvc:          importSvc,
	}
}

// SetEventPublisher sets the optional graph event publisher for cache invalidation.
func (h *Handler) SetEventPublisher(p appevents.Publisher) {
	h.eventPublisher = p
}

// enqueueRecommendationTasks queues recommendation refresh tasks for affected notes
func (h *Handler) enqueueRecommendationTasks(ctx context.Context, noteID uuid.UUID) {
	if h.affectedNotesSvc == nil || h.taskQueue == nil {
		return
	}

	affected, err := h.affectedNotesSvc.GetAffectedNotes(ctx, noteID)
	if err != nil {
		// Log error but don't fail the request
		return
	}

	for _, nid := range affected {
		if err := h.taskQueue.EnqueueRefreshRecommendations(ctx, nid, h.taskDelay); err != nil {
			// Log error but continue
			_ = err
		}
	}
}

// enqueueBackupOnNoteChange schedules a database backup after note mutations.
// Multiple changes within the unique window are deduplicated by the task queue.
func (h *Handler) enqueueBackupOnNoteChange(ctx context.Context) {
	if h.taskQueue == nil {
		return
	}
	if err := h.taskQueue.EnqueueBackupOnNoteChange(ctx); err != nil {
		log.Printf("[NoteHandler] Failed to enqueue backup on note change: %v", err)
	}
}

type createNoteRequest struct {
	Title    string                 `json:"title" binding:"required,max=200"`
	Content  string                 `json:"content" binding:"omitempty,max=50000"`
	Type     string                 `json:"type" binding:"omitempty,oneof=star planet comet nebula galaxy asteroid debris blackhole satellite dust moon technical unknown reality_rift chromatic_maw void_whisper cosmic_abomination"`
	Metadata map[string]interface{} `json:"metadata"`
}

type deleteBatchRequest struct {
	IDs []string `json:"ids" binding:"required,dive,uuid"`
}

//nolint:unused
type noteResponse struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Type     string                 `json:"type"`
	Metadata map[string]interface{} `json:"metadata"`
}

//nolint:unused
var noteValidationMessages = map[string]string{
	"title.required": "Title is required",
	"title.max":      "Title must not exceed 200 characters",
	"content.max":    "Content must not exceed 50000 characters",
	"type.oneof":     "Type must be one of: star, planet, comet, nebula, galaxy, asteroid, debris, blackhole, satellite, dust, moon, technical, unknown, reality_rift, chromatic_maw, void_whisper, cosmic_abomination",
}

// NoteValidationErrors defines human-readable error messages for note validation
var NoteValidationErrors = map[string]string{
	"title.required": "Title is required",
	"title.max":      "Title must not exceed 200 characters",
	"content.max":    "Content must not exceed 50000 characters",
	"type.oneof":     "Type must be one of: star, planet, comet, galaxy, asteroid, satellite, debris, nebula",
}

func (h *Handler) Create(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "create")

	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Provide structured validation error response
		errStr := err.Error()
		var details []apicommon.FieldError
		for key, msg := range NoteValidationErrors {
			if strings.Contains(errStr, key) {
				parts := strings.Split(key, ".")
				if len(parts) >= 2 {
					details = append(details, apicommon.NewFieldError(parts[0], apicommon.ReasonInvalidValue, msg))
				}
			}
		}
		if len(details) == 0 {
			details = append(details, apicommon.NewFieldError("request", apicommon.ReasonInvalidValue, errStr))
		}
		apicommon.BadRequest(c, details)
		return
	}

	title, err := note.NewTitle(req.Title)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("title", apicommon.ReasonInvalidValue, err.Error(), req.Title),
		})
		return
	}
	content, err := note.NewContent(req.Content)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("content", apicommon.ReasonInvalidValue, err.Error(), req.Content),
		})
		return
	}
	metadata, err := note.NewMetadata(req.Metadata)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("metadata", apicommon.ReasonInvalidValue, err.Error(), req.Metadata),
		})
		return
	}

	// Определяем тип: сначала из корня запроса, затем из metadata
	noteType := req.Type
	if noteType == "" && req.Metadata != nil {
		if t, ok := req.Metadata["type"]; ok {
			if ts, ok := t.(string); ok && ts != "" {
				noteType = ts
			}
		}
	}

	// Get authenticated user ID if available
	var newNote *note.Note
	if userID, exists := middleware.GetUserID(c); exists {
		newNote = note.NewNoteWithCreator(title, content, noteType, metadata, userID)
	} else {
		newNote = note.NewNote(title, content, noteType, metadata)
	}

	if err := h.repo.Save(c.Request.Context(), newNote); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedSaveNote)
		return
	}

	// Notify graph-service cache invalidation subscribers
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishNoteCreated(context.Background(), newNote.ID().String(), getUserIDString(c)); err != nil {
			log.Printf("[NoteHandler] Failed to publish NoteCreated event: %v", err)
		}
	}

	// Ставим задачи в очередь
	log.Printf("taskQueue is nil? %v", h.taskQueue == nil)
	if h.taskQueue != nil {
		noteID := newNote.ID().String()
		log.Printf("Enqueuing tasks for note %s", noteID)
		if err := h.taskQueue.EnqueueExtractKeywords(c.Request.Context(), noteID, 10); err != nil {
			log.Printf("Failed to enqueue extract keywords: %v", err)
		}
		if err := h.taskQueue.EnqueueComputeEmbedding(c.Request.Context(), noteID); err != nil {
			log.Printf("Failed to enqueue compute embedding: %v", err)
		}
		if err := h.taskQueue.EnqueueRecalculateLinkWeights(c.Request.Context(), newNote.ID(), h.taskDelay); err != nil {
			log.Printf("Failed to enqueue link weight recalculation: %v", err)
		}
	} else {
		log.Println("taskQueue is nil, tasks not enqueued")
	}

	// Enqueue recommendation refresh tasks for affected notes
	h.enqueueRecommendationTasks(c.Request.Context(), newNote.ID())

	// Schedule backup after note change
	h.enqueueBackupOnNoteChange(c.Request.Context())

	// Invalidate graph cache for the user
	if userID, exists := middleware.GetUserID(c); exists && h.graphCache != nil {
		if err := h.graphCache.InvalidateUserGraph(c.Request.Context(), userID.String()); err != nil {
			log.Printf("[NoteHandler] Failed to invalidate graph cache: %v", err)
		}
	}

	responseData := gin.H{
		"id":         newNote.ID(),
		"title":      newNote.Title().String(),
		"content":    newNote.Content().String(),
		"type":       newNote.Type(),
		"metadata":   newNote.Metadata().Value(),
		"is_public":  newNote.IsPublic(),
		"created_at": newNote.CreatedAt(),
		"updated_at": newNote.UpdatedAt(),
	}
	apicommon.JSONWithMessage(c, 201, responseData, apicommon.MsgResourceCreated)
}

type bookmarkletRequest struct {
	Title string `json:"title" binding:"required,max=200"`
	URL   string `json:"url"   binding:"required,url,max=2048"`
	Text  string `json:"text"  binding:"max=50000"`
	Type  string `json:"type"  binding:"omitempty,oneof=star planet comet nebula galaxy asteroid debris blackhole satellite dust moon technical unknown reality_rift chromatic_maw void_whisper cosmic_abomination"`
}

type bookmarkletResponse struct {
	NoteID string `json:"note_id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
}

const maxBookmarkletContent = 10000

// buildBookmarkletContent creates Markdown body with title, URL and selected text.
// It truncates text so the total length does not exceed the domain Content limit.
func buildBookmarkletContent(title, url, text string) string {
	prefix := fmt.Sprintf("## [%s](%s)\n\n", title, url)
	remaining := maxBookmarkletContent - len(prefix)
	if remaining < 0 {
		remaining = 0
	}
	if len(text) > remaining {
		text = text[:remaining]
	}
	return prefix + text
}

// Bookmarklet creates a note from a captured web page.
func (h *Handler) Bookmarklet(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "create")

	var req bookmarkletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errStr := err.Error()
		var details []apicommon.FieldError
		for key, msg := range NoteValidationErrors {
			if strings.Contains(errStr, key) {
				parts := strings.Split(key, ".")
				if len(parts) >= 2 {
					details = append(details, apicommon.NewFieldError(parts[0], apicommon.ReasonInvalidValue, msg))
				}
			}
		}
		if len(details) == 0 {
			details = append(details, apicommon.NewFieldError("request", apicommon.ReasonInvalidValue, errStr))
		}
		apicommon.BadRequest(c, details)
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		apicommon.Unauthorized(c)
		return
	}

	noteType := req.Type
	if noteType == "" {
		noteType = "asteroid"
	}

	content := buildBookmarkletContent(req.Title, req.URL, req.Text)

	title, err := note.NewTitle(req.Title)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("title", apicommon.ReasonInvalidValue, err.Error(), req.Title),
		})
		return
	}

	contentVO, err := note.NewContent(content)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("content", apicommon.ReasonInvalidValue, err.Error(), content),
		})
		return
	}

	metadata, err := note.NewMetadata(map[string]interface{}{
		"source_url": req.URL,
		"type":       noteType,
	})
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("metadata", apicommon.ReasonInvalidValue, err.Error(), nil),
		})
		return
	}

	newNote := note.NewNoteWithCreator(title, contentVO, noteType, metadata, userID)

	if err := h.repo.Save(c.Request.Context(), newNote); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedSaveNote)
		return
	}

	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishNoteCreated(context.Background(), newNote.ID().String(), getUserIDString(c)); err != nil {
			log.Printf("[NoteHandler] Failed to publish NoteCreated event: %v", err)
		}
	}

	if h.taskQueue != nil {
		noteID := newNote.ID().String()
		if err := h.taskQueue.EnqueueExtractKeywords(c.Request.Context(), noteID, 10); err != nil {
			log.Printf("Failed to enqueue extract keywords: %v", err)
		}
		if err := h.taskQueue.EnqueueComputeEmbedding(c.Request.Context(), noteID); err != nil {
			log.Printf("Failed to enqueue compute embedding: %v", err)
		}
		if err := h.taskQueue.EnqueueRecalculateLinkWeights(c.Request.Context(), newNote.ID(), h.taskDelay); err != nil {
			log.Printf("Failed to enqueue link weight recalculation: %v", err)
		}
	}

	h.enqueueRecommendationTasks(c.Request.Context(), newNote.ID())
	h.enqueueBackupOnNoteChange(c.Request.Context())

	if h.graphCache != nil {
		if err := h.graphCache.InvalidateUserGraph(c.Request.Context(), userID.String()); err != nil {
			log.Printf("[NoteHandler] Failed to invalidate graph cache: %v", err)
		}
	}

	c.JSON(201, bookmarkletResponse{
		NoteID: newNote.ID().String(),
		Title:  newNote.Title().String(),
		Type:   newNote.Type(),
	})
}

// --- Mass bookmark import handlers ---

type importItem struct {
	Title string `json:"title" binding:"max=200"`
	URL   string `json:"url"   binding:"required,url,max=2048"`
	Text  string `json:"text"  binding:"max=50000"`
	Type  string `json:"type"  binding:"omitempty,oneof=star planet comet nebula galaxy asteroid debris blackhole satellite dust moon technical unknown reality_rift chromatic_maw void_whisper cosmic_abomination"`
}

type importOptions struct {
	DefaultType    string `json:"default_type" binding:"omitempty,oneof=star planet comet nebula galaxy asteroid debris blackhole satellite dust moon technical unknown reality_rift chromatic_maw void_whisper cosmic_abomination"`
	ExtractContent bool   `json:"extract_content"`
}

type importBookmarksPreviewRequest struct {
	Items   []importItem  `json:"items" binding:"required,max=50,dive"`
	Options importOptions `json:"options"`
}

type importBookmarksCreateRequest struct {
	Items   []importItem  `json:"items" binding:"required,max=50,dive"`
	Options importOptions `json:"options"`
}

func resolveImportType(itemType, defaultType string) string {
	if itemType != "" {
		return itemType
	}
	if defaultType != "" {
		return defaultType
	}
	return "asteroid"
}

// ImportBookmarksPreview returns a preview with titles, normalized URLs,
// duplicates and per-item validation errors.
func (h *Handler) ImportBookmarksPreview(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "import_preview")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		apicommon.Unauthorized(c)
		return
	}

	var req importBookmarksPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("request", apicommon.ReasonInvalidValue, err.Error()),
		})
		return
	}

	if len(req.Items) == 0 {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("items", apicommon.ReasonInvalidValue, "at least one item is required"),
		})
		return
	}
	if len(req.Items) > 50 {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("items", apicommon.ReasonInvalidValue, "max 50 items per batch"),
		})
		return
	}

	if h.importSvc == nil {
		apicommon.InternalErrorWithMessage(c, "Import service is not configured")
		return
	}

	items := make([]importer.Item, len(req.Items))
	for i, it := range req.Items {
		items[i] = importer.Item{
			Title:          it.Title,
			URL:            it.URL,
			Text:           it.Text,
			Type:           resolveImportType(it.Type, req.Options.DefaultType),
			ExtractContent: req.Options.ExtractContent,
		}
	}

	preview, err := h.importSvc.Preview(c.Request.Context(), userID, items)
	if err != nil {
		apicommon.BadRequestSimple(c, err.Error())
		return
	}

	apicommon.JSON(c, 200, gin.H{"items": preview})
}

// ImportBookmarks starts an async import task and returns its task_id.
func (h *Handler) ImportBookmarks(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "import_create")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		apicommon.Unauthorized(c)
		return
	}

	var req importBookmarksCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("request", apicommon.ReasonInvalidValue, err.Error()),
		})
		return
	}

	if len(req.Items) == 0 {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("items", apicommon.ReasonInvalidValue, "at least one item is required"),
		})
		return
	}
	if len(req.Items) > 50 {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("items", apicommon.ReasonInvalidValue, "max 50 items per batch"),
		})
		return
	}

	if h.importSvc == nil {
		apicommon.InternalErrorWithMessage(c, "Import service is not configured")
		return
	}

	items := make([]importer.Item, len(req.Items))
	for i, it := range req.Items {
		items[i] = importer.Item{
			Title:          it.Title,
			URL:            it.URL,
			Text:           it.Text,
			Type:           resolveImportType(it.Type, req.Options.DefaultType),
			ExtractContent: req.Options.ExtractContent,
		}
	}

	taskID, err := h.importSvc.StartImport(c.Request.Context(), userID, items)
	if err != nil {
		apicommon.BadRequestSimple(c, err.Error())
		return
	}

	apicommon.JSON(c, 202, gin.H{"task_id": taskID})
}

// ImportBookmarksStatus returns the current status of an async import task.
func (h *Handler) ImportBookmarksStatus(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "import_status")

	_, exists := middleware.GetUserID(c)
	if !exists {
		apicommon.Unauthorized(c)
		return
	}

	taskID := c.Param("task_id")
	if _, err := uuid.Parse(taskID); err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("task_id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, taskID),
		})
		return
	}

	if h.importSvc == nil {
		apicommon.InternalErrorWithMessage(c, "Import service is not configured")
		return
	}

	status, err := h.importSvc.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, dcache.ErrCacheMiss) {
			apicommon.NotFound(c, "Import task")
			return
		}
		apicommon.InternalErrorWithMessage(c, "Failed to fetch import status")
		return
	}

	apicommon.JSON(c, 200, status)
}

type updateNoteRequest struct {
	Title    string                 `json:"title" binding:"omitempty,max=200"`
	Content  string                 `json:"content" binding:"omitempty,max=50000"`
	Type     string                 `json:"type" binding:"omitempty,oneof=star planet comet nebula galaxy asteroid debris blackhole satellite dust moon technical unknown reality_rift chromatic_maw void_whisper cosmic_abomination"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (h *Handler) Update(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "update")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	existing, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNote)
		return
	}
	if existing == nil {
		apicommon.NotFound(c, "Note")
		return
	}

	var req updateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errStr := err.Error()
		var details []apicommon.FieldError
		for key, msg := range NoteValidationErrors {
			if strings.Contains(errStr, key) {
				parts := strings.Split(key, ".")
				if len(parts) >= 2 {
					details = append(details, apicommon.NewFieldError(parts[0], apicommon.ReasonInvalidValue, msg))
				}
			}
		}
		if len(details) == 0 {
			details = append(details, apicommon.NewFieldError("request", apicommon.ReasonInvalidValue, errStr))
		}
		apicommon.BadRequest(c, details)
		return
	}

	textChanged := false

	// Update title if provided
	if req.Title != "" {
		title, err := note.NewTitle(req.Title)
		if err != nil {
			apicommon.BadRequest(c, []apicommon.FieldError{
				apicommon.NewFieldErrorWithValue("title", apicommon.ReasonInvalidValue, err.Error(), req.Title),
			})
			return
		}
		if err := existing.UpdateTitle(title); err != nil {
			apicommon.BadRequest(c, []apicommon.FieldError{
				apicommon.NewFieldErrorWithValue("title", apicommon.ReasonInvalidValue, err.Error(), req.Title),
			})
			return
		}
		textChanged = true
	}
	// Update content if provided
	if req.Content != "" {
		content, err := note.NewContent(req.Content)
		if err != nil {
			apicommon.BadRequest(c, []apicommon.FieldError{
				apicommon.NewFieldErrorWithValue("content", apicommon.ReasonInvalidValue, err.Error(), req.Content),
			})
			return
		}
		if err := existing.UpdateContent(content); err != nil {
			apicommon.BadRequest(c, []apicommon.FieldError{
				apicommon.NewFieldErrorWithValue("content", apicommon.ReasonInvalidValue, err.Error(), req.Content),
			})
			return
		}
		textChanged = true
	}
	if req.Metadata != nil {
		metadata, err := note.NewMetadata(req.Metadata)
		if err != nil {
			apicommon.BadRequest(c, []apicommon.FieldError{
				apicommon.NewFieldErrorWithValue("metadata", apicommon.ReasonInvalidValue, err.Error(), req.Metadata),
			})
			return
		}
		if err := existing.UpdateMetadata(metadata); err != nil {
			apicommon.BadRequest(c, []apicommon.FieldError{
				apicommon.NewFieldErrorWithValue("metadata", apicommon.ReasonInvalidValue, err.Error(), req.Metadata),
			})
			return
		}
	}
	if req.Type != "" {
		existing.SetType(req.Type)
	}

	if err := h.repo.Save(c.Request.Context(), existing); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}

	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishNoteUpdated(context.Background(), existing.ID().String(), getUserIDString(c)); err != nil {
			log.Printf("[NoteHandler] Failed to publish NoteUpdated event: %v", err)
		}
	}

	if textChanged && h.taskQueue != nil {
		noteID := existing.ID().String()
		_ = h.taskQueue.EnqueueExtractKeywords(c.Request.Context(), noteID, 10)
		_ = h.taskQueue.EnqueueComputeEmbedding(c.Request.Context(), noteID)
		_ = h.taskQueue.EnqueueRecalculateLinkWeights(c.Request.Context(), existing.ID(), h.taskDelay)
	}

	h.enqueueRecommendationTasks(c.Request.Context(), existing.ID())
	h.enqueueBackupOnNoteChange(c.Request.Context())

	// Invalidate graph cache for the user
	if userID, exists := middleware.GetUserID(c); exists && h.graphCache != nil {
		if err := h.graphCache.InvalidateUserGraph(c.Request.Context(), userID.String()); err != nil {
			log.Printf("[NoteHandler] Failed to invalidate graph cache: %v", err)
		}
	}

	responseData := gin.H{
		"id":         existing.ID(),
		"title":      existing.Title().String(),
		"content":    existing.Content().String(),
		"type":       existing.Type(),
		"metadata":   existing.Metadata().Value(),
		"is_public":  existing.IsPublic(),
		"created_at": existing.CreatedAt(),
		"updated_at": existing.UpdatedAt(),
	}
	apicommon.JSONWithMessage(c, 200, responseData, apicommon.MsgResourceUpdated)
}

// Publish makes a note publicly visible.
func (h *Handler) Publish(c *gin.Context) {
	h.setNotePublic(c, true)
}

// Unpublish hides a note from the public graph.
func (h *Handler) Unpublish(c *gin.Context) {
	h.setNotePublic(c, false)
}

func (h *Handler) setNotePublic(c *gin.Context, isPublic bool) {
	middleware.SetDBEntity(c, "notes")
	if isPublic {
		middleware.SetDBOperation(c, "publish")
	} else {
		middleware.SetDBOperation(c, "unpublish")
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		apicommon.Unauthorized(c)
		return
	}

	existing, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNote)
		return
	}
	if existing == nil {
		apicommon.NotFound(c, "Note")
		return
	}

	if existing.CreatorID() == nil || *existing.CreatorID() != userID {
		apicommon.Forbidden(c)
		return
	}

	existing.SetIsPublic(isPublic)
	if err := h.repo.Save(c.Request.Context(), existing); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}

	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishNoteUpdated(context.Background(), existing.ID().String(), getUserIDString(c)); err != nil {
			log.Printf("[NoteHandler] Failed to publish NoteUpdated event for publish/unpublish: %v", err)
		}
	}

	h.invalidateGraphServiceCaches(c.Request.Context(), userID.String(), id.String())

	if h.graphCache != nil {
		if err := h.graphCache.InvalidateUserGraph(c.Request.Context(), userID.String()); err != nil {
			log.Printf("[NoteHandler] Failed to invalidate graph cache: %v", err)
		}
	}

	responseData := gin.H{
		"id":         existing.ID(),
		"title":      existing.Title().String(),
		"content":    existing.Content().String(),
		"type":       existing.Type(),
		"metadata":   existing.Metadata().Value(),
		"is_public":  existing.IsPublic(),
		"created_at": existing.CreatedAt(),
		"updated_at": existing.UpdatedAt(),
	}
	apicommon.JSONWithMessage(c, 200, responseData, apicommon.MsgResourceUpdated)
}

func (h *Handler) invalidateGraphServiceCaches(ctx context.Context, userID, noteID string) {
	if h.cacheClient == nil {
		return
	}

	patterns := []string{
		fmt.Sprintf("graph-service:full:%s", userID),
		fmt.Sprintf("graph-service:delta:%s:*", userID),
		"graph-service:full:public",
		"graph-service:delta:public:*",
		"graph-service:public:*",
		fmt.Sprintf("graph-service:note:*:%s:*", noteID),
	}

	for _, pattern := range patterns {
		var cursor uint64
		for {
			keys, nextCursor, err := h.cacheClient.Scan(ctx, cursor, pattern, 100)
			if err != nil {
				log.Printf("[NoteHandler] Failed to scan cache pattern %s: %v", pattern, err)
				break
			}
			if len(keys) > 0 {
				if err := h.cacheClient.Del(ctx, keys...); err != nil {
					log.Printf("[NoteHandler] Failed to delete cache keys for pattern %s: %v", pattern, err)
				}
			}
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
	}
}

func (h *Handler) Delete(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "delete")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	existing, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNote)
		return
	}
	if existing == nil {
		apicommon.NotFound(c, "Note")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedDeleteNote)
		return
	}

	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishNoteDeleted(context.Background(), id.String(), getUserIDString(c)); err != nil {
			log.Printf("[NoteHandler] Failed to publish NoteDeleted event: %v", err)
		}
	}

	// Invalidate graph cache for the user
	if userID, exists := middleware.GetUserID(c); exists && h.graphCache != nil {
		if err := h.graphCache.InvalidateUserGraph(c.Request.Context(), userID.String()); err != nil {
			log.Printf("[NoteHandler] Failed to invalidate graph cache: %v", err)
		}
	}

	h.enqueueBackupOnNoteChange(c.Request.Context())

	apicommon.NoContent(c)
}

func (h *Handler) DeleteBatch(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "delete_batch")

	var req deleteBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("ids", apicommon.ReasonInvalidValue, "ids must be a non-empty array of valid UUIDs", req.IDs),
		})
		return
	}

	if len(req.IDs) == 0 {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("ids", apicommon.ReasonInvalidValue, "ids must be a non-empty array of valid UUIDs"),
		})
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, idStr := range req.IDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			apicommon.BadRequest(c, []apicommon.FieldError{
				apicommon.NewFieldErrorWithValue("ids", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
			})
			return
		}
		ids = append(ids, id)
	}

	if err := h.repo.DeleteBatch(c.Request.Context(), ids); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedDeleteNote)
		return
	}

	if h.eventPublisher != nil {
		userID := getUserIDString(c)
		for _, id := range ids {
			if err := h.eventPublisher.PublishNoteDeleted(context.Background(), id.String(), userID); err != nil {
				log.Printf("[NoteHandler] Failed to publish NoteDeleted event for batch: %v", err)
			}
		}
	}

	// Invalidate graph cache for the user
	if userID, exists := middleware.GetUserID(c); exists && h.graphCache != nil {
		if err := h.graphCache.InvalidateUserGraph(c.Request.Context(), userID.String()); err != nil {
			log.Printf("[NoteHandler] Failed to invalidate graph cache: %v", err)
		}
	}

	h.enqueueBackupOnNoteChange(c.Request.Context())

	apicommon.NoContent(c)
}

func (h *Handler) Restore(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "restore")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	if err := h.repo.Restore(c.Request.Context(), id); err != nil {
		if errors.Is(err, note.ErrNoteNotFound) {
			apicommon.NotFound(c, "Note")
			return
		}
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedSaveNote)
		return
	}

	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishNoteUpdated(context.Background(), id.String(), getUserIDString(c)); err != nil {
			log.Printf("[NoteHandler] Failed to publish NoteUpdated event for restore: %v", err)
		}
	}

	// Invalidate graph cache for the user
	if userID, exists := middleware.GetUserID(c); exists && h.graphCache != nil {
		if err := h.graphCache.InvalidateUserGraph(c.Request.Context(), userID.String()); err != nil {
			log.Printf("[NoteHandler] Failed to invalidate graph cache: %v", err)
		}
	}

	h.enqueueBackupOnNoteChange(c.Request.Context())

	apicommon.NoContent(c)
}

func (h *Handler) Get(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "read")

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	n, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNote)
		return
	}
	if n == nil {
		apicommon.NotFound(c, "Note")
		return
	}

	responseData := gin.H{
		"id":         n.ID(),
		"title":      n.Title().String(),
		"content":    n.Content().String(),
		"type":       n.Type(),
		"metadata":   n.Metadata().Value(),
		"is_public":  n.IsPublic(),
		"created_at": n.CreatedAt(),
		"updated_at": n.UpdatedAt(),
	}
	apicommon.JSON(c, 200, responseData)
}

// GetSuggestions returns precomputed recommendations for a note
// with fallback to semantic neighbors and Redis cache
func (h *Handler) GetSuggestions(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse note ID
	idStr := c.Param("id")
	noteID, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	// Parse limit parameter (default from config)
	limit := h.cfg.RecommendationTopN
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Pass the authenticated user ID to downstream graph-service calls.
	ctx = graphdomain.WithUserID(ctx, getUserIDString(c))

	// 1. Try to get precomputed recommendations from database
	if h.recRepo != nil {
		recs, err := h.recRepo.GetRecommendations(ctx, noteID, limit)
		if err == nil && len(recs) > 0 {
			// Check staleness by comparing recommendation timestamp with note update time
			stale := false
			note, _ := h.repo.FindByID(ctx, noteID)
			if note != nil && len(recs) > 0 {
				if recs[0].UpdatedAt.Before(note.UpdatedAt()) {
					stale = true
					// Trigger background refresh
					h.enqueueRefreshWithDelay(noteID)
				}
			}

			// Convert to response format
			suggestionsResp := SuggestionsResponse{
				Suggestions: make([]Suggestion, 0, len(recs)),
				GeneratedAt: recs[0].UpdatedAt,
			}
			for _, rec := range recs {
				suggestionsResp.Suggestions = append(suggestionsResp.Suggestions, Suggestion{
					NoteID: rec.RecommendedNoteID.String(),
					Score:  rec.Score,
				})
			}

			c.Header("X-Recommendations-Source", "table")
			if stale {
				c.Header("X-Recommendations-Stale", "true")
			}
			c.JSON(200, suggestionsResp)
			return
		}
	}

	// 2. Try live graph analytics via graph-service (with in-memory BFS fallback).
	if h.suggestionsHandler != nil {
		dtos, err := h.suggestionsHandler.Handle(ctx, graphQueries.GetSuggestionsQuery{NoteID: noteID, Limit: limit})
		if err == nil && len(dtos) > 0 {
			suggestions := make([]Suggestion, 0, len(dtos))
			for _, s := range dtos {
				suggestions = append(suggestions, Suggestion{
					NoteID: s.NoteID.String(),
					Title:  s.Title,
					Score:  s.Score,
				})
			}

			c.Header("X-Recommendations-Source", "graph-service")
			c.Header("X-Recommendations-Stale", "true")
			c.JSON(200, SuggestionsResponse{Suggestions: suggestions, GeneratedAt: time.Now()})
			return
		}
	}

	// 3. Fallback to semantic neighbors (if enabled)
	if h.cfg.RecommendationFallbackSemanticEnabled && h.embeddingRepo != nil {
		neighbors, err := h.embeddingRepo.FindSimilarNotes(ctx, noteID, limit)
		if err == nil && len(neighbors) > 0 {
			suggestions := make([]Suggestion, 0, len(neighbors))
			for _, n := range neighbors {
				suggestions = append(suggestions, Suggestion{
					NoteID: n.NoteID.String(),
					Score:  n.Score,
				})
			}

			c.Header("X-Recommendations-Source", "semantic")
			c.Header("X-Recommendations-Stale", "true")
			c.JSON(200, SuggestionsResponse{Suggestions: suggestions})
			h.enqueueRefreshWithDelay(noteID)
			return
		}
	}

	// 3. Fallback to cache (if enabled)
	if h.cfg.RecommendationFallbackEnabled && h.cacheClient != nil {
		cacheKey := "recommendations:" + noteID.String()
		cached, err := h.cacheClient.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var suggestions []Suggestion
			if err := json.Unmarshal([]byte(cached), &suggestions); err == nil {
				c.Header("X-Recommendations-Source", "redis")
				c.Header("X-Recommendations-Stale", "true")
				c.JSON(200, SuggestionsResponse{Suggestions: suggestions})
				h.enqueueRefreshWithDelay(noteID)
				return
			}
		}
	}

	// 4. Nothing available - trigger background calculation and return Accepted
	h.enqueueRefreshWithDelay(noteID)
	c.Header("X-Recommendations-Source", "empty")
	c.Header("X-Recommendations-Stale", "true")
	c.JSON(202, SuggestionsResponse{Suggestions: []Suggestion{}})
}

// enqueueRefreshWithDelay creates and enqueues a refresh task with configured delay
func (h *Handler) enqueueRefreshWithDelay(noteID uuid.UUID) {
	if h.taskQueue == nil {
		return
	}

	delay := time.Duration(h.cfg.RecommendationTaskDelaySeconds) * time.Second
	if err := h.taskQueue.EnqueueRefreshRecommendations(context.Background(), noteID, delay); err != nil {
		log.Printf("failed to enqueue refresh task: %v", err)
	}
}

// SearchRequest - search query parameters
type SearchRequest struct {
	Q    string `form:"q"`    // search query
	Page int    `form:"page"` // page number (default 1)
	Size int    `form:"size"` // page size (default 20)
}

// SearchValidationErrors defines validation error messages
var SearchValidationErrors = map[string]string{
	"q.too_long":   "Search query too long (max 200 characters)",
	"page.invalid": "Invalid page number",
	"size.invalid": "Invalid page size",
}

// SearchResponse - search response structure
type SearchResponse struct {
	Data       []*note.Note `json:"data"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	Size       int          `json:"size"`
	TotalPages int          `json:"totalPages"`
}

// validateSearchQuery validates the search query for security
func validateSearchQuery(query string) []apicommon.FieldError {
	var errors []apicommon.FieldError
	maxQueryLength := 200

	if len(query) > maxQueryLength {
		errors = append(errors, apicommon.NewFieldErrorWithValue("q", apicommon.ReasonTooLong,
			SearchValidationErrors["q.too_long"], query))
	}

	// Check for potentially dangerous characters in search query
	sanitized := validation.SanitizeString(query)
	if sanitized != query {
		errors = append(errors, apicommon.NewFieldError("q", apicommon.ReasonInvalidValue,
			"Search query contains invalid characters"))
	}

	return errors
}

// Search performs full-text search on notes
func (h *Handler) Search(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "search")

	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		apicommon.BadRequestSimple(c, err.Error())
		return
	}

	// Validate search query
	if validationErrors := validateSearchQuery(req.Q); len(validationErrors) > 0 {
		apicommon.BadRequest(c, validationErrors)
		return
	}

	// Default values
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = h.cfg.PaginationDefaultLimit
	}

	// Validate pagination limits
	if req.Size > h.cfg.PaginationMaxLimit {
		req.Size = h.cfg.PaginationMaxLimit
	}

	// Perform search using repository directly (for simplicity, could use service layer).
	// Scope by current user: uuid.Nil for anonymous means public notes only.
	userID, _ := middleware.GetUserID(c)
	notes, total, err := h.repo.Search(c.Request.Context(), userID, req.Q, req.Size, (req.Page-1)*req.Size)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedSearchNotes)
		return
	}

	// Convert domain notes to JSON-структуры (как в методе List)
	responseNotes := make([]gin.H, len(notes))
	for i, n := range notes {
		responseNotes[i] = gin.H{
			"id":         n.ID(),
			"title":      n.Title().String(),
			"content":    n.Content().String(),
			"type":       n.Type(),
			"metadata":   n.Metadata().Value(),
			"is_public":  n.IsPublic(),
			"created_at": n.CreatedAt(),
			"updated_at": n.UpdatedAt(),
		}
	}

	// Calculate total pages
	totalPages := int((total + int64(req.Size) - 1) / int64(req.Size))

	c.JSON(200, gin.H{
		"data":       responseNotes,
		"total":      total,
		"page":       req.Page,
		"size":       req.Size,
		"totalPages": totalPages,
	})
}

// List возвращает список заметок с пагинацией
func (h *Handler) List(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "list")

	// Получаем параметры пагинации из query
	limitStr := c.DefaultQuery("limit", strconv.Itoa(h.cfg.PaginationDefaultLimit))
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = h.cfg.PaginationDefaultLimit
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	if limit > h.cfg.PaginationMaxLimit {
		limit = h.cfg.PaginationMaxLimit
	}

	// Получаем заметки из репозитория с учётом текущего пользователя.
	// uuid.Nil для анонима — только публичные заметки.
	userID, _ := middleware.GetUserID(c)
	notes, total, err := h.repo.List(c.Request.Context(), userID, limit, offset)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNotes)
		return
	}

	// Преобразуем доменные модели в JSON-структуры
	responseNotes := make([]gin.H, len(notes))
	for i, n := range notes {
		responseNotes[i] = gin.H{
			"id":         n.ID(),
			"title":      n.Title().String(),
			"content":    n.Content().String(),
			"type":       n.Type(),
			"metadata":   n.Metadata().Value(),
			"is_public":  n.IsPublic(),
			"created_at": n.CreatedAt(),
			"updated_at": n.UpdatedAt(),
		}
	}

	c.JSON(200, gin.H{
		"notes":  responseNotes,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
