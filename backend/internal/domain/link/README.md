# Link Domain

Defines the `Link` entity and value objects that connect notes in the graph.

## Files

- `entity.go` — `Link` entity and constructors
- `value_objects.go` — `LinkType`, `Weight`, `Metadata`, `SourceType`
- `*_test.go` — unit tests

## Example

```go
import (
    "github.com/Killaret/knowledge-graph/backend/internal/domain/link"
    "github.com/google/uuid"
)

source := uuid.MustParse("...")
target := uuid.MustParse("...")
l := link.NewLink(source, target, link.TypeReference, link.NewWeight(0.8), link.Metadata{})
```

## AI Agent Notes

- Use `link.NewLink` for normal user-created links.
- Use `link.NewGammaLink` for recommendation-generated links.
- Link types are defined in `value_objects.go`.
