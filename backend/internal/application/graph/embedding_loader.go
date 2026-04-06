package graph

import (
	"context"
	"log"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/google/uuid"
)

// embeddingNeighborLoader загружает соседей на основе семантического сходства эмбеддингов.
type embeddingNeighborLoader struct {
	embeddingRepo *postgres.EmbeddingRepository
	limit         int
}

// NewEmbeddingNeighborLoader создаёт загрузчик для эмбеддингов.
// limit — сколько самых похожих заметок возвращать (рекомендуется 20-50).
func NewEmbeddingNeighborLoader(embeddingRepo *postgres.EmbeddingRepository, limit int) graph.NeighborLoader {
	return &embeddingNeighborLoader{
		embeddingRepo: embeddingRepo,
		limit:         limit,
	}
}

// GetNeighbors возвращает рёбра к семантически похожим заметкам.
// Вес ребра = косинусное сходство эмбеддингов (от 0 до 1).
func (l *embeddingNeighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]graph.Edge, error) {
	similar, err := l.embeddingRepo.FindSimilarNotes(ctx, nodeID, l.limit)
	if err != nil {
		log.Printf("embeddingNeighborLoader: failed to find similar notes for %s: %v", nodeID, err)
		return nil, err
	}

	edges := make([]graph.Edge, len(similar))
	for i, s := range similar {
		edges[i] = graph.Edge{
			From:   nodeID,
			To:     s.NoteID,
			Weight: s.Score,
		}
	}
	return edges, nil
}
