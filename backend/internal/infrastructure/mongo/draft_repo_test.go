package mongo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDraftModelStructure(t *testing.T) {
	noteID := uuid.New()
	userID := uuid.New()
	objID := primitive.NewObjectID()

	model := &DraftModel{
		ID:        objID,
		NoteID:    noteID.String(),
		UserID:    userID.String(),
		Content:   "Test content",
		Title:     "Test title",
		State:     "active",
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	if model.NoteID != noteID.String() {
		t.Errorf("DraftModel.NoteID = %v, want %v", model.NoteID, noteID.String())
	}

	if model.UserID != userID.String() {
		t.Errorf("DraftModel.UserID = %v, want %v", model.UserID, userID.String())
	}

	if model.Content != "Test content" {
		t.Errorf("DraftModel.Content = %v, want %v", model.Content, "Test content")
	}

	if model.Title != "Test title" {
		t.Errorf("DraftModel.Title = %v, want %v", model.Title, "Test title")
	}

	if model.State != "active" {
		t.Errorf("DraftModel.State = %v, want %v", model.State, "active")
	}
}
