package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"knowledge-graph/internal/domain/cache"
	"knowledge-graph/internal/domain/note"
	contextkeys "knowledge-graph/internal/shared/context"

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

// applyNoteScope фильтрует заметки по пользователю: авторизованный — только свои,
// анонимный — только публичные.
func applyNoteScope(db *gorm.DB, userID uuid.UUID) *gorm.DB {
	// In SKIP_AUTH mode the test user has uuid.Nil, which would normally be treated as
	// an anonymous (public-only) request. Use the request context flag to keep the test
	// user scope unmodified and return all notes.
	if db.Statement != nil && db.Statement.Context != nil {
		if isSkip, _ := db.Statement.Context.Value(contextkeys.SkipAuthKey).(bool); isSkip {
			return db
		}
	}

	// A verified identity scopes to its own notes — including the seeded test
	// user who legitimately owns notes as uuid.Nil. Only a request without a
	// verified identity (anonymous) is limited to public notes.
	if db.Statement != nil && db.Statement.Context != nil {
		if authed, _ := db.Statement.Context.Value(contextkeys.AuthenticatedKey).(bool); authed {
			return db.Where("creator_id = ?", userID.String())
		}
	}

	if userID == uuid.Nil {
		return db.Where("is_public = ?", true)
	}
	return db.Where("creator_id = ?", userID.String())
}

// invalidateCache удаляет кэш списка заметок
func (r *NoteRepository) invalidateCache(ctx context.Context) {
	if r.cache != nil {
		if err := r.cache.Del(ctx, notesCacheKey); err != nil {
			log.Printf("failed to invalidate notes cache: %v", err)
		}
	}
}

// NoteAffectedRow is the (id, creator_id) pair a write touched, reported via
// RETURNING so the outbox decorator (SYNC-1 A2) knows which events to record.
type NoteAffectedRow struct {
	ID        uuid.UUID
	CreatorID *uuid.UUID
}

func (r *NoteRepository) Save(ctx context.Context, n *note.Note) error {
	_, err := r.SaveReturning(ctx, n)
	return err
}

// SaveReturning behaves like Save and reports whether the row was created
// (true) or updated (false) — the outbox decorator maps that to
// NoteCreated/NoteUpdated.
func (r *NoteRepository) SaveReturning(ctx context.Context, n *note.Note) (bool, error) {
	created := false
	// Use explicit transaction to ensure clean state
	err := dbFromContext(ctx, r.db).Transaction(func(tx *gorm.DB) error {
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
			created = true
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
		// Select("*") ensures boolean zero values (e.g. is_public=false) are persisted.
		return tx.Model(&existing).Select("*").Updates(model).Error
	})
	return created, err
}

func (r *NoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return r.findByID(ctx, id, false)
}

// FindByIDIncludingDeleted is FindByID without the soft-delete scope: the
// restore route's access check needs the row that sits in the trash.
func (r *NoteRepository) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return r.findByID(ctx, id, true)
}

func (r *NoteRepository) findByID(ctx context.Context, id uuid.UUID, includeDeleted bool) (*note.Note, error) {
	var model NoteModel
	query := r.db.WithContext(ctx)
	if includeDeleted {
		query = query.Unscoped()
	}
	err := query.Where("id = ?", id).First(&model).Error
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
	return r.DeleteBatch(ctx, []uuid.UUID{id})
}

// DeleteBatch soft-deletes multiple notes by ID in a single transaction.
// The notes' still-live links are soft-deleted alongside and marked via
// deleted_via_note_id so Restore can tell them from links removed on their
// own.
func (r *NoteRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	_, err := r.DeleteBatchReturning(ctx, ids)
	return err
}

// DeleteBatchReturning is DeleteBatch plus the (id, creator_id) pairs of the
// soft-deleted notes, collected via RETURNING for the outbox decorator.
func (r *NoteRepository) DeleteBatchReturning(ctx context.Context, ids []uuid.UUID) ([]NoteAffectedRow, error) {
	var affected []NoteAffectedRow
	if len(ids) == 0 {
		return nil, nil
	}
	err := dbFromContext(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&LinkModel{}).
			Where("deleted_at IS NULL AND (source_note_id IN ? OR target_note_id IN ?)", ids, ids).
			Updates(map[string]any{
				"deleted_at":          gorm.Expr("now()"),
				"deleted_via_note_id": gorm.Expr("CASE WHEN source_note_id IN ? THEN source_note_id ELSE target_note_id END", ids),
			}).Error; err != nil {
			return err
		}
		var notes []NoteModel
		if err := tx.Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "creator_id"}}}).
			Where("id IN ?", ids).Delete(&notes).Error; err != nil {
			return err
		}
		for i := range notes {
			affected = append(affected, NoteAffectedRow{ID: notes[i].ID, CreatorID: notes[i].CreatorID})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Инвалидация кэша при удалении заметок
	r.invalidateCache(ctx)
	return affected, nil
}

// Restore recovers a soft-deleted note by clearing its deleted_at timestamp.
// Links that went down with the note come back once both endpoints are
// alive again; links deleted on their own stay deleted.
func (r *NoteRepository) Restore(ctx context.Context, id uuid.UUID) error {
	_, err := r.RestoreReturning(ctx, id)
	return err
}

