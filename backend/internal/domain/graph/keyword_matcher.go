package graph

import (
	"context"

	"github.com/google/uuid"
)

// KeywordMatcher — интерфейс для keyword-based similarity
// Будет реализован через JaccardSimilarity на основе NoteKeywordRepository
type KeywordMatcher interface {
	// Match возвращает similarity scores между sourceID и candidateIDs
	// based on keyword overlap (Jaccard index)
	Match(ctx context.Context, sourceID uuid.UUID, candidateIDs []uuid.UUID) (map[uuid.UUID]float64, error)
}

// NoOpKeywordMatcher — заглушка, возвращающая пустую мапу
// Используется когда keyword-компонент отключен (gamma = 0)
type NoOpKeywordMatcher struct{}

// Match implements KeywordMatcher
func (n *NoOpKeywordMatcher) Match(ctx context.Context, sourceID uuid.UUID, candidateIDs []uuid.UUID) (map[uuid.UUID]float64, error) {
	return make(map[uuid.UUID]float64), nil
}

// Ensure NoOpKeywordMatcher implements KeywordMatcher
var _ KeywordMatcher = (*NoOpKeywordMatcher)(nil)
