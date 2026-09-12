package textutil

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestTruncateToMaxBytes(t *testing.T) {
	russian := "Полное руководство" // 34 bytes, 18 runes

	tests := []struct {
		name     string
		input    string
		maxBytes int
		want     string
	}{
		{"no truncation", russian, 100, russian},
		{"zero", russian, 0, ""},
		{"split inside multi-byte rune", russian, 33, russian[:33]},
		{"exact leading rune", "Привет", 2, "П"},
		{"inside leading rune", "Привет", 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateToMaxBytes(tt.input, tt.maxBytes)
			require.Equal(t, tt.want, got)
			require.True(t, utf8.ValidString(got))
		})
	}
}

func TestTruncateToMaxRunes(t *testing.T) {
	russian := "Полное руководство" // 18 runes

	tests := []struct {
		name     string
		input    string
		maxRunes int
		want     string
	}{
		{"no truncation", russian, 100, russian},
		{"zero", russian, 0, ""},
		{"truncate", russian, 5, "Полно"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateToMaxRunes(tt.input, tt.maxRunes)
			require.Equal(t, tt.want, got)
			require.LessOrEqual(t, utf8.RuneCountInString(got), tt.maxRunes)
		})
	}
}

func TestSanitizeUTF8(t *testing.T) {
	// A string that ends with an incomplete 2-byte UTF-8 sequence.
	invalid := "test\xD0"
	got := SanitizeUTF8(invalid)
	require.True(t, utf8.ValidString(got))
	require.True(t, strings.HasPrefix(got, "test"))
}
