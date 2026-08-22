---
name: knowledge-graph-performance
description: "A custom agent for profiling, caching strategies, P95 latency, Redis, pgvector, asynq offloading, and frontend bundle optimization in the Knowledge Graph project."
applyTo:
  - "backend/**/*.go"
  - "frontend/**/*.ts"
  - "frontend/**/*.svelte"
  - "frontend/vite.config.ts"
  - "*.md"
---

This agent is specialized for the current `knowledge-graph` repository and should be selected when the user is asking for:

- P95 latency targets and request profiling
- Redis caching patterns and key naming
- pgvector IVFFlat index tuning and cosine similarity queries
- asynq task offloading for heavy operations
- frontend bundle splitting and D3-force/Three.js performance
- connection pool sizing for PostgreSQL and Redis

## Key Targets

| Endpoint | P95 target |
|----------|-----------|
| `GET /api/v1/notes` | < 100ms |
| `GET /api/v1/notes/:id` | < 50ms |
| `GET /api/v1/graph/all` | < 300ms |
| `GET /me/graph/cached` | < 30ms |
| D3-force render (200 nodes) | < 16ms/frame |

## Constraints

- Cache GORM models in Redis, not domain entities (unexported fields serialize as `{}`).
- Use Redis `Scan` iterator; never `KEYS *` in production.
- Offload NLP embeddings and keyword extraction to asynq workers.
- Redis v9 API uses `ConnMaxLifetime` / `ConnMaxIdleTime`, not `MaxConnAge` / `IdleTimeout`.
