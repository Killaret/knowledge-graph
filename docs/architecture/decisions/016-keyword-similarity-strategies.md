# ADR 016: Keyword Similarity Strategies

## Status
Accepted

## Context
Knowledge Graph uses keyword extraction and similarity matching to improve note recommendations. Notes have keywords extracted from their content, and similarity between keywords helps find related notes.

As the system evolved, several requirements emerged:
- **Flexible similarity metrics**: Different use cases require different similarity measures
- **Configurable weights**: Keywords have different importance levels
- **Performance**: Similarity calculations must be fast
- **Extensibility**: Easy to add new similarity strategies without modifying core logic
- **A/B testing**: Ability to test different strategies in production

### Current State Analysis
The system has:
- Basic keyword extraction from note content
- Simple similarity calculation (likely Jaccard)
- Hard-coded similarity logic in recommendation engine
- No support for keyword weights

Current challenges:
- **Inflexible**: Cannot easily change similarity metric
- **No weights**: All keywords treated equally regardless of importance
- **Hard to test**: Cannot A/B test different strategies
- **Scattered logic**: Similarity calculation mixed with recommendation logic
- **Limited metrics**: Only basic similarity measures supported

## Problem Statement
How do we create a flexible, extensible keyword similarity system that supports multiple metrics, configurable weights, and easy A/B testing?

## Decision Drivers
- **Flexibility**: Support multiple similarity metrics
- **Performance**: Similarity calculations must be efficient
- **Extensibility**: Easy to add new strategies
- **Configurability**: Weights and parameters should be configurable
- **Testability**: Easy to A/B test different strategies
- **Maintainability**: Clear separation of similarity logic from recommendation logic

## Considered Options

### Option 1: Hard-coded Similarity Function
Single similarity function hard-coded in recommendation engine.

**Pros:**
- ✅ Simplest implementation
- ✅ Fast (no overhead)
- ✅ No configuration needed

**Cons:**
- ❌ Inflexible (cannot change metric)
- ❌ No weight support
- ❌ Cannot A/B test
- ❌ Hard to extend (need to modify core logic)
- ❌ One-size-fits-all (not optimal for all use cases)

### Option 2: Configuration-based Similarity
Similarity metric selected via config, single implementation.

**Pros:**
- ✅ Can switch between metrics via config
- ✅ Simpler than strategy pattern
- ✅ Some flexibility

**Cons:**
- ❌ Still need to modify code to add new metrics
- ❌ Cannot test multiple metrics simultaneously
- ❌ No weight support (unless added separately)
- ❌ Limited extensibility
- ❌ Configuration complexity grows with each metric

### Option 3: Strategy Pattern with Interface
Define similarity strategy interface, implement multiple strategies, select at runtime.

**Pros:**
- ✅ Easy to add new strategies (implement interface)
- ✅ Can A/B test (use multiple strategies)
- ✅ Clear separation of concerns
- ✅ Type-safe (compile-time checking)
- ✅ Can support weights in interface
- ✅ Easy to test (mock interface)

**Cons:**
- ❌ More complex than simple function
- ❌ Need to manage strategy lifecycle
- ❌ Slight performance overhead (interface dispatch)
- ❌ More code to maintain

### Option 4: Plugin System
Load similarity strategies as external plugins.

**Pros:**
- ✅ Maximum extensibility (add without recompiling)
- ✅ Can develop strategies independently
- ✅ Dynamic loading

**Cons:**
- ❌ Complex infrastructure (plugin loading, security)
- ❌ Overkill for current needs
- ❌ Deployment complexity
- ❌ Security concerns (arbitrary code execution)
- ❌ Type safety issues (dynamic loading)
- ❌ Not necessary for current scale

### Option 5: Strategy Pattern + Weight Support
Strategy pattern with built-in weight support in interface.

**Pros:**
- ✅ All benefits of strategy pattern
- ✅ Native weight support
- ✅ Configurable weights per keyword
- ✅ Flexible (weights can be frequency-based, manual, etc.)
- ✅ Can combine with other features

**Cons:**
- ❌ More complex interface
- ❌ Need weight calculation/management
- ❌ Additional storage for weights

## Decision
**Chosen Approach: Option 5 - Strategy Pattern with Weight Support**

