package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"knowledge-graph/internal/application/common"
	appGraph "knowledge-graph/internal/application/graph"
	graphQueries "knowledge-graph/internal/application/queries/graph"
	"knowledge-graph/internal/config"
	graphDomain "knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/queue"
	"knowledge-graph/internal/interfaces/api/graphhandler"
	"knowledge-graph/internal/interfaces/api/linkhandler"
	"knowledge-graph/internal/interfaces/api/notehandler"
)

func main() {
	cfg := config.Load()

	log.Printf("Config loaded: alpha=%.2f, beta=%.2f, depth=%d, decay=%.2f, cacheTTL=%v, embeddingLimit=%d, graphLoadDepth=%d",
		cfg.RecommendationAlpha, cfg.RecommendationBeta,
		cfg.RecommendationDepth, cfg.RecommendationDecay,
		cfg.RecommendationCacheTTL, cfg.EmbeddingSimilarityLimit, cfg.GraphLoadDepth)

	db.Init()
	if db.DB == nil {
		log.Fatal("database connection is nil")
	}
	log.Println("Connected to PostgreSQL")

	// Redis
	redisAddr := cfg.RedisURL
	log.Printf("Redis address: %s", redisAddr)
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
	}()

	noteRepo := postgres.NewNoteRepository(db.DB, redisClient)
	linkRepo := postgres.NewLinkRepository(db.DB)
	embeddingRepo := postgres.NewEmbeddingRepository(db.DB)

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
	}

	// Загрузчики графа
	linkLoader := appGraph.NewNeighborLoader(linkRepo, noteRepo)
	embeddingLoader := appGraph.NewEmbeddingNeighborLoader(embeddingRepo, cfg.EmbeddingSimilarityLimit)

	compositeLoader := appGraph.NewCompositeNeighborLoaderWithWeights(
		[]graphDomain.NeighborLoader{linkLoader, embeddingLoader},
		[]float64{cfg.RecommendationAlpha, cfg.RecommendationBeta},
	)

	traversalSvc := graphDomain.NewTraversalService(compositeLoader, cfg.RecommendationDepth, cfg.RecommendationDecay)

	suggestionsHandler := graphQueries.NewGetSuggestionsHandler(traversalSvc, noteRepo, redisClient, cfg.RecommendationCacheTTL)

	// Хендлеры
	noteHandler := notehandler.New(noteRepo, taskQueue, suggestionsHandler)
	linkHandler := linkhandler.New(linkRepo, noteRepo)
	graphHandler := graphhandler.New(noteRepo, linkRepo, cfg.GraphLoadDepth)

	// Роуты
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/db-check", func(c *gin.Context) {
		sqlDB, err := db.DB.DB()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(500, gin.H{"status": "db not reachable"})
			return
		}
		c.JSON(200, gin.H{"status": "db ok"})
	})

	r.POST("/notes", noteHandler.Create)
	r.GET("/notes/:id", noteHandler.Get)
	r.PUT("/notes/:id", noteHandler.Update)
	r.DELETE("/notes/:id", noteHandler.Delete)
	r.GET("/notes/:id/suggestions", noteHandler.GetSuggestions)
	r.GET("/notes", noteHandler.List)
	r.GET("/notes/search", noteHandler.Search)

	r.POST("/links", linkHandler.Create)
	r.GET("/links/:id", linkHandler.Get)
	r.GET("/notes/:id/links", linkHandler.GetByNote)
	r.DELETE("/links/:id", linkHandler.Delete)
	r.DELETE("/notes/:id/links", linkHandler.DeleteByNote)

	r.GET("/notes/:id/graph", graphHandler.GetGraph)
	r.GET("/graph/all", graphHandler.GetFullGraph)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
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
