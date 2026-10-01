package postgres

import (
	"context"
	"time"

	"knowledge-graph/internal/domain/notification"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NotificationModel is the GORM model for the notifications table.
type NotificationModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	NoteID    *uuid.UUID `gorm:"type:uuid"`
	Type      string     `gorm:"type:varchar(50);not null"`
	Title     string     `gorm:"type:text;not null"`
	Body      string     `gorm:"type:text;not null;default:''"`
	DedupeKey *string    `gorm:"type:text"`
	CreatedAt time.Time
	ReadAt    *time.Time
}

func (NotificationModel) TableName() string {
	return "notifications"
}

// NotificationRepository implements notification.Repository using GORM.
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository.
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func toDomainNotification(m *NotificationModel) *notification.Notification {
	dedupe := ""
	if m.DedupeKey != nil {
		dedupe = *m.DedupeKey
	}
	return notification.Reconstruct(m.ID, m.UserID, m.NoteID, m.Type, m.Title, m.Body, dedupe, m.CreatedAt, m.ReadAt)
}

func (r *NotificationRepository) CreateIfAbsent(ctx context.Context, n *notification.Notification) (bool, error) {
	m := &NotificationModel{
		ID:        n.ID(),
		UserID:    n.UserID(),
		NoteID:    n.NoteID(),
		Type:      n.Type(),
		Title:     n.Title(),
		Body:      n.Body(),
		CreatedAt: n.CreatedAt(),
		ReadAt:    n.ReadAt(),
	}
	if key := n.DedupeKey(); key != "" {
		m.DedupeKey = &key
	}
	res := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:     []clause.Column{{Name: "dedupe_key"}},
			TargetWhere: clause.Where{Exprs: []clause.Expression{clause.Expr{SQL: "dedupe_key IS NOT NULL"}}},
			DoNothing:   true,
		}).
		Create(m)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*notification.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	var models []NotificationModel
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	out := make([]*notification.Notification, 0, len(models))
	for i := range models {
		out = append(out, toDomainNotification(&models[i]))
	}
	return out, nil
}

func (r *NotificationRepository) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Count(&count).Error
	return count, err
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	res := r.db.WithContext(ctx).
		Model(&NotificationModel{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL", id, userID).
		Update("read_at", time.Now())
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
