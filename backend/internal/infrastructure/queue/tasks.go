package queue

const (
	// TypeExtractKeywords — task type for extracting keywords
	TypeExtractKeywords = "extract:keywords"
	// TypeComputeEmbedding — task type for computing embedding
	TypeComputeEmbedding = "compute:embedding"
	TypeBackupToCloud    = "backup:cloud"
	// TypeNotificationAchievement — task type for achievement notifications
	TypeNotificationAchievement = "notification:achievement"
	// TypeImportBookmarks — task type for async batch bookmark import
	TypeImportBookmarks = "import:bookmarks"
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

// ImportBookmarksPayload contains data for a batch bookmark import task
type ImportBookmarksPayload struct {
	TaskID string `json:"task_id"`
	UserID string `json:"user_id"`
	Items  []byte `json:"items"`
}
