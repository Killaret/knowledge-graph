package note

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNote(t *testing.T) {
	tests := []struct {
		name         string
		title        string
		content      string
		noteType     string
		metadata     map[string]interface{}
		wantErr      bool
		expectedType string
	}{
		{
			name:         "valid note with star type",
			title:        "Test Title",
			content:      "Test Content",
			noteType:     "star",
			metadata:     map[string]interface{}{"tag": "test"},
			wantErr:      false,
			expectedType: "star",
		},
		{
			name:         "valid note with planet type",
			title:        "Planet",
			content:      "Important excerpt",
			noteType:     "planet",
			metadata:     map[string]interface{}{},
			wantErr:      false,
			expectedType: "planet",
		},
		{
			name:         "invalid note type falls back to unknown",
			title:        "Invalid",
			content:      "https://example.com",
			noteType:     "link",
			metadata:     map[string]interface{}{"url": "https://example.com"},
			wantErr:      false,
			expectedType: "unknown",
		},
		{
			name:         "empty note type defaults to star",
			title:        "Test",
			content:      "Content",
			noteType:     "",
			metadata:     map[string]interface{}{},
			wantErr:      false,
			expectedType: "star",
		},
		{
			name:         "nil metadata defaults to empty map",
			title:        "Test",
			content:      "Content",
			noteType:     "star",
			metadata:     nil,
			wantErr:      false,
			expectedType: "star",
		},
		{
			name:         "title with leading/trailing spaces",
			title:        "  Spaced Title  ",
			content:      "Content",
			noteType:     "star",
			metadata:     map[string]interface{}{},
			wantErr:      false,
			expectedType: "star",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, err := NewTitle(tt.title)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			content, err := NewContent(tt.content)
			require.NoError(t, err)

			var metadata Metadata
			if tt.metadata != nil {
				metadata, err = NewMetadata(tt.metadata)
				require.NoError(t, err)
			} else {
				metadata = Metadata{value: map[string]interface{}{}}
			}

			var nt NoteType
			if tt.noteType == "" {
				nt = Default()
			} else {
				var err error
				nt, err = NewType(tt.noteType)
				if err != nil {
					nt = NoteType{}
				}
			}
			note := NewNote(title, content, nt, metadata)

			assert.NotEqual(t, uuid.Nil, note.ID())
			assert.Equal(t, tt.expectedType, note.Type())
			assert.NotNil(t, note.Metadata().Value())
			assert.False(t, note.CreatedAt().IsZero())
			assert.False(t, note.UpdatedAt().IsZero())
			assert.Nil(t, note.CreatorID())
		})
	}
}

func TestNewNoteWithCreator(t *testing.T) {
	creatorID := uuid.New()
	title, _ := NewTitle("Test")
	content, _ := NewContent("Content")
	metadata, _ := NewMetadata(map[string]interface{}{})

	note := NewNoteWithCreator(title, content, MustType("star"), metadata, creatorID)

	assert.NotEqual(t, uuid.Nil, note.ID())
	assert.NotNil(t, note.CreatorID())
	assert.Equal(t, creatorID, *note.CreatorID())
}

func TestReconstructNote(t *testing.T) {
	id := uuid.New()
	title, _ := NewTitle("Reconstructed")
	content, _ := NewContent("Content")
	metadata, _ := NewMetadata(map[string]interface{}{})
	createdAt := time.Now().Add(-1 * time.Hour)
	updatedAt := time.Now()

	note := ReconstructNote(id, title, content, MustType("star"), metadata, createdAt, updatedAt)

	assert.Equal(t, id, note.ID())
	assert.Equal(t, "Reconstructed", note.Title().String())
	assert.Equal(t, createdAt, note.CreatedAt())
	assert.Equal(t, updatedAt, note.UpdatedAt())
}

func TestReconstructNoteWithCreator(t *testing.T) {
	id := uuid.New()
	creatorID := uuid.New()
	title, _ := NewTitle("Reconstructed")
	content, _ := NewContent("Content")
	metadata, _ := NewMetadata(map[string]interface{}{})
	createdAt := time.Now().Add(-1 * time.Hour)
	updatedAt := time.Now()

	note := ReconstructNoteWithCreator(id, title, content, MustType("star"), metadata, &creatorID, createdAt, updatedAt)

	assert.Equal(t, id, note.ID())
	assert.NotNil(t, note.CreatorID())
	assert.Equal(t, creatorID, *note.CreatorID())
}

func TestNote_SetCreatorID(t *testing.T) {
	title, _ := NewTitle("Test")
	content, _ := NewContent("Content")
	metadata, _ := NewMetadata(map[string]interface{}{})
	note := NewNote(title, content, MustType("star"), metadata)

	assert.Nil(t, note.CreatorID())

	creatorID := uuid.New()
	note.SetCreatorID(creatorID)

	assert.NotNil(t, note.CreatorID())
	assert.Equal(t, creatorID, *note.CreatorID())
}

