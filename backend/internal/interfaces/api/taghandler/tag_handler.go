package taghandler

import (
	"net/http"
	"strings"
	"time"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	apicommon "knowledge-graph/internal/interfaces/api/common"
	"knowledge-graph/internal/interfaces/api/common/validation"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler — HTTP-хендлер для работы с тегами
type Handler struct {
	tagRepo  *postgres.TagRepository
	noteRepo note.Repository
}

// New создает новый хендлер тегов
func New(tagRepo *postgres.TagRepository, noteRepo note.Repository) *Handler {
	return &Handler{
		tagRepo:  tagRepo,
		noteRepo: noteRepo,
	}
}

// TagResponse — структура ответа с тегом
type TagResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateRequest — запрос на создание тега
type CreateRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// UpdateRequest — запрос на обновление тега
type UpdateRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

// TagValidationErrors defines human-readable error messages for tag validation
var TagValidationErrors = map[string]string{
	"name.required": "Tag name is required",
	"name.min":      "Tag name must be at least 1 character",
	"name.max":      "Tag name must not exceed 50 characters",
}

// validateTagName performs comprehensive tag name validation
func validateTagName(name string) []apicommon.FieldError {
	var errors []apicommon.FieldError

	// Basic binding validation equivalent
	if strings.TrimSpace(name) == "" {
		errors = append(errors, apicommon.NewFieldError("name", apicommon.ReasonRequired, TagValidationErrors["name.required"]))
		return errors
	}

	if len(name) > 50 {
		errors = append(errors, apicommon.NewFieldErrorWithValue("name", apicommon.ReasonTooLong,
			TagValidationErrors["name.max"], name))
		return errors
	}

	// Security validation - check for safe characters
	result := validation.IsSafeTag(name, 50)
	if !result.Valid {
		errors = append(errors, apicommon.NewFieldErrorWithValue("name", apicommon.ReasonInvalidValue,
			"Tag name contains invalid characters", name))
	}

	return errors
}

// toTagResponse преобразует модель в ответ
func toTagResponse(tag *postgres.TagModel) TagResponse {
	return TagResponse{
		ID:        tag.ID.String(),
		Name:      tag.Name,
		CreatedAt: tag.CreatedAt,
	}
}

// Create создает новый тег
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errStr := err.Error()
		var details []apicommon.FieldError
		for key, msg := range TagValidationErrors {
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

	// Additional security validation
	if validationErrors := validateTagName(req.Name); len(validationErrors) > 0 {
		apicommon.BadRequest(c, validationErrors)
		return
	}

	// Sanitize the name
	sanitizedName := validation.SanitizeString(req.Name)
	sanitizedName = validation.TrimAndNormalize(sanitizedName)

	ctx := c.Request.Context()

	// Проверяем уникальность имени
	existing, err := h.tagRepo.FindByName(ctx, sanitizedName)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось проверить существование тега")
		return
	}
	if existing != nil {
		apicommon.Conflict(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("name", apicommon.ReasonAlreadyExists,
				"Тег с таким именем уже существует", sanitizedName),
		})
		return
	}

	tag := &postgres.TagModel{
		ID:   uuid.New(),
		Name: sanitizedName,
	}

	if err := h.tagRepo.Create(ctx, tag); err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось создать тег")
		return
	}

	apicommon.JSON(c, http.StatusCreated, toTagResponse(tag))
}

// List возвращает список всех тегов
func (h *Handler) List(c *gin.Context) {
	ctx := c.Request.Context()

	tags, err := h.tagRepo.FindAll(ctx)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось получить список тегов")
		return
	}

	response := make([]TagResponse, len(tags))
	for i, tag := range tags {
		response[i] = toTagResponse(tag)
	}

	apicommon.JSON(c, http.StatusOK, response)
}

// Get возвращает тег по ID
func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat,
				apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	ctx := c.Request.Context()
	tag, err := h.tagRepo.FindByID(ctx, id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось получить тег")
		return
	}
	if tag == nil {
		apicommon.NotFound(c, "Тег")
		return
	}

	apicommon.JSON(c, http.StatusOK, toTagResponse(tag))
}

