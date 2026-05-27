package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"knowledge-graph/internal/application/achievement"
	"knowledge-graph/internal/application/cache"
	"knowledge-graph/internal/application/common"
	appGraph "knowledge-graph/internal/application/graph"
	"knowledge-graph/internal/application/queries/graph"
	"knowledge-graph/internal/application/recommendation"
	userApp "knowledge-graph/internal/application/user"
	"knowledge-graph/internal/auth"
	authpkg "knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	graphDomain "knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/infrastructure/cloud"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/nlp" // Восстановление импорта nlp пакета
	"knowledge-graph/internal/infrastructure/queue"
	"knowledge-graph/internal/interfaces/api/graphhandler"
	achievementhandler "knowledge-graph/internal/interfaces/api/handlers/achievement"
	authhandler "knowledge-graph/internal/interfaces/api/handlers/auth"
	backphandler "knowledge-graph/internal/interfaces/api/handlers/backup"
	userhandler "knowledge-graph/internal/interfaces/api/handlers/user"
	"knowledge-graph/internal/interfaces/api/linkhandler"
	"knowledge-graph/internal/interfaces/api/middleware"
	"knowledge-graph/internal/interfaces/api/notehandler"
	"knowledge-graph/internal/interfaces/api/taghandler"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	// Migrations directory path
	defaultMigrationsDir = "./migrations"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Printf("FATAL: Failed to load configuration: %v", err)
		os.Exit(1)
	}

	log.Printf("Config loaded: alpha=%.2f, beta=%.2f, depth=%d, decay=%.2f, cacheTTL=%v, embeddingLimit=%d, graphLoadDepth=%d",
		cfg.RecommendationAlpha, cfg.RecommendationBeta,
		cfg.RecommendationDepth, cfg.RecommendationDecay,
		cfg.RecommendationCacheTTL, cfg.EmbeddingSimilarityLimit, cfg.GraphLoadDepth)

	db.Init()
	if db.DB == nil {
		retryDelay := cfg.DatabaseRetryDelaySeconds
		log.Printf("CRITICAL: database connection is nil, retrying in %ds...", retryDelay)
		time.Sleep(time.Duration(retryDelay) * time.Second)
		db.Init()
		if db.DB == nil {
			log.Printf("FATAL: database connection failed after retry")
			os.Exit(1)
		}
	}
	log.Println("Connected to PostgreSQL")

	// Применяем миграции
	migrationsDir := defaultMigrationsDir
	if err := postgres.RunMigrations(db.DB, migrationsDir); err != nil {
		log.Printf("ERROR: Failed to run migrations: %v", err)
		log.Printf("WARNING: Continuing without migrations - database may be inconsistent")
	} else {
		log.Println("Migrations applied successfully")
	}

	// Redis
	redisAddr := cfg.RedisURL
	log.Printf("Redis address: %s", redisAddr)
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
	}()

	// Graceful shutdown
	defer func() {
		log.Println("Server shutdown, closing database connection...")
		if db.DB != nil {
			sqlDB, _ := db.DB.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		}
	}()

	noteRepo := postgres.NewNoteRepository(db.DB, redisClient)
	linkRepo := postgres.NewLinkRepository(db.DB)
	embeddingRepo := postgres.NewEmbeddingRepository(db.DB)

		// Yandex.Disk backup service
		var yandexBackupService *cloud.YandexBackupService
	// Очередь
	var taskQueue common.TaskQueue
	asynqClient, err := queue.NewAsynqClient(redisAddr)
	if err != nil {
		log.Printf("WARNING: failed to create asynq client: %v", err)
	} else {
		log.Printf("Asynq client created successfully")
		taskQueue = asynqClient
		defer func() {
			if err := asynqClient.Close(); err != nil {
				log.Printf("Error closing asynq client: %v", err)
			}
		}()
		if cfg.BackupCloudProvider == "yandex" && cfg.BackupYandexOAuthToken != "" {
		yandexCfg := cloud.YandexConfig{
		OAuthToken:   cfg.BackupYandexOAuthToken,
		BackupFolder: cfg.BackupYandexFolder,
		MaxBackups:   cfg.BackupYandexMaxBackups,
		}
		yandexBackupService, err = cloud.NewYandexBackupService(yandexCfg)
		if err != nil {
		log.Printf("WARNING: failed to create Yandex.Disk backup service: %v", err)
		} else {
			log.Printf("Yandex.Disk backup service initialized successfully")
		}
		}
	}

	// Загрузчики графа
	linkLoader := appGraph.NewNeighborLoader(linkRepo, noteRepo)
	embeddingLoader := appGraph.NewEmbeddingNeighborLoader(embeddingRepo, cfg.EmbeddingSimilarityLimit)

	compositeLoader := appGraph.NewCompositeNeighborLoaderWithWeights(
		[]graphDomain.NeighborLoader{linkLoader, embeddingLoader},
		[]float64{cfg.RecommendationAlpha, cfg.RecommendationBeta},
	)

	traversalSvc := graphDomain.NewTraversalService(compositeLoader, cfg.RecommendationDepth, cfg.RecommendationDecay, cfg.BFSAggregation, cfg.BFSNormalize)

	suggestionsHandler := graph.NewGetSuggestionsHandler(traversalSvc, noteRepo, redisClient, cfg.RecommendationCacheTTL)

	// Recommendation repository and affected notes service
	recRepo := postgres.NewRecommendationRepository(db.DB)
	affectedNotesSvc := recommendation.NewAffectedNotesService(recRepo)
	taskDelay := time.Duration(cfg.RecommendationTaskDelaySeconds) * time.Second

	// Achievement service
	achievementRepo := postgres.NewAchievementRepository(db.DB)
	achievementEngine := achievement.NewEngine(db.DB)
	userSettingsRepo := postgres.NewUserSettingsRepository(db.DB)
	settingsService := userApp.NewSettingsService(userSettingsRepo, redisClient)
	achievementService := achievement.NewService(achievementEngine, achievementRepo, settingsService, redisClient)
	achievementHandler := achievementhandler.NewHandler(achievementService)

	// Graph cache
	graphCache := cache.NewGraphCache(redisClient)

	// Очистка граф кэша при старте сервера
	log.Printf("[Cache] Clearing graph cache on startup...")
	if err := graphCache.InvalidateAll(ctx); err != nil {
		log.Printf("[Cache] WARNING: failed to clear graph cache on startup: %v", err)
	} else {
		log.Printf("[Cache] SUCCESS: Graph cache cleared on startup")
	}

	// Хендлеры с новыми параметрами
	noteHandler := notehandler.New(noteRepo, taskQueue, suggestionsHandler, affectedNotesSvc, taskDelay, recRepo, embeddingRepo, redisClient, cfg, graphCache, achievementService)
	linkHandler := linkhandler.New(linkRepo, noteRepo, taskQueue, affectedNotesSvc, taskDelay, achievementService, graphCache)
	graphHandler := graphhandler.New(noteRepo, linkRepo, cfg, graphCache)
	tagRepo := postgres.NewTagRepository(db.DB)
	tagHandler := taghandler.New(tagRepo, noteRepo)
	// Backup handler
	backupHandler := backphandler.NewHandler(cfg, yandexBackupService, taskQueue)

	// Auth handler
	jwtManager := authpkg.NewJWTManager(cfg.JWTSecret, time.Hour*24, time.Hour*24*7) // 24h access, 7d refresh
	tokenStore := authpkg.NewRedisTokenStore(redisClient)
	authHandler := authhandler.NewHandler(db.DB, jwtManager, tokenStore, cfg)

	// User handler
	passwordConfig := &auth.PasswordConfig{
		Time:    3,
		Memory:  65536,
		Threads: 4,
		KeyLen:  32,
	}
	passwordPolicy := auth.DefaultPasswordPolicy()
	userHandler := userhandler.NewHandler(db.DB, passwordConfig, passwordPolicy)

	// Роуты
	r := gin.Default()

	// Gzip middleware - compress responses
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// Recovery middleware - catches panics and returns 500
	r.Use(middleware.RecoveryMiddleware())

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

	// CORS middleware - разрешаем запросы с frontend
	r.Use(func(c *gin.Context) {
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
	})

	// Structured logging middleware with token data
	r.Use(middleware.LoggingMiddleware())

	// SkipAuth middleware for testing (when SKIP_AUTH=true)
	if cfg.SkipAuth {
		r.Use(middleware.SkipAuth(middleware.DefaultSkipAuthConfig(true)))
		log.Println("[Auth] SKIP_AUTH enabled - authentication disabled for testing")
	}

	// JWT middleware for authentication
	jwtConfig := middleware.DefaultJWTConfig(jwtManager, tokenStore)
	r.Use(middleware.JWTAuth(jwtConfig))

	// API Key middleware for authentication
	apiKeyConfig := middleware.DefaultAPIKeyConfig(db.DB, cfg.APIKeyEnabled, cfg.StaticAPIKey)
	r.Use(middleware.APIKey(apiKeyConfig))

	// Rate limiting middleware (conditional)
	var writeLimiter gin.HandlerFunc
	if cfg.ServerRateLimitEnabled {
		rateWindow := time.Duration(cfg.ServerRateLimitWindowSeconds) * time.Second
		r.Use(middleware.RateLimitMiddleware(cfg.ServerRateLimitRequests, rateWindow))

		// Stricter rate limiting for write operations - build endpoint map from config
		endpointLimits := map[string]int{
			"/notes":     cfg.ServerRateLimitEndpoints["notes_create"],
			"/links":     cfg.ServerRateLimitEndpoints["links_create"],
			"/notes/:id": cfg.ServerRateLimitEndpoints["notes_update"],
		}
		writeLimiter = middleware.RateLimitByEndpoint(endpointLimits, cfg.ServerRateLimitRequests, rateWindow)
	} else {
		// No-op handler when rate limiting is disabled
		writeLimiter = func(c *gin.Context) { c.Next() }
	}

	// Comprehensive health check with all dependencies
	r.GET("/health", func(c *gin.Context) {
		health := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		}
		status := http.StatusOK

		// Check database
		sqlDB, err := db.DB.DB()
		if err != nil {
			health["database"] = gin.H{"status": "unhealthy", "error": err.Error()}
			status = http.StatusServiceUnavailable
		} else if err := sqlDB.Ping(); err != nil {
			health["database"] = gin.H{"status": "unhealthy", "error": err.Error()}
			status = http.StatusServiceUnavailable
		} else {
			health["database"] = gin.H{"status": "healthy"}
		}

		// Check Redis
		if redisClient != nil {
			if err := redisClient.Ping(ctx).Err(); err != nil {
				health["redis"] = gin.H{"status": "unhealthy", "error": err.Error()}
				status = http.StatusServiceUnavailable
			} else {
				health["redis"] = gin.H{"status": "healthy"}
			}
		} else {
			health["redis"] = gin.H{"status": "disabled"}
		}

		// Check NLP service
		nlpClient := nlp.NewNLPClient(cfg.NLPServiceURL, redisClient, cfg.RecommendationCacheTTL)
		if err := nlpClient.HealthCheck(ctx); err != nil {
			health["nlp"] = gin.H{"status": "unhealthy", "error": err.Error()}
			// Don't mark as unhealthy if NLP is optional
		} else {
			health["nlp"] = gin.H{"status": "healthy"}
		}

		c.JSON(status, health)
	})

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

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("CRITICAL: Failed to start server: %v", err)
			// Attempt to restart server on fallback port
			if len(cfg.ServerFallbackPorts) > 0 {
				fallbackPort := cfg.ServerFallbackPorts[0]
				srv.Addr = ":" + fallbackPort
				log.Printf("Attempting to restart on port %s...", fallbackPort)
			} else {
				log.Printf("FATAL: No fallback ports configured")
				os.Exit(1)
			}
			if err := srv.ListenAndServe(); err != nil {
				log.Printf("FATAL: Server restart failed: %v", err)
				os.Exit(1)
			}
		}
	}()

	log.Printf("Server started on port %s", cfg.ServerPort)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
