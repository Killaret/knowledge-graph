// Package share provides HTTP handlers for note sharing
package share

import (
	"net/http"
	"time"

	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler handles note sharing requests
type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new share handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
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

	sharedWithUserID, _ := uuid.Parse(req.UserID)

	// Check if note exists and user is the creator
	var note postgres.NoteModel
	if result := h.db.First(&note, noteID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	// Verify ownership
	if note.CreatorID == nil || *note.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can share this note"})
		return
	}

	// Check if user exists
	var targetUser postgres.UserModel
	if result := h.db.First(&targetUser, sharedWithUserID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "target user not found"})
		return
	}

	// Set default permission
	permission := req.Permission
	if permission == "" {
		permission = "read"
	}

	// Check if already shared
	var existingShare postgres.NoteShareModel
	result := h.db.Where("note_id = ? AND shared_with_user_id = ?", noteID, sharedWithUserID).First(&existingShare)
	if result.Error == nil {
		// Update existing share
		existingShare.Permission = permission
		existingShare.ExpiresAt = req.ExpiresAt
		h.db.Save(&existingShare)

		c.JSON(http.StatusOK, ShareNoteResponse{
			ID:               existingShare.ID,
			NoteID:           noteID,
			SharedWithUserID: sharedWithUserID,
			SharedWithLogin:  targetUser.Login,
			Permission:       permission,
			CreatedAt:        existingShare.CreatedAt,
			ExpiresAt:        req.ExpiresAt,
		})
		return
	}

	// Create new share
	share := postgres.NoteShareModel{
		ID:               uuid.New(),
		NoteID:           noteID,
		SharedByUserID:   userID,
		SharedWithUserID: sharedWithUserID,
		Permission:       permission,
		CreatedAt:        time.Now(),
		ExpiresAt:        req.ExpiresAt,
	}

	if result := h.db.Create(&share); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create share"})
		return
	}

	c.JSON(http.StatusCreated, ShareNoteResponse{
		ID:               share.ID,
		NoteID:           noteID,
		SharedWithUserID: sharedWithUserID,
		SharedWithLogin:  targetUser.Login,
		Permission:       permission,
		CreatedAt:        share.CreatedAt,
		ExpiresAt:        req.ExpiresAt,
	})
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

	// Check if note exists and user is the creator
	var note postgres.NoteModel
	if result := h.db.First(&note, noteID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	// Verify ownership
	if note.CreatorID == nil || *note.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can create share links"})
		return
	}

	// Generate token
	token, err := generateShareToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate share token"})
		return
	}

	// Set default permission
	permission := req.Permission
	if permission == "" {
		permission = "read"
	}

	// Create share link
	shareLink := postgres.ShareLinkModel{
		ID:             uuid.New(),
		NoteID:         noteID,
		SharedByUserID: userID,
		Token:          token,
		Permission:     permission,
		CreatedAt:      time.Now(),
		ExpiresAt:      req.ExpiresAt,
		MaxUses:        req.MaxUses,
		IsActive:       true,
	}

	if result := h.db.Create(&shareLink); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create share link"})
		return
	}

	c.JSON(http.StatusCreated, ShareLinkResponse{
		ID:         shareLink.ID,
		Token:      token,
		Permission: permission,
		ExpiresAt:  req.ExpiresAt,
		MaxUses:    req.MaxUses,
		CreatedAt:  shareLink.CreatedAt,
	})
}

// generateShareToken generates a unique share token
func generateShareToken() (string, error) {
	return uuid.New().String(), nil
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

	// Verify ownership and revoke
	result := h.db.Model(&postgres.ShareLinkModel{}).
		Where("id = ? AND shared_by_user_id = ?", linkID, userID).
		Update("is_active", false)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke share link"})
		return
	}

	if result.RowsAffected == 0 {
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

	// Check if note exists and user is the creator
	var note postgres.NoteModel
	if result := h.db.First(&note, noteID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	// Verify ownership
	if note.CreatorID == nil || *note.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only the creator can view shares"})
		return
	}

	// Get user shares
	var shares []postgres.NoteShareModel
	h.db.Preload("SharedWithUser").Where("note_id = ?", noteID).Find(&shares)

	userShares := make([]gin.H, 0, len(shares))
	for _, share := range shares {
		login := ""
		if share.SharedWithUser.ID != uuid.Nil {
			login = share.SharedWithUser.Login
		}
		userShares = append(userShares, gin.H{
			"id":                 share.ID,
			"shared_with_user_id": share.SharedWithUserID,
			"shared_with_login":  login,
			"permission":         share.Permission,
			"created_at":         share.CreatedAt,
			"expires_at":         share.ExpiresAt,
		})
	}

	// Get share links
	var links []postgres.ShareLinkModel
	h.db.Where("note_id = ? AND is_active = ?", noteID, true).Find(&links)

	shareLinks := make([]gin.H, 0, len(links))
	for _, link := range links {
		shareLinks = append(shareLinks, gin.H{
			"id":         link.ID,
			"token":      link.Token,
			"permission": link.Permission,
			"created_at": link.CreatedAt,
			"expires_at": link.ExpiresAt,
			"max_uses":   link.MaxUses,
			"uses_count": link.UsesCount,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user_shares": userShares,
		"share_links": shareLinks,
	})
}

// AccessSharedNote handles access to a shared note via token
func (h *Handler) AccessSharedNote(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "share token required"})
		return
	}

	// Find share link
	var shareLink postgres.ShareLinkModel
	if result := h.db.Where("token = ? AND is_active = ?", token, true).First(&shareLink); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invalid or expired share link"})
		return
	}

	// Check expiration
	if shareLink.ExpiresAt != nil && shareLink.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "share link has expired"})
		return
	}

	// Check max uses
	if shareLink.MaxUses != nil && shareLink.UsesCount >= *shareLink.MaxUses {
		c.JSON(http.StatusGone, gin.H{"error": "share link has reached maximum uses"})
		return
	}

	// Increment uses count
	h.db.Model(&shareLink).Update("uses_count", shareLink.UsesCount+1)

	// Get the note
	var note postgres.NoteModel
	if result := h.db.First(&note, shareLink.NoteID); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "note not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"note": gin.H{
			"id":         note.ID,
			"title":      note.Title,
			"content":    note.Content,
			"type":       note.Type,
			"metadata":   note.Metadata,
			"created_at": note.CreatedAt,
			"updated_at": note.UpdatedAt,
		},
		"permission": shareLink.Permission,
	})
}
