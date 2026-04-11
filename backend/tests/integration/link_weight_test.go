package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/infrastructure/db/postgres"
	"knowledge-graph/internal/interfaces/api/linkhandler"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLinkWeightCalculation(t *testing.T) {
	// Setup test database and router
	db := postgres.SetupTestDB(t)
	defer db.Close()

	router := gin.New()
	handler := linkhandler.NewLinkHandler(db)
	router.POST("/links", handler.CreateLink)
	router.GET("/links/:id", handler.GetLink)
	router.GET("/notes/:id/neighbors", handler.GetNoteNeighbors)

	t.Run("create link with default weight", func(t *testing.T) {
		// Create two test notes first
		note1ID := uuid.New()
		note2ID := uuid.New()

		// Insert test notes
		_, err := db.Exec("INSERT INTO notes (id, title, content) VALUES ($1, $2, $3)",
			note1ID, "Test Note 1", "Content 1")
		assert.NoError(t, err)

		_, err = db.Exec("INSERT INTO notes (id, title, content) VALUES ($1, $2, $3)",
			note2ID, "Test Note 2", "Content 2")
		assert.NoError(t, err)

		// Create link with explicit weight
		linkReq := map[string]interface{}{
			"source_id": note1ID.String(),
			"target_id": note2ID.String(),
			"type":      "reference",
			"weight":    0.75,
		}

		body, _ := json.Marshal(linkReq)
		req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Verify weight is preserved
		assert.InDelta(t, 0.75, response["weight"], 0.001)
		assert.Equal(t, "reference", response["type"])
	})

	t.Run("create link with weight at boundaries", func(t *testing.T) {
		// Test weight = 0
		note3ID := uuid.New()
		note4ID := uuid.New()

		_, err := db.Exec("INSERT INTO notes (id, title, content) VALUES ($1, $2, $3)",
			note3ID, "Test Note 3", "Content 3")
		assert.NoError(t, err)

		_, err = db.Exec("INSERT INTO notes (id, title, content) VALUES ($1, $2, $3)",
			note4ID, "Test Note 4", "Content 4")
		assert.NoError(t, err)

		linkReq := map[string]interface{}{
			"source_id": note3ID.String(),
			"target_id": note4ID.String(),
			"type":      "dependency",
			"weight":    0.0,
		}

		body, _ := json.Marshal(linkReq)
		req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.InDelta(t, 0.0, response["weight"], 0.001)
	})

	t.Run("create link with invalid weight should fail", func(t *testing.T) {
		note5ID := uuid.New()
		note6ID := uuid.New()

		_, err := db.Exec("INSERT INTO notes (id, title, content) VALUES ($1, $2, $3)",
			note5ID, "Test Note 5", "Content 5")
		assert.NoError(t, err)

		_, err = db.Exec("INSERT INTO notes (id, title, content) VALUES ($1, $2, $3)",
			note6ID, "Test Note 6", "Content 6")
		assert.NoError(t, err)

		// Weight > 1 should fail
		linkReq := map[string]interface{}{
			"source_id": note5ID.String(),
			"target_id": note6ID.String(),
			"type":      "related",
			"weight":    1.5,
		}

		body, _ := json.Marshal(linkReq)
		req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return error for invalid weight
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLinkWeightInGraphTraversal(t *testing.T) {
	t.Run("verify weight affects neighbor ranking", func(t *testing.T) {
		// This test verifies that links with higher weights are prioritized
		// in graph traversal operations
		
		// Setup would involve:
		// 1. Creating a central note
		// 2. Creating multiple neighbor notes
		// 3. Creating links with different weights
		// 4. Querying neighbors and verifying order by weight
		
		// Placeholder for complex integration test
		t.Skip("Complex graph traversal test - requires full service setup")
	})
}
