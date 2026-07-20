package graph

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAggregateMax(t *testing.T) {
	assert.Equal(t, 0.5, AggregateMax(0.3, 0.5))
	assert.Equal(t, 0.3, AggregateMax(0.3, 0.1))
}

func TestAggregateWeighted(t *testing.T) {
	total, components := AggregateWeighted(0.8, 0.2, 0.5, 1.0, 1.0, 1.0)
	assert.InDelta(t, (0.8+0.2+0.5)/3.0, total, 0.001)
	assert.Equal(t, 0.8, components.Graph)
	assert.Equal(t, 0.2, components.Semantic)
	assert.Equal(t, 0.5, components.Keyword)

	total2, _ := AggregateWeighted(0.8, 0.2, 0.5, 0, 0, 0)
	assert.Equal(t, 0.8, total2)
}

func TestNormalizeMap(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	scores := map[uuid.UUID]float64{id1: 0.5, id2: 1.0}
	normalized := NormalizeMap(scores)
	assert.InDelta(t, 0.5, normalized[id1], 0.001)
	assert.InDelta(t, 1.0, normalized[id2], 0.001)

	zero := map[uuid.UUID]float64{id1: 0.0, id2: -1.0}
	normalizedZero := NormalizeMap(zero)
	assert.Equal(t, 0.0, normalizedZero[id1])
	assert.Equal(t, -1.0, normalizedZero[id2])
}

func TestTraversalService_RunBFS(t *testing.T) {
	ctx := context.Background()
	startID := uuid.New()
	targetID := uuid.New()

	loader := new(MockNeighborLoader)
	loader.On("GetNeighbors", ctx, startID).Return([]Edge{
		{From: startID, To: targetID, Weight: 0.9},
	}, nil)
	loader.On("GetNeighbors", ctx, targetID).Return([]Edge{}, nil)

	svc := NewTraversalService(loader, 2, 0.5, "max", false)
	result := svc.RunBFS(ctx, startID)

	assert.Contains(t, result, targetID)
	assert.InDelta(t, 0.9, result[targetID].weight, 0.001)
}

func TestTraversalService_SetKeywordMatcher(t *testing.T) {
	loader := new(MockNeighborLoader)
	svc := NewTraversalService(loader, 1, 0.5, "max", false)

	matcher := &NoOpKeywordMatcher{}
	svc.SetKeywordMatcher(matcher)
	assert.Equal(t, matcher, svc.keywordMatcher)
}

func TestNewTraversalServiceWithWeights_InvalidAggregation(t *testing.T) {
	loader := new(MockNeighborLoader)
	svc := NewTraversalServiceWithWeights(loader, 2, 0.5, "invalid", false, 0.5, 0.3, 0.2)

	assert.Equal(t, "max", svc.aggregation)
}

func TestTraversalService_GetSuggestions_WithKeywordMatcher(t *testing.T) {
	ctx := context.Background()
	startID := uuid.New()
	targetID := uuid.New()

	loader := new(MockNeighborLoader)
	loader.On("GetNeighbors", ctx, startID).Return([]Edge{
		{From: startID, To: targetID, Weight: 0.8},
	}, nil)
	loader.On("GetNeighbors", ctx, targetID).Return([]Edge{}, nil)

	svc := NewTraversalServiceWithWeights(loader, 2, 0.5, "max", false, 0.7, 0.0, 0.3)
	svc.SetKeywordMatcher(&mockKeywordMatcher{})

	results, err := svc.GetSuggestions(ctx, startID, 10)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, targetID, results[0].NodeID)
}

type mockKeywordMatcher struct{}

func (m *mockKeywordMatcher) Match(ctx context.Context, sourceID uuid.UUID, candidateIDs []uuid.UUID) (map[uuid.UUID]float64, error) {
	scores := make(map[uuid.UUID]float64, len(candidateIDs))
	for _, id := range candidateIDs {
		scores[id] = 0.5
	}
	return scores, nil
}

var _ KeywordMatcher = (*mockKeywordMatcher)(nil)
