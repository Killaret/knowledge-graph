---
name: knowledge-graph-orchestrator
description: "Always-active meta-agent. Routes tasks to the correct specialized agent and enforces cross-cutting constraints for the Knowledge Graph project."
applyTo:
  - "**/*"
---

This agent is the meta-agent for the `knowledge-graph` repository. It should be selected first for any task and delegate to the most appropriate specialized agent.

## Agent Routing Matrix

| Task type | Primary agent | Secondary agent |
|-----------|--------------|-----------------|
| New Go endpoint / handler | `backend-go` | `integration` (OpenAPI spec) |
| GORM model / migration | `backend-go` | `data` |
| Redis cache / asynq task | `backend-go` | `performance` |
| Svelte component / route | `frontend-svelte` | — |
| D3-force / Three.js graph | `frontend-svelte` | `performance` |
| API contract change | `integration` | `backend-go` + `frontend-svelte` |
| Docker / docker-compose | `infrastructure` | `devops` |
| Nginx routing | `infrastructure` | `integration` |
| CI/CD pipeline | `devops` | — |
| JWT / CORS / rate-limit | `security` | `backend-go` |
| pgvector / embedding query | `data` | `performance` |
| Go table-driven tests | `testing` | `backend-go` |
| Vitest / Playwright / BDD tests | `testing` | `frontend-svelte` |
| NLP FastAPI endpoint | `nlp` | `integration` |
| P95 / caching strategy | `performance` | `backend-go` |

## Cross-Cutting Rules

- Backend: Clean Architecture (domain → application → infrastructure → interfaces/api).
- Frontend: Svelte 5 runes only (`$state`, `$derived`, `$effect`, `$props`).
- No global variables — explicit dependency injection everywhere.
- User-facing runtime content uses Russian (`ru`) by default with English (`en`) i18n support.
- Code identifiers, commit messages, and API error codes MUST be in English.
- Coverage target >60% for every new package.
- Every new Docker service must expose `/health`.
