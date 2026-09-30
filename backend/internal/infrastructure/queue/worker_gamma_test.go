package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeGammaRunner struct {
	links []*link.Link
	err   error
	calls []uuid.UUID
}

func (f *fakeGammaRunner) GenerateForNote(_ context.Context, noteID uuid.UUID) ([]*link.Link, error) {
	f.calls = append(f.calls, noteID)
	return f.links, f.err
}

// LinkCreated events are no longer emitted by the worker — the outbox
// decorator on the link repository records them inside the write
// transaction (SYNC-1 A2), covered by the outbox integration tests.

type fakeRefreshEnqueuer struct {
	noteIDs []uuid.UUID
}

func (f *fakeRefreshEnqueuer) EnqueueRefreshRecommendations(_ context.Context, noteID uuid.UUID, _ time.Duration) error {
	f.noteIDs = append(f.noteIDs, noteID)
	return nil
}

func newGammaLink(t *testing.T, source, target uuid.UUID) *link.Link {
	t.Helper()
	lt, err := link.NewLinkType("related")
	require.NoError(t, err)
	w, err := link.NewWeight(0.9)
	require.NoError(t, err)
	md, err := link.NewMetadata(map[string]interface{}{"source": "gamma"})
	require.NoError(t, err)
	return link.NewGammaLink(source, target, lt, w, md)
}

func newWorkerNote(t *testing.T, creatorID *uuid.UUID) *note.Note {
	t.Helper()
	title, _ := note.NewTitle("Gamma source")
	content, _ := note.NewContent("content")
	md, _ := note.NewMetadata(nil)
	n := note.NewNote(title, content, note.MustType("star"), md)
	if creatorID != nil {
		n.SetCreatorID(*creatorID)
	}
	return n
}

// LINKS-1: each created gamma link enqueues a recommendations refresh for the
// source and every target — a target gained an incoming neighbour.
func TestWorker_GenerateGammaLinks_RefreshesAffected(t *testing.T) {
	sourceID := uuid.New()
	targetA := uuid.New()
	targetB := uuid.New()
	creator := uuid.New()

	runner := &fakeGammaRunner{links: []*link.Link{
		newGammaLink(t, sourceID, targetA),
		newGammaLink(t, sourceID, targetB),
	}}
	enq := &fakeRefreshEnqueuer{}

	w := NewWorker(nil, nil, nil, nil, nil, nil, runner, enq, 5*time.Second, nil, false, "")
	n := newWorkerNote(t, &creator)

	err := w.generateGammaLinks(context.Background(), n, sourceID)
	require.NoError(t, err)

	require.Len(t, runner.calls, 1)
	assert.Equal(t, sourceID, runner.calls[0])

	// targets first, then source
	require.Len(t, enq.noteIDs, 3)
	assert.Equal(t, targetA, enq.noteIDs[0])
	assert.Equal(t, targetB, enq.noteIDs[1])
	assert.Equal(t, sourceID, enq.noteIDs[2])
}

// No candidates: no refreshes, no error.
func TestWorker_GenerateGammaLinks_NoLinks(t *testing.T) {
	sourceID := uuid.New()
	runner := &fakeGammaRunner{links: nil}
	enq := &fakeRefreshEnqueuer{}

	w := NewWorker(nil, nil, nil, nil, nil, nil, runner, enq, 0, nil, false, "")
	err := w.generateGammaLinks(context.Background(), newWorkerNote(t, nil), sourceID)
	require.NoError(t, err)
	assert.Empty(t, enq.noteIDs)
}

// Generator failure propagates so asynq retries the task (idempotent).
func TestWorker_GenerateGammaLinks_GeneratorErrorPropagates(t *testing.T) {
	runner := &fakeGammaRunner{err: errors.New("db down")}
	w := NewWorker(nil, nil, nil, nil, nil, nil, runner, &fakeRefreshEnqueuer{}, 0, nil, false, "")

	err := w.generateGammaLinks(context.Background(), newWorkerNote(t, nil), uuid.New())
	require.Error(t, err)
}
