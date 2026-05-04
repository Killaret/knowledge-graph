package config

import (
	"os"
	"testing"
	"time"
)

// TestEnvVarPriority tests that environment variables take priority over JSON config
func TestEnvVarPriority(t *testing.T) {
	// Save original env vars and restore after test
	originalEnv := make(map[string]string)
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	// Store original values
	envVars := []string{
		"DATABASE_URL",
		"RECOMMENDATION_DEPTH",
		"RECOMMENDATION_ALPHA",
		"RECOMMENDATION_CACHE_TTL_SECONDS",
		"GRAPH_LOAD_DEPTH",
		"PAGINATION_DEFAULT_LIMIT",
		"PAGINATION_MAX_LIMIT",
		"SERVER_RATE_LIMIT_ENABLED",
		"SERVER_RATE_LIMIT_REQUESTS",
	}

	for _, v := range envVars {
		originalEnv[v] = os.Getenv(v)
	}

	// Set required DATABASE_URL
	os.Setenv("DATABASE_URL", "postgres://test@localhost/test")

	// Test 1: Integer values - env var should override JSON
	t.Run("IntegerEnvPriority", func(t *testing.T) {
		os.Setenv("RECOMMENDATION_DEPTH", "42")
		defer os.Unsetenv("RECOMMENDATION_DEPTH")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.RecommendationDepth != 42 {
			t.Errorf("Expected RecommendationDepth=42 from env, got %d", cfg.RecommendationDepth)
		}
	})

	// Test 2: Float values - env var should override JSON
	t.Run("FloatEnvPriority", func(t *testing.T) {
		os.Setenv("RECOMMENDATION_ALPHA", "0.75")
		defer os.Unsetenv("RECOMMENDATION_ALPHA")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.RecommendationAlpha != 0.75 {
			t.Errorf("Expected RecommendationAlpha=0.75 from env, got %f", cfg.RecommendationAlpha)
		}
	})

	// Test 3: Duration from integer seconds - env var should override JSON
	t.Run("DurationEnvPriority", func(t *testing.T) {
		os.Setenv("RECOMMENDATION_CACHE_TTL_SECONDS", "600")
		defer os.Unsetenv("RECOMMENDATION_CACHE_TTL_SECONDS")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}
		expected := 600 * time.Second

		if cfg.RecommendationCacheTTL != expected {
			t.Errorf("Expected RecommendationCacheTTL=%v from env, got %v", expected, cfg.RecommendationCacheTTL)
		}
	})

	// Test 4: Boolean values - env var should override JSON
	t.Run("BoolEnvPriority", func(t *testing.T) {
		os.Setenv("SERVER_RATE_LIMIT_ENABLED", "false")
		defer os.Unsetenv("SERVER_RATE_LIMIT_ENABLED")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.ServerRateLimitEnabled != false {
			t.Errorf("Expected ServerRateLimitEnabled=false from env, got %v", cfg.ServerRateLimitEnabled)
		}
	})

	// Test 5: Multiple env vars take priority over JSON
	t.Run("MultipleEnvVarsPriority", func(t *testing.T) {
		os.Setenv("GRAPH_LOAD_DEPTH", "5")
		os.Setenv("PAGINATION_DEFAULT_LIMIT", "50")
		os.Setenv("PAGINATION_MAX_LIMIT", "200")
		defer os.Unsetenv("GRAPH_LOAD_DEPTH")
		defer os.Unsetenv("PAGINATION_DEFAULT_LIMIT")
		defer os.Unsetenv("PAGINATION_MAX_LIMIT")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.GraphLoadDepth != 5 {
			t.Errorf("Expected GraphLoadDepth=5 from env, got %d", cfg.GraphLoadDepth)
		}
		if cfg.PaginationDefaultLimit != 50 {
			t.Errorf("Expected PaginationDefaultLimit=50 from env, got %d", cfg.PaginationDefaultLimit)
		}
		if cfg.PaginationMaxLimit != 200 {
			t.Errorf("Expected PaginationMaxLimit=200 from env, got %d", cfg.PaginationMaxLimit)
		}
	})

	// Test 6: Without env var, JSON value is used (or default if no JSON)
	t.Run("JsonValueWhenNoEnv", func(t *testing.T) {
		// Clear env vars to test JSON fallback
		os.Unsetenv("RECOMMENDATION_DEPTH")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Should have a positive value from JSON or default
		if cfg.RecommendationDepth <= 0 {
			t.Errorf("Expected positive RecommendationDepth from JSON/default, got %d", cfg.RecommendationDepth)
		}
	})
}

