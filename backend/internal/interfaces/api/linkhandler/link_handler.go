package linkhandler

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"knowledge-graph/internal/application/common"
	"knowledge-graph/internal/application/recommendation"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/queue/tasks"
	apicommon "knowledge-graph/internal/interfaces/api/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	linkRepo         link.Repository
	noteRepo         note.Repository
	taskQueue        common.TaskQueue
	affectedNotesSvc *recommendation.AffectedNotesService
	taskDelay        time.Duration
}

func New(linkRepo link.Repository, noteRepo note.Repository, taskQueue common.TaskQueue, affectedNotesSvc *recommendation.AffectedNotesService, taskDelay time.Duration) *Handler {
	return &Handler{
		linkRepo:         linkRepo,
		noteRepo:         noteRepo,
		taskQueue:        taskQueue,
		affectedNotesSvc: affectedNotesSvc,
		taskDelay:        taskDelay,
	}
}

// enqueueRecommendationTasks queues recommendation refresh tasks for affected notes
func (h *Handler) enqueueRecommendationTasks(ctx context.Context, noteID uuid.UUID) {
	if h.affectedNotesSvc == nil || h.taskQueue == nil {
		return
	}

	affected, err := h.affectedNotesSvc.GetAffectedNotes(ctx, noteID)
	if err != nil {
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

type createLinkRequest struct {
	SourceNoteID string                 `json:"source_note_id" binding:"required,uuid"`
	TargetNoteID string                 `json:"target_note_id" binding:"required,uuid"`
	LinkType     string                 `json:"link_type" binding:"required,oneof=reference dependency related custom"`
	Weight       float64                `json:"weight" binding:"omitempty,min=0,max=1"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// LinkValidationErrors defines human-readable error messages for link validation
var LinkValidationErrors = map[string]string{
	"source_note_id.required": "Source note ID is required",
	"source_note_id.uuid":     "Source note ID must be a valid UUID",
	"target_note_id.required": "Target note ID is required",
	"target_note_id.uuid":     "Target note ID must be a valid UUID",
	"link_type.required":      "Link type is required",
	"link_type.oneof":         "Link type must be one of: reference, dependency, related, custom",
	"weight.min":              "Weight must be between 0 and 1",
	"weight.max":              "Weight must be between 0 and 1",
}

func (h *Handler) Create(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errStr := err.Error()
		var details []apicommon.FieldError
		for key, msg := range LinkValidationErrors {
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

	sourceID, err := uuid.Parse(req.SourceNoteID)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("source_note_id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, req.SourceNoteID),
		})
		return
	}
	targetID, err := uuid.Parse(req.TargetNoteID)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("target_note_id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, req.TargetNoteID),
		})
		return
	}

	ctx := c.Request.Context()
	sourceNote, err := h.noteRepo.FindByID(ctx, sourceID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNote)
		return
	}
	if sourceNote == nil {
		apicommon.NotFound(c, apicommon.MsgSourceNotFound)
		return
	}

	targetNote, err := h.noteRepo.FindByID(ctx, targetID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNote)
		return
	}
	if targetNote == nil {
		apicommon.NotFound(c, apicommon.MsgTargetNotFound)
		return
	}

	linkType, err := link.NewLinkType(req.LinkType)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("link_type", apicommon.ReasonInvalidValue, err.Error(), req.LinkType),
		})
		return
	}

	weight := 1.0
	if req.Weight > 0 {
		weight = req.Weight
	}
	weightVO, err := link.NewWeight(weight)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("weight", apicommon.ReasonOutOfRange, err.Error(), weight),
		})
		return
	}

	metadata, err := link.NewMetadata(req.Metadata)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("metadata", apicommon.ReasonInvalidValue, err.Error(), req.Metadata),
		})
		return
	}

	newLink := link.NewLink(sourceID, targetID, linkType, weightVO, metadata)

	if err := h.linkRepo.Save(ctx, newLink); err != nil {
		log.Printf("[LinkHandler.Create] Failed to save link: source=%s target=%s type=%s error=%v",
			newLink.SourceNoteID(), newLink.TargetNoteID(), newLink.LinkType().String(), err)
		if errors.Is(err, postgres.ErrDuplicateLink) {
			apicommon.Conflict(c, []apicommon.FieldError{
				apicommon.NewFieldErrorFull("link", apicommon.ReasonAlreadyExists, apicommon.MsgDuplicateLink,
					map[string]string{"source_note_id": req.SourceNoteID, "target_note_id": req.TargetNoteID, "link_type": req.LinkType},
					"unique combination of source, target and type"),
			})
			return
		}
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedSaveLink)
		return
	}

	responseData := toLinkResponse(newLink)
	apicommon.JSONWithMessage(c, 201, responseData, apicommon.MsgResourceCreated)
}

func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	l, err := h.linkRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchLink)
		return
	}
	if l == nil {
		apicommon.NotFound(c, apicommon.MsgLinkNotFound)
		return
	}

	apicommon.JSON(c, 200, toLinkResponse(l))
}

func (h *Handler) GetByNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, noteIDStr),
		})
		return
	}

	ctx := c.Request.Context()
	n, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNote)
		return
	}
	if n == nil {
		apicommon.NotFound(c, apicommon.MsgNoteNotFound)
		return
	}

	outgoing, err := h.linkRepo.FindBySource(ctx, noteID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchLink)
		return
	}
	incoming, err := h.linkRepo.FindByTarget(ctx, noteID)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchLink)
		return
	}

	outgoingResp := make([]linkResponse, 0, len(outgoing))
	for _, l := range outgoing {
		outgoingResp = append(outgoingResp, toLinkResponse(l))
	}
	incomingResp := make([]linkResponse, 0, len(incoming))
	for _, l := range incoming {
		incomingResp = append(incomingResp, toLinkResponse(l))
	}

	apicommon.JSON(c, 200, gin.H{
		"outgoing": outgoingResp,
		"incoming": incomingResp,
	})
}

func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, idStr),
		})
		return
	}

	ctx := c.Request.Context()
	l, err := h.linkRepo.FindByID(ctx, id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchLink)
		return
	}
	if l == nil {
		apicommon.NotFound(c, apicommon.MsgLinkNotFound)
		return
	}

	if err := h.linkRepo.Delete(ctx, id); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedDeleteLink)
		return
	}
	apicommon.NoContent(c)
}

func (h *Handler) DeleteByNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, noteIDStr),
		})
		return
	}

	ctx := c.Request.Context()
	if err := h.linkRepo.DeleteBySource(ctx, noteID); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedDeleteLink)
		return
	}
	apicommon.NoContent(c)
}
