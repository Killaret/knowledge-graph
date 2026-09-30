package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	appquality "knowledge-graph/internal/application/quality"
	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/mongo"
)

// QualityEnqueuer schedules quality:assess tasks. Implemented by AsynqClient;
// kept narrow so tests inject a spy.
type QualityEnqueuer interface {
	EnqueueAssessQuality(ctx context.Context, noteID string, trigger string) error
}

// qualityNoteReader adapts note.Repository to quality.NoteReader.
type qualityNoteReader struct {
	repo note.Repository
}

func (r *qualityNoteReader) FindByID(ctx context.Context, id uuid.UUID) (*appquality.Note, error) {
	n, err := r.repo.FindByID(ctx, id)
	if err != nil || n == nil {
		return nil, err
	}
	meta := n.Metadata().Value()
	out := &appquality.Note{
		ID:       n.ID(),
		Title:    n.Title().String(),
		Content:  n.Content().String(),
		Metadata: meta,
	}
	if u, ok := meta["source_url"].(string); ok {
		out.SourceURL = u
	}
	return out, nil
}

// qualityArtifactReader adapts the nlp_artifacts store to
// quality.ArtifactReader, binding the pipeline version.
type qualityArtifactReader struct {
	repo *mongo.NlpArtifactsRepository
}

func (r *qualityArtifactReader) CurrentArtifact(ctx context.Context, noteID uuid.UUID) (*appquality.Artifact, error) {
	doc, err := r.repo.FindCurrent(ctx, noteID, NlpPipelineVersion)
	if err != nil || doc == nil {
		return nil, err
	}
	chunks := make([]string, len(doc.Chunks))
	for i, c := range doc.Chunks {
		chunks[i] = c.Text
	}
	return &appquality.Artifact{
		SourceHash:     doc.SourceHash,
		NormalizedText: doc.NormalizedText,
		Chunks:         chunks,
	}, nil
}

// qualityArtifactWriter adapts SetQuality(*Record) to the value-typed port.
type qualityArtifactWriter struct {
	repo *mongo.NlpArtifactsRepository
}

func (w *qualityArtifactWriter) SetQuality(ctx context.Context, noteID uuid.UUID, sourceHash string, rec appquality.Record) (bool, error) {
	return w.repo.SetQuality(ctx, noteID, sourceHash, &rec)
}

// NewQualityAssessor builds the assessor from infrastructure pieces —
// nil artifacts repo degrades to raw-text signals without chunk coherence.
func NewQualityAssessor(noteRepo note.Repository, artifacts *mongo.NlpArtifactsRepository,
	logStore appquality.LogStore, stats appquality.StatsReader,
	embedder appquality.Embedder, th appquality.Thresholds) *appquality.Assessor {
	var ar appquality.ArtifactReader
	var aw appquality.ArtifactQualityWriter
	if artifacts != nil {
		ar = &qualityArtifactReader{repo: artifacts}
		aw = &qualityArtifactWriter{repo: artifacts}
	}
	return appquality.NewAssessor(&qualityNoteReader{repo: noteRepo}, ar, aw, logStore, stats, embedder, th)
}

// UseQuality installs the NOTE-QUALITY-1 dependencies: the assessor that
// HandleAssessQuality runs and the enqueuer that schedules follow-up
// assessments after enrichment tasks. Nil assessor = quality off.
func (w *Worker) UseQuality(assessor *appquality.Assessor, enq QualityEnqueuer) {
	w.qualityAssessor = assessor
	w.qualityEnq = enq
}

// scheduleQuality enqueues an automatic re-assessment after an enrichment
// step changed the note's data. Best-effort: a queue error is logged, not
// returned — the enrichment itself already succeeded.
func (w *Worker) scheduleQuality(ctx context.Context, noteID uuid.UUID) {
	if w.qualityAssessor == nil || w.qualityEnq == nil {
		return
	}
	if err := w.qualityEnq.EnqueueAssessQuality(ctx, noteID.String(), appquality.TriggerAuto); err != nil {
		log.Printf("quality: failed to enqueue assessment for %s: %v", noteID, err)
	}
}

// HandleAssessQuality runs one NOTE-QUALITY-1 pass. No-op when the assessor
// is not configured (flag off or Mongo missing) — defence in depth behind
// the client-side gate.
func (w *Worker) HandleAssessQuality(ctx context.Context, t *asynq.Task) error {
	var p AssessQualityPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	noteID, err := uuid.Parse(p.NoteID)
	if err != nil {
		return fmt.Errorf("invalid note id: %w", err)
	}
	if w.qualityAssessor == nil {
		return nil
	}
	trigger := p.Trigger
	if trigger != appquality.TriggerManual {
		trigger = appquality.TriggerAuto
	}
	rec, err := w.qualityAssessor.Assess(ctx, noteID, trigger)
	if err != nil {
		return fmt.Errorf("quality assess: %w", err)
	}
	if rec != nil {
		log.Printf("HandleAssessQuality: note %s -> verdict=%s reasons=%v attempt=%d",
			noteID, rec.Verdict, rec.Reasons, rec.Attempt)
	}
	return nil
}
