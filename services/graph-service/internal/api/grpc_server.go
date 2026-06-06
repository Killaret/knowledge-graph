package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/db"
	"knowledge-graph-graph-service/internal/engine"
	graphservice "knowledge-graph-graph-service/proto"
)

// graphService implements the GraphServiceServer interface
type graphService struct {
	graphservice.UnimplementedGraphServiceServer
	cache           *cache.RedisCache
	postgres        db.PostgresClient
	fullLimit       int
	defaultDepth    int
	streamChunkSize int
}

// NewGraphService creates a new graph service instance
func NewGraphService(postgres db.PostgresClient, cache *cache.RedisCache, fullLimit, defaultDepth, streamChunkSize int) graphservice.GraphServiceServer {
	return &graphService{
		cache:           cache,
		postgres:        postgres,
		fullLimit:       fullLimit,
		defaultDepth:    defaultDepth,
		streamChunkSize: streamChunkSize,
	}
}

// GetNoteLayout returns the layout for a specific note and its neighbors
func (s *graphService) GetNoteLayout(ctx context.Context, req *graphservice.NoteLayoutRequest) (*graphservice.LayoutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	noteID := req.NoteId
	depth := req.Depth
	userID := req.UserId

	if userID == "" {
		userID = "public"
	}

	if depth <= 0 {
		depth = int32(s.defaultDepth)
	}

	log.Printf("[GraphService] GetNoteLayout: noteID=%s, depth=%d, userID=%s", noteID, depth, userID)

	// Try cache first
	if cached, hash, err := s.cache.LoadNoteLayout(ctx, noteID, int(depth)); err == nil && cached != nil {
		log.Printf("[GraphService] Cache hit for note layout: %s", noteID)
		return convertLayoutResponse(cached, hash), nil
	}

	// Load from database
	notes, links, err := s.postgres.GetNotes(ctx, noteID, int(depth))
	if err != nil {
		log.Printf("[GraphService] Failed to load notes from DB: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to load graph: %v", err)
	}

	// Generate layout
	layout := engine.Layout2D(notes, links, noteID)
	hash := computeLayoutHash(layout)

	// Cache the result
	if err := s.cache.SaveNoteLayout(ctx, noteID, int(depth), layout, hash); err != nil {
		log.Printf("[GraphService] Warning: failed to cache note layout: %v", err)
	}

	return convertLayoutResponse(layout, hash), nil
}

