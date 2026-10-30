# Graph Domain

Provides the graph traversal service used for recommendations: BFS neighbor loading, keyword matching, and weighted scoring.

## Files

- `traversal_service.go` — `TraversalService` orchestrator
- `neighbor_loader.go` / `keyword_matcher.go` — ports injected into the service
- `*_test.go` — unit tests

## Example

```go
import "github.com/Killaret/knowledge-graph/backend/internal/domain/graph"

svc := graph.NewTraversalServiceWithWeights(
    loader, 3, 0.5, "max", true, 0.5, 0.3, 0.2,
)
results, err := svc.Recommend(ctx, targetNoteID, limit)
```

## AI Agent Notes

- `alpha + beta + gamma` must equal `1.0`.
- `aggregation` is either `"max"` or `"sum"`.
- Implement `NeighborLoader` and `KeywordMatcher` in the application or infrastructure layer.
