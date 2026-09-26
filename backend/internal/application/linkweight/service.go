// Package linkweight provides application logic for recalculating link
// weights using semantic similarity between connected notes.
package linkweight

import (
	"context"
	"fmt"
	"log"

	appevents "knowledge-graph/internal/application/events"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
)

// SimilarityClient abstracts the call to the NLP service for text similarity.
type SimilarityClient interface {
	Similarity(ctx context.Context, textA, textB string) (float64, error)
}

// Recalculator recalculates link weights from note content similarity.
type Recalculator struct {
	linkRepo       link.Repository
	noteRepo       note.Repository
	nlpClient      SimilarityClient
	eventPublisher appevents.Publisher
}

// NewRecalculator creates a new link-weight recalculator.
func NewRecalculator(linkRepo link.Repository, noteRepo note.Repository, nlpClient SimilarityClient) *Recalculator {
	return &Recalculator{
		linkRepo:  linkRepo,
		noteRepo:  noteRepo,
		nlpClient: nlpClient,
	}
}

// SetEventPublisher sets the optional graph event publisher. A weight change
// is a write to graph data — it must publish LinkUpdated or the change stays
// invisible to open graphs until TTL.
func (r *Recalculator) SetEventPublisher(p appevents.Publisher) {
	r.eventPublisher = p
}

// RecalculateForNote recalculates weights for all links connected to the note.
func (r *Recalculator) RecalculateForNote(ctx context.Context, noteID uuid.UUID) error {
	outgoing, err := r.linkRepo.FindBySource(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to find outgoing links: %w", err)
	}

	incoming, err := r.linkRepo.FindByTarget(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to find incoming links: %w", err)
	}

	links := make([]*link.Link, 0, len(outgoing)+len(incoming))
	links = append(links, outgoing...)
	links = append(links, incoming...)

	for _, l := range links {
		source, err := r.noteRepo.FindByID(ctx, l.SourceNoteID())
		if err != nil {
			log.Printf("[LinkWeightRecalculator] failed to fetch source note %s: %v", l.SourceNoteID(), err)
			continue
		}
		if source == nil {
			continue
		}

		target, err := r.noteRepo.FindByID(ctx, l.TargetNoteID())
		if err != nil {
			log.Printf("[LinkWeightRecalculator] failed to fetch target note %s: %v", l.TargetNoteID(), err)
			continue
		}
		if target == nil {
			continue
		}

		textA := noteText(source)
		textB := noteText(target)

		similarity, err := r.nlpClient.Similarity(ctx, textA, textB)
		if err != nil {
			log.Printf("[LinkWeightRecalculator] similarity failed for link %s: %v", l.ID(), err)
			continue
		}

		weight, err := link.NewWeight(similarity)
		if err != nil {
			log.Printf("[LinkWeightRecalculator] invalid weight %f for link %s: %v", similarity, l.ID(), err)
			continue
		}

		l.UpdateWeight(weight)
		if err := r.linkRepo.Update(ctx, l); err != nil {
			log.Printf("[LinkWeightRecalculator] failed to update link %s: %v", l.ID(), err)
			continue
		}

		if r.eventPublisher != nil {
			userID := ""
			if source.CreatorID() != nil {
				userID = source.CreatorID().String()
			} else if target.CreatorID() != nil {
				userID = target.CreatorID().String()
			}
			if err := r.eventPublisher.PublishLinkUpdated(ctx, l.SourceNoteID().String(), l.TargetNoteID().String(), userID); err != nil {
				log.Printf("[LinkWeightRecalculator] failed to publish LinkUpdated for %s: %v", l.ID(), err)
			}
		}

		log.Printf("[LinkWeightRecalculator] updated link %s weight to %.4f", l.ID(), weight.Value())
	}

	return nil
}

func noteText(n *note.Note) string {
	text := n.Title().String() + " " + n.Content().String()
	return text
}