func TestNote_SetType(t *testing.T) {
	tests := []struct {
		name       string
		current    string
		newType    string
		expectType string
	}{
		{
			name:       "valid type update",
			current:    "star",
			newType:    "planet",
			expectType: "planet",
		},
		{
			name:       "empty type not updated",
			current:    "star",
			newType:    "",
			expectType: "star",
		},
		{
			name:       "different type changes",
			current:    "planet",
			newType:    "star",
			expectType: "star",
		},
		{
			name:       "invalid type not updated",
			current:    "star",
			newType:    "invalid",
			expectType: "star",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, _ := NewTitle("Test")
			content, _ := NewContent("Content")
			metadata, _ := NewMetadata(map[string]interface{}{})
			note := NewNote(title, content, MustType(tt.current), metadata)

			oldUpdated := note.UpdatedAt()
			time.Sleep(10 * time.Millisecond) // Ensure time difference

			var newType NoteType
			if tt.newType != "" {
				var err error
				newType, err = NewType(tt.newType)
				if err != nil {
					newType = NoteType{}
				}
			}
			note.SetType(newType)

			assert.Equal(t, tt.expectType, note.Type())
			if newType.IsValid() && note.Type() != newType.String() {
				assert.True(t, note.UpdatedAt().After(oldUpdated))
			}
		})
	}
}

func TestNote_UpdateTitle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr bool
	}{
		{
			name:    "valid title update",
			title:   "New Title",
			wantErr: false,
		},
		{
			name:    "empty title fails",
			title:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only title fails",
			title:   "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldTitle, _ := NewTitle("Old")
			content, _ := NewContent("Content")
			metadata, _ := NewMetadata(map[string]interface{}{})
			note := NewNote(oldTitle, content, MustType("star"), metadata)

			newTitle, err := NewTitle(tt.title)
			if tt.wantErr {
				assert.Error(t, note.UpdateTitle(newTitle))
			} else {
				require.NoError(t, err)
				oldUpdated := note.UpdatedAt()
				time.Sleep(10 * time.Millisecond)

				err := note.UpdateTitle(newTitle)
				assert.NoError(t, err)
				assert.Equal(t, tt.title, note.Title().String())
				assert.True(t, note.UpdatedAt().After(oldUpdated))
			}
		})
	}
}

func TestNote_UpdateContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid content update",
			content: "New content here",
			wantErr: false,
		},
		{
			name:    "empty content fails (validation in UpdateContent)",
			content: "",
			wantErr: true,
		},
		{
			name:    "long content succeeds",
			content: string(make([]byte, 5000)),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, _ := NewTitle("Test")
			oldContent, _ := NewContent("Old")
			metadata, _ := NewMetadata(map[string]interface{}{})
			note := NewNote(title, oldContent, MustType("star"), metadata)

			newContent, err := NewContent(tt.content)
			require.NoError(t, err)

			oldUpdated := note.UpdatedAt()
			time.Sleep(10 * time.Millisecond)

			err = note.UpdateContent(newContent)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.content, note.Content().String())
				assert.True(t, note.UpdatedAt().After(oldUpdated))
			}
		})
	}
}

func TestNote_UpdateMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "valid metadata update",
			metadata: map[string]interface{}{"tag": "updated", "count": 5},
			wantErr:  false,
		},
		{
			name:     "empty metadata map succeeds",
			metadata: map[string]interface{}{},
			wantErr:  false,
		},
		{
			name:     "nil metadata fails",
			metadata: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, _ := NewTitle("Test")
			content, _ := NewContent("Content")
			oldMetadata, _ := NewMetadata(map[string]interface{}{"old": "value"})
			note := NewNote(title, content, MustType("star"), oldMetadata)

			var newMetadata Metadata
			if tt.metadata != nil {
				newMetadata, _ = NewMetadata(tt.metadata)
			} else {
				newMetadata = Metadata{value: nil}
			}

			err := note.UpdateMetadata(newMetadata)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.metadata, note.Metadata().Value())
			}
		})
	}
}

func TestNote_Getters(t *testing.T) {
	id := uuid.New()
	creatorID := uuid.New()
	title, _ := NewTitle("Test Title")
	content, _ := NewContent("Test Content")
	metadata, _ := NewMetadata(map[string]interface{}{"key": "value"})
	createdAt := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	note := ReconstructNoteWithCreator(id, title, content, MustType("star"), metadata, &creatorID, createdAt, updatedAt)

	assert.Equal(t, id, note.ID())
	assert.Equal(t, "Test Title", note.Title().String())
	assert.Equal(t, "Test Content", note.Content().String())
	assert.Equal(t, "star", note.Type())
	assert.Equal(t, creatorID, *note.CreatorID())
	assert.Equal(t, createdAt, note.CreatedAt())
	assert.Equal(t, updatedAt, note.UpdatedAt())
	assert.Equal(t, map[string]interface{}{"key": "value"}, note.Metadata().Value())
	assert.False(t, note.IsPublic())
}

func TestNote_SetIsPublic(t *testing.T) {
	title, _ := NewTitle("Test")
	content, _ := NewContent("Content")
	metadata, _ := NewMetadata(map[string]interface{}{"key": "value"})
	note := NewNote(title, content, MustType("star"), metadata)

	assert.False(t, note.IsPublic())

	beforeUpdate := note.UpdatedAt()
	note.SetIsPublic(true)

	assert.True(t, note.IsPublic())
	assert.False(t, note.UpdatedAt().Before(beforeUpdate))

	note.SetIsPublic(false)
	assert.False(t, note.IsPublic())
}
