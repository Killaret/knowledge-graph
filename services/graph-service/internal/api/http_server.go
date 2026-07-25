package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/config"
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

func (s *HTTPServer) cacheUserID(ctx context.Context) string {
	if userID, ok := userIDFromContext(ctx); ok && userID != "" {
		return userID
	}
	return "public"
}

func (s *HTTPServer) notesFilter(ctx context.Context) db.NotesFilter {
	filter := db.NotesFilter{}
	if isPublicRequest(ctx) {
		filter.IsPublic = true
	} else if userID, ok := userIDFromContext(ctx); ok && userID != "" {
		filter.UserID = userID
	}
	return filter
}

// parseLayout returns the requested layout type (2d or 3d).
func parseLayout(r *http.Request) string {
	layoutType := r.URL.Query().Get("layout")
	if layoutType == "3d" {
		return "3d"
	}
	return "2d"
}

// GetNoteGraphHandler handles GET /api/v1/graph/note/:id
func (s *HTTPServer) GetNoteGraphHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/graph/note/")
	if idx := strings.Index(path, "?"); idx != -1 {
		path = path[:idx]
	}
	noteID := strings.TrimSpace(path)
	if noteID == "" {
		http.Error(w, "Note ID is required", http.StatusBadRequest)
		return
	}
	if _, err := uuid.Parse(noteID); err != nil {
		http.Error(w, "Invalid note ID format (must be UUID)", http.StatusBadRequest)
		return
	}

	depth := s.defaultDepth
	if depthStr := r.URL.Query().Get("depth"); depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d > 0 {
			depth = d
		}
	}
	layoutType := parseLayout(r)

	filter := s.notesFilter(ctx)
	filter.RootID = noteID
	filter.Depth = depth
	cacheUserID := s.cacheUserID(ctx)

	log.Printf("[GraphService] HTTP GetNoteGraph: noteID=%s, depth=%d, layout=%s, userID=%s, public=%v", noteID, depth, layoutType, filter.UserID, filter.IsPublic)

	if layoutType == "2d" {
		if cached, hash, err := s.cache.LoadNoteLayout(ctx, cacheUserID, noteID, depth); err == nil && cached != nil {
			log.Printf("[GraphService] Cache hit for note layout user=%s note=%s depth=%d (took %v)", cacheUserID, noteID, depth, time.Since(startTime))
			s.sendGraphData(w, cached, hash)
			return
		}
	}

	notes, links, err := s.postgres.GetNotes(ctx, filter)
	if err != nil {
		log.Printf("[GraphService] Failed to load notes: %v", err)
		http.Error(w, "Failed to load graph", http.StatusInternalServerError)
		return
	}

	var layout *engine.LayoutResponse
	if layoutType == "3d" {
		layout = engine.Layout3D(notes, links)
	} else {
		layout = engine.Layout2D(notes, links, noteID)
	}
	hash := computeLayoutHash(layout)

	if layoutType == "2d" {
		if err := s.cache.SaveNoteLayout(ctx, cacheUserID, noteID, depth, layout, hash); err != nil {
			log.Printf("[GraphService] Failed to cache layout: %v", err)
		}
	}

	log.Printf("[GraphService] GetNoteGraph completed in %v", time.Since(startTime))
	s.sendGraphData(w, layout, hash)
}

// GetFullGraphHandler handles GET /api/v1/graph/full with chunked transfer
func (s *HTTPServer) GetFullGraphHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	limit := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	} else if s.limit > 0 {
		limit = s.limit
	}
	nocache := r.URL.Query().Get("nocache") == "1" || r.URL.Query().Get("nocache") == "true"

	filter := s.notesFilter(ctx)
	cacheUserID := s.cacheUserID(ctx)

	log.Printf("[GraphService] HTTP GetFullGraph: userID=%s, public=%v, limit=%d, nocache=%v", filter.UserID, filter.IsPublic, limit, nocache)

	if !nocache {
		if cached, hash, err := s.cache.LoadFullLayout(ctx, cacheUserID); err == nil && cached != nil {
			log.Printf("[GraphService] Cache hit for full graph user=%s (took %v)", cacheUserID, time.Since(startTime))
			s.sendGraphData(w, cached, hash)
			return
		}
	}

	notes, links, err := s.postgres.GetNotes(ctx, filter)
	if err != nil {
		log.Printf("[GraphService] Failed to load full graph: %v", err)
		http.Error(w, "Failed to load full graph", http.StatusInternalServerError)
		return
	}

	if limit > 0 && len(notes) > limit {
		notes = notes[:limit]
	}
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

	layout := engine.Layout3D(notes, links)
	hash := computeLayoutHash(layout)

	if err := s.cache.SaveFullLayout(ctx, cacheUserID, layout, hash); err != nil {
		log.Printf("[GraphService] Failed to cache full layout: %v", err)
	}

	log.Printf("[GraphService] GetFullGraph completed in %v", time.Since(startTime))
	s.sendGraphData(w, layout, hash)
}

