package queue

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/pgvector/pgvector-go"

	importer "knowledge-graph/internal/application/import"
	dcache "knowledge-graph/internal/domain/cache"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/mongo"
	"knowledge-graph/internal/infrastructure/nlp"
)

// GammaLinkRunner generates automatic (gamma) links for a note and returns
// the created links. Implemented by recommendation.GammaLinkGenerator; kept as
// an interface so the worker can be tested without the generator.
type GammaLinkRunner interface {
	GenerateForNote(ctx context.Context, noteID uuid.UUID) ([]*link.Link, error)
}

// LinkEventPublisher publishes LinkCreated events (events.Publisher or a stub).
type LinkEventPublisher interface {
	PublishLinkCreated(ctx context.Context, sourceNoteID, targetNoteID, userID string) error
}

// RecommendationsEnqueuer schedules a recommendations refresh for a note.
type RecommendationsEnqueuer interface {
	EnqueueRefreshRecommendations(ctx context.Context, noteID uuid.UUID, delay time.Duration) error
}

// NlpArtifactsStore persists NLP-4 artifacts (mongo.NlpArtifactsRepository in
// production; fakes in tests). Nil store = pipeline storage unavailable.
type NlpArtifactsStore interface {
	FindCurrentSourceHash(ctx context.Context, noteID uuid.UUID, pipelineVersion string) (string, bool, error)
	SaveCurrent(ctx context.Context, doc *mongo.NlpArtifact, historyEnabled bool) error
	DeleteByNoteID(ctx context.Context, noteID uuid.UUID) (int64, error)
}

type Worker struct {
	noteRepo          note.Repository
	keywordRepo       *postgres.KeywordRepository
	embeddingRepo     *postgres.EmbeddingRepository
	nlpClient         *nlp.NLPClient
	cacheClient       dcache.CacheClient
	importSvc         *importer.Service
	gammaGen          GammaLinkRunner
	eventPublisher    LinkEventPublisher
	taskQueue         RecommendationsEnqueuer
	taskDelay         time.Duration
	artifactsStore    NlpArtifactsStore
	nlpHistoryEnabled bool
	nlpModelVersion   string
}

// NlpPipelineVersion identifies the normalizer ruleset; bump on rule changes
// so stale artifacts are distinguishable and recompute can refill.
const NlpPipelineVersion = "norm-v1"

func NewWorker(
	noteRepo note.Repository,
	keywordRepo *postgres.KeywordRepository,
	embeddingRepo *postgres.EmbeddingRepository,
	nlpClient *nlp.NLPClient,
	cacheClient dcache.CacheClient,
	importSvc *importer.Service,
	gammaGen GammaLinkRunner,
	eventPublisher LinkEventPublisher,
	taskQueue RecommendationsEnqueuer,
	taskDelay time.Duration,
	artifactsStore NlpArtifactsStore,
	nlpHistoryEnabled bool,
	nlpModelVersion string,
) *Worker {
	return &Worker{
		noteRepo:          noteRepo,
		keywordRepo:       keywordRepo,
		embeddingRepo:     embeddingRepo,
		nlpClient:         nlpClient,
		cacheClient:       cacheClient,
		importSvc:         importSvc,
		gammaGen:          gammaGen,
		eventPublisher:    eventPublisher,
		taskQueue:         taskQueue,
		taskDelay:         taskDelay,
		artifactsStore:    artifactsStore,
		nlpHistoryEnabled: nlpHistoryEnabled,
		nlpModelVersion:   nlpModelVersion,
	}
}

func (w *Worker) HandleExtractKeywords(ctx context.Context, t *asynq.Task) error {
	log.Println("HandleExtractKeywords: received task", string(t.Payload()))
	var p ExtractKeywordsTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Printf("HandleExtractKeywords: unmarshal error: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	noteID, err := uuid.Parse(p.NoteID)
	if err != nil {
		return fmt.Errorf("invalid note id: %w", err)
	}

	n, err := w.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to fetch note: %w", err)
	}
	if n == nil {
		return nil
	}

	title := n.Title().String()
	content := n.Content().String()
	text := title + " " + content
	if text == "" {
		// Удаляем ключевые слова
		return w.keywordRepo.DeleteAll(ctx, noteID)
	}

	wordCount := len(strings.Fields(text))
	topN := 5
	if wordCount > 0 {
		dynamic := wordCount / 100
		if dynamic < 5 {
			topN = 5
		} else if dynamic > 20 {
			topN = 20
		} else {
			topN = dynamic
		}
	}
	kwResult, err := w.nlpClient.ExtractKeywords(ctx, content, title, topN)
	if err != nil {
		log.Printf("HandleExtractKeywords: failed to extract keywords: %v", err)
		return fmt.Errorf("failed to extract keywords: %w", err)
	}
	keywords := kwResult.Keywords
	log.Printf("HandleExtractKeywords: extracted %d keywords for note %s", len(keywords), p.NoteID)

	// Преобразуем в модели GORM
	models := make([]postgres.NoteKeywordModel, 0, len(keywords))
	for _, kw := range keywords {
		models = append(models, postgres.NoteKeywordModel{
			NoteID:    noteID,
			Keyword:   kw.Keyword,
			Surface:   kw.Surface,
			Extractor: kwResult.Extractor,
			Weight:    kw.Weight,
		})
	}
	err = w.keywordRepo.SaveAll(ctx, noteID, models)
	if err != nil {
		log.Printf("HandleExtractKeywords: failed to save keywords: %v", err)
		return err
	}
	log.Printf("HandleExtractKeywords: successfully processed note %s with %d keywords", noteID, len(keywords))
	return nil
}