// TestConfigValuesArePositive tests that loaded config values are valid
func TestConfigValuesArePositive(t *testing.T) {
	// Set required DATABASE_URL
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
		defer os.Unsetenv("DATABASE_URL")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name  string
		value int
	}{
		{"RecommendationDepth", cfg.RecommendationDepth},
		{"GraphLoadDepth", cfg.GraphLoadDepth},
		{"PaginationDefaultLimit", cfg.PaginationDefaultLimit},
		{"PaginationMaxLimit", cfg.PaginationMaxLimit},
		{"ServerRateLimitRequests", cfg.ServerRateLimitRequests},
		{"ServerRateLimitWindowSeconds", cfg.ServerRateLimitWindowSeconds},
		{"DatabaseRetryMaxAttempts", cfg.DatabaseRetryMaxAttempts},
		{"DatabaseRetryDelaySeconds", cfg.DatabaseRetryDelaySeconds},
		{"EmbeddingSimilarityLimit", cfg.EmbeddingSimilarityLimit},
		{"GraphDefaultLimit", cfg.GraphDefaultLimit},
		{"GraphMaxLimit", cfg.GraphMaxLimit},
		{"GraphLinkDefaultLimit", cfg.GraphLinkDefaultLimit},
		{"GraphLinkMaxLimit", cfg.GraphLinkMaxLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value <= 0 {
				t.Errorf("%s must be > 0, got %d", tt.name, tt.value)
			}
		})
	}
}

// TestConfigLoadNeverNil ensures Load() never returns nil
func TestConfigLoadNeverNil(t *testing.T) {
	// Set required DATABASE_URL
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
		defer os.Unsetenv("DATABASE_URL")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil, expected valid Config")
	}
}

// TestPaginationLimitConsistency ensures MaxLimit >= DefaultLimit
func TestPaginationLimitConsistency(t *testing.T) {
	// Set required DATABASE_URL
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
		defer os.Unsetenv("DATABASE_URL")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.PaginationMaxLimit < cfg.PaginationDefaultLimit {
		t.Errorf("PaginationMaxLimit (%d) must be >= PaginationDefaultLimit (%d)",
			cfg.PaginationMaxLimit, cfg.PaginationDefaultLimit)
	}
}

// TestRecommendationWeightsRange ensures alpha, beta, gamma are in [0,1]
func TestRecommendationWeightsRange(t *testing.T) {
	// Set required DATABASE_URL
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
		defer os.Unsetenv("DATABASE_URL")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	weights := []struct {
		name  string
		value float64
	}{
		{"Alpha", cfg.RecommendationAlpha},
		{"Beta", cfg.RecommendationBeta},
		{"Gamma", cfg.RecommendationGamma},
	}

	for _, w := range weights {
		t.Run(w.name, func(t *testing.T) {
			if w.value < 0 || w.value > 1 {
				t.Errorf("Recommendation%s must be in [0,1], got %f", w.name, w.value)
			}
		})
	}
}

// TestAuthConfigComplete ensures all auth configuration fields are present and valid
func TestAuthConfigComplete(t *testing.T) {
	// Set required DATABASE_URL
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
		defer os.Unsetenv("DATABASE_URL")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test 1: JWT configuration
	t.Run("JWTConfiguration", func(t *testing.T) {
		if cfg.JWTSecret == "" {
			t.Error("JWTSecret is empty")
		}
		if cfg.JWTAccessTTL <= 0 {
			t.Errorf("JWTAccessTTL must be > 0, got %v", cfg.JWTAccessTTL)
		}
		if cfg.JWTRefreshTTL <= 0 {
			t.Errorf("JWTRefreshTTL must be > 0, got %v", cfg.JWTRefreshTTL)
		}
	})

	// Test 2: Argon2 configuration
	t.Run("Argon2Configuration", func(t *testing.T) {
		if cfg.Argon2Time == 0 {
			t.Error("Argon2Time is not set")
		}
		if cfg.Argon2Memory == 0 {
			t.Error("Argon2Memory is not set")
		}
		if cfg.Argon2Threads == 0 {
			t.Error("Argon2Threads is not set")
		}
	})

	// Test 3: Password policy configuration
	t.Run("PasswordPolicyConfiguration", func(t *testing.T) {
		if cfg.PasswordPolicyMinLength < 8 {
			t.Errorf("PasswordPolicyMinLength should be at least 8, got %d", cfg.PasswordPolicyMinLength)
		}
	})

	// Test 4: Password reset TTL
	t.Run("PasswordResetTTL", func(t *testing.T) {
		if cfg.PasswordResetTTL <= 0 {
			t.Errorf("PasswordResetTTL must be > 0, got %v", cfg.PasswordResetTTL)
		}
	})
}

