# Backend Domain Layer

This directory contains the pure domain layer of the Knowledge Graph backend. It follows Clean Architecture: no imports from `internal/infrastructure`, `internal/interfaces`, or any framework-specific packages (GORM, Gin, Redis, etc.).

## Packages

| Package | Purpose |
|---------|---------|
| `achievement` | Achievement definitions, conditions, user achievements, engine/counter ports |
| `cache` | `CacheClient` port implemented by Redis/miniredis/fake adapters |
| `graph` | Graph traversal service for recommendations (BFS, keyword matching, aggregation) |
| `link` | Link entity and value objects (LinkType, Weight, Metadata, SourceType) |
| `note` | Note entity and value objects (Title, Content, Metadata) |
| `permission` | Repository port for access checks on notes and resources |
| `share` | NoteShare and ShareLink entities |
| `tag` | Tag entity with name validation |
| `user` | User and APIKey aggregates, repository ports |

## Conventions

- Entities have private fields and public getters.
- Use `New...` factory functions for creation; they validate input and return errors.
- Repository interfaces are defined here and implemented in `internal/infrastructure/db/postgres` or `internal/infrastructure/cache`.
- Each package has its own `*_test.go` files using table-driven tests.

## Example

```go
import (
    "github.com/Killaret/knowledge-graph/backend/internal/domain/note"
    "github.com/google/uuid"
)

title, _ := note.NewTitle("My Note")
content, _ := note.NewContent("Note body")
n, err := note.NewNote(title, content, "star", note.Metadata{})
if err != nil {
    return err
}
_ = n.ID()
```
