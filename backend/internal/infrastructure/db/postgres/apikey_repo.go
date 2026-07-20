package postgres

import (
	"context"
	"errors"
	"time"

	"knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// APIKeyRepository implements user.APIKeyRepository using GORM.
type APIKeyRepository struct {
	db *gorm.DB
}

// NewAPIKeyRepository creates a new API key repository.
func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func toDomainAPIKey(m *APIKeyModel) (*user.APIKey, error) {
	return user.NewAPIKey(m.ID, m.UserID, m.KeyHash, m.Name, m.Scopes, m.CreatedAt)
}

func fromDomainAPIKey(k *user.APIKey) *APIKeyModel {
	return &APIKeyModel{
		ID:         k.ID(),
		UserID:     k.UserID(),
		KeyHash:    k.KeyHash(),
		Name:       k.Name(),
		Scopes:     k.Scopes(),
		CreatedAt:  k.CreatedAt(),
		ExpiresAt:  k.ExpiresAt(),
		LastUsedAt: k.LastUsedAt(),
		IsActive:   k.IsActive(),
	}
}

// FindByUserID returns all active API keys for a user.
func (r *APIKeyRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]user.APIKey, error) {
	var models []APIKeyModel
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	keys := make([]user.APIKey, 0, len(models))
	for _, m := range models {
		k, err := toDomainAPIKey(&m)
		if err != nil {
			return nil, err
		}
		keys = append(keys, *k)
	}
	return keys, nil
}

// Create inserts a new API key.
func (r *APIKeyRepository) Create(ctx context.Context, key *user.APIKey) error {
	model := fromDomainAPIKey(key)
	model.IsActive = true
	model.CreatedAt = time.Now()
	if model.CreatedAt.IsZero() {
		model.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// Revoke marks an API key as inactive if it belongs to the given user.
func (r *APIKeyRepository) Revoke(ctx context.Context, keyID, userID uuid.UUID) (bool, error) {
	result := r.db.WithContext(ctx).
		Model(&APIKeyModel{}).
		Where("id = ? AND user_id = ?", keyID, userID).
		Update("is_active", false)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// FindActiveByHash returns the active API key with the given hash.
func (r *APIKeyRepository) FindActiveByHash(ctx context.Context, hash string) (*user.APIKey, error) {
	var model APIKeyModel
	err := r.db.WithContext(ctx).
		Where("key_hash = ? AND is_active = ?", hash, true).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainAPIKey(&model)
}

// UpdateLastUsed updates the last_used_at timestamp for the given key ID.
func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, keyID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&APIKeyModel{}).
		Where("id = ?", keyID).
		Update("last_used_at", "now()").Error
}
