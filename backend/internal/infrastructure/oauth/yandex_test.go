package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewYandex(t *testing.T) {
	p := NewYandex("id", "secret", "http://localhost/cb")
	assert.Equal(t, "id", p.clientID)
	assert.Equal(t, "secret", p.clientSecret)
	assert.Equal(t, "http://localhost/cb", p.redirectURI)
}

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

func TestYandexProvider_Exchange_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	provider := &YandexProvider{
		tokenURL: ts.URL + "/token",
		client:   ts.Client(),
	}

	_, err := provider.Exchange(context.Background(), "code", "")
	assert.Error(t, err)
}

func TestYandexProvider_Exchange_MissingAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"token_type": "bearer"})
	}))
	defer ts.Close()

	provider := &YandexProvider{
		tokenURL: ts.URL + "/token",
		client:   ts.Client(),
	}

	_, err := provider.Exchange(context.Background(), "code", "")
	assert.Error(t, err)
}

func TestYandexProvider_UserInfo_EmptyEmail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":     "123",
			"login":  "user",
			"emails": []string{},
		})
	}))
	defer ts.Close()

	provider := &YandexProvider{
		infoURL: ts.URL + "/info",
		client:  ts.Client(),
	}

	info, err := provider.UserInfo(context.Background(), "token")
	assert.NoError(t, err)
	assert.Equal(t, "", info.Email)
}

func TestYandexProvider_UserInfo_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not-json"))
	}))
	defer ts.Close()

	provider := &YandexProvider{
		infoURL: ts.URL + "/info",
		client:  ts.Client(),
	}

	_, err := provider.UserInfo(context.Background(), "token")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "parse"))
}

func TestYandexProvider_UserInfo_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	provider := &YandexProvider{
		infoURL: ts.URL + "/info",
		client:  ts.Client(),
	}

	_, err := provider.UserInfo(context.Background(), "token")
	assert.Error(t, err)
}
