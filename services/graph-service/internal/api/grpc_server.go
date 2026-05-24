package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"knowledge-graph-graph-service/internal/cache"
	"knowledge-graph-graph-service/internal/db"
	"knowledge-graph-graph-service/internal/engine"
)

// GraphServiceServer defines the gRPC service interface
type GraphServiceServer interface {
	GetFullLayout(context.Context, *NoteLayoutRequest) (GraphService_GetFullLayoutServer, error)
	GetDelta(context.Context, *DeltaRequest) (*ProtoDeltaResponse, error)
}

// Message definitions for protobuf

type NoteLayoutRequest struct {
	UserId string
}

type DeltaRequest struct {
	UserId   string
	LastHash string
}

type ProtoDeltaResponse struct {
	AddedNodes   []*ProtoLayoutNode `json:"added_nodes,omitempty"`
	RemovedNodes []string           `json:"removed_nodes,omitempty"`
	UpdatedNodes []*ProtoLayoutNode `json:"updated_nodes,omitempty"`
	AddedLinks   []*ProtoLayoutLink `json:"added_links,omitempty"`
	RemovedLinks []*ProtoLayoutLink `json:"removed_links,omitempty"`
	CurrentHash  string             `json:"current_hash,omitempty"`
}

type ProtoLayoutNode struct {
	Id    string  `json:"id"`
	Title string  `json:"title"`
	Type  string  `json:"type"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Z     float64 `json:"z"`
	Size  float64 `json:"size"`
}

type ProtoLayoutLink struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Weight   float64 `json:"weight"`
	LinkType string  `json:"link_type"`
}

type LayoutChunk struct {
	Nodes   []*ProtoLayoutNode `json:"nodes,omitempty"`
	Links   []*ProtoLayoutLink `json:"links,omitempty"`
	ChunkId string             `json:"chunk_id,omitempty"`
}

type GraphService_GetFullLayoutServer interface {
	Send(*LayoutChunk) error
	Context() context.Context
}

// graphService implements the GraphServiceServer interface
type graphService struct {
	cache     *cache.RedisCache
	postgres  db.PostgresClient
	fullLimit int
}

// NewGraphService creates a new graph service instance
func NewGraphService(postgres db.PostgresClient, cache *cache.RedisCache, fullLimit int) GraphServiceServer {
	return &graphService{
		cache:     cache,
		postgres:  postgres,
		fullLimit: fullLimit,
	}
}

// Convert engine.LayoutNode to ProtoLayoutNode
func convertLayoutNode(node *engine.LayoutNode) *ProtoLayoutNode {
	return &ProtoLayoutNode{
		Id:    node.ID,
		Title: node.Title,
		Type:  node.Type,
		X:     node.X,
		Y:     node.Y,
		Z:     node.Z,
		Size:  node.Size,
	}
}

// Convert engine.LayoutLink to ProtoLayoutLink
func convertLayoutLink(link *engine.LayoutLink) *ProtoLayoutLink {
	return &ProtoLayoutLink{
		Source:   link.Source,
		Target:   link.Target,
		Weight:   link.Weight,
		LinkType: link.LinkType,
	}
}

// Convert []*engine.LayoutNode to []*ProtoLayoutNode
func convertLayoutNodes(nodes []*engine.LayoutNode) []*ProtoLayoutNode {
	result := make([]*ProtoLayoutNode, len(nodes))
	for i, node := range nodes {
		result[i] = convertLayoutNode(node)
	}
	return result
}

// Convert []*engine.LayoutLink to []*ProtoLayoutLink
func convertLayoutLinks(links []*engine.LayoutLink) []*ProtoLayoutLink {
	result := make([]*ProtoLayoutLink, len(links))
	for i, link := range links {
		result[i] = convertLayoutLink(link)
	}
	return result
}

// Convert engine.DeltaResponse to ProtoDeltaResponse
func convertDeltaResponse(delta *engine.DeltaResponse) *ProtoDeltaResponse {
	return &ProtoDeltaResponse{
		AddedNodes:   convertLayoutNodes(delta.AddedNodes),
		RemovedNodes: delta.RemovedNodes,
		UpdatedNodes: convertLayoutNodes(delta.UpdatedNodes),
		AddedLinks:   convertLayoutLinks(delta.AddedLinks),
		RemovedLinks: convertLayoutLinks(delta.RemovedLinks),
		CurrentHash:  delta.CurrentHash,
	}
}

// GetFullLayout streams the full graph layout
func (s *graphService) GetFullLayout(ctx context.Context, req *NoteLayoutRequest) (GraphService_GetFullLayoutServer, error) {
	return nil, status.Error(codes.Unimplemented, "streaming not implemented in stub")
}

// GetDelta returns the delta between two layout versions
func (s *graphService) GetDelta(ctx context.Context, req *DeltaRequest) (*ProtoDeltaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be empty")
	}
	if req.UserId == "" {
		req.UserId = "public"
	}
	if req.LastHash == "" {
		return nil, status.Error(codes.InvalidArgument, "last_hash is required")
	}

	// Try to load from cache first
	if delta, err := s.cache.LoadDelta(ctx, req.UserId, req.LastHash); err == nil && delta != nil {
		return convertDeltaResponse(delta), nil
	}

	// Load current layout
	notes, links, err := s.postgres.GetNotes(ctx, "", 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to load current layout: %v", err)
	}
	current := engine.Layout3D(notes, links)
	currentHash := computeLayoutHash(current)

	// Load old layout for comparison
	oldLayout, _, err := s.cache.LoadFullLayout(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to load old layout: %v", err)
	}

	// Compute delta
	delta := engine.ComputeDelta(oldLayout, current)
	delta.CurrentHash = currentHash

	// Cache the delta
	if err := s.cache.SaveDelta(ctx, req.UserId, req.LastHash, delta); err != nil {
		// Log warning but don't fail
		// log.Printf("[GraphService] Warning: failed to cache delta: %v", err)
	}

	// Update cached full layout
	if err := s.cache.SaveFullLayout(ctx, req.UserId, current, currentHash); err != nil {
		// Log warning but don't fail
		// log.Printf("[GraphService] Warning: failed to update full layout cache: %v", err)
	}

	return convertDeltaResponse(delta), nil
}

// RegisterGraphServiceServer registers the graph service with the gRPC server
func RegisterGraphServiceServer(s *grpc.Server, srv GraphServiceServer) {
	// TODO: Implement proper gRPC registration when protobuf is generated
}

// computeLayoutHash generates a hash for the layout
func computeLayoutHash(layout *engine.LayoutResponse) string {
	payload, _ := json.Marshal(layout)
	h := sha256.Sum256(payload)
	return hex.EncodeToString(h[:])
}
