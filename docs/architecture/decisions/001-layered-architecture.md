# ADR 001: Layered Architecture with Domain-Driven Design

## Status
Accepted (with pending SaaS evolution)

## Context
The project aims to demonstrate advanced architectural patterns suitable for a Senior System Architect portfolio. We need a structure that:
- Enforces clear separation of concerns
- Supports complex business logic (note management, AI embeddings, recommendations)
- Can evolve from a single-user demo to a multi-tenant SaaS
- Makes business rules explicit and testable

### Current State Analysis
Our current implementation follows a 4-layer DDD approach:
1. **Domain Layer**: Entities, Value Objects, Aggregates, Domain Services
2. **Application Layer**: Use cases, orchestration, transaction management
3. **Infrastructure Layer**: Database, external APIs, file storage
4. **Presentation Layer**: API controllers, UI

**Critical Gap Identified**: The current architecture lacks multi-tenancy support. All entities are user-agnostic, creating a fundamental security risk if exposed as SaaS.

## Problem Statement
How do we maintain clean DDD boundaries while evolving from a single-user demonstration to a secure multi-tenant SaaS platform without compromising the educational value of the architecture?

## Decision Drivers
- **Educational Value**: Must demonstrate proper DDD implementation
- **Security**: Data isolation between tenants is non-negotiable
- **Maintainability**: Changes should be localized, not ripple through all layers
- **Performance**: Tenant filtering must not degrade query performance
- **Testability**: Business rules must be testable in isolation

## Considered Options

### Option 1: Anemic Domain Model + Service Layer
Move all business logic to Application Services, keep entities as data containers.
- ✅ Simpler initial implementation
- ✅ Easier ORM mapping
- ❌ Violates DDD principles we aim to demonstrate
- ❌ Business rules scattered across services
- ❌ Harder to enforce invariants

### Option 2: Rich Domain Model (Current) + TenantID Injection
Keep rich domain model but add TenantID to all aggregates and enforce at repository level.
- ✅ Maintains DDD purity
- ✅ Business rules stay in domain
- ✅ Clear ownership of invariants
- ⚠️ Requires careful context propagation
- ⚠️ Repository interfaces need redesign

### Option 3: Database-per-Tenant
Separate database schema/instance per tenant.
- ✅ Maximum isolation
- ✅ Simplifies backup/restore per tenant
- ❌ Operational complexity (migration, scaling)
- ❌ Overkill for MVP/educational project
- ❌ Harder to demonstrate cross-tenant features later

## Decision Chosen
**Option 2: Rich Domain Model with Explicit Tenant Context**

We will:
1. Keep the 4-layer DDD structure as it demonstrates proper separation of concerns
2. Add `TenantId` as a fundamental property of all Aggregate Roots
3. Implement Row Level Security (RLS) in PostgreSQL as the primary isolation mechanism
4. Require tenant context in all repository queries (compile-time safety)
5. Perform authorization checks in Application Services before domain invocation

### Evolution Path
- **Phase 1 (Current)**: Single-user mode with TenantId = UserId for future compatibility
- **Phase 2 (SaaS)**: Full multi-tenancy with RLS and shared infrastructure
- **Phase 3 (Enterprise)**: Optional database-per-tenant for premium customers

## Consequences

### Positive
- ✅ Demonstrates senior-level understanding of DDD trade-offs
- ✅ Security built into the architecture, not bolted on
- ✅ Business rules remain testable and isolated
- ✅ Clear migration path to production SaaS

### Negative
- ⚠️ Increased initial complexity in repository layer
- ⚠️ Need for careful context management across layers
- ⚠️ Additional testing required for tenant isolation

### Risks & Mitigations
| Risk | Impact | Mitigation |
|------|--------|------------|
| Accidental cross-tenant data leak | Critical | RLS policies + integration tests with multiple tenants |
| Tenant context lost in async operations | High | Explicit context passing, no ambient state |
| Performance degradation from RLS | Medium | Proper indexing, query plan analysis |

## Compliance Notes
This decision aligns with:
- OWASP Multi-Tenancy Security Guidelines
- SOC 2 Data Isolation Requirements
- GDPR Data Segregation Principles

## Related Decisions
- ADR-002: CQRS Pattern Implementation
- ADR-003: Link Entity Design
- ADR-015: SaaS Migration Strategy (pending)

## References
- Evans, E. "Domain-Driven Design: Tackling Complexity in the Heart of Software"
- Microsoft Azure Architecture Center: Multi-tenancy guidance
- PostgreSQL RLS Documentation
