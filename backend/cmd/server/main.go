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

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	_ "knowledge-graph/docs"

	"knowledge-graph/internal/application/achievement"
	"knowledge-graph/internal/application/cache"
	"knowledge-graph/internal/application/common"
	draftApp "knowledge-graph/internal/application/draft"
	appGraph "knowledge-graph/internal/application/graph"
	importer "knowledge-graph/internal/application/import"
	"knowledge-graph/internal/application/queries/graph"
	"knowledge-graph/internal/application/recommendation"
	userApp "knowledge-graph/internal/application/user"
	"knowledge-graph/internal/auth"
	"knowledge-graph/internal/config"
	graphDomain "knowledge-graph/internal/domain/graph"
	infracache "knowledge-graph/internal/infrastructure/cache"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/email"
	"knowledge-graph/internal/infrastructure/events"
	graphinfra "knowledge-graph/internal/infrastructure/graph"
	"knowledge-graph/internal/infrastructure/mongo"
	"knowledge-graph/internal/infrastructure/nlp"
	oauthpkg "knowledge-graph/internal/infrastructure/oauth"
	"knowledge-graph/internal/infrastructure/queue"
	"knowledge-graph/internal/infrastructure/web"
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

	database, err := connectDatabaseWithRetry(ctx, cfg)
	if err != nil {
		log.Fatalf("FATAL: database connection failed: %v", err)
	}

	redisClient := newRedisClient(cfg)
	mongoClient := newMongoClient(ctx, cfg)
	taskQueue := newAsynqClient(cfg)

	srv, cleanup, err := run(
		ctx,
		cfg,
		database,
		redisClient,
		mongoClient,
		taskQueue,
		defaultMigrationsDir,
	)
	if err != nil {
		log.Fatalf("FATAL: failed to build server: %v", err)
	}
	defer cleanup()

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// run wires all application dependencies and returns an HTTP server ready to start.
// All external resources (database, Redis, MongoDB, task queue) are injected from main
// so that run can be covered with unit tests using mocks/fakes.
func run(
	ctx context.Context,
	cfg *config.Config,
	database *gorm.DB,
	redisClient *redis.Client,
	mongoClient *mongo.Client,
	taskQueue common.TaskQueue,
	migrationsDir string,
) (*http.Server, func(), error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("config is nil")
	}
	if database == nil {
		return nil, nil, fmt.Errorf("database is nil")
	}

	// Apply migrations (non-fatal)
	if migrationsDir != "" {
		if err := postgres.RunMigrations(database, migrationsDir); err != nil {
			log.Printf("ERROR: Failed to run migrations: %v", err)
			log.Printf("WARNING: Continuing without migrations - database may be inconsistent")
		} else {
			log.Println("Migrations applied successfully")
		}
	}

	cacheClient := infracache.NewRedisCacheClient(redisClient)

	noteRepo := postgres.NewNoteRepository(database, cacheClient)
	linkRepo := postgres.NewLinkRepository(database)
	embeddingRepo := postgres.NewEmbeddingRepository(database)

	// Draft service and handler (only if MongoDB is available)
	var draftHandler *drafthandler.Handler
	if mongoClient != nil {
		draftRepo := mongo.NewDraftRepository(mongoClient)
		draftService := draftApp.NewService(draftRepo, noteRepo, "")
		draftHandler = drafthandler.NewHandler(draftService)
		log.Println("[Draft] Draft service initialized")
	} else {
		log.Println("[Draft] Draft service disabled (MongoDB unavailable)")
	}

	// Graph loaders
	linkLoader := appGraph.NewNeighborLoader(linkRepo, noteRepo)
	embeddingLoader := appGraph.NewEmbeddingNeighborLoader(embeddingRepo, cfg.EmbeddingSimilarityLimit)

	compositeLoader := appGraph.NewCompositeNeighborLoaderWithWeights(
		[]graphDomain.NeighborLoader{linkLoader, embeddingLoader},
		[]float64{cfg.RecommendationAlpha, cfg.RecommendationBeta},
	)

	traversalSvc := graphDomain.NewTraversalService(compositeLoader, cfg.RecommendationDepth, cfg.RecommendationDecay, cfg.BFSAggregation, cfg.BFSNormalize)

	// Graph-service client for analytics (falls back to in-memory BFS if unavailable).
	if cfg.GraphServiceURL != "" {
		graphServiceClient := graphinfra.NewClient(cfg)
		traversalSvc.SetGraphServiceClient(graphServiceClient)
		log.Printf("[GraphService] Graph service client configured: %s", cfg.GraphServiceURL)
	}

	suggestionsHandler := graph.NewGetSuggestionsHandler(traversalSvc, noteRepo, cacheClient, cfg.RecommendationCacheTTL)

	// Recommendation repository and affected notes service
	recRepo := postgres.NewRecommendationRepository(database)
	affectedNotesSvc := recommendation.NewAffectedNotesService(recRepo)
	taskDelay := time.Duration(cfg.RecommendationTaskDelaySeconds) * time.Second

	// Achievement service
	achievementRepo := postgres.NewAchievementRepository(database)
	achievementCounter := postgres.NewAchievementCounter(database)
	achievementEngine := achievement.NewEngine(achievementCounter)
	userSettingsRepo := postgres.NewUserSettingsRepository(database)
	settingsService := userApp.NewSettingsService(userSettingsRepo, cacheClient)
	achievementService := achievement.NewService(achievementEngine, achievementRepo, settingsService, cacheClient, taskQueue)
	achievementHandler := achievementhandler.NewHandler(achievementService)

	// Graph cache (nil when Redis is unavailable)
	var graphCache *cache.GraphCache
	if cacheClient != nil {
		graphCache = cache.NewGraphCache(cacheClient)
		log.Printf("[Cache] Clearing graph cache on startup...")
		if err := graphCache.InvalidateAll(ctx); err != nil {
			log.Printf("[Cache] WARNING: failed to clear graph cache on startup: %v", err)
		} else {
			log.Printf("[Cache] SUCCESS: Graph cache cleared on startup")
		}
	}

	// Graph event publisher (nil when Redis is unavailable)
	var eventPublisher *events.Publisher
	if redisClient != nil {
		eventPublisher = events.NewPublisher(redisClient, cfg.EventChannel)
	}

	// Import service and handlers
	importService := importer.NewService(noteRepo, cacheClient, taskQueue, web.NewImportFetcher())

	// Handlers with new parameters
	noteHandler := notehandler.New(noteRepo, taskQueue, suggestionsHandler, affectedNotesSvc, taskDelay, recRepo, embeddingRepo, cacheClient, cfg, graphCache, achievementService, importService)
	linkHandler := linkhandler.New(linkRepo, noteRepo, achievementService, graphCache)
	if eventPublisher != nil {
		noteHandler.SetEventPublisher(eventPublisher)
		linkHandler.SetEventPublisher(eventPublisher)
	}
	graphHandler := graphhandler.New(noteRepo, linkRepo, cfg, graphCache)
	tagRepo := postgres.NewTagRepository(database)
	tagHandler := taghandler.New(tagRepo, noteRepo)
	// Backup handler
	backupHandler := backphandler.NewHandler(cfg, taskQueue)

	// Auth handler
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	var tokenStore auth.TokenStore
	if redisClient != nil {
		tokenStore = auth.NewRedisTokenStore(redisClient)
	}
	roleRepo := postgres.NewRoleRepository(database)
	userRepo := postgres.NewUserRepository(database, roleRepo)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(database)
	apiKeyRepo := postgres.NewAPIKeyRepository(database)
	emailSender := email.FromConfig(cfg)
	oauthProviderFactory := func(clientID, clientSecret, redirectURI string) auth.OAuthProvider {
		return oauthpkg.NewYandex(clientID, clientSecret, redirectURI)
	}
	authHandler := authhandler.NewHandler(userRepo, refreshTokenRepo, tokenStore, jwtManager, cfg, emailSender, oauthProviderFactory)

	// User handler
	passwordConfig := &auth.PasswordConfig{
		Time:    3,
		Memory:  65536,
		Threads: 4,
		KeyLen:  32,
	}
	passwordPolicy := auth.DefaultPasswordPolicy()
	userHandler := userhandler.NewHandler(userRepo, apiKeyRepo, passwordConfig, passwordPolicy)

	// Settings handler
	settingsHandler := settingshandler.NewHandler(settingsService)

	// Share handler
	shareRepo := postgres.NewShareRepository(database)
	shareHandler := sharehandler.NewHandler(noteRepo, userRepo, shareRepo)

	// Health handler
	sqlDB, _ := database.DB()
	nlpClient := nlp.NewNLPClient(cfg.NLPServiceURL, redisClient, cfg.RecommendationCacheTTL)
	var redisPinger RedisPinger
	if redisClient != nil {
		redisPinger = &redisPingAdapter{client: redisClient}
	}
	healthHandler := newHealthHandler(sqlDB, redisPinger, nlpClient)

	// Router setup with all middleware and routes
	writeLimiter := newWriteLimiter(cfg)
	jwtConfig := newJWTConfig(jwtManager, tokenStore)
	apiKeyConfig := newAPIKeyConfig(apiKeyRepo, cfg.APIKeyEnabled, cfg.StaticAPIKey)
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
		settingsHandler,
		shareHandler,
		draftHandler,
		cfg,
		healthHandler,
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

	cleanup := func() {
		if asynq, ok := taskQueue.(interface{ Close() error }); ok && asynq != nil {
			if err := asynq.Close(); err != nil {
				log.Printf("Error closing asynq client: %v", err)
			}
		}
		if mongoClient != nil {
			if err := mongoClient.Close(ctx); err != nil {
				log.Printf("[MongoDB] Error closing MongoDB client: %v", err)
			}
		}
		if redisClient != nil {
			if err := redisClient.Close(); err != nil {
				log.Printf("Error closing redis client: %v", err)
			}
		}
		if database != nil {
			if sqlDB, err := database.DB(); err == nil && sqlDB != nil {
				sqlDB.Close()
			}
		}
	}

	return srv, cleanup, nil
}

