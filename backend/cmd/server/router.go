package main

import (
	"fmt"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/interfaces/api/graphhandler"
	achievementhandler "knowledge-graph/internal/interfaces/api/handlers/achievement"
	authhandler "knowledge-graph/internal/interfaces/api/handlers/auth"
	backphandler "knowledge-graph/internal/interfaces/api/handlers/backup"
	userhandler "knowledge-graph/internal/interfaces/api/handlers/user"
	"knowledge-graph/internal/interfaces/api/linkhandler"
	"knowledge-graph/internal/interfaces/api/middleware"
	"knowledge-graph/internal/interfaces/api/notehandler"
	"knowledge-graph/internal/interfaces/api/taghandler"
)

// setupRouter configures the Gin router with all routes and middleware
func setupRouter(
	noteHandler *notehandler.Handler,
	linkHandler *linkhandler.Handler,
	graphHandler *graphhandler.Handler,
	tagHandler *taghandler.Handler,
	achievementHandler *achievementhandler.Handler,
	authHandler *authhandler.Handler,
	userHandler *userhandler.Handler,
	backupHandler *backphandler.Handler,
	cfg *config.Config,
	database *gorm.DB,
	redisClient *redis.Client,
	writeLimiter gin.HandlerFunc,
	jwtConfig *middleware.JWTConfig,
	apiKeyConfig *middleware.APIKeyConfig,
	skipAuthConfig *middleware.SkipAuthConfig,
) *gin.Engine {
	r := gin.Default()

	// Gzip middleware - compress responses
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// Recovery middleware - catches panics and returns 500
	r.Use(middleware.RecoveryMiddleware())

	// CORS middleware - allow requests from frontend
	r.Use(corsMiddleware())

	// Structured logging middleware with token data
	r.Use(middleware.LoggingMiddleware())

	// SkipAuth middleware for testing (when SKIP_AUTH=true)
	if cfg.SkipAuth {
		r.Use(middleware.SkipAuth(skipAuthConfig))
	}

	// JWT middleware for authentication
	r.Use(middleware.JWTAuth(jwtConfig))

	// API Key middleware for authentication
	r.Use(middleware.APIKey(apiKeyConfig))

	// Rate limiting middleware (conditional)
	if cfg.ServerRateLimitEnabled {
		rateWindow := time.Duration(cfg.ServerRateLimitWindowSeconds) * time.Second
		r.Use(middleware.RateLimitMiddleware(cfg.ServerRateLimitRequests, rateWindow))
	}

	// Cache control middleware for static data endpoints
	cacheControlMiddleware := func(maxAge int) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Writer.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
			c.Next()
		}
	}

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/openapi.yaml"),
		ginSwagger.DeepLinking(true),
		ginSwagger.DocExpansion("list"),
	))

	// Serve OpenAPI spec
	r.StaticFile("/openapi.yaml", "./openAPI.yaml")

	// Health check endpoint (created in health.go)
	r.GET("/health", newHealthHandler(database, redisClient, cfg))

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Auth routes
		v1.POST("/auth/register", authHandler.Register)
		v1.POST("/auth/login", authHandler.Login)
		v1.POST("/auth/logout", authHandler.Logout)
		v1.POST("/auth/refresh", authHandler.Refresh)
		v1.POST("/auth/forgot-password", authHandler.ForgotPassword)
		v1.POST("/auth/reset-password", authHandler.ResetPassword)
		v1.GET("/auth/yandex", authHandler.YandexLogin)
		v1.GET("/auth/yandex/callback", authHandler.YandexCallback)

		// User routes
		v1.GET("/users/me", userHandler.GetMe)

		// Achievements
		v1.GET("/achievements", achievementHandler.ListAchievements)
		v1.GET("/users/me/achievements", achievementHandler.GetUserAchievements)
		v1.POST("/users/me/achievements/:id/mark-seen", achievementHandler.MarkSeen)

		// Write operations with stricter rate limiting
		v1.POST("/notes", writeLimiter, noteHandler.Create)
		v1.GET("/notes/:id", cacheControlMiddleware(60), noteHandler.Get)
		v1.PUT("/notes/:id", writeLimiter, noteHandler.Update)
		v1.DELETE("/notes/:id", writeLimiter, noteHandler.Delete)
		v1.GET("/notes/:id/suggestions", cacheControlMiddleware(60), noteHandler.GetSuggestions)
		v1.GET("/notes", cacheControlMiddleware(60), noteHandler.List)
		v1.GET("/notes/search", cacheControlMiddleware(30), noteHandler.Search)

		v1.POST("/links", writeLimiter, linkHandler.Create)
		v1.GET("/links/:id", linkHandler.Get)
		v1.GET("/notes/:id/links", linkHandler.GetByNote)
		v1.DELETE("/links/:id", writeLimiter, linkHandler.Delete)
		v1.DELETE("/notes/:id/links", writeLimiter, linkHandler.DeleteByNote)

		v1.GET("/notes/:id/graph", cacheControlMiddleware(300), graphHandler.GetGraph)
		v1.GET("/graph/all", cacheControlMiddleware(300), graphHandler.GetFullGraph)
		v1.GET("/me/graph/cached", cacheControlMiddleware(60), graphHandler.GetCachedGraph)
		v1.GET("/me/graph/fresh", cacheControlMiddleware(0), graphHandler.GetFreshGraph)

		// Tag routes
		v1.POST("/tags", writeLimiter, tagHandler.Create)
		v1.GET("/tags", tagHandler.List)
		v1.GET("/tags/:id", tagHandler.Get)
		v1.PUT("/tags/:id", writeLimiter, tagHandler.Update)
		v1.DELETE("/tags/:id", writeLimiter, tagHandler.Delete)
		v1.POST("/notes/:id/tags", writeLimiter, tagHandler.AddTagToNote)
		v1.DELETE("/notes/:id/tags/:tagId", writeLimiter, tagHandler.RemoveTagFromNote)
		v1.GET("/notes/:id/tags", tagHandler.GetTagsByNote)

		// Backup routes
		v1.POST("/backup/cloud", writeLimiter, backupHandler.TriggerCloudBackup)
		v1.GET("/backup/status", backupHandler.GetBackupStatus)
	}

	// Legacy routes (deprecated - kept for backward compatibility)
	// TODO(# Issue): Remove legacy routes after all clients migrate to /api/v1
	r.POST("/notes", writeLimiter, noteHandler.Create)
	r.GET("/notes/:id", noteHandler.Get)
	r.PUT("/notes/:id", writeLimiter, noteHandler.Update)
	r.DELETE("/notes/:id", writeLimiter, noteHandler.Delete)
	r.GET("/notes/:id/suggestions", noteHandler.GetSuggestions)
	r.GET("/notes", noteHandler.List)
	r.GET("/notes/search", noteHandler.Search)

	r.POST("/links", writeLimiter, linkHandler.Create)
	r.GET("/links/:id", linkHandler.Get)
	r.GET("/notes/:id/links", linkHandler.GetByNote)
	r.DELETE("/links/:id", writeLimiter, linkHandler.Delete)
	r.DELETE("/notes/:id/links", writeLimiter, linkHandler.DeleteByNote)

	r.GET("/notes/:id/graph", graphHandler.GetGraph)
	r.GET("/graph/all", graphHandler.GetFullGraph)

	// Tag routes
	r.POST("/tags", writeLimiter, tagHandler.Create)
	r.GET("/tags", tagHandler.List)
	r.GET("/tags/:id", tagHandler.Get)
	r.PUT("/tags/:id", writeLimiter, tagHandler.Update)
	r.DELETE("/tags/:id", writeLimiter, tagHandler.Delete)
	r.POST("/notes/:id/tags", writeLimiter, tagHandler.AddTagToNote)
	r.DELETE("/notes/:id/tags/:tagId", writeLimiter, tagHandler.RemoveTagFromNote)
	r.GET("/notes/:id/tags", tagHandler.GetTagsByNote)

	return r
}