// TestAuthEnvVars ensures auth-related env vars are read correctly
func TestAuthEnvVars(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	authEnvVars := []string{
		"JWT_SECRET",
		"JWT_ACCESS_TTL_SECONDS",
		"JWT_REFRESH_TTL_SECONDS",
		"ARGON2_TIME",
		"ARGON2_MEMORY",
		"ARGON2_THREADS",
		"API_KEY_ENABLED",
		"PASSWORD_RESET_TTL_SECONDS",
		"PASSWORD_POLICY_MIN_LENGTH",
	}

	for _, v := range authEnvVars {
		originalEnv[v] = os.Getenv(v)
	}

	// Set required DATABASE_URL
	os.Setenv("DATABASE_URL", "postgres://test@localhost/test")

	// Test 1: JWT TTL from env
	t.Run("JWTAccessTTLFromEnv", func(t *testing.T) {
		os.Setenv("JWT_ACCESS_TTL_SECONDS", "1800")
		defer os.Unsetenv("JWT_ACCESS_TTL_SECONDS")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		expected := 1800 * time.Second
		if cfg.JWTAccessTTL != expected {
			t.Errorf("Expected JWTAccessTTL=%v from env, got %v", expected, cfg.JWTAccessTTL)
		}
	})

	// Test 2: Argon2 params from env
	t.Run("Argon2ParamsFromEnv", func(t *testing.T) {
		os.Setenv("ARGON2_TIME", "3")
		os.Setenv("ARGON2_MEMORY", "65536")
		os.Setenv("ARGON2_THREADS", "4")
		defer os.Unsetenv("ARGON2_TIME")
		defer os.Unsetenv("ARGON2_MEMORY")
		defer os.Unsetenv("ARGON2_THREADS")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.Argon2Time != 3 {
			t.Errorf("Expected Argon2Time=3 from env, got %d", cfg.Argon2Time)
		}
		if cfg.Argon2Memory != 65536 {
			t.Errorf("Expected Argon2Memory=65536 from env, got %d", cfg.Argon2Memory)
		}
		if cfg.Argon2Threads != 4 {
			t.Errorf("Expected Argon2Threads=4 from env, got %d", cfg.Argon2Threads)
		}
	})

	// Test 3: API Key enabled from env
	t.Run("APIKeyEnabledFromEnv", func(t *testing.T) {
		os.Setenv("API_KEY_ENABLED", "true")
		defer os.Unsetenv("API_KEY_ENABLED")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if !cfg.APIKeyEnabled {
			t.Error("Expected APIKeyEnabled=true from env")
		}
	})

	// Test 4: Password policy from env
	t.Run("PasswordPolicyFromEnv", func(t *testing.T) {
		os.Setenv("PASSWORD_POLICY_MIN_LENGTH", "12")
		defer os.Unsetenv("PASSWORD_POLICY_MIN_LENGTH")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if cfg.PasswordPolicyMinLength != 12 {
			t.Errorf("Expected PasswordPolicyMinLength=12 from env, got %d", cfg.PasswordPolicyMinLength)
		}
	})
}

// TestSecurityConfig validates security-related configuration
func TestSecurityConfig(t *testing.T) {
	// Set required DATABASE_URL
	if os.Getenv("DATABASE_URL") == "" {
		os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
		defer os.Unsetenv("DATABASE_URL")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test JWT TTL values (should be reasonable)
	t.Run("JWTAccessTTLSecurity", func(t *testing.T) {
		// Access token should be short-lived (15 min to 24 hours)
		if cfg.JWTAccessTTL < 15*time.Minute {
			t.Errorf("JWTAccessTTL too short for security: %v", cfg.JWTAccessTTL)
		}
		if cfg.JWTAccessTTL > 24*time.Hour {
			t.Errorf("JWTAccessTTL too long for security: %v", cfg.JWTAccessTTL)
		}
	})

	t.Run("JWTRefreshTTLSecurity", func(t *testing.T) {
		// Refresh token should be longer (7 to 90 days)
		if cfg.JWTRefreshTTL < 7*24*time.Hour {
			t.Errorf("JWTRefreshTTL too short: %v", cfg.JWTRefreshTTL)
		}
		if cfg.JWTRefreshTTL > 90*24*time.Hour {
			t.Errorf("JWTRefreshTTL too long: %v", cfg.JWTRefreshTTL)
		}
	})

	// Test Argon2 parameters (should be secure)
	t.Run("Argon2Security", func(t *testing.T) {
		// Memory should be at least 64MB
		minMemory := uint32(64 * 1024)
		if cfg.Argon2Memory < minMemory {
			t.Errorf("Argon2Memory too low for security: %d, minimum: %d", cfg.Argon2Memory, minMemory)
		}
		// Time should be at least 1 iteration
		if cfg.Argon2Time < 1 {
			t.Errorf("Argon2Time too low: %d", cfg.Argon2Time)
		}
		// Threads should be reasonable
		if cfg.Argon2Threads < 1 || cfg.Argon2Threads > 16 {
			t.Errorf("Argon2Threads out of reasonable range: %d", cfg.Argon2Threads)
		}
	})
}
