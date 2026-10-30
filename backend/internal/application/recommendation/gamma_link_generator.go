package recommendation

import (
	"context"
	"errors"
	"fmt"

	"knowledge-graph/internal/domain/link"

	"github.com/google/uuid"
)

// BatchLinkRepository extends the domain link repository with batch lookups.
type BatchLinkRepository interface {
	link.Repository
	link.SuppressionRepository
	FindBySourceIDs(ctx context.Context, sourceIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error)
}

// GammaLinkGenerator creates gamma (automatic, embedding-based) links for notes.
// It limits the outgoing degree per note: the transitive closure materialized
// view (note_links_closure) is bounded by pairs x depth since migration 033,
// but every extra edge still multiplies the number of reachable pairs.
type GammaLinkGenerator struct {
	embeddingRepo EmbeddingRepository
	linkRepo      BatchLinkRepository
	maxOutDegree  int
	minScore      float64
}

// NewGammaLinkGenerator creates a generator with the given out-degree cap and
// minimum similarity score. maxOutDegree must be > 0; keep it small (2) —
// every gamma edge makes the graph denser and the closure refresh (migration
// 033, shortest-paths bounded at depth 5) grows with it. minScore must be in
// [0, 1]; without it every note would get links to its nearest neighbours no
// matter how distant they are.
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

	suppressed, err := g.suppressedPairs(ctx, []uuid.UUID{noteID})
	if err != nil {
		return nil, fmt.Errorf("finding suppressions for %s: %w", noteID, err)
	}

	return g.saveMissingGammaLinks(ctx, noteID, similar, existing, suppressed, nil)
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

	suppressed, err := g.suppressedPairs(ctx, noteIDs)
	if err != nil {
		return nil, fmt.Errorf("finding suppressions in batch: %w", err)
	}

	created := make(map[uuid.UUID][]*link.Link, len(noteIDs))
	for _, noteID := range noteIDs {
		links, err := g.saveMissingGammaLinks(ctx, noteID, similarMap[noteID], existingMap[noteID], suppressed, nil)
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
// blockers because regeneration removes them before generating. The second
// return value is how many candidates were discarded by pair rejections.
func (g *GammaLinkGenerator) PlanForNotes(ctx context.Context, noteIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, int, error) {
	if len(noteIDs) == 0 {
		return nil, 0, nil
	}

	similarMap, err := g.embeddingRepo.FindSimilarNotesBatch(ctx, noteIDs, g.maxOutDegree)
	if err != nil {
		return nil, 0, fmt.Errorf("finding similar notes in batch: %w", err)
	}

	existingMap, err := g.linkRepo.FindBySourceIDs(ctx, noteIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("finding existing links in batch: %w", err)
	}

	suppressed, err := g.suppressedPairs(ctx, noteIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("finding suppressions in batch: %w", err)
	}

	planned := make(map[uuid.UUID][]*link.Link, len(noteIDs))
	suppressedCount := 0
	for _, noteID := range noteIDs {
		planned[noteID] = g.selectGammaLinks(noteID, similarMap[noteID], existingMap[noteID], suppressed, true, &suppressedCount)
	}
	return planned, suppressedCount, nil
}

// pairKey is a normalized (direction-free) pair of note IDs.
type pairKey struct{ a, b uuid.UUID }

// suppressedPairs loads rejections involving the given notes and returns the
// normalized pairs that block gamma ("related") proposals.
func (g *GammaLinkGenerator) suppressedPairs(ctx context.Context, noteIDs []uuid.UUID) (map[pairKey]bool, error) {
	suppressions, err := g.linkRepo.FindSuppressionsForNotes(ctx, noteIDs)
	if err != nil {
		return nil, err
	}
	set := make(map[pairKey]bool, len(suppressions))
	for _, s := range suppressions {
		// NULL-type rejections block any link; typed rejections block only
		// that type — the generator only ever proposes "related".
		if s.LinkType() == nil || *s.LinkType() == "related" {
			set[pairKey{s.NoteAID(), s.NoteBID()}] = true
		}
	}
	return set, nil
}

func (g *GammaLinkGenerator) saveMissingGammaLinks(ctx context.Context, sourceID uuid.UUID, similar []SimilarNote, existing []*link.Link, suppressed map[pairKey]bool, suppressedCount *int) ([]*link.Link, error) {
	planned := g.selectGammaLinks(sourceID, similar, existing, suppressed, false, suppressedCount)

	for _, l := range planned {
		if err := g.linkRepo.Save(ctx, l); err != nil {
			// A manual link created concurrently on the same pair wins the
			// race (UNIQUE on source+target+type) — the human's decision is
			// stronger than the proposal, so the duplicate is skipped instead
			// of aborting the whole regeneration run.
			if errors.Is(err, link.ErrDuplicateLink) {
				continue
			}
			return nil, fmt.Errorf("saving gamma link %s -> %s: %w", sourceID, l.TargetNoteID(), err)
		}
	}
	return planned, nil
}

// selectGammaLinks picks up to maxOutDegree candidates from similar, skipping
// the source itself, scores below minScore, targets that already have a
// "related" link from this source and pairs rejected by a human (in either
// direction, with this type or NULL). When ignoreGammaExisting is true,
// existing gamma links do not block a candidate (they are about to be
// regenerated). suppressedCount, when non-nil, is incremented for every
// candidate discarded by a rejection.
func (g *GammaLinkGenerator) selectGammaLinks(sourceID uuid.UUID, similar []SimilarNote, existing []*link.Link, suppressed map[pairKey]bool, ignoreGammaExisting bool, suppressedCount *int) []*link.Link {
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
		pa, pb := link.NormalizePair(sourceID, s.NoteID)
		if suppressed[pairKey{pa, pb}] {
			if suppressedCount != nil {
				*suppressedCount++
			}
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
