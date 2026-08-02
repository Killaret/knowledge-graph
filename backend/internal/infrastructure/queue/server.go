package queue

import (
	"context"

	"knowledge-graph/internal/infrastructure/queue/tasks"

	"github.com/hibiken/asynq"
)

// TypeRefreshRecommendations re-exports the tasks package constant so callers only import queue.
const TypeRefreshRecommendations = tasks.TypeRefreshRecommendations

// TypeRecalculateLinkWeights re-exports the tasks package constant.
const TypeRecalculateLinkWeights = tasks.TypeRecalculateLinkWeights

// Server wraps an asynq server so that cmd/worker does not import asynq directly.
type Server struct {
	srv *asynq.Server
}

// NewServer creates a new asynq server for the given Redis address and config.
func NewServer(redisAddr string, concurrency int, queues map[string]int) *Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: concurrency,
			Queues:      queues,
		},
	)
	return &Server{srv: srv}
}

// Run starts the server with the given mux.
func (s *Server) Run(mux *ServeMux) error {
	return s.srv.Run(mux.mux)
}

// Stop signals the server to stop accepting new tasks and wait for in-flight tasks.
func (s *Server) Stop() {
	s.srv.Stop()
}

// ServeMux wraps an asynq ServeMux.
type ServeMux struct {
	mux *asynq.ServeMux
}

// NewServeMux creates a new asynq ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{mux: asynq.NewServeMux()}
}

// HandleFunc registers a handler for the given task type.
func (m *ServeMux) HandleFunc(taskType string, handler func(context.Context, *asynq.Task) error) {
	m.mux.HandleFunc(taskType, handler)
}

// RefreshRecommendationsHandler returns a handler that dispatches to tasks.HandleRefreshRecommendations.
func RefreshRecommendationsHandler(refreshSvc tasks.RefreshServiceInterface) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		return tasks.HandleRefreshRecommendations(ctx, t, refreshSvc)
	}
}

// RecalculateLinkWeightsHandler returns a handler that dispatches to tasks.HandleRecalculateLinkWeights.
func RecalculateLinkWeightsHandler(svc tasks.LinkWeightRecalculator) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, t *asynq.Task) error {
		return tasks.HandleRecalculateLinkWeights(ctx, t, svc)
	}
}
