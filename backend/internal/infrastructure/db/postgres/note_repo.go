package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	notesCacheKey = "notes:all"
	notesCacheTTL = 5 * time.Minute
)

type NoteRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewNoteRepository(db *gorm.DB, redis *redis.Client) *NoteRepository {
	return &NoteRepository{db: db, redis: redis}
}

// invalidateCache removes the notes-list cache
func (r *NoteRepository) invalidateCache(ctx context.Context) {
	if r.redis != nil {
		r.redis.Del(ctx, notesCacheKey)
	}
}

func (r *NoteRepository) Save(ctx context.Context, n *note.Note) error {
	var existing NoteModel
	err := r.db.WithContext(ctx).Where("id = ?", n.ID()).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model, err := toGormNote(n)
		if err != nil {
			return err
		}
		if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
			return err
		}
		// Invalidate the cache when a new note is created
		r.invalidateCache(ctx)
		return nil
	}
	if err != nil {
		return err
	}
	model, err := toGormNote(n)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(model).Error
}

func (r *NoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	var model NoteModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("[INFO] note not found: id=%s", id.String())
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainNote(&model)
}

func (r *NoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&NoteModel{}, "id = ?", id).Error; err != nil {
		return err
	}
	// Invalidate the cache when a note is deleted
	r.invalidateCache(ctx)
	return nil
}

// FindAllPaginated returns notes with pagination at the database level
// limit=0 means "all records"
func (r *NoteRepository) FindAllPaginated(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	var total int64

	// Count the total number
	if err := r.db.WithContext(ctx).Model(&NoteModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Paginated query
	query := r.db.WithContext(ctx).Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	var models []NoteModel
	if err := query.Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return toDomainNotes(models), total, nil
}

// FindAll returns all notes without pagination, cached in Redis
// DEPRECATED: use FindAllPaginated for large data sets
func (r *NoteRepository) FindAll(ctx context.Context) ([]*note.Note, error) {
	// 1. Check the Redis cache (we cache NoteModel, not Note, because Note has unexported fields)
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, notesCacheKey).Bytes()
		if err == nil {
			var models []NoteModel
			if err := json.Unmarshal(cached, &models); err == nil {
				// Convert models into domain objects
				return toDomainNotes(models), nil
			}
		}
	}

	// 2. Fetch from the database
	var models []NoteModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	notes := toDomainNotes(models)

	// 3. Store in the cache (NoteModel with exported fields)
	if r.redis != nil {
		if data, err := json.Marshal(models); err == nil {
			r.redis.Set(ctx, notesCacheKey, data, notesCacheTTL)
		}
	}

	return notes, nil
}

// List returns notes with pagination
func (r *NoteRepository) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	var models []NoteModel
	var total int64

	db := r.db.WithContext(ctx).Model(&NoteModel{})

	// Count total
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	return toDomainNotes(models), total, nil
}

// Search performs multilingual full-text search on notes (Russian + English)
// Falls back to ILIKE search if full-text search returns no results
func (r *NoteRepository) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
	var models []NoteModel
	var total int64

	// Try full-text search first
	if query != "" {
		db := r.db.WithContext(ctx).Model(&NoteModel{})

		// Multilingual search using tsvector
		db = db.Where(`
			search_vector @@ plainto_tsquery('russian', ?) OR 
			search_vector @@ plainto_tsquery('simple', ?)
		`, query, query)

		// Safe sorting: use a placeholder for the query in ts_rank
		db = db.Order(clause.Expr{
			SQL:  "COALESCE(ts_rank(search_vector, plainto_tsquery('russian', ?)), 0) + COALESCE(ts_rank(search_vector, plainto_tsquery('simple', ?)), 0) DESC",
			Vars: []interface{}{query, query},
		})

		// Count and get results
		if err := db.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		if total > 0 {
			// Full-text search returned results, use them
			err := db.Limit(limit).Offset(offset).Find(&models).Error
			if err != nil {
				return nil, 0, err
			}
			return toDomainNotes(models), total, nil
		}

		// Fallback: use ILIKE search if full-text returned nothing
		dbLike := r.db.WithContext(ctx).Model(&NoteModel{})
		dbLike = dbLike.Where(`
			title ILIKE ? OR content ILIKE ?
		`, "%"+query+"%", "%"+query+"%")
		dbLike = dbLike.Order("created_at DESC")

		if err := dbLike.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		err := dbLike.Limit(limit).Offset(offset).Find(&models).Error
		if err != nil {
			return nil, 0, err
		}
		return toDomainNotes(models), total, nil
	}

	// Empty query - return all notes
	db := r.db.WithContext(ctx).Model(&NoteModel{}).Order("created_at DESC")
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	return toDomainNotes(models), total, nil
}

// toGormNote converts a domain note into a GORM model
func toGormNote(n *note.Note) (NoteModel, error) {
	metadataJSON, err := json.Marshal(n.Metadata().Value())
	if err != nil {
		return NoteModel{}, err
	}
	noteType := n.Type()
	if noteType == "" {
		noteType = "star"
	}
	return NoteModel{
		ID:        n.ID(),
		Title:     n.Title().String(),
		Content:   n.Content().String(),
		Type:      noteType,
		Metadata:  datatypes.JSON(metadataJSON),
		CreatedAt: n.CreatedAt(),
		UpdatedAt: n.UpdatedAt(),
	}, nil
}

// toDomainNote converts a GORM model into a domain note
func toDomainNote(m *NoteModel) (*note.Note, error) {
	title, err := note.NewTitle(m.Title)
	if err != nil {
		return nil, err
	}
	content, err := note.NewContent(m.Content)
	if err != nil {
		return nil, err
	}
	var metadataMap map[string]interface{}
	if len(m.Metadata) > 0 {
		if err := json.Unmarshal(m.Metadata, &metadataMap); err != nil {
			return nil, err
		}
	}
	metadata, err := note.NewMetadata(metadataMap)
	if err != nil {
		return nil, err
	}
	noteType := m.Type
	if noteType == "" {
		noteType = "star"
	}
	return note.ReconstructNote(m.ID, title, content, noteType, metadata, m.CreatedAt, m.UpdatedAt), nil
}

// toDomainNotes converts a list of GORM models into a list of domain entities
func toDomainNotes(models []NoteModel) []*note.Note {
	result := make([]*note.Note, 0, len(models))
	for _, m := range models {
		n, err := toDomainNote(&m)
		if err != nil {
			// Log the error but continue (skip broken records)
			continue
		}
		result = append(result, n)
	}
	return result
}
