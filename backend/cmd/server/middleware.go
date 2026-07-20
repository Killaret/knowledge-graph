package main

import (
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	authpkg "knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	domainuser "knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/interfaces/api/middleware"
)

// corsMiddleware returns CORS middleware that allows requests from whitelisted origins.
// In development, localhost origins are allowed. In production, only configured origins are allowed.
func corsMiddleware() gin.HandlerFunc {
	// Default localhost origins for development
	defaultOrigins := map[string]bool{
		"http://localhost:3000": true,
		"http://localhost:3001": true,
		"http://localhost:5173": true,
		"http://localhost:8080": true,
		"http://localhost:8081": true,
		"http://localhost:8082": true,
		"http://localhost:8083": true,
		"http://127.0.0.1:3000": true,
		"http://127.0.0.1:3001": true,
		"http://127.0.0.1:5173": true,
		"http://127.0.0.1:8080": true,
		"http://127.0.0.1:8082": true,
	}

	// Read CORS_ALLOWED_ORIGINS from environment variable
	corsOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	allowedOrigins := defaultOrigins

	if corsOriginsEnv != "" {
		// Parse comma-separated origins from environment variable
		origins := strings.Split(corsOriginsEnv, ",")
		allowedOrigins = make(map[string]bool)
		for _, origin := range origins {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins[origin] = true
			}
		}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Only set CORS headers if origin is in whitelist
		if origin != "" && allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With, X-Backend-Url")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
			c.Writer.Header().Set("Vary", "Origin")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// newWriteLimiter creates rate limiter for write operations based on config
func newWriteLimiter(cfg *config.Config) gin.HandlerFunc {
	if !cfg.ServerRateLimitEnabled {
		return func(c *gin.Context) { c.Next() }
	}

	rateWindow := time.Duration(cfg.ServerRateLimitWindowSeconds) * time.Second
	endpointLimits := map[string]int{
		"/notes":     cfg.ServerRateLimitEndpoints["notes_create"],
		"/links":     cfg.ServerRateLimitEndpoints["links_create"],
		"/notes/:id": cfg.ServerRateLimitEndpoints["notes_update"],
	}
	return middleware.RateLimitByEndpoint(endpointLimits, cfg.ServerRateLimitRequests, rateWindow)
}

// newJWTConfig creates JWT middleware configuration
func newJWTConfig(jwtManager *authpkg.JWTManager, tokenStore authpkg.TokenStore) *middleware.JWTConfig {
	return middleware.DefaultJWTConfig(jwtManager, tokenStore)
}

// newAPIKeyConfig creates API key middleware configuration
func newAPIKeyConfig(repo domainuser.APIKeyRepository, enabled bool, staticKey string) *middleware.APIKeyConfig {
	return middleware.DefaultAPIKeyConfig(repo, enabled, staticKey)
}
