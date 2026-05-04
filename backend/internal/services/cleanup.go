// Package services provides background service implementations
package services

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CleanupService handles cleanup of soft-deleted records and expired tokens
type CleanupService struct {
	db *gorm.DB
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(db *gorm.DB) *CleanupService {
	return &CleanupService{db: db}
}

// CleanupSoftDeleted permanently deletes records that have been soft-deleted for more than N days
func (s *CleanupService) CleanupSoftDeleted(ctx context.Context, days int, tables []string) error {
	if days <= 0 {
		days = 90
	}

	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Default tables to clean up
	if len(tables) == 0 {
		tables = []string{"users", "notes", "links"}
	}

	for _, table := range tables {
		if err := s.cleanupTable(ctx, table, cutoffDate); err != nil {
			return fmt.Errorf("failed to cleanup table %s: %w", table, err)
		}
	}

	return nil
}

// cleanupTable deletes soft-deleted records from a specific table
func (s *CleanupService) cleanupTable(ctx context.Context, tableName string, cutoffDate time.Time) error {
	// Execute raw SQL for hard delete
	sql := fmt.Sprintf(
		"DELETE FROM %s WHERE deleted_at IS NOT NULL AND deleted_at < ?",
		tableName,
	)
	
	result := s.db.WithContext(ctx).Exec(sql, cutoffDate)
	if result.Error != nil {
		return result.Error
	}

	return nil
}

// CleanupExpiredTokens deletes expired refresh tokens and blacklisted tokens
func (s *CleanupService) CleanupExpiredTokens(ctx context.Context, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 1000
	}

	// Delete expired refresh tokens
	if err := s.db.WithContext(ctx).
		Exec("DELETE FROM refresh_tokens WHERE expires_at < now()").Error; err != nil {
		return fmt.Errorf("failed to cleanup expired refresh tokens: %w", err)
	}

	// Delete revoked refresh tokens older than 7 days
	if err := s.db.WithContext(ctx).
		Exec("DELETE FROM refresh_tokens WHERE revoked_at IS NOT NULL AND revoked_at < now() - interval '7 days'").Error; err != nil {
		return fmt.Errorf("failed to cleanup revoked refresh tokens: %w", err)
	}

	// Delete old audit logs (keep 90 days)
	if err := s.db.WithContext(ctx).
		Exec("DELETE FROM audit_log WHERE created_at < now() - interval '90 days'").Error; err != nil {
		return fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}

	return nil
}

// CleanupExpiredShares disables expired note shares
func (s *CleanupService) CleanupExpiredShares(ctx context.Context) error {
	// Disable expired note shares
	if err := s.db.WithContext(ctx).
		Model(&struct{ TableName string }{TableName: "note_shares"}).
		Where("expires_at IS NOT NULL AND expires_at < now()").
		Update("permission", "expired").Error; err != nil {
		return fmt.Errorf("failed to cleanup expired note shares: %w", err)
	}

	// Disable expired share links
	if err := s.db.WithContext(ctx).
		Exec("UPDATE share_links SET is_active = false WHERE expires_at IS NOT NULL AND expires_at < now()").Error; err != nil {
		return fmt.Errorf("failed to cleanup expired share links: %w", err)
	}

	return nil
}

// CleanupExpiredAPIKeys disables expired API keys
func (s *CleanupService) CleanupExpiredAPIKeys(ctx context.Context) error {
	return s.db.WithContext(ctx).
		Exec("UPDATE api_keys SET is_active = false WHERE expires_at IS NOT NULL AND expires_at < now()").Error
}
