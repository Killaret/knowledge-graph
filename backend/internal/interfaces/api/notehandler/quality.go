package notehandler

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appquality "knowledge-graph/internal/application/quality"
)

// QualityReader fetches NOTE-QUALITY-1 log entries for a note — the log is
// authoritative because it exists even without an NLP-4 artifact.
type QualityReader interface {
	LatestQuality(ctx context.Context, noteID uuid.UUID) (*appquality.LogEntry, error)
	ListQuality(ctx context.Context, noteID uuid.UUID, limit int) ([]appquality.LogEntry, error)
}

// QualityEnqueuer schedules an assessment — the manual "Доработать" action.
type QualityEnqueuer interface {
	EnqueueAssessQuality(ctx context.Context, noteID string, trigger string) error
}

// SetQuality installs the NOTE-QUALITY-1 dependencies. Called only when the
// feature is enabled and Mongo is connected; otherwise both endpoints keep
// answering {"enabled": false}.
func (h *Handler) SetQuality(enabled bool, reader QualityReader, enq QualityEnqueuer) {
	h.qualityEnabled = enabled
	h.qualityReader = reader
	h.qualityEnq = enq
}

// GetQuality — GET /api/v1/notes/:id/quality
// Returns the latest assessment for the note. Disabled feature answers
// {"enabled": false} so the frontend can hide the row entirely.
func (h *Handler) GetQuality(c *gin.Context) {
	if !h.qualityEnabled || h.qualityReader == nil {
		c.JSON(http.StatusOK, gin.H{"enabled": false})
		return
	}
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}
	entry, err := h.qualityReader.LatestQuality(c.Request.Context(), noteID)
	if err != nil {
		log.Printf("GetQuality: read failed for note %s: %v", noteID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read quality"})
		return
	}
	logEntries, err := h.qualityReader.ListQuality(c.Request.Context(), noteID, appquality.LogKeepPerNote)
	if err != nil {
		log.Printf("GetQuality: history read failed for note %s: %v", noteID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read quality log"})
		return
	}
	resp := gin.H{"enabled": true, "quality": nil, "log": logEntries}
	if entry != nil {
		resp["quality"] = gin.H{
			"signals":             entry.Record.Signals,
			"gates":               entry.Record.Gates,
			"verdict":             entry.Record.Verdict,
			"reasons":             entry.Record.Reasons,
			"attempt":             entry.Record.Attempt,
			"needs_manual_review": entry.Record.NeedsManualReview,
			"computed_at":         entry.Record.ComputedAt,
			"pipeline_version":    entry.Record.PipelineVersion,
			"source_hash":         entry.SourceHash,
			"trigger":             entry.Trigger,
			// can_refetch drives the "Перезабрать страницу" action: stub and
			// truncated imports that still carry their source URL.
			"can_refetch": entry.Record.Signals.HasSourceLink &&
				(entry.Record.Signals.TruncatedByImport || entry.Record.Signals.Kind == appquality.KindStub),
		}
	}
	c.JSON(http.StatusOK, resp)
}

// AssessQuality — POST /api/v1/notes/:id/quality/assess
// The manual trigger: enqueues quality:assess and answers 202. The result
// lands via GetQuality once the worker finishes.
func (h *Handler) AssessQuality(c *gin.Context) {
	if !h.qualityEnabled {
		c.JSON(http.StatusOK, gin.H{"enabled": false})
		return
	}
	if h.qualityEnq == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "quality queue unavailable"})
		return
	}
	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note id"})
		return
	}
	if err := h.qualityEnq.EnqueueAssessQuality(c.Request.Context(), noteID.String(), appquality.TriggerManual); err != nil {
		log.Printf("AssessQuality: enqueue failed for note %s: %v", noteID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to enqueue assessment"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"enabled": true, "enqueued": true})
}
