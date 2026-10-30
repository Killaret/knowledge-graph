package user

import (
	"time"

	"github.com/google/uuid"
)

// User represents a registered user in the domain layer.
type User struct {
	id           uuid.UUID
	login        string
	email        string
	passwordHash string
	role         string
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewUser creates a validated user aggregate.
func NewUser(id uuid.UUID, login, email, passwordHash, role string, createdAt, updatedAt time.Time, deletedAt *time.Time) (*User, error) {
	if login == "" {
		return nil, ErrLoginRequired
	}
	if email == "" {
		return nil, ErrEmailRequired
	}
	if passwordHash == "" {
		return nil, ErrPasswordHashRequired
	}
	return &User{
		id:           id,
		login:        login,
		email:        email,
		passwordHash: passwordHash,
		role:         role,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}, nil
}

func (u *User) ID() uuid.UUID         { return u.id }
func (u *User) Login() string         { return u.login }
func (u *User) Email() string         { return u.email }
func (u *User) PasswordHash() string  { return u.passwordHash }
func (u *User) Role() string          { return u.role }
func (u *User) CreatedAt() time.Time  { return u.createdAt }
func (u *User) UpdatedAt() time.Time  { return u.updatedAt }
func (u *User) DeletedAt() *time.Time { return u.deletedAt }
func (u *User) IsDeleted() bool       { return u.deletedAt != nil }

// SetEmail changes the user's email.
func (u *User) SetEmail(email string) {
	u.email = email
}

// SetPasswordHash updates the password hash (caller must hash the password first).
func (u *User) SetPasswordHash(hash string) {
	u.passwordHash = hash
}

// APIKey represents a user API key in the domain layer.
type APIKey struct {
	id         uuid.UUID
	userID     uuid.UUID
	keyHash    string
	name       string
	scopes     []string
	createdAt  time.Time
	expiresAt  *time.Time
	lastUsedAt *time.Time
	isActive   bool
}

// NewAPIKey creates a new active API key.
func NewAPIKey(id, userID uuid.UUID, keyHash, name string, scopes []string, createdAt time.Time) (*APIKey, error) {
	if keyHash == "" {
		return nil, ErrAPIKeyHashRequired
	}
	if name == "" {
		return nil, ErrAPIKeyNameRequired
	}
	return &APIKey{
		id:        id,
		userID:    userID,
		keyHash:   keyHash,
		name:      name,
		scopes:    scopes,
		createdAt: createdAt,
		isActive:  true,
	}, nil
}

func (k *APIKey) ID() uuid.UUID          { return k.id }
func (k *APIKey) UserID() uuid.UUID      { return k.userID }
func (k *APIKey) KeyHash() string        { return k.keyHash }
func (k *APIKey) Name() string           { return k.name }
func (k *APIKey) Scopes() []string       { return k.scopes }
func (k *APIKey) CreatedAt() time.Time   { return k.createdAt }
func (k *APIKey) ExpiresAt() *time.Time  { return k.expiresAt }
func (k *APIKey) LastUsedAt() *time.Time { return k.lastUsedAt }
func (k *APIKey) IsActive() bool         { return k.isActive }

// SetInactive marks the key as revoked.
func (k *APIKey) SetInactive() {
	k.isActive = false
}
