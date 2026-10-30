package tag

import "errors"

// Domain errors for Tag entity.
var (
	ErrEmptyName   = errors.New("tag name cannot be empty")
	ErrNameTooLong = errors.New("tag name cannot exceed 50 characters")
)
