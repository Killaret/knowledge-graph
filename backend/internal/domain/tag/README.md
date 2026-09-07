# Tag Domain

Defines the `Tag` entity for note labels.

## Files

- `entity.go` — `Tag` entity, `New`, `Reconstruct`, validation
- `*_test.go` — unit tests

## Example

```go
import "github.com/Killaret/knowledge-graph/backend/internal/domain/tag"

t, err := tag.New("my-tag")
if err != nil { return err }
```

## AI Agent Notes

- Tag names have a maximum length of 50 characters.
- Use `Reconstruct` only when loading persisted tags.
