package link

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewLink(t *testing.T) {
	sourceID := uuid.New()
	targetID := uuid.New()
	linkType, _ := NewLinkType("reference")
	weight, _ := NewWeight(0.8)
	metadata, _ := NewMetadata(nil)

	link := NewLink(sourceID, targetID, linkType, weight, metadata)

	if link.ID() == uuid.Nil {
		t.Error("ID should not be nil")
	}
	if link.SourceNoteID() != sourceID {
		t.Error("source ID mismatch")
	}
	if link.TargetNoteID() != targetID {
		t.Error("target ID mismatch")
	}
	if link.LinkType().String() != "reference" {
		t.Error("link type mismatch")
	}
	if link.Weight().Value() != 0.8 {
		t.Error("weight mismatch")
	}
}

func TestLinkUpdateWeight(t *testing.T) {
	sourceID := uuid.New()
	targetID := uuid.New()
	linkType, _ := NewLinkType("reference")
	weight, _ := NewWeight(0.5)
	metadata, _ := NewMetadata(nil)
	link := NewLink(sourceID, targetID, linkType, weight, metadata)

	newWeight, _ := NewWeight(0.9)
	link.UpdateWeight(newWeight)
	if link.Weight().Value() != 0.9 {
		t.Error("weight not updated")
	}
}
