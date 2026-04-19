package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RecommendationRepository provides access to precomputed note recommendations
type RecommendationRepository struct {
	db *gorm.DB
}

// NewRecommendationRepository creates a new recommendation repository
func NewRecommendationRepository(db *gorm.DB) *RecommendationRepository {
	return &RecommendationRepository{db: db}
}

// Get retrieves recommendations for a given note, sorted by score descending
func (r *RecommendationRepository) Get(ctx context.Context, noteID uuid.UUID, limit int) ([]RecommendationModel, error) {
	var recommendations []RecommendationModel
	err := r.db.WithContext(ctx).
		Where("note_id = ?", noteID).
		Order("score DESC").
		Limit(limit).
		Find(&recommendations).Error
	return recommendations, err
}

// SaveBatch saves a batch of recommendations for a note using upsert (ON CONFLICT DO UPDATE)
func (r *RecommendationRepository) SaveBatch(ctx context.Context, noteID uuid.UUID, recs map[uuid.UUID]float64) error {
	if len(recs) == 0 {
		return nil
	}

	models := make([]RecommendationModel, 0, len(recs))
	for targetID, score := range recs {
		models = append(models, RecommendationModel{
			NoteID:            noteID,
			RecommendedNoteID: targetID,
			Score:             score,
		})
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "note_id"},
				{Name: "recommended_note_id"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"score", "updated_at"}),
		}).
		Create(&models).Error
}

// DeleteByNote removes all recommendations for a given note
func (r *RecommendationRepository) DeleteByNote(ctx context.Context, noteID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("note_id = ?", noteID).
		Delete(&RecommendationModel{}).Error
}

// DeleteNotInBatch removes recommendations that are not in the keep set
// This is used to clean up stale recommendations after recalculation
func (r *RecommendationRepository) DeleteNotInBatch(ctx context.Context, noteID uuid.UUID, keep map[uuid.UUID]float64) error {
	if len(keep) == 0 {
		// If keep is empty, delete all recommendations for this note
		return r.DeleteByNote(ctx, noteID)
	}

	// Build slice of IDs to keep
	keepIDs := make([]uuid.UUID, 0, len(keep))
	for id := range keep {
		keepIDs = append(keepIDs, id)
	}

	return r.db.WithContext(ctx).
		Where("note_id = ? AND recommended_note_id NOT IN ?", noteID, keepIDs).
		Delete(&RecommendationModel{}).Error
}

// GetNotesThatRecommend returns all note IDs that recommend the given note
func (r *RecommendationRepository) GetNotesThatRecommend(ctx context.Context, recommendedID uuid.UUID) ([]uuid.UUID, error) {
	var noteIDs []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&RecommendationModel{}).
		Where("recommended_note_id = ?", recommendedID).
		Pluck("note_id", &noteIDs).Error
	return noteIDs, err
}

// GetStaleRecommendations returns recommendations older than the given time
func (r *RecommendationRepository) GetStaleRecommendations(ctx context.Context, olderThan int, limit int) ([]RecommendationModel, error) {
	var recommendations []RecommendationModel
	err := r.db.WithContext(ctx).
		Where("updated_at < now() - interval '? hours'", olderThan).
		Order("updated_at ASC").
		Limit(limit).
		Find(&recommendations).Error
	return recommendations, err
}

// Count returns the total number of recommendations for a note
func (r *RecommendationRepository) Count(ctx context.Context, noteID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&RecommendationModel{}).
		Where("note_id = ?", noteID).
		Count(&count).Error
	return count, err
}
