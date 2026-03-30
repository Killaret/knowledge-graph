package linkhandler

import (
	"time"

	"knowledge-graph/internal/domain/link"
)

type linkResponse struct {
	ID           string                 `json:"id"`
	SourceNoteID string                 `json:"source_note_id"`
	TargetNoteID string                 `json:"target_note_id"`
	LinkType     string                 `json:"link_type"`
	Weight       float64                `json:"weight"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

func toLinkResponse(l *link.Link) linkResponse {
	return linkResponse{
		ID:           l.ID().String(),
		SourceNoteID: l.SourceNoteID().String(),
		TargetNoteID: l.TargetNoteID().String(),
		LinkType:     l.LinkType().String(),
		Weight:       l.Weight().Value(),
		Metadata:     l.Metadata().Value(),
		CreatedAt:    l.CreatedAt(),
	}
}
