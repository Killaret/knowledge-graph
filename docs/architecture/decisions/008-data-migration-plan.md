# ADR 008: Data Migration Plan

## Status
Accepted

## Context
We are transitioning the Knowledge Graph platform from single-user MVP to multi-tenant SaaS. This requires:
- Creating tenant structure for existing users
- Adding `tenant_id` to all existing tables
- Enabling RLS policies
- Preserving all existing data

The migration must be reversible and minimize downtime.

## Problem
How do we migrate production data with these constraints:
1. Zero data loss (all notes, links, embeddings preserved)
2. Minimal downtime (users cannot be locked out for hours)
3. Reversible (rollback plan if issues arise)
4. All existing users become tenants with appropriate roles
5. RLS policies active immediately after migration

## Decision
**Maintenance Window Migration with Read-Only Phase**

### Pre-Migration Checklist
```bash
# 1. Full database backup (rollback point)
pg_dump -Fc knowledge_graph > backup_$(date +%Y%m%d_%H%M%S).dump

# 2. Verify backup integrity
pg_restore -l backup_*.dump > /dev/null && echo "Backup valid"

# 3. Announce maintenance window to users
# 4. Enable read-only mode in application
```

### Migration Script
```sql
-- MIGRATION SCRIPT: Single-Tenant to Multi-Tenant
-- Execute during maintenance window

BEGIN;

-- Step 1: Create tenants table
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Step 2: Create roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]',
    is_system BOOLEAN DEFAULT false,
    UNIQUE(tenant_id, name)
);

-- Step 3: Create tenant_memberships table
CREATE TABLE tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id),
    joined_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(tenant_id, user_id)
);

-- Step 4: Create one tenant per existing user
INSERT INTO tenants (id, name, slug)
SELECT 
    gen_random_uuid(),
    COALESCE(display_name, email) || '''s Workspace',
    LOWER(REGEXP_REPLACE(COALESCE(display_name, email), '[^a-zA-Z0-9]', '-', 'g'))
FROM users;

-- Step 5: Add tenant_id to users table
ALTER TABLE users ADD COLUMN tenant_id UUID;

-- Step 6: Link users to their personal tenant
UPDATE users u
SET tenant_id = t.id
FROM tenants t
WHERE t.name LIKE u.display_name || '%' 
   OR t.name LIKE u.email || '%';

-- Step 7: Create owner roles for each tenant
INSERT INTO roles (tenant_id, name, permissions, is_system)
SELECT t.id, 'owner', '["*"]', true
FROM tenants t;

-- Step 8: Seed tenant_memberships with owners
INSERT INTO tenant_memberships (tenant_id, user_id, role_id)
SELECT t.id, u.id, r.id
FROM tenants t
JOIN users u ON u.tenant_id = t.id
JOIN roles r ON r.tenant_id = t.id AND r.name = 'owner';

-- Step 9: Add tenant_id to notes table
ALTER TABLE notes ADD COLUMN tenant_id UUID;

-- Step 10: Backfill notes with tenant_id from owner
UPDATE notes n
SET tenant_id = u.tenant_id
FROM users u
WHERE n.owner_id = u.id;

-- Step 11: Make tenant_id NOT NULL after backfill
ALTER TABLE notes ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE notes ADD CONSTRAINT fk_notes_tenant 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

-- Step 12: Add tenant_id to links table
ALTER TABLE links ADD COLUMN tenant_id UUID;
UPDATE links l
SET tenant_id = n.tenant_id
FROM notes n
WHERE l.source_id = n.id;
ALTER TABLE links ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE links ADD CONSTRAINT fk_links_tenant 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

-- Step 13: Add tenant_id to embeddings table
ALTER TABLE embeddings ADD COLUMN tenant_id UUID;
UPDATE embeddings e
SET tenant_id = n.tenant_id
FROM notes n
WHERE e.note_id = n.id;
ALTER TABLE embeddings ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE embeddings ADD CONSTRAINT fk_embeddings_tenant 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id);

-- Step 14: Enable RLS on all tenant tables
ALTER TABLE notes ENABLE ROW LEVEL SECURITY;
ALTER TABLE links ENABLE ROW LEVEL SECURITY;
ALTER TABLE embeddings ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_memberships ENABLE ROW LEVEL SECURITY;

-- Step 15: Force RLS (even for table owners)
ALTER TABLE notes FORCE ROW LEVEL SECURITY;
ALTER TABLE links FORCE ROW LEVEL SECURITY;
ALTER TABLE embeddings FORCE ROW LEVEL SECURITY;
ALTER TABLE tenant_memberships FORCE ROW LEVEL SECURITY;

-- Step 16: Create RLS policies
CREATE POLICY tenant_isolation ON notes
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

CREATE POLICY tenant_isolation ON links
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

CREATE POLICY tenant_isolation ON embeddings
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

CREATE POLICY tenant_membership_isolation ON tenant_memberships
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Step 17: Add deleted_at for soft delete
ALTER TABLE notes ADD COLUMN deleted_at TIMESTAMP;
ALTER TABLE links ADD COLUMN deleted_at TIMESTAMP;
ALTER TABLE embeddings ADD COLUMN deleted_at TIMESTAMP;

-- Step 18: Update RLS to exclude soft-deleted
DROP POLICY tenant_isolation ON notes;
CREATE POLICY tenant_isolation ON notes
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID 
           AND deleted_at IS NULL);

-- Step 19: Create indexes for performance
CREATE INDEX idx_notes_tenant_deleted ON notes(tenant_id, deleted_at) 
    WHERE deleted_at IS NULL;
CREATE INDEX idx_links_tenant ON links(tenant_id);
CREATE INDEX idx_embeddings_tenant ON embeddings(tenant_id);

COMMIT;
```

