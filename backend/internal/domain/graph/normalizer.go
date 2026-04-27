package graph

import "github.com/google/uuid"

// NormalizeWeights — нормализация весов по максимальному значению
func NormalizeWeights(weights map[uuid.UUID]weightedPath) map[uuid.UUID]weightedPath {
	maxW := 0.0
	for _, wp := range weights {
		if wp.weight > maxW {
			maxW = wp.weight
		}
	}

	if maxW <= 0 {
		return weights
	}

	normalized := make(map[uuid.UUID]weightedPath, len(weights))
	for id, wp := range weights {
		normalized[id] = weightedPath{
			weight: wp.weight / maxW,
			depth:  wp.depth,
		}
	}

	return normalized
}

// NormalizeMap — нормализация простой мапы float64
func NormalizeMap(scores map[uuid.UUID]float64) map[uuid.UUID]float64 {
	maxScore := 0.0
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}

	if maxScore <= 0 {
		return scores
	}

	normalized := make(map[uuid.UUID]float64, len(scores))
	for id, score := range scores {
		normalized[id] = score / maxScore
	}

	return normalized
}
