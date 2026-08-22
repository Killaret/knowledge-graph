package link

import "errors"

// ErrDuplicateLink is returned when a link of the same type already exists between two notes.
var ErrDuplicateLink = errors.New("link of this type already exists between these notes")

// ErrLinkNotFound is returned when a link with the given ID does not exist.
var ErrLinkNotFound = errors.New("link not found")
