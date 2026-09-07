package graph

import (
	"context"
	"fmt"
	"sort"

	"knowledge-graph/internal/domain/link"
	"knowledge-graph/internal/domain/note"

	"github.com/google/uuid"
)

// AnalyticsResult contains computed graph metrics for a set of nodes and links.
type AnalyticsResult struct {
	PageRank   map[string]float64 `json:"page_rank"`
	Clusters   [][]string         `json:"clusters"`
	TopCenters []string           `json:"top_centers"`
}

// Analytics computes graph analytics for the whole visible graph.
type Analytics struct {
	linkRepo link.Repository
	noteRepo note.Repository
}

// NewAnalytics creates a new graph analytics service.
func NewAnalytics(linkRepo link.Repository, noteRepo note.Repository) *Analytics {
	return &Analytics{linkRepo: linkRepo, noteRepo: noteRepo}
}

// ComputeForAll runs PageRank and cluster detection over all non-deleted links.
func (a *Analytics) ComputeForAll(ctx context.Context) (*AnalyticsResult, error) {
	links, err := a.linkRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load links: %w", err)
	}

	// Build adjacency list with weights.
	adj := make(map[uuid.UUID]map[uuid.UUID]float64)
	nodeSet := make(map[uuid.UUID]bool)
	for _, l := range links {
		source := l.SourceNoteID()
		target := l.TargetNoteID()
		w := l.Weight().Value()
		if w <= 0 {
			w = 0.01 // avoid zero-weight edges breaking PageRank
		}
		if adj[source] == nil {
			adj[source] = make(map[uuid.UUID]float64)
		}
		adj[source][target] = w
		nodeSet[source] = true
		nodeSet[target] = true
	}

	// Convert to stable node slice.
	nodes := make([]uuid.UUID, 0, len(nodeSet))
	for id := range nodeSet {
		nodes = append(nodes, id)
	}

	// PageRank with weighted edges.
	pageRank := a.pagerank(nodes, adj)

	// Connected components (clusters).
	clusters := a.clusters(nodes, adj)

	// Top 10 nodes by PageRank.
	top := make([]string, 0, 10)
	type pair struct {
		id    string
		score float64
	}
	pairs := make([]pair, 0, len(pageRank))
	for id, score := range pageRank {
		pairs = append(pairs, pair{id: id, score: score})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].score == pairs[j].score {
			return pairs[i].id < pairs[j].id
		}
		return pairs[i].score > pairs[j].score
	})
	for i := 0; i < len(pairs) && i < 10; i++ {
		top = append(top, pairs[i].id)
	}

	// Convert clusters to string IDs.
	clusterIDs := make([][]string, len(clusters))
	for i, c := range clusters {
		ids := make([]string, len(c))
		for j, id := range c {
			ids[j] = id.String()
		}
		clusterIDs[i] = ids
	}

	result := &AnalyticsResult{
		PageRank:   pageRank,
		Clusters:   clusterIDs,
		TopCenters: top,
	}
	return result, nil
}

func (a *Analytics) pagerank(nodes []uuid.UUID, adj map[uuid.UUID]map[uuid.UUID]float64) map[string]float64 {
	n := len(nodes)
	if n == 0 {
		return map[string]float64{}
	}

	// Map node -> index.
	idx := make(map[uuid.UUID]int, n)
	for i, id := range nodes {
		idx[id] = i
	}

	// Build weighted out-degree.
	outWeight := make([]float64, n)
	for i, u := range nodes {
		for _, w := range adj[u] {
			outWeight[i] += w
		}
	}

	damping := 0.85
	tol := 1e-6
	maxIter := 100

	rank := make([]float64, n)
	base := 1.0 / float64(n)
	for i := range rank {
		rank[i] = base
	}

	for iter := 0; iter < maxIter; iter++ {
		newRank := make([]float64, n)
		for j := 0; j < n; j++ {
			newRank[j] = (1.0 - damping) / float64(n)
		}

		for i, u := range nodes {
			if outWeight[i] == 0 {
				// dangling node: distribute equally
				share := damping * rank[i] / float64(n)
				for j := 0; j < n; j++ {
					newRank[j] += share
				}
				continue
			}

			for v, w := range adj[u] {
				j := idx[v]
				newRank[j] += damping * rank[i] * (w / outWeight[i])
			}
		}

		// Convergence check.
		diff := 0.0
		for i := 0; i < n; i++ {
			d := newRank[i] - rank[i]
			if d < 0 {
				d = -d
			}
			diff += d
		}
		rank = newRank
		if diff < tol {
			break
		}
	}

	result := make(map[string]float64, n)
	for i, id := range nodes {
		result[id.String()] = rank[i]
	}
	return result
}

func (a *Analytics) clusters(nodes []uuid.UUID, adj map[uuid.UUID]map[uuid.UUID]float64) [][]uuid.UUID {
	visited := make(map[uuid.UUID]bool)
	var clusters [][]uuid.UUID

	for _, start := range nodes {
		if visited[start] {
			continue
		}

		component := []uuid.UUID{}
		stack := []uuid.UUID{start}
		visited[start] = true

		for len(stack) > 0 {
			u := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			component = append(component, u)

			for v := range adj[u] {
				if !visited[v] {
					visited[v] = true
					stack = append(stack, v)
				}
			}

			// Check incoming edges by scanning all adjacency lists.
			for x, neighbors := range adj {
				if _, ok := neighbors[u]; ok && !visited[x] {
					visited[x] = true
					stack = append(stack, x)
				}
			}
		}

		clusters = append(clusters, component)
	}

	return clusters
}
