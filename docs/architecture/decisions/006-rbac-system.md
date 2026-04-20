# ADR 006: RBAC System

## Status
Accepted

## Context
Multi-tenant SaaS requires fine-grained access control. Different roles need different permissions:
- Owner: full tenant administration
- Admin: manage members, billing
- Member: create/edit notes
- Viewer: read-only access

Permissions must be flexible (tenant-specific features) and performant (checked on every request).

## Problem
How do we implement RBAC that:
1. Supports customizable roles per tenant
2. Performs permission checks without DB round-trip per request
3. Integrates with JWT authentication
4. Allows role hierarchy (admin inherits member permissions)
5. Supports permission caching with invalidation

## Decision
**JSONB Permissions in Postgres + JWT Claims + App-Layer Caching**

### Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────────────────────┐
│   Client    │────▶│  Auth       │────▶│  JWT with claims:           │
│             │     │  Provider   │     │  { tenant_id, role, perms } │
└─────────────┘     └─────────────┘     └─────────────────────────────┘
                                                 │
                                                 ▼
┌─────────────┐     ┌─────────────┐     ┌─────────────────────────────┐
│   API       │◀────│  Go App     │◀────│  Check JWT perms first      │
│   Response  │     │  (Cache)    │     │  Fallback to DB if needed   │
└─────────────┘     └─────────────┘     └─────────────────────────────┘
                            │
                            ▼
                   ┌─────────────────┐
                   │  roles table:   │
                   │  {              │
                   │    name,        │
                   │    tenant_id,   │
                   │    permissions: │
                   │     JSONB[      │
                   │      "notes:read",
                   │      "notes:write"
                   │     ]           │
                   │  }              │
                   └─────────────────┘
```

### Implementation

#### 1. Database Schema
```sql
-- Roles defined per tenant
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]',
    is_system BOOLEAN DEFAULT false,  -- Built-in roles (owner, admin, member)
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(tenant_id, name)
);

-- Tenant memberships link users to roles
CREATE TABLE tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id),
    invited_by UUID REFERENCES users(id),
    joined_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(tenant_id, user_id)
);

-- Index for quick lookups
CREATE INDEX idx_memberships_user ON tenant_memberships(user_id);
CREATE INDEX idx_memberships_tenant ON tenant_memberships(tenant_id);

-- System roles (auto-created with each tenant)
INSERT INTO roles (tenant_id, name, permissions, is_system) VALUES
('00000000-0000-0000-0000-000000000000', 'owner', '[
    "tenant:*",
    "billing:*",
    "members:*",
    "notes:*",
    "links:*",
    "settings:*"
]', true),
('00000000-0000-0000-0000-000000000000', 'admin', '[
    "members:read", "members:write", "members:delete",
    "notes:*", "links:*", "settings:read", "settings:write"
]', true),
('00000000-0000-0000-0000-000000000000', 'member', '[
    "notes:read", "notes:write", "notes:delete:own",
    "links:read", "links:write:own",
    "settings:read"
]', true);
```

#### 2. Permission Structure (JSONB)
```json
{
  "permissions": [
    "notes:read",
    "notes:write",
    "notes:delete:own",
    "links:read",
    "links:write:own",
    "members:read",
    "settings:read"
  ]
}
```

Permission format: `{resource}:{action}:{scope?}`
- `*` = wildcard (all resources/actions)
- `:own` = scoped to user's own resources
- `:any` = can act on any resource in tenant

#### 3. JWT Claims Integration
```go
type JWTClaims struct {
    Sub       string   `json:"sub"`        // user_id
    TenantID  string   `json:"tenant_id"`  // current tenant
    Role      string   `json:"role"`       // role name
    Perms     []string `json:"permissions"` // cached permissions
    Exp       int64    `json:"exp"`
}

func (c *JWTClaims) HasPermission(required string) bool {
    for _, perm := range c.Perms {
        if matches(perm, required) {
            return true
        }
    }
    return false
}

