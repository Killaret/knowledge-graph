package share

import (
	"context"

	"github.com/google/uuid"
)

// Repository handles persistence for note sharing.
type Repository interface {
	FindShareByNoteAndUser(ctx context.Context, noteID, sharedWithUserID uuid.UUID) (*NoteShare, error)
	CreateShare(ctx context.Context, share *NoteShare) error
	UpdateShare(ctx context.Context, share *NoteShare) error
	ListSharesByNote(ctx context.Context, noteID uuid.UUID) ([]NoteShare, error)
	RevokeShare(ctx context.Context, noteID, shareID, sharedByUserID uuid.UUID) (bool, error)

	CreateShareLink(ctx context.Context, link *ShareLink) error
	FindActiveShareLinkByToken(ctx context.Context, token string) (*ShareLink, error)
	RevokeShareLink(ctx context.Context, linkID, sharedByUserID uuid.UUID) (bool, error)
	ListShareLinksByNote(ctx context.Context, noteID uuid.UUID) ([]ShareLink, error)
	IncrementShareLinkUsage(ctx context.Context, linkID uuid.UUID) error
}
