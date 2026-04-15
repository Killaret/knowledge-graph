# Knowledge Graph - Полный код проекта

> **Исключено**: NLP-сервис (Python) по запросу

---

## Содержание

1. [Backend (Go)](#1-backend-go)
2. [Frontend (Svelte/TypeScript)](#2-frontend-sveltetypescript)
3. [Database Migrations](#3-database-migrations)
4. [Tests](#4-tests)
5. [Docker & Configuration](#5-docker--configuration)

---

## 1. Backend (Go)

### 1.1 Main Entry Points

#### `backend/cmd/server/main.go`

```go
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
```

#### `backend/cmd/worker/main.go`

```go
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
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Printf("Error closing redis client: %v", err)
		}
	}()

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
```

#### `backend/go.mod`

```go
module knowledge-graph

go 1.26.1

require (
	github.com/gin-gonic/gin v1.12.0
	github.com/google/uuid v1.6.0
	github.com/hibiken/asynq v0.26.0
	github.com/joho/godotenv v1.5.1
	github.com/pgvector/pgvector-go v0.3.0
	github.com/redis/go-redis/v9 v9.18.0
	github.com/stretchr/testify v1.11.1
	gorm.io/datatypes v1.2.7
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)
```

### 1.2 Domain Layer

#### `backend/internal/config/config.go`

```go
package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Сервер
	ServerPort string

	// База данных
	DatabaseURL string

	// Redis
	RedisURL string

	// NLP
	NLPServiceURL string

	// Рекомендации (граф)
	RecommendationAlpha      float64
	RecommendationBeta       float64
	RecommendationDepth      int
	RecommendationDecay      float64
	RecommendationCacheTTL   time.Duration
	EmbeddingSimilarityLimit int

	// Загрузка графа (визуализация)
	GraphLoadDepth int
}

// Load загружает конфигурацию из переменных окружения.
func Load() *Config {
	return &Config{
		ServerPort:               getEnv("SERVER_PORT", "8080"),
		DatabaseURL:              mustGetEnv("DATABASE_URL"),
		RedisURL:                 getEnv("REDIS_URL", "localhost:6379"),
		NLPServiceURL:            getEnv("NLP_SERVICE_URL", "http://localhost:5000"),
		RecommendationAlpha:      getFloatEnv("RECOMMENDATION_ALPHA", 0.5),
		RecommendationBeta:       getFloatEnv("RECOMMENDATION_BETA", 0.5),
		RecommendationDepth:      getIntEnv("RECOMMENDATION_DEPTH", 3),
		RecommendationDecay:      getFloatEnv("RECOMMENDATION_DECAY", 0.5),
		RecommendationCacheTTL:   time.Duration(getIntEnv("RECOMMENDATION_CACHE_TTL_SECONDS", 300)) * time.Second,
		EmbeddingSimilarityLimit: getIntEnv("EMBEDDING_SIMILARITY_LIMIT", 30),
		GraphLoadDepth:           getIntEnv("GRAPH_LOAD_DEPTH", 2),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	if str := os.Getenv(key); str != "" {
		if val, err := strconv.Atoi(str); err == nil {
			return val
		}
	}
	return defaultValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	if str := os.Getenv(key); str != "" {
		if val, err := strconv.ParseFloat(str, 64); err == nil {
			return val
		}
	}
	return defaultValue
}
```

#### `backend/internal/domain/note/entity.go`

```go
package note

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	id        uuid.UUID
	title     Title
	content   Content
	metadata  Metadata
	createdAt time.Time
	updatedAt time.Time
}

func NewNote(title Title, content Content, metadata Metadata) *Note {
	now := time.Now()
	return &Note{
		id:        uuid.New(),
		title:     title,
		content:   content,
		metadata:  metadata,
		createdAt: now,
		updatedAt: now,
	}
}

// ReconstructNote восстанавливает заметку из сохранённых данных
func ReconstructNote(id uuid.UUID, title Title, content Content, metadata Metadata, createdAt, updatedAt time.Time) *Note {
	return &Note{
		id:        id,
		title:     title,
		content:   content,
		metadata:  metadata,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Геттеры
func (n *Note) ID() uuid.UUID {
	return n.id
}

func (n *Note) Title() Title {
	return n.title
}

func (n *Note) Content() Content {
	return n.content
}

func (n *Note) Metadata() Metadata {
	return n.metadata
}

func (n *Note) CreatedAt() time.Time {
	return n.createdAt
}

func (n *Note) UpdatedAt() time.Time {
	return n.updatedAt
}

// Методы изменения с валидацией
func (n *Note) UpdateTitle(newTitle Title) error {
	if newTitle.String() == "" {
		return fmt.Errorf("cannot update with empty title")
	}
	n.title = newTitle
	n.updatedAt = time.Now()
	return nil
}

func (n *Note) UpdateContent(newContent Content) error {
	if newContent.String() == "" {
		return fmt.Errorf("cannot update with empty content")
	}
	n.content = newContent
	n.updatedAt = time.Now()
	return nil
}

func (n *Note) UpdateMetadata(newMetadata Metadata) error {
	if newMetadata.Value() == nil {
		return fmt.Errorf("cannot update with nil metadata")
	}
	n.metadata = newMetadata
	n.updatedAt = time.Now()
	return nil
}
```

#### `backend/internal/domain/note/value_objects.go`

```go
package note

import (
	"errors"
	"strings"
)

// Title — заголовок заметки (не может быть пустым, макс 200 символов)
type Title struct {
	value string
}

func NewTitle(value string) (Title, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) == 0 {
		return Title{}, errors.New("title cannot be empty")
	}
	if len(trimmed) > 200 {
		return Title{}, errors.New("title too long (max 200 characters)")
	}
	return Title{value: trimmed}, nil
}

func (t Title) String() string {
	return t.value
}

// Content — содержимое заметки (простой текст)
type Content struct {
	value string
}

func NewContent(value string) (Content, error) {
	if len(value) > 10000 {
		return Content{}, errors.New("content too long (max 10000 characters)")
	}
	return Content{value: value}, nil
}

func (c Content) String() string {
	return c.value
}

// Metadata — дополнительные данные заметки (теги, статус и т.п.)
type Metadata struct {
	value map[string]interface{}
}

func NewMetadata(value map[string]interface{}) (Metadata, error) {
	return Metadata{value: value}, nil
}

func (m Metadata) Value() map[string]interface{} {
	return m.value
}
```

#### `backend/internal/domain/note/repository.go`

```go
package note

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Common repository errors
var (
	ErrNoteNotFound = errors.New("note not found")
)

type Repository interface {
	Save(ctx context.Context, note *Note) error
	FindByID(ctx context.Context, id uuid.UUID) (*Note, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Note, int64, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*Note, int64, error)
	FindAll(ctx context.Context) ([]*Note, error)
}
```

#### `backend/internal/domain/link/entity.go`

```go
package link

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	id           uuid.UUID
	sourceNoteID uuid.UUID
	targetNoteID uuid.UUID
	linkType     LinkType
	weight       Weight
	metadata     Metadata
	createdAt    time.Time
}

func NewLink(sourceID, targetID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata) *Link {
	return &Link{
		id:           uuid.New(),
		sourceNoteID: sourceID,
		targetNoteID: targetID,
		linkType:     linkType,
		weight:       weight,
		metadata:     metadata,
		createdAt:    time.Now(),
	}
}

// ReconstructLink восстанавливает связь из сохранённых данных
func ReconstructLink(id uuid.UUID, sourceID, targetID uuid.UUID, linkType LinkType, weight Weight, metadata Metadata, createdAt time.Time) *Link {
	return &Link{
		id:           id,
		sourceNoteID: sourceID,
		targetNoteID: targetID,
		linkType:     linkType,
		weight:       weight,
		metadata:     metadata,
		createdAt:    createdAt,
	}
}

// Геттеры
func (l *Link) ID() uuid.UUID {
	return l.id
}

func (l *Link) SourceNoteID() uuid.UUID {
	return l.sourceNoteID
}

func (l *Link) TargetNoteID() uuid.UUID {
	return l.targetNoteID
}

func (l *Link) LinkType() LinkType {
	return l.linkType
}

func (l *Link) Weight() Weight {
	return l.weight
}

func (l *Link) Metadata() Metadata {
	return l.metadata
}

func (l *Link) CreatedAt() time.Time {
	return l.createdAt
}

// UpdateWeight обновляет вес связи
func (l *Link) UpdateWeight(newWeight Weight) {
	l.weight = newWeight
}
```

#### `backend/internal/domain/link/value_objects.go`

```go
package link

import "errors"

type LinkType struct {
	value string
}

func NewLinkType(value string) (LinkType, error) {
	switch value {
	case "reference", "dependency", "related", "custom":
		return LinkType{value: value}, nil
	default:
		return LinkType{}, errors.New("invalid link type")
	}
}

func (t LinkType) String() string {
	return t.value
}

type Weight struct {
	value float64
}

func NewWeight(value float64) (Weight, error) {
	if value < 0 || value > 1 {
		return Weight{}, errors.New("weight must be between 0 and 1")
	}
	return Weight{value: value}, nil
}

func (w Weight) Value() float64 {
	return w.value
}

// Metadata — дополнительные данные связи
type Metadata struct {
	value map[string]interface{}
}

func NewMetadata(value map[string]interface{}) (Metadata, error) {
	return Metadata{value: value}, nil
}

func (m Metadata) Value() map[string]interface{} {
	return m.value
}
```

#### `backend/internal/domain/link/repository.go`

```go
package link

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, link *Link) error
	FindByID(ctx context.Context, id uuid.UUID) (*Link, error)
	FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*Link, error)
	FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*Link, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteBySource(ctx context.Context, sourceID uuid.UUID) error
	FindAll(ctx context.Context) ([]*Link, error)
}
```

#### `backend/internal/domain/graph/traversal.go`

```go
package graph

import (
	"context"
	"sort"

	"github.com/google/uuid"
)

type Node struct {
	ID uuid.UUID
}

type Edge struct {
	From   uuid.UUID
	To     uuid.UUID
	Weight float64
}

type NeighborLoader interface {
	GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]Edge, error)
}

type TraversalService struct {
	loader NeighborLoader
	depth  int
	decay  float64
}

func NewTraversalService(loader NeighborLoader, depth int, decay float64) *TraversalService {
	return &TraversalService{
		loader: loader,
		depth:  depth,
		decay:  decay,
	}
}

type SuggestionResult struct {
	NodeID uuid.UUID
	Score  float64
}

func (s *TraversalService) GetSuggestions(ctx context.Context, startID uuid.UUID, topN int) ([]SuggestionResult, error) {
	depth := s.depth
	decay := s.decay
	if depth < 1 {
		depth = 1
	}
	if decay <= 0 || decay > 1 {
		decay = 0.5
	}

	scores := make(map[uuid.UUID]float64)
	type bfsItem struct {
		nodeID uuid.UUID
		weight float64
		depth  int
	}
	queue := []bfsItem{{nodeID: startID, weight: 1.0, depth: 0}}
	visited := map[uuid.UUID]bool{startID: true}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.depth >= depth {
			continue
		}

		neighbors, err := s.loader.GetNeighbors(ctx, current.nodeID)
		if err != nil {
			return nil, err
		}

		for _, edge := range neighbors {
			if edge.To == startID {
				continue
			}
			newWeight := current.weight * edge.Weight
			if current.depth > 0 {
				newWeight *= decay
			}
			if _, ok := scores[edge.To]; !ok {
				scores[edge.To] = newWeight
			} else {
				scores[edge.To] += newWeight
			}
			if !visited[edge.To] {
				visited[edge.To] = true
				queue = append(queue, bfsItem{
					nodeID: edge.To,
					weight: newWeight,
					depth:  current.depth + 1,
				})
			}
		}
	}

	results := make([]SuggestionResult, 0, len(scores))
	for id, score := range scores {
		results = append(results, SuggestionResult{NodeID: id, Score: score})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topN {
		results = results[:topN]
	}
	return results, nil
}
```

### 1.3 Infrastructure Layer

#### `backend/internal/infrastructure/db/db.go`

```go
package db

import (
	"log"
	"os"

	"knowledge-graph/internal/infrastructure/db/postgres"

	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var err error
	DB, err = gorm.Open(pgdriver.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	log.Println("Connected to PostgreSQL")

	// Run migrations
	if err := DB.AutoMigrate(
		&postgres.NoteModel{},
		&postgres.LinkModel{},
		&postgres.NoteKeywordModel{},
		&postgres.NoteEmbeddingModel{},
		&postgres.NoteTagModel{},
		&postgres.UserModel{},
		&postgres.NoteLikeModel{},
		&postgres.SuggestionFeedbackModel{},
		&postgres.ShareLinkModel{},
	); err != nil {
		log.Fatal("failed to run migrations:", err)
	}
	log.Println("Database migrations completed successfully")
}
```

#### `backend/internal/infrastructure/db/postgres/models.go`

```go
package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
	"gorm.io/datatypes"
)

// NoteModel — модель заметки
type NoteModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Title        string         `gorm:"not null"`
	Content      string         `gorm:"type:text"`
	Metadata     datatypes.JSON `gorm:"type:jsonb"`
	SearchVector string         `gorm:"column:search_vector;type:tsvector;->"` // read-only
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (NoteModel) TableName() string {
	return "notes"
}

// LinkModel — связь между заметками
type LinkModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	SourceNoteID uuid.UUID      `gorm:"type:uuid;not null;index"`
	TargetNoteID uuid.UUID      `gorm:"type:uuid;not null;index"`
	LinkType     string         `gorm:"default:'reference'"`
	Weight       float64        `gorm:"default:1.0"`
	Metadata     datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt    time.Time

	SourceNote NoteModel `gorm:"foreignKey:SourceNoteID"`
	TargetNote NoteModel `gorm:"foreignKey:TargetNoteID"`
}

func (LinkModel) TableName() string {
	return "links"
}

// NoteKeywordModel — ключевые слова заметки
type NoteKeywordModel struct {
	NoteID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	Keyword string    `gorm:"primaryKey"`
	Weight  float64

	Note NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteKeywordModel) TableName() string {
	return "note_keywords"
}

// NoteEmbeddingModel — векторное представление заметки
type NoteEmbeddingModel struct {
	NoteID    uuid.UUID       `gorm:"type:uuid;primaryKey"`
	Embedding pgvector.Vector `gorm:"type:vector(384)"`
	UpdatedAt time.Time

	Note NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteEmbeddingModel) TableName() string {
	return "note_embeddings"
}

// NoteTagModel — теги заметки
type NoteTagModel struct {
	NoteID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Tag    string    `gorm:"primaryKey"`
	Note   NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteTagModel) TableName() string { return "note_tags" }

// UserModel — пользователь
type UserModel struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Login        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"default:'user'"`
	CreatedAt    time.Time
}

func (UserModel) TableName() string { return "users" }

// NoteLikeModel — лайк/дизлайк
type NoteLikeModel struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	NoteID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	LikeType  string    `gorm:"not null"`
	CreatedAt time.Time
	User      UserModel `gorm:"foreignKey:UserID"`
	Note      NoteModel `gorm:"foreignKey:NoteID"`
}

func (NoteLikeModel) TableName() string { return "note_likes" }

// SuggestionFeedbackModel — обратная связь
type SuggestionFeedbackModel struct {
	UserID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	SourceNoteID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	SuggestedNoteID uuid.UUID `gorm:"type:uuid;primaryKey"`
	FeedbackType    string    `gorm:"not null"`
	CreatedAt       time.Time
	User            UserModel `gorm:"foreignKey:UserID"`
	SourceNote      NoteModel `gorm:"foreignKey:SourceNoteID"`
	SuggestedNote   NoteModel `gorm:"foreignKey:SuggestedNoteID"`
}

func (SuggestionFeedbackModel) TableName() string { return "suggestion_feedback" }

// ShareLinkModel — расшаривание
type ShareLinkModel struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	NoteID         uuid.UUID `gorm:"type:uuid;not null;index"`
	SharedByUserID uuid.UUID `gorm:"type:uuid;not null"`
	ShareToken     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	Note           NoteModel `gorm:"foreignKey:NoteID"`
	SharedBy       UserModel `gorm:"foreignKey:SharedByUserID"`
}

func (ShareLinkModel) TableName() string { return "share_links" }
```

#### `backend/internal/infrastructure/db/postgres/note_repo.go`

```go
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type NoteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

func (r *NoteRepository) Save(ctx context.Context, n *note.Note) error {
	var existing NoteModel
	err := r.db.WithContext(ctx).Where("id = ?", n.ID()).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model, err := toGormNote(n)
		if err != nil {
			return err
		}
		return r.db.WithContext(ctx).Create(&model).Error
	}
	if err != nil {
		return err
	}
	model, err := toGormNote(n)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(model).Error
}

func (r *NoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
	var model NoteModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainNote(&model)
}

func (r *NoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&NoteModel{}, "id = ?", id).Error
}