We implement a strategy pattern with a well-defined interface that supports weighted keywords. This provides maximum flexibility while keeping complexity manageable.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              Recommendation Engine                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │         TraversalService                            │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │  SimilarityStrategy (interface)             │   │   │
│  │  │  ┌─────────────────────────────────────┐   │   │   │
│  │  │  │ CalculateSimilarity(keywords1,      │   │   │   │
│  │  │  │                  keywords2, weights) │   │   │   │
│  │  │  │  -> float64                         │   │   │   │
│  │  │  └─────────────────────────────────────┘   │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  │                       │                            │   │
│  │         ┌─────────────┼─────────────┐              │   │
│  │         │             │             │              │   │
│  │  ┌──────▼─────┐ ┌────▼─────┐ ┌───▼──────┐        │   │
│  │  │ Jaccard    │ │ Overlap  │ │ Tversky  │  ...   │   │
│  │  │ Strategy   │ │ Strategy │ │ Strategy │        │   │
│  │  └────────────┘ └──────────┘ └──────────┘        │   │
│  │                                                   │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │  Weighted Jaccard Strategy                   │   │   │
│  │  │  (supports keyword weights)                  │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  │                                                   │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │  Cosine Similarity Strategy                │   │   │
│  │  │  (for vector-based comparison)             │   │   │
│  │  └─────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ Selected via Config
                          │
┌─────────────────────────▼───────────────────────────────────┐
│              knowledge-graph.config.json                    │
│  {                                                         │
│    "keyword_similarity_method": "weighted_jaccard",        │
│    "keyword_tversky_alpha": 0.5,                           │
│    "keyword_tversky_beta": 0.5                             │
│  }                                                         │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Details

#### 1. Similarity Strategy Interface

```go
package similarity

// KeywordWithWeight represents a keyword with its importance weight
type KeywordWithWeight struct {
    Keyword string
    Weight  float64 // 0.0 to 1.0, higher = more important
}

// SimilarityStrategy defines the interface for keyword similarity calculations
type SimilarityStrategy interface {
    // CalculateSimilarity computes similarity between two sets of keywords
    CalculateSimilarity(keywords1, keywords2 []KeywordWithWeight) float64
    
    // Name returns the strategy name for logging/configuration
    Name() string
}
```

#### 2. Strategy Implementations

**Jaccard Strategy:**
```go
type JaccardStrategy struct{}

func (s *JaccardStrategy) CalculateSimilarity(kw1, kw2 []KeywordWithWeight) float64 {
    set1 := keywordSet(kw1)
    set2 := keywordSet(kw2)
    
    intersection := 0.0
    for k := range set1 {
        if set2[k] {
            intersection++
        }
    }
    
    union := float64(len(set1) + len(set2) - int(intersection))
    if union == 0 {
        return 0.0
    }
    
    return intersection / union
}

func (s *JaccardStrategy) Name() string {
    return "jaccard"
}
```

**Overlap Coefficient Strategy:**
```go
type OverlapStrategy struct{}

func (s *OverlapStrategy) CalculateSimilarity(kw1, kw2 []KeywordWithWeight) float64 {
    set1 := keywordSet(kw1)
    set2 := keywordSet(kw2)
    
    intersection := 0.0
    for k := range set1 {
        if set2[k] {
            intersection++
        }
    }
    
    smallerSet := math.Min(float64(len(set1)), float64(len(set2)))
    if smallerSet == 0 {
        return 0.0
    }
    
    return intersection / smallerSet
}

func (s *OverlapStrategy) Name() string {
    return "overlap"
}
```

**Tversky Index Strategy:**
```go
type TverskyStrategy struct {
    Alpha float64
    Beta  float64
}

func NewTverskyStrategy(alpha, beta float64) *TverskyStrategy {
    return &TverskyStrategy{Alpha: alpha, Beta: beta}
}

func (s *TverskyStrategy) CalculateSimilarity(kw1, kw2 []KeywordWithWeight) float64 {
    set1 := keywordSet(kw1)
    set2 := keywordSet(kw2)
    
    intersection := 0.0
    difference1 := 0.0
    difference2 := 0.0
    
    for k := range set1 {
        if set2[k] {
            intersection++
        } else {
            difference1++
        }
    }
    
    for k := range set2 {
        if !set1[k] {
            difference2++
        }
    }
    
    denominator := intersection + s.Alpha*difference1 + s.Beta*difference2
    if denominator == 0 {
        return 0.0
    }
    
    return intersection / denominator
}

func (s *TverskyStrategy) Name() string {
    return fmt.Sprintf("tversky(α=%.2f,β=%.2f)", s.Alpha, s.Beta)
}
```