func (w *Worker) HandleComputeEmbedding(ctx context.Context, t *asynq.Task) error {
	log.Println("HandleComputeEmbedding: received task", string(t.Payload()))
	var p ComputeEmbeddingTaskPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Printf("HandleComputeEmbedding: unmarshal error: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	noteID, err := uuid.Parse(p.NoteID)
	if err != nil {
		log.Printf("HandleComputeEmbedding: invalid note id %s: %v", p.NoteID, err)
		return fmt.Errorf("invalid note id: %w", err)
	}

	n, err := w.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		log.Printf("HandleComputeEmbedding: failed to fetch note %s: %v", noteID, err)
		return fmt.Errorf("failed to fetch note: %w", err)
	}
	if n == nil {
		log.Printf("HandleComputeEmbedding: note %s not found", noteID)
		return nil
	}
	log.Printf("HandleComputeEmbedding: found note %s, processing...", noteID)

	title := n.Title().String()
	content := n.Content().String()
	text := title + " " + content
	if text == "" {
		// Удаляем эмбеддинг
		return w.embeddingRepo.Delete(ctx, noteID)
	}

	embedding, err := w.nlpClient.Embed(ctx, content, title)
	if err != nil {
		log.Printf("HandleComputeEmbedding: failed to compute embedding: %v", err)
		return fmt.Errorf("failed to compute embedding: %w", err)
	}
	log.Printf("HandleComputeEmbedding: computed embedding for note %s (size=%d)", noteID, len(embedding))

	// Преобразуем в pgvector.Vector
	vec := pgvector.NewVector(embedding)
	err = w.embeddingRepo.Upsert(ctx, noteID, vec)
	if err != nil {
		log.Printf("HandleComputeEmbedding: failed to upsert embedding: %v", err)
		return err
	}
	log.Printf("HandleComputeEmbedding: successfully processed note %s", noteID)

	// LINKS-1: automatic gamma links are generated only after the embedding
	// exists — never on note creation (no embedding yet) and never on a timer.
	if w.gammaGen != nil {
		if err := w.generateGammaLinks(ctx, n, noteID); err != nil {
			log.Printf("HandleComputeEmbedding: gamma link generation failed for %s: %v", noteID, err)
			return err
		}
	}
	return nil
}

// generateGammaLinks creates gamma links for the note, publishes LinkCreated
// for each and enqueues a recommendations refresh for source and targets —
// a target gained an incoming neighbour.
func (w *Worker) generateGammaLinks(ctx context.Context, n *note.Note, noteID uuid.UUID) error {
	created, err := w.gammaGen.GenerateForNote(ctx, noteID)
	if err != nil {
		return err
	}
	if len(created) == 0 {
		return nil
	}

	userID := ""
	if n != nil && n.CreatorID() != nil {
		userID = n.CreatorID().String()
	}

	for _, l := range created {
		if w.eventPublisher != nil {
			if err := w.eventPublisher.PublishLinkCreated(ctx, l.SourceNoteID().String(), l.TargetNoteID().String(), userID); err != nil {
				log.Printf("HandleComputeEmbedding: failed to publish LinkCreated %s -> %s: %v",
					l.SourceNoteID(), l.TargetNoteID(), err)
			}
		}
		if w.taskQueue != nil {
			if err := w.taskQueue.EnqueueRefreshRecommendations(ctx, l.TargetNoteID(), w.taskDelay); err != nil {
				log.Printf("HandleComputeEmbedding: failed to enqueue refresh for target %s: %v", l.TargetNoteID(), err)
			}
		}
	}
	if w.taskQueue != nil {
		if err := w.taskQueue.EnqueueRefreshRecommendations(ctx, noteID, w.taskDelay); err != nil {
			log.Printf("HandleComputeEmbedding: failed to enqueue refresh for source %s: %v", noteID, err)
		}
	}
	log.Printf("HandleComputeEmbedding: created %d gamma links for note %s", len(created), noteID)
	return nil
}

