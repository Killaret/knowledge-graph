package engine

// AnalyticsNode represents a neighbor returned by graph analytics queries.
type AnalyticsNode struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Type     string  `json:"type"`
	Weight   float64 `json:"weight"`
	Distance int     `json:"distance"`
}

// AnalyticsPath represents a shortest path between two notes.
type AnalyticsPath struct {
	NoteIDs  []string `json:"note_ids"`
	Distance int      `json:"distance"`
	Weight   float64  `json:"weight"`
}

// AnalyticsRecommendation represents a single recommendation result.
type AnalyticsRecommendation struct {
	NoteID        string  `json:"note_id"`
	Title         string  `json:"title"`
	Score         float64 `json:"score"`
	GraphScore    float64 `json:"graph_score"`
	SemanticScore float64 `json:"semantic_score"`
}
