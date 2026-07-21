# User Domain

Defines the `User` aggregate, `APIKey` value object, and repository/role ports.

## Files

- `user.go` — `User` and `APIKey` entities
- `errors.go` — domain errors
- `*_test.go` — unit tests

## Example

```go
import (
    "time"
    "github.com/Killaret/knowledge-graph/backend/internal/domain/user"
    "github.com/google/uuid"
)

u, err := user.NewUser(
    uuid.New(), "login", "user@example.com", "hash", "user",
    time.Now(), time.Now(), nil,
)
```

## AI Agent Notes

- `NewUser` validates login, email, and password hash.
- Use `user.Repository` and `user.RoleRepository` ports from application/handlers.
- API keys are created via `APIKey` value object, not the `User` aggregate.
