package recommendation

import (
	"math"
	"testing"
)

func TestJaccardSimilarity(t *testing.T) {
	j := &JaccardSimilarity{}

	tests := []struct {
		name      string
		source    []string
		target    []string
		expected  float64
		tolerance float64
	}{
		{
			name:      "identical sets",
			source:    []string{"a", "b", "c"},
			target:    []string{"a", "b", "c"},
			expected:  1.0,
			tolerance: 0.0001,
		},
		{
			name:      "completely different",
			source:    []string{"a", "b", "c"},
			target:    []string{"x", "y", "z"},
			expected:  0.0,
			tolerance: 0.0001,
		},
		{
			name:      "partial overlap",
			source:    []string{"a", "b", "c"},
			target:    []string{"b", "c", "d"},
			expected:  0.5, // intersection=2, union=4
			tolerance: 0.0001,
		},
		{
			name:      "one empty source",
			source:    []string{},
			target:    []string{"a", "b"},
			expected:  0.0,
			tolerance: 0.0001,
		},
		{
			name:      "one empty target",
			source:    []string{"a", "b"},
			target:    []string{},
			expected:  0.0,
			tolerance: 0.0001,
		},
		{
			name:      "both empty",
			source:    []string{},
			target:    []string{},
			expected:  1.0,
			tolerance: 0.0001,
		},
		{
			name:      "subset",
			source:    []string{"a", "b"},
			target:    []string{"a", "b", "c", "d"},
			expected:  0.5, // intersection=2, union=4
			tolerance: 0.0001,
		},
		{
			name:      "duplicates in source (should be ignored)",
			source:    []string{"a", "a", "b", "b"},
			target:    []string{"a", "b", "c"},
			expected:  0.667, // intersection=2, union=3
			tolerance: 0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := j.Similarity(tt.source, tt.target, nil, nil)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("JaccardSimilarity() = %v, want %v (tolerance: %v)", result, tt.expected, tt.tolerance)
			}
		})
	}
}

func TestJaccardSimilarity_RequiresWeights(t *testing.T) {
	j := &JaccardSimilarity{}
	if j.RequiresWeights() {
		t.Error("JaccardSimilarity.RequiresWeights() should return false")
	}
}

