# ADR 003: Multi-Tenancy Strategy

## Status
Accepted

## Context
The Knowledge Graph platform is transitioning from single-user to multi-tenant SaaS. Data isolation is the highest priority security requirement. We must choose an approach that balances:
- Security (no cross-tenant data leakage)
- Operational complexity (migrations, backups, scaling)
- Cost efficiency (infrastructure per tenant)

## Problem
How do we isolate tenant data with the following constraints:
1. Small team - cannot manage hundreds of separate databases
2. SOC 2 compliance required (auditable isolation)
3. Need to support both shared workspaces and private notes
4. Must prevent accidents (developer error, SQL injection)

## Decision
**Shared Database + PostgreSQL Row-Level Security (RLS)**

We accept RLS complexity to avoid DB-per-tenant operational overhead.

### Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────────────────────┐
│   Client    │────▶│  Go App     │────▶│  PostgreSQL (Single DB)     │
│  (JWT)      │     │  (Set RLS)  │     │  ┌─────────────────────┐    │
└─────────────┘     └─────────────┘     │  │  RLS Policy Layer   │    │
                                        │  │  (Last line of      │    │
                                        │  │   defense)          │    │
                                        │  └─────────────────────┘    │
                                        │         │                   │
                                        │    ┌────┴────┐              │
                                        │    ▼         ▼              │
                                        │  tenant_a  tenant_b tables │
                                        └─────────────────────────────┘
```

### Implementation Details

#### 1. JWT Claims Structure
```json
{
  "sub": "user-uuid-123",
  "tenant_id": "tenant-uuid-456",
  "role": "member",
  "permissions": ["notes:read", "notes:write"],
  "iat": 1234567890,
  "exp": 1234571490
}
```

#### 2. Application Layer: Context Propagation
```go
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims := extractJWTClaims(r)
        
        // Set tenant context for this request
        ctx := WithTenantID(r.Context(), claims.TenantID)
        ctx = WithUserID(ctx, claims.Sub)
        ctx = WithRole(ctx, claims.Role)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (s *NoteService) CreateNote(ctx context.Context, cmd CreateNoteCommand) error {
    tenantID := TenantIDFromContext(ctx)
    
    // App-layer authorization check
    if !s.authz.HasPermission(ctx, "notes:write") {
        return ErrForbidden
    }
    
    // Set RLS context before DB operations
    if err := s.db.SetTenantContext(ctx, tenantID); err != nil {
        return err
    }
    
    return s.repo.Save(ctx, note)
}
```

#### 3. Database Layer: RLS Policies
```sql
-- Every table has tenant_id column
ALTER TABLE notes ADD COLUMN tenant_id UUID NOT NULL;

-- Enable RLS on all tenant tables
ALTER TABLE notes ENABLE ROW LEVEL SECURITY;
ALTER TABLE links ENABLE ROW LEVEL SECURITY;
ALTER TABLE embeddings ENABLE ROW LEVEL SECURITY;

-- Create policy that uses app.current_tenant_id
CREATE POLICY tenant_isolation ON notes
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Force RLS even for table owner (defense in depth)
ALTER TABLE notes FORCE ROW LEVEL SECURITY;

-- Helper function to set tenant context
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id UUID)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', tenant_id::text, false);
END;
$$ LANGUAGE plpgsql;
```

#### 4. Defense in Depth
```
┌─────────────────┐
│  1. JWT Validation │  ← Auth0/Okta verifies signature
├─────────────────┤
│  2. App Check     │  ← Handler checks permissions from JWT
├─────────────────┤
│  3. RLS Policy    │  ← Postgres filters rows by tenant_id
├─────────────────┤
│  4. Query Binding │  ← Parameterized queries prevent bypass
└─────────────────┘
```

## Consequences

### Positive
- ✅ Database-level enforcement: Bugs cannot bypass isolation
- ✅ Operational simplicity: One database to backup, migrate, monitor
- ✅ Cost efficiency: Shared resources, pay for what you use
- ✅ Cross-tenant analytics possible (with elevated privileges)
- ✅ Easy tenant provisioning (just insert row, no schema creation)

### Negative
- ⚠️ RLS adds ~5-10% query overhead (acceptable)
- ⚠️ Complex queries with JOINs need careful policy design
- ⚠️ Connection pooling requires context reset between requests
- ⚠️ Superuser can bypass RLS (operational risk to manage)

### Compliance
| Requirement | Implementation |
|-------------|----------------|
| Data Isolation | RLS policies + FORCE RLS |
| Audit Trail | Explicit audit_log table + MongoDB |
| Access Control | JWT claims + RBAC table |
| Breach Containment | Single tenant cannot access others' data even with SQL injection |

## Migration Notes
- See ADR 008 for migration from single-tenant to multi-tenant
- All existing data assigned to `default` tenant during migration

## References
- PostgreSQL RLS Documentation: https://www.postgresql.org/docs/current/ddl-rowsecurity.html
- OWASP Multi-Tenancy Security Guidelines
- SOC 2 CC6.1 Logical Access Security
