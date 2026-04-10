package main

import (
	"log"

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

	log.Printf("Config loaded: alpha=%.2f, beta=%.2f, depth=%d, decay=%.2f, cacheTTL=%v, embeddingLimit=%d",
		cfg.RecommendationAlpha, cfg.RecommendationBeta,
		cfg.RecommendationDepth, cfg.RecommendationDecay,
		cfg.RecommendationCacheTTL, cfg.EmbeddingSimilarityLimit)

	db.Init()
	if db.DB == nil {
		log.Fatal("database connection is nil")
	}
	log.Println("Connected to PostgreSQL")

	noteRepo := postgres.NewNoteRepository(db.DB)
	linkRepo := postgres.NewLinkRepository(db.DB)
	embeddingRepo := postgres.NewEmbeddingRepository(db.DB)

	redisAddr := cfg.RedisURL
	log.Printf("Redis address: %s", redisAddr)

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

	// Redis
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
	}()

	suggestionsHandler := graphQueries.NewGetSuggestionsHandler(traversalSvc, noteRepo, redisClient, cfg.RecommendationCacheTTL)

	// Хендлеры
	noteHandler := notehandler.New(noteRepo, taskQueue, suggestionsHandler)
	linkHandler := linkhandler.New(linkRepo, noteRepo)
	graphHandler := graphhandler.New(noteRepo, linkRepo)

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

	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
