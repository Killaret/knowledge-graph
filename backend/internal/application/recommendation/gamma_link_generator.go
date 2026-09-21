package recommendation

import (
	"context"
	"fmt"

	"knowledge-graph/internal/domain/link"

	"github.com/google/uuid"
)

// BatchLinkRepository extends the domain link repository with batch lookups.
type BatchLinkRepository interface {
	link.Repository
	FindBySourceIDs(ctx context.Context, sourceIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error)
}

// GammaLinkGenerator creates gamma (automatic, embedding-based) links for notes.
// It limits the outgoing degree per note to avoid blowing up the transitive
// closure materialized view (note_links_closure), which is exponential in the
// worst case for dense graphs.
type GammaLinkGenerator struct {
	embeddingRepo EmbeddingRepository
	linkRepo      BatchLinkRepository
	maxOutDegree  int
	minScore      float64
}

// NewGammaLinkGenerator creates a generator with the given out-degree cap and
// minimum similarity score. maxOutDegree must be > 0; values above 2 should not
// be used with the current materialized view because path enumeration becomes
// prohibitively expensive. minScore must be in [0, 1]; without it every note
// would get links to its nearest neighbours no matter how distant they are.
func NewGammaLinkGenerator(embeddingRepo EmbeddingRepository, linkRepo BatchLinkRepository, maxOutDegree int, minScore float64) *GammaLinkGenerator {
	if maxOutDegree <= 0 {
		maxOutDegree = 2
	}
	if minScore < 0 {
		minScore = 0
	}
	if minScore > 1 {
		minScore = 1
	}
	return &GammaLinkGenerator{
		embeddingRepo: embeddingRepo,
		linkRepo:      linkRepo,
		maxOutDegree:  maxOutDegree,
		minScore:      minScore,
	}
}

// GenerateForNote creates up to maxOutDegree gamma links from noteID and
// returns the created links so the caller can publish LinkCreated events.
func (g *GammaLinkGenerator) GenerateForNote(ctx context.Context, noteID uuid.UUID) ([]*link.Link, error) {
	similar, err := g.embeddingRepo.FindSimilarNotes(ctx, noteID, g.maxOutDegree)
	if err != nil {
		return nil, fmt.Errorf("finding similar notes for %s: %w", noteID, err)
	}

	existing, err := g.linkRepo.FindBySource(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("finding existing links for %s: %w", noteID, err)
	}

	return g.saveMissingGammaLinks(ctx, noteID, similar, existing)
}

// GenerateForNotes creates gamma links for multiple notes in a single batch
// query and returns the created links grouped by source note.
func (g *GammaLinkGenerator) GenerateForNotes(ctx context.Context, noteIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error) {
	if len(noteIDs) == 0 {
		return nil, nil
	}

	similarMap, err := g.embeddingRepo.FindSimilarNotesBatch(ctx, noteIDs, g.maxOutDegree)
	if err != nil {
		return nil, fmt.Errorf("finding similar notes in batch: %w", err)
	}

	existingMap, err := g.linkRepo.FindBySourceIDs(ctx, noteIDs)
	if err != nil {
		return nil, fmt.Errorf("finding existing links in batch: %w", err)
	}

	created := make(map[uuid.UUID][]*link.Link, len(noteIDs))
	for _, noteID := range noteIDs {
		links, err := g.saveMissingGammaLinks(ctx, noteID, similarMap[noteID], existingMap[noteID])
		if err != nil {
			return created, err
		}
		created[noteID] = links
	}
	return created, nil
}

// PlanForNotes computes which gamma links would be created for each note if
// all existing gamma links were deleted first. Nothing is saved — used by the
// regenerate command's --dry-run mode. Existing gamma links are ignored as
// blockers because regeneration removes them before generating.
func (g *GammaLinkGenerator) PlanForNotes(ctx context.Context, noteIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error) {
	if len(noteIDs) == 0 {
		return nil, nil
	}

	similarMap, err := g.embeddingRepo.FindSimilarNotesBatch(ctx, noteIDs, g.maxOutDegree)
	if err != nil {
		return nil, fmt.Errorf("finding similar notes in batch: %w", err)
	}

	existingMap, err := g.linkRepo.FindBySourceIDs(ctx, noteIDs)
	if err != nil {
		return nil, fmt.Errorf("finding existing links in batch: %w", err)
	}

	planned := make(map[uuid.UUID][]*link.Link, len(noteIDs))
	for _, noteID := range noteIDs {
		planned[noteID] = g.selectGammaLinks(noteID, similarMap[noteID], existingMap[noteID], true)
	}
	return planned, nil
}

func (g *GammaLinkGenerator) saveMissingGammaLinks(ctx context.Context, sourceID uuid.UUID, similar []SimilarNote, existing []*link.Link) ([]*link.Link, error) {
	planned := g.selectGammaLinks(sourceID, similar, existing, false)

	for _, l := range planned {
		if err := g.linkRepo.Save(ctx, l); err != nil {
			return nil, fmt.Errorf("saving gamma link %s -> %s: %w", sourceID, l.TargetNoteID(), err)
		}
	}
	return planned, nil
}

// selectGammaLinks picks up to maxOutDegree candidates from similar, skipping
// the source itself, scores below minScore and targets that already have a
// "related" link from this source. When ignoreGammaExisting is true, existing
// gamma links do not block a candidate (they are about to be regenerated).
func (g *GammaLinkGenerator) selectGammaLinks(sourceID uuid.UUID, similar []SimilarNote, existing []*link.Link, ignoreGammaExisting bool) []*link.Link {
	relatedType, err := link.NewLinkType("related")
	if err != nil {
		return nil
	}

	existingTargets := make(map[uuid.UUID]bool)
	for _, l := range existing {
		if l == nil || l.LinkType().String() != relatedType.String() {
			continue
		}
		if ignoreGammaExisting && l.SourceType().IsGamma() {
			continue
		}
		existingTargets[l.TargetNoteID()] = true
	}

	metadata, err := link.NewMetadata(map[string]interface{}{"source": "gamma"})
	if err != nil {
		return nil
	}

	var selected []*link.Link
	for _, s := range similar {
		if len(selected) >= g.maxOutDegree {
			break
		}
		if s.NoteID == sourceID {
			continue
		}
		if existingTargets[s.NoteID] {
			continue
		}
		if s.Score < g.minScore {
			continue
		}

		weight, err := link.NewWeight(s.Score)
		if err != nil {
			continue
		}

		l := link.NewGammaLink(sourceID, s.NoteID, relatedType, weight, metadata)
		existingTargets[s.NoteID] = true
		selected = append(selected, l)
	}

	return selected
}
