package recommendation

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// KeywordRepositoryWithWeights — интерфейс для получения ключевых слов с весами
type KeywordRepositoryWithWeights interface {
	GetKeywordsWithWeights(ctx context.Context, noteID uuid.UUID) (map[string]float64, error)
	GetKeywordsBatchWithWeights(ctx context.Context, noteIDs []uuid.UUID) (map[uuid.UUID]map[string]float64, error)
}

// KeywordMatcherImpl — реализация KeywordMatcher с использованием KeywordSimilarity
type KeywordMatcherImpl struct {
	repo             KeywordRepositoryWithWeights
	similarity       KeywordSimilarity
	requiresWeights  bool
}

// NewKeywordMatcherImpl создаёт новый KeywordMatcherImpl
func NewKeywordMatcherImpl(repo KeywordRepositoryWithWeights, similarity KeywordSimilarity) *KeywordMatcherImpl {
	return &KeywordMatcherImpl{
		repo:            repo,
		similarity:      similarity,
		requiresWeights: similarity.RequiresWeights(),
	}
}

// Match возвращает similarity scores между sourceID и candidateIDs
// based on keyword similarity using the configured strategy
func (k *KeywordMatcherImpl) Match(ctx context.Context, sourceID uuid.UUID, candidateIDs []uuid.UUID) (map[uuid.UUID]float64, error) {
	if len(candidateIDs) == 0 {
		return make(map[uuid.UUID]float64), nil
	}

	// Получаем ключевые слова для source
	sourceKeywordsMap, err := k.repo.GetKeywordsWithWeights(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get keywords for source %s: %w", sourceID, err)
	}

	// Получаем ключевые слова для всех кандидатов
	allIDs := append([]uuid.UUID{sourceID}, candidateIDs...)
	allKeywordsMap, err := k.repo.GetKeywordsBatchWithWeights(ctx, allIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get keywords batch: %w", err)
	}

	// Преобразуем мапу source в слайс строк и мапу весов
	sourceKeywords := make([]string, 0, len(sourceKeywordsMap))
	for kw := range sourceKeywordsMap {
		sourceKeywords = append(sourceKeywords, kw)
	}

	// Вычисляем сходство для каждого кандидата
	results := make(map[uuid.UUID]float64, len(candidateIDs))
	for _, candidateID := range candidateIDs {
		candidateKeywordsMap := allKeywordsMap[candidateID]
		
		// Преобразуем в слайс строк
		candidateKeywords := make([]string, 0, len(candidateKeywordsMap))
		for kw := range candidateKeywordsMap {
			candidateKeywords = append(candidateKeywords, kw)
		}

		// Вычисляем сходство
		var similarity float64
		if k.requiresWeights {
			similarity = k.similarity.Similarity(sourceKeywords, candidateKeywords, sourceKeywordsMap, candidateKeywordsMap)
		} else {
			similarity = k.similarity.Similarity(sourceKeywords, candidateKeywords, nil, nil)
		}
		
		results[candidateID] = similarity
	}

	return results, nil
}
