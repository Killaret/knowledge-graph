package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appquality "knowledge-graph/internal/application/quality"
	"knowledge-graph/internal/domain/note"
)

type fakeQualityLogReader struct {
	entries map[uuid.UUID]*appquality.LogEntry
	err     error
}

func (f *fakeQualityLogReader) LatestQuality(_ context.Context, noteID uuid.UUID) (*appquality.LogEntry, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.entries[noteID], nil
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

func entryFor(n *note.Note) *appquality.LogEntry {
	return &appquality.LogEntry{
		NoteID:     n.ID(),
		SourceHash: appquality.SourceHash(n.Title().String(), n.Content().String()),
		Trigger:    appquality.TriggerAuto,
		Record: appquality.Record{
			Verdict:         appquality.VerdictEnrich,
			Gates:           []string{appquality.GateStub},
			Attempt:         1,
			ComputedAt:      time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC),
			PipelineVersion: appquality.PipelineVersion,
		},
	}
}

// NOTE-QUALITY-1: --dry-run planning marks a note due when its current
// source_hash has no matching log entry; a second run over a completed
// backfill produces an empty due list.
func TestPlanRecompute_SkipsCurrentAssessments(t *testing.T) {
	fresh := newTestNote(t, "Fresh", "no assessment yet")
	current := newTestNote(t, "Current", "assessment already stored")
	stale := newTestNote(t, "Stale", "content changed after assessment")

	store := &fakeQualityLogReader{entries: map[uuid.UUID]*appquality.LogEntry{
		current.ID(): entryFor(current),
		stale.ID(): {
			NoteID:     stale.ID(),
			SourceHash: "a-different-hash",
			Record:     appquality.Record{Verdict: appquality.VerdictCreate},
		},
	}}

	due, assessed, err := planRecompute(context.Background(), []*note.Note{fresh, current, stale}, store)
	require.NoError(t, err)
	assert.ElementsMatch(t, []uuid.UUID{fresh.ID(), stale.ID()}, due)
	assert.Equal(t, 1, assessed)
}

func TestPlanRecompute_SecondRunEnqueuesNothing(t *testing.T) {
	n1 := newTestNote(t, "One", "content one")
	n2 := newTestNote(t, "Two", "content two")

	store := &fakeQualityLogReader{entries: map[uuid.UUID]*appquality.LogEntry{
		n1.ID(): entryFor(n1),
		n2.ID(): entryFor(n2),
	}}

	due, assessed, err := planRecompute(context.Background(), []*note.Note{n1, n2}, store)
	require.NoError(t, err)
	assert.Empty(t, due, "second run must enqueue nothing")
	assert.Equal(t, 2, assessed)
}

func TestPlanRecompute_StoreErrorPropagates(t *testing.T) {
	n := newTestNote(t, "Err", "content")
	bad := &fakeQualityLogReader{err: errors.New("mongo down")}
	_, _, err := planRecompute(context.Background(), []*note.Note{n}, bad)
	require.Error(t, err)
	assert.Contains(t, err.Error(), n.ID().String())
}

// --export writes JSONL rows with ids, hashes, gates and signals — never
// titles or note text.
func TestExportSnapshot_WritesTextFreeRows(t *testing.T) {
	withEntry := newTestNote(t, "Secret title", "secret body")
	noEntry := newTestNote(t, "Other", "other body")

	store := &fakeQualityLogReader{entries: map[uuid.UUID]*appquality.LogEntry{
		withEntry.ID(): entryFor(withEntry),
	}}

	path := filepath.Join(t.TempDir(), "quality.jsonl")
	written, err := exportSnapshot(context.Background(), []*note.Note{withEntry, noEntry}, store, path)
	require.NoError(t, err)
	assert.Equal(t, 1, written)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "Secret title")
	assert.NotContains(t, string(data), "secret body")

	var row exportRow
	require.NoError(t, json.Unmarshal(data, &row))
	assert.Equal(t, withEntry.ID(), row.NoteID)
	assert.Equal(t, appquality.VerdictEnrich, row.Verdict)
	assert.Equal(t, []string{appquality.GateStub}, row.Gates)
}
