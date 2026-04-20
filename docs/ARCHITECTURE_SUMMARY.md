# Architecture Summary: Knowledge Graph SaaS

## Executive Summary

The Knowledge Graph platform is a **multi-tenant SaaS application** built on **Clean Architecture principles** with a **Go backend** and **Svelte frontend**. The architecture prioritizes:

1. **Data Isolation**: PostgreSQL Row-Level Security (RLS) enforces tenant boundaries at the database level
2. **Scalability**: CQRS-Lite pattern separates read/write concerns; Redis queues enable async processing
3. **Resilience**: Circuit breakers and fallback strategies prevent cascade failures
4. **Compliance**: Comprehensive audit logging with 90-day retention

### Key Architectural Decisions

| Decision | Rationale |
|----------|-----------|
| **Shared Database + RLS** | Accept RLS complexity to avoid DB-per-tenant operational overhead |
| **CQRS-Lite (Single DB)** | Optimize read/write paths without event sourcing complexity |
| **MongoDB for Logs/Drafts** | High-volume, write-heavy workloads isolated from transactional DB |
| **Rich Domain Model** | Business logic in entities, not anemic services |
| **Defense in Depth** | App-layer auth checks + RLS policies as last line |

## C4 Container Diagram

```mermaid
C4Container
    title Container Diagram - Knowledge Graph SaaS
    
    Person(user, "User", "Knowledge worker accessing notes")
    
    System_Boundary(saas, "Knowledge Graph SaaS") {
        Container(frontend, "Svelte SPA", "SvelteKit, TypeScript", "Note editor, graph visualization, search")
        
        Container(backend, "Go API", "Go 1.21, Chi Router", "Business logic, CQRS handlers, auth")
        
        ContainerDb(postgres, "PostgreSQL", "PostgreSQL 15", "Notes, links, users, tenants, RBAC<br/>RLS-enforced, transactional")
        
        ContainerDb(mongo, "MongoDB", "MongoDB 6", "Audit logs, draft autosaves<br/>High-volume, TTL expiry")
        
        Container(redis, "Redis", "Redis 7", "Async job queues<br/>Permission cache, rate limiting")
        
        Container(worker, "Background Workers", "Go", "Audit log persistence<br/>Draft sync, cleanup jobs")
    }
    
    System_Ext(auth0, "Auth0", "OIDC/OAuth2 provider<br/>JWT issuance")
    System_Ext(openai, "OpenAI API", "Embedding generation<br/>Subject to circuit breaker")
    
    Rel(user, frontend, "Uses", "HTTPS")
    Rel(frontend, backend, "API calls", "HTTPS/JSON, Bearer JWT")
    Rel(frontend, auth0, "Authenticates", "OAuth2 PKCE")
    
    Rel(backend, postgres, "Read/Write", "SQL with RLS")
    Rel(backend, mongo, "Write logs/drafts", "BSON")
    Rel(backend, redis, "Enqueue jobs", "LPUSH/BRPOP")
    Rel(backend, redis, "Cache permissions", "GET/SET")
    Rel(backend, openai, "Generate embeddings", "HTTPS (CB protected)")
    
    Rel(worker, postgres, "Process jobs", "SQL")
    Rel(worker, mongo, "Persist audit logs", "BSON bulk insert")
    Rel(worker, redis, "Dequeue jobs", "BRPOP")
    
    UpdateLayoutConfig($c4ShapeInRow="3", $c4BoundaryInRow="1")
```

## Data Flow: User Saves Draft

```mermaid
sequenceDiagram
    actor User
    participant Browser as Svelte SPA
    participant API as Go API
    participant Auth as Auth Middleware
    participant Redis as Redis Queue
    participant Mongo as MongoDB Drafts
    participant Postgres as PostgreSQL
    participant Worker as Background Worker

    User->>Browser: Types in note editor
    loop Every 30 seconds
        Browser->>API: POST /api/drafts/autosave
        Note over API: JWT validation
        API->>Auth: Extract tenant_id from JWT
        Auth-->>API: tenant_id, user_id
        API->>Mongo: Upsert draft document
        Mongo-->>API: Saved (version incremented)
        API-->>Browser: 200 OK {version: 5}
    end

    User->>Browser: Clicks "Publish"
    Browser->>API: POST /api/notes/publish
    API->>Mongo: Load draft by ID
    Mongo-->>API: Draft document
    
    API->>Auth: Check RBAC (notes:write)
    Auth-->>API: Authorized
    
    API->>Postgres: BEGIN TRANSACTION
    API->>Postgres: INSERT/UPDATE notes
    Note over Postgres: RLS validates tenant_id
    Postgres-->>API: Note persisted
    
    API->>Redis: Enqueue embedding job
    Redis-->>API: Job queued
    API->>Mongo: Delete draft (best effort)
    API->>Postgres: COMMIT
    
    API-->>Browser: 201 Created {note_id}
    
    Worker->>Redis: BRPOP embedding:queue
    Redis-->>Worker: Job payload
    Worker->>Postgres: Load note content
    Postgres-->>Worker: Note data
    Worker->>OpenAI: Generate embedding
    Worker->>Postgres: Save embedding
```

## Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Frontend** | SvelteKit + TypeScript | SPA with graph visualization |
| **API** | Go 1.21 + Chi | REST API, CQRS handlers |
| **Domain** | Pure Go structs | Rich entities, value objects |
| **Primary DB** | PostgreSQL 15 | Notes, links, users, tenants |
| **NoSQL** | MongoDB 6 | Audit logs, draft autosaves |
| **Cache/Queue** | Redis 7 | Job queues, permission cache |
| **Auth** | Auth0 (OIDC) | JWT issuance, MFA support |
| **Embeddings** | OpenAI API | Vector generation (CB protected) |
| **Circuit Breaker** | sony/gobreaker | Resilience patterns |

## Security Architecture

### Defense in Depth Layers

```
┌─────────────────────────────────────┐
│  Layer 4: PostgreSQL RLS          │
│  - Database enforces tenant_id      │
│  - FORCE RLS prevents bypass        │
├─────────────────────────────────────┤
│  Layer 3: Application RBAC          │
│  - JWT permission claims            │
│  - Role-based access control        │
│  - Resource ownership checks        │
├─────────────────────────────────────┤
│  Layer 2: Authentication            │
│  - Auth0 OIDC validates JWT         │
│  - Token expiration enforced          │
├─────────────────────────────────────┤
│  Layer 1: Transport Security        │
│  - TLS 1.3 for all connections        │
│  - CORS policy restrictions         │
└─────────────────────────────────────┘
```

### JWT Claims Structure

```json
{
  "sub": "user-uuid-123",
  "tenant_id": "tenant-uuid-456",
  "role": "member",
  "permissions": [
    "notes:read",
    "notes:write:own",
    "links:read"
  ],
  "iat": 1705312800,
  "exp": 1705399200
}
```

## Data Lifecycle

### Soft Delete Flow

```mermaid
stateDiagram-v2
    [*] --> Active: Create Note
    Active --> SoftDeleted: User deletes
    SoftDeleted --> Active: User restores (within 30 days)
    SoftDeleted --> HardDeleted: Cleanup job (after 30 days)
    HardDeleted --> [*]: Data permanently removed
    
    Active --> GDPRDeleted: GDPR erasure request
    GDPRDeleted --> [*]: Immediate hard delete
```

### Draft Synchronization

```mermaid
stateDiagram-v2
    [*] --> Editing: User opens editor
    Editing --> DraftSaved: Autosave (MongoDB)
    DraftSaved --> DraftSaved: Continue editing
    DraftSaved --> Published: User clicks Publish
    Published --> PostgresSync: Sync to PostgreSQL
    PostgresSync --> DraftDeleted: Remove from MongoDB
    DraftDeleted --> [*]: Draft lifecycle complete
    
    DraftSaved --> Abandoned: 30 days no activity
    Abandoned --> [*]: TTL cleanup
```

## Performance Characteristics

| Operation | Latency Target | Implementation |
|-----------|----------------|----------------|
| Draft autosave | < 100ms | MongoDB direct write |
| Note publish | < 500ms | Postgres transaction |
| Query list | < 200ms | CQRS query handler + indexes |
| Permission check | < 10ms | JWT claims (no DB hit) |
| Embedding generation | < 2s | Async worker, not blocking |

## Operational Considerations

### Monitoring
- Circuit breaker state changes emit alerts
- Redis queue depth monitored for backlog
- MongoDB TTL expiry tracked for compliance
- RLS policy effectiveness via query plans

### Backup Strategy
- PostgreSQL: Daily snapshots + WAL archiving
- MongoDB: Daily snapshots (audit logs = 90 days only)
- Redis: RDB snapshots (queues = ephemeral)

### Scaling Vectors
- **Read scaling**: Read replicas for CQRS queries
- **Write scaling**: Shard by tenant_id for MongoDB
- **Worker scaling**: Horizontal pod autoscaler on queue depth

## References

- [ADR 001: Layered Architecture](./architecture/decisions/001-layered-architecture.md)
- [ADR 003: Multi-Tenancy Strategy](./architecture/decisions/003-multi-tenancy-strategy.md)
- [ADR 009: Resilience Patterns](./architecture/decisions/009-resilience-patterns.md)
- [SaaS Database Schema](./SaaS_DATABASE_SCHEMA.md)
