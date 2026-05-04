package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	config := DefaultPasswordConfig()
	password := "TestPassword123!"

	hash, err := HashPassword(password, config)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}

	if hash == password {
		t.Error("Hash should not be the same as password")
	}
}

func TestVerifyPassword(t *testing.T) {
	config := DefaultPasswordConfig()
	password := "TestPassword123!"

	hash, err := HashPassword(password, config)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test correct password
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if !valid {
		t.Error("Valid password should be accepted")
	}

	// Test incorrect password
	valid, err = VerifyPassword("WrongPassword", hash)
	if err != nil {
		t.Fatalf("Failed to verify password: %v", err)
	}
	if valid {
		t.Error("Invalid password should be rejected")
	}
}

func TestValidatePassword(t *testing.T) {
	policy := DefaultPasswordPolicy()

	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{
			name:     "valid password",
			password: "TestPass123!",
			valid:    true,
		},
		{
			name:     "too short",
			password: "Test1!",
			valid:    false,
		},
		{
			name:     "no uppercase",
			password: "testpass123!",
			valid:    false,
		},
		{
			name:     "no lowercase",
			password: "TESTPASS123!",
			valid:    false,
		},
		{
			name:     "no digit",
			password: "TestPassword!",
			valid:    false,
		},
		{
			name:     "no special",
			password: "TestPass123",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password, policy)
			if tt.valid && err != nil {
				t.Errorf("Expected valid password, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("Expected invalid password, got no error")
			}
		})
	}
}

func TestGenerateRandomToken(t *testing.T) {
	token1, err := GenerateRandomToken(32)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	token2, err := GenerateRandomToken(32)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token1 == token2 {
		t.Error("Generated tokens should be unique")
	}

	if len(token1) == 0 || len(token2) == 0 {
		t.Error("Generated tokens should not be empty")
	}
}

func TestGeneratePKCE(t *testing.T) {
	pkce, err := GeneratePKCE(128)
	if err != nil {
		t.Fatalf("Failed to generate PKCE: %v", err)
	}

	if pkce.CodeChallenge == "" {
		t.Error("Code challenge should not be empty")
	}

	if pkce.CodeChallengeMethod != "plain" {
		t.Errorf("Expected method 'plain', got '%s'", pkce.CodeChallengeMethod)
	}

	if pkce.CodeVerifier == "" {
		t.Error("Code verifier should not be empty")
	}
}
