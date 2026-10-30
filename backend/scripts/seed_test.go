package main

import (
	"testing"
)

func TestGenerateNotes(t *testing.T) {
	notes := generateNotes(20)

	if len(notes) != 20 {
		t.Fatalf("expected 20 notes, got %d", len(notes))
	}

	seenTypes := make(map[string]bool)
	for _, note := range notes {
		if note.ID == "" {
			t.Error("note ID should not be empty")
		}
		if note.Title == "" {
			t.Error("note title should not be empty")
		}
		if note.Content == "" {
			t.Error("note content should not be empty")
		}
		if note.Type == "" {
			t.Error("note type should not be empty")
		}
		if note.CreatedAt.IsZero() {
			t.Error("note created_at should not be zero")
		}
		seenTypes[note.Type] = true
	}

	if len(seenTypes) == 0 {
		t.Error("expected at least one note type")
	}
}

func TestGenerateLinks(t *testing.T) {
	noteIDs := []string{"a", "b", "c", "d", "e"}
	links := generateLinks(noteIDs, 10)

	if len(links) != 10 {
		t.Fatalf("expected 10 links, got %d", len(links))
	}

	for _, link := range links {
		if link.SourceNoteID == "" || link.TargetNoteID == "" {
			t.Error("link source and target should not be empty")
		}
		if link.SourceNoteID == link.TargetNoteID {
			t.Error("source and target should be different")
		}
		if link.LinkType == "" {
			t.Error("link type should not be empty")
		}
		if link.Weight < 0.5 || link.Weight > 1.0 {
			t.Errorf("link weight %f out of expected range", link.Weight)
		}
	}
}
