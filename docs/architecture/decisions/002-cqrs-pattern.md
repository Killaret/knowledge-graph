# ADR 002: CQRS-Lite Pattern

## Status
Accepted

## Context
As the Knowledge Graph platform grows from a single-user tool to a multi-tenant SaaS, we observe divergent access patterns:
- **Commands** (writes) require strict validation, business rule enforcement, and side-effect orchestration
- **Queries** (reads) need flexible projections, pagination, and search capabilities with different performance characteristics

We need a pattern that optimizes each path without over-engineering for our current scale.

## Problem
How do we structure our data access layer to:
1. Support complex read models without polluting the domain model
2. Maintain transactional consistency for writes
3. Avoid premature complexity (event sourcing, separate read DBs)
4. Keep the codebase maintainable for a small team

## Decision
**CQRS-Lite: Separate Command/Query handlers with Single Database**

We adopt a lightweight CQRS approach where:

### Command Path (Writes)
```go
// Application Layer
type CreateNoteCommand struct {
    TenantID    uuid.UUID
    OwnerID     uuid.UUID
    Title       string
    Content     string
    Tags        []string
}

type CreateNoteHandler struct {
    repo        domain.NoteRepository
    validator   *validation.Validator
    eventBus    EventBus
}

func (h *CreateNoteHandler) Execute(ctx context.Context, cmd CreateNoteCommand) (*NoteResult, error) {
    // 1. Validate input syntax
    if err := h.validator.Validate(cmd); err != nil {
        return nil, err
    }
    
    // 2. Check permissions (RBAC)
    if !h.authz.CanCreateNote(ctx, cmd.TenantID, cmd.OwnerID) {
        return nil, ErrUnauthorized
    }
    
    // 3. Execute domain logic
    note, err := domain.NewNote(cmd.TenantID, cmd.OwnerID, cmd.Title, cmd.Content)
    if err != nil {
        return nil, err // Business rule violation
    }
    
    // 4. Persist
    if err := h.repo.Save(ctx, note); err != nil {
        return nil, err
    }
    
    // 5. Emit side effects
    h.eventBus.Publish(NoteCreatedEvent{...})
    
    return &NoteResult{ID: note.ID}, nil
}
```

### Query Path (Reads)
```go
// Separate Query Handlers - no domain logic
type NoteListQuery struct {
    TenantID   uuid.UUID
    UserID     uuid.UUID
    Page       int
    PageSize   int
    SearchTerm string
    Tags       []string
}

type NoteListHandler struct {
    db *sqlx.DB  // Direct DB access, bypasses domain repository
}

func (h *NoteListHandler) Execute(ctx context.Context, q NoteListQuery) (*PaginatedResult, error) {
    // RLS automatically filters by tenant_id via app.current_tenant_id
    rows, err := h.db.QueryContext(ctx, `
        SELECT n.id, n.title, n.created_at, array_agg(t.name) as tags
        FROM notes n
        LEFT JOIN note_tags t ON n.id = t.note_id
        WHERE n.deleted_at IS NULL
          AND ($4 = '' OR n.title ILIKE '%' || $4 || '%')
          AND ($5 = '{}' OR t.name = ANY($5))
        GROUP BY n.id
        ORDER BY n.created_at DESC
        LIMIT $2 OFFSET $3
    `, q.TenantID, q.PageSize, (q.Page-1)*q.PageSize, q.SearchTerm, pq.Array(q.Tags))
    // ...
}
```

### Key Principles
1. **Single Database**: PostgreSQL handles both reads and writes. No event sourcing.
2. **No Shared Models**: Commands use domain entities; Queries use DTOs/projections
3. **Explicit Handlers**: Every use case has a dedicated handler class
4. **Read-Side Flexibility**: Queries can use raw SQL, materialized views, or computed columns

## Consequences

### Positive
- ✅ Commands benefit from rich domain model validation
- ✅ Queries can be optimized independently (indexes, denormalized views)
- ✅ Clear separation makes testing easier
- ✅ No infrastructure complexity of event sourcing or dual databases
- ✅ Team can evolve read models without touching domain logic

### Negative
- ⚠️ Code duplication between command and query models
- ⚠️ Eventual consistency not available (we accept this trade-off)
- ⚠️ Developers must understand which path to use
- ⚠️ No built-in audit log from event stream (we implement explicit audit instead)

### Trade-offs Accepted
| Trade-off | Rationale |
|-----------|-----------|
| Single DB for both | Operational simplicity; can split later if needed |
| No Event Sourcing | Team expertise; adds massive complexity |
| No Read Replicas | Current scale doesn't justify; add when QPS > 1000 |

## Compliance
- Aligns with Clean Architecture: Application layer orchestrates, Domain enforces rules
- Supports RLS: Both paths set `app.current_tenant_id` before queries

## References
- Vernon, V. "Implementing Domain-Driven Design" (Chapter on CQRS)
- Microsoft: "CQRS Pattern" - https://docs.microsoft.com/en-us/azure/architecture/patterns/cqrs
