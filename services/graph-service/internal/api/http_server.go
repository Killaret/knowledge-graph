package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/db"
	"knowledge-graph-graph-service/internal/engine"
	graphservice "knowledge-graph-graph-service/proto"

	"github.com/google/uuid"
)

// HTTPServer handles HTTP requests for the graph service
type HTTPServer struct {
	cache        *cache.RedisCache
	postgres     db.PostgresClient
	limit        int
	defaultDepth int
}

// NewHTTPServer creates a new HTTP server
func NewHTTPServer(postgres db.PostgresClient, cache *cache.RedisCache, limit, defaultDepth int) *HTTPServer {
	return &HTTPServer{
		cache:        cache,
		postgres:     postgres,
		limit:        limit,
		defaultDepth: defaultDepth,
	}
}

// GraphData represents the response structure
type GraphData struct {
	Nodes []*engine.LayoutNode `json:"nodes"`
	Links []*engine.LayoutLink `json:"links"`
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

	// Extract note ID from path - handle /api/v1/graph/note/{id}
	log.Printf("[GraphService] Raw URL.Path: '%s', URL.RawPath: '%s'", r.URL.Path, r.URL.RawPath)

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/graph/note/")

	// Remove query parameters if any
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}

	noteID := strings.TrimSpace(path)
	if noteID == "" {
		log.Printf("[GraphService] Empty noteID from path: '%s'", r.URL.Path)
		http.Error(w, "Note ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("[GraphService] Extracted noteID: '%s' from path: '%s'", noteID, r.URL.Path)

	// Validate UUID format
	if _, err := uuid.Parse(noteID); err != nil {
		log.Printf("[GraphService] Invalid UUID format: %s", noteID)
		http.Error(w, "Invalid note ID format (must be UUID)", http.StatusBadRequest)
		return
	}

	// Parse depth parameter
	depth := s.defaultDepth
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

	// Filter links to only include those between the limited nodes
	if limit > 0 && len(notes) < len(links) {
		limitedNodeIDs := make(map[string]bool)
		for _, note := range notes {
			limitedNodeIDs[note.ID] = true
		}

		filteredLinks := make([]*db.Link, 0)
		for _, link := range links {
			if limitedNodeIDs[link.Source] && limitedNodeIDs[link.Target] {
				filteredLinks = append(filteredLinks, link)
			}
		}
		links = filteredLinks
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
			Nodes: layout.Nodes,
			Links: layout.Links,
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
	// Create a flat list of all nodes (added + updated, removed are just IDs)
	allNodes := append([]*engine.LayoutNode{}, delta.AddedNodes...)
	allNodes = append(allNodes, delta.UpdatedNodes...)

	response := GraphApiResponse{
		Data: GraphData{
			Nodes: allNodes,
			Links: append([]*engine.LayoutLink{}, delta.AddedLinks...),
		},
		Meta: &struct {
			TotalNodes int    `json:"total_nodes,omitempty"`
			TotalLinks int    `json:"total_links,omitempty"`
			Hash       string `json:"hash,omitempty"`
		}{
			TotalNodes: len(allNodes),
			TotalLinks: len(delta.AddedLinks),
			Hash:       delta.CurrentHash,
		},
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[GraphService] Failed to encode delta response: %v", err)
	}
}

// RegisterHTTPHandlers registers all HTTP handlers
func RegisterHTTPHandlers(mux *http.ServeMux, srv graphservice.GraphServiceServer) {
	if gs, ok := srv.(*graphService); ok {
		httpServer := NewHTTPServer(gs.postgres, gs.cache, gs.fullLimit, gs.defaultDepth)

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
