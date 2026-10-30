package textutil

import "unicode/utf8"

// TruncateToMaxBytes returns the longest valid UTF-8 prefix of s that does not
// exceed maxBytes bytes. It never splits a multi-byte rune.
func TruncateToMaxBytes(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}

	b := []byte(s)
	cut := maxBytes
	// Walk backwards past UTF-8 continuation bytes (10xxxxxx).
	for cut > 0 && b[cut]&0xC0 == 0x80 {
		cut--
	}
	return string(b[:cut])
}

// TruncateToMaxRunes returns a string containing at most maxRunes runes.
// It preserves valid UTF-8 and replaces invalid byte sequences with the
// Unicode replacement character before counting runes.
func TruncateToMaxRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}

// SanitizeUTF8 replaces invalid UTF-8 byte sequences with the Unicode
// replacement character. The returned string is guaranteed to be valid UTF-8.
func SanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return string([]rune(s))
}
