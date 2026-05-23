package api

import (
	"context"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: Implement full gRPC server with proper type conversion
// This is a temporary stub to allow compilation

type GraphServiceServer interface {
	GetFullLayout(context.Context, *NoteLayoutRequest) (GraphService_GetFullLayoutServer, error)
	GetDelta(context.Context, *DeltaRequest) (*ProtoDeltaResponse, error)
}

type NoteLayoutRequest struct {
	UserId string
}

type DeltaRequest struct {
	UserId   string
	LastHash string
}

type ProtoDeltaResponse struct {
	AddedNodes   []*ProtoLayoutNode
	RemovedNodes []string
	UpdatedNodes []*ProtoLayoutNode
	AddedLinks   []*ProtoLayoutLink
	RemovedLinks []*ProtoLayoutLink
	CurrentHash  string
}

type ProtoLayoutNode struct {
	Id    string
	Title string
	Type  string
	X     float64
	Y     float64
	Z     float64
	Size  float64
}

type ProtoLayoutLink struct {
	Source   string
	Target   string
	Weight   float64
	LinkType string
}

type LayoutChunk struct {
	Nodes   []*ProtoLayoutNode
	Links   []*ProtoLayoutLink
	ChunkId string
}

type GraphService_GetFullLayoutServer interface {
	Send(*LayoutChunk) error
	Context() context.Context
}

type graphService struct {
	cache     interface{}
	postgres  interface{}
	fullLimit int
}

func NewGraphServer(postgres, cache interface{}, fullLimit int) *graphService {
	return &graphService{cache: cache, postgres: postgres, fullLimit: fullLimit}
}

func (s *graphService) GetFullLayout(ctx context.Context, req *NoteLayoutRequest) (GraphService_GetFullLayoutServer, error) {
	return nil, status.Error(codes.Unimplemented, "method GetFullLayout not implemented")
}

func (s *graphService) GetDelta(ctx context.Context, req *DeltaRequest) (*ProtoDeltaResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetDelta not implemented")
}

func RegisterGraphServiceServer(s *grpc.Server, srv GraphServiceServer) {
	// TODO: Implement registration
}

func RegisterHTTPHandlers(mux *http.ServeMux, service interface{}) {
	// TODO: Implement HTTP handlers
	mux.HandleFunc("/api/v1/graph/note/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})
	mux.HandleFunc("/api/v1/graph/full", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})
}
