package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/nlp"
)

// newHealthHandler returns a health check handler that checks all dependencies
func newHealthHandler(database *gorm.DB, redisClient *redis.Client, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		health := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		}
		status := http.StatusOK

		// Check database
		sqlDB, err := database.DB()
		if err != nil {
			health["database"] = gin.H{"status": "unhealthy", "error": err.Error()}
			status = http.StatusServiceUnavailable
		} else if err := sqlDB.Ping(); err != nil {
			health["database"] = gin.H{"status": "unhealthy", "error": err.Error()}
			status = http.StatusServiceUnavailable
		} else {
			health["database"] = gin.H{"status": "healthy"}
		}

		// Check Redis
		if redisClient != nil {
			if err := redisClient.Ping(ctx).Err(); err != nil {
				health["redis"] = gin.H{"status": "unhealthy", "error": err.Error()}
				status = http.StatusServiceUnavailable
			} else {
				health["redis"] = gin.H{"status": "healthy"}
			}
		} else {
			health["redis"] = gin.H{"status": "disabled"}
		}

		// Check NLP service
		nlpClient := nlp.NewNLPClient(cfg.NLPServiceURL, redisClient, cfg.RecommendationCacheTTL)
		if err := nlpClient.HealthCheck(ctx); err != nil {
			health["nlp"] = gin.H{"status": "unhealthy", "error": err.Error()}
			// Don't mark as unhealthy if NLP is optional
		} else {
			health["nlp"] = gin.H{"status": "healthy"}
		}

		c.JSON(status, health)
	}
}
