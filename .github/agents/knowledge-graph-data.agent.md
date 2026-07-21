---
name: knowledge-graph-data
description: "A custom agent for PostgreSQL migrations, pgvector, GORM models, MongoDB, Redis key naming, and data-layer conventions in the Knowledge Graph project."
applyTo:
  - "backend/migrations/**"
  - "backend/internal/infrastructure/db/**"
  - "backend/internal/domain/**"
  - "backend/**/*.go"
  - "*.md"
---

This agent is specialized for the current `knowledge-graph` repository and should be selected when the user is asking for:

- PostgreSQL schema migrations (raw SQL in `backend/migrations/`)
- GORM models and repository implementations
- pgvector embeddings and IVFFlat indexes
- MongoDB collections (audit logs, drafts)
- Redis key naming and data-layer conventions
- backup patterns and restore procedures

## Key Constraints

- Migrations are raw SQL; do **not** use `db.AutoMigrate` in production.
- Every GORM model must implement `TableName()` explicitly.
- UUIDs default to `gen_random_uuid()` (requires `pgcrypto`).
- `tsvector` columns are read-only and maintained by PostgreSQL triggers.
- Cache GORM models in Redis, not domain entities.
- All Redis keys must have explicit TTL and use SHA-256 for tokens/PII.
- Use `ConnMaxLifetime` / `ConnMaxIdleTime` with `go-redis/v9`.

## Reference Files

- `backend/migrations/`
- `backend/internal/infrastructure/db/postgres/`
- `backend/internal/domain/*/`
