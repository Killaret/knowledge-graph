package graph

import (
	"context"
	"sort"

	"github.com/google/uuid"
)

// TraversalService — сервис-оркестратор для рекомендаций
type TraversalService struct {
	loader        NeighborLoader
	keywordMatcher KeywordMatcher
	depth         int
	decay         float64
	aggregation   string // "max" или "sum"
	normalize     bool
	// Веса компонентов (alpha + beta + gamma = 1.0)
	alpha float64 // вес графового компонента
	beta  float64 // вес семантического компонента
	gamma float64 // вес keyword компонента
}

// NewTraversalService — конструктор с дефолтными весами
func NewTraversalService(
	loader NeighborLoader,
	depth int,
	decay float64,
	aggregation string,
	normalize bool,
) *TraversalService {
	return NewTraversalServiceWithWeights(
		loader, depth, decay, aggregation, normalize,
		1.0, 0.0, 0.0, // по умолчанию только графовый компонент
	)
}

// NewTraversalServiceWithWeights — конструктор с явными весами компонентов
func NewTraversalServiceWithWeights(
	loader NeighborLoader,
	depth int,
	decay float64,
	aggregation string,
	normalize bool,
	alpha, beta, gamma float64,
) *TraversalService {
	if aggregation != "max" && aggregation != "sum" {
		aggregation = "max"
	}

	// Нормализация весов
	total := alpha + beta + gamma
	if total > 0 {
		alpha /= total
		beta /= total
		gamma /= total
	} else {
		alpha = 1.0
		beta = 0.0
		gamma = 0.0
	}

	return &TraversalService{
		loader:         loader,
		keywordMatcher: &NoOpKeywordMatcher{}, // по умолчанию заглушка
		depth:          depth,
		decay:          decay,
		aggregation:    aggregation,
		normalize:      normalize,
		alpha:          alpha,
		beta:           beta,
		gamma:          gamma,
	}
}

// SetKeywordMatcher — установить реализацию KeywordMatcher
func (s *TraversalService) SetKeywordMatcher(matcher KeywordMatcher) {
	s.keywordMatcher = matcher
}

// GetSuggestions — основной метод получения рекомендаций
func (s *TraversalService) GetSuggestions(ctx context.Context, startID uuid.UUID, topN int) ([]SuggestionResult, error) {
	// 1. Графовый компонент (BFS)
	graphWeights := runBFS(ctx, startID, s.loader, s.depth, s.decay, s.aggregation)
	if s.normalize {
		graphWeights = NormalizeWeights(graphWeights)
	}

	// 2. Получаем список кандидатов
	candidateIDs := make([]uuid.UUID, 0, len(graphWeights))
	for id := range graphWeights {
		candidateIDs = append(candidateIDs, id)
	}

	// 3. Keyword компонент (если gamma > 0)
	keywordScores := make(map[uuid.UUID]float64)
	if s.gamma > 0 && s.keywordMatcher != nil {
		var err error
		keywordScores, err = s.keywordMatcher.Match(ctx, startID, candidateIDs)
		if err != nil {
			// При ошибке продолжаем без keyword-компонента
			keywordScores = make(map[uuid.UUID]float64)
		}
		if s.normalize {
			keywordScores = NormalizeMap(keywordScores)
		}
	}

	// 4. Комбинирование компонентов
	results := make([]SuggestionResult, 0, len(graphWeights))
	for id, graphPath := range graphWeights {
		// Получаем keyword score (если нет — 0)
		keywordScore := keywordScores[id]

		// Пока semantic score не реализован — 0
		semanticScore := 0.0

		// Агрегация
		totalScore, components := AggregateWeighted(
			graphPath.weight, semanticScore, keywordScore,
			s.alpha, s.beta, s.gamma,
		)

		results = append(results, SuggestionResult{
			NodeID:        id,
			Score:         totalScore,
			GraphScore:    components.Graph,
			SemanticScore: components.Semantic,
			KeywordScore:  components.Keyword,
		})
	}

	// 5. Сортировка по убыванию итогового скора
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 6. Ограничение topN
	if len(results) > topN {
		results = results[:topN]
	}

	return results, nil
}
