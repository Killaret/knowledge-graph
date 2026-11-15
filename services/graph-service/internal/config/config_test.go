package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CONFIG-AUDIT-1: documented precedence is env > knowledge-graph.config.json >
// built-in default. Each test pins one step of that chain.

func TestFullLimitEnvOverridesJSONFile(t *testing.T) {
	// The repo config sets graph_service.full_limit=500; env must win.
	t.Setenv("GRAPH_FULL_LIMIT", "777")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 777, cfg.FullLimit)
}

func TestFullLimitFromJSONWhenNoEnv(t *testing.T) {
	if _, set := os.LookupEnv("GRAPH_FULL_LIMIT"); set {
		t.Skip("GRAPH_FULL_LIMIT set in the environment")
	}
	// loadJSONConfig resolves the repo file via runtime.Caller; if that fails
	// the built-in default is 500 too, so assert against the JSON value.
	jsonCfg := loadJSONConfig()
	require.NotNil(t, jsonCfg, "knowledge-graph.config.json must load in tests")
	require.Equal(t, 500, jsonCfg.GraphService.FullLimit, "repo config must pin full_limit")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, jsonCfg.GraphService.FullLimit, cfg.FullLimit)
}

func TestLimitHelperPrecedence(t *testing.T) {
	t.Setenv("CONFIG_AUDIT_TEST_LIMIT", "42")

	// env > JSON getter > built-in default
	got := getIntEnv("CONFIG_AUDIT_TEST_LIMIT", getJSONInt(&JSONConfig{}, func(j *JSONConfig) int { return 7 }, 3))
	assert.Equal(t, 42, got)

	// no env → JSON value; nil JSON → default
	os.Unsetenv("CONFIG_AUDIT_TEST_LIMIT")
	jsonCfg := &JSONConfig{}
	jsonCfg.GraphService.FullLimit = 7
	assert.Equal(t, 7, getIntEnv("CONFIG_AUDIT_TEST_LIMIT", getJSONInt(jsonCfg, func(j *JSONConfig) int { return j.GraphService.FullLimit }, 3)))
	assert.Equal(t, 3, getIntEnv("CONFIG_AUDIT_TEST_LIMIT", getJSONInt(nil, func(j *JSONConfig) int { return j.GraphService.FullLimit }, 3)))
}
