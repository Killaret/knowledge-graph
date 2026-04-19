package graph

import (
	"context"
	"log"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/infrastructure/db/postgres"

	"github.com/google/uuid"
)

// embeddingRepositoryWithBatch — интерфейс для batch-загрузки эмбеддингов
type embeddingRepositoryWithBatch interface {
	FindSimilarNotes(ctx context.Context, noteID uuid.UUID, limit int) ([]postgres.SimilarNote, error)
	FindSimilarNotesBatch(ctx context.Context, noteIDs []uuid.UUID, limit int) (map[uuid.UUID][]postgres.SimilarNote, error)
}

// embeddingNeighborLoader загружает соседей на основе семантического сходства эмбеддингов.
type embeddingNeighborLoader struct {
	embeddingRepo embeddingRepositoryWithBatch
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

// GetNeighborsBatch возвращает соседей для нескольких узлов на основе эмбеддингов (batch-запрос)
func (l *embeddingNeighborLoader) GetNeighborsBatch(ctx context.Context, nodeIDs []uuid.UUID) (map[uuid.UUID][]graph.Edge, error) {
	if len(nodeIDs) == 0 {
		return make(map[uuid.UUID][]graph.Edge), nil
	}

	// Пробуем привести репозиторий к интерфейсу с batch-методом
	batchRepo, ok := l.embeddingRepo.(interface {
		FindSimilarNotesBatch(ctx context.Context, noteIDs []uuid.UUID, limit int) (map[uuid.UUID][]postgres.SimilarNote, error)
	})

	if !ok {
		// Fallback на последовательные запросы
		result := make(map[uuid.UUID][]graph.Edge, len(nodeIDs))
		for _, nodeID := range nodeIDs {
			edges, err := l.GetNeighbors(ctx, nodeID)
			if err != nil {
				continue
			}
			result[nodeID] = edges
		}
		return result, nil
	}

	// Batch-запрос
	similarMap, err := batchRepo.FindSimilarNotesBatch(ctx, nodeIDs, l.limit)
	if err != nil {
		// Fallback на последовательные запросы при ошибке
		result := make(map[uuid.UUID][]graph.Edge, len(nodeIDs))
		for _, nodeID := range nodeIDs {
			edges, err := l.GetNeighbors(ctx, nodeID)
			if err != nil {
				continue
			}
			result[nodeID] = edges
		}
		return result, nil
	}

	// Преобразуем результат в Edge
	result := make(map[uuid.UUID][]graph.Edge, len(nodeIDs))
	for nodeID, similar := range similarMap {
		edges := make([]graph.Edge, len(similar))
		for i, s := range similar {
			edges[i] = graph.Edge{
				From:   nodeID,
				To:     s.NoteID,
				Weight: s.Score,
			}
		}
		result[nodeID] = edges
	}

	// Убеждаемся, что все запрошенные узлы есть в результате (даже если пустые)
	for _, nodeID := range nodeIDs {
		if _, ok := result[nodeID]; !ok {
			result[nodeID] = []graph.Edge{}
		}
	}

	return result, nil
}
