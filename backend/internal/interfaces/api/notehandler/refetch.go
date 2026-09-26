package notehandler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/domain/note"
	apicommon "knowledge-graph/internal/interfaces/api/common"
)

// RefetchExtractor fetches and re-extracts the note's source URL with the
// stage-A pipeline. The quality loop itself never touches the network —
// re-fetch exists only as these explicit user actions.
type RefetchExtractor interface {
	Extract(ctx context.Context, rawURL string) (*importer.ExtractedPage, error)
}

// SetRefetch installs the extractor used by the re-fetch endpoints. Without
// it the endpoints answer 503 — same shape as other optional deps.
func (h *Handler) SetRefetch(x RefetchExtractor) {
	h.refetchExtractor = x
}

// noteSourceURL reads metadata.source_url — set by URL-HEADING-1 imports.
func noteSourceURL(n *note.Note) string {
	if n == nil {
		return ""
	}
	u, _ := n.Metadata().Value()["source_url"].(string)
	return u
}

func (h *Handler) loadNoteForRefetch(c *gin.Context) (*note.Note, uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldErrorWithValue("id", apicommon.ReasonInvalidFormat, apicommon.MsgInvalidUUID, c.Param("id")),
		})
		return nil, uuid.Nil, false
	}
	n, err := h.repo.FindByID(c.Request.Context(), id)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNote)
		return nil, uuid.Nil, false
	}
	if n == nil {
		apicommon.NotFound(c, "Note")
		return nil, uuid.Nil, false
	}
	return n, id, true
}

// extractSource loads the note's source URL and runs stage-A extraction.
// Returns nil page (and already-answered response) on failure.
func (h *Handler) extractSource(c *gin.Context, n *note.Note) *importer.ExtractedPage {
	if h.refetchExtractor == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "refetch unavailable"})
		return nil
	}
	url := noteSourceURL(n)
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "note has no source_url"})
		return nil
	}
	page, err := h.refetchExtractor.Extract(c.Request.Context(), url)
	if err != nil {
		log.Printf("RefetchPreview: extract failed for %s: %v", url, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch source page"})
		return nil
	}
	return page
}

// RefetchPreview — POST /api/v1/notes/:id/refetch/preview
// Fetches the source page and answers with what the new version would look
// like. Pure read: nothing is written until RefetchApply.
func (h *Handler) RefetchPreview(c *gin.Context) {
	n, _, ok := h.loadNoteForRefetch(c)
	if !ok {
		return
	}
	page := h.extractSource(c, n)
	if page == nil {
		return
	}
	title := page.Title
	if len(page.TitleCandidates) > 0 {
		title = page.TitleCandidates[0]
	}
	content := importer.BuildContent(title, noteSourceURL(n), page.Text)
	c.JSON(http.StatusOK, gin.H{
		"suggested_title":  title,
		"title_candidates": page.TitleCandidates,
		"title_source":     page.TitleSource,
		"outline":          page.Outline,
		"noise_dropped":    page.NoiseDropped,
		"sections_dropped": page.SectionsDropped,
		"length_runes":     len([]rune(content)),
		"current_runes":    len([]rune(n.Content().String())),
	})
}

