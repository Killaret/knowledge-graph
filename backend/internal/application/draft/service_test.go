package draft

import (
	"context"
	"errors"
	"testing"
	"time"

	noteDomain "knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDraftRepository is a mock implementation of DraftRepository
type MockDraftRepository struct {
	mock.Mock
}

func (m *MockDraftRepository) Save(ctx context.Context, draft *noteDomain.Draft) error {
	args := m.Called(ctx, draft)
	return args.Error(0)
}

func (m *MockDraftRepository) FindByNoteAndUser(ctx context.Context, noteID, userID uuid.UUID) (*noteDomain.Draft, error) {
	args := m.Called(ctx, noteID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*noteDomain.Draft), args.Error(1)
}

func (m *MockDraftRepository) FindActiveByUser(ctx context.Context, userID uuid.UUID) ([]*noteDomain.Draft, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*noteDomain.Draft), args.Error(1)
}

func (m *MockDraftRepository) FindByID(ctx context.Context, id uuid.UUID) (*noteDomain.Draft, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*noteDomain.Draft), args.Error(1)
}

func (m *MockDraftRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDraftRepository) DeleteExpired(ctx context.Context, before time.Time) (int, error) {
	args := m.Called(ctx, before)
	return args.Int(0), args.Error(1)
}

func (m *MockDraftRepository) Update(ctx context.Context, draft *noteDomain.Draft) error {
	args := m.Called(ctx, draft)
	return args.Error(0)
}

// MockNoteRepository is a mock implementation of note.Repository
type MockNoteRepository struct {
	mock.Mock
}

func (m *MockNoteRepository) Save(ctx context.Context, n *noteDomain.Note) error {
	args := m.Called(ctx, n)
	return args.Error(0)
}

func (m *MockNoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*noteDomain.Note, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*noteDomain.Note), args.Error(1)
}

func (m *MockNoteRepository) FindAll(ctx context.Context) ([]*noteDomain.Note, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*noteDomain.Note), args.Error(1)
}

func (m *MockNoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNoteRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockNoteRepository) Restore(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNoteRepository) List(ctx context.Context, limit, offset int) ([]*noteDomain.Note, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, int64(args.Int(1)), args.Error(2)
	}
	return args.Get(0).([]*noteDomain.Note), int64(args.Int(1)), args.Error(2)
}

func (m *MockNoteRepository) Search(ctx context.Context, query string, limit, offset int) ([]*noteDomain.Note, int64, error) {
	args := m.Called(ctx, query, limit, offset)
	if args.Get(0) == nil {
		return nil, int64(args.Int(1)), args.Error(2)
	}
	return args.Get(0).([]*noteDomain.Note), int64(args.Int(1)), args.Error(2)
}

func (m *MockNoteRepository) GetNotesByType(ctx context.Context, noteType string) ([]*noteDomain.Note, error) {
	args := m.Called(ctx, noteType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*noteDomain.Note), args.Error(1)
}

func (m *MockNoteRepository) GetNotesByCreator(ctx context.Context, creatorID uuid.UUID) ([]*noteDomain.Note, error) {
	args := m.Called(ctx, creatorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*noteDomain.Note), args.Error(1)
}

func (m *MockNoteRepository) FindAllPaginated(ctx context.Context, limit, offset int) ([]*noteDomain.Note, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, int64(args.Int(1)), args.Error(2)
	}
	return args.Get(0).([]*noteDomain.Note), int64(args.Int(1)), args.Error(2)
}

func TestNewService(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	assert.NotNil(t, service)
	assert.Equal(t, mockDraftRepo, service.repo)
	assert.Equal(t, mockNoteRepo, service.noteRepo)
	assert.Equal(t, "", service.syncEndpoint)
	assert.Equal(t, 3, service.maxRetries)
}

func TestSaveDraft_NewDraft(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	noteID := uuid.New()
	userID := uuid.New()
	content := "Test content"
	title := "Test title"

	// No existing draft
	mockDraftRepo.On("FindByNoteAndUser", ctx, noteID, userID).Return(nil, nil)
	// Save new draft
	mockDraftRepo.On("Save", ctx, mock.AnythingOfType("*note.Draft")).Return(nil)

	draft, err := service.SaveDraft(ctx, noteID, userID, content, title)

	assert.NoError(t, err)
	assert.NotNil(t, draft)
	assert.Equal(t, noteID, draft.NoteID())
	assert.Equal(t, userID, draft.UserID())
	assert.Equal(t, content, draft.Content())
	assert.Equal(t, title, draft.Title())
	assert.Equal(t, noteDomain.DraftStateActive, draft.State())

	mockDraftRepo.AssertExpectations(t)
}

