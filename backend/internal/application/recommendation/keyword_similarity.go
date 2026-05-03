package recommendation

import (
	"fmt"
	"math"
)

// KeywordSimilarity — интерфейс для вычисления сходства между наборами ключевых слов
type KeywordSimilarity interface {
	// Similarity вычисляет сходство между source и target keywords
	// weightsSource и weightsTarget — веса ключевых слов (для weighted методов)
	// Если веса не требуются, передаются nil или пустые мапы
	Similarity(source []string, target []string, weightsSource map[string]float64, weightsTarget map[string]float64) float64

	// RequiresWeights возвращает true, если метод требует веса ключевых слов
	RequiresWeights() bool
}

// JaccardSimilarity — классический коэффициент Жаккара: |A ∩ B| / |A ∪ B|
type JaccardSimilarity struct{}

// Similarity implements KeywordSimilarity
func (j *JaccardSimilarity) Similarity(source []string, target []string, weightsSource map[string]float64, weightsTarget map[string]float64) float64 {
	if len(source) == 0 && len(target) == 0 {
		return 1.0 // Оба пустые — полное сходство
	}
	if len(source) == 0 || len(target) == 0 {
		return 0.0 // Один пустой — нет сходства
	}

	set1 := make(map[string]struct{}, len(source))
	for _, k := range source {
		set1[k] = struct{}{}
	}

	set2 := make(map[string]struct{}, len(target))
	for _, k := range target {
		set2[k] = struct{}{}
	}

	intersection := 0
	for k := range set1 {
		if _, ok := set2[k]; ok {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// RequiresWeights implements KeywordSimilarity
func (j *JaccardSimilarity) RequiresWeights() bool {
	return false
}

// OverlapSimilarity — коэффициент перекрытия: |A ∩ B| / min(|A|, |B|)
type OverlapSimilarity struct{}

// Similarity implements KeywordSimilarity
func (o *OverlapSimilarity) Similarity(source []string, target []string, weightsSource map[string]float64, weightsTarget map[string]float64) float64 {
	if len(source) == 0 && len(target) == 0 {
		return 1.0
	}
	if len(source) == 0 || len(target) == 0 {
		return 0.0
	}

	set1 := make(map[string]struct{}, len(source))
	for _, k := range source {
		set1[k] = struct{}{}
	}

	intersection := 0
	for _, k := range target {
		if _, ok := set1[k]; ok {
			intersection++
		}
	}

	minLen := len(source)
	if len(target) < minLen {
		minLen = len(target)
	}

	if minLen == 0 {
		return 0.0
	}

	return float64(intersection) / float64(minLen)
}

// RequiresWeights implements KeywordSimilarity
func (o *OverlapSimilarity) RequiresWeights() bool {
	return false
}

// TverskySimilarity — индекс Тверски с параметрами alpha и beta
// Формула: |A ∩ B| / (|A ∩ B| + alpha*|A\B| + beta*|B\A|)
// При alpha = beta = 0.5 — это симметричный индекс Дайса
// При alpha = beta = 1 — это коэффициент Жаккара
type TverskySimilarity struct {
	Alpha float64
	Beta  float64
}

// Similarity implements KeywordSimilarity
func (t *TverskySimilarity) Similarity(source []string, target []string, weightsSource map[string]float64, weightsTarget map[string]float64) float64 {
	if len(source) == 0 && len(target) == 0 {
		return 1.0
	}
	if len(source) == 0 || len(target) == 0 {
		return 0.0
	}

	set1 := make(map[string]struct{}, len(source))
	for _, k := range source {
		set1[k] = struct{}{}
	}

	set2 := make(map[string]struct{}, len(target))
	for _, k := range target {
		set2[k] = struct{}{}
	}

	intersection := 0
	diff1 := 0 // В set1, но не в set2
	diff2 := 0 // В set2, но не в set1

	for k := range set1 {
		if _, ok := set2[k]; ok {
			intersection++
		} else {
			diff1++
		}
	}

	for k := range set2 {
		if _, ok := set1[k]; !ok {
			diff2++
		}
	}

	denominator := float64(intersection) + t.Alpha*float64(diff1) + t.Beta*float64(diff2)
	if denominator == 0 {
		return 0.0
	}

	return float64(intersection) / denominator
}

// RequiresWeights implements KeywordSimilarity
func (t *TverskySimilarity) RequiresWeights() bool {
	return false
}

// WeightedJaccardSimilarity — взвешенный Жаккард: sum(min(w1, w2)) / sum(max(w1, w2))
// Если веса не предоставлены, падает обратно к обычному Жаккару
type WeightedJaccardSimilarity struct{}

// Similarity implements KeywordSimilarity
func (w *WeightedJaccardSimilarity) Similarity(source []string, target []string, weightsSource map[string]float64, weightsTarget map[string]float64) float64 {
	// Если веса не предоставлены — используем обычный Жаккард
	if len(weightsSource) == 0 || len(weightsTarget) == 0 {
		jaccard := &JaccardSimilarity{}
		return jaccard.Similarity(source, target, nil, nil)
	}

	if len(source) == 0 && len(target) == 0 {
		return 1.0
	}
	if len(source) == 0 || len(target) == 0 {
		return 0.0
	}

	// Собираем все уникальные ключевые слова
	allKeywords := make(map[string]struct{})
	for _, k := range source {
		allKeywords[k] = struct{}{}
	}
	for _, k := range target {
		allKeywords[k] = struct{}{}
	}

	minSum := 0.0
	maxSum := 0.0

	for k := range allKeywords {
		w1, ok1 := weightsSource[k]
		w2, ok2 := weightsTarget[k]

		if !ok1 {
			w1 = 0.0
		}
		if !ok2 {
			w2 = 0.0
		}

		if w1 < w2 {
			minSum += w1
			maxSum += w2
		} else {
			minSum += w2
			maxSum += w1
		}
	}

	if maxSum == 0 {
		return 0.0
	}

	return minSum / maxSum
}

// RequiresWeights implements KeywordSimilarity
func (w *WeightedJaccardSimilarity) RequiresWeights() bool {
	return true
}

// CosineSimilarity — косинусное сходство между векторами весов
// При отсутствии весов использует единичные веса
type CosineSimilarity struct{}

// Similarity implements KeywordSimilarity
func (c *CosineSimilarity) Similarity(source []string, target []string, weightsSource map[string]float64, weightsTarget map[string]float64) float64 {
	if len(source) == 0 && len(target) == 0 {
		return 1.0
	}
	if len(source) == 0 || len(target) == 0 {
		return 0.0
	}

	// Собираем все уникальные ключевые слова
	allKeywords := make(map[string]struct{})
	for _, k := range source {
		allKeywords[k] = struct{}{}
	}
	for _, k := range target {
		allKeywords[k] = struct{}{}
	}

	// Если веса не предоставлены — используем единичные веса
	useUnitWeights := len(weightsSource) == 0 || len(weightsTarget) == 0

	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	for k := range allKeywords {
		var w1, w2 float64

		if useUnitWeights {
			// Единичные веса (1 если слово присутствует, 0 если нет)
			for _, s := range source {
				if s == k {
					w1 = 1.0
					break
				}
			}
			for _, t := range target {
				if t == k {
					w2 = 1.0
					break
				}
			}
		} else {
			var ok1, ok2 bool
			w1, ok1 = weightsSource[k]
			w2, ok2 = weightsTarget[k]
			if !ok1 {
				w1 = 0.0
			}
			if !ok2 {
				w2 = 0.0
			}
		}

		dotProduct += w1 * w2
		norm1 += w1 * w1
		norm2 += w2 * w2
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	cosine := dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
	
	// Защита от численных ошибок
	if cosine > 1.0 {
		return 1.0
	}
	if cosine < 0.0 {
		return 0.0
	}

	return cosine
}

// RequiresWeights implements KeywordSimilarity
func (c *CosineSimilarity) RequiresWeights() bool {
	return true
}

// NewKeywordSimilarity — фабричная функция для создания стратегий сходства
// method: "jaccard", "overlap", "tversky", "weighted_jaccard", "cosine"
// tverskyAlpha, tverskyBeta — используются только для метода "tversky"
func NewKeywordSimilarity(method string, tverskyAlpha, tverskyBeta float64) (KeywordSimilarity, error) {
	switch method {
	case "jaccard":
		return &JaccardSimilarity{}, nil
	case "overlap":
		return &OverlapSimilarity{}, nil
	case "tversky":
		return &TverskySimilarity{
			Alpha: tverskyAlpha,
			Beta:  tverskyBeta,
		}, nil
	case "weighted_jaccard":
		return &WeightedJaccardSimilarity{}, nil
	case "cosine":
		return &CosineSimilarity{}, nil
	default:
		return nil, fmt.Errorf("unknown keyword similarity method: %s (valid: jaccard, overlap, tversky, weighted_jaccard, cosine)", method)
	}
}

// Ensure all strategies implement KeywordSimilarity
var _ KeywordSimilarity = (*JaccardSimilarity)(nil)
var _ KeywordSimilarity = (*OverlapSimilarity)(nil)
var _ KeywordSimilarity = (*TverskySimilarity)(nil)
var _ KeywordSimilarity = (*WeightedJaccardSimilarity)(nil)
var _ KeywordSimilarity = (*CosineSimilarity)(nil)
