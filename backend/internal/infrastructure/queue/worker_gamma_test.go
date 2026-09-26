package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"
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

type fakeLinkPublisher struct {
	events [][3]string
}

func (f *fakeLinkPublisher) PublishLinkCreated(_ context.Context, source, target, user string) error {
	f.events = append(f.events, [3]string{source, target, user})
	return nil
}

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

// LINKS-1: each created gamma link produces one LinkCreated event, and a
// recommendations refresh is enqueued for the source and every target.
func TestWorker_GenerateGammaLinks_PublishesAndRefreshes(t *testing.T) {
	sourceID := uuid.New()
	targetA := uuid.New()
	targetB := uuid.New()
	creator := uuid.New()

	runner := &fakeGammaRunner{links: []*link.Link{
		newGammaLink(t, sourceID, targetA),
		newGammaLink(t, sourceID, targetB),
	}}
	pub := &fakeLinkPublisher{}
	enq := &fakeRefreshEnqueuer{}

	w := NewWorker(nil, nil, nil, nil, nil, nil, runner, pub, enq, 5*time.Second, nil, false, "")
	n := newWorkerNote(t, &creator)

	err := w.generateGammaLinks(context.Background(), n, sourceID)
	require.NoError(t, err)

	require.Len(t, runner.calls, 1)
	assert.Equal(t, sourceID, runner.calls[0])

	require.Len(t, pub.events, 2)
	assert.Equal(t, [3]string{sourceID.String(), targetA.String(), creator.String()}, pub.events[0])
	assert.Equal(t, [3]string{sourceID.String(), targetB.String(), creator.String()}, pub.events[1])

	// targets first, then source
	require.Len(t, enq.noteIDs, 3)
	assert.Equal(t, targetA, enq.noteIDs[0])
	assert.Equal(t, targetB, enq.noteIDs[1])
	assert.Equal(t, sourceID, enq.noteIDs[2])
}

// No candidates: no events, no refreshes, no error.
func TestWorker_GenerateGammaLinks_NoLinks(t *testing.T) {
	sourceID := uuid.New()
	runner := &fakeGammaRunner{links: nil}
	pub := &fakeLinkPublisher{}
	enq := &fakeRefreshEnqueuer{}

	w := NewWorker(nil, nil, nil, nil, nil, nil, runner, pub, enq, 0, nil, false, "")
	err := w.generateGammaLinks(context.Background(), newWorkerNote(t, nil), sourceID)
	require.NoError(t, err)
	assert.Empty(t, pub.events)
	assert.Empty(t, enq.noteIDs)
}

// Generator failure propagates so asynq retries the task (idempotent).
func TestWorker_GenerateGammaLinks_GeneratorErrorPropagates(t *testing.T) {
	runner := &fakeGammaRunner{err: errors.New("db down")}
	w := NewWorker(nil, nil, nil, nil, nil, nil, runner, &fakeLinkPublisher{}, &fakeRefreshEnqueuer{}, 0, nil, false, "")

	err := w.generateGammaLinks(context.Background(), newWorkerNote(t, nil), uuid.New())
	require.Error(t, err)
}
