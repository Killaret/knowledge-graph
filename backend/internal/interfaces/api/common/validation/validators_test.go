package validation

import (
	"testing"
)

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid UUID v4", "550e8400-e29b-41d4-a716-446655440000", true},
		{"valid UUID v4 lowercase", "550e8400-e29b-41d4-a716-446655440000", true},
		{"invalid UUID - too short", "550e8400-e29b-41d4", false},
		{"invalid UUID - wrong format", "not-a-uuid", false},
		{"invalid UUID - empty", "", false},
		{"invalid UUID - special chars", "550e8400@e29b#41d4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidUUID(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidUUID(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsSafeName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantValid bool
		wantMsg   string
	}{
		{"valid simple name", "Test Note", 50, true, ""},
		{"valid with hyphen", "Test-Note", 50, true, ""},
		{"valid with underscore", "Test_Note", 50, true, ""},
		{"valid with dot", "Test.Note", 50, true, ""},
		{"valid with spaces", "Test Note Name", 50, true, ""},
		{"valid unicode", "Тест Заметка", 50, true, ""},
		{"valid numbers", "Note123", 50, true, ""},
		{"empty name", "", 50, false, MsgRequired},
		{"too long", "This is a very long name that exceeds the limit", 20, false, MsgTooLong},
		{"sql injection attempt", "Note'; DROP TABLE notes;--", 50, false, MsgInvalidCharacters},
		{"xss attempt", "<script>alert('xss')</script>", 50, false, MsgInvalidCharacters},
		{"null byte", "Test\x00Note", 50, false, MsgInvalidCharacters},
		{"control char", "Test\x01Note", 50, false, MsgInvalidCharacters},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSafeName(tt.input, tt.maxLength)
			if result.Valid != tt.wantValid {
				t.Errorf("IsSafeName(%q, %d).Valid = %v, want %v",
					tt.input, tt.maxLength, result.Valid, tt.wantValid)
			}
			if !tt.wantValid && result.Message != tt.wantMsg {
				t.Errorf("IsSafeName(%q, %d).Message = %q, want %q",
					tt.input, tt.maxLength, result.Message, tt.wantMsg)
			}
		})
	}
}

func TestIsSafeTag(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantValid bool
		wantMsg   string
	}{
		{"valid tag", "important", 50, true, ""},
		{"valid with hyphen", "high-priority", 50, true, ""},
		{"valid with underscore", "urgent_tag", 50, true, ""},
		{"valid numbers", "tag123", 50, true, ""},
		{"valid unicode", "важно", 50, true, ""},
		{"empty tag", "", 50, false, MsgRequired},
		{"too long", "this-is-a-very-long-tag-name", 20, false, MsgTooLong},
		{"spaces not allowed", "important tag", 50, false, MsgInvalidCharacters},
		{"dot not allowed", "important.tag", 50, false, MsgInvalidCharacters},
		{"sql injection", "tag'; DROP TABLE tags;--", 50, false, MsgInvalidCharacters},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSafeTag(tt.input, tt.maxLength)
			if result.Valid != tt.wantValid {
				t.Errorf("IsSafeTag(%q, %d).Valid = %v, want %v",
					tt.input, tt.maxLength, result.Valid, tt.wantValid)
			}
			if !tt.wantValid && result.Message != tt.wantMsg {
				t.Errorf("IsSafeTag(%q, %d).Message = %q, want %q",
					tt.input, tt.maxLength, result.Message, tt.wantMsg)
			}
		})
	}
}