func TestOverlapSimilarity(t *testing.T) {
	o := &OverlapSimilarity{}

	tests := []struct {
		name      string
		source    []string
		target    []string
		expected  float64
		tolerance float64
	}{
		{
			name:      "identical sets",
			source:    []string{"a", "b", "c"},
			target:    []string{"a", "b", "c"},
			expected:  1.0,
			tolerance: 0.0001,
		},
		{
			name:      "completely different",
			source:    []string{"a", "b", "c"},
			target:    []string{"x", "y", "z"},
			expected:  0.0,
			tolerance: 0.0001,
		},
		{
			name:      "partial overlap",
			source:    []string{"a", "b", "c"},
			target:    []string{"b", "c", "d"},
			expected:  0.667, // intersection=2, min=3
			tolerance: 0.01,
		},
		{
			name:      "subset",
			source:    []string{"a", "b"},
			target:    []string{"a", "b", "c", "d"},
			expected:  1.0, // intersection=2, min(len(source), len(target))=2
			tolerance: 0.0001,
		},
		{
			name:      "one empty",
			source:    []string{},
			target:    []string{"a", "b"},
			expected:  0.0,
			tolerance: 0.0001,
		},
		{
			name:      "both empty",
			source:    []string{},
			target:    []string{},
			expected:  1.0,
			tolerance: 0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := o.Similarity(tt.source, tt.target, nil, nil)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("OverlapSimilarity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestOverlapSimilarity_RequiresWeights(t *testing.T) {
	o := &OverlapSimilarity{}
	if o.RequiresWeights() {
		t.Error("OverlapSimilarity.RequiresWeights() should return false")
	}
}

func TestTverskySimilarity(t *testing.T) {
	// Test with alpha=beta=1 (should be equivalent to Jaccard)
	tverskyJaccard := &TverskySimilarity{Alpha: 1.0, Beta: 1.0}

	// Test with alpha=beta=0.5 (should be Dice coefficient)
	tverskyDice := &TverskySimilarity{Alpha: 0.5, Beta: 0.5}

	tests := []struct {
		name      string
		sim       *TverskySimilarity
		source    []string
		target    []string
		expected  float64
		tolerance float64
	}{
		{
			name:      "jaccard-like: identical",
			sim:       tverskyJaccard,
			source:    []string{"a", "b", "c"},
			target:    []string{"a", "b", "c"},
			expected:  1.0,
			tolerance: 0.0001,
		},
		{
			name:      "jaccard-like: partial",
			sim:       tverskyJaccard,
			source:    []string{"a", "b", "c"},
			target:    []string{"b", "c", "d"},
			expected:  0.5, // intersection=2, diff1=1, diff2=1, denom=2+1+1=4
			tolerance: 0.0001,
		},
		{
			name:      "dice-like: partial",
			sim:       tverskyDice,
			source:    []string{"a", "b", "c"},
			target:    []string{"b", "c", "d"},
			expected:  0.667, // intersection=2, diff1=1*0.5=0.5, diff2=1*0.5=0.5, denom=2+0.5+0.5=3
			tolerance: 0.01,
		},
		{
			name:      "empty sets",
			sim:       tverskyJaccard,
			source:    []string{},
			target:    []string{},
			expected:  1.0,
			tolerance: 0.0001,
		},
		{
			name:      "one empty",
			sim:       tverskyJaccard,
			source:    []string{"a", "b"},
			target:    []string{},
			expected:  0.0,
			tolerance: 0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.sim.Similarity(tt.source, tt.target, nil, nil)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("TverskySimilarity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTverskySimilarity_RequiresWeights(t *testing.T) {
	tvrs := &TverskySimilarity{Alpha: 0.5, Beta: 0.5}
	if tvrs.RequiresWeights() {
		t.Error("TverskySimilarity.RequiresWeights() should return false")
	}
}

func TestWeightedJaccardSimilarity(t *testing.T) {
	w := &WeightedJaccardSimilarity{}

	tests := []struct {
		name          string
		source        []string
		target        []string
		weightsSource map[string]float64
		weightsTarget map[string]float64
		expected      float64
		tolerance     float64
	}{
		{
			name:          "with weights: identical",
			source:        []string{"a", "b"},
			target:        []string{"a", "b"},
			weightsSource: map[string]float64{"a": 1.0, "b": 2.0},
			weightsTarget: map[string]float64{"a": 1.0, "b": 2.0},
			expected:      1.0,
			tolerance:     0.0001,
		},
		{
			name:          "with weights: partial",
			source:        []string{"a", "b", "c"},
			target:        []string{"b", "c", "d"},
			weightsSource: map[string]float64{"a": 1.0, "b": 2.0, "c": 3.0},
			weightsTarget: map[string]float64{"b": 2.0, "c": 2.5, "d": 1.0},
			// min(w1,w2): a=0, b=2, c=2.5, d=0 = 4.5
			// max(w1,w2): a=1, b=2, c=3, d=1 = 7
			// Result: 4.5/7 = 0.643
			expected:  0.643,
			tolerance: 0.01,
		},
		{
			name:          "no weights: falls back to jaccard",
			source:        []string{"a", "b", "c"},
			target:        []string{"b", "c", "d"},
			weightsSource: map[string]float64{},
			weightsTarget: map[string]float64{},
			expected:      0.5, // Jaccard
			tolerance:     0.0001,
		},
		{
			name:          "empty weights: falls back to jaccard",
			source:        []string{"a", "b"},
			target:        []string{"a", "b", "c"},
			weightsSource: nil,
			weightsTarget: map[string]float64{"a": 1.0, "b": 1.0, "c": 1.0},
			// Falls back to Jaccard: intersection=2, union=3 -> 2/3 = 0.667
			expected:  0.667,
			tolerance: 0.01,
		},
		{
			name:          "with weights: completely different",
			source:        []string{"a", "b"},
			target:        []string{"c", "d"},
			weightsSource: map[string]float64{"a": 1.0, "b": 2.0},
			weightsTarget: map[string]float64{"c": 1.0, "d": 2.0},
			expected:      0.0,
			tolerance:     0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := w.Similarity(tt.source, tt.target, tt.weightsSource, tt.weightsTarget)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("WeightedJaccardSimilarity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWeightedJaccardSimilarity_RequiresWeights(t *testing.T) {
	w := &WeightedJaccardSimilarity{}
	if !w.RequiresWeights() {
		t.Error("WeightedJaccardSimilarity.RequiresWeights() should return true")
	}
}

func TestCosineSimilarity(t *testing.T) {
	c := &CosineSimilarity{}

	tests := []struct {
		name          string
		source        []string
		target        []string
		weightsSource map[string]float64
		weightsTarget map[string]float64
		expected      float64
		tolerance     float64
	}{
		{
			name:          "unit weights: identical",
			source:        []string{"a", "b", "c"},
			target:        []string{"a", "b", "c"},
			weightsSource: map[string]float64{},
			weightsTarget: map[string]float64{},
			expected:      1.0,
			tolerance:     0.0001,
		},
		{
			name:          "unit weights: orthogonal (no overlap)",
			source:        []string{"a", "b"},
			target:        []string{"c", "d"},
			weightsSource: map[string]float64{},
			weightsTarget: map[string]float64{},
			expected:      0.0,
			tolerance:     0.0001,
		},
		{
			name:          "unit weights: partial overlap",
			source:        []string{"a", "b", "c"},
			target:        []string{"b", "c", "d"},
			weightsSource: map[string]float64{},
			weightsTarget: map[string]float64{},
			// dot=2 (b and c), norm1=sqrt(3), norm2=sqrt(3) -> 2/3 = 0.667
			expected:  0.667,
			tolerance: 0.01,
		},
		{
			name:          "with weights: simple case",
			source:        []string{"a", "b"},
			target:        []string{"a", "b"},
			weightsSource: map[string]float64{"a": 1.0, "b": 0.0},
			weightsTarget: map[string]float64{"a": 1.0, "b": 0.0},
			expected:      1.0,
			tolerance:     0.0001,
		},
		{
			name:          "with weights: orthogonal",
			source:        []string{"a"},
			target:        []string{"b"},
			weightsSource: map[string]float64{"a": 1.0},
			weightsTarget: map[string]float64{"b": 1.0},
			expected:      0.0,
			tolerance:     0.0001,
		},
		{
			name:          "empty sets",
			source:        []string{},
			target:        []string{},
			weightsSource: map[string]float64{},
			weightsTarget: map[string]float64{},
			expected:      1.0,
			tolerance:     0.0001,
		},
		{
			name:          "one empty",
			source:        []string{"a", "b"},
			target:        []string{},
			weightsSource: map[string]float64{},
			weightsTarget: map[string]float64{},
			expected:      0.0,
			tolerance:     0.0001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Similarity(tt.source, tt.target, tt.weightsSource, tt.weightsTarget)
			if math.Abs(result-tt.expected) > tt.tolerance {
				t.Errorf("CosineSimilarity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCosineSimilarity_RequiresWeights(t *testing.T) {
	c := &CosineSimilarity{}
	if !c.RequiresWeights() {
		t.Error("CosineSimilarity.RequiresWeights() should return true")
	}
}

func TestNewKeywordSimilarity(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		tverskyAlpha float64
		tverskyBeta  float64
		wantErr      bool
		expectedType string
	}{
		{
			name:         "jaccard",
			method:       "jaccard",
			wantErr:      false,
			expectedType: "*recommendation.JaccardSimilarity",
		},
		{
			name:         "overlap",
			method:       "overlap",
			wantErr:      false,
			expectedType: "*recommendation.OverlapSimilarity",
		},
		{
			name:         "tversky",
			method:       "tversky",
			tverskyAlpha: 0.5,
			tverskyBeta:  0.5,
			wantErr:      false,
			expectedType: "*recommendation.TverskySimilarity",
		},
		{
			name:         "weighted_jaccard",
			method:       "weighted_jaccard",
			wantErr:      false,
			expectedType: "*recommendation.WeightedJaccardSimilarity",
		},
		{
			name:         "cosine",
			method:       "cosine",
			wantErr:      false,
			expectedType: "*recommendation.CosineSimilarity",
		},
		{
			name:    "unknown method",
			method:  "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim, err := NewKeywordSimilarity(tt.method, tt.tverskyAlpha, tt.tverskyBeta)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKeywordSimilarity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && sim == nil {
				t.Error("NewKeywordSimilarity() returned nil similarity")
			}
		})
	}
}

func TestTverskyParameters(t *testing.T) {
	// Test that Tversky with alpha=1, beta=1 equals Jaccard
	tversky := &TverskySimilarity{Alpha: 1.0, Beta: 1.0}
	jaccard := &JaccardSimilarity{}

	source := []string{"a", "b", "c", "d"}
	target := []string{"b", "c", "e", "f"}

	tverskyResult := tversky.Similarity(source, target, nil, nil)
	jaccardResult := jaccard.Similarity(source, target, nil, nil)

	if math.Abs(tverskyResult-jaccardResult) > 0.0001 {
		t.Errorf("Tversky(1,1) = %v, Jaccard = %v, should be equal", tverskyResult, jaccardResult)
	}
}

// Benchmarks

func BenchmarkJaccardSimilarity(b *testing.B) {
	j := &JaccardSimilarity{}
	source := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	target := []string{"b", "c", "d", "e", "k", "l", "m", "n", "o", "p"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j.Similarity(source, target, nil, nil)
	}
}

func BenchmarkWeightedJaccardSimilarity(b *testing.B) {
	w := &WeightedJaccardSimilarity{}
	source := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	target := []string{"b", "c", "d", "e", "k", "l", "m", "n", "o", "p"}
	weightsSource := map[string]float64{
		"a": 1.0, "b": 2.0, "c": 3.0, "d": 4.0, "e": 5.0,
		"f": 6.0, "g": 7.0, "h": 8.0, "i": 9.0, "j": 10.0,
	}
	weightsTarget := map[string]float64{
		"b": 2.0, "c": 3.0, "d": 4.0, "e": 5.0, "k": 1.0,
		"l": 2.0, "m": 3.0, "n": 4.0, "o": 5.0, "p": 6.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Similarity(source, target, weightsSource, weightsTarget)
	}
}

func BenchmarkCosineSimilarity(b *testing.B) {
	c := &CosineSimilarity{}
	source := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	target := []string{"b", "c", "d", "e", "k", "l", "m", "n", "o", "p"}
	weightsSource := map[string]float64{
		"a": 1.0, "b": 2.0, "c": 3.0, "d": 4.0, "e": 5.0,
		"f": 6.0, "g": 7.0, "h": 8.0, "i": 9.0, "j": 10.0,
	}
	weightsTarget := map[string]float64{
		"b": 2.0, "c": 3.0, "d": 4.0, "e": 5.0, "k": 1.0,
		"l": 2.0, "m": 3.0, "n": 4.0, "o": 5.0, "p": 6.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Similarity(source, target, weightsSource, weightsTarget)
	}
}
