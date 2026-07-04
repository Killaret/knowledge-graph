package note

import (
	"errors"
	"strings"
)

// Title is the note title (cannot be empty, max 200 characters)
type Title struct {
	value string
}

func NewTitle(value string) (Title, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) == 0 {
		return Title{}, errors.New("title cannot be empty")
	}
	if len(trimmed) > 200 {
		return Title{}, errors.New("title too long (max 200 characters)")
	}
	return Title{value: trimmed}, nil
}

func (t Title) String() string {
	return t.value
}

// Content is the note body (plain text)
type Content struct {
	value string
}

func NewContent(value string) (Content, error) {
	// Constraints could be added here, e.g. no more than 10000 characters
	if len(value) > 10000 {
		return Content{}, errors.New("content too long (max 10000 characters)")
	}
	return Content{value: value}, nil
}

func (c Content) String() string {
	return c.value
}

// Metadata holds additional note data (tags, status, etc.)
type Metadata struct {
	value map[string]interface{}
}

func NewMetadata(value map[string]interface{}) (Metadata, error) {
	// Validation could be added here, but for now we leave it as is
	return Metadata{value: value}, nil
}

func (m Metadata) Value() map[string]interface{} {
	return m.value
}
