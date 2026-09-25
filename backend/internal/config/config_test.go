package config

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-jwt-secret")
	}
}

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

func TestConfigHelpers(t *testing.T) {
	t.Run("getEnv returns env value", func(t *testing.T) {
		t.Setenv("TEST_CONFIG_KEY", "value")
		if got := getEnv("TEST_CONFIG_KEY", "default"); got != "value" {
			t.Errorf("expected 'value', got %q", got)
		}
	})

	t.Run("getEnv returns default", func(t *testing.T) {
		t.Setenv("TEST_CONFIG_KEY_UNKNOWN", "")
		if got := getEnv("TEST_CONFIG_KEY_UNKNOWN_XYZ", "default"); got != "default" {
			t.Errorf("expected 'default', got %q", got)
		}
	})

	t.Run("mustGetEnv returns value", func(t *testing.T) {
		t.Setenv("TEST_MUST_KEY", "value")
		val, err := mustGetEnv("TEST_MUST_KEY")
		if err != nil || val != "value" {
			t.Errorf("expected 'value', got %q, err %v", val, err)
		}
	})

	t.Run("mustGetEnv returns error when missing", func(t *testing.T) {
		t.Setenv("TEST_MUST_KEY_MISSING", "")
		if _, err := mustGetEnv("TEST_MUST_KEY_MISSING_XYZ"); err == nil {
			t.Error("expected error for missing env var")
		}
	})

	t.Run("getIntEnv parses integer", func(t *testing.T) {
		t.Setenv("TEST_INT_KEY", "42")
		if got := getIntEnv("TEST_INT_KEY", 0); got != 42 {
			t.Errorf("expected 42, got %d", got)
		}
	})

	t.Run("getIntEnv returns default on invalid", func(t *testing.T) {
		t.Setenv("TEST_INT_KEY_INVALID", "not-a-number")
		if got := getIntEnv("TEST_INT_KEY_INVALID", 7); got != 7 {
			t.Errorf("expected default 7, got %d", got)
		}
	})

	t.Run("getFloatEnv parses float", func(t *testing.T) {
		t.Setenv("TEST_FLOAT_KEY", "3.14")
		if got := getFloatEnv("TEST_FLOAT_KEY", 0); got != 3.14 {
			t.Errorf("expected 3.14, got %f", got)
		}
	})

	t.Run("getBoolEnv parses bool", func(t *testing.T) {
		t.Setenv("TEST_BOOL_KEY", "true")
		if got := getBoolEnv("TEST_BOOL_KEY", false); !got {
			t.Error("expected true")
		}
	})

	t.Run("getUint32Env parses uint32", func(t *testing.T) {
		t.Setenv("TEST_UINT32_KEY", "100")
		if got := getUint32Env("TEST_UINT32_KEY", 0); got != 100 {
			t.Errorf("expected 100, got %d", got)
		}
	})

	t.Run("getUint8Env parses uint8", func(t *testing.T) {
		t.Setenv("TEST_UINT8_KEY", "8")
		if got := getUint8Env("TEST_UINT8_KEY", 0); got != 8 {
			t.Errorf("expected 8, got %d", got)
		}
	})
}

func TestMergeJSONObjects(t *testing.T) {
	dst := map[string]any{
		"backend": map[string]any{
			"server": map[string]any{
				"port": "8080",
			},
		},
	}
	src := map[string]any{
		"backend": map[string]any{
			"server": map[string]any{
				"port": "9090",
				"host": "127.0.0.1",
			},
			"new": true,
		},
	}

	merged := mergeJSONObjects(dst, src)
	backend := merged["backend"].(map[string]any)
	server := backend["server"].(map[string]any)

	if server["port"] != "9090" {
		t.Errorf("expected port 9090, got %v", server["port"])
	}
	if server["host"] != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %v", server["host"])
	}
	if backend["new"] != true {
		t.Error("expected new key")
	}
}

func TestMustJSON(t *testing.T) {
	data := map[string]any{"key": "value"}
	jsonStr := mustJSON(data)
	if jsonStr != `{"key":"value"}` {
		t.Errorf("unexpected json: %s", jsonStr)
	}
}

func TestSkipAuthRequiresTestEnv(t *testing.T) {
	// Set required variables for config.Load
	t.Setenv("DATABASE_URL", "postgres://test@localhost/test")
	t.Setenv("JWT_SECRET", "test-jwt-secret")

	tests := []struct {
		name     string
		appEnv   string
		skipAuth string
		wantErr  bool
	}{
		{"skip auth in dev", "development", "true", true},
		{"skip auth in production", "production", "true", true},
		{"skip auth in test", "test", "true", false},
		{"no skip auth in dev", "development", "false", false},
		{"empty env defaults to development", "", "false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tt.appEnv)
			t.Setenv("SKIP_AUTH", tt.skipAuth)

			cfg, err := Load()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "SKIP_AUTH")
				assert.Contains(t, err.Error(), "APP_ENV")
			} else {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				if tt.appEnv == "" {
					assert.Equal(t, "development", cfg.AppEnv)
				} else {
					assert.Equal(t, tt.appEnv, cfg.AppEnv)
				}
			}
		})
	}
}

