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

		cfg := Load()

		if cfg.RecommendationDepth != 42 {
			t.Errorf("Expected RecommendationDepth=42 from env, got %d", cfg.RecommendationDepth)
		}
	})

	// Test 2: Float values - env var should override JSON
	t.Run("FloatEnvPriority", func(t *testing.T) {
		os.Setenv("RECOMMENDATION_ALPHA", "0.75")
		defer os.Unsetenv("RECOMMENDATION_ALPHA")

		cfg := Load()

		if cfg.RecommendationAlpha != 0.75 {
			t.Errorf("Expected RecommendationAlpha=0.75 from env, got %f", cfg.RecommendationAlpha)
		}
	})

	// Test 3: Duration from integer seconds - env var should override JSON
	t.Run("DurationEnvPriority", func(t *testing.T) {
		os.Setenv("RECOMMENDATION_CACHE_TTL_SECONDS", "600")
		defer os.Unsetenv("RECOMMENDATION_CACHE_TTL_SECONDS")

		cfg := Load()
		expected := 600 * time.Second

		if cfg.RecommendationCacheTTL != expected {
			t.Errorf("Expected RecommendationCacheTTL=%v from env, got %v", expected, cfg.RecommendationCacheTTL)
		}
	})

	// Test 4: Boolean values - env var should override JSON
	t.Run("BoolEnvPriority", func(t *testing.T) {
		os.Setenv("SERVER_RATE_LIMIT_ENABLED", "false")
		defer os.Unsetenv("SERVER_RATE_LIMIT_ENABLED")

		cfg := Load()

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

		cfg := Load()

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

		cfg := Load()

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

	cfg := Load()

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

	cfg := Load()

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

	cfg := Load()

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

	cfg := Load()

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
