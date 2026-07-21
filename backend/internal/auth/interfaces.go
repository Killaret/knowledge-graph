package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TokenStore abstracts the Redis-backed token cache used by auth handlers and middleware.
// RedisTokenStore is the production implementation.
type TokenStore interface {
	BlacklistToken(ctx context.Context, token string, ttl time.Duration) error
	IsTokenBlacklisted(ctx context.Context, token string) (bool, error)
	StoreRefreshToken(ctx context.Context, userID string, token string, expiresAt time.Time) error
	ValidateRefreshToken(ctx context.Context, token string) (string, error)
	RevokeRefreshToken(ctx context.Context, token string, ttl time.Duration) error
	StorePasswordResetToken(ctx context.Context, userID string, token string, ttl time.Duration) error
	ValidatePasswordResetToken(ctx context.Context, token string) (string, error)
	DeletePasswordResetToken(ctx context.Context, token string) error
	StorePKCE(ctx context.Context, state string, pkce *PKCE, ttl time.Duration) error
	GetPKCE(ctx context.Context, state string) (*PKCE, error)
	CachePermission(ctx context.Context, userID, resource, action string, allowed bool, ttl time.Duration) error
	CheckCachedPermission(ctx context.Context, userID, resource, action string) (bool, bool, error)
	InvalidatePermissionCache(ctx context.Context, userID string) error
}

// RefreshToken represents a persisted refresh token.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	IPAddress string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// RefreshTokenRepository persists refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
}

// EmailSender abstracts sending password-reset emails.
type EmailSender interface {
	// SendPasswordReset sends a password-reset email to the given address.
	// The resetLink is a fully qualified URL that the user should follow.
	SendPasswordReset(ctx context.Context, to, resetLink string) error
}

// OAuthUserInfo represents the subset of an OAuth provider profile needed by auth handlers.
type OAuthUserInfo struct {
	ID     string
	Login  string
	Email  string
	Emails []string
}

// OAuthProvider abstracts OAuth token exchange and user info retrieval.
type OAuthProvider interface {
	Exchange(ctx context.Context, code, codeVerifier string) (string, error)
	UserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error)
}

// OAuthProviderFactory creates an OAuth provider with runtime redirect URI.
type OAuthProviderFactory func(clientID, clientSecret, redirectURI string) OAuthProvider
