package graph

import (
	"context"

	"github.com/google/uuid"
)

// KeywordMatcher — interface for keyword-based similarity
// Will be implemented via JaccardSimilarity based on NoteKeywordRepository
type KeywordMatcher interface {
	// Match returns similarity scores between sourceID and candidateIDs
	// based on keyword overlap (Jaccard index)
	Match(ctx context.Context, sourceID uuid.UUID, candidateIDs []uuid.UUID) (map[uuid.UUID]float64, error)
}

// NoOpKeywordMatcher — stub that returns an empty map
// Used when the keyword component is disabled (gamma = 0)
type NoOpKeywordMatcher struct{}

// Match implements KeywordMatcher
func (n *NoOpKeywordMatcher) Match(ctx context.Context, sourceID uuid.UUID, candidateIDs []uuid.UUID) (map[uuid.UUID]float64, error) {
	return make(map[uuid.UUID]float64), nil
}

// Ensure NoOpKeywordMatcher implements KeywordMatcher
var _ KeywordMatcher = (*NoOpKeywordMatcher)(nil)
