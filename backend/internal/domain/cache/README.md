# Cache Domain

Defines the `CacheClient` port used by the application and infrastructure layers.

## Files

- `client.go` — `CacheClient` interface
- `cachetest/fake.go` — in-memory fake implementation for tests

## Example

```go
import (
    "context"
    "time"
    "github.com/Killaret/knowledge-graph/backend/internal/domain/cache"
)

ctx := context.Background()
err := client.Set(ctx, "key", "value", 5*time.Minute)
v, err := client.Get(ctx, "key") // returns "value" or cache.ErrCacheMiss
```

## AI Agent Notes

- Always accept `cache.CacheClient` as a dependency; never import Redis here.
- Use `cachetest.NewFakeClient()` in domain/application unit tests.
