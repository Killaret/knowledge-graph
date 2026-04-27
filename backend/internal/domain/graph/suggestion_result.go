package graph

import (
	"github.com/google/uuid"
)

// SuggestionResult — результат рекомендации с разложением по компонентам
type SuggestionResult struct {
	NodeID        uuid.UUID
	Title         string
	Score         float64 // итоговый комбинированный скор (alpha*Graph + beta*Semantic + gamma*Keyword)
	GraphScore    float64 // вклад графовых связей (alpha)
	SemanticScore float64 // вклад семантической близости (beta)
	KeywordScore  float64 // вклад ключевых слов (gamma)
}

// SuggestionComponents — компоненты скора для комбинирования
type SuggestionComponents struct {
	Graph    float64
	Semantic float64
	Keyword  float64
}
