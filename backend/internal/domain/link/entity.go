package link

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	id               uuid.UUID
	sourceNoteID     uuid.UUID
	targetNoteID     uuid.UUID
	linkType         LinkType
	weight           Weight
	metadata         Metadata
	sourceType       SourceType
	creatorID        *uuid.UUID
	createdAt        time.Time
	updatedAt        time.Time
	lastWeightUpdate *time.Time
}

func NewLink(sourceID, targetID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata) *Link {
	now := time.Now()
	return &Link{
		id:           uuid.New(),
		sourceNoteID: sourceID,
		targetNoteID: targetID,
		linkType:     linkType,
		weight:       weight,
		metadata:     metadata,
		sourceType:   DefaultSourceType(),
		creatorID:    nil,
		createdAt:    now,
		updatedAt:    now,
	}
}

func NewLinkWithCreator(sourceID, targetID, creatorID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata) *Link {
	now := time.Now()
	return &Link{
		id:           uuid.New(),
		sourceNoteID: sourceID,
		targetNoteID: targetID,
		linkType:     linkType,
		weight:       weight,
		metadata:     metadata,
		sourceType:   DefaultSourceType(),
		creatorID:    &creatorID,
		createdAt:    now,
		updatedAt:    now,
	}
}

// NewGammaLink creates a link from recommendations (gamma source type)
func NewGammaLink(sourceID, targetID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata) *Link {
	sourceType, _ := NewSourceType("gamma")
	now := time.Now()
	return &Link{
		id:           uuid.New(),
		sourceNoteID: sourceID,
		targetNoteID: targetID,
		linkType:     linkType,
		weight:       weight,
		metadata:     metadata,
		sourceType:   sourceType,
		creatorID:    nil,
		createdAt:    now,
		updatedAt:    now,
	}
}

// ReconstructLink reconstructs a link from saved data
func ReconstructLink(id uuid.UUID, sourceID, targetID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata, sourceType SourceType, createdAt, updatedAt time.Time, lastWeightUpdate *time.Time) *Link {
	return &Link{
		id:               id,
		sourceNoteID:     sourceID,
		targetNoteID:     targetID,
		linkType:         linkType,
		weight:           weight,
		metadata:         metadata,
		sourceType:       sourceType,
		creatorID:        nil,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		lastWeightUpdate: lastWeightUpdate,
	}
}

func ReconstructLinkWithCreator(id uuid.UUID, sourceID, targetID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata, sourceType SourceType, creatorID *uuid.UUID, createdAt, updatedAt time.Time, lastWeightUpdate *time.Time) *Link {
	return &Link{
		id:               id,
		sourceNoteID:     sourceID,
		targetNoteID:     targetID,
		linkType:         linkType,
		weight:           weight,
		metadata:         metadata,
		sourceType:       sourceType,
		creatorID:        creatorID,
		createdAt:        createdAt,
		updatedAt:        updatedAt,
		lastWeightUpdate: lastWeightUpdate,
	}
}

// Getters
func (l *Link) ID() uuid.UUID {
	return l.id
}

func (l *Link) SourceNoteID() uuid.UUID {
	return l.sourceNoteID
}

func (l *Link) TargetNoteID() uuid.UUID {
	return l.targetNoteID
}

func (l *Link) LinkType() LinkType {
	return l.linkType
}

func (l *Link) Weight() Weight {
	return l.weight
}

func (l *Link) Metadata() Metadata {
	return l.metadata
}

func (l *Link) CreatedAt() time.Time {
	return l.createdAt
}

func (l *Link) UpdatedAt() time.Time {
	return l.updatedAt
}

func (l *Link) LastWeightUpdate() *time.Time {
	return l.lastWeightUpdate
}

func (l *Link) CreatorID() *uuid.UUID {
	return l.creatorID
}

func (l *Link) SetCreatorID(creatorID uuid.UUID) {
	l.creatorID = &creatorID
}

func (l *Link) SourceType() SourceType {
	return l.sourceType
}

func (l *Link) SetSourceType(sourceType SourceType) {
	l.sourceType = sourceType
}

// UpdateWeight updates the link weight and records when the weight was last changed.
func (l *Link) UpdateWeight(newWeight Weight) {
	now := time.Now()
	l.weight = newWeight
	l.updatedAt = now
	l.lastWeightUpdate = &now
}

// UpdateLinkType updates the link type and touches updated_at.
func (l *Link) UpdateLinkType(newLinkType LinkType) {
	l.linkType = newLinkType
	l.updatedAt = time.Now()
}

// HasGammaProvenance reports whether the link records that the model once
// proposed this pair (either still gamma, or promoted earlier).
func (l *Link) HasGammaProvenance() bool {
	_, ok := l.metadata.Value()["gamma"]
	return ok
}

// gammaProvenance builds the metadata.gamma payload: the score the model
// proposed the pair with and when the proposal was created.
func (l *Link) gammaProvenance() map[string]interface{} {
	return map[string]interface{}{
		"score":        l.weight.Value(),
		"generated_at": l.createdAt.UTC().Format(time.RFC3339Nano),
	}
}

// PromoteToUser turns a gamma link into a user-confirmed link in place:
// weight, type and metadata come from the manual request, created_at and id
// are preserved, and the gamma origin is kept in metadata.gamma.
func (l *Link) PromoteToUser(creatorID *uuid.UUID, newLinkType LinkType, newWeight Weight, newMetadata Metadata) {
	provenance := l.gammaProvenance()
	merged := map[string]interface{}{"gamma": provenance}
	for k, v := range newMetadata.Value() {
		merged[k] = v
	}
	l.metadata, _ = NewMetadata(merged)
	l.linkType = newLinkType
	l.weight = newWeight
	l.sourceType = DefaultSourceType()
	l.creatorID = creatorID
	l.updatedAt = time.Now()
}

// InheritGammaProvenance copies the gamma origin of another (removed) link
// into this link's metadata without overwriting request-provided keys.
func (l *Link) InheritGammaProvenance(from *Link) {
	merged := make(map[string]interface{}, len(l.metadata.Value())+1)
	for k, v := range l.metadata.Value() {
		merged[k] = v
	}
	merged["gamma"] = from.gammaProvenance()
	l.metadata, _ = NewMetadata(merged)
}