**Weighted Jaccard Strategy:**
```go
type WeightedJaccardStrategy struct{}

func (s *WeightedJaccardStrategy) CalculateSimilarity(kw1, kw2 []KeywordWithWeight) float64 {
    // Build keyword -> weight maps
    weights1 := make(map[string]float64)
    weights2 := make(map[string]float64)
    
    for _, kw := range kw1 {
        weights1[kw.Keyword] = kw.Weight
    }
    for _, kw := range kw2 {
        weights2[kw.Keyword] = kw.Weight
    }
    
    // Calculate weighted intersection
    intersection := 0.0
    allKeywords := make(map[string]bool)
    
    for k := range weights1 {
        allKeywords[k] = true
        if w2, ok := weights2[k]; ok {
            // Both have keyword: use average weight
            intersection += (weights1[k] + w2) / 2.0
        }
    }
    
    // Calculate weighted union
    union := 0.0
    for k := range allKeywords {
        w1 := weights1[k]
        w2 := weights2[k]
        union += math.Max(w1, w2) // Use max weight for union
    }
    
    if union == 0 {
        return 0.0
    }
    
    return intersection / union
}

func (s *WeightedJaccardStrategy) Name() string {
    return "weighted_jaccard"
}
```

**Cosine Similarity Strategy:**
```go
type CosineStrategy struct{}

func (s *CosineStrategy) CalculateSimilarity(kw1, kw2 []KeywordWithWeight) float64 {
    // Build keyword -> weight vectors
    vec1 := make(map[string]float64)
    vec2 := make(map[string]float64)
    
    for _, kw := range kw1 {
        vec1[kw.Keyword] = kw.Weight
    }
    for _, kw := range kw2 {
        vec2[kw.Keyword] = kw.Weight
    }
    
    // Calculate dot product
    dotProduct := 0.0
    for k, w1 := range vec1 {
        if w2, ok := vec2[k]; ok {
            dotProduct += w1 * w2
        }
    }
    
    // Calculate magnitudes
    mag1 := 0.0
    for _, w := range vec1 {
        mag1 += w * w
    }
    mag1 = math.Sqrt(mag1)
    
    mag2 := 0.0
    for _, w := range vec2 {
        mag2 += w * w
    }
    mag2 = math.Sqrt(mag2)
    
    if mag1 == 0 || mag2 == 0 {
        return 0.0
    }
    
    return dotProduct / (mag1 * mag2)
}

func (s *CosineStrategy) Name() string {
    return "cosine"
}
```

#### 3. Strategy Factory

```go
type StrategyFactory struct{}

func (f *StrategyFactory) CreateStrategy(method string, params map[string]interface{}) (SimilarityStrategy, error) {
    switch method {
    case "jaccard":
        return &JaccardStrategy{}, nil
    case "overlap":
        return &OverlapStrategy{}, nil
    case "tversky":
        alpha := getFloatParam(params, "alpha", 0.5)
        beta := getFloatParam(params, "beta", 0.5)
        return NewTverskyStrategy(alpha, beta), nil
    case "weighted_jaccard":
        return &WeightedJaccardStrategy{}, nil
    case "cosine":
        return &CosineStrategy{}, nil
    default:
        return nil, fmt.Errorf("unknown similarity method: %s", method)
    }
}
```

#### 4. Configuration

```json
{
  "keyword_similarity_method": "weighted_jaccard",
  "keyword_tversky_alpha": 0.5,
  "keyword_tversky_beta": 0.5
}
```

#### 5. Weight Calculation

Weights can be calculated using various strategies:

