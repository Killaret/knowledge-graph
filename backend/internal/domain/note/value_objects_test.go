package note

import (
	"testing"
)

func TestNewTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"too long", string(make([]byte, 201)), true},
		{"valid", "Hello", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTitle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewContent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", false},
		{"too long", string(make([]byte, 10001)), true},
		{"valid", "Content", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewContent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMetadata(t *testing.T) {
	_, err := NewMetadata(map[string]interface{}{"key": "value"})
	if err != nil {
		t.Errorf("NewMetadata() error = %v", err)
	}
}
