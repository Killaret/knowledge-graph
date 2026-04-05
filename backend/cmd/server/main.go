package main

import (
	"log"
	"os"

	"knowledge-graph/internal/application/common"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/infrastructure/queue"
	"knowledge-graph/internal/interfaces/api/linkhandler"
	"knowledge-graph/internal/interfaces/api/notehandler"

	appGraph "knowledge-graph/internal/application/graph"
	graphQueries "knowledge-graph/internal/application/queries/graph"
	graphDomain "knowledge-graph/internal/domain/graph"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db.Init()
	if db.DB == nil {
		log.Fatal("database connection is nil")
	}
	log.Println("Connected to PostgreSQL")

	noteRepo := postgres.NewNoteRepository(db.DB)
	linkRepo := postgres.NewLinkRepository(db.DB)

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	log.Printf("Redis address: %s", redisAddr)

	// === 1. Инициализация очереди (asynq) ===
	var taskQueue common.TaskQueue
	asynqClient, err := queue.NewAsynqClient(redisAddr)
	if err != nil {
		log.Printf("WARNING: failed to create asynq client: %v", err)
	} else {
		log.Printf("Asynq client created successfully")
		taskQueue = asynqClient
		defer asynqClient.Close()
	}

	// === 2. Инициализация компонентов графа и рекомендаций ===
	neighborLoader := appGraph.NewNeighborLoader(linkRepo, noteRepo)
	traversalSvc := graphDomain.NewTraversalService(neighborLoader)

	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer redisClient.Close()

	suggestionsHandler := graphQueries.NewGetSuggestionsHandler(traversalSvc, noteRepo, redisClient, 5*time.Minute)

	// === 3. Создание хендлеров (теперь suggestionsHandler уже существует) ===
	noteHandler := notehandler.New(noteRepo, taskQueue, suggestionsHandler)
	linkHandler := linkhandler.New(linkRepo, noteRepo)

	// === 4. Роуты ===
	r := gin.Default()
	// ... health, db-check ...

	// Заметки
	r.POST("/notes", noteHandler.Create)
	r.GET("/notes/:id", noteHandler.Get)
	r.PUT("/notes/:id", noteHandler.Update)
	r.DELETE("/notes/:id", noteHandler.Delete)
	r.GET("/notes/:id/suggestions", noteHandler.GetSuggestions) // ← добавлен корректно

	// Связи
	r.POST("/links", linkHandler.Create)
	r.GET("/links/:id", linkHandler.Get)
	r.GET("/notes/:id/links", linkHandler.GetByNote)
	r.DELETE("/links/:id", linkHandler.Delete)
	r.DELETE("/notes/:id/links", linkHandler.DeleteByNote)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