// RefetchApply — POST /api/v1/notes/:id/refetch
// Replaces the note content with a fresh stage-A extraction. The previous
// title+content move to metadata.previous_content — byte-for-byte for
// RefetchRestore. Optional body {"title": "..."} overrides the suggestion.
func (h *Handler) RefetchApply(c *gin.Context) {
	n, id, ok := h.loadNoteForRefetch(c)
	if !ok {
		return
	}
	page := h.extractSource(c, n)
	if page == nil {
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	// Body is optional — ignore bind errors on empty bodies.
	_ = c.ShouldBindJSON(&req)

	title := page.Title
	if len(page.TitleCandidates) > 0 {
		title = page.TitleCandidates[0]
	}
	if req.Title != "" {
		title = req.Title
	}

	newContent, err := note.NewContent(importer.BuildContent(title, noteSourceURL(n), page.Text))
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "refetched content is invalid")
		return
	}
	newTitle, err := note.NewTitle(title)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "refetched title is invalid")
		return
	}

	// Save the previous version before overwriting — restore needs the
	// exact bytes.
	meta := n.Metadata().Value()
	if meta == nil {
		meta = map[string]any{}
	}
	meta["previous_content"] = map[string]any{
		"title":    n.Title().String(),
		"content":  n.Content().String(),
		"saved_at": time.Now().UTC().Format(time.RFC3339),
	}
	newMeta, err := note.NewMetadata(meta)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if err := n.UpdateMetadata(newMeta); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if err := n.UpdateTitle(newTitle); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if err := n.UpdateContent(newContent); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if err := h.repo.Save(c.Request.Context(), n); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishNoteUpdated(context.Background(), n.ID().String(), getUserIDString(c)); err != nil {
			log.Printf("[NoteHandler] failed to publish NoteUpdated for refetch %s: %v", n.ID(), err)
		}
	}

	// Same text-changed triggers as Update: keywords, embedding, normalize,
	// link weights — the worker then re-assesses quality itself.
	if h.taskQueue != nil {
		noteID := id.String()
		_ = h.taskQueue.EnqueueExtractKeywords(c.Request.Context(), noteID, 10)
		_ = h.taskQueue.EnqueueComputeEmbedding(c.Request.Context(), noteID)
		_ = h.taskQueue.EnqueueNormalizeNote(c.Request.Context(), noteID)
		_ = h.taskQueue.EnqueueRecalculateLinkWeights(c.Request.Context(), id, h.taskDelay)
	}

	c.JSON(http.StatusOK, gin.H{
		"title":         n.Title().String(),
		"content_runes": len([]rune(n.Content().String())),
		"restorable":    true,
	})
}

// RefetchRestore — POST /api/v1/notes/:id/refetch/restore
// Byte-for-byte restore from metadata.previous_content; drops the key so a
// second restore is a no-op answered with 404.
func (h *Handler) RefetchRestore(c *gin.Context) {
	n, id, ok := h.loadNoteForRefetch(c)
	if !ok {
		return
	}
	prev, ok := n.Metadata().Value()["previous_content"].(map[string]any)
	if !ok || prev == nil {
		apicommon.NotFound(c, "previous_content")
		return
	}
	oldContent, _ := prev["content"].(string)
	oldTitle, _ := prev["title"].(string)

	content, err := note.NewContent(oldContent)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if err := n.UpdateContent(content); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if oldTitle != "" {
		if title, err := note.NewTitle(oldTitle); err == nil {
			if err := n.UpdateTitle(title); err != nil {
				apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
				return
			}
		}
	}
	meta := n.Metadata().Value()
	delete(meta, "previous_content")
	newMeta, err := note.NewMetadata(meta)
	if err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if err := n.UpdateMetadata(newMeta); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if err := h.repo.Save(c.Request.Context(), n); err != nil {
		apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedUpdateNote)
		return
	}
	if h.eventPublisher != nil {
		if err := h.eventPublisher.PublishNoteUpdated(context.Background(), n.ID().String(), getUserIDString(c)); err != nil {
			log.Printf("[NoteHandler] failed to publish NoteUpdated for refetch restore %s: %v", n.ID(), err)
		}
	}

	if h.taskQueue != nil {
		noteID := id.String()
		_ = h.taskQueue.EnqueueExtractKeywords(c.Request.Context(), noteID, 10)
		_ = h.taskQueue.EnqueueComputeEmbedding(c.Request.Context(), noteID)
		_ = h.taskQueue.EnqueueNormalizeNote(c.Request.Context(), noteID)
		_ = h.taskQueue.EnqueueRecalculateLinkWeights(c.Request.Context(), id, h.taskDelay)
	}

	c.JSON(http.StatusOK, gin.H{"restored": true})
}
