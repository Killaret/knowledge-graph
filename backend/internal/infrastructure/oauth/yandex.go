// Package oauth provides OAuth 2.0 provider integrations.
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	yandexTokenURL = "https://oauth.yandex.com/token"
	yandexInfoURL  = "https://login.yandex.ru/info"
	yandexTimeout  = 10 * time.Second
)

// YandexProvider handles Yandex OAuth token exchange and user info retrieval.
type YandexProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
	tokenURL     string
	infoURL      string
	client       *http.Client
}

// NewYandex creates a Yandex OAuth provider.
func NewYandex(clientID, clientSecret, redirectURI string) *YandexProvider {
	return &YandexProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		tokenURL:     yandexTokenURL,
		infoURL:      yandexInfoURL,
		client:       &http.Client{Timeout: yandexTimeout},
	}
}

// UserInfo represents the subset of the Yandex Passport response we need.
type UserInfo struct {
	ID     string   `json:"id"`
	Login  string   `json:"login"`
	Email  string   `json:"default_email"`
	Emails []string `json:"emails"`
}

// TokenResponse represents the Yandex token endpoint response.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Exchange swaps an authorization code (and optional PKCE verifier) for an access token.
func (p *YandexProvider) Exchange(ctx context.Context, code, codeVerifier string) (string, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", p.clientID)
	data.Set("client_secret", p.clientSecret)
	if p.redirectURI != "" {
		data.Set("redirect_uri", p.redirectURI)
	}
	if codeVerifier != "" {
		data.Set("code_verifier", codeVerifier)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed: status %d, body %s", resp.StatusCode, string(body))
	}

	var tok tokenResponse
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("token response contained no access_token")
	}
	return tok.AccessToken, nil
}

// UserInfo fetches the user profile from Yandex Passport.
func (p *YandexProvider) UserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.infoURL+"?format=json", nil)
	if err != nil {
		return nil, fmt.Errorf("build info request: %w", err)
	}
	req.Header.Set("Authorization", "OAuth "+accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read userinfo response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo failed: status %d, body %s", resp.StatusCode, string(body))
	}

	var info UserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse userinfo response: %w", err)
	}
	if info.Email == "" && len(info.Emails) > 0 {
		info.Email = info.Emails[0]
	}
	return &info, nil
}
