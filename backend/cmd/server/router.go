package main

import (
	"time"

	"github.com/gin-gonic/gin"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/interfaces/api/graphhandler"
	"knowledge-graph/internal/interfaces/api/linkhandler"
	"knowledge-graph/internal/interfaces/api/middleware"
	"knowledge-graph/internal/interfaces/api/notehandler"
)

// setupRouter wires up middleware and all HTTP routes.
func setupRouter(
	noteHandler *notehandler.Handler,
	linkHandler *linkhandler.Handler,
	graphHandler *graphhandler.Handler,
	cfg *config.Config,
	writeLimiter gin.HandlerFunc,
	healthCheck gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	// Rate limiting
	if cfg.ServerRateLimitEnabled {
		rateWindow := time.Duration(cfg.ServerRateLimitWindowSeconds) * time.Second
		r.Use(middleware.RateLimitMiddleware(cfg.ServerRateLimitRequests, rateWindow))
	}

	r.GET("/health", healthCheck)

	// Notes
	r.POST("/notes", writeLimiter, noteHandler.Create)
	r.GET("/notes/:id", noteHandler.Get)
	r.PUT("/notes/:id", writeLimiter, noteHandler.Update)
	r.DELETE("/notes/:id", writeLimiter, noteHandler.Delete)
	r.GET("/notes/:id/suggestions", noteHandler.GetSuggestions)
	r.GET("/notes", noteHandler.List)
	r.GET("/notes/search", noteHandler.Search)

	// Links
	r.POST("/links", writeLimiter, linkHandler.Create)
	r.GET("/links/:id", linkHandler.Get)
	r.GET("/notes/:id/links", linkHandler.GetByNote)
	r.DELETE("/links/:id", writeLimiter, linkHandler.Delete)
	r.DELETE("/notes/:id/links", writeLimiter, linkHandler.DeleteByNote)

	// Graph
	r.GET("/notes/:id/graph", graphHandler.GetGraph)
	r.GET("/graph/all", graphHandler.GetFullGraph)

	return r
}
