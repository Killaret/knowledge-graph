package recommendation

import (
	"context"
	"testing"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockEmbeddingRepoForGamma struct{ mock.Mock }

func (m *mockEmbeddingRepoForGamma) FindSimilarNotes(ctx context.Context, noteID uuid.UUID, limit int) ([]SimilarNote, error) {
	args := m.Called(ctx, noteID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]SimilarNote), args.Error(1)
}

func (m *mockEmbeddingRepoForGamma) FindSimilarNotesBatch(ctx context.Context, noteIDs []uuid.UUID, limit int) (map[uuid.UUID][]SimilarNote, error) {
	args := m.Called(ctx, noteIDs, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID][]SimilarNote), args.Error(1)
}

type mockBatchLinkRepoForGamma struct{ mock.Mock }

func (m *mockBatchLinkRepoForGamma) Save(ctx context.Context, l *link.Link) error {
	return m.Called(ctx, l).Error(0)
}

func (m *mockBatchLinkRepoForGamma) Update(ctx context.Context, l *link.Link) error {
	return m.Called(ctx, l).Error(0)
}

func (m *mockBatchLinkRepoForGamma) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*link.Link), args.Error(1)
}

func (m *mockBatchLinkRepoForGamma) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	args := m.Called(ctx, sourceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*link.Link), args.Error(1)
}

func (m *mockBatchLinkRepoForGamma) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	args := m.Called(ctx, targetID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*link.Link), args.Error(1)
}

func (m *mockBatchLinkRepoForGamma) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockBatchLinkRepoForGamma) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	return m.Called(ctx, sourceID).Error(0)
}

func (m *mockBatchLinkRepoForGamma) FindAll(ctx context.Context) ([]*link.Link, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*link.Link), args.Error(1)
}

func (m *mockBatchLinkRepoForGamma) FindAllPaginated(ctx context.Context, limit, offset int) ([]*link.Link, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*link.Link), args.Get(1).(int64), args.Error(2)
}

func (m *mockBatchLinkRepoForGamma) FindBySourceIDs(ctx context.Context, sourceIDs []uuid.UUID) (map[uuid.UUID][]*link.Link, error) {
	args := m.Called(ctx, sourceIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[uuid.UUID][]*link.Link), args.Error(1)
}

func newNote(t *testing.T, title string) *note.Note {
	titleV, err := note.NewTitle(title)
	assert.NoError(t, err)
	content, err := note.NewContent("content")
	assert.NoError(t, err)
	metadata, err := note.NewMetadata(nil)
	assert.NoError(t, err)
	return note.NewNote(titleV, content, "star", metadata)
}

func TestGammaLinkGenerator_RespectsMaxOutDegree(t *testing.T) {
	embRepo := new(mockEmbeddingRepoForGamma)
	linkRepo := new(mockBatchLinkRepoForGamma)
	gen := NewGammaLinkGenerator(embRepo, linkRepo, 2)

	sourceID := uuid.New()
	target1 := uuid.New()
	target2 := uuid.New()
	target3 := uuid.New()

	embRepo.On("FindSimilarNotes", context.Background(), sourceID, 2).Return([]SimilarNote{
		{NoteID: target1, Score: 0.9},
		{NoteID: target2, Score: 0.8},
		{NoteID: target3, Score: 0.7},
	}, nil)
	linkRepo.On("FindBySource", context.Background(), sourceID).Return([]*link.Link{}, nil)

	var saved []*link.Link
	linkRepo.On("Save", context.Background(), mock.AnythingOfType("*link.Link")).Run(func(args mock.Arguments) {
		l := args.Get(1).(*link.Link)
		saved = append(saved, l)
	}).Return(nil)

	err := gen.GenerateForNote(context.Background(), sourceID)
	assert.NoError(t, err)

	assert.Len(t, saved, 2)
	assert.Equal(t, target1, saved[0].TargetNoteID())
	assert.Equal(t, target2, saved[1].TargetNoteID())
	assert.Equal(t, "related", saved[0].LinkType().String())
	assert.Equal(t, "gamma", saved[0].SourceType().String())
	assert.InDelta(t, 0.9, saved[0].Weight().Value(), 0.0001)
	embRepo.AssertExpectations(t)
	linkRepo.AssertExpectations(t)
}

func TestGammaLinkGenerator_SkipsSelfLoopsAndExistingLinks(t *testing.T) {
	embRepo := new(mockEmbeddingRepoForGamma)
	linkRepo := new(mockBatchLinkRepoForGamma)
	gen := NewGammaLinkGenerator(embRepo, linkRepo, 3)

	sourceID := uuid.New()
	target1 := uuid.New()
	target2 := sourceID
	target3 := uuid.New()

	linkType, err := link.NewLinkType("related")
	assert.NoError(t, err)
	weight, err := link.NewWeight(0.6)
	assert.NoError(t, err)
	metadata, err := link.NewMetadata(map[string]interface{}{"source": "gamma"})
	assert.NoError(t, err)
	existing := link.NewGammaLink(sourceID, target1, linkType, weight, metadata)

	embRepo.On("FindSimilarNotes", context.Background(), sourceID, 3).Return([]SimilarNote{
		{NoteID: target1, Score: 0.9},
		{NoteID: target2, Score: 0.85},
		{NoteID: target3, Score: 0.7},
	}, nil)
	linkRepo.On("FindBySource", context.Background(), sourceID).Return([]*link.Link{existing}, nil)

	var saved []*link.Link
	linkRepo.On("Save", context.Background(), mock.AnythingOfType("*link.Link")).Run(func(args mock.Arguments) {
		l := args.Get(1).(*link.Link)
		saved = append(saved, l)
	}).Return(nil)

	err = gen.GenerateForNote(context.Background(), sourceID)
	assert.NoError(t, err)

	// target1 already exists, target2 is a self-loop, only target3 should be created.
	assert.Len(t, saved, 1)
	assert.Equal(t, target3, saved[0].TargetNoteID())
}

func TestGammaLinkGenerator_BatchCreatesAtMostMaxOutDegreePerNote(t *testing.T) {
	embRepo := new(mockEmbeddingRepoForGamma)
	linkRepo := new(mockBatchLinkRepoForGamma)
	gen := NewGammaLinkGenerator(embRepo, linkRepo, 1)

	noteA := uuid.New()
	noteB := uuid.New()
	targetA := uuid.New()
	targetB := uuid.New()
	extraA := uuid.New()

	embRepo.On("FindSimilarNotesBatch", context.Background(), []uuid.UUID{noteA, noteB}, 1).Return(map[uuid.UUID][]SimilarNote{
		noteA: {{NoteID: targetA, Score: 0.9}, {NoteID: extraA, Score: 0.8}},
		noteB: {{NoteID: targetB, Score: 0.85}},
	}, nil)
	linkRepo.On("FindBySourceIDs", context.Background(), []uuid.UUID{noteA, noteB}).Return(map[uuid.UUID][]*link.Link{}, nil)

	var saved []*link.Link
	linkRepo.On("Save", context.Background(), mock.AnythingOfType("*link.Link")).Run(func(args mock.Arguments) {
		l := args.Get(1).(*link.Link)
		saved = append(saved, l)
	}).Return(nil)

	err := gen.GenerateForNotes(context.Background(), []uuid.UUID{noteA, noteB})
	assert.NoError(t, err)

	assert.Len(t, saved, 2)
	assert.Equal(t, targetA, saved[0].TargetNoteID())
	assert.Equal(t, targetB, saved[1].TargetNoteID())
}
