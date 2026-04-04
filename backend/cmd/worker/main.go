package main

import (
	"log"
	"os"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/nlp"
	"knowledge-graph/internal/infrastructure/queue"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	// Инициализация БД
	db.Init()
	if db.DB == nil {
		log.Fatal("database connection is nil")
	}
	log.Println("Worker: Connected to PostgreSQL")

	// Репозитории
	noteRepo := postgres.NewNoteRepository(db.DB)
	keywordRepo := postgres.NewKeywordRepository(db.DB)
	embeddingRepo := postgres.NewEmbeddingRepository(db.DB)

	// NLP клиент (с Redis для кэша)
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	log.Println("Starting worker, connecting to Redis at", redisAddr)
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer redisClient.Close()

	// URL Python-сервиса (внутри Docker – nlp:5000, локально – localhost:5000)
	nlpURL := os.Getenv("NLP_SERVICE_URL")
	if nlpURL == "" {
		nlpURL = "http://localhost:5000"
	}
	nlpClient := nlp.NewNLPClient(nlpURL, redisClient, 24*time.Hour)

	// Воркер (обработчик задач)
	worker := queue.NewWorker(noteRepo, keywordRepo, embeddingRepo, nlpClient)

	// Asynq сервер
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default": 1,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeExtractKeywords, worker.HandleExtractKeywords)
	mux.HandleFunc(queue.TypeComputeEmbedding, worker.HandleComputeEmbedding)

	log.Println("Worker started, listening for tasks...")
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
