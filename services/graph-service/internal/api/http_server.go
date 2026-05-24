package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/db"
	"knowledge-graph-graph-service/internal/engine"
)

// HTTPServer handles HTTP requests for the graph service
type HTTPServer struct {
	cache    *cache.RedisCache
	postgres db.PostgresClient
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(postgres db.PostgresClient, cache *cache.RedisCache) *HTTPServer {
	return &HTTPServer{
		cache:    cache,
		postgres: postgres,
	}
}

// GraphData represents the response structure
type GraphData struct {
	Nodes []*ProtoLayoutNode `json:"nodes"`
	Links []*ProtoLayoutLink `json:"links"`
}

// GraphApiResponse wraps the graph data with metadata
type GraphApiResponse struct {
	Data GraphData `json:"data"`
	Meta *struct {
		TotalNodes int `json:"total_nodes,omitempty"`
		TotalLinks int `json:"total_links,omitempty"`
	} `json:"meta,omitempty"`
}

// GetNoteGraphHandler handles GET /api/v1/graph/note/:id
func (s *HTTPServer) GetNoteGraphHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract note ID from path
	noteID := r.URL.Path[len("/api/v1/graph/note/"):]
	if noteID == "" {
		http.Error(w, "Note ID is required", http.StatusBadRequest)
		return
	}

	// Parse depth parameter
	depth := 2
	if depthStr := r.URL.Query().Get("depth"); depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d > 0 {
			depth = d
		}
	}

	// Parse user_id parameter
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "public"
	}

	// Try cache first
	cacheKey := fmt.Sprintf("note:%s:depth-%d", noteID, depth)
	if cached, hash, err := s.cache.LoadNoteLayout(ctx, noteID, depth); err == nil && cached != nil {
		log.Printf("[GraphService] Cache hit for %s", cacheKey)
		s.sendGraphData(w, cached, hash)
		return
	}

	// Load from database
	notes, links, err := s.postgres.GetNotes(ctx, noteID, depth)
	if err != nil {
		log.Printf("[GraphService] Failed to load notes: %v", err)
		http.Error(w, "Failed to load graph", http.StatusInternalServerError)
		return
	}

	// Generate layout
	layout := engine.Layout2D(notes, links, noteID)
	hash := computeLayoutHash(layout)

	// Cache the result
	if err := s.cache.SaveNoteLayout(ctx, noteID, depth, layout, hash); err != nil {
		log.Printf("[GraphService] Failed to cache layout: %v", err)
	}

	s.sendGraphData(w, layout, hash)
}

// GetFullGraphHandler handles GET /api/v1/graph/full
func (s *HTTPServer) GetFullGraphHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse limit parameter
	limit := 1000
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Parse user_id parameter
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "public"
	}

	// Try cache first
	if cached, hash, err := s.cache.LoadFullLayout(ctx, userID); err == nil && cached != nil {
		log.Printf("[GraphService] Cache hit for full graph user=%s", userID)
		s.sendGraphData(w, cached, hash)
		return
	}

	// Load from database
	notes, links, err := s.postgres.GetNotes(ctx, "", 0)
	if err != nil {
		log.Printf("[GraphService] Failed to load full graph: %v", err)
		http.Error(w, "Failed to load full graph", http.StatusInternalServerError)
		return
	}

	// Apply limit
	if limit > 0 && len(notes) > limit {
		notes = notes[:limit]
	}

	// Generate layout
	layout := engine.Layout3D(notes, links)
	hash := computeLayoutHash(layout)

	// Cache the result
	if err := s.cache.SaveFullLayout(ctx, userID, layout, hash); err != nil {
		log.Printf("[GraphService] Failed to cache full layout: %v", err)
	}

	s.sendGraphData(w, layout, hash)
}

// sendGraphData sends the layout data as JSON response
func (s *HTTPServer) sendGraphData(w http.ResponseWriter, layout *engine.LayoutResponse, hash string) {
	response := GraphApiResponse{
		Data: GraphData{
			Nodes: convertLayoutNodes(layout.Nodes),
			Links: convertLayoutLinks(layout.Links),
		},
		Meta: &struct {
			TotalNodes int `json:"total_nodes,omitempty"`
			TotalLinks int `json:"total_links,omitempty"`
		}{
			TotalNodes: len(layout.Nodes),
			TotalLinks: len(layout.Links),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Layout-Hash", hash)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[GraphService] Failed to encode response: %v", err)
	}
}

// RegisterHTTPHandlers registers all HTTP handlers
func RegisterHTTPHandlers(mux *http.ServeMux, service GraphServiceServer) {
	if gs, ok := service.(*graphService); ok {
		httpServer := NewHTTPServer(gs.postgres, gs.cache)

		// Health check
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Graph endpoints
		mux.HandleFunc("/api/v1/graph/note/", httpServer.GetNoteGraphHandler)
		mux.HandleFunc("/api/v1/graph/full", httpServer.GetFullGraphHandler)
	}
}
