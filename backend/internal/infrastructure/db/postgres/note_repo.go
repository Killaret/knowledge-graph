package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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

// invalidateCache удаляет кэш списка заметок
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
		// Инвалидация кэша при создании новой заметки
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
	// Инвалидация кэша при удалении заметки
	r.invalidateCache(ctx)
	return nil
}

// FindAll возвращает все заметки без пагинации с кэшированием в Redis
func (r *NoteRepository) FindAll(ctx context.Context) ([]*note.Note, error) {
	// 1. Проверяем кэш Redis
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, notesCacheKey).Bytes()
		if err == nil {
			var notes []*note.Note
			if err := json.Unmarshal(cached, &notes); err == nil {
				return notes, nil
			}
		}
	}

	// 2. Получаем из БД
	var models []NoteModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	notes := toDomainNotes(models)

	// 3. Сохраняем в кэш
	if r.redis != nil {
		if data, err := json.Marshal(notes); err == nil {
			r.redis.Set(ctx, notesCacheKey, data, notesCacheTTL)
		}
	}

	return notes, nil
}

// List возвращает заметки с пагинацией
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

		db = db.Order(fmt.Sprintf(`
			COALESCE(ts_rank(search_vector, plainto_tsquery('russian', '%s')), 0) +
			COALESCE(ts_rank(search_vector, plainto_tsquery('simple', '%s')), 0) DESC
		`, query, query))

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

// toGormNote преобразует доменную заметку в GORM-модель
func toGormNote(n *note.Note) (NoteModel, error) {
	metadataJSON, err := json.Marshal(n.Metadata().Value())
	if err != nil {
		return NoteModel{}, err
	}
	return NoteModel{
		ID:        n.ID(),
		Title:     n.Title().String(),
		Content:   n.Content().String(),
		Metadata:  datatypes.JSON(metadataJSON),
		CreatedAt: n.CreatedAt(),
		UpdatedAt: n.UpdatedAt(),
	}, nil
}

// toDomainNote преобразует GORM-модель в доменную заметку
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
	return note.ReconstructNote(m.ID, title, content, metadata, m.CreatedAt, m.UpdatedAt), nil
}

// toDomainNotes преобразует список GORM-моделей в список доменных сущностей
func toDomainNotes(models []NoteModel) []*note.Note {
	result := make([]*note.Note, 0, len(models))
	for _, m := range models {
		n, err := toDomainNote(&m)
		if err != nil {
			// Логируем ошибку, но продолжаем (пропускаем битые записи)
			continue
		}
		result = append(result, n)
	}
	return result
}
