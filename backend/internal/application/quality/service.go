package quality

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Ports — everything the assessor needs, kept narrow for fakes in tests.

// NoteReader loads the note under assessment.
type NoteReader interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Note, error)
}

// Note is the small projection the assessor reads.
type Note struct {
	ID        uuid.UUID
	Title     string
	Content   string
	SourceURL string // metadata.source_url when present
	Metadata  map[string]any
}

// ArtifactReader exposes the current NLP-4 artifact (normalized text + chunks)
// when the pipeline produced one.
type ArtifactReader interface {
	CurrentArtifact(ctx context.Context, noteID uuid.UUID) (*Artifact, error)
}

// Artifact is the piece of nlp_artifacts the assessor consumes.
type Artifact struct {
	SourceHash     string
	NormalizedText string
	Chunks         []string // chunk texts in order
}

// ArtifactQualityWriter stamps the assessment onto the current artifact of the
// same text version. A missing artifact is not an error — the record still
// lands in the log.
type ArtifactQualityWriter interface {
	SetQuality(ctx context.Context, noteID uuid.UUID, sourceHash string, rec Record) (bool, error)
}

// LogStore is the quality_log collection: append, list, keep last N.
type LogStore interface {
	Append(ctx context.Context, entry LogEntry) error
	List(ctx context.Context, noteID uuid.UUID, limit int) ([]LogEntry, error)
	// MarkManualReview sets needs_manual_review on the latest entry of the
	// given text version. Returns true when an entry was updated.
	MarkManualReview(ctx context.Context, noteID uuid.UUID, sourceHash string) (bool, error)
}

// StatsReader supplies the database counters (keywords, links, embedding).
type StatsReader interface {
	KeywordCount(ctx context.Context, noteID uuid.UUID) (int, error)
	LinkCount(ctx context.Context, noteID uuid.UUID) (int, error)
	HasEmbedding(ctx context.Context, noteID uuid.UUID) (bool, error)
}

// Embedder embeds a text for coherence and title/text similarity. The NLP
// service client satisfies this; tests inject vectors.
type Embedder interface {
	Embed(ctx context.Context, text string, title string) ([]float32, error)
}

// LogEntry is one row of quality_log.
type LogEntry struct {
	NoteID     uuid.UUID `json:"note_id" bson:"note_id"`
	SourceHash string    `json:"source_hash" bson:"source_hash"`
	Trigger    string    `json:"trigger" bson:"trigger"`
	Record     Record    `json:"record" bson:"record"`
	Changed    bool      `json:"changed" bson:"changed"` // signals differ from the previous entry
}

// LogKeepPerNote — the log keeps at most this many entries per note.
const LogKeepPerNote = 10

// titleSimilarityWindow — how much of the body the title is compared against.
const titleSimilarityRunes = 500

// Assessor computes stage-1 signals, evaluates the gates and writes the
// result to the artifact + the log under the stop rule.
type Assessor struct {
	notes      NoteReader
	artifacts  ArtifactReader
	artifactW  ArtifactQualityWriter
	log        LogStore
	stats      StatsReader
	embedder   Embedder // may be nil — coherence/similarity become null
	thresholds Thresholds
	now        func() time.Time
}

// NewAssessor wires the assessor; any store may be nil (degrades to fewer
// signals), embedder nil means coherence and title similarity stay null.
func NewAssessor(notes NoteReader, artifacts ArtifactReader, artifactW ArtifactQualityWriter,
	log LogStore, stats StatsReader, embedder Embedder, th Thresholds) *Assessor {
	return &Assessor{
		notes: notes, artifacts: artifacts, artifactW: artifactW,
		log: log, stats: stats, embedder: embedder,
		thresholds: th,
		now:        time.Now,
	}
}

// SourceHash mirrors worker.nlpArtifactsSourceHash: sha256 of title + NUL +
// content — the text version the assessment refers to.
func SourceHash(title, content string) string {
	sum := sha256.Sum256([]byte(title + "\x00" + content))
	return hex.EncodeToString(sum[:])
}

