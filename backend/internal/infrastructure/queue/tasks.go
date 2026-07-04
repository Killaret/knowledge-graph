package queue

const (
	// TypeExtractKeywords is the task type for keyword extraction
	TypeExtractKeywords = "extract:keywords"
	// TypeComputeEmbedding is the task type for embedding computation
	TypeComputeEmbedding = "compute:embedding"
)

// ExtractKeywordsTaskPayload holds the data for a keyword extraction task
type ExtractKeywordsTaskPayload struct {
	NoteID string `json:"note_id"`
	TopN   int    `json:"top_n"`
}

// ComputeEmbeddingTaskPayload holds the data for an embedding computation task
type ComputeEmbeddingTaskPayload struct {
	NoteID string `json:"note_id"`
}
