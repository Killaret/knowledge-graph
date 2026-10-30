package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// DBPinger abstracts a database connection that can be pinged.
type DBPinger interface {
	PingContext(ctx context.Context) error
}

// RedisPinger abstracts a Redis client that can be pinged.
type RedisPinger interface {
	Ping(ctx context.Context) error
}

// NLPHealthChecker abstracts an NLP service client that exposes a health check.
type NLPHealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// newHealthHandler returns a health check handler that checks all dependencies
func newHealthHandler(database DBPinger, redisClient RedisPinger, nlpClient NLPHealthChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		health := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		}
		status := http.StatusOK

		// Check database
		if database != nil {
			if err := database.PingContext(ctx); err != nil {
				health["database"] = gin.H{"status": "unhealthy", "error": err.Error()}
				status = http.StatusServiceUnavailable
			} else {
				health["database"] = gin.H{"status": "healthy"}
			}
		} else {
			health["database"] = gin.H{"status": "disabled"}
		}

		// Check Redis
		if redisClient != nil {
			if err := redisClient.Ping(ctx); err != nil {
				health["redis"] = gin.H{"status": "unhealthy", "error": err.Error()}
				status = http.StatusServiceUnavailable
			} else {
				health["redis"] = gin.H{"status": "healthy"}
			}
		} else {
			health["redis"] = gin.H{"status": "disabled"}
		}

		// Check NLP service
		if nlpClient != nil {
			if err := nlpClient.HealthCheck(ctx); err != nil {
				health["nlp"] = gin.H{"status": "unhealthy", "error": err.Error()}
				// Don't mark as unhealthy if NLP is optional
			} else {
				health["nlp"] = gin.H{"status": "healthy"}
			}
		} else {
			health["nlp"] = gin.H{"status": "disabled"}
		}

		c.JSON(status, health)
	}
}

// compile-time interface assertions
var _ DBPinger = (*sql.DB)(nil)
