package linkhandler

import (
	"context"
	"time"

	"knowledge-graph/internal/application/common"
	"knowledge-graph/internal/application/recommendation"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/queue/tasks"

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

type createLinkRequest struct {
	SourceNoteID string                 `json:"source_note_id" binding:"required"`
	TargetNoteID string                 `json:"target_note_id" binding:"required"`
	LinkType     string                 `json:"link_type" binding:"required"`
	Weight       float64                `json:"weight"`
	Metadata     map[string]interface{} `json:"metadata"`
}

func (h *Handler) Create(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	sourceID, err := uuid.Parse(req.SourceNoteID)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid source_note_id"})
		return
	}
	targetID, err := uuid.Parse(req.TargetNoteID)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid target_note_id"})
		return
	}

	ctx := c.Request.Context()
	sourceNote, err := h.noteRepo.FindByID(ctx, sourceID)
	if err != nil || sourceNote == nil {
		c.JSON(404, gin.H{"error": "source note not found"})
		return
	}
	targetNote, err := h.noteRepo.FindByID(ctx, targetID)
	if err != nil || targetNote == nil {
		c.JSON(404, gin.H{"error": "target note not found"})
		return
	}

	linkType, err := link.NewLinkType(req.LinkType)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	weight := 1.0
	if req.Weight > 0 {
		weight = req.Weight
	}
	weightVO, err := link.NewWeight(weight)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	metadata, err := link.NewMetadata(req.Metadata)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	newLink := link.NewLink(sourceID, targetID, linkType, weightVO, metadata)

	if err := h.linkRepo.Save(ctx, newLink); err != nil {
		c.JSON(500, gin.H{"error": "failed to save link"})
		return
	}

	c.JSON(201, gin.H{
		"id":             newLink.ID(),
		"source_note_id": newLink.SourceNoteID(),
		"target_note_id": newLink.TargetNoteID(),
		"link_type":      newLink.LinkType().String(),
		"weight":         newLink.Weight().Value(),
		"metadata":       newLink.Metadata().Value(),
		"created_at":     newLink.CreatedAt(),
	})
}

func (h *Handler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	l, err := h.linkRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch link"})
		return
	}
	if l == nil {
		c.JSON(404, gin.H{"error": "link not found"})
		return
	}

	c.JSON(200, gin.H{
		"id":             l.ID(),
		"source_note_id": l.SourceNoteID(),
		"target_note_id": l.TargetNoteID(),
		"link_type":      l.LinkType().String(),
		"weight":         l.Weight().Value(),
		"metadata":       l.Metadata().Value(),
		"created_at":     l.CreatedAt(),
	})
}

func (h *Handler) GetByNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid note id"})
		return
	}

	ctx := c.Request.Context()
	n, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil || n == nil {
		c.JSON(404, gin.H{"error": "note not found"})
		return
	}

	outgoing, err := h.linkRepo.FindBySource(ctx, noteID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch outgoing links"})
		return
	}
	incoming, err := h.linkRepo.FindByTarget(ctx, noteID)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch incoming links"})
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

	c.JSON(200, gin.H{
		"outgoing": outgoingResp,
		"incoming": incomingResp,
	})
}

func (h *Handler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	l, err := h.linkRepo.FindByID(ctx, id)
	if err != nil || l == nil {
		c.JSON(404, gin.H{"error": "link not found"})
		return
	}

	if err := h.linkRepo.Delete(ctx, id); err != nil {
		c.JSON(500, gin.H{"error": "failed to delete link"})
		return
	}
	c.JSON(204, nil)
}

func (h *Handler) DeleteByNote(c *gin.Context) {
	noteIDStr := c.Param("id")
	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid note id"})
		return
	}

	ctx := c.Request.Context()
	if err := h.linkRepo.DeleteBySource(ctx, noteID); err != nil {
		c.JSON(500, gin.H{"error": "failed to delete links for note"})
		return
	}
	c.JSON(204, nil)
}
