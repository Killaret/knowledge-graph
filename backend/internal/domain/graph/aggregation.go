package graph

// AggregateMax — maximum path weight strategy
func AggregateMax(currentWeight, newWeight float64) float64 {
	if newWeight > currentWeight {
		return newWeight
	}
	return currentWeight
}

// AggregateSum — weight summation strategy
func AggregateSum(currentWeight, newWeight float64) float64 {
	return currentWeight + newWeight
}

// AggregateWeighted — combined score from three components
// alpha — graph component weight
// beta — semantic component weight
// gamma — keyword component weight
func AggregateWeighted(graphWeight, semanticWeight, keywordWeight, alpha, beta, gamma float64) (total float64, components SuggestionComponents) {
	components = SuggestionComponents{
		Graph:    graphWeight,
		Semantic: semanticWeight,
		Keyword:  keywordWeight,
	}

	// Normalize weights (if sum > 0)
	totalWeight := alpha + beta + gamma
	if totalWeight > 0 {
		total = (alpha*graphWeight + beta*semanticWeight + gamma*keywordWeight) / totalWeight
	} else {
		total = graphWeight // fallback to graph score
	}

	return total, components
}
