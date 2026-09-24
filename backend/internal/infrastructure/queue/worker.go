package queue

import (
	"context"
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

type Worker struct {
	noteRepo       note.Repository
	keywordRepo    *postgres.KeywordRepository
	embeddingRepo  *postgres.EmbeddingRepository
	nlpClient      *nlp.NLPClient
	cacheClient    dcache.CacheClient
	importSvc      *importer.Service
	gammaGen       GammaLinkRunner
	eventPublisher LinkEventPublisher
	taskQueue      RecommendationsEnqueuer
	taskDelay      time.Duration
}

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
) *Worker {
	return &Worker{
		noteRepo:       noteRepo,
		keywordRepo:    keywordRepo,
		embeddingRepo:  embeddingRepo,
		nlpClient:      nlpClient,
		cacheClient:    cacheClient,
		importSvc:      importSvc,
		gammaGen:       gammaGen,
		eventPublisher: eventPublisher,
		taskQueue:      taskQueue,
		taskDelay:      taskDelay,
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
