# ADR 004: Soft Delete Strategy

## Status
Accepted

## Context
In a multi-tenant SaaS environment, accidental data deletion has severe consequences:
- User error ("I didn't mean to delete that note")
- Malicious insider (disgruntled employee)
- Bug in sync logic (draft overwrites published content)

We need a deletion strategy that balances user recovery needs with data retention compliance (GDPR "right to erasure").

## Problem
How do we handle data deletion with these requirements:
1. Allow user-initiated recovery within reasonable window
2. Prevent permanent loss from bugs or accidents
3. Comply with GDPR data erasure requests
4. Handle cascading deletes for linked resources
5. Maintain referential integrity in graph structures

## Decision
**Soft Delete with 30-Day Hard Delete Cleanup**

All entities use `deleted_at` timestamp. Background job permanently removes after retention period.

### Implementation

#### 1. Database Schema
```sql
-- All tenant tables include deleted_at
CREATE TABLE notes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    owner_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,  -- NULL = active, NOT NULL = soft-deleted
    
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (owner_id) REFERENCES users(id)
);

-- Index for efficient cleanup queries
CREATE INDEX idx_notes_deleted_at ON notes(deleted_at) 
    WHERE deleted_at IS NOT NULL;
```

#### 2. RLS Policy (includes soft-delete filter)
```sql
-- Active records only (excludes soft-deleted)
CREATE POLICY tenant_active_records ON notes
    FOR SELECT
    USING (
        tenant_id = current_setting('app.current_tenant_id')::UUID
        AND deleted_at IS NULL
    );

-- Update/Delete only non-deleted records
CREATE POLICY tenant_modify_active ON notes
    FOR ALL
    USING (
        tenant_id = current_setting('app.current_tenant_id')::UUID
        AND deleted_at IS NULL
    );
```

#### 3. Domain Layer
```go
type Note struct {
    id         uuid.UUID
    tenantID   uuid.UUID
    ownerID    uuid.UUID
    title      string
    content    string
    createdAt  time.Time
    updatedAt  time.Time
    deletedAt  *time.Time  // nil = active
}

// Soft delete - sets deleted_at
type SoftDeleteNoteCommand struct {
    NoteID   uuid.UUID
    TenantID uuid.UUID
    UserID   uuid.UUID
}

func (n *Note) SoftDelete(now time.Time) error {
    if n.deletedAt != nil {
        return ErrAlreadyDeleted
    }
    n.deletedAt = &now
    n.updatedAt = now
    return nil
}

// Restore from soft-delete
type RestoreNoteCommand struct {
    NoteID   uuid.UUID
    TenantID uuid.UUID
}

func (n *Note) Restore() error {
    if n.deletedAt == nil {
        return ErrNotDeleted
    }
    n.deletedAt = nil
    n.updatedAt = time.Now()
    return nil
}
```

#### 4. Cascade Rules
```go
// When a note is soft-deleted:
// 1. Soft-delete child notes (owned by same user)
// 2. Delete links FROM this note (graph edges)
// 3. Keep links TO this note (broken reference handling)

func (h *SoftDeleteNoteHandler) Execute(ctx context.Context, cmd SoftDeleteNoteCommand) error {
    note, err := h.repo.GetByID(ctx, cmd.NoteID)
    if err != nil {
        return err
    }
    
    // Verify ownership
    if note.OwnerID() != cmd.UserID {
        return ErrNotOwner
    }
    
    now := time.Now()
    
    // Soft-delete this note
    if err := note.SoftDelete(now); err != nil {
        return err
    }
    
    // Cascade: soft-delete children
    children, _ := h.repo.GetChildren(ctx, cmd.NoteID)
    for _, child := range children {
        child.SoftDelete(now)
        h.repo.Save(ctx, child)
    }
    
    // Cascade: remove outgoing links
    h.linkRepo.DeleteBySource(ctx, cmd.NoteID)
    
    return h.repo.Save(ctx, note)
}
```

#### 5. Hard Delete Cleanup Job
```go
// Background worker runs daily
type CleanupJob struct {
    db     *sql.DB
    logger Logger
}

func (j *CleanupJob) Run(ctx context.Context) error {
    retentionDays := 30
    
    // Permanently delete records past retention
    result, err := j.db.ExecContext(ctx, `
        DELETE FROM notes
        WHERE deleted_at < NOW() - INTERVAL '$1 days'
    `, retentionDays)
    
    if err != nil {
        j.logger.Error("cleanup failed", err)
        return err
    }
    
    rowsDeleted, _ := result.RowsAffected()
    j.logger.Info("hard deleted records", "count", rowsDeleted)
    return nil
}
```

#### 6. GDPR Hard Delete
```go
// Immediate permanent deletion for GDPR requests
type GDPRDeleteCommand struct {
    UserID   uuid.UUID
    TenantID uuid.UUID
}

func (h *GDPRDeleteHandler) Execute(ctx context.Context, cmd GDPRDeleteCommand) error {
    // Immediate hard delete, bypassing soft-delete
    // Requires elevated admin privileges
    return h.repo.HardDeleteByUser(ctx, cmd.UserID)
}
```

## Consequences

### Positive
- ✅ User recovery: 30-day window to restore accidentally deleted notes
- ✅ Bug resilience: Can undo batch operations gone wrong
- ✅ Audit trail: deleted_at timestamp records when deletion occurred
- ✅ Referential integrity: Links can detect broken references

### Negative
- ⚠️ Storage overhead: Deleted records consume space for 30 days
- ⚠️ Query complexity: Every query needs `deleted_at IS NULL` check (handled by RLS)
- ⚠️ Index bloat: Soft-deleted records remain in indexes
- ⚠️ Backup size: Larger backups include soft-deleted data

### Trade-offs
| Decision | Rationale |
|----------|-----------|
| 30-day retention | Balances recovery needs vs storage costs |
| Soft-delete children | Parent recovery should restore children |
| Hard-delete outgoing links | Prevents orphaned edges; incoming links stay for reference |

## Compliance
- **GDPR Article 17**: GDPRDeleteCommand provides immediate erasure
- **Default 30-day**: Provides "reasonable time" for user reconsideration
- **Audit requirement**: deleted_at timestamp logged to audit system

## References
- OWASP: "Secure Delete" patterns
- GDPR Article 17 - Right to erasure