func TestJSONDefaultHelpers(t *testing.T) {
	var cfg JSONConfig
	cfg.Backend.Recommendation.Depth = 5
	cfg.Backend.Recommendation.Alpha = 0.5

	if got := getJSONIntOrDefault(&cfg, func(j *JSONConfig) int { return j.Backend.Recommendation.Depth }, 1); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
	if got := getJSONIntOrDefault(nil, nil, 3); got != 3 {
		t.Errorf("expected default 3, got %d", got)
	}
	if got := getJSONFloatOrDefault(&cfg, func(j *JSONConfig) float64 { return j.Backend.Recommendation.Alpha }, 0.1); got != 0.5 {
		t.Errorf("expected 0.5, got %f", got)
	}
	if got := getJSONStringOrDefault(nil, nil, "default"); got != "default" {
		t.Errorf("expected default string, got %q", got)
	}
}

// DB-POOL-1: the env knob must reach the loaded config — POSTGRES_MAX_OPEN_CONNS=7
// has to land in cfg.DatabasePoolMaxOpenConns.
func TestDatabasePoolEnvVars(t *testing.T) {
	vars := []string{
		"DATABASE_URL",
		"POSTGRES_MAX_OPEN_CONNS",
		"POSTGRES_MAX_IDLE_CONNS",
		"POSTGRES_CONN_MAX_LIFETIME_SECONDS",
		"POSTGRES_CONN_MAX_IDLE_TIME_SECONDS",
	}
	original := make(map[string]string)
	for _, v := range vars {
		original[v] = os.Getenv(v)
	}
	defer func() {
		for k, v := range original {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
	os.Setenv("POSTGRES_MAX_OPEN_CONNS", "7")
	os.Setenv("POSTGRES_MAX_IDLE_CONNS", "3")
	os.Setenv("POSTGRES_CONN_MAX_LIFETIME_SECONDS", "111")
	os.Setenv("POSTGRES_CONN_MAX_IDLE_TIME_SECONDS", "22")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg.DatabasePoolMaxOpenConns != 7 {
		t.Errorf("expected DatabasePoolMaxOpenConns 7, got %d", cfg.DatabasePoolMaxOpenConns)
	}
	if cfg.DatabasePoolMaxIdleConns != 3 {
		t.Errorf("expected DatabasePoolMaxIdleConns 3, got %d", cfg.DatabasePoolMaxIdleConns)
	}
	if cfg.DatabasePoolConnMaxLifetimeSeconds != 111 {
		t.Errorf("expected lifetime 111, got %d", cfg.DatabasePoolConnMaxLifetimeSeconds)
	}
	if cfg.DatabasePoolConnMaxIdleTimeSeconds != 22 {
		t.Errorf("expected idle time 22, got %d", cfg.DatabasePoolConnMaxIdleTimeSeconds)
	}
}

// TestGammaLinkMinScore ensures the LINKS-1 threshold is wired through env,
// JSON config and the Go default of 0.6.
func TestGammaLinkMinScore(t *testing.T) {
	vars := []string{"DATABASE_URL", "GAMMA_LINK_MIN_SCORE"}
	original := make(map[string]string)
	for _, v := range vars {
		original[v] = os.Getenv(v)
	}
	defer func() {
		for k, v := range original {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
	os.Unsetenv("GAMMA_LINK_MIN_SCORE")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg.GammaLinkMinScore != 0.6 {
		t.Errorf("expected default GammaLinkMinScore 0.6, got %f", cfg.GammaLinkMinScore)
	}

	os.Setenv("GAMMA_LINK_MIN_SCORE", "0.55")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg.GammaLinkMinScore != 0.55 {
		t.Errorf("expected env override GammaLinkMinScore 0.55, got %f", cfg.GammaLinkMinScore)
	}
}

// A loaded JSON config without the gamma_link_min_score key must resolve to
// the 0.6 default — via the seeded struct for a real load, and via the <=0
// guard for an explicit non-positive value.
func TestGammaLinkMinScore_MissingJSONKey(t *testing.T) {
	original := os.Getenv("GAMMA_LINK_MIN_SCORE")
	defer func() {
		if original == "" {
			os.Unsetenv("GAMMA_LINK_MIN_SCORE")
		} else {
			os.Setenv("GAMMA_LINK_MIN_SCORE", original)
		}
	}()
	os.Unsetenv("GAMMA_LINK_MIN_SCORE")

	// CONFIG-1 path: seeded struct + file that omits the key -> default kept.
	jsonCfg := defaultJSONConfig()
	require.NoError(t, json.Unmarshal([]byte(`{"backend":{"pagination":{"default_limit":50}}}`), jsonCfg))
	if v := resolveGammaLinkMinScore(jsonCfg); v != 0.6 {
		t.Errorf("expected fallback 0.6 for missing JSON key, got %f", v)
	}
	assert.Equal(t, 50, jsonCfg.Backend.Pagination.DefaultLimit, "present key must override the seed")
	assert.Equal(t, 100, jsonCfg.Backend.Pagination.MaxLimit, "absent sibling key must keep its default")

	// Explicit non-positive still clamps through the guard.
	jsonCfg.Backend.Recommendation.GammaLinkMinScore = 0
	if v := resolveGammaLinkMinScore(jsonCfg); v != 0.6 {
		t.Errorf("expected fallback 0.6 for explicit 0, got %f", v)
	}

	// An explicit positive value still wins.
	jsonCfg.Backend.Recommendation.GammaLinkMinScore = 0.7
	if v := resolveGammaLinkMinScore(jsonCfg); v != 0.7 {
		t.Errorf("expected JSON value 0.7, got %f", v)
	}
}

// NLP-4: pipeline switch defaults to off, history to on, cosine floor to
// 0.7; env overrides JSON; out-of-range cosine falls back to the default.
func TestNLP4Config(t *testing.T) {
	vars := []string{
		"DATABASE_URL", "NLP_PIPELINE_ENABLED", "NLP_HISTORY_ENABLED",
		"NLP_NORMALIZATION_MIN_COSINE",
	}
	original := make(map[string]string)
	for _, v := range vars {
		original[v] = os.Getenv(v)
	}
	defer func() {
		for k, v := range original {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	os.Setenv("DATABASE_URL", "postgres://test@localhost/test")
	for _, v := range vars[1:] {
		os.Unsetenv(v)
	}

	cfg, err := Load()
	require.NoError(t, err)
	assert.False(t, cfg.NLPPipelineEnabled, "pipeline must default to off")
	assert.True(t, cfg.NLPHistoryEnabled, "history must default to on")
	assert.Equal(t, 0.7, cfg.NLPNormalizationMinCosine)

	os.Setenv("NLP_PIPELINE_ENABLED", "true")
	os.Setenv("NLP_HISTORY_ENABLED", "false")
	os.Setenv("NLP_NORMALIZATION_MIN_COSINE", "0.55")
	cfg, err = Load()
	require.NoError(t, err)
	assert.True(t, cfg.NLPPipelineEnabled)
	assert.False(t, cfg.NLPHistoryEnabled)
	assert.Equal(t, 0.55, cfg.NLPNormalizationMinCosine)

	os.Setenv("NLP_NORMALIZATION_MIN_COSINE", "1.5")
	cfg, err = Load()
	require.NoError(t, err)
	assert.Equal(t, 0.7, cfg.NLPNormalizationMinCosine, "out-of-range cosine must fall back")
}

// CONFIG-1 drift guard: the seeded struct must mirror every call-site
// default. resolveConfig(nil) resolves through the helpers' defaultValue
// arguments; resolveConfig(defaultJSONConfig()) resolves through the seed —
// any default added to one but not the other shows up as a diff here.
func TestDefaultJSONConfig_MatchesCallSiteDefaults(t *testing.T) {
	saved := os.Environ()
	os.Clearenv()
	t.Cleanup(func() {
		os.Clearenv()
		for _, kv := range saved {
			if k, v, ok := strings.Cut(kv, "="); ok {
				os.Setenv(k, v)
			}
		}
	})
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	t.Setenv("JWT_SECRET", "test-jwt-secret")

	withoutFile, err := resolveConfig(nil)
	require.NoError(t, err)
	withSeed, err := resolveConfig(defaultJSONConfig())
	require.NoError(t, err)

	assert.Equal(t, *withoutFile, *withSeed,
		"seeded JSONConfig must reproduce every built-in default — a missing field means defaultJSONConfig drifted from the call sites")
}

// CONFIG-1: explicit zero-values in the file must still apply — the fix is
// pre-seeding, not zero-detection, so `false`/`0`/`""` are real settings.
func TestDefaultJSONConfig_ExplicitValuesWin(t *testing.T) {
	jsonCfg := defaultJSONConfig()
	require.NoError(t, json.Unmarshal([]byte(`{
		"backend": {
			"search": {"fallback_to_ilike": false},
			"server": {"rate_limit": {"enabled": false, "endpoints": {"notes_create": 5}}}
		}
	}`), jsonCfg))

	assert.False(t, jsonCfg.Backend.Search.FallbackToILike, "explicit false must override the seeded true")
	assert.False(t, jsonCfg.Backend.Server.RateLimit.Enabled)
	// Partial maps merge onto the seeded defaults: unspecified entries keep
	// their default rates instead of disappearing.
	assert.Equal(t, 5, jsonCfg.Backend.Server.RateLimit.Endpoints["notes_create"])
	assert.Equal(t, 50, jsonCfg.Backend.Server.RateLimit.Endpoints["links_create"],
		"absent map entry must keep its seeded default")
}