// Assess runs one pass. Returns the stored record, or nil when the stop rule
// suppressed it. Errors are returned for storage/IO failures.
func (a *Assessor) Assess(ctx context.Context, noteID uuid.UUID, trigger string) (*Record, error) {
	n, err := a.notes.FindByID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("quality: load note: %w", err)
	}
	if n == nil {
		return nil, nil // note deleted before the task ran
	}
	hash := SourceHash(n.Title, n.Content)

	var artifact *Artifact
	if a.artifacts != nil {
		if artifact, err = a.artifacts.CurrentArtifact(ctx, noteID); err != nil {
			return nil, fmt.Errorf("quality: load artifact: %w", err)
		}
	}
	// The artifact is only authoritative for the text it was built from.
	var text string
	var chunks []string
	if artifact != nil && artifact.SourceHash == hash {
		text = artifact.NormalizedText
		chunks = artifact.Chunks
	} else {
		text = n.Content
	}

	in := Input{
		Title:           n.Title,
		Content:         text,
		SourceURL:       n.SourceURL,
		ImportTruncated: metadataImportTruncated(n.Metadata),
	}
	if a.stats != nil {
		if in.Keywords, err = a.stats.KeywordCount(ctx, noteID); err != nil {
			return nil, fmt.Errorf("quality: keyword count: %w", err)
		}
		if in.Links, err = a.stats.LinkCount(ctx, noteID); err != nil {
			return nil, fmt.Errorf("quality: link count: %w", err)
		}
		if in.HasEmbedding, err = a.stats.HasEmbedding(ctx, noteID); err != nil {
			return nil, fmt.Errorf("quality: embedding probe: %w", err)
		}
	}
	if a.embedder != nil {
		if min, med := a.chunkCoherence(ctx, chunks); med != nil {
			in.CoherenceMin, in.CoherenceMedian = min, med
		}
		if sim := a.titleSimilarity(ctx, n.Title, text); sim != nil {
			in.TitleTextSimilarity = sim
		}
	}

	signals := ComputeSignals(in, a.thresholds)
	verdict, reasons := Evaluate(signals)

	// Stop rule — per text version: unchanged signals suppress the entry; the
	// count of automatic entries is capped at MaxAutoAttempts.
	prior, err := a.logEntries(ctx, noteID)
	if err != nil {
		return nil, err
	}
	sameVersion := entriesForHash(prior, hash)
	autoAttempts := 0
	var lastSame *LogEntry
	for i := range sameVersion {
		e := sameVersion[i]
		if e.Trigger == TriggerAuto {
			autoAttempts++
		}
		if lastSame == nil {
			lastSame = &e
		}
	}
	changed := lastSame == nil || !SameSignals(lastSame.Record.Signals, signals)
	if trigger != TriggerManual {
		if autoAttempts >= MaxAutoAttempts {
			// Fourth+ automatic assessment of the same text never lands. A
			// gate that still holds flags the version for manual review.
			if verdict != VerdictCreate && lastSame != nil && a.log != nil {
				if _, err := a.log.MarkManualReview(ctx, noteID, hash); err != nil {
					return nil, fmt.Errorf("quality: flag manual review: %w", err)
				}
			}
			if lastSame != nil {
				rec := lastSame.Record
				rec.NeedsManualReview = rec.NeedsManualReview || verdict != VerdictCreate
				return &rec, nil
			}
			return nil, nil
		}
		if !changed {
			return nil, nil
		}
	}

	rec := Record{
		Signals:         signals,
		Gates:           append([]string{}, reasons...),
		Verdict:         verdict,
		Reasons:         append([]string{}, reasons...),
		Attempt:         autoAttempts + 1,
		ComputedAt:      a.now().UTC(),
		PipelineVersion: PipelineVersion,
	}

	entry := LogEntry{
		NoteID:     noteID,
		SourceHash: hash,
		Trigger:    trigger,
		Record:     rec,
		Changed:    changed,
	}
	if err := a.log.Append(ctx, entry); err != nil {
		return nil, fmt.Errorf("quality: append log: %w", err)
	}
	if a.artifactW != nil {
		if _, err := a.artifactW.SetQuality(ctx, noteID, hash, rec); err != nil {
			return nil, fmt.Errorf("quality: stamp artifact: %w", err)
		}
	}
	return &rec, nil
}

func (a *Assessor) logEntries(ctx context.Context, noteID uuid.UUID) ([]LogEntry, error) {
	if a.log == nil {
		return nil, nil
	}
	return a.log.List(ctx, noteID, LogKeepPerNote*2)
}

func entriesForHash(entries []LogEntry, hash string) []LogEntry {
	var out []LogEntry
	for _, e := range entries {
		if e.SourceHash == hash {
			out = append(out, e)
		}
	}
	return out
}

// chunkCoherence embeds adjacent chunk texts and reports min/median cosine.
// Empty title on purpose — coherence must measure chunks, not the title.
func (a *Assessor) chunkCoherence(ctx context.Context, chunks []string) (min, median *float64) {
	if len(chunks) < 2 {
		return nil, nil
	}
	vectors := make([][]float32, 0, len(chunks))
	for _, c := range chunks {
		v, err := a.embedder.Embed(ctx, c, "")
		if err != nil || len(v) == 0 {
			return nil, nil // model down or unusable — coherence stays null
		}
		vectors = append(vectors, v)
	}
	cos := make([]float64, 0, len(vectors)-1)
	for i := 1; i < len(vectors); i++ {
		cos = append(cos, cosine(vectors[i-1], vectors[i]))
	}
	sort.Float64s(cos)
	mn := cos[0]
	md := cos[len(cos)/2]
	return &mn, &md
}

// titleSimilarity — cosine(title, first titleSimilarityRunes of the body).
func (a *Assessor) titleSimilarity(ctx context.Context, title, text string) *float64 {
	if title == "" || text == "" {
		return nil
	}
	body := text
	if utf8.RuneCountInString(body) > titleSimilarityRunes {
		runes := []rune(body)
		body = string(runes[:titleSimilarityRunes])
	}
	vt, err := a.embedder.Embed(ctx, title, "")
	if err != nil || len(vt) == 0 {
		return nil
	}
	vb, err := a.embedder.Embed(ctx, body, "")
	if err != nil || len(vb) == 0 {
		return nil
	}
	sim := cosine(vt, vb)
	return &sim
}

func cosine(a, b []float32) float64 {
	var dot, na, nb float64
	for i := 0; i < len(a) && i < len(b); i++ {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / math.Sqrt(na) / math.Sqrt(nb)
}

// metadataImportTruncated reads metadata.import_truncated written by the
// stage-A import ({"sections_dropped": N} or a bare true).
func metadataImportTruncated(meta map[string]any) bool {
	v, ok := meta["import_truncated"]
	if !ok {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case map[string]any:
		return len(t) > 0
	default:
		return true
	}
}
