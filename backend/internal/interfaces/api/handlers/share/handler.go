// Package share provides HTTP handlers for note sharing
package share

import (
	"net/http"
	"time"

	domainnote "knowledge-graph/internal/domain/note"
	domainshare "knowledge-graph/internal/domain/share"
	domainuser "knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles note sharing requests
type Handler struct {
	noteRepo  domainnote.Repository
	userRepo  domainuser.Repository
	shareRepo domainshare.Repository
}

// NewHandler creates a new share handler
func NewHandler(
	noteRepo domainnote.Repository,
	userRepo domainuser.Repository,
	shareRepo domainshare.Repository,
) *Handler {
	return &Handler{
		noteRepo:  noteRepo,
		userRepo:  userRepo,
		shareRepo: shareRepo,
	}
}

// ShareNoteRequest represents a request to share a note with a user
type ShareNoteRequest struct {
	UserID     string     `json:"user_id" binding:"required,uuid"`
	Permission string     `json:"permission" binding:"omitempty,oneof=read write"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// ShareNoteResponse represents a note share response
type ShareNoteResponse struct {
	ID               uuid.UUID  `json:"id"`
	NoteID           uuid.UUID  `json:"note_id"`
	SharedWithUserID uuid.UUID  `json:"shared_with_user_id"`
	SharedWithLogin  string     `json:"shared_with_login"`
	Permission       string     `json:"permission"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
}

// CreateShareLinkRequest represents a request to create a share link
type CreateShareLinkRequest struct {
	Permission string     `json:"permission" binding:"omitempty,oneof=read write"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	MaxUses    *int       `json:"max_uses,omitempty"`
}

// ShareLinkResponse represents a share link response
type ShareLinkResponse struct {
	ID         uuid.UUID  `json:"id"`
	Token      string     `json:"token"`
	Permission string     `json:"permission"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	MaxUses    *int       `json:"max_uses,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ShareNote shares a note with a specific user
func (h *Handler) ShareNote(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note ID"})
		return
	}

	var req ShareNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sharedWithUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	ctx := c.Request.Context()

	note, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch note"})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	if note.CreatorID() == nil || *note.CreatorID() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can share this note"})
		return
	}

	targetUser, err := h.userRepo.FindByID(ctx, sharedWithUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch target user"})
		return
	}
	if targetUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user not found"})
		return
	}

	permission := req.Permission
	if permission == "" {
		permission = "read"
	}

	existing, err := h.shareRepo.FindShareByNoteAndUser(ctx, noteID, sharedWithUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing share"})
		return
	}

	if existing != nil {
		updatedShare, err := domainshare.NewNoteShare(existing.ID(), noteID, userID, sharedWithUserID, permission, req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update share"})
			return
		}
		if err := h.shareRepo.UpdateShare(ctx, updatedShare); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update share"})
			return
		}
		c.JSON(http.StatusOK, ShareNoteResponse{
			ID:               updatedShare.ID(),
			NoteID:           noteID,
			SharedWithUserID: sharedWithUserID,
			SharedWithLogin:  targetUser.Login(),
			Permission:       updatedShare.Permission(),
			CreatedAt:        updatedShare.CreatedAt(),
			ExpiresAt:        updatedShare.ExpiresAt(),
		})
		return
	}

	newShare, err := domainshare.NewNoteShare(uuid.New(), noteID, userID, sharedWithUserID, permission, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create share"})
		return
	}
	if err := h.shareRepo.CreateShare(ctx, newShare); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create share"})
		return
	}

	c.JSON(http.StatusCreated, ShareNoteResponse{
		ID:               newShare.ID(),
		NoteID:           noteID,
		SharedWithUserID: sharedWithUserID,
		SharedWithLogin:  targetUser.Login(),
		Permission:       newShare.Permission(),
		CreatedAt:        newShare.CreatedAt(),
		ExpiresAt:        newShare.ExpiresAt(),
	})
}

// CreateShareLink creates a public share link for a note
func (h *Handler) CreateShareLink(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note ID"})
		return
	}

	var req CreateShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	note, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch note"})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	if note.CreatorID() == nil || *note.CreatorID() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can create share links"})
		return
	}

	permission := req.Permission
	if permission == "" {
		permission = "read"
	}

	token := uuid.New().String()
	link, err := domainshare.NewShareLink(uuid.New(), noteID, userID, token, permission, req.ExpiresAt, req.MaxUses, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create share link"})
		return
	}

	if err := h.shareRepo.CreateShareLink(ctx, link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create share link"})
		return
	}

	c.JSON(http.StatusCreated, ShareLinkResponse{
		ID:         link.ID(),
		Token:      link.Token(),
		Permission: link.Permission(),
		ExpiresAt:  link.ExpiresAt(),
		MaxUses:    link.MaxUses(),
		CreatedAt:  link.CreatedAt(),
	})
}

// RevokeShareLink revokes a share link
func (h *Handler) RevokeShareLink(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	linkID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid share link ID"})
		return
	}

	revoked, err := h.shareRepo.RevokeShareLink(c.Request.Context(), linkID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke share link"})
		return
	}
	if !revoked {
		c.JSON(http.StatusNotFound, gin.H{"error": "share link not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "share link revoked successfully"})
}

// ListNoteShares returns all shares for a specific note
func (h *Handler) ListNoteShares(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note ID"})
		return
	}

	ctx := c.Request.Context()

	note, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch note"})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	if note.CreatorID() == nil || *note.CreatorID() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can view shares"})
		return
	}

	shares, err := h.shareRepo.ListSharesByNote(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list shares"})
		return
	}

	userShares := make([]gin.H, 0, len(shares))
	for _, s := range shares {
		userShares = append(userShares, gin.H{
			"id":                  s.ID(),
			"shared_with_user_id": s.SharedWithUserID(),
			"shared_with_login":   s.SharedWithLogin(),
			"permission":          s.Permission(),
			"created_at":          s.CreatedAt(),
			"expires_at":          s.ExpiresAt(),
		})
	}

	links, err := h.shareRepo.ListShareLinksByNote(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list share links"})
		return
	}

	shareLinks := make([]gin.H, 0, len(links))
	for _, link := range links {
		shareLinks = append(shareLinks, gin.H{
			"id":         link.ID(),
			"token":      link.Token(),
			"permission": link.Permission(),
			"created_at": link.CreatedAt(),
			"expires_at": link.ExpiresAt(),
			"max_uses":   link.MaxUses(),
			"uses_count": link.UsesCount(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_shares": userShares,
		"share_links": shareLinks,
	})
}

// RevokeShare revokes a direct user-to-user share
func (h *Handler) RevokeShare(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	noteID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid note ID"})
		return
	}

	shareID, err := uuid.Parse(c.Param("shareId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid share ID"})
		return
	}

	ctx := c.Request.Context()

	note, err := h.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch note"})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}
	if note.CreatorID() == nil || *note.CreatorID() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can revoke shares"})
		return
	}

	revoked, err := h.shareRepo.RevokeShare(ctx, noteID, shareID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke share"})
		return
	}
	if !revoked {
		c.JSON(http.StatusNotFound, gin.H{"error": "share not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "share revoked successfully"})
}

// AccessSharedNote handles access to a shared note via token
func (h *Handler) AccessSharedNote(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "share token required"})
		return
	}

	ctx := c.Request.Context()

	shareLink, err := h.shareRepo.FindActiveShareLinkByToken(ctx, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch share link"})
		return
	}
	if shareLink == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid or expired share link"})
		return
	}

	// Check expiration
	if shareLink.ExpiresAt() != nil && shareLink.ExpiresAt().Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "share link has expired"})
		return
	}

	// Check max uses
	if shareLink.MaxUses() != nil && shareLink.UsesCount() >= *shareLink.MaxUses() {
		c.JSON(http.StatusGone, gin.H{"error": "share link has reached maximum uses"})
		return
	}

	// Increment uses count
	_ = h.shareRepo.IncrementShareLinkUsage(ctx, shareLink.ID())

	// Get the note
	note, err := h.noteRepo.FindByID(ctx, shareLink.NoteID())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch note"})
		return
	}
	if note == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"note": gin.H{
			"id":         note.ID(),
			"title":      note.Title().String(),
			"content":    note.Content().String(),
			"type":       note.Type(),
			"metadata":   note.Metadata().Value(),
			"created_at": note.CreatedAt(),
			"updated_at": note.UpdatedAt(),
		},
		"permission": shareLink.Permission(),
	})
}
