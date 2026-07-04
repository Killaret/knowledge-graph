package main

import (
	"time"

	"github.com/gin-gonic/gin"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/interfaces/api/middleware"
)

// corsMiddleware allows cross-origin requests from the frontend.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Origin, X-Requested-With")
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

// newWriteLimiter returns stricter per-endpoint rate limiting for write
// operations, or a no-op handler when rate limiting is disabled.
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
