package main

import (
	"testing"
)

func TestFindSubstring(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   int
	}{
		{"hello world", "world", 6},
		{"hello world", "foo", -1},
		{"", "foo", -1},
		{"foo", "foobar", -1},
		{"abcabc", "abc", 0},
		{"hello", "", 0},
	}

	for _, tt := range tests {
		got := findSubstring(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("findSubstring(%q, %q) = %d, want %d", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestMaskURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"", "(empty)"},
		{"postgres://user:pass@localhost/db", "postgres://user:***@localhost/db"},
		{"postgres://user@localhost/db", "postgres://user@localhost/db"},
		{"postgres://localhost/db", "postgres://localhost/db"},
		{"redis://localhost:6379", "redis://localhost:6379"},
		{"mongodb://user:secret@mongo:27017/db", "mongodb://user:***@mongo:27017/db"},
	}

	for _, tt := range tests {
		got := maskURL(tt.url)
		if got != tt.want {
			t.Errorf("maskURL(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
