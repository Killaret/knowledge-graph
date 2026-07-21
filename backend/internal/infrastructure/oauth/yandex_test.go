package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestYandexProvider_ExchangeAndUserInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			assert.Equal(t, http.MethodPost, r.Method)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "test-access-token",
				"token_type":   "bearer",
				"expires_in":   3600,
			})
		case "/info":
			assert.Equal(t, "OAuth test-access-token", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":            "12345",
				"login":         "testuser",
				"default_email": "test@example.com",
				"emails":        []string{"test@example.com"},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	provider := &YandexProvider{
		clientID:     "client-id",
		clientSecret: "client-secret",
		redirectURI:  "http://localhost/callback",
		tokenURL:     ts.URL + "/token",
		infoURL:      ts.URL + "/info",
		client:       ts.Client(),
	}

	accessToken, err := provider.Exchange(context.Background(), "auth-code", "verifier")
	require.NoError(t, err)
	assert.Equal(t, "test-access-token", accessToken)

	info, err := provider.UserInfo(context.Background(), accessToken)
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", info.Email)
	assert.Equal(t, "testuser", info.Login)
}
