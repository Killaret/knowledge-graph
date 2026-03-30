package link

import (
	"testing"
)

func TestNewLinkType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"reference", "reference", false},
		{"dependency", "dependency", false},
		{"related", "related", false},
		{"custom", "custom", false},
		{"invalid", "invalid", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLinkType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLinkType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewWeight(t *testing.T) {
	tests := []struct {
		name    string
		input   float64
		wantErr bool
	}{
		{"zero", 0, false},
		{"half", 0.5, false},
		{"one", 1, false},
		{"negative", -0.1, true},
		{"too high", 1.1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWeight(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMetadata(t *testing.T) {
	_, err := NewMetadata(map[string]interface{}{"description": "test"})
	if err != nil {
		t.Errorf("NewMetadata() error = %v", err)
	}
}
