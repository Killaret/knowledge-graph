package note

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewNote(t *testing.T) {
	title, _ := NewTitle("Test")
	content, _ := NewContent("Content")
	metadata, _ := NewMetadata(map[string]interface{}{})

	note := NewNote(title, content, metadata)

	if note.ID() == uuid.Nil {
		t.Error("ID should not be nil")
	}
	if note.Title().String() != "Test" {
		t.Error("title mismatch")
	}
	if note.Content().String() != "Content" {
		t.Error("content mismatch")
	}
	if note.Metadata().Value() == nil {
		t.Error("metadata should not be nil")
	}
}

func TestNoteUpdateTitle(t *testing.T) {
	title, _ := NewTitle("Old")
	content, _ := NewContent("Content")
	metadata, _ := NewMetadata(map[string]interface{}{})
	note := NewNote(title, content, metadata)

	newTitle, _ := NewTitle("New")
	err := note.UpdateTitle(newTitle)
	if err != nil {
		t.Errorf("UpdateTitle failed: %v", err)
	}
	if note.Title().String() != "New" {
		t.Error("title not updated")
	}
	if note.UpdatedAt().Before(time.Now().Add(-time.Second)) {
		t.Error("UpdatedAt not updated")
	}
}