func (r *NoteRepository) FindAll(ctx context.Context) ([]*note.Note, error) {
	var models []NoteModel
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainNotes(models), nil
}

func (r *NoteRepository) List(ctx context.Context, limit, offset int) ([]*note.Note, int64, error) {
	var models []NoteModel
	var total int64

	db := r.db.WithContext(ctx).Model(&NoteModel{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	return toDomainNotes(models), total, nil
}

func (r *NoteRepository) Search(ctx context.Context, query string, limit, offset int) ([]*note.Note, int64, error) {
	var models []NoteModel
	var total int64

	if query != "" {
		db := r.db.WithContext(ctx).Model(&NoteModel{})

		db = db.Where(`
			search_vector @@ plainto_tsquery('russian', ?) OR 
			search_vector @@ plainto_tsquery('simple', ?)
		`, query, query)

		db = db.Order(fmt.Sprintf(`
			COALESCE(ts_rank(search_vector, plainto_tsquery('russian', '%s')), 0) +
			COALESCE(ts_rank(search_vector, plainto_tsquery('simple', '%s')), 0) DESC
		`, query, query))

		if err := db.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		if total > 0 {
			err := db.Limit(limit).Offset(offset).Find(&models).Error
			if err != nil {
				return nil, 0, err
			}
			return toDomainNotes(models), total, nil
		}

		dbLike := r.db.WithContext(ctx).Model(&NoteModel{})
		dbLike = dbLike.Where(`
			title ILIKE ? OR content ILIKE ?
		`, "%"+query+"%", "%"+query+"%")
		dbLike = dbLike.Order("created_at DESC")

		if err := dbLike.Count(&total).Error; err != nil {
			return nil, 0, err
		}

		err := dbLike.Limit(limit).Offset(offset).Find(&models).Error
		if err != nil {
			return nil, 0, err
		}
		return toDomainNotes(models), total, nil
	}

	db := r.db.WithContext(ctx).Model(&NoteModel{}).Order("created_at DESC")
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := db.Limit(limit).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}
	return toDomainNotes(models), total, nil
}

func toGormNote(n *note.Note) (NoteModel, error) {
	metadataJSON, err := json.Marshal(n.Metadata().Value())
	if err != nil {
		return NoteModel{}, err
	}
	return NoteModel{
		ID:        n.ID(),
		Title:     n.Title().String(),
		Content:   n.Content().String(),
		Metadata:  datatypes.JSON(metadataJSON),
		CreatedAt: n.CreatedAt(),
		UpdatedAt: n.UpdatedAt(),
	}, nil
}

func toDomainNote(m *NoteModel) (*note.Note, error) {
	title, err := note.NewTitle(m.Title)
	if err != nil {
		return nil, err
	}
	content, err := note.NewContent(m.Content)
	if err != nil {
		return nil, err
	}
	var metadataMap map[string]interface{}
	if len(m.Metadata) > 0 {
		if err := json.Unmarshal(m.Metadata, &metadataMap); err != nil {
			return nil, err
		}
	}
	metadata, err := note.NewMetadata(metadataMap)
	if err != nil {
		return nil, err
	}
	return note.ReconstructNote(m.ID, title, content, metadata, m.CreatedAt, m.UpdatedAt), nil
}

func toDomainNotes(models []NoteModel) []*note.Note {
	result := make([]*note.Note, 0, len(models))
	for _, m := range models {
		n, err := toDomainNote(&m)
		if err != nil {
			continue
		}
		result = append(result, n)
	}
	return result
}
```

#### `backend/internal/infrastructure/db/postgres/link_repo.go`

```go
package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"knowledge-graph/internal/domain/link"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LinkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Save(ctx context.Context, l *link.Link) error {
	var existing LinkModel
	err := r.db.WithContext(ctx).Where("id = ?", l.ID()).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		model, err := toGormLink(l)
		if err != nil {
			return err
		}
		return r.db.WithContext(ctx).Create(&model).Error
	}
	if err != nil {
		return err
	}
	model, err := toGormLink(l)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&existing).Updates(model).Error
}

