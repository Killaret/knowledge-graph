package engine

import (
	"context"
	"math"
	"sort"

	"knowledge-graph-graph-service/internal/db"
)

const (
	defaultRecDepth  = 2
	defaultRecLimit  = 10
	defaultRecAlpha  = 0.5
	defaultRecBeta   = 0.5
	defaultRecDecay  = 0.5
	maxRecDepth      = 5
	maxRecCandidates = 100
)

// Recommendations returns ranked note recommendations for a source note.
// It combines graph proximity (transitive closure distance and link weights)
// with cosine similarity of pgvector embeddings when available.
func Recommendations(ctx context.Context, client db.PostgresClient, filter db.NotesFilter, noteID string, depth, limit int, alpha, beta float64) ([]*AnalyticsRecommendation, error) {
	if depth <= 0 {
		depth = defaultRecDepth
	}
	if depth > maxRecDepth {
		depth = maxRecDepth
	}
	if limit <= 0 {
		limit = defaultRecLimit
	}
	if alpha <= 0 && beta <= 0 {
		alpha = defaultRecAlpha
		beta = defaultRecBeta
	}
	total := alpha + beta
	alpha /= total
	beta /= total

	// Load more candidates than requested so the top-N remains meaningful after scoring.
	candidateLimit := limit * 5
	if candidateLimit > maxRecCandidates {
		candidateLimit = maxRecCandidates
	}

	candidates, err := client.GetRecommendationCandidates(ctx, filter, noteID, depth, candidateLimit)
	if err != nil {
		return nil, err
	}

	candidateIDs := make([]string, 0, len(candidates))
	for _, c := range candidates {
		candidateIDs = append(candidateIDs, c.ID)
	}

	sourceEmbeddings, err := client.GetEmbeddings(ctx, []string{noteID})
	if err != nil {
		sourceEmbeddings = nil
	}
	candidateEmbeddings, err := client.GetEmbeddings(ctx, candidateIDs)
	if err != nil {
		candidateEmbeddings = nil
	}

	sourceVector := sourceEmbeddings[noteID]
	hasSourceVector := len(sourceVector) > 0

	results := make([]*AnalyticsRecommendation, 0, len(candidates))
	for _, c := range candidates {
		graphScore := c.Weight * math.Pow(defaultRecDecay, float64(c.Distance))

		var semanticScore float64
		if hasSourceVector {
			if vec, ok := candidateEmbeddings[c.ID]; ok && len(vec) > 0 {
				similarity := cosineSimilarity(sourceVector, vec)
				semanticScore = (similarity + 1.0) / 2.0
			}
		}

		score := alpha*graphScore + beta*semanticScore
		results = append(results, &AnalyticsRecommendation{
			NoteID:        c.ID,
			Title:         c.Title,
			Score:         score,
			GraphScore:    graphScore,
			SemanticScore: semanticScore,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		x := float64(a[i])
		y := float64(b[i])
		dot += x * y
		normA += x * x
		normB += y * y
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
