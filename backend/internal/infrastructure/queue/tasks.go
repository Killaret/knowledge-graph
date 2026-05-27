package queue

const (
	// TypeExtractKeywords — тип задачи для извлечения ключевых слов
	TypeExtractKeywords = "extract:keywords"
	// TypeComputeEmbedding — тип задачи для вычисления эмбеддинга
	TypeComputeEmbedding = "compute:embedding"
	TypeBackupToCloud    = "backup:cloud"
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

// BackupToCloudPayload содержит данные для задачи бэкапа в облако
type BackupToCloudPayload struct {
	LocalPath string `json:"local_path"`
	RemoteKey string `json:"remote_key"`
}
