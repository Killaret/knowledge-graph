package mongo

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDraftModelStructure(t *testing.T) {
	noteID := uuid.New()
	userID := uuid.New()

	model := &draftModel{
		ID:        uuid.New(),
		NoteID:    noteID,
		UserID:    userID,
		Content:   "Test content",
		Title:     "Test title",
		State:     "active",
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	if model.NoteID != noteID {
		t.Errorf("draftModel.NoteID = %v, want %v", model.NoteID, noteID)
	}

	if model.UserID != userID {
		t.Errorf("draftModel.UserID = %v, want %v", model.UserID, userID)
	}

	if model.Content != "Test content" {
		t.Errorf("draftModel.Content = %v, want %v", model.Content, "Test content")
	}

	if model.Title != "Test title" {
		t.Errorf("draftModel.Title = %v, want %v", model.Title, "Test title")
	}

	if model.State != "active" {
		t.Errorf("draftModel.State = %v, want %v", model.State, "active")
	}
}
