package mongo

import (
	"testing"
	"time"

	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDraftToModelAndBack(t *testing.T) {
	noteID := uuid.New()
	userID := uuid.New()

	draft := note.NewDraft(noteID, userID, "draft content", "Draft title")
	draft.StartPublishing()

	model := draftToModel(draft)
	assert.Equal(t, draft.ID(), model.ID)
	assert.Equal(t, draft.Content(), model.Content)
	assert.Equal(t, string(draft.State()), model.State)

	reconstructed := model.toDomain()
	assert.Equal(t, draft.ID(), reconstructed.ID())
	assert.Equal(t, draft.NoteID(), reconstructed.NoteID())
	assert.Equal(t, draft.UserID(), reconstructed.UserID())
	assert.Equal(t, draft.Content(), reconstructed.Content())
	assert.Equal(t, draft.Title(), reconstructed.Title())
	assert.Equal(t, draft.State(), reconstructed.State())
}

func TestDraftModel_toDomain(t *testing.T) {
	model := &draftModel{
		ID:        uuid.New(),
		NoteID:    uuid.New(),
		UserID:    uuid.New(),
		Content:   "content",
		Title:     "title",
		State:     string(note.DraftStateActive),
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	d := model.toDomain()
	assert.Equal(t, model.ID, d.ID())
	assert.Equal(t, note.DraftStateActive, d.State())
}
