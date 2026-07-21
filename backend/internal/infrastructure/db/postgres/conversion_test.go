package postgres

import (
	"testing"
	"time"

	"knowledge-graph/internal/domain/share"
	"knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
)

func TestToFromDomainUser(t *testing.T) {
	uid := uuid.New()
	u, err := user.NewUser(uid, "login", "email@example.com", "hash", "user", time.Now(), time.Now(), nil)
	require.NoError(t, err)

	m := fromDomainUser(u)
	assert.Equal(t, uid, m.ID)
	assert.Equal(t, "login", m.Login)
	assert.Equal(t, "email@example.com", m.Email)
	assert.Equal(t, "hash", m.PasswordHash)

	out, err := toDomainUser(m, "user")
	require.NoError(t, err)
	assert.Equal(t, u.ID(), out.ID())
	assert.Equal(t, u.Email(), out.Email())
}

func TestToFromDomainAPIKey(t *testing.T) {
	uid := uuid.New()
	key, err := user.NewAPIKey(uuid.New(), uid, "hash", "name", []string{"read"}, time.Now())
	require.NoError(t, err)

	m := fromDomainAPIKey(key)
	assert.Equal(t, key.ID(), m.ID)
	assert.Equal(t, uid, m.UserID)
	assert.Equal(t, "hash", m.KeyHash)

	out, err := toDomainAPIKey(m)
	require.NoError(t, err)
	assert.Equal(t, key.ID(), out.ID())
	assert.Equal(t, key.Name(), out.Name())
}

func TestToFromDomainNoteShare(t *testing.T) {
	sid := uuid.New()
	noteID := uuid.New()
	by := uuid.New()
	to := uuid.New()

	s, err := share.NewNoteShare(sid, noteID, by, to, "read", nil)
	require.NoError(t, err)
	m := fromDomainNoteShare(s)
	assert.Equal(t, sid, m.ID)
	assert.Equal(t, noteID, m.NoteID)
	assert.Equal(t, by, m.SharedByUserID)
	assert.Equal(t, to, m.SharedWithUserID)
	assert.Equal(t, "read", m.Permission)

	out := toDomainNoteShare(m)
	assert.Equal(t, sid, out.ID())
	assert.Equal(t, "read", out.Permission())
}

func TestToFromDomainShareLink(t *testing.T) {
	lid := uuid.New()
	noteID := uuid.New()
	by := uuid.New()

	maxUses := 10
	l, err := share.NewShareLink(lid, noteID, by, "token", "read", nil, &maxUses, 0)
	require.NoError(t, err)
	m := fromDomainShareLink(l)
	assert.Equal(t, lid, m.ID)
	assert.Equal(t, "token", m.Token)
	assert.Equal(t, 10, *m.MaxUses)

	out := toDomainShareLink(m)
	assert.Equal(t, lid, out.ID())
	assert.Equal(t, "token", out.Token())
	assert.NotNil(t, out.MaxUses())
	assert.Equal(t, 10, *out.MaxUses())
}

func TestToDomainNote_Conversion(t *testing.T) {
	uid := uuid.New()
	m := &NoteModel{
		ID:        uid,
		Title:     "Title",
		Content:   "Content",
		Type:      "star",
		Metadata:  datatypes.JSON(`{"key":"value"}`),
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now(),
	}

	n, err := toDomainNote(m)
	require.NoError(t, err)
	assert.Equal(t, uid, n.ID())
	assert.Equal(t, "Title", n.Title().String())

	models := []NoteModel{*m}
	out := toDomainNotes(models)
	assert.Len(t, out, 1)

	// Invalid metadata should be skipped by toDomainNotes
	m.Metadata = datatypes.JSON(`{invalid`)
	out2 := toDomainNotes([]NoteModel{*m})
	assert.Len(t, out2, 0)
}

func TestToDomainLink_Conversion(t *testing.T) {
	uid := uuid.New()
	source := uuid.New()
	target := uuid.New()
	m := &LinkModel{
		ID:           uid,
		SourceNoteID: source,
		TargetNoteID: target,
		LinkType:     "reference",
		Weight:       0.8,
		SourceType:   "user",
		Metadata:     datatypes.JSON(`{}`),
		CreatedAt:    time.Now().Add(-time.Hour),
		UpdatedAt:    time.Now(),
	}

	l, err := toDomainLink(m)
	require.NoError(t, err)
	assert.Equal(t, uid, l.ID())
	assert.Equal(t, source, l.SourceNoteID())
	assert.Equal(t, target, l.TargetNoteID())

	models := []LinkModel{*m}
	out := toDomainLinks(models)
	assert.Len(t, out, 1)

	// Invalid metadata should be skipped
	m.Metadata = datatypes.JSON(`{invalid`)
	out2 := toDomainLinks([]LinkModel{*m})
	assert.Len(t, out2, 0)
}

func TestToDomainAchievement(t *testing.T) {
	uid := uuid.New()
	name := "Name"
	desc := "Desc"
	icon := "🚀"
	m := &AchievementModel{
		ID:            uid,
		Code:          "code",
		NameEn:        &name,
		DescriptionEn: &desc,
		IconEmoji:     &icon,
		Points:        10,
		Hidden:        false,
		ConditionJSON: []byte(`{"type":"count_notes"}`),
		CreatedAt:     time.Now().Add(-time.Hour),
		UpdatedAt:     time.Now(),
	}

	a, err := toDomainAchievement(m)
	require.NoError(t, err)
	assert.Equal(t, uid, a.ID())
	assert.Equal(t, "code", a.Code())
	assert.Equal(t, "Name", a.Title())
}
