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

	healthHandler := newHealthHandler(nil, nil, nil)
	r := setupRouter(
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, nil,
		cfg,
		healthHandler,
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

	// Regression check: user profile and API key management routes must be
	// wired up (previously only GET /users/me was registered, silently
	// breaking profile updates, account deletion, and API key management).
	methodPaths := make(map[string]bool)
	for _, route := range routes {
		methodPaths[route.Method+" "+route.Path] = true
	}

	for _, mp := range []string{
		"GET /api/v1/users/me",
		"PUT /api/v1/users/me",
		"DELETE /api/v1/users/me",
		"GET /api/v1/users/me/api-keys",
		"POST /api/v1/users/me/api-keys",
		"DELETE /api/v1/users/me/api-keys/:id",
	} {
		assert.True(t, methodPaths[mp], "expected route to be registered: %s", mp)
	}
}
