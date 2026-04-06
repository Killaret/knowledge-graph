package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type GetSuggestionsQuery struct {
	NoteID uuid.UUID
	Limit  int
}

type SuggestionDTO struct {
	NoteID uuid.UUID `json:"note_id"`
	Title  string    `json:"title"`
	Score  float64   `json:"score"`
}

type GetSuggestionsHandler struct {
	traversalSvc *graph.TraversalService
	noteRepo     note.Repository
	cache        *redis.Client
	cacheTTL     time.Duration
}

func NewGetSuggestionsHandler(
	traversalSvc *graph.TraversalService,
	noteRepo note.Repository,
	cache *redis.Client,
	cacheTTL time.Duration,
) *GetSuggestionsHandler {
	return &GetSuggestionsHandler{
		traversalSvc: traversalSvc,
		noteRepo:     noteRepo,
		cache:        cache,
		cacheTTL:     cacheTTL,
	}
}

func (h *GetSuggestionsHandler) Handle(ctx context.Context, query GetSuggestionsQuery) ([]SuggestionDTO, error) {
	// 1. Проверяем кэш
	cacheKey := fmt.Sprintf("suggestions:%s:%d", query.NoteID.String(), query.Limit)
	if h.cache != nil {
		cached, err := h.cache.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var result []SuggestionDTO
			if err := json.Unmarshal(cached, &result); err == nil {
				return result, nil
			}
		}
	}

	// 2. Получаем рекомендации через доменный сервис (глубина 3, затухание 0.5, берём с запасом)
	suggestions, err := h.traversalSvc.GetSuggestions(ctx, query.NoteID, query.Limit*2)
	if err != nil {
		return nil, err
	}

	// 3. Загружаем заголовки заметок (можно оптимизировать, загружая все сразу)
	result := make([]SuggestionDTO, 0, len(suggestions))
	for _, s := range suggestions {
		noteEntity, err := h.noteRepo.FindByID(ctx, s.NodeID)
		if err != nil || noteEntity == nil {
			continue
		}
		result = append(result, SuggestionDTO{
			NoteID: s.NodeID,
			Title:  noteEntity.Title().String(),
			Score:  s.Score,
		})
		if len(result) >= query.Limit {
			break
		}
	}

	// 4. Сохраняем в кэш
	if h.cache != nil && len(result) > 0 {
		data, _ := json.Marshal(result)
		h.cache.Set(ctx, cacheKey, data, h.cacheTTL)
	}

	return result, nil
}
