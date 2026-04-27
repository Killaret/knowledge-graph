package notehandler

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"knowledge-graph/internal/application/common"
	graphQueries "knowledge-graph/internal/application/queries/graph"
	"knowledge-graph/internal/application/recommendation"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/queue/tasks"
	apicommon "knowledge-graph/internal/interfaces/api/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Handler struct {
	repo               note.Repository
	taskQueue          common.TaskQueue
	suggestionsHandler *graphQueries.GetSuggestionsHandler
	affectedNotesSvc   *recommendation.AffectedNotesService
	taskDelay          time.Duration
	recRepo            *postgres.RecommendationRepository
	embeddingRepo      *postgres.EmbeddingRepository
	redis              *redis.Client
	cfg                *config.Config
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

func New(repo note.Repository, taskQueue common.TaskQueue, suggestionsHandler *graphQueries.GetSuggestionsHandler, affectedNotesSvc *recommendation.AffectedNotesService, taskDelay time.Duration, recRepo *postgres.RecommendationRepository, embeddingRepo *postgres.EmbeddingRepository, redis *redis.Client, cfg *config.Config) *Handler {
	return &Handler{
		repo:               repo,
		taskQueue:          taskQueue,
		suggestionsHandler: suggestionsHandler,
		affectedNotesSvc:   affectedNotesSvc,
		taskDelay:          taskDelay,
		recRepo:            recRepo,
		embeddingRepo:      embeddingRepo,
		redis:              redis,
		cfg:                cfg,
	}
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
		task, err := tasks.NewRefreshRecommendationsTask(nid, h.taskDelay)
		if err != nil {
			continue
		}
		if err := h.taskQueue.Enqueue(ctx, task); err != nil {
			// Log error but continue
		}
	}
}

type createNoteRequest struct {
	Title    string                 `json:"title" binding:"required,max=200"`
	Content  string                 `json:"content" binding:"max=50000"`
	Type     string                 `json:"type" binding:"omitempty,oneof=star planet comet galaxy asteroid satellite debris nebula"`
	Metadata map[string]interface{} `json:"metadata"`
}

// NoteValidationErrors defines human-readable error messages for note validation
var NoteValidationErrors = map[string]string{
	"title.required": "Title is required",
	"title.max":      "Title must not exceed 200 characters",
	"content.max":    "Content must not exceed 50000 characters",
	"type.oneof":     "Type must be one of: star, planet, comet, galaxy, asteroid, satellite, debris, nebula",
}

func (h *Handler) Create(c *gin.Context) {
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

	newNote := note.NewNote(title, content, noteType, metadata)

	if err := h.repo.Save(c.Request.Context(), newNote); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedSaveNote)
		return
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
	} else {
		log.Println("taskQueue is nil, tasks not enqueued")
	}

	// Enqueue recommendation refresh tasks for affected notes
	h.enqueueRecommendationTasks(c.Request.Context(), newNote.ID())

	responseData := gin.H{
		"id":         newNote.ID(),
		"title":      newNote.Title().String(),
		"content":    newNote.Content().String(),
		"type":       newNote.Type(),
		"metadata":   newNote.Metadata().Value(),
		"created_at": newNote.CreatedAt(),
		"updated_at": newNote.UpdatedAt(),
	}
	apicommon.JSONWithMessage(c, 201, responseData, apicommon.MsgResourceCreated)
}

type updateNoteRequest struct {
	Title    string                 `json:"title" binding:"omitempty,min=1,max=200"`
	Content  string                 `json:"content" binding:"omitempty,max=50000"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (h *Handler) Update(c *gin.Context) {
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
		apicommon.NotFound(c, "Заметка")
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

	if err := h.repo.Save(c.Request.Context(), existing); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}

	if textChanged && h.taskQueue != nil {
		noteID := existing.ID().String()
		_ = h.taskQueue.EnqueueExtractKeywords(c.Request.Context(), noteID, 10)
		_ = h.taskQueue.EnqueueComputeEmbedding(c.Request.Context(), noteID)
	}

	h.enqueueRecommendationTasks(c.Request.Context(), existing.ID())

	responseData := gin.H{
		"id":         existing.ID(),
		"title":      existing.Title().String(),
		"content":    existing.Content().String(),
		"type":       existing.Type(),
		"metadata":   existing.Metadata().Value(),
		"created_at": existing.CreatedAt(),
		"updated_at": existing.UpdatedAt(),
	}
	apicommon.JSONWithMessage(c, 200, responseData, apicommon.MsgResourceUpdated)
}

func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	existing, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch note"})
		return
	}
	if existing == nil {
		c.JSON(404, gin.H{"error": "note not found"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": "failed to delete note"})
		return
	}
	c.JSON(204, nil)
}

func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	n, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch note"})
		return
	}
	if n == nil {
		c.JSON(404, gin.H{"error": "note not found"})
		return
	}

	responseData := gin.H{
		"id":         n.ID(),
		"title":      n.Title().String(),
		"content":    n.Content().String(),
		"type":       n.Type(),
		"metadata":   n.Metadata().Value(),
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
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	// Parse limit parameter (default from config)
	limit := h.cfg.RecommendationTopN
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// 1. Try to get precomputed recommendations from database
	if h.recRepo != nil {
		recs, err := h.recRepo.Get(ctx, noteID, limit)
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

	// 2. Fallback to semantic neighbors (if enabled)
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

	// 3. Fallback to Redis cache (if enabled)
	if h.cfg.RecommendationFallbackEnabled && h.redis != nil {
		cacheKey := "recommendations:" + noteID.String()
		cached, err := h.redis.Get(ctx, cacheKey).Result()
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
	task, err := tasks.NewRefreshRecommendationsTask(noteID, delay)
	if err != nil {
		log.Printf("failed to create refresh task: %v", err)
		return
	}

	if err := h.taskQueue.Enqueue(context.Background(), task); err != nil {
		log.Printf("failed to enqueue refresh task: %v", err)
	}
}

// SearchRequest - search query parameters
type SearchRequest struct {
	Q    string `form:"q"`    // search query
	Page int    `form:"page"` // page number (default 1)
	Size int    `form:"size"` // page size (default 20)
}

// SearchResponse - search response structure
type SearchResponse struct {
	Data       []*note.Note `json:"data"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	Size       int          `json:"size"`
	TotalPages int          `json:"totalPages"`
}

// Search performs full-text search on notes
func (h *Handler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Default values
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = h.cfg.PaginationDefaultLimit
	}

	// Perform search using repository directly (for simplicity, could use service layer)
	notes, total, err := h.repo.Search(c.Request.Context(), req.Q, req.Size, (req.Page-1)*req.Size)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to search notes"})
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
	// Получаем параметры пагинации из query
	limitStr := c.DefaultQuery("limit", strconv.Itoa(h.cfg.PaginationDefaultLimit))
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	if limit <= 0 || limit > h.cfg.PaginationMaxLimit {
		limit = h.cfg.PaginationDefaultLimit
	}

	// Получаем заметки из репозитория
	notes, total, err := h.repo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch notes"})
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