// GetPublicGraphHandler handles GET /api/v1/graph/public without authentication.
func (s *HTTPServer) GetPublicGraphHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	limit := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	} else if s.limit > 0 {
		limit = s.limit
	}
	layoutType := parseLayout(r)

	filter := db.NotesFilter{IsPublic: true}

	log.Printf("[GraphService] HTTP GetPublicGraph: limit=%d, layout=%s", limit, layoutType)

	if cached, hash, err := s.cache.LoadFullLayout(ctx, "public"); err == nil && cached != nil {
		log.Printf("[GraphService] Cache hit for public graph (took %v)", time.Since(startTime))
		s.sendGraphData(w, cached, hash)
		return
	}

	notes, links, err := s.postgres.GetNotes(ctx, filter)
	if err != nil {
		log.Printf("[GraphService] Failed to load public graph: %v", err)
		http.Error(w, "Failed to load public graph", http.StatusInternalServerError)
		return
	}

	if limit > 0 && len(notes) > limit {
		notes = notes[:limit]
	}
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

	var layout *engine.LayoutResponse
	if layoutType == "3d" {
		layout = engine.Layout3D(notes, links)
	} else {
		layout = engine.Layout2D(notes, links, "")
	}
	hash := computeLayoutHash(layout)

	if err := s.cache.SaveFullLayout(ctx, "public", layout, hash); err != nil {
		log.Printf("[GraphService] Failed to cache public layout: %v", err)
	}

	log.Printf("[GraphService] GetPublicGraph completed in %v", time.Since(startTime))
	s.sendGraphData(w, layout, hash)
}

// GetDeltaHandler handles GET /api/v1/graph/delta
func (s *HTTPServer) GetDeltaHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	startTime := time.Now()

	filter := s.notesFilter(ctx)
	cacheUserID := s.cacheUserID(ctx)

	lastHash := r.URL.Query().Get("last_hash")
	if lastHash == "" {
		http.Error(w, "last_hash is required", http.StatusBadRequest)
		return
	}

	log.Printf("[GraphService] HTTP GetDelta: userID=%s, public=%v, lastHash=%s", filter.UserID, filter.IsPublic, lastHash)

	if delta, err := s.cache.LoadDelta(ctx, cacheUserID, lastHash); err == nil && delta != nil {
		log.Printf("[GraphService] Cache hit for delta (took %v)", time.Since(startTime))
		s.sendDeltaData(w, delta)
		return
	}

	notes, links, err := s.postgres.GetNotes(ctx, filter)
	if err != nil {
		log.Printf("[GraphService] Failed to load current layout: %v", err)
		http.Error(w, "Failed to load current layout", http.StatusInternalServerError)
		return
	}

	current := engine.Layout3D(notes, links)
	currentHash := computeLayoutHash(current)

	oldLayout, _, err := s.cache.LoadFullLayout(ctx, cacheUserID)
	if err != nil {
		delta := &engine.DeltaResponse{
			AddedNodes:  current.Nodes,
			AddedLinks:  current.Links,
			CurrentHash: currentHash,
		}
		s.sendDeltaData(w, delta)
		return
	}

	delta := engine.ComputeDelta(oldLayout, current)
	delta.CurrentHash = currentHash

	if err := s.cache.SaveDelta(ctx, cacheUserID, lastHash, delta); err != nil {
		log.Printf("[GraphService] Warning: failed to cache delta: %v", err)
	}
	if err := s.cache.SaveFullLayout(ctx, cacheUserID, current, currentHash); err != nil {
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

// sendDeltaData sends the delta data as JSON response with the canonical
// added_nodes/removed_nodes/updated_nodes/added_links/removed_links shape.
func (s *HTTPServer) sendDeltaData(w http.ResponseWriter, delta *engine.DeltaResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Layout-Hash", delta.CurrentHash)

	if err := json.NewEncoder(w).Encode(delta); err != nil {
		log.Printf("[GraphService] Failed to encode delta response: %v", err)
	}
}

// RegisterHTTPHandlers registers all HTTP handlers
func RegisterHTTPHandlers(mux *http.ServeMux, srv graphservice.GraphServiceServer, cfg *config.Config) {
	if gs, ok := srv.(*graphService); ok {
		httpServer := NewHTTPServer(gs.postgres, gs.cache, gs.fullLimit, gs.defaultDepth)

		// Health check (no auth)
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok","service":"graph-service"}`))
		})

		// Public graph endpoint (no auth)
		mux.HandleFunc("/api/v1/graph/public", AuthMiddleware(cfg, true, httpServer.GetPublicGraphHandler))

		// Private graph endpoints (require authentication)
		mux.HandleFunc("/api/v1/graph/note/", AuthMiddleware(cfg, false, httpServer.GetNoteGraphHandler))
		mux.HandleFunc("/api/v1/graph/full", AuthMiddleware(cfg, false, httpServer.GetFullGraphHandler))
		mux.HandleFunc("/api/v1/graph/delta", AuthMiddleware(cfg, false, httpServer.GetDeltaHandler))
	}
}
