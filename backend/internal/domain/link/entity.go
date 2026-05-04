package link

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	id           uuid.UUID
	sourceNoteID uuid.UUID
	targetNoteID uuid.UUID
	linkType     LinkType
	weight       Weight
	metadata     Metadata
	creatorID    *uuid.UUID
	createdAt    time.Time
}

func NewLink(sourceID, targetID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata) *Link {
	return &Link{
		id:           uuid.New(),
		sourceNoteID: sourceID,
		targetNoteID: targetID,
		linkType:     linkType,
		weight:       weight,
		metadata:     metadata,
		creatorID:    nil,
		createdAt:    time.Now(),
	}
}

func NewLinkWithCreator(sourceID, targetID, creatorID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata) *Link {
	return &Link{
		id:           uuid.New(),
		sourceNoteID: sourceID,
		targetNoteID: targetID,
		linkType:     linkType,
		weight:       weight,
		metadata:     metadata,
		creatorID:    &creatorID,
		createdAt:    time.Now(),
	}
}

// ReconstructLink восстанавливает связь из сохранённых данных
func ReconstructLink(id uuid.UUID, sourceID, targetID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata, createdAt time.Time) *Link {
	return &Link{
		id:           id,
		sourceNoteID: sourceID,
		targetNoteID: targetID,
		linkType:     linkType,
		weight:       weight,
		metadata:     metadata,
		creatorID:    nil,
		createdAt:    createdAt,
	}
}

func ReconstructLinkWithCreator(id uuid.UUID, sourceID, targetID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata, creatorID *uuid.UUID, createdAt time.Time) *Link {
	return &Link{
		id:           id,
		sourceNoteID: sourceID,
		targetNoteID: targetID,
		linkType:     linkType,
		weight:       weight,
		metadata:     metadata,
		creatorID:    creatorID,
		createdAt:    createdAt,
	}
}

// Геттеры
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

func (l *Link) CreatorID() *uuid.UUID {
	return l.creatorID
}

func (l *Link) SetCreatorID(creatorID uuid.UUID) {
	l.creatorID = &creatorID
}

// UpdateWeight обновляет вес связи
func (l *Link) UpdateWeight(newWeight Weight) {
	l.weight = newWeight
}
