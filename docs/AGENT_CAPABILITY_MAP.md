# Agent Capability Map

> Quick reference for architectural sessions (DeepSeek and others).  
> For full agent descriptions see [AGENTS_EN.md](AGENTS_EN.md).

This document answers two questions before an architecture session starts:

1. **What can each agent implement?**
2. **What is intentionally out of scope for the project?**

---

## Agent Scope

| Agent | Can implement | Cannot implement | Typical handoff |
|-------|--------------|------------------|---------------|
| `knowledge-graph-orchestrator` | Task routing, multi-agent coordination, plan validation | Concrete code changes, DB migrations, UI implementation | Reads all agent rules; produces handoff blocks |
| `knowledge-graph-backend-go` | Go API endpoints, GORM repositories, PostgreSQL queries, Redis/asynq, JWT middleware, domain layer | Frontend code, NLP model training, infrastructure beyond Docker Compose basics | `integration` for DTO/OpenAPI, `testing` for Go tests |
| `knowledge-graph-frontend-svelte` | Svelte 5 components/routes, stores (runes), D3-force/Three.js graphs, ky API clients, Vitest unit tests | Backend code, DB schemas, CI/CD pipelines | `integration` for TS types, `testing` for E2E |
| `knowledge-graph-integration` | OpenAPI/Swagger annotations, request/response DTOs, nginx proxy routing, TypeScript API types | Business logic, DB queries, UI behavior | `backend-go` + `frontend-svelte` for contract sync |
| `knowledge-graph-infrastructure` | Docker multi-stage builds, docker-compose services, nginx config, health checks, volumes | Application code, security policies, CI/CD pipeline logic | `devops` for deployment, `security` for secrets/TLS |
| `knowledge-graph-devops` | CI/CD scripts, deployment automation, log aggregation, rollback procedures | Feature code, DB migrations, UI/UX | `infrastructure` for stack changes |
| `knowledge-graph-performance` | Redis caching, P95 profiling, asynq offloading, query/index optimization, bundle analysis | Security controls, domain invariants | `backend-go` + `data` for fixes |
| `knowledge-graph-security` | JWT/CORS/rate-limit reviews, secret management audits, vulnerability scans, auth flows | Performance optimization, feature implementation | `backend-go` for fixes, `infrastructure` for TLS/secrets |
| `knowledge-graph-testing` | Unit/integration/E2E/BDD tests, coverage analysis, testcontainers, Playwright/Cucumber | Production feature code | All agents whose code is being tested |
| `knowledge-graph-nlp` | Python FastAPI endpoints, sentence-transformers embeddings, HuggingFace offline inference, lazy model loading | Training new models from scratch, GPU cluster orchestration | `integration` for API contracts, `infrastructure` for model cache |
| `knowledge-graph-data` | SQL/GORM migrations, pgvector schemas/indexes, Redis key naming tables, MongoDB schemas | Application business logic, UI components | `backend-go` + `testing` for model/test sync |

---

## Cross-Cutting Out-of-Scope

Do **not** design or require the following in an architecture session unless the user explicitly asks:

- **WebSocket / real-time server-push** — the current stack is request/response (HTTP/gRPC) only.
- **CDN / edge caching** — no CDN agents; caching is in-app (Redis) or nginx static files.
- **Mobile apps / React Native / Flutter** — frontend is SvelteKit web only.
- **Push notifications** — email only (if configured); no push service.
- **Kubernetes / Helm / Terraform** — deployment is Docker Compose based.
- **Multi-region / active-active replication** — single-region Docker Compose stacks.
- **Custom ML model training** — use pre-cached HuggingFace sentence-transformers models.
- **OAuth providers beyond Yandex** — only Yandex OAuth is supported.
- **GraphQL** — REST/OpenAPI only.
- **Microservices beyond the existing set** — backend monolith + graph-service + NLP service.

---

## Conflict Resolution

When an architectural suggestion touches two agents with conflicting goals, use this priority:

1. **Security** — do not weaken auth/validation/secrets.
2. **Correctness** — domain invariants and API contracts.
3. **Performance** — caching, indexing, async offloading.
4. **Convenience** — developer experience, helper scripts.

If a proposed architecture requires an agent that does not exist (e.g., "mobile agent"), flag it as out-of-scope and propose a simpler alternative that uses existing agents.