func (r *LinkRepository) FindByID(ctx context.Context, id uuid.UUID) (*link.Link, error) {
	var model LinkModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainLink(&model)
}

func (r *LinkRepository) FindBySource(ctx context.Context, sourceID uuid.UUID) ([]*link.Link, error) {
	var models []LinkModel
	err := r.db.WithContext(ctx).Where("source_note_id = ?", sourceID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

func (r *LinkRepository) FindByTarget(ctx context.Context, targetID uuid.UUID) ([]*link.Link, error) {
	var models []LinkModel
	err := r.db.WithContext(ctx).Where("target_note_id = ?", targetID).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

func (r *LinkRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&LinkModel{}, "id = ?", id).Error
}

func (r *LinkRepository) DeleteBySource(ctx context.Context, sourceID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("source_note_id = ?", sourceID).Delete(&LinkModel{}).Error
}

func (r *LinkRepository) FindAll(ctx context.Context) ([]*link.Link, error) {
	var models []LinkModel
	err := r.db.WithContext(ctx).Find(&models).Error
	if err != nil {
		return nil, err
	}
	return toDomainLinks(models), nil
}

func toGormLink(l *link.Link) (LinkModel, error) {
	metadataJSON, err := json.Marshal(l.Metadata().Value())
	if err != nil {
		return LinkModel{}, err
	}
	return LinkModel{
		ID:           l.ID(),
		SourceNoteID: l.SourceNoteID(),
		TargetNoteID: l.TargetNoteID(),
		LinkType:     l.LinkType().String(),
		Weight:       l.Weight().Value(),
		Metadata:     datatypes.JSON(metadataJSON),
		CreatedAt:    l.CreatedAt(),
	}, nil
}

func toDomainLink(m *LinkModel) (*link.Link, error) {
	linkType, err := link.NewLinkType(m.LinkType)
	if err != nil {
		return nil, err
	}
	weight, err := link.NewWeight(m.Weight)
	if err != nil {
		return nil, err
	}
	var metadataMap map[string]interface{}
	if len(m.Metadata) > 0 {
		if err := json.Unmarshal(m.Metadata, &metadataMap); err != nil {
			return nil, err
		}
	}
	metadata, err := link.NewMetadata(metadataMap)
	if err != nil {
		return nil, err
	}
	return link.ReconstructLink(m.ID, m.SourceNoteID, m.TargetNoteID, linkType, weight, metadata, m.CreatedAt), nil
}

func toDomainLinks(models []LinkModel) []*link.Link {
	result := make([]*link.Link, 0, len(models))
	for _, m := range models {
		l, err := toDomainLink(&m)
		if err != nil {
			continue
		}
		result = append(result, l)
	}
	return result
}
```
