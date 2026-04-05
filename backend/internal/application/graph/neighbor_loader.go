package graph

import (
	"context"

	"knowledge-graph/internal/domain/graph"
	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
)

// neighborLoader — структура, которая реализует интерфейс graph.NeighborLoader.
// Она знает, как загрузить соседей заметки из базы данных через репозитории.
type neighborLoader struct {
	linkRepo link.Repository // репозиторий для работы со связями
	noteRepo note.Repository // репозиторий для заметок (здесь не используется, но может пригодиться)
}

// NewNeighborLoader — конструктор, возвращает реализацию интерфейса graph.NeighborLoader.
// Принимает репозитории (зависимости), которые нужны для загрузки данных.
func NewNeighborLoader(linkRepo link.Repository, noteRepo note.Repository) graph.NeighborLoader {
	return &neighborLoader{
		linkRepo: linkRepo,
		noteRepo: noteRepo,
	}
}

// GetNeighbors — метод, который для заданной заметки (nodeID) возвращает список рёбер (связей) к соседним заметкам.
// В рамках рекомендаций мы делаем граф неориентированным: учитываем как исходящие, так и входящие связи.
// Каждое ребро содержит:
//   - From: ID исходной заметки (всегда nodeID, но для входящих связей мы разворачиваем направление)
//   - To:   ID соседней заметки
//   - Weight: вес связи (уже рассчитанный и сохранённый в БД)
func (l *neighborLoader) GetNeighbors(ctx context.Context, nodeID uuid.UUID) ([]graph.Edge, error) {
	// 1. Получаем исходящие связи: где наша заметка является источником (source)
	outgoing, err := l.linkRepo.FindBySource(ctx, nodeID)
	if err != nil {
		// Если ошибка при запросе к БД – возвращаем её, чтобы вышестоящий слой мог обработать
		return nil, err
	}

	// 2. Получаем входящие связи: где наша заметка является целью (target)
	incoming, err := l.linkRepo.FindByTarget(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	// 3. Подготавливаем срез для рёбер (примерный размер = сумма исходящих и входящих)
	edges := make([]graph.Edge, 0, len(outgoing)+len(incoming))

	// 4. Добавляем исходящие связи как есть (направление от nodeID к target)
	for _, ln := range outgoing {
		edges = append(edges, graph.Edge{
			From:   ln.SourceNoteID(), // это nodeID
			To:     ln.TargetNoteID(),
			Weight: ln.Weight().Value(), // вес хранится во Value Object Weight, берём float64
		})
	}

	// 5. Добавляем входящие связи, но разворачиваем направление (чтобы граф был неориентированным)
	//    Для рекомендаций не важно, кто на кого ссылается – важна сама связь.
	for _, ln := range incoming {
		edges = append(edges, graph.Edge{
			From:   ln.TargetNoteID(), // в развёрнутом виде: от целевой заметки к источнику
			To:     ln.SourceNoteID(),
			Weight: ln.Weight().Value(),
		})
	}

	// 6. Возвращаем список рёбер
	return edges, nil
}
