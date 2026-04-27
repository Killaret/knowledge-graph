package graph

// AggregateMax — стратегия максимального веса пути
func AggregateMax(currentWeight, newWeight float64) float64 {
	if newWeight > currentWeight {
		return newWeight
	}
	return currentWeight
}

// AggregateSum — стратегия суммирования весов
func AggregateSum(currentWeight, newWeight float64) float64 {
	return currentWeight + newWeight
}

// AggregateWeighted — комбинированный скор из трёх компонентов
// alpha — вес графового компонента
// beta — вес семантического компонента  
// gamma — вес keyword компонента
func AggregateWeighted(graphWeight, semanticWeight, keywordWeight, alpha, beta, gamma float64) (total float64, components SuggestionComponents) {
	components = SuggestionComponents{
		Graph:    graphWeight,
		Semantic: semanticWeight,
		Keyword:  keywordWeight,
	}

	// Нормализация весов (если сумма > 0)
	totalWeight := alpha + beta + gamma
	if totalWeight > 0 {
		total = (alpha*graphWeight + beta*semanticWeight + gamma*keywordWeight) / totalWeight
	} else {
		total = graphWeight // fallback на графовый скор
	}

	return total, components
}
