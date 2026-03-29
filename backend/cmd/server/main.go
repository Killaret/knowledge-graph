package main

import (
	"log"
	"os"

	"knowledge-graph/internal/infrastructure/db"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	db.Init() // 👈 подключение к БД

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

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

	r.Run(":" + port)

}
