package graph

import (
	"github.com/google/uuid"
)

// SuggestionResult — recommendation result with component breakdown
type SuggestionResult struct {
	NodeID        uuid.UUID
	Title         string
	Score         float64 // final combined score (alpha*Graph + beta*Semantic + gamma*Keyword)
	GraphScore    float64 // graph connections contribution (alpha)
	SemanticScore float64 // semantic similarity contribution (beta)
	KeywordScore  float64 // keyword contribution (gamma)
}

// SuggestionComponents — score components for combining
type SuggestionComponents struct {
	Graph    float64
	Semantic float64
	Keyword  float64
}
