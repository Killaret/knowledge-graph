package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePKCE_S256(t *testing.T) {
	pkce, err := GeneratePKCE(128)
	require.NoError(t, err)
	require.NotNil(t, pkce)

	assert.Equal(t, "S256", pkce.CodeChallengeMethod)
	assert.NotEmpty(t, pkce.CodeVerifier)
	assert.NotEmpty(t, pkce.CodeChallenge)

	// RFC 7636 S256: code_challenge = BASE64URL-ENCODE(SHA256(ASCII(code_verifier)))
	hash := sha256.Sum256([]byte(pkce.CodeVerifier))
	wantChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	assert.Equal(t, wantChallenge, pkce.CodeChallenge)

	// No padding characters must be present.
	assert.NotContains(t, pkce.CodeChallenge, "=")
	assert.NotContains(t, pkce.CodeChallenge, "+")
	assert.NotContains(t, pkce.CodeChallenge, "/")
}

func TestGeneratePKCE_VerifierLength(t *testing.T) {
	pkce, err := GeneratePKCE(128)
	require.NoError(t, err)

	// GenerateRandomToken(128) produces base64url of 128 random bytes.
	// base64url encoding of 128 bytes is ceil(128*4/3) = 172 chars without padding.
	assert.Equal(t, 172, len(pkce.CodeVerifier))
}
