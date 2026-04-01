package queue

const (
	// TypeExtractKeywords — тип задачи для извлечения ключевых слов
	TypeExtractKeywords = "extract:keywords"
	// TypeComputeEmbedding — тип задачи для вычисления эмбеддинга
	TypeComputeEmbedding = "compute:embedding"
)

// ExtractKeywordsTaskPayload содержит данные для задачи извлечения ключевых слов
type ExtractKeywordsTaskPayload struct {
	NoteID string `json:"note_id"`
	TopN   int    `json:"top_n"`
}

// ComputeEmbeddingTaskPayload содержит данные для задачи вычисления эмбеддинга
type ComputeEmbeddingTaskPayload struct {
	NoteID string `json:"note_id"`
}
