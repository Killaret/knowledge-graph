package queue

const (
	// TypeExtractKeywords — task type for extracting keywords
	TypeExtractKeywords = "extract:keywords"
	// TypeComputeEmbedding — task type for computing embedding
	TypeComputeEmbedding = "compute:embedding"
	TypeBackupToCloud    = "backup:cloud"
)

// ExtractKeywordsTaskPayload contains data for the keyword extraction task
type ExtractKeywordsTaskPayload struct {
	NoteID string `json:"note_id"`
	TopN   int    `json:"top_n"`
}

// ComputeEmbeddingTaskPayload contains data for the embedding computation task
type ComputeEmbeddingTaskPayload struct {
	NoteID string `json:"note_id"`
}

// BackupToCloudPayload contains data for the cloud backup task
type BackupToCloudPayload struct {
	LocalPath string `json:"local_path"`
	RemoteKey string `json:"remote_key"`
}
