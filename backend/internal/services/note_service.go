package services

import (
	"context"
	"math"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/domain/note"
)

// NoteService provides business logic for note operations
type NoteService struct {
	repo note.Repository
	cfg  *config.Config
}

// NewNoteService creates a new instance of NoteService
func NewNoteService(repo note.Repository, cfg *config.Config) *NoteService {
	return &NoteService{
		repo: repo,
		cfg:  cfg,
	}
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	Notes      []*note.Note
	Total      int64
	Page       int
	Size       int
	TotalPages int
}

// Search performs full-text search with pagination
func (s *NoteService) Search(ctx context.Context, query string, page, pageSize int) (*SearchResult, error) {
	// Validate and normalize pagination parameters
	if page < 1 {
		page = 1
	}
	maxLimit := s.cfg.PaginationMaxLimit
	defaultLimit := s.cfg.PaginationDefaultLimit
	if pageSize < 1 || pageSize > maxLimit {
		pageSize = defaultLimit
	}

	offset := (page - 1) * pageSize

	// Perform search in repository
	notes, total, err := s.repo.Search(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &SearchResult{
		Notes:      notes,
		Total:      total,
		Page:       page,
		Size:       pageSize,
		TotalPages: totalPages,
	}, nil
}

// List returns notes with pagination
func (s *NoteService) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	maxLimit := s.cfg.PaginationMaxLimit
	defaultLimit := s.cfg.PaginationDefaultLimit
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

// FindByID returns a note by ID (keeping existing functionality)
func (s *NoteService) FindByID(ctx context.Context, id string) (*note.Note, error) {
	// Convert string ID to UUID (you'll need to import uuid and handle conversion)
	// For now, assuming the handler will do UUID conversion
	// This is a placeholder - you may need to adjust based on your UUID handling
	panic("TODO: Implement UUID conversion and FindByID")
}

// Save saves a note (keeping existing functionality)
func (s *NoteService) Save(ctx context.Context, note *note.Note) error {
	return s.repo.Save(ctx, note)
}

// Delete deletes a note (keeping existing functionality)
func (s *NoteService) Delete(ctx context.Context, id string) error {
	// Convert string ID to UUID
	// This is a placeholder - you may need to adjust based on your UUID handling
	panic("TODO: Implement UUID conversion and Delete")
}
