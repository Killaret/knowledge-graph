# ADR 005: Validation Strategy

## Status
Accepted

## Context
Validation occurs at multiple layers in a Clean Architecture application. Without clear boundaries, we risk:
- Duplicating the same checks across layers
- Leaking infrastructure concerns into domain
- Allowing invalid states to persist

We need explicit rules for where different types of validation belong.

## Problem
How do we distribute validation responsibilities across layers:
1. Syntax validation (email format, required fields, string length)
2. Business rule validation (duplicate titles, circular references, quota limits)
3. Authorization (can this user perform this action)
4. Database constraints (foreign key existence, unique constraints)

## Decision
**Syntax in Application, Business Rules in Domain**

### Validation Pyramid
```
┌─────────────────────────────────────┐
│  Layer 4: Presentation              │
│  - Field format hints               │
│  - Client-side regex                │
│  (UX optimization, not security)  │
├─────────────────────────────────────┤
│  Layer 3: Application               │
│  - Syntax validation                │
│  - Schema validation                │
│  - Input sanitization               │
│  - Authorization checks             │
├─────────────────────────────────────┤
│  Layer 2: Domain                    │
│  - Business invariants              │
│  - Aggregate consistency            │
│  - Domain rule violations           │
├─────────────────────────────────────┤
│  Layer 1: Infrastructure            │
│  - Database constraints             │
│  - RLS policies                     │
│  (Last line of defense)             │
└─────────────────────────────────────┘
```

### Implementation

#### 1. Application Layer: Syntax Validation
```go
package application

import "github.com/go-playground/validator/v10"

type CreateNoteCommand struct {
    TenantID uuid.UUID `validate:"required"`
    OwnerID  uuid.UUID `validate:"required"`
    Title    string    `validate:"required,min=1,max=255"`
    Content  string    `validate:"max=100000"`
    Tags     []string  `validate:"dive,max=50"`
}

type CreateNoteHandler struct {
    validator *validator.Validate
    authz     Authorizer
    repo      domain.NoteRepository
}

func (h *CreateNoteHandler) Execute(ctx context.Context, cmd CreateNoteCommand) error {
    // Layer 3: Syntax validation
    if err := h.validator.Struct(cmd); err != nil {
        return ValidationError{Errors: err}
    }
    
    // Layer 3: Authorization (RBAC check)
    if !h.authz.HasPermission(ctx, cmd.TenantID, "notes:create") {
        return ErrForbidden
    }
    
    // Layer 3: Quota check (application concern)
    count, _ := h.repo.CountByOwner(ctx, cmd.TenantID, cmd.OwnerID)
    if count >= 1000 { // tenant-specific quota
        return ErrQuotaExceeded
    }
    
    // Hand off to domain for business rules
    note, err := domain.NewNote(cmd.TenantID, cmd.OwnerID, cmd.Title, cmd.Content)
    if err != nil {
        return err // Business rule violation from domain
    }
    
    return h.repo.Save(ctx, note)
}
```

#### 2. Domain Layer: Business Rules
```go
package domain

import (
    "errors"
    "strings"
)

var (
    ErrEmptyTitle      = errors.New("note title cannot be empty")
    ErrTitleTooLong    = errors.New("title exceeds 255 characters")
    ErrCircularLink    = errors.New("circular note link detected")
    ErrDuplicateTag    = errors.New("duplicate tag not allowed")
)

type Note struct {
    id       uuid.UUID
    tenantID uuid.UUID
    ownerID  uuid.UUID
    title    string
    content  string
    tags     []string
    links    []Link
}

// NewNote enforces creation-time invariants
func NewNote(tenantID, ownerID uuid.UUID, title, content string) (*Note, error) {
    // Business rule: title must be meaningful
    trimmed := strings.TrimSpace(title)
    if trimmed == "" {
        return nil, ErrEmptyTitle
    }
    if len(trimmed) > 255 {
        return nil, ErrTitleTooLong
    }
    
    return &Note{
        id:       uuid.New(),
        tenantID: tenantID,
        ownerID:  ownerID,
        title:    trimmed,
        content:  content,
        tags:     []string{},
        links:    []Link{},
    }, nil
}

// AddLink enforces graph invariants
func (n *Note) AddLink(targetID uuid.UUID, linkType LinkType) error {
    // Business rule: prevent self-reference
    if targetID == n.id {
        return ErrCircularLink
    }
    
    // Business rule: prevent duplicate links
    for _, link := range n.links {
        if link.TargetID() == targetID && link.Type() == linkType {
            return ErrDuplicateLink
        }
    }
    
    // Business rule: max links per note
    if len(n.links) >= 100 {
        return ErrMaxLinksExceeded
    }
    
    n.links = append(n.links, NewLink(n.id, targetID, linkType))
    return nil
}

// AddTag enforces tag invariants
func (n *Note) AddTag(tag string) error {
    normalized := strings.ToLower(strings.TrimSpace(tag))
    
    // Business rule: no duplicates
    for _, existing := range n.tags {
        if existing == normalized {
            return ErrDuplicateTag
        }
    }
    
    // Business rule: max tags
    if len(n.tags) >= 20 {
        return ErrMaxTagsExceeded
    }
    
    n.tags = append(n.tags, normalized)
    return nil
}
```

#### 3. Infrastructure Layer: Database Constraints
```sql
-- Physical constraints (last defense)
ALTER TABLE notes 
    ADD CONSTRAINT chk_title_not_empty CHECK (LENGTH(TRIM(title)) > 0),
    ADD CONSTRAINT chk_title_length CHECK (LENGTH(title) <= 255);

-- Unique constraints
ALTER TABLE users 
    ADD CONSTRAINT uniq_email_per_tenant 
    UNIQUE (tenant_id, email);

-- Foreign key constraints
ALTER TABLE notes 
    ADD CONSTRAINT fk_notes_tenant 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) 
    ON DELETE CASCADE;
```

## Consequences

### Positive
- ✅ Clear separation: each layer has distinct validation responsibility
- ✅ Domain purity: business rules live in domain, not scattered
- ✅ Testability: domain rules testable without infrastructure
- ✅ Security: defense in depth (each layer validates)
- ✅ Performance: syntax validation fails fast before heavy operations

### Negative
- ⚠️ Duplicate effort: some checks exist in multiple layers (necessary for security)
- ⚠️ Error handling complexity: different error types from each layer
- ⚠️ Maintenance: rules in two places can drift (tests must catch this)

### Error Mapping
```go
// HTTP status codes by validation layer
switch err := err.(type) {
case SyntaxValidationError:
    return 400 // Bad Request
case domain.BusinessRuleError:
    return 422 // Unprocessable Entity
case AuthorizationError:
    return 403 // Forbidden
case domain.ConstraintViolation:
    return 409 // Conflict
default:
    return 500 // Internal Server Error
}
```

## Compliance
- OWASP Input Validation Cheat Sheet
- Domain-Driven Design: "Make implicit concepts explicit"

## References
- Evans, E. "Domain-Driven Design" (Chapter on Specifications)
- Vernon, V. "Implementing Domain-Driven Design" (Validation chapter)
