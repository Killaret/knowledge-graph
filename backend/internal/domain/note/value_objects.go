package note

import (
	"errors"
	"strings"
)

// Title — заголовок заметки (не может быть пустым, макс 200 символов)
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

// Content — содержимое заметки (простой текст)
type Content struct {
	value string
}

func NewContent(value string) (Content, error) {
	// Можно добавить ограничения, например, не больше 10000 символов
	if len(value) > 10000 {
		return Content{}, errors.New("content too long (max 10000 characters)")
	}
	return Content{value: value}, nil
}

func (c Content) String() string {
	return c.value
}

// Metadata — дополнительные данные заметки (теги, статус и т.п.)
type Metadata struct {
	value map[string]interface{}
}

func NewMetadata(value map[string]interface{}) (Metadata, error) {
	// Можно добавить валидацию, но пока оставляем как есть
	return Metadata{value: value}, nil
}

func (m Metadata) Value() map[string]interface{} {
	return m.value
}
