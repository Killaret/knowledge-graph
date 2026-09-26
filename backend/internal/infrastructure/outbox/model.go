// Package outbox implements the transactional outbox for graph events
// (SYNC-1 stage A2, owner decision 71). Repository writes run inside the
// decorator's transaction together with a graph_outbox row, so a committed
// write always has its event and a rolled-back write never does. The relayer
// publishes unsent rows to Redis; a crash between commit and publish is
// repaired by the next relay pass (at-least-once delivery).
package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Event types mirror events.Publisher's channel vocabulary — the graph
// service subscriber switches on these exact strings.
const (
	EventNoteCreated = "NoteCreated"
	EventNoteUpdated = "NoteUpdated"
	EventNoteDeleted = "NoteDeleted"
	EventLinkCreated = "LinkCreated"
	EventLinkUpdated = "LinkUpdated"
	EventLinkDeleted = "LinkDeleted"
)

// Model is a row of the graph_outbox table (migration 036).
type Model struct {
	ID        int64          `gorm:"primaryKey;autoIncrement"`
	EventType string         `gorm:"column:event_type;not null"`
	EntityID  uuid.UUID      `gorm:"type:uuid;column:entity_id;not null"`
	UserID    *uuid.UUID     `gorm:"type:uuid;column:user_id"`
	Payload   datatypes.JSON `gorm:"type:jsonb;column:payload;not null"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	SentAt    *time.Time     `gorm:"column:sent_at"`
}

func (Model) TableName() string {
	return "graph_outbox"
}

// noteEvent builds a row whose payload matches events.NoteEventPayload.
func noteEvent(eventType string, noteID uuid.UUID, userID *uuid.UUID) Model {
	uid := ""
	if userID != nil {
		uid = userID.String()
	}
	payload, _ := json.Marshal(map[string]string{"note_id": noteID.String(), "user_id": uid})
	return Model{EventType: eventType, EntityID: noteID, UserID: userID, Payload: datatypes.JSON(payload)}
}

// linkEvent builds a row whose payload matches events.LinkEventPayload.
func linkEvent(eventType string, linkID, sourceID, targetID uuid.UUID, userID *uuid.UUID) Model {
	uid := ""
	if userID != nil {
		uid = userID.String()
	}
	payload, _ := json.Marshal(map[string]string{
		"source_note_id": sourceID.String(),
		"target_note_id": targetID.String(),
		"user_id":        uid,
	})
	return Model{EventType: eventType, EntityID: linkID, UserID: userID, Payload: datatypes.JSON(payload)}
}

// insertEvents records the outbox rows inside the caller's transaction.
func insertEvents(tx *gorm.DB, rows []Model) error {
	if len(rows) == 0 {
		return nil
	}
	return tx.Create(&rows).Error
}
