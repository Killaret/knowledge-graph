# Note Domain

Defines the `Note` aggregate and value objects (`Title`, `Content`, `Metadata`).

## Files

- `entity.go` — `Note` entity and constructors
- `value_objects.go` — `Title`, `Content`, `Metadata`
- `*_test.go` — unit tests

## Example

```go
import "github.com/Killaret/knowledge-graph/backend/internal/domain/note"

title, err := note.NewTitle("My Note")
if err != nil { return err }
content, err := note.NewContent("Body text")
if err != nil { return err }

n, err := note.NewNote(title, content, "star", note.Metadata{})
```

## AI Agent Notes

- Default note type is `"star"`.
- `Title` and `Content` validate non-empty and max-length rules.
- Repositories return `*note.Note`, not DB models.
