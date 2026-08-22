# Achievement Domain

Defines the achievement system: achievements, conditions, user achievements, and the engine/counter/repository ports.

## Files

- `entity.go` — `Achievement`, `UserAchievement`, `Condition` entities; `Repository`, `Engine`, `Counter` ports
- `*_test.go` — table-driven unit tests

## Example

```go
import "github.com/Killaret/knowledge-graph/backend/internal/domain/achievement"

cond := achievement.Condition{
    Type:      "count",
    Entity:    "note",
    Action:    "create",
    Threshold: 10,
}
if err := cond.Validate(); err != nil {
    return err
}
```

## AI Agent Notes

- Use `achievement.NewAchievement` or repository ports from the application layer.
- Do not construct `Achievement` with zero values; use factories and validation.