func TestSaveDraft_UpdateExisting(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	noteID := uuid.New()
	userID := uuid.New()
	newContent := "Updated content"
	newTitle := "Updated title"

	// Existing draft
	existingDraft := noteDomain.NewDraft(noteID, userID, "Old content", "Old title")
	mockDraftRepo.On("FindByNoteAndUser", ctx, noteID, userID).Return(existingDraft, nil)
	// Update draft
	mockDraftRepo.On("Update", ctx, existingDraft).Return(nil)

	draft, err := service.SaveDraft(ctx, noteID, userID, newContent, newTitle)

	assert.NoError(t, err)
	assert.NotNil(t, draft)
	assert.Equal(t, newContent, draft.Content())
	assert.Equal(t, newTitle, draft.Title())

	mockDraftRepo.AssertExpectations(t)
}

func TestSaveDraft_ErrorCheckingExisting(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	noteID := uuid.New()
	userID := uuid.New()

	mockDraftRepo.On("FindByNoteAndUser", ctx, noteID, userID).Return(nil, errors.New("database error"))

	_, err := service.SaveDraft(ctx, noteID, userID, "content", "title")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check existing draft")

	mockDraftRepo.AssertExpectations(t)
}

func TestSaveDraft_ErrorSavingNew(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	noteID := uuid.New()
	userID := uuid.New()

	mockDraftRepo.On("FindByNoteAndUser", ctx, noteID, userID).Return(nil, nil)
	mockDraftRepo.On("Save", ctx, mock.AnythingOfType("*note.Draft")).Return(errors.New("save error"))

	_, err := service.SaveDraft(ctx, noteID, userID, "content", "title")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save draft")

	mockDraftRepo.AssertExpectations(t)
}

func TestSyncDraft_Success(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	draftID := uuid.New()

	draft := noteDomain.NewDraft(uuid.New(), uuid.New(), "content", "title")
	// Set ID to match the expected draft ID
	draft = noteDomain.ReconstructDraft(draftID, draft.NoteID(), draft.UserID(), draft.Content(), draft.Title(), noteDomain.DraftStateActive, draft.CreatedAt(), draft.UpdatedAt())

	// Existing note owned by the draft author.
	title, _ := noteDomain.NewTitle("Old title")
	content, _ := noteDomain.NewContent("Old content")
	metadata, _ := noteDomain.NewMetadata(nil)
	userID := draft.UserID()
	existingNote := noteDomain.ReconstructNoteWithCreator(
		draft.NoteID(), title, content, "star", metadata, &userID,
		time.Now().Add(-time.Hour), time.Now().Add(-time.Hour),
	)

	mockDraftRepo.On("FindByID", ctx, draftID).Return(draft, nil)
	mockDraftRepo.On("Update", ctx, draft).Return(nil).Times(2) // StartPublishing and MarkAsPublished
	mockDraftRepo.On("DeleteByID", ctx, draftID).Return(nil)
	mockNoteRepo.On("FindByID", ctx, draft.NoteID()).Return(existingNote, nil)
	mockNoteRepo.On("Save", ctx, mock.AnythingOfType("*note.Note")).Return(nil)

	err := service.SyncDraft(ctx, draftID)

	assert.NoError(t, err)
	assert.Equal(t, noteDomain.DraftStatePublished, draft.State())

	mockDraftRepo.AssertExpectations(t)
}

func TestSyncDraft_NoteNotFound(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")
	service.SetMaxRetries(1)

	ctx := context.Background()
	draftID := uuid.New()

	draft := noteDomain.NewDraft(uuid.New(), uuid.New(), "content", "title")
	draft = noteDomain.ReconstructDraft(draftID, draft.NoteID(), draft.UserID(), draft.Content(), draft.Title(), noteDomain.DraftStateActive, draft.CreatedAt(), draft.UpdatedAt())

	mockDraftRepo.On("FindByID", ctx, draftID).Return(draft, nil)
	mockDraftRepo.On("Update", ctx, draft).Return(nil).Twice()
	mockNoteRepo.On("FindByID", ctx, draft.NoteID()).Return(nil, nil)

	err := service.SyncDraft(ctx, draftID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "note not found")
}