func TestIsSafeContent(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		wantValid bool
		wantMsg   string
	}{
		{"valid content", "This is normal content", 1000, true, ""},
		{"valid with punctuation", "Hello! How are you? (Test).", 1000, true, ""},
		{"valid empty", "", 1000, true, ""},
		{"valid multiline", "Line 1\nLine 2\r\nLine 3", 1000, true, ""},
		{"valid unicode", "Привет мир! 你好世界", 1000, true, ""},
		{"valid code snippet", "function test() { return 42; }", 1000, true, ""},
		{"too long", "This content is way too long for the limit", 20, false, MsgTooLong},
		{"null byte", "Content\x00WithNull", 1000, false, MsgInvalidCharacters},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSafeContent(tt.input, tt.maxLength)
			if result.Valid != tt.wantValid {
				t.Errorf("IsSafeContent(%q, %d).Valid = %v, want %v",
					tt.input, tt.maxLength, result.Valid, tt.wantValid)
			}
			if !tt.wantValid && result.Message != tt.wantMsg {
				t.Errorf("IsSafeContent(%q, %d).Message = %q, want %q",
					tt.input, tt.maxLength, result.Message, tt.wantMsg)
			}
		})
	}
}

func TestIsValidCelestialBodyType(t *testing.T) {
	// Must match the oneof validation in createNoteRequest.Type: star planet comet galaxy asteroid satellite debris nebula
	validTypes := []string{"star", "planet", "comet", "galaxy", "asteroid", "satellite", "debris", "nebula"}
	invalidTypes := []string{"invalid", "unknown", "", "STAR", "Planet", "moon", "blackhole"} // case-sensitive check

	for _, tt := range validTypes {
		t.Run("valid "+tt, func(t *testing.T) {
			if !IsValidCelestialBodyType(tt) {
				t.Errorf("IsValidCelestialBodyType(%q) = false, want true", tt)
			}
		})
	}

	for _, tt := range invalidTypes {
		t.Run("invalid "+tt, func(t *testing.T) {
			if IsValidCelestialBodyType(tt) {
				t.Errorf("IsValidCelestialBodyType(%q) = true, want false", tt)
			}
		})
	}
}

func TestIsValidLinkType(t *testing.T) {
	validTypes := []string{"reference", "dependency", "related", "custom"}
	invalidTypes := []string{"invalid", "unknown", "", "REFERENCE"}

	for _, tt := range validTypes {
		t.Run("valid "+tt, func(t *testing.T) {
			if !IsValidLinkType(tt) {
				t.Errorf("IsValidLinkType(%q) = false, want true", tt)
			}
		})
	}

	for _, tt := range invalidTypes {
		t.Run("invalid "+tt, func(t *testing.T) {
			if IsValidLinkType(tt) {
				t.Errorf("IsValidLinkType(%q) = true, want false", tt)
			}
		})
	}
}

func TestIsValidWeight(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected bool
	}{
		{"valid 0", 0, true},
		{"valid 1", 1, true},
		{"valid 0.5", 0.5, true},
		{"valid 0.0", 0.0, true},
		{"invalid negative", -0.1, false},
		{"invalid > 1", 1.1, false},
		{"invalid large", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidWeight(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidWeight(%f) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal string", "Hello World", "Hello World"},
		{"with null byte", "Hello\x00World", "HelloWorld"},
		{"with control char", "Hello\x01World", "HelloWorld"},
		{"preserve tabs", "Hello\tWorld", "Hello\tWorld"},
		{"preserve newlines", "Hello\nWorld", "Hello\nWorld"},
		{"preserve carriage return", "Hello\rWorld", "Hello\rWorld"},
		{"multiple null bytes", "\x00H\x00e\x00l\x00l\x00o\x00", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTrimAndNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal string", "Hello World", "Hello World"},
		{"leading space", "  Hello World", "Hello World"},
		{"trailing space", "Hello World  ", "Hello World"},
		{"multiple spaces", "Hello   World", "Hello World"},
		{"mixed whitespace", "Hello\t\t\tWorld", "Hello World"},
		{"newlines normalized", "Hello\n\n\nWorld", "Hello World"},
		{"empty string", "", ""},
		{"only spaces", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TrimAndNormalize(tt.input)
			if result != tt.expected {
				t.Errorf("TrimAndNormalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
