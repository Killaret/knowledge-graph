package main

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"knowledge-graph/internal/config"
	drafthandler "knowledge-graph/internal/interfaces/api/handlers/draft"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// serviceRoutes are registered on the router but are not part of the API
// contract: /health is a probe, /swagger/*any and /openapi.yaml serve the
// documentation itself. The list is deliberately short — every API route
// must be described in openAPI.yaml.
var serviceRoutes = map[string]bool{
	"GET /openapi.yaml":  true, // serves the specification file itself
	"HEAD /openapi.yaml": true, // StaticFile registers HEAD alongside GET
	"GET /swagger/*any":  true, // the Swagger UI bundle
}

var ginParamPattern = regexp.MustCompile(`[:*]([A-Za-z0-9_]+)`)

func ginPathToSpec(path string) string {
	return ginParamPattern.ReplaceAllString(path, "{$1}")
}

// TestRouterMatchesOpenAPISpec guards the hand-written contract: every API
// route registered on the real router must be described in openAPI.yaml, and
// every operation in openAPI.yaml must exist on the router. Both directions
// are checked — a route without a description and a description without a
// route are the same defect.
func TestRouterMatchesOpenAPISpec(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{}
	jwtConfig := middleware.DefaultJWTConfig(nil, nil)
	apiKeyConfig := middleware.DefaultAPIKeyConfig(nil, false, "")
	skipAuthConfig := middleware.DefaultSkipAuthConfig(false)

	// The router is built in full kit: every handler whose presence changes
	// the route set is passed non-nil (the draft handler registers its routes
	// conditionally — a nil there would report drift that does not exist).
	r := setupRouter(
		nil, nil, nil, nil, nil,
		nil, nil, nil, nil, nil, &drafthandler.Handler{},
		cfg,
		newHealthHandler(nil, nil, nil),
		newWriteLimiter(cfg),
		jwtConfig,
		apiKeyConfig,
		skipAuthConfig,
		nil,
	)

	routerOps := make(map[string]bool)
	for _, route := range r.Routes() {
		key := route.Method + " " + route.Path
		if serviceRoutes[key] {
			continue
		}
		routerOps[route.Method+" "+ginPathToSpec(route.Path)] = true
	}

	// A test that finds no routes guards nothing.
	require.GreaterOrEqual(t, len(routerOps), 70,
		"expected the full API route set; got %d — did the router lose routes or the exclusion list grow?", len(routerOps))

	specRaw, err := os.ReadFile("../../openAPI.yaml")
	require.NoError(t, err, "openAPI.yaml must be readable from backend/cmd/server")

	var spec struct {
		Paths map[string]map[string]yaml.Node `yaml:"paths"`
	}
	require.NoError(t, yaml.Unmarshal(specRaw, &spec))
	require.NotEmpty(t, spec.Paths, "openAPI.yaml parsed but has no paths")

	specOps := make(map[string]bool)
	for path, item := range spec.Paths {
		for method := range item {
			specOps[strings.ToUpper(method)+" "+path] = true
		}
	}

	var missing []string
	for op := range routerOps {
		if !specOps[op] {
			missing = append(missing, op+" — in the router, not described in openAPI.yaml")
		}
	}
	for op := range specOps {
		if !routerOps[op] {
			missing = append(missing, op+" — described in openAPI.yaml, not in the router")
		}
	}
	sort.Strings(missing)
	require.Empty(t, missing, "the API contract drifted from the code:\n  %s", strings.Join(missing, "\n  "))
}
