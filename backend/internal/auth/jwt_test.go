package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewJWTManager(t *testing.T) {
	secret := "test-secret-key"
	accessTTL := 15 * time.Minute
	refreshTTL := 7 * 24 * time.Hour

	manager := NewJWTManager(secret, accessTTL, refreshTTL)
	if manager == nil {
		t.Fatal("JWTManager should not be nil")
	}
}

func TestGenerateAndValidateTokenPair(t *testing.T) {
	secret := "test-secret-key"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	login := "testuser"
	role := "user"

	tokens, err := manager.GenerateTokenPair(userID, login, role)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Access token should not be empty")
	}

	if tokens.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}

	if tokens.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got '%s'", tokens.TokenType)
	}

	// Validate access token
	claims, err := manager.ValidateToken(tokens.AccessToken, "access")
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, claims.UserID)
	}

	if claims.Login != login {
		t.Errorf("Expected login %s, got %s", login, claims.Login)
	}

	if claims.Role != role {
		t.Errorf("Expected role %s, got %s", role, claims.Role)
	}

	if claims.TokenType != "access" {
		t.Errorf("Expected token type 'access', got '%s'", claims.TokenType)
	}

	// Validate refresh token
	refreshClaims, err := manager.ValidateToken(tokens.RefreshToken, "refresh")
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if refreshClaims.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, refreshClaims.UserID)
	}

	if refreshClaims.TokenType != "refresh" {
		t.Errorf("Expected token type 'refresh', got '%s'", refreshClaims.TokenType)
	}
}

func TestValidateToken_InvalidType(t *testing.T) {
	secret := "test-secret-key"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	userID := uuid.New()
	tokens, _ := manager.GenerateTokenPair(userID, "testuser", "user")

	// Try to validate access token as refresh token
	_, err := manager.ValidateToken(tokens.AccessToken, "refresh")
	if err == nil {
		t.Error("Should fail when validating access token as refresh token")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	secret := "test-secret-key"
	manager := NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)

	// Validate invalid token
	_, err := manager.ValidateToken("invalid-token", "access")
	if err == nil {
		t.Error("Should fail when validating invalid token")
	}
}

func TestTokenClaims_Expiry(t *testing.T) {
	secret := "test-secret-key"
	// Use very short TTL for testing
	manager := NewJWTManager(secret, 1*time.Millisecond, 1*time.Millisecond)

	userID := uuid.New()
	tokens, err := manager.GenerateTokenPair(userID, "testuser", "user")
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Validate expired token
	_, err = manager.ValidateToken(tokens.AccessToken, "access")
	if err == nil {
		t.Error("Should fail when validating expired token")
	}
}
