package user

import (
	"context"

	"github.com/google/uuid"
)

// Repository handles persistence for User aggregates.
type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByLogin(ctx context.Context, login string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	EmailExists(ctx context.Context, email string, excludeID uuid.UUID) (bool, error)
}

// APIKeyRepository handles persistence for API keys.
type APIKeyRepository interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]APIKey, error)
	Create(ctx context.Context, key *APIKey) error
	Revoke(ctx context.Context, keyID, userID uuid.UUID) (bool, error)
}
