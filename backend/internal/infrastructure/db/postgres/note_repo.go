package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"knowledge-graph/internal/domain/cache"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
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
	cache cache.CacheClient
}

func NewNoteRepository(db *gorm.DB, cacheClient cache.CacheClient) *NoteRepository {
	return &NoteRepository{db: db, cache: cacheClient}
}

// invalidateCache удаляет кэш списка заметок
func (r *NoteRepository) invalidateCache(ctx context.Context) {
	if r.cache != nil {
		if err := r.cache.Del(ctx, notesCacheKey); err != nil {
			log.Printf("failed to invalidate notes cache: %v", err)
		}
	}
}

func (r *NoteRepository) Save(ctx context.Context, n *note.Note) error {
	// Use explicit transaction to ensure clean state
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing NoteModel
		err := tx.Where("id = ?", n.ID()).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			model, err := toGormNote(n)
			if err != nil {
				return err
			}
			if err := tx.Create(&model).Error; err != nil {
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
		return tx.Model(&existing).Updates(model).Error
	})
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

// DeleteBatch soft-deletes multiple notes by ID in a single transaction.
func (r *NoteRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Delete(&NoteModel{}, "id IN ?", ids).Error; err != nil {
		return err
	}
	// Инвалидация кэша при удалении заметок
	r.invalidateCache(ctx)
	return nil
}

// Restore recovers a soft-deleted note by clearing its deleted_at timestamp.
func (r *NoteRepository) Restore(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Unscoped().
		Model(&NoteModel{}).
		Where("id = ?", id).
		Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return note.ErrNoteNotFound
	}
	// Инвалидация кэша при восстановлении заметки
	r.invalidateCache(ctx)
	return nil
}

// FindAllPaginated возвращает заметки с пагинацией на уровне БД
// limit=0 означает "все записи"
func (r *NoteRepository) FindAllPaginated(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	var total int64

	// Считаем общее количество
	if err := r.db.WithContext(ctx).Model(&NoteModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Запрос с пагинацией
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

// FindAll возвращает все заметки без пагинации с кэшированием
// DEPRECATED: используйте FindAllPaginated для больших наборов данных
func (r *NoteRepository) FindAll(ctx context.Context) ([]*note.Note, error) {
	// 1. Проверяем кэш (кэшируем NoteModel, а не Note, т.к. у Note неэкспортированные поля)
	if r.cache != nil {
		cached, err := r.cache.Get(ctx, notesCacheKey)
		if err == nil {
			var models []NoteModel
			if err := json.Unmarshal([]byte(cached), &models); err == nil {
				// Конвертируем модели в доменные объекты
				return toDomainNotes(models), nil
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

	// 3. Сохраняем в кэш (NoteModel с экспортированными полями)
	if r.cache != nil {
		if data, err := json.Marshal(models); err == nil {
			if err := r.cache.Set(ctx, notesCacheKey, string(data), notesCacheTTL); err != nil {
				log.Printf("failed to cache notes: %v", err)
			}
		}
	}

	return notes, nil
}

// List возвращает заметки с пагинацией
func (r *NoteRepository) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	var models []NoteModel
	var total int64

	// Use explicit transaction to ensure clean state
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Count total
		if err := tx.Model(&NoteModel{}).Count(&total).Error; err != nil {
			return err
		}

		// Get paginated results
		err := tx.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error
		if err != nil {
			return err
		}
		return nil
	})
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

		// Безопасная сортировка: используем placeholder для query в ts_rank
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

// toGormNote преобразует доменную заметку в GORM-модель
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
		CreatorID: n.CreatorID(),
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
	noteType := m.Type
	if noteType == "" {
		noteType = "star"
	}
	return note.ReconstructNoteWithCreator(m.ID, title, content, noteType, metadata, m.CreatorID, m.CreatedAt, m.UpdatedAt), nil
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
