# Share Domain

Defines direct user-to-user note shares (`NoteShare`) and public share links (`ShareLink`).

## Files

- `share.go` — `NoteShare` and `ShareLink` entities, validation errors
- `*_test.go` — unit tests

## Example

```go
import (
    "github.com/Killaret/knowledge-graph/backend/internal/domain/share"
    "github.com/google/uuid"
)

s, err := share.NewNoteShare(
    uuid.New(), noteID, sharedByUserID, sharedWithUserID, "read", nil,
)
if err != nil { return err }
```

## AI Agent Notes

- Valid permissions are `"read"` and `"write"`.
- `expiresAt` is optional.
- Repositories are responsible for checking that `noteID` exists and the sharer has permission.
