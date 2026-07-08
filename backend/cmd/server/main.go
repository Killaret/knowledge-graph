package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	_ "knowledge-graph/docs"

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

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		retryDelay := cfg.DatabaseRetryDelaySeconds
		log.Printf("CRITICAL: database connection failed: %v, retrying in %ds...", err, retryDelay)
		time.Sleep(time.Duration(retryDelay) * time.Second)
		database, err = db.Connect(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("FATAL: database connection failed after retry: %v", err)
		}
	}
	log.Println("Connected to PostgreSQL")

	// Apply migrations
	migrationsDir := defaultMigrationsDir
	if err := postgres.RunMigrations(database, migrationsDir); err != nil {
		log.Printf("ERROR: Failed to run migrations: %v", err)
		log.Printf("WARNING: Continuing without migrations - database may be inconsistent")
	} else {
		log.Println("Migrations applied successfully")
	}

	// Redis with connection pool settings
	redisAddr := cfg.RedisURL
	log.Printf("Redis address: %s", redisAddr)
	redisClient := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		PoolSize:     10,              // Maximum number of connections
		MinIdleConns: 3,               // Minimum number of idle connections
		PoolTimeout:  4 * time.Second, // Timeout for getting connection from pool
	})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
	}()

	log.Println("Redis configured with connection pool: PoolSize=10, MinIdle=3, MaxConnAge=5m")

	// Redis flush on startup if configured
	if cfg.RedisFlushOnStartup {
		keysBefore, _ := redisClient.DBSize(ctx).Result()
		log.Printf("[Cache] Redis keys before flush: %d", keysBefore)
		if err := redisClient.FlushDB(ctx).Err(); err != nil {
			log.Printf("[Cache] WARNING: failed to flush Redis cache on startup: %v", err)
		} else {
			keysAfter, _ := redisClient.DBSize(ctx).Result()
			log.Printf("[Cache] SUCCESS: Redis cache flushed on startup (keys after: %d)", keysAfter)
		}
	} else {
		log.Printf("[Cache] Redis flush on startup is disabled")
	}

	// Graceful shutdown
	defer func() {
		log.Println("Server shutdown, closing database connection...")
		if database != nil {
			sqlDB, _ := database.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		}
	}()

	noteRepo := postgres.NewNoteRepository(database, redisClient)
	linkRepo := postgres.NewLinkRepository(database)
	embeddingRepo := postgres.NewEmbeddingRepository(database)

	// Yandex.Disk backup service
	var yandexBackupService *cloud.YandexBackupService
	// Task queue
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

	// Graph loaders
	linkLoader := appGraph.NewNeighborLoader(linkRepo, noteRepo)
	embeddingLoader := appGraph.NewEmbeddingNeighborLoader(embeddingRepo, cfg.EmbeddingSimilarityLimit)

	compositeLoader := appGraph.NewCompositeNeighborLoaderWithWeights(
		[]graphDomain.NeighborLoader{linkLoader, embeddingLoader},
		[]float64{cfg.RecommendationAlpha, cfg.RecommendationBeta},
	)

	traversalSvc := graphDomain.NewTraversalService(compositeLoader, cfg.RecommendationDepth, cfg.RecommendationDecay, cfg.BFSAggregation, cfg.BFSNormalize)

	suggestionsHandler := graph.NewGetSuggestionsHandler(traversalSvc, noteRepo, redisClient, cfg.RecommendationCacheTTL)

	// Recommendation repository and affected notes service
	recRepo := postgres.NewRecommendationRepository(database)
	affectedNotesSvc := recommendation.NewAffectedNotesService(recRepo)
	taskDelay := time.Duration(cfg.RecommendationTaskDelaySeconds) * time.Second

	// Achievement service
	achievementRepo := postgres.NewAchievementRepository(database)
	achievementEngine := achievement.NewEngine(database)
	userSettingsRepo := postgres.NewUserSettingsRepository(database)
	settingsService := userApp.NewSettingsService(userSettingsRepo, redisClient)
	achievementService := achievement.NewService(achievementEngine, achievementRepo, settingsService, redisClient)
	achievementHandler := achievementhandler.NewHandler(achievementService)

	// Graph cache
	graphCache := cache.NewGraphCache(redisClient)

	// Clear graph cache on startup
	log.Printf("[Cache] Clearing graph cache on startup...")
	if err := graphCache.InvalidateAll(ctx); err != nil {
		log.Printf("[Cache] WARNING: failed to clear graph cache on startup: %v", err)
	} else {
		log.Printf("[Cache] SUCCESS: Graph cache cleared on startup")
	}

	// Handlers with new parameters
	noteHandler := notehandler.New(noteRepo, taskQueue, suggestionsHandler, affectedNotesSvc, taskDelay, recRepo, embeddingRepo, redisClient, cfg, graphCache, achievementService)
	linkHandler := linkhandler.New(linkRepo, noteRepo, taskQueue, affectedNotesSvc, taskDelay, achievementService, graphCache)
	graphHandler := graphhandler.New(noteRepo, linkRepo, cfg, graphCache)
	tagRepo := postgres.NewTagRepository(database)
	tagHandler := taghandler.New(tagRepo, noteRepo)
	// Backup handler
	backupHandler := backphandler.NewHandler(cfg, yandexBackupService, taskQueue)

	// Auth handler
	jwtManager := authpkg.NewJWTManager(cfg.JWTSecret, time.Hour*24, time.Hour*24*7) // 24h access, 7d refresh
	tokenStore := authpkg.NewRedisTokenStore(redisClient)
	authHandler := authhandler.NewHandler(database, jwtManager, tokenStore, cfg)

	// User handler
	passwordConfig := &auth.PasswordConfig{
		Time:    3,
		Memory:  65536,
		Threads: 4,
		KeyLen:  32,
	}
	passwordPolicy := auth.DefaultPasswordPolicy()
	userHandler := userhandler.NewHandler(database, passwordConfig, passwordPolicy)

	// Router setup with all middleware and routes
	writeLimiter := newWriteLimiter(cfg)
	jwtConfig := newJWTConfig(jwtManager, tokenStore)
	apiKeyConfig := newAPIKeyConfig(database, cfg.APIKeyEnabled, cfg.StaticAPIKey)
	skipAuthConfig := middleware.DefaultSkipAuthConfig(cfg.SkipAuth)

	r := setupRouter(
		noteHandler,
		linkHandler,
		graphHandler,
		tagHandler,
		achievementHandler,
		authHandler,
		userHandler,
		backupHandler,
		cfg,
		database,
		redisClient,
		writeLimiter,
		jwtConfig,
		apiKeyConfig,
		skipAuthConfig,
	)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// Graceful shutdown channel (declared early for goroutines)
	quit := make(chan os.Signal, 1)

	// Start periodic pool statistics logging
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				stats := db.GetPoolStats(database)
				log.Printf("[Connection Pool] Stats: MaxOpen=%v, Open=%v, InUse=%v, Idle=%v, WaitCount=%v, WaitDuration=%v",
					stats["max_open_connections"],
					stats["open_connections"],
					stats["in_use"],
					stats["idle"],
					stats["wait_count"],
					stats["wait_duration"])
			case <-quit:
				return
			}
		}
	}()

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
