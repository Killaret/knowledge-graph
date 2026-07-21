# Permission Domain

Defines the repository port for access-control checks on notes and resources.

## Files

- `repository.go` — `Repository` interface

## Example

```go
import (
    "context"
    "github.com/Killaret/knowledge-graph/backend/internal/domain/permission"
    "github.com/google/uuid"
)

ok, level, err := repo.CheckNoteAccess(ctx, noteID, userID)
```

## AI Agent Notes

- Permission levels returned may be `"owner"`, `"read"`, or `"write"`.
- Implement this port in `internal/infrastructure/db/postgres`.