### Application Changes
```go
// Enable read-only mode during migration
func MaintenanceMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if os.Getenv("MAINTENANCE_MODE") == "true" && r.Method != "GET" {
            http.Error(w, `{"error": "Service in maintenance mode"}`, 
                http.StatusServiceUnavailable)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// JWT update to include tenant context
func GenerateJWT(user User, tenant Tenant) (string, error) {
    claims := jwt.MapClaims{
        "sub":        user.ID.String(),
        "tenant_id":  tenant.ID.String(),
        "role":       "owner",
        "permissions": []string{"*"},
        "exp":        time.Now().Add(time.Hour * 24).Unix(),
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}
```

### Rollback Procedure
```bash
#!/bin/bash
# rollback.sh - Execute if migration fails

# 1. Stop application
kubectl scale deployment app --replicas=0

# 2. Restore from backup
pg_restore -d knowledge_graph --clean --if-exists backup_*.dump

# 3. Verify data integrity
psql -d knowledge_graph -c "SELECT COUNT(*) FROM notes;"

# 4. Restart application
kubectl scale deployment app --replicas=3

# 5. Notify users of extended maintenance
```

## Consequences

### Positive
- ✅ Reversible: Full backup enables complete rollback
- ✅ No data loss: All notes, links, embeddings preserved
- ✅ Minimal complexity: Single transaction ensures consistency
- ✅ Fast execution: Backfills use efficient UPDATE...FROM patterns

### Negative
- ⚠️ Downtime required: Users locked out during migration window
- ⚠️ Risk of partial failure: Long transaction may have issues
- ⚠️ One-way migration: Rollback loses changes made post-migration

### Migration Timeline
| Phase | Duration | User Impact |
|-------|----------|-------------|
| Backup | 5-10 min | Read-only mode active |
| Schema changes | 2-3 min | Read-only mode active |
| Data backfill | 10-30 min | Read-only mode active |
| RLS activation | 1 min | Read-only mode active |
| Verification | 5 min | Read-only mode active |
| **Total** | **25-50 min** | **Read-only window** |

## Compliance
- Backup retention: 30 days minimum
- Migration runbook: Documented and tested in staging
- Rollback tested: Full restore procedure validated quarterly

## References
- PostgreSQL pg_dump/pg_restore documentation
- Zero-downtime migration patterns (GitHub blog)
