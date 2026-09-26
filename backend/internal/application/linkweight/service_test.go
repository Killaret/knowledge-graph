package linkweight

import (
	"context"
	"testing"

	appevents "knowledge-graph/internal/application/events"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// fakeLinkRepo is a minimal in-memory link.Repository for unit tests.
type fakeLinkRepo struct {
	links   []*link.Link
	updated []*link.Link
}

func (r *fakeLinkRepo) Save(ctx context.Context, l *link.Link) error { return nil }

func (r *fakeLinkRepo) Update(ctx context.Context, l *link.Link) error {
	r.updated = append(r.updated, l)
	return nil
}

func (r *fakeLinkRepo) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	return nil, nil
}

func (r *fakeLinkRepo) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	var out []*link.Link
	for _, l := range r.links {
		if l.SourceNoteID() == sourceID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r *fakeLinkRepo) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	var out []*link.Link
	for _, l := range r.links {
		if l.TargetNoteID() == targetID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r *fakeLinkRepo) FindByPair(ctx context.Context, sourceID, targetID uuid.UUID) ([]*link.Link, error) {
	return nil, nil
}

func (r *fakeLinkRepo) SaveUserLink(ctx context.Context, l *link.Link) (*link.Link, bool, error) {
	return l, true, nil
}

func (r *fakeLinkRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

func (r *fakeLinkRepo) DeleteAndSuppress(ctx context.Context, l *link.Link, s *link.Suppression) error {
	return nil
}

func (r *fakeLinkRepo) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error { return nil }

func (r *fakeLinkRepo) FindAll(ctx context.Context) ([]*link.Link, error) { return r.links, nil }

func (r *fakeLinkRepo) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	return r.links, int64(len(r.links)), nil
}

// fakeNoteRepo returns prebuilt notes by id.
type fakeNoteRepo struct {
	notes map[uuid.UUID]*note.Note
}

func (r *fakeNoteRepo) Save(ctx context.Context, n *note.Note) error { return nil }

func (r *fakeNoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	return r.notes[id], nil
}

func (r *fakeNoteRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

func (r *fakeNoteRepo) DeleteBatch(ctx context.Context, ids []uuid.UUID) error { return nil }

func (r *fakeNoteRepo) Restore(ctx context.Context, id uuid.UUID) error { return nil }

func (r *fakeNoteRepo) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

func (r *fakeNoteRepo) Search(ctx context.Context, userID uuid.UUID, query string, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

func (r *fakeNoteRepo) FindAll(ctx context.Context) ([]*note.Note, error) { return nil, nil }

func (r *fakeNoteRepo) FindAllPaginated(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*note.Note, int64, error) {
	return nil, 0, nil
}

// fakeSimilarity returns a fixed similarity score.
type fakeSimilarity struct{}

func (f fakeSimilarity) Similarity(ctx context.Context, textA, textB string) (float64, error) {
	return 0.75, nil
}

// fakePublisher records published link events.
type fakePublisher struct {
	linkUpdated [][3]string
}

func (p *fakePublisher) PublishNoteCreated(ctx context.Context, noteID, userID string) error {
	return nil
}
func (p *fakePublisher) PublishNoteUpdated(ctx context.Context, noteID, userID string) error {
	return nil
}
func (p *fakePublisher) PublishNoteDeleted(ctx context.Context, noteID, userID string) error {
	return nil
}
func (p *fakePublisher) PublishLinkCreated(ctx context.Context, sourceNoteID, targetNoteID, userID string) error {
	return nil
}
func (p *fakePublisher) PublishLinkUpdated(ctx context.Context, sourceNoteID, targetNoteID, userID string) error {
	p.linkUpdated = append(p.linkUpdated, [3]string{sourceNoteID, targetNoteID, userID})
	return nil
}
func (p *fakePublisher) PublishLinkDeleted(ctx context.Context, sourceNoteID, targetNoteID, userID string) error {
	return nil
}

var (
	_ link.Repository     = (*fakeLinkRepo)(nil)
	_ note.Repository     = (*fakeNoteRepo)(nil)
	_ SimilarityClient    = fakeSimilarity{}
	_ appevents.Publisher = (*fakePublisher)(nil)
)

func mustNote(t *testing.T, creator uuid.UUID) *note.Note {
	t.Helper()
	title, err := note.NewTitle("T")
	require.NoError(t, err)
	content, err := note.NewContent("body")
	require.NoError(t, err)
	meta, err := note.NewMetadata(map[string]interface{}{})
	require.NoError(t, err)
	return note.NewNoteWithCreator(title, content, note.MustType("asteroid"), meta, creator)
}

// SYNC-1: a weight recalculation changes graph data — every updated link must
// publish LinkUpdated with the owner's id, or open graphs keep stale weights
// until TTL.
func TestRecalculateForNote_PublishesLinkUpdated(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	src := mustNote(t, userID)
	dst := mustNote(t, userID)

	w, err := link.NewWeight(0.5)
	require.NoError(t, err)
	ltype, err := link.NewLinkType("related")
	require.NoError(t, err)
	lmeta, err := link.NewMetadata(map[string]interface{}{})
	require.NoError(t, err)
	l := link.NewLink(src.ID(), dst.ID(), ltype, w, lmeta)

	linkRepo := &fakeLinkRepo{links: []*link.Link{l}}
	noteRepo := &fakeNoteRepo{notes: map[uuid.UUID]*note.Note{src.ID(): src, dst.ID(): dst}}
	rec := NewRecalculator(linkRepo, noteRepo, fakeSimilarity{})
	pub := &fakePublisher{}
	rec.SetEventPublisher(pub)

	require.NoError(t, rec.RecalculateForNote(ctx, src.ID()))
	require.Len(t, linkRepo.updated, 1)
	require.Len(t, pub.linkUpdated, 1)
	require.Equal(t, [3]string{src.ID().String(), dst.ID().String(), userID.String()}, pub.linkUpdated[0])
}
