package link

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewLinkWithCreator(t *testing.T) {
	sourceID := uuid.New()
	targetID := uuid.New()
	creatorID := uuid.New()
	linkType, _ := NewLinkType("reference")
	weight, _ := NewWeight(0.8)
	metadata, _ := NewMetadata(map[string]interface{}{"note": "test"})

	link := NewLinkWithCreator(sourceID, targetID, creatorID, linkType, weight, metadata)

	assert.NotEqual(t, uuid.Nil, link.ID())
	assert.Equal(t, sourceID, link.SourceNoteID())
	assert.Equal(t, targetID, link.TargetNoteID())
	assert.Equal(t, creatorID, *link.CreatorID())
	assert.Equal(t, "reference", link.LinkType().String())
	assert.Equal(t, 0.8, link.Weight().Value())
	assert.Equal(t, "test", link.Metadata().Value()["note"])
}

func TestNewGammaLink(t *testing.T) {
	sourceID := uuid.New()
	targetID := uuid.New()
	linkType, _ := NewLinkType("related")
	weight, _ := NewWeight(0.5)
	metadata, _ := NewMetadata(nil)

	link := NewGammaLink(sourceID, targetID, linkType, weight, metadata)

	assert.True(t, link.SourceType().IsGamma())
	assert.False(t, link.SourceType().IsUser())
	assert.Equal(t, sourceID, link.SourceNoteID())
}

func TestReconstructLink(t *testing.T) {
	id := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()
	linkType, _ := NewLinkType("dependency")
	weight, _ := NewWeight(0.3)
	metadata, _ := NewMetadata(nil)
	sourceType, _ := NewSourceType("user")
	createdAt := time.Now()

	link := ReconstructLink(id, sourceID, targetID, linkType, weight, metadata, sourceType, createdAt)

	assert.Equal(t, id, link.ID())
	assert.Equal(t, sourceID, link.SourceNoteID())
	assert.Equal(t, targetID, link.TargetNoteID())
	assert.Equal(t, linkType, link.LinkType())
	assert.Equal(t, weight, link.Weight())
	assert.Equal(t, sourceType, link.SourceType())
	assert.True(t, createdAt.Equal(link.CreatedAt()) || createdAt.Before(link.CreatedAt()))
}

func TestReconstructLinkWithCreator(t *testing.T) {
	id := uuid.New()
	sourceID := uuid.New()
	targetID := uuid.New()
	creatorID := uuid.New()
	linkType, _ := NewLinkType("custom")
	weight, _ := NewWeight(0.9)
	metadata, _ := NewMetadata(nil)
	sourceType, _ := NewSourceType("gamma")
	createdAt := time.Now()

	link := ReconstructLinkWithCreator(id, sourceID, targetID, linkType, weight, metadata, sourceType, &creatorID, createdAt)

	assert.Equal(t, id, link.ID())
	assert.Equal(t, &creatorID, link.CreatorID())
}

func TestLink_Setters(t *testing.T) {
	sourceID := uuid.New()
	targetID := uuid.New()
	creatorID := uuid.New()
	linkType, _ := NewLinkType("reference")
	weight, _ := NewWeight(0.5)
	metadata, _ := NewMetadata(nil)

	link := NewLink(sourceID, targetID, linkType, weight, metadata)

	link.SetCreatorID(creatorID)
	assert.Equal(t, creatorID, *link.CreatorID())

	newSourceType, _ := NewSourceType("gamma")
	link.SetSourceType(newSourceType)
	assert.Equal(t, "gamma", link.SourceType().String())
}

func TestLink_Metadata_Value(t *testing.T) {
	metadata, _ := NewMetadata(map[string]interface{}{"key": "value"})
	assert.Equal(t, "value", metadata.Value()["key"])
}
