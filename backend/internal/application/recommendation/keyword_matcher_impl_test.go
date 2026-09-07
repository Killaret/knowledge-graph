package recommendation

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockKeywordRepo struct {
	mock.Mock
}

func (m *mockKeywordRepo) GetKeywordsWithWeights(ctx context.Context, noteID uuid.UUID) (map[string]float64, error) {
	args := m.Called(ctx, noteID)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *mockKeywordRepo) GetKeywordsBatchWithWeights(ctx context.Context, noteIDs []uuid.UUID) (map[uuid.UUID]map[string]float64, error) {
	args := m.Called(ctx, noteIDs)
	return args.Get(0).(map[uuid.UUID]map[string]float64), args.Error(1)
}

func TestKeywordMatcherImpl_Match_EmptyCandidates(t *testing.T) {
	repo := new(mockKeywordRepo)
	sim, _ := NewKeywordSimilarity("jaccard", 0, 0)
	matcher := NewKeywordMatcherImpl(repo, sim)

	result, err := matcher.Match(context.Background(), uuid.New(), nil)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestKeywordMatcherImpl_Match_WithWeights(t *testing.T) {
	sourceID := uuid.New()
	candidateID := uuid.New()
	repo := new(mockKeywordRepo)
	sim, _ := NewKeywordSimilarity("weighted_jaccard", 0, 0)
	matcher := NewKeywordMatcherImpl(repo, sim)

	keywords := map[string]float64{"go": 1.0, "test": 0.5}
	batch := map[uuid.UUID]map[string]float64{
		sourceID:    keywords,
		candidateID: {"go": 0.8, "test": 0.3},
	}

	repo.On("GetKeywordsWithWeights", mock.Anything, sourceID).Return(keywords, nil)
	repo.On("GetKeywordsBatchWithWeights", mock.Anything, mock.Anything).Return(batch, nil)

	result, err := matcher.Match(context.Background(), sourceID, []uuid.UUID{candidateID})
	assert.NoError(t, err)
	assert.Contains(t, result, candidateID)
}

func TestKeywordMatcherImpl_Match_WithoutWeights(t *testing.T) {
	sourceID := uuid.New()
	candidateID := uuid.New()
	repo := new(mockKeywordRepo)
	sim, _ := NewKeywordSimilarity("jaccard", 0, 0)
	matcher := NewKeywordMatcherImpl(repo, sim)

	keywords := map[string]float64{"go": 1.0}
	batch := map[uuid.UUID]map[string]float64{
		sourceID:    keywords,
		candidateID: {"go": 1.0},
	}

	repo.On("GetKeywordsWithWeights", mock.Anything, sourceID).Return(keywords, nil)
	repo.On("GetKeywordsBatchWithWeights", mock.Anything, mock.Anything).Return(batch, nil)

	result, err := matcher.Match(context.Background(), sourceID, []uuid.UUID{candidateID})
	assert.NoError(t, err)
	assert.Contains(t, result, candidateID)
}

func TestKeywordMatcherImpl_Match_RepoError(t *testing.T) {
	repo := new(mockKeywordRepo)
	sim, _ := NewKeywordSimilarity("jaccard", 0, 0)
	matcher := NewKeywordMatcherImpl(repo, sim)

	repo.On("GetKeywordsWithWeights", mock.Anything, mock.Anything).Return(map[string]float64{}, assert.AnError)

	_, err := matcher.Match(context.Background(), uuid.New(), []uuid.UUID{uuid.New()})
	assert.Error(t, err)
}
