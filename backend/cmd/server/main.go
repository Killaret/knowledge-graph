package main

import (
	"log"
	"os"

	"knowledge-graph/internal/domain/note"
	"knowledge-graph/internal/infrastructure/db"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db.Init() // инициализация глобального DB

	// Проверка, что подключение установлено
	if db.DB == nil {
		log.Fatal("database connection is nil")
	}
	log.Println("Connected to PostgreSQL")

	// Репозитории
	noteRepo := postgres.NewNoteRepository(db.DB)

	// Обработчик (временно здесь)
	noteHandler := &NoteHandler{repo: noteRepo}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

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

	// Добавляем эндпоинт для создания заметки
	r.POST("/notes", noteHandler.Create)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

// NoteHandler — временный обработчик
type NoteHandler struct {
	repo note.Repository
}

type createNoteRequest struct {
	Title    string                 `json:"title" binding:"required"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (h *NoteHandler) Create(c *gin.Context) {
	var req createNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Создаём Value Objects
	title, err := note.NewTitle(req.Title)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	content, err := note.NewContent(req.Content)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	metadata, err := note.NewMetadata(req.Metadata)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Создаём доменную заметку
	newNote := note.NewNote(title, content, metadata)

	// Сохраняем через репозиторий
	if err := h.repo.Save(c.Request.Context(), newNote); err != nil {
		c.JSON(500, gin.H{"error": "failed to save note"})
		return
	}

	c.JSON(201, gin.H{
		"id":         newNote.ID(),
		"title":      newNote.Title().String(),
		"content":    newNote.Content().String(),
		"metadata":   newNote.Metadata().Value(),
		"created_at": newNote.CreatedAt(),
		"updated_at": newNote.UpdatedAt(),
	})
}
