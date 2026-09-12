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
}

// NewGammaLinkGenerator creates a generator with the given out-degree cap.
// maxOutDegree must be > 0; values above 2 should not be used with the current
// materialized view because path enumeration becomes prohibitively expensive.
func NewGammaLinkGenerator(embeddingRepo EmbeddingRepository, linkRepo BatchLinkRepository, maxOutDegree int) *GammaLinkGenerator {
	if maxOutDegree <= 0 {
		maxOutDegree = 2
	}
	return &GammaLinkGenerator{
		embeddingRepo: embeddingRepo,
		linkRepo:      linkRepo,
		maxOutDegree:  maxOutDegree,
	}
}

// GenerateForNote creates up to MaxGammaOutDegree gamma links from noteID.
func (g *GammaLinkGenerator) GenerateForNote(ctx context.Context, noteID uuid.UUID) error {
	similar, err := g.embeddingRepo.FindSimilarNotes(ctx, noteID, g.maxOutDegree)
	if err != nil {
		return fmt.Errorf("finding similar notes for %s: %w", noteID, err)
	}

	existing, err := g.linkRepo.FindBySource(ctx, noteID)
	if err != nil {
		return fmt.Errorf("finding existing links for %s: %w", noteID, err)
	}

	return g.saveMissingGammaLinks(ctx, noteID, similar, existing)
}

// GenerateForNotes creates gamma links for multiple notes in a single batch query.
func (g *GammaLinkGenerator) GenerateForNotes(ctx context.Context, noteIDs []uuid.UUID) error {
	if len(noteIDs) == 0 {
		return nil
	}

	similarMap, err := g.embeddingRepo.FindSimilarNotesBatch(ctx, noteIDs, g.maxOutDegree)
	if err != nil {
		return fmt.Errorf("finding similar notes in batch: %w", err)
	}

	existingMap, err := g.linkRepo.FindBySourceIDs(ctx, noteIDs)
	if err != nil {
		return fmt.Errorf("finding existing links in batch: %w", err)
	}

	for _, noteID := range noteIDs {
		if err := g.saveMissingGammaLinks(ctx, noteID, similarMap[noteID], existingMap[noteID]); err != nil {
			return err
		}
	}
	return nil
}

func (g *GammaLinkGenerator) saveMissingGammaLinks(ctx context.Context, sourceID uuid.UUID, similar []SimilarNote, existing []*link.Link) error {
	relatedType, err := link.NewLinkType("related")
	if err != nil {
		return fmt.Errorf("creating link type: %w", err)
	}

	existingTargets := make(map[uuid.UUID]bool)
	for _, l := range existing {
		if l != nil && l.LinkType().String() == relatedType.String() {
			existingTargets[l.TargetNoteID()] = true
		}
	}

	metadata, err := link.NewMetadata(map[string]interface{}{"source": "gamma"})
	if err != nil {
		return fmt.Errorf("creating link metadata: %w", err)
	}

	created := 0
	for _, s := range similar {
		if created >= g.maxOutDegree {
			break
		}
		if s.NoteID == sourceID {
			continue
		}
		if existingTargets[s.NoteID] {
			continue
		}
		if s.Score <= 0 {
			continue
		}

		weight, err := link.NewWeight(s.Score)
		if err != nil {
			continue
		}

		l := link.NewGammaLink(sourceID, s.NoteID, relatedType, weight, metadata)
		if err := g.linkRepo.Save(ctx, l); err != nil {
			return fmt.Errorf("saving gamma link %s -> %s: %w", sourceID, s.NoteID, err)
		}
		existingTargets[s.NoteID] = true
		created++
	}

	return nil
}
