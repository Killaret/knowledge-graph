package share

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNoteShare(t *testing.T) {
	id := uuid.New()
	noteID := uuid.New()
	by := uuid.New()
	with := uuid.New()

	s, err := NewNoteShare(id, noteID, by, with, "read", nil)
	require.NoError(t, err)
	assert.Equal(t, id, s.ID())
	assert.Equal(t, noteID, s.NoteID())
	assert.Equal(t, by, s.SharedByUserID())
	assert.Equal(t, with, s.SharedWithUserID())
	assert.Equal(t, "read", s.Permission())
	assert.False(t, s.CreatedAt().IsZero())

	_, err = NewNoteShare(id, noteID, by, with, "admin", nil)
	assert.ErrorIs(t, err, ErrInvalidPermission)
}

func TestNoteShare_UpdatePermission(t *testing.T) {
	s, _ := NewNoteShare(uuid.New(), uuid.New(), uuid.New(), uuid.New(), "read", nil)
	err := s.UpdatePermission("write")
	require.NoError(t, err)
	assert.Equal(t, "write", s.Permission())

	assert.ErrorIs(t, s.UpdatePermission("delete"), ErrInvalidPermission)
}

func TestNoteShare_SetSharedWithLogin(t *testing.T) {
	s, _ := NewNoteShare(uuid.New(), uuid.New(), uuid.New(), uuid.New(), "read", nil)
	s.SetSharedWithLogin("alice")
	assert.Equal(t, "alice", s.SharedWithLogin())
}

func TestNewShareLink(t *testing.T) {
	id := uuid.New()
	noteID := uuid.New()
	by := uuid.New()
	expires := time.Now().Add(time.Hour)
	maxUses := 5

	l, err := NewShareLink(id, noteID, by, "token-123", "read", &expires, &maxUses, 0)
	require.NoError(t, err)
	assert.Equal(t, id, l.ID())
	assert.Equal(t, noteID, l.NoteID())
	assert.Equal(t, by, l.SharedByUserID())
	assert.Equal(t, "token-123", l.Token())
	assert.Equal(t, "read", l.Permission())
	assert.Equal(t, &expires, l.ExpiresAt())
	assert.Equal(t, &maxUses, l.MaxUses())
	assert.Equal(t, 0, l.UsesCount())
	assert.True(t, l.IsActive())

	_, err = NewShareLink(id, noteID, by, "", "read", nil, nil, 0)
	assert.Error(t, err)

	_, err = NewShareLink(id, noteID, by, "token-123", "admin", nil, nil, 0)
	assert.ErrorIs(t, err, ErrInvalidPermission)
}

func TestShareLink_IncrementUsage(t *testing.T) {
	l, _ := NewShareLink(uuid.New(), uuid.New(), uuid.New(), "token", "read", nil, nil, 0)
	l.IncrementUsage()
	assert.Equal(t, 1, l.UsesCount())
}

func TestShareLink_Deactivate(t *testing.T) {
	l, _ := NewShareLink(uuid.New(), uuid.New(), uuid.New(), "token", "read", nil, nil, 0)
	l.Deactivate()
	assert.False(t, l.IsActive())
}
