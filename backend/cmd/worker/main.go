package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"knowledge-graph/internal/application/graph"
	"knowledge-graph/internal/application/recommendation"
	"knowledge-graph/internal/config"
	graphDomain "knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/nlp"
	"knowledge-graph/internal/infrastructure/queue"
	"knowledge-graph/internal/infrastructure/queue/tasks"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("FATAL: Failed to load configuration: %v", err)
		os.Exit(1)
	}
	log.Printf("Worker config loaded: DatabaseURL=%s, RedisURL=%s, NLPServiceURL=%s",
		maskURL(cfg.DatabaseURL), cfg.RedisURL, cfg.NLPServiceURL)
	// Инициализация БД
	db.Init()
	if db.DB == nil {
		log.Fatal("database connection is nil")
	}
	log.Println("Worker: Connected to PostgreSQL")

	// Redis для кэша
	redisAddr := cfg.RedisURL
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	log.Println("Starting worker, connecting to Redis at", redisAddr)
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
	}()

	// Репозитории
	noteRepo := postgres.NewNoteRepository(db.DB, redisClient)
	keywordRepo := postgres.NewKeywordRepository(db.DB)
	embeddingRepo := postgres.NewEmbeddingRepository(db.DB)

	// URL Python-сервиса (внутри Docker – nlp:5000, локально – localhost:5000)
	nlpURL := cfg.NLPServiceURL
	if nlpURL == "" {
		nlpURL = "http://localhost:5000"
	}
	nlpClient := nlp.NewNLPClient(nlpURL, redisClient, 24*time.Hour)

	// Воркер (обработчик задач)
	worker := queue.NewWorker(noteRepo, keywordRepo, embeddingRepo, nlpClient)

	// Graph traversal service for recommendations
	linkRepo := postgres.NewLinkRepository(db.DB)
	neighborLoader := graph.NewNeighborLoader(linkRepo, noteRepo)

	// Create keyword similarity strategy from config
	keywordSimilarity, err := recommendation.NewKeywordSimilarity(
		cfg.RecommendationKeywordSimilarityMethod,
		cfg.RecommendationKeywordTverskyAlpha,
		cfg.RecommendationKeywordTverskyBeta,
	)
	if err != nil {
		log.Printf("FATAL: Invalid keyword similarity method: %v", err)
		os.Exit(1)
	}
	log.Printf("Worker: Using keyword similarity method: %s", cfg.RecommendationKeywordSimilarityMethod)

	// Create keyword matcher with the configured similarity
	keywordMatcher := recommendation.NewKeywordMatcherImpl(keywordRepo, keywordSimilarity)

	traversalSvc := graphDomain.NewTraversalServiceWithWeights(
		neighborLoader,
		cfg.RecommendationDepth,
		cfg.RecommendationDecay,
		cfg.BFSAggregation,
		cfg.BFSNormalize,
		cfg.RecommendationAlpha,
		cfg.RecommendationBeta,
		cfg.RecommendationGamma,
	)

	// Set keyword matcher if gamma > 0
	if cfg.RecommendationGamma > 0 {
		traversalSvc.SetKeywordMatcher(keywordMatcher)
		log.Printf("Worker: Keyword component enabled (gamma=%.2f)", cfg.RecommendationGamma)
	} else {
		log.Printf("Worker: Keyword component disabled (gamma=%.2f)", cfg.RecommendationGamma)
	}

	// Refresh service for background recommendation calculation
	refreshSvc := recommendation.NewRefreshService(db.DB, redisClient, traversalSvc, cfg.RecommendationTopN)

	// Asynq сервер с конфигурацией
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: cfg.AsynqConcurrency,
			Queues: map[string]int{
				"default": cfg.AsynqQueueDefault,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeExtractKeywords, worker.HandleExtractKeywords)
	mux.HandleFunc(queue.TypeComputeEmbedding, worker.HandleComputeEmbedding)
	mux.HandleFunc(tasks.TypeRefreshRecommendations, func(ctx context.Context, t *asynq.Task) error {
		return tasks.HandleRefreshRecommendations(ctx, t, refreshSvc)
	})

	log.Printf("Worker started with config: Concurrency=%d, QueueMaxLen=%d", cfg.AsynqConcurrency, cfg.AsynqQueueMaxLen)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Println("Worker started, listening for tasks...")
		if err := srv.Run(mux); err != nil {
			errChan <- err
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		// Stop accepting new tasks and wait for ongoing tasks to complete
		srv.Stop()
		// Close resources
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
		log.Println("Worker shut down gracefully")
	case err := <-errChan:
		log.Fatalf("Worker error: %v", err)
	}
}

// maskURL hides sensitive info in URLs for logging
func maskURL(url string) string {
	if url == "" {
		return "(empty)"
	}
	// Simple masking - replace password with ***
	if idx := findSubstring(url, "://"); idx != -1 {
		remaining := url[idx+3:]
		if atIdx := findSubstring(remaining, "@"); atIdx != -1 {
			// Has credentials
			if colonIdx := findSubstring(remaining[:atIdx], ":"); colonIdx != -1 {
				return url[:idx+3+colonIdx+1] + "***" + url[idx+3+atIdx:]
			}
		}
	}
	return url
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
