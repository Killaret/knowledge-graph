package main

import (
	"fmt"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/interfaces/api/graphhandler"
	achievementhandler "knowledge-graph/internal/interfaces/api/handlers/achievement"
	authhandler "knowledge-graph/internal/interfaces/api/handlers/auth"
	backphandler "knowledge-graph/internal/interfaces/api/handlers/backup"
	drafthandler "knowledge-graph/internal/interfaces/api/handlers/draft"
	settingshandler "knowledge-graph/internal/interfaces/api/handlers/settings"
	sharehandler "knowledge-graph/internal/interfaces/api/handlers/share"
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
	settingsHandler *settingshandler.Handler,
	shareHandler *sharehandler.Handler,
	draftHandler *drafthandler.Handler,
	cfg *config.Config,
	healthHandler gin.HandlerFunc,
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

	// Health check endpoint
	r.GET("/health", healthHandler)

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		// Auth routes (all write endpoints are rate-limited)
		v1.POST("/auth/register", writeLimiter, authHandler.Register)
		v1.POST("/auth/login", writeLimiter, authHandler.Login)
		v1.POST("/auth/logout", writeLimiter, authHandler.Logout)
		v1.POST("/auth/refresh", writeLimiter, authHandler.Refresh)
		v1.POST("/auth/forgot-password", writeLimiter, authHandler.ForgotPassword)
		v1.POST("/auth/reset-password", writeLimiter, authHandler.ResetPassword)
		v1.GET("/auth/yandex", authHandler.YandexLogin)
		v1.GET("/auth/yandex/callback", authHandler.YandexCallback)

		// User routes
		v1.GET("/users/me", userHandler.GetMe)

		// Settings routes
		v1.GET("/users/me/settings", settingsHandler.GetMySettings)
		v1.GET("/users/me/settings/:key", settingsHandler.GetSetting)
		v1.PUT("/users/me/settings", writeLimiter, settingsHandler.UpdateSetting)
		v1.DELETE("/users/me/settings/:key", writeLimiter, settingsHandler.DeleteSetting)
		v1.GET("/users/me/settings/galactic_mode", settingsHandler.GetGalacticMode)
		v1.POST("/users/me/settings/galactic_mode/toggle", writeLimiter, settingsHandler.ToggleGalacticMode)

		// Share routes
		v1.POST("/notes/:id/share", writeLimiter, shareHandler.ShareNote)
		v1.POST("/notes/:id/share-link", writeLimiter, shareHandler.CreateShareLink)
		v1.GET("/notes/:id/shares", shareHandler.ListNoteShares)
		v1.DELETE("/notes/:id/shares/:shareId", writeLimiter, shareHandler.RevokeShare)
		v1.DELETE("/share-links/:id", writeLimiter, shareHandler.RevokeShareLink)
		v1.GET("/share/:token", shareHandler.AccessSharedNote)

		// Draft routes (only if MongoDB is configured)
		if draftHandler != nil {
			v1.POST("/notes/:id/draft", writeLimiter, draftHandler.SaveDraft)
			v1.GET("/notes/:id/draft", draftHandler.GetDraft)
			v1.POST("/drafts/:draft_id/sync", writeLimiter, draftHandler.SyncDraft)
			v1.POST("/drafts/:draft_id/resolve", writeLimiter, draftHandler.ResolveConflict)
			v1.DELETE("/drafts/:draft_id", writeLimiter, draftHandler.DeleteDraft)
			v1.GET("/users/me/drafts", draftHandler.GetActiveDrafts)
		}

		// Achievements
		v1.GET("/achievements", achievementHandler.ListAchievements)
		v1.GET("/users/me/achievements", achievementHandler.GetUserAchievements)
		v1.POST("/users/me/achievements/:id/mark-seen", achievementHandler.MarkSeen)

		// Write operations with stricter rate limiting
		v1.POST("/notes", writeLimiter, noteHandler.Create)
		v1.GET("/notes/:id", cacheControlMiddleware(60), noteHandler.Get)
		v1.PUT("/notes/:id", writeLimiter, noteHandler.Update)
		v1.DELETE("/notes/:id", writeLimiter, noteHandler.Delete)
		v1.POST("/notes/batch", writeLimiter, noteHandler.DeleteBatch)
		v1.POST("/notes/:id/restore", writeLimiter, noteHandler.Restore)
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
	// TODO: Remove legacy /notes routes once all clients migrate to /api/v1/notes.
	// Migration progress should be tracked in docs/DEPRECATIONS.md before removal.
	r.POST("/notes", writeLimiter, noteHandler.Create)
	r.GET("/notes/:id", noteHandler.Get)
	r.PUT("/notes/:id", writeLimiter, noteHandler.Update)
	r.DELETE("/notes/:id", writeLimiter, noteHandler.Delete)
	r.POST("/notes/batch", writeLimiter, noteHandler.DeleteBatch)
	r.POST("/notes/:id/restore", writeLimiter, noteHandler.Restore)
	r.GET("/notes/:id/suggestions", noteHandler.GetSuggestions)
	r.GET("/notes", noteHandler.List)
	r.GET("/notes/search", noteHandler.Search)

	r.POST("/links", writeLimiter, linkHandler.Create)
	r.GET("/links/:id", linkHandler.Get)
	r.GET("/notes/:id/links", linkHandler.GetByNote)
	r.DELETE("/links/:id", writeLimiter, linkHandler.Delete)
	r.DELETE("/notes/:id/links", writeLimiter, linkHandler.DeleteByNote)

	r.GET("/notes/:id/graph", graphHandler.GetGraph)

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
