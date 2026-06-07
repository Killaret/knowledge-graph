package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	authpkg "knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	"knowledge-graph/internal/interfaces/api/middleware"
)

// corsMiddleware returns CORS middleware that allows requests from frontend
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With, X-Backend-Url")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Vary", "Origin")

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
func newJWTConfig(jwtManager *authpkg.JWTManager, tokenStore *authpkg.RedisTokenStore) *middleware.JWTConfig {
	return middleware.DefaultJWTConfig(jwtManager, tokenStore)
}

// newAPIKeyConfig creates API key middleware configuration
func newAPIKeyConfig(db *gorm.DB, enabled bool, staticKey string) *middleware.APIKeyConfig {
	return middleware.DefaultAPIKeyConfig(db, enabled, staticKey)
}