// RestoreReturning is Restore plus the restored (id, creator_id) pair — empty
// when nothing was restored (ErrNoteNotFound is returned as before).
func (r *NoteRepository) RestoreReturning(ctx context.Context, id uuid.UUID) ([]NoteAffectedRow, error) {
	var restored []NoteModel
	err := dbFromContext(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		result := tx.Unscoped().
			Model(&restored).
			Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "creator_id"}}}).
			Where("id = ? AND deleted_at IS NOT NULL", id).
			Update("deleted_at", nil)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return note.ErrNoteNotFound
		}
		return tx.Exec(`
			UPDATE links l
			SET deleted_at = NULL, deleted_via_note_id = NULL
			WHERE l.deleted_via_note_id IS NOT NULL
			  AND (l.source_note_id = ? OR l.target_note_id = ?)
			  AND NOT EXISTS (
			      SELECT 1 FROM notes n
			      WHERE n.id IN (l.source_note_id, l.target_note_id)
			        AND n.deleted_at IS NOT NULL
			  )`, id, id).Error
	})
	if err != nil {
		return nil, err
	}
	affected := make([]NoteAffectedRow, 0, len(restored))
	for i := range restored {
		affected = append(affected, NoteAffectedRow{ID: restored[i].ID, CreatorID: restored[i].CreatorID})
	}
	// Инвалидация кэша при восстановлении заметки
	r.invalidateCache(ctx)
	return affected, nil
}

// PurgeDeletedBefore hard-deletes notes soft-deleted before cutoff; their
// link rows go through the FK cascade. Links soft-deleted on their own are
// removed by the same horizon. Returns the number of notes purged.
func (r *NoteRepository) PurgeDeletedBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	purged, _, err := r.PurgeDeletedBeforeReturning(ctx, cutoff)
	return purged, err
}

// PurgeDeletedBeforeReturning additionally reports the purged notes'
// (id, creator_id) pairs so the outbox decorator can log events for them.
func (r *NoteRepository) PurgeDeletedBeforeReturning(ctx context.Context, cutoff time.Time) (int64, []NoteAffectedRow, error) {
	var purged int64
	var notes []NoteModel
	err := dbFromContext(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM links WHERE deleted_at IS NOT NULL AND deleted_at < ?", cutoff).Error; err != nil {
			return err
		}
		res := tx.Unscoped().
			Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}, {Name: "creator_id"}}}).
			Where("deleted_at IS NOT NULL AND deleted_at < ?", cutoff).
			Delete(&notes)
		if res.Error != nil {
			return res.Error
		}
		purged = res.RowsAffected
		return nil
	})
	affected := make([]NoteAffectedRow, 0, len(notes))
	for i := range notes {
		affected = append(affected, NoteAffectedRow{ID: notes[i].ID, CreatorID: notes[i].CreatorID})
	}
	return purged, affected, err
}

// FindAllPaginated возвращает заметки с пагинацией на уровне БД.
// userID = uuid.Nil — только публичные, иначе только заметки пользователя.
// limit=0 означает "все записи"
func (r *NoteRepository) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	var total int64

	// Считаем общее количество с учётом видимости
	countQuery := applyNoteScope(r.db.WithContext(ctx).Model(&NoteModel{}), userID)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Запрос с пагинацией
	query := applyNoteScope(r.db.WithContext(ctx), userID).Order("created_at DESC")
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

// List возвращает заметки с пагинацией. userID = uuid.Nil — только публичные.
func (r *NoteRepository) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	// NoteRepository has its own per-user scope; no need for a transaction here.
	return r.FindAllPaginated(ctx, userID, limit, offset)
}

// FindComets возвращает незакрытые кометы пользователя для «Ближайших дел»:
// датированные по возрастанию due_at (просроченные естественно первыми — их
// дата меньше now), бездатные — отдельной группой в конце. COMET-1 этап A.
func (r *NoteRepository) FindComets(ctx context.Context, userID uuid.UUID) ([]*note.Note, error) {
	var models []NoteModel
	err := r.db.WithContext(ctx).
		Where("type = ?", "comet").
		Where("creator_id = ?", userID).
		Where("done_at IS NULL").
		Order("due_at ASC NULLS LAST, created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainNotes(models), nil
}

// Search performs multilingual full-text search on notes (Russian + English).
// userID = uuid.Nil — ищет только по публичным, иначе только по заметкам пользователя.
// Falls back to ILIKE search if full-text search returns no results
func (r *NoteRepository) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	var models []NoteModel
	var total int64

	// Try full-text search first
	if query != "" {
		db := applyNoteScope(r.db.WithContext(ctx).Model(&NoteModel{}), userID)

		// Multilingual search using tsvector
		db = db.Where(`(
			search_vector @@ plainto_tsquery('russian', ?) OR
			search_vector @@ plainto_tsquery('simple', ?)
		)`, query, query)

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
		dbLike := applyNoteScope(r.db.WithContext(ctx).Model(&NoteModel{}), userID)
		dbLike = dbLike.Where(`
			(title ILIKE ? OR content ILIKE ?)
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

	// Empty query - return all notes scoped by user
	db := applyNoteScope(r.db.WithContext(ctx).Model(&NoteModel{}), userID).Order("created_at DESC")
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
		noteType = "unknown"
	}
	return NoteModel{
		ID:                  n.ID(),
		Title:               n.Title().String(),
		Content:             n.Content().String(),
		Type:                noteType,
		Metadata:            datatypes.JSON(metadataJSON),
		CreatorID:           n.CreatorID(),
		IsPublic:            n.IsPublic(),
		DueAt:               n.DueAt(),
		RemindBeforeSeconds: n.RemindBeforeSeconds(),
		DoneAt:              n.DoneAt(),
		CreatedAt:           n.CreatedAt(),
		UpdatedAt:           n.UpdatedAt(),
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
	noteType := note.NewTypeOrUnknown(m.Type)
	return note.ReconstructNoteWithCreator(m.ID, title, content, noteType, metadata, m.CreatorID, m.CreatedAt, m.UpdatedAt,
		note.WithIsPublic(m.IsPublic),
		note.WithCometFields(m.DueAt, m.RemindBeforeSeconds, m.DoneAt)), nil
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
