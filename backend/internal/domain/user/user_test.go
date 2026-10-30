package user

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_Valid(t *testing.T) {
	now := time.Now()
	u, err := NewUser(uuid.New(), "login", "user@example.com", "hash", "user", now, now, nil)
	require.NoError(t, err)
	assert.Equal(t, "login", u.Login())
	assert.Equal(t, "user@example.com", u.Email())
	assert.Equal(t, "hash", u.PasswordHash())
	assert.Equal(t, "user", u.Role())
	assert.False(t, u.IsDeleted())
}

func TestNewUser_Validation(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	_, err := NewUser(id, "", "email", "hash", "user", now, now, nil)
	assert.ErrorIs(t, err, ErrLoginRequired)

	_, err = NewUser(id, "login", "", "hash", "user", now, now, nil)
	assert.ErrorIs(t, err, ErrEmailRequired)

	_, err = NewUser(id, "login", "email", "", "user", now, now, nil)
	assert.ErrorIs(t, err, ErrPasswordHashRequired)
}

func TestUser_Setters(t *testing.T) {
	now := time.Now()
	u, err := NewUser(uuid.New(), "login", "email", "hash", "user", now, now, nil)
	require.NoError(t, err)

	u.SetEmail("new@example.com")
	assert.Equal(t, "new@example.com", u.Email())

	u.SetPasswordHash("newhash")
	assert.Equal(t, "newhash", u.PasswordHash())
}

func TestNewAPIKey_Valid(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	userID := uuid.New()
	key, err := NewAPIKey(id, userID, "keyhash", "my-key", []string{"read"}, now)
	require.NoError(t, err)
	assert.Equal(t, id, key.ID())
	assert.Equal(t, userID, key.UserID())
	assert.Equal(t, "my-key", key.Name())
	assert.Equal(t, "keyhash", key.KeyHash())
	assert.Equal(t, []string{"read"}, key.Scopes())
	assert.True(t, key.IsActive())
}

func TestNewAPIKey_Validation(t *testing.T) {
	now := time.Now()
	_, err := NewAPIKey(uuid.New(), uuid.New(), "", "name", nil, now)
	assert.ErrorIs(t, err, ErrAPIKeyHashRequired)

	_, err = NewAPIKey(uuid.New(), uuid.New(), "hash", "", nil, now)
	assert.ErrorIs(t, err, ErrAPIKeyNameRequired)
}

func TestAPIKey_SetInactive(t *testing.T) {
	now := time.Now()
	key, err := NewAPIKey(uuid.New(), uuid.New(), "hash", "name", nil, now)
	require.NoError(t, err)

	key.SetInactive()
	assert.False(t, key.IsActive())
}
