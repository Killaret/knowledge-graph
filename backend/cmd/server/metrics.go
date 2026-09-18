package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"knowledge-graph/internal/infrastructure/db"
)

// newMetricsHandler returns a handler that exposes database connection pool
// statistics. Intended for admin-only diagnostics.
func newMetricsHandler(database *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"timestamp":     time.Now().UTC(),
			"database_pool": db.GetPoolStats(database),
		})
	}
}