func connectDatabaseWithRetry(ctx context.Context, cfg *config.Config) (*gorm.DB, error) {
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		retryDelay := cfg.DatabaseRetryDelaySeconds
		log.Printf("CRITICAL: database connection failed: %v, retrying in %ds...", err, retryDelay)
		time.Sleep(time.Duration(retryDelay) * time.Second)
		database, err = db.Connect(cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("database connection failed after retry: %w", err)
		}
	}
	log.Println("Connected to PostgreSQL")
	return database, nil
}

func newRedisClient(cfg *config.Config) *redis.Client {
	if cfg.RedisURL == "" {
		log.Println("[Redis] Redis URL not configured")
		return nil
	}

	redisAddr := cfg.RedisURL
	log.Printf("Redis address: %s", redisAddr)
	redisClient := redis.NewClient(&redis.Options{
		Addr:            redisAddr,
		PoolSize:        10,
		MinIdleConns:    3,
		PoolTimeout:     4 * time.Second,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	})

	// Redis flush on startup if configured
	if cfg.RedisFlushOnStartup {
		ctx := context.Background()
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

	log.Println("Redis configured with connection pool: PoolSize=10, MinIdle=3, ConnMaxLifetime=5m, ConnMaxIdleTime=1m")
	return redisClient
}

// redisPingAdapter adapts *redis.Client to the RedisPinger interface used by the health handler.
type redisPingAdapter struct {
	client *redis.Client
}

func (a *redisPingAdapter) Ping(ctx context.Context) error {
	return a.client.Ping(ctx).Err()
}

func newMongoClient(ctx context.Context, cfg *config.Config) *mongo.Client {
	if cfg.MongoDBURL == "" {
		log.Printf("[MongoDB] MongoDB URL not configured, draft feature disabled")
		return nil
	}

	mongoClient, err := mongo.NewClient(ctx, cfg.MongoDBURL, cfg.MongoDBDatabase)
	if err != nil {
		log.Printf("[MongoDB] WARNING: failed to connect to MongoDB at %s: %v", cfg.MongoDBURL, err)
		return nil
	}
	log.Printf("[MongoDB] Connected to MongoDB at %s/%s", cfg.MongoDBURL, cfg.MongoDBDatabase)
	return mongoClient
}

func newAsynqClient(cfg *config.Config) common.TaskQueue {
	if cfg.RedisURL == "" {
		log.Println("[Asynq] Redis URL not configured, task queue disabled")
		return nil
	}

	asynqClient, err := queue.NewAsynqClient(cfg.RedisURL)
	if err != nil {
		log.Printf("WARNING: failed to create asynq client: %v", err)
		return nil
	}
	log.Printf("Asynq client created successfully")
	return asynqClient
}
