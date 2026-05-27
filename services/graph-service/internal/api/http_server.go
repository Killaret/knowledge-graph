package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/db"
	"knowledge-graph-graph-service/internal/engine"
	graphservice "knowledge-graph-graph-service/proto"
)

// HTTPServer handles HTTP requests for the graph service
type HTTPServer struct {
	cache    *cache.RedisCache
	postgres db.PostgresClient
	limit    int
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(postgres db.PostgresClient, cache *cache.RedisCache, limit int) *HTTPServer {
	return &HTTPServer{
		cache:    cache,
		postgres: postgres,
		limit:    limit,
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
		TotalNodes int    `json:"total_nodes,omitempty"`
		TotalLinks int    `json:"total_links,omitempty"`
		Hash       string `json:"hash,omitempty"`
	} `json:"meta,omitempty"`
}

// GetNoteGraphHandler handles GET /api/v1/graph/note/:id
func (s *HTTPServer) GetNoteGraphHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

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

	log.Printf("[GraphService] HTTP GetNoteGraph: noteID=%s, depth=%d, userID=%s", noteID, depth, userID)

	// Try cache first
	cacheKey := fmt.Sprintf("note:%s:depth-%d", noteID, depth)
	if cached, hash, err := s.cache.LoadNoteLayout(ctx, noteID, depth); err == nil && cached != nil {
		log.Printf("[GraphService] Cache hit for %s (took %v)", cacheKey, time.Since(startTime))
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

	log.Printf("[GraphService] GetNoteGraph completed in %v", time.Since(startTime))
	s.sendGraphData(w, layout, hash)
}

// GetFullGraphHandler handles GET /api/v1/graph/full with chunked transfer
func (s *HTTPServer) GetFullGraphHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Parse limit parameter
	limit := s.limit
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

	log.Printf("[GraphService] HTTP GetFullGraph: userID=%s, limit=%d", userID, limit)

	// Try cache first
	if cached, hash, err := s.cache.LoadFullLayout(ctx, userID); err == nil && cached != nil {
		log.Printf("[GraphService] Cache hit for full graph user=%s (took %v)", userID, time.Since(startTime))
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

	log.Printf("[GraphService] GetFullGraph completed in %v", time.Since(startTime))
	s.sendGraphData(w, layout, hash)
}

// GetDeltaHandler handles GET /api/v1/graph/delta
func (s *HTTPServer) GetDeltaHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	// Parse user_id parameter
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "public"
	}

	// Parse last_hash parameter
	lastHash := r.URL.Query().Get("last_hash")
	if lastHash == "" {
		http.Error(w, "last_hash is required", http.StatusBadRequest)
		return
	}

	log.Printf("[GraphService] HTTP GetDelta: userID=%s, lastHash=%s", userID, lastHash)

	// Try to load delta from cache first
	if delta, err := s.cache.LoadDelta(ctx, userID, lastHash); err == nil && delta != nil {
		log.Printf("[GraphService] Cache hit for delta (took %v)", time.Since(startTime))
		s.sendDeltaData(w, delta)
		return
	}

	// Load current layout
	notes, links, err := s.postgres.GetNotes(ctx, "", 0)
	if err != nil {
		log.Printf("[GraphService] Failed to load current layout: %v", err)
		http.Error(w, "Failed to load current layout", http.StatusInternalServerError)
		return
	}

	current := engine.Layout3D(notes, links)
	currentHash := computeLayoutHash(current)

	// Load old layout for comparison
	oldLayout, _, err := s.cache.LoadFullLayout(ctx, userID)
	if err != nil {
		log.Printf("[GraphService] Failed to load old layout for delta: %v", err)
		// If no old layout, return everything as added
		delta := &engine.DeltaResponse{
			AddedNodes:  current.Nodes,
			AddedLinks:  current.Links,
			CurrentHash: currentHash,
		}
		s.sendDeltaData(w, delta)
		return
	}

	// Compute delta
	delta := engine.ComputeDelta(oldLayout, current)
	delta.CurrentHash = currentHash

	// Cache the delta
	if err := s.cache.SaveDelta(ctx, userID, lastHash, delta); err != nil {
		log.Printf("[GraphService] Warning: failed to cache delta: %v", err)
	}

	// Update cached full layout
	if err := s.cache.SaveFullLayout(ctx, userID, current, currentHash); err != nil {
		log.Printf("[GraphService] Warning: failed to update full layout cache: %v", err)
	}

	log.Printf("[GraphService] GetDelta completed in %v", time.Since(startTime))
	s.sendDeltaData(w, delta)
}

// sendGraphData sends the layout data as JSON response
func (s *HTTPServer) sendGraphData(w http.ResponseWriter, layout *engine.LayoutResponse, hash string) {
	response := GraphApiResponse{
		Data: GraphData{
			Nodes: convertLayoutNodes(layout.Nodes),
			Links: convertLayoutLinks(layout.Links),
		},
		Meta: &struct {
			TotalNodes int    `json:"total_nodes,omitempty"`
			TotalLinks int    `json:"total_links,omitempty"`
			Hash       string `json:"hash,omitempty"`
		}{
			TotalNodes: len(layout.Nodes),
			TotalLinks: len(layout.Links),
			Hash:       hash,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Layout-Hash", hash)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[GraphService] Failed to encode response: %v", err)
	}
}

// sendDeltaData sends the delta data as JSON response
func (s *HTTPServer) sendDeltaData(w http.ResponseWriter, delta *engine.DeltaResponse) {
	response := convertDeltaResponse(delta)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[GraphService] Failed to encode delta response: %v", err)
	}
}

// RegisterHTTPHandlers registers all HTTP handlers
func RegisterHTTPHandlers(mux *http.ServeMux, srv graphservice.GraphServiceServer) {
	if gs, ok := srv.(*graphService); ok {
		httpServer := NewHTTPServer(gs.postgres, gs.cache, gs.fullLimit)

		// Health check
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"graph-service"}`))
		})

		// Graph endpoints
		mux.HandleFunc("/api/v1/graph/note/", httpServer.GetNoteGraphHandler)
		mux.HandleFunc("/api/v1/graph/full", httpServer.GetFullGraphHandler)
		mux.HandleFunc("/api/v1/graph/delta", httpServer.GetDeltaHandler)
	}
}
