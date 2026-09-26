package queue

const (
	// TypeExtractKeywords — task type for extracting keywords
	TypeExtractKeywords = "extract:keywords"
	// TypeComputeEmbedding — task type for computing embedding
	TypeComputeEmbedding = "compute:embedding"
	TypeBackupToCloud    = "backup:cloud"
	TypeDatabaseBackup   = "backup:database"
	// TypeNotificationAchievement — task type for achievement notifications
	TypeNotificationAchievement = "notification:achievement"
	// TypeImportBookmarks — task type for async batch bookmark import
	TypeImportBookmarks = "import:bookmarks"
	// TypeNormalizeNote — NLP-4: normalize a note and store the artifact in Mongo
	TypeNormalizeNote = "nlp:normalize"
	// TypeNlpArtifactsCleanup — cascade: remove a deleted note's nlp_artifacts
	TypeNlpArtifactsCleanup = "nlp:artifacts-cleanup"
	// TypeAssessQuality — NOTE-QUALITY-1: compute signals/gates for a note
	TypeAssessQuality = "quality:assess"
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

// NormalizeNotePayload identifies the note to normalize (NLP-4).
type NormalizeNotePayload struct {
	NoteID string `json:"note_id"`
}

// NlpArtifactsCleanupPayload identifies the note whose artifacts are removed.
type NlpArtifactsCleanupPayload struct {
	NoteID string `json:"note_id"`
}

// AssessQualityPayload identifies the note and what triggered the assessment
// (quality.TriggerAuto | quality.TriggerManual — kept as string to avoid an
// application import in the payload file).
type AssessQualityPayload struct {
	NoteID  string `json:"note_id"`
	Trigger string `json:"trigger"`
}

// ImportBookmarksPayload contains data for a batch bookmark import task
type ImportBookmarksPayload struct {
	TaskID string `json:"task_id"`
	UserID string `json:"user_id"`
	Items  []byte `json:"items"`
}