// Update обновляет тег
func (h *Handler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat,
				apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errStr := err.Error()
		var details []apicommon.FieldError
		for key, msg := range TagValidationErrors {
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

	// Additional security validation
	if validationErrors := validateTagName(req.Name); len(validationErrors) > 0 {
		apicommon.BadRequest(c, validationErrors)
		return
	}

	// Sanitize the name
	sanitizedName := validation.SanitizeString(req.Name)
	sanitizedName = validation.TrimAndNormalize(sanitizedName)

	ctx := c.Request.Context()

	// Проверяем существование тега
	tag, err := h.tagRepo.FindByID(ctx, id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось получить тег")
		return
	}
	if tag == nil {
		apicommon.NotFound(c, "Тег")
		return
	}

	// Проверяем уникальность нового имени (если оно изменилось)
	if sanitizedName != tag.Name {
		existing, err := h.tagRepo.FindByName(ctx, sanitizedName)
		if err != nil {
			apicommon.InternalErrorWithMessage(c, "Не удалось проверить существование тега")
			return
		}
		if existing != nil {
			apicommon.Conflict(c, []apicommon.FieldError{
				apicommon.NewFieldErrorWithValue("name", apicommon.ReasonAlreadyExists,
					"Тег с таким именем уже существует", sanitizedName),
			})
			return
		}
	}

	tag.Name = sanitizedName
	if err := h.tagRepo.Update(ctx, tag); err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось обновить тег")
		return
	}

	apicommon.JSON(c, http.StatusOK, toTagResponse(tag))
}

// Delete удаляет тег
func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat,
				apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	ctx := c.Request.Context()

	// Проверяем существование тега
	tag, err := h.tagRepo.FindByID(ctx, id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось получить тег")
		return
	}
	if tag == nil {
		apicommon.NotFound(c, "Тег")
		return
	}

	if err := h.tagRepo.Delete(ctx, id); err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось удалить тег")
		return
	}

	apicommon.NoContent(c)
}

// AddTagToNote привязывает тег к заметке
func (h *Handler) AddTagToNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat,
				apicommon.MsgInvalidUUID, noteIDStr),
		})
		return
	}

	var req struct {
		TagID string `json:"tag_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("tag_id", apicommon.ReasonRequired, "Tag ID is required"),
		})
		return
	}

	tagID, err := uuid.Parse(req.TagID)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("tag_id", apicommon.ReasonInvalidFormat,
				apicommon.MsgInvalidUUID, req.TagID),
		})
		return
	}

	ctx := c.Request.Context()

	// Проверяем существование заметки
	note, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось проверить заметку")
		return
	}
	if note == nil {
		apicommon.NotFound(c, "Заметка")
		return
	}

	// Проверяем существование тега
	tag, err := h.tagRepo.FindByID(ctx, tagID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось проверить тег")
		return
	}
	if tag == nil {
		apicommon.NotFound(c, "Тег")
		return
	}

	// Проверяем, не привязан ли уже тег
	exists, err := h.tagRepo.IsTagAssignedToNote(ctx, noteID, tagID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось проверить привязку тега")
		return
	}
	if exists {
		apicommon.Conflict(c, []apicommon.FieldError{
			apicommon.NewFieldError("tag_id", apicommon.ReasonAlreadyExists,
				"Тег уже привязан к этой заметке"),
		})
		return
	}

	if err := h.tagRepo.AddTagToNote(ctx, noteID, tagID); err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось привязать тег")
		return
	}

	apicommon.NoContent(c)
}

// RemoveTagFromNote отвязывает тег от заметки
func (h *Handler) RemoveTagFromNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	tagIDStr := c.Param("tagId")

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat,
				apicommon.MsgInvalidUUID, noteIDStr),
		})
		return
	}

	tagID, err := uuid.Parse(tagIDStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("tagId", apicommon.ReasonInvalidFormat,
				apicommon.MsgInvalidUUID, tagIDStr),
		})
		return
	}

	ctx := c.Request.Context()

	if err := h.tagRepo.RemoveTagFromNote(ctx, noteID, tagID); err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось отвязать тег")
		return
	}

	apicommon.NoContent(c)
}

// GetTagsByNote возвращает теги заметки
func (h *Handler) GetTagsByNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat,
				apicommon.MsgInvalidUUID, noteIDStr),
		})
		return
	}

	ctx := c.Request.Context()

	// Проверяем существование заметки
	note, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось проверить заметку")
		return
	}
	if note == nil {
		apicommon.NotFound(c, "Заметка")
		return
	}

	tags, err := h.tagRepo.GetTagsByNoteID(ctx, noteID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Не удалось получить теги")
		return
	}

	response := make([]TagResponse, len(tags))
	for i, tag := range tags {
		response[i] = toTagResponse(tag)
	}

	apicommon.JSON(c, http.StatusOK, response)
}
