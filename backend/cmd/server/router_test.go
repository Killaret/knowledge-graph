package main

import (
	"testing"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupRouter(t *testing.T) {
	cfg := &config.Config{}

	jwtConfig := middleware.DefaultJWTConfig(nil, nil)
	apiKeyConfig := middleware.DefaultAPIKeyConfig(nil, false, "")
	skipAuthConfig := middleware.DefaultSkipAuthConfig(false)

	r := setupRouter(
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil,
		cfg, nil, nil,
		newWriteLimiter(cfg),
		jwtConfig,
		apiKeyConfig,
		skipAuthConfig,
	)

	require.NotNil(t, r)
	routes := r.Routes()
	assert.NotEmpty(t, routes)

	// Ensure expected groups/paths are registered
	paths := make(map[string]bool)
	for _, route := range routes {
		paths[route.Path] = true
	}

	assert.True(t, paths["/health"])
	assert.True(t, paths["/swagger/*any"])
	assert.True(t, paths["/api/v1/notes"])
	assert.True(t, paths["/api/v1/tags"])
	assert.True(t, paths["/api/v1/links"])
}