func TestSyncDraft_DifferentCreator(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")
	service.SetMaxRetries(1)

	ctx := context.Background()
	draftID := uuid.New()

	draft := noteDomain.NewDraft(uuid.New(), uuid.New(), "content", "title")
	draft = noteDomain.ReconstructDraft(draftID, draft.NoteID(), draft.UserID(), draft.Content(), draft.Title(), noteDomain.DraftStateActive, draft.CreatedAt(), draft.UpdatedAt())

	title, _ := noteDomain.NewTitle("Old title")
	content, _ := noteDomain.NewContent("Old content")
	metadata, _ := noteDomain.NewMetadata(nil)
	differentCreator := uuid.New()
	existingNote := noteDomain.ReconstructNoteWithCreator(
		draft.NoteID(), title, content, "star", metadata, &differentCreator,
		time.Now().Add(-time.Hour), time.Now().Add(-time.Hour),
	)

	mockDraftRepo.On("FindByID", ctx, draftID).Return(draft, nil)
	mockDraftRepo.On("Update", ctx, draft).Return(nil).Twice()
	mockNoteRepo.On("FindByID", ctx, draft.NoteID()).Return(existingNote, nil)

	err := service.SyncDraft(ctx, draftID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "owner")
}

func TestSyncDraft_RemoteEndpoint(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "http://sync")
	service.SetMaxRetries(1)

	ctx := context.Background()
	draftID := uuid.New()

	draft := noteDomain.NewDraft(uuid.New(), uuid.New(), "content", "title")
	draft = noteDomain.ReconstructDraft(draftID, draft.NoteID(), draft.UserID(), draft.Content(), draft.Title(), noteDomain.DraftStateActive, draft.CreatedAt(), draft.UpdatedAt())

	mockDraftRepo.On("FindByID", ctx, draftID).Return(draft, nil)
	mockDraftRepo.On("Update", ctx, draft).Return(nil).Twice()

	err := service.SyncDraft(ctx, draftID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remote sync endpoint not implemented")
}

func TestSyncDraft_NotFound(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	draftID := uuid.New()

	mockDraftRepo.On("FindByID", ctx, draftID).Return(nil, nil)

	err := service.SyncDraft(ctx, draftID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "draft not found")

	mockDraftRepo.AssertExpectations(t)
}

func TestResolveConflict_Success(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	draftID := uuid.New()

	draft := noteDomain.NewDraft(uuid.New(), uuid.New(), "content", "title")
	// Manually set to conflict state
	draft = noteDomain.ReconstructDraft(draftID, draft.NoteID(), draft.UserID(), draft.Content(), draft.Title(), noteDomain.DraftStateConflict, draft.CreatedAt(), draft.UpdatedAt())

	mockDraftRepo.On("FindByID", ctx, draftID).Return(draft, nil)
	mockDraftRepo.On("Update", ctx, draft).Return(nil)

	err := service.ResolveConflict(ctx, draftID)

	assert.NoError(t, err)
	assert.Equal(t, noteDomain.DraftStateActive, draft.State())

	mockDraftRepo.AssertExpectations(t)
}

func TestResolveConflict_NotFound(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	draftID := uuid.New()

	mockDraftRepo.On("FindByID", ctx, draftID).Return(nil, nil)

	err := service.ResolveConflict(ctx, draftID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "draft not found")

	mockDraftRepo.AssertExpectations(t)
}

func TestGetLatestDraft(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	noteID := uuid.New()
	userID := uuid.New()

	expectedDraft := noteDomain.NewDraft(noteID, userID, "content", "title")

	mockDraftRepo.On("FindByNoteAndUser", ctx, noteID, userID).Return(expectedDraft, nil)

	draft, err := service.GetLatestDraft(ctx, noteID, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedDraft, draft)

	mockDraftRepo.AssertExpectations(t)
}

func TestGetActiveDrafts(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	userID := uuid.New()

	expectedDrafts := []*noteDomain.Draft{
		noteDomain.NewDraft(uuid.New(), userID, "content1", "title1"),
		noteDomain.NewDraft(uuid.New(), userID, "content2", "title2"),
	}

	mockDraftRepo.On("FindActiveByUser", ctx, userID).Return(expectedDrafts, nil)

	drafts, err := service.GetActiveDrafts(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedDrafts, drafts)
	assert.Len(t, drafts, 2)

	mockDraftRepo.AssertExpectations(t)
}

func TestDeleteDraft(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	draftID := uuid.New()

	mockDraftRepo.On("DeleteByID", ctx, draftID).Return(nil)

	err := service.DeleteDraft(ctx, draftID)

	assert.NoError(t, err)

	mockDraftRepo.AssertExpectations(t)
}

func TestDeleteDraft_Error(t *testing.T) {
	mockDraftRepo := new(MockDraftRepository)
	mockNoteRepo := new(MockNoteRepository)

	service := NewService(mockDraftRepo, mockNoteRepo, "")

	ctx := context.Background()
	draftID := uuid.New()

	mockDraftRepo.On("DeleteByID", ctx, draftID).Return(errors.New("delete error"))

	err := service.DeleteDraft(ctx, draftID)

	assert.Error(t, err)

	mockDraftRepo.AssertExpectations(t)
}
