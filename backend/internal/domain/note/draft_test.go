package note

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDraft(t *testing.T) {
	noteID := uuid.New()
	userID := uuid.New()
	d := NewDraft(noteID, userID, "content", "title")

	assert.NotEqual(t, uuid.Nil, d.ID())
	assert.Equal(t, noteID, d.NoteID())
	assert.Equal(t, userID, d.UserID())
	assert.Equal(t, "content", d.Content())
	assert.Equal(t, "title", d.Title())
	assert.Equal(t, DraftStateActive, d.State())
	assert.False(t, d.CreatedAt().IsZero())
	assert.False(t, d.UpdatedAt().IsZero())
	assert.Empty(t, d.Events())
}

func TestReconstructDraft(t *testing.T) {
	now := time.Now()
	d := ReconstructDraft(uuid.New(), uuid.New(), uuid.New(), "c", "t", DraftStatePublished, now, now)
	assert.Equal(t, DraftStatePublished, d.State())
}

func TestDraftStateTransitions(t *testing.T) {
	d := NewDraft(uuid.New(), uuid.New(), "content", "title")

	assert.Error(t, d.MarkAsPublished())
	assert.Error(t, d.MarkAsConflict())
	assert.Error(t, d.ResolveConflict())

	require.NoError(t, d.StartPublishing())
	assert.Equal(t, DraftStatePublishing, d.State())

	assert.Error(t, d.StartPublishing())
	assert.Error(t, d.UpdateContent("new"))

	require.NoError(t, d.MarkAsConflict())
	assert.Equal(t, DraftStateConflict, d.State())

	require.NoError(t, d.ResolveConflict())
	assert.Equal(t, DraftStateActive, d.State())

	require.NoError(t, d.StartPublishing())
	require.NoError(t, d.MarkAsPublished())
	assert.Equal(t, DraftStatePublished, d.State())
	assert.Len(t, d.Events(), 1)

	d.ClearEvents()
	assert.Empty(t, d.Events())
}

func TestDraftUpdate(t *testing.T) {
	d := NewDraft(uuid.New(), uuid.New(), "old content", "old title")

	require.NoError(t, d.UpdateContent("new content"))
	assert.Equal(t, "new content", d.Content())

	require.NoError(t, d.UpdateTitle("new title"))
	assert.Equal(t, "new title", d.Title())

	d.StartPublishing()
	assert.Error(t, d.UpdateContent("x"))
	assert.Error(t, d.UpdateTitle("y"))
}