// HandleImportBookmarks processes an async batch bookmark import task.
func (w *Worker) HandleImportBookmarks(ctx context.Context, t *asynq.Task) error {
	log.Println("HandleImportBookmarks: received task", string(t.Payload()))

	var p ImportBookmarksPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		log.Printf("HandleImportBookmarks: unmarshal error: %v", err)
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	userID, err := uuid.Parse(p.UserID)
	if err != nil {
		log.Printf("HandleImportBookmarks: invalid user id %s: %v", p.UserID, err)
		return fmt.Errorf("invalid user id: %w", err)
	}

	var items []importer.Item
	if err := json.Unmarshal(p.Items, &items); err != nil {
		log.Printf("HandleImportBookmarks: failed to unmarshal items: %v", err)
		return fmt.Errorf("failed to unmarshal items: %w", err)
	}

	if w.importSvc == nil {
		return fmt.Errorf("import service is not configured")
	}

	return w.importSvc.ProcessImportTask(ctx, userID, p.TaskID, items)
}

// HandleNormalizeNote runs NLP-4 for one note: /normalize via the NLP
// service, then the artifact lands in Mongo as the single current document
// for (note_id, pipeline_version). Unchanged source_hash skips the call —
// makes the task idempotent under retries and recompute runs.
func (w *Worker) HandleNormalizeNote(ctx context.Context, t *asynq.Task) error {
	var p NormalizeNotePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	noteID, err := uuid.Parse(p.NoteID)
	if err != nil {
		return fmt.Errorf("invalid note id: %w", err)
	}

	if w.artifactsStore == nil {
		log.Printf("HandleNormalizeNote: artifacts store unavailable for note %s, skipping", noteID)
		return nil
	}

	n, err := w.noteRepo.FindByID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to fetch note: %w", err)
	}
	if n == nil {
		// Note deleted before the task ran — nothing to keep artifacts for.
		_, err := w.artifactsStore.DeleteByNoteID(ctx, noteID)
		return err
	}

	title := n.Title().String()
	content := n.Content().String()
	sourceHash := nlpArtifactsSourceHash(title, content)

	existingHash, found, err := w.artifactsStore.FindCurrentSourceHash(ctx, noteID, NlpPipelineVersion)
	if err != nil {
		return fmt.Errorf("failed to read current artifact: %w", err)
	}
	if found && existingHash == sourceHash {
		return nil // unchanged since last run — idempotent skip
	}

	res, err := w.nlpClient.Normalize(ctx, content, title)
	if err != nil {
		return fmt.Errorf("failed to normalize note: %w", err)
	}

	doc := &mongo.NlpArtifact{
		NoteID:          noteID,
		SourceHash:      sourceHash,
		PipelineVersion: NlpPipelineVersion,
		ModelVersion:    w.nlpModelVersion,
		NormalizedText:  res.NormalizedText,
		Chunks:          mapNormalizeChunks(res.Chunks),
		Metrics: mongo.NlpArtifactMetrics{
			RawTokens:      res.Metrics.RawTokens,
			NormTokens:     res.Metrics.NormTokens,
			Compression:    res.Metrics.Compression,
			Iterations:     res.Metrics.Iterations,
			StopReason:     res.Metrics.StopReason,
			EmbCosine:      res.Metrics.EmbCosine,
			RolledBack:     res.RolledBack,
			RollbackReason: res.RollbackReason,
			Skipped:        res.Skipped,
		},
	}
	if err := w.artifactsStore.SaveCurrent(ctx, doc, w.nlpHistoryEnabled); err != nil {
		return fmt.Errorf("failed to save artifact: %w", err)
	}
	log.Printf("HandleNormalizeNote: stored artifact for note %s (rolled_back=%v, chunks=%d)",
		noteID, res.RolledBack, len(res.Chunks))
	return nil
}

// HandleNlpArtifactsCleanup removes every nlp_artifacts document of a
// deleted note. No-op when the store is absent (pipeline never ran).
func (w *Worker) HandleNlpArtifactsCleanup(ctx context.Context, t *asynq.Task) error {
	var p NlpArtifactsCleanupPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	noteID, err := uuid.Parse(p.NoteID)
	if err != nil {
		return fmt.Errorf("invalid note id: %w", err)
	}
	if w.artifactsStore == nil {
		return nil
	}
	deleted, err := w.artifactsStore.DeleteByNoteID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to delete artifacts: %w", err)
	}
	if deleted > 0 {
		log.Printf("HandleNlpArtifactsCleanup: deleted %d artifacts for note %s", deleted, noteID)
	}
	return nil
}

// nlpArtifactsSourceHash hashes the normalization inputs (title + content)
// so a change to either triggers a rebuild.
func nlpArtifactsSourceHash(title, content string) string {
	sum := sha256.Sum256([]byte(title + "\x00" + content))
	return hex.EncodeToString(sum[:])
}

func mapNormalizeChunks(chunks []nlp.NormalizeChunk) []mongo.NlpArtifactChunk {
	out := make([]mongo.NlpArtifactChunk, len(chunks))
	for i, c := range chunks {
		out[i] = mongo.NlpArtifactChunk{
			Idx:         c.Idx,
			Text:        c.Text,
			HeadingPath: c.HeadingPath,
			CharSpan:    c.CharSpan,
			TokenCount:  c.TokenCount,
			Kind:        c.Kind,
			ForcedSplit: c.ForcedSplit,
		}
	}
	return out
}
