// Package validation provides common input validation utilities
package validation

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Common error messages
const (
	MsgInvalidUUID       = "Неверный формат UUID"
	MsgInvalidCharacters = "Поле содержит недопустимые символы"
	MsgTooLong           = "Поле слишком длинное"
	msgTooShort          = "Поле слишком короткое" // unused, reserved for future
	MsgRequired          = "Поле обязательно для заполнения"
	msgOutOfRange        = "Значение вне допустимого диапазона" // unused, reserved for future
	MsgInvalidEnum       = "Недопустимое значение перечисления"
)

// Allowed characters patterns for safe input
var (
	// SafeNamePattern allows letters, numbers, spaces, hyphens, underscores, dots
	// Prevents SQL injection and XSS attempts
	SafeNamePattern = regexp.MustCompile(`^[\p{L}\p{N}\s\-_\.]+$`)

	// SafeContentPattern allows broader range for content but still blocks dangerous chars
	// Allows unicode letters, numbers, common punctuation, spaces, newlines
	SafeContentPattern = regexp.MustCompile(`^[\p{L}\p{N}\s\-_.,!?;:()\[\]{}"'/@#%&*+=|<>^~\r\n]*$`)

	// SafeTagPattern for tag names - more restrictive
	SafeTagPattern = regexp.MustCompile(`^[\p{L}\p{N}\-_]+$`)

	// strictIDPattern for external IDs that should be alphanumeric only (unused, reserved)
	strictIDPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
)

// ValidCelestialBodyTypes contains all allowed celestial body types
// Must match the oneof validation in createNoteRequest.Type: star planet comet galaxy asteroid satellite debris nebula
var ValidCelestialBodyTypes = map[string]bool{
	"star":      true,
	"planet":    true,
	"comet":     true,
	"galaxy":    true,
	"asteroid":  true,
	"satellite": true,
	"debris":    true,
	"nebula":    true,
}

// ValidLinkTypes contains all allowed link types
var ValidLinkTypes = map[string]bool{
	"reference":  true,
	"dependency": true,
	"related":    true,
	"custom":     true,
}

// ValidationResult holds the result of a validation check
type ValidationResult struct {
	Valid   bool
	Field   string
	Message string
	Value   interface{}
}

// IsValidUUID checks if a string is a valid UUID
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// IsSafeName validates that a string contains only safe characters for names
func IsSafeName(name string, maxLength int) *ValidationResult {
	if name == "" {
		return &ValidationResult{Valid: false, Message: MsgRequired}
	}
	if len(name) > maxLength {
		return &ValidationResult{Valid: false, Message: MsgTooLong}
	}
	if !SafeNamePattern.MatchString(name) {
		return &ValidationResult{Valid: false, Message: MsgInvalidCharacters}
	}
	return &ValidationResult{Valid: true}
}

// IsSafeContent validates content string with length check
func IsSafeContent(content string, maxLength int) *ValidationResult {
	if len(content) > maxLength {
		return &ValidationResult{Valid: false, Message: MsgTooLong}
	}
	// Content can be empty, but if not empty, must be safe
	if content != "" && !SafeContentPattern.MatchString(content) {
		return &ValidationResult{Valid: false, Message: MsgInvalidCharacters}
	}
	return &ValidationResult{Valid: true}
}

// IsSafeTag validates tag name
func IsSafeTag(tag string, maxLength int) *ValidationResult {
	if tag == "" {
		return &ValidationResult{Valid: false, Message: MsgRequired}
	}
	if len(tag) > maxLength {
		return &ValidationResult{Valid: false, Message: MsgTooLong}
	}
	if !SafeTagPattern.MatchString(tag) {
		return &ValidationResult{Valid: false, Message: MsgInvalidCharacters}
	}
	return &ValidationResult{Valid: true}
}

// IsValidCelestialBodyType checks if the given type is valid (case-sensitive)
func IsValidCelestialBodyType(t string) bool {
	return ValidCelestialBodyTypes[t]
}

// IsValidLinkType checks if the given link type is valid (case-sensitive)
func IsValidLinkType(t string) bool {
	return ValidLinkTypes[t]
}

// IsValidWeight checks if weight is in valid range [0, 1]
func IsValidWeight(weight float64) bool {
	return weight >= 0 && weight <= 1
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(input string) string {
	// Remove null bytes and control characters (except common whitespace)
	var result strings.Builder
	for _, r := range input {
		if r == 0 {
			continue // Skip null bytes
		}
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			continue // Skip control characters except tab, newline, carriage return
		}
		result.WriteRune(r)
	}
	return result.String()
}

// TrimAndNormalize normalizes whitespace in a string
func TrimAndNormalize(input string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	normalized := re.ReplaceAllString(strings.TrimSpace(input), " ")
	return normalized
}
