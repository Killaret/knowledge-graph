package main

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knowledge-graph/internal/domain/note"
)

type fakeArtifactHashReader struct {
	hashes map[uuid.UUID]string
}

func (f *fakeArtifactHashReader) FindCurrentSourceHash(_ context.Context, noteID uuid.UUID, _ string) (string, bool, error) {
	h, ok := f.hashes[noteID]
	return h, ok, nil
}

func newTestNote(t *testing.T, title, content string) *note.Note {
	t.Helper()
	ti, err := note.NewTitle(title)
	require.NoError(t, err)
	c, err := note.NewContent(content)
	require.NoError(t, err)
	m, err := note.NewMetadata(nil)
	require.NoError(t, err)
	return note.NewNote(ti, c, note.MustType("star"), m)
}

// NLP-4 criterion 6: --dry-run planning splits notes into create vs
// skip-by-hash; a second run over artifacts the worker already wrote
// produces an empty create list.
func TestPlanRecompute_SkipsCurrentArtifacts(t *testing.T) {
	fresh := newTestNote(t, "Fresh", "no artifact yet")
	current := newTestNote(t, "Current", "artifact already stored")
	stale := newTestNote(t, "Stale", "content changed after artifact")

	store := &fakeArtifactHashReader{hashes: map[uuid.UUID]string{
		current.ID(): sourceHash(current.Title().String(), current.Content().String()),
		stale.ID():   "a-different-hash",
	}}

	toCreate, skipped, err := planRecompute(context.Background(), []*note.Note{fresh, current, stale}, store)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{fresh.ID().String(), stale.ID().String()}, toCreate)
	assert.Equal(t, 1, skipped)
}

func TestPlanRecompute_SecondRunCreatesNothing(t *testing.T) {
	n1 := newTestNote(t, "One", "content one")
	n2 := newTestNote(t, "Two", "content two")

	// Simulate the state after a completed first run: the worker wrote a
	// current artifact per note with the same hash planRecompute computes.
	store := &fakeArtifactHashReader{hashes: map[uuid.UUID]string{
		n1.ID(): sourceHash(n1.Title().String(), n1.Content().String()),
		n2.ID(): sourceHash(n2.Title().String(), n2.Content().String()),
	}}

	toCreate, skipped, err := planRecompute(context.Background(), []*note.Note{n1, n2}, store)
	require.NoError(t, err)
	assert.Empty(t, toCreate, "second run must create nothing")
	assert.Equal(t, 2, skipped)
}

func TestPlanRecompute_StoreErrorPropagates(t *testing.T) {
	n := newTestNote(t, "Err", "content")
	bad := &failingReader{err: errors.New("mongo down")}
	_, _, err := planRecompute(context.Background(), []*note.Note{n}, bad)
	require.Error(t, err)
	assert.Contains(t, err.Error(), n.ID().String())
}

type failingReader struct{ err error }

func (f *failingReader) FindCurrentSourceHash(context.Context, uuid.UUID, string) (string, bool, error) {
	return "", false, f.err
}