// GetFullLayout streams the full graph layout in chunks
func (s *graphService) GetFullLayout(req *graphservice.FullLayoutRequest, stream graphservice.GraphService_GetFullLayoutServer) error {
	if req == nil {
		return status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	ctx := stream.Context()
	userID := req.UserId
	limit := req.Limit

	if userID == "" {
		userID = "public"
	}

	if limit <= 0 {
		limit = int32(s.fullLimit)
	}

	log.Printf("[GraphService] GetFullLayout: userID=%s, limit=%d", userID, limit)

	// Try cache first
	if cached, hash, err := s.cache.LoadFullLayout(ctx, userID); err == nil && cached != nil {
		log.Printf("[GraphService] Cache hit for full layout: user=%s", userID)
		return s.streamLayout(cached, hash, stream)
	}

	// Load from database
	notes, links, err := s.postgres.GetNotes(ctx, "", 0)
	if err != nil {
		log.Printf("[GraphService] Failed to load full graph from DB: %v", err)
		return status.Errorf(codes.Internal, "failed to load full graph: %v", err)
	}

	// Apply limit
	if int(limit) > 0 && len(notes) > int(limit) {
		notes = notes[:limit]
	}

	// Generate layout
	layout := engine.Layout3D(notes, links)
	hash := computeLayoutHash(layout)

	// Cache the result
	if err := s.cache.SaveFullLayout(ctx, userID, layout, hash); err != nil {
		log.Printf("[GraphService] Warning: failed to cache full layout: %v", err)
	}

	return s.streamLayout(layout, hash, stream)
}

// streamLayout sends the layout in chunks
func (s *graphService) streamLayout(layout *engine.LayoutResponse, hash string, stream graphservice.GraphService_GetFullLayoutServer) error {
	chunkSize := s.streamChunkSize
	if chunkSize <= 0 {
		chunkSize = 100
	}
	totalNodes := len(layout.Nodes)

	for i := 0; i < totalNodes; i += chunkSize {
		end := i + chunkSize
		if end > totalNodes {
			end = totalNodes
		}

		chunk := &graphservice.LayoutChunk{
			Nodes:   convertLayoutNodes(layout.Nodes[i:end]),
			ChunkId: fmt.Sprintf("nodes-%d-%d", i, end),
		}

		if err := stream.Send(chunk); err != nil {
			log.Printf("[GraphService] Failed to send node chunk: %v", err)
			return err
		}
	}

	// Send links in a separate chunk
	linksChunk := &graphservice.LayoutChunk{
		Links:   convertLayoutLinks(layout.Links),
		ChunkId: "links",
	}

	if err := stream.Send(linksChunk); err != nil {
		log.Printf("[GraphService] Failed to send links chunk: %v", err)
		return err
	}

	// Send final chunk with hash
	finalChunk := &graphservice.LayoutChunk{
		ChunkId: fmt.Sprintf("hash:%s", hash),
	}

	if err := stream.Send(finalChunk); err != nil {
		log.Printf("[GraphService] Failed to send hash chunk: %v", err)
		return err
	}

	return nil
}

// GetDelta returns the delta between two layout versions
func (s *graphService) GetDelta(ctx context.Context, req *graphservice.DeltaRequest) (*graphservice.DeltaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}

	userID := req.UserId
	lastHash := req.LastHash

	if userID == "" {
		userID = "public"
	}

	if lastHash == "" {
		return nil, status.Error(codes.InvalidArgument, "last_hash is required")
	}

	log.Printf("[GraphService] GetDelta: userID=%s, lastHash=%s", userID, lastHash)

	// Try to load delta from cache first
	if delta, err := s.cache.LoadDelta(ctx, userID, lastHash); err == nil && delta != nil {
		log.Printf("[GraphService] Cache hit for delta: %s", lastHash)
		return convertDeltaResponse(delta), nil
	}

	// Load current layout
	notes, links, err := s.postgres.GetNotes(ctx, "", 0)
	if err != nil {
		log.Printf("[GraphService] Failed to load current layout: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to load current layout: %v", err)
	}

	current := engine.Layout3D(notes, links)
	currentHash := computeLayoutHash(current)

	// Load old layout for comparison
	oldLayout, _, err := s.cache.LoadFullLayout(ctx, userID)
	if err != nil {
		log.Printf("[GraphService] Failed to load old layout for delta: %v", err)
		// If no old layout, return everything as added
		return &graphservice.DeltaResponse{
			AddedNodes:  convertLayoutNodes(current.Nodes),
			AddedLinks:  convertLayoutLinks(current.Links),
			CurrentHash: currentHash,
		}, nil
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

	return convertDeltaResponse(delta), nil
}

// Conversion functions

func convertLayoutResponse(layout *engine.LayoutResponse, hash string) *graphservice.LayoutResponse {
	return &graphservice.LayoutResponse{
		Nodes: convertLayoutNodes(layout.Nodes),
		Links: convertLayoutLinks(layout.Links),
		Hash:  hash,
	}
}

func convertLayoutNodes(nodes []*engine.LayoutNode) []*graphservice.LayoutNode {
	result := make([]*graphservice.LayoutNode, len(nodes))
	for i, node := range nodes {
		result[i] = &graphservice.LayoutNode{
			Id:    node.ID,
			Title: node.Title,
			Type:  node.Type,
			X:     node.X,
			Y:     node.Y,
			Z:     node.Z,
			Size:  node.Size,
		}
	}
	return result
}

func convertLayoutLinks(links []*engine.LayoutLink) []*graphservice.LayoutLink {
	result := make([]*graphservice.LayoutLink, len(links))
	for i, link := range links {
		result[i] = &graphservice.LayoutLink{
			Source:   link.Source,
			Target:   link.Target,
			Weight:   link.Weight,
			LinkType: link.LinkType,
		}
	}
	return result
}

func convertDeltaResponse(delta *engine.DeltaResponse) *graphservice.DeltaResponse {
	return &graphservice.DeltaResponse{
		AddedNodes:   convertLayoutNodes(delta.AddedNodes),
		RemovedNodes: delta.RemovedNodes,
		UpdatedNodes: convertLayoutNodes(delta.UpdatedNodes),
		AddedLinks:   convertLayoutLinks(delta.AddedLinks),
		RemovedLinks: convertLayoutLinks(delta.RemovedLinks),
		CurrentHash:  delta.CurrentHash,
	}
}

// computeLayoutHash computes a hash of the layout for cache validation
func computeLayoutHash(layout *engine.LayoutResponse) string {
	data, _ := json.Marshal(layout)
	return fmt.Sprintf("%x", data)[:32]
}

// RegisterGraphServiceServer is a wrapper for the generated registration function
func RegisterGraphServiceServer(s interface{}, srv graphservice.GraphServiceServer) {
	// This is handled by the generated code
}