**Frequency-based:**
```go
func CalculateFrequencyWeights(keywords []string) []KeywordWithWeight {
    freq := make(map[string]int)
    for _, kw := range keywords {
        freq[kw]++
    }
    
    maxFreq := 0
    for _, f := range freq {
        if f > maxFreq {
            maxFreq = f
        }
    }
    
    result := make([]KeywordWithWeight, 0, len(freq))
    for kw, f := range freq {
        weight := float64(f) / float64(maxFreq)
        result = append(result, KeywordWithWeight{Keyword: kw, Weight: weight})
    }
    
    return result
}
```

**TF-IDF:**
```go
func CalculateTFIDFWeights(keywords []string, corpus []string) []KeywordWithWeight {
    // Calculate term frequency
    tf := make(map[string]float64)
    for _, kw := range keywords {
        tf[kw]++
    }
    total := float64(len(keywords))
    for kw := range tf {
        tf[kw] = tf[kw] / total
    }
    
    // Calculate inverse document frequency
    idf := make(map[string]float64)
    n := float64(len(corpus))
    for kw := range tf {
        docCount := 0
        for _, doc := range corpus {
            if contains(doc, kw) {
                docCount++
            }
        }
        idf[kw] = math.Log(n / float64(docCount+1))
    }
    
    // Calculate TF-IDF
    result := make([]KeywordWithWeight, 0, len(tf))
    for kw := range tf {
        weight := tf[kw] * idf[kw]
        result = append(result, KeywordWithWeight{Keyword: kw, Weight: weight})
    }
    
    return result
}
```

### Strategy Comparison

| Strategy | Best For | Weights | Parameters | Complexity |
|----------|----------|---------|------------|------------|
| Jaccard | Set similarity | ❌ | None | Low |
| Overlap | Small sets | ❌ | None | Low |
| Tversky | Asymmetric similarity | ❌ | α, β | Medium |
| Weighted Jaccard | Weighted sets | ✅ | None | Medium |
| Cosine | Vector similarity | ✅ | None | Medium |

### A/B Testing Strategy

```go
type ABTestStrategy struct {
    strategyA SimilarityStrategy
    strategyB SimilarityStrategy
    split     float64 // 0.0 to 1.0
}

func (s *ABTestStrategy) CalculateSimilarity(kw1, kw2 []KeywordWithWeight) float64 {
    if rand.Float64() < s.split {
        return s.strategyA.CalculateSimilarity(kw1, kw2)
    }
    return s.strategyB.CalculateSimilarity(kw1, kw2)
}
```

## Consequences

### Positive Consequences
- ✅ **Flexibility**: Easy to switch between similarity metrics
- ✅ **Extensibility**: Add new strategies without modifying core logic
- ✅ **Weight support**: Can prioritize important keywords
- ✅ **A/B testing**: Can test multiple strategies simultaneously
- ✅ **Type safety**: Compile-time checking via interface
- ✅ **Testability**: Easy to mock and unit test
- ✅ **Performance**: Each strategy optimized for its use case

### Negative Consequences
- ❌ **Complexity**: More code than simple function
- ❌ **Interface overhead**: Slight performance cost (negligible)
- ❌ **Configuration complexity**: Need to manage strategy selection
- ❌ **Weight management**: Need to calculate/store weights
- ❌ **Decision paralysis**: Too many strategy choices

### Mitigation Strategies
- **Complexity**: Clear documentation, examples for common use cases
- **Interface overhead**: Negligible in practice (< 1ns overhead)
- **Configuration complexity**: Sensible defaults, validation
- **Weight management**: Provide weight calculation utilities
- **Decision paralysis**: Document when to use each strategy

## When to Reconsider
- If performance profiling shows interface overhead is significant
- If strategy selection becomes too complex
- If need more advanced features (e.g., ensemble strategies)
- If weights become too complex to manage

## Alternatives for Future
- **Ensemble strategies**: Combine multiple strategies
- **Machine learning**: Learn optimal similarity metric
- **Hybrid approaches**: Combine embedding-based with keyword-based

## References
- [RECOMMENDATION_ARCHITECTURE.md](../../RECOMMENDATION_ARCHITECTURE.md)
- [Strategy Pattern (GoF)](https://en.wikipedia.org/wiki/Strategy_pattern)
- [Similarity Metrics in NLP](https://nlp.stanford.edu/IR-book/html/htmledition/definition-of-similarity-1.html)
