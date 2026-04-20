package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
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
	cfg := config.Load()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
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
	nlpURL := os.Getenv("NLP_SERVICE_URL")
	if nlpURL == "" {
		nlpURL = "http://localhost:5000"
	}
	nlpClient := nlp.NewNLPClient(nlpURL, redisClient, 24*time.Hour)

	// Воркер (обработчик задач)
	worker := queue.NewWorker(noteRepo, keywordRepo, embeddingRepo, nlpClient)

	// Graph traversal service for recommendations
	linkRepo := postgres.NewLinkRepository(db.DB)
	neighborLoader := graph.NewNeighborLoader(linkRepo, noteRepo)
	traversalSvc := graphDomain.NewTraversalService(
		neighborLoader,
		cfg.RecommendationDepth,
		cfg.RecommendationDecay,
		cfg.BFSAggregation,
		cfg.BFSNormalize,
	)

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

	log.Println("Worker started, listening for tasks...")
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
