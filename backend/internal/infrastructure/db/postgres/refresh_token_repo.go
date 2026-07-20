package postgres

import (
	"context"

	"knowledge-graph/internal/auth"

	"gorm.io/gorm"
)

// RefreshTokenRepository implements auth.RefreshTokenRepository using GORM.
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create saves a refresh token.
func (r *RefreshTokenRepository) Create(ctx context.Context, token *auth.RefreshToken) error {
	model := RefreshTokenModel{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		IPAddress: token.IPAddress,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}