// Wildcard matching: "notes:*" matches "notes:write"
func matches(granted, required string) bool {
    gParts := strings.Split(granted, ":")
    rParts := strings.Split(required, ":")
    
    for i := 0; i < len(rParts); i++ {
        if i >= len(gParts) {
            return false
        }
        if gParts[i] == "*" {
            return true
        }
        if gParts[i] != rParts[i] {
            return false
        }
    }
    return len(gParts) == len(rParts)
}
```

#### 4. Application Layer Authorization
```go
type Authorizer struct {
    roleRepo    RoleRepository
    cache       Cache  // Redis for permission caching
    cacheTTL    time.Duration
}

func (a *Authorizer) Authorize(ctx context.Context, userID, tenantID uuid.UUID, requiredPerm string) error {
    // 1. Check JWT claims first (fast path)
    claims := JWTFromContext(ctx)
    if claims.TenantID == tenantID.String() && claims.HasPermission(requiredPerm) {
        return nil
    }
    
    // 2. Fallback: load from cache
    cacheKey := fmt.Sprintf("perms:%s:%s", tenantID, userID)
    if cached, ok := a.cache.Get(cacheKey); ok {
        perms := cached.([]string)
        if hasPermission(perms, requiredPerm) {
            return nil
        }
        return ErrForbidden
    }
    
    // 3. Fallback: load from DB
    membership, err := a.roleRepo.GetMembership(ctx, tenantID, userID)
    if err != nil {
        return err
    }
    
    role, err := a.roleRepo.GetByID(ctx, membership.RoleID)
    if err != nil {
        return err
    }
    
    // Cache permissions
    a.cache.Set(cacheKey, role.Permissions, a.cacheTTL)
    
    if hasPermission(role.Permissions, requiredPerm) {
        return nil
    }
    return ErrForbidden
}

// Middleware integration
func PermissionMiddleware(authorizer *Authorizer, required string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := JWTFromContext(r.Context())
            
            if err := authorizer.Authorize(r.Context(), 
                uuid.MustParse(claims.Sub),
                uuid.MustParse(claims.TenantID),
                required); err != nil {
                
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

#### 5. Resource Ownership Check
```go
func (a *Authorizer) CanModifyNote(ctx context.Context, userID uuid.UUID, note *domain.Note) error {
    // Check generic permission
    if err := a.Authorize(ctx, userID, note.TenantID(), "notes:write:any"); err == nil {
        return nil
    }
    
    // Check ownership-scoped permission
    if err := a.Authorize(ctx, userID, note.TenantID(), "notes:write:own"); err != nil {
        return err
    }
    
    // Verify ownership
    if note.OwnerID() != userID {
        return ErrNotOwner
    }
    
    return nil
}
```

## Consequences

### Positive
- ✅ Fast permission checks via JWT claims (no DB hit)
- ✅ Flexible JSONB schema allows tenant customization
- ✅ Cache reduces DB load for permission lookups
- ✅ Wildcard permissions reduce JWT size
- ✅ Scope modifiers (:own/:any) enable fine-grained control

### Negative
- ⚠️ Permission changes require JWT refresh or cache invalidation
- ⚠️ JSONB queries slightly slower than normalized tables
- ⚠️ Cache invalidation complexity on role updates
- ⚠️ Permission drift possible if cache not invalidated properly

### Cache Invalidation Strategy
```go
func (a *Authorizer) InvalidateForTenant(tenantID uuid.UUID) error {
    // Pattern delete: perms:{tenantID}:*
    return a.cache.DeletePattern(fmt.Sprintf("perms:%s:*", tenantID))
}

func (a *Authorizer) InvalidateForUser(userID uuid.UUID) error {
    // Get all tenant memberships for user
    memberships, _ := a.roleRepo.GetUserMemberships(context.Background(), userID)
    for _, m := range memberships {
        a.cache.Delete(fmt.Sprintf("perms:%s:%s", m.TenantID, userID))
    }
    return nil
}
```

## Compliance
- Principle of least privilege: Default roles follow minimal permission sets
- Defense in depth: App check + RLS validation
- Audit trail: All authorization decisions logged

## References
- NIST RBAC Standard
- OWASP Authorization Cheat Sheet
