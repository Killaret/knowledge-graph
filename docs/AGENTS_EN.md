# AI Agents in Knowledge Graph

This document describes the 11 AI agent roles used across Cursor AI, Koda (Continue.dev), and Devin.

**Agent rules locations:**
- Cursor AI: `.cursor/rules/*.md`
- Koda (VSCode): `.continue/rules/*.md`
- Devin: `.devin/skills/knowledge-graph/SKILL.md`

---

## Agent List

### `knowledge-graph-orchestrator`
- **Focus:** Task routing & delegation across all agents
- **Ideal for:** Analyzing user requests, delegating to specialized agents

**Key Responsibilities:**
- Always active on startup — analyze request before delegating
- Route tasks to appropriate specialized agents
- Coordinate multi-agent tasks (e.g., backend + testing)
- Enforce consistency across domains

**Routing Matrix:**

| Task | Agent |
|------|-------|
| Go API, endpoints, DB queries | `backend-go` |
| Svelte components, UI, stores | `frontend-svelte` |
| OpenAPI, DTOs, webhooks | `integration` |
| Docker, nginx, volumes | `infrastructure` |
| CI/CD, deployment, logs | `devops` |
| Redis caching, P95, profiling | `performance` |
| Auth, JWT, vulnerabilities | `security` |
| Unit/E2E/BDD tests | `testing` |
| Python NLP, embeddings | `nlp` |
| Migrations, pgvector, schemas | `data` |

**Example Prompts:**
- "Route this task to the right agent"
- "Which agent handles pgvector queries?"
- "Coordinate backend + testing for this feature"

---

### `knowledge-graph-backend-go`
- **Focus:** `backend/`, Go 1.25, Gin, GORM, PostgreSQL, Redis, MongoDB, JWT, asynq
- **Ideal for:** Backend API, server logic, database queries, async tasks

**Key Responsibilities:**
- Go API and server logic (Gin v1.12)
- Clean Architecture layers: `domain/ → application/ → infrastructure/ → interfaces/api/`
- PostgreSQL (GORM + pgx/v5), Redis (go-redis/v9), MongoDB (audit logs, drafts)
- JWT authorization (golang-jwt/jwt/v5)
- Async task processing (asynq v0.23)
- Vector search (pgvector-go v0.2)
- Build, testing, documentation

**Anti-patterns:**
- ❌ Global `var db *sql.DB` — use dependency injection
- ❌ Panic in business logic — return errors explicitly
- ❌ Mixing domain and infrastructure layers
- ❌ Redis v8 API (MaxConnAge/IdleTimeout) — use v9 (ConnMaxIdleTime/ConnMaxLifetime)

**Example Prompts:**
- "Write a new Gin endpoint for notes"
- "Add Redis caching with go-redis/v9"
- "Implement JWT middleware"
- "Optimize PostgreSQL query with pgvector"

---

### `knowledge-graph-frontend-svelte`
- **Focus:** `frontend/`, Svelte 5, TypeScript strict, SvelteKit, D3-force, Three.js
- **Ideal for:** UI components, state management, graph visualization

**Key Responsibilities:**
- Svelte 5 runes ONLY: `$state`, `$derived`, `$effect`, `$props`, `$bindable`
- Atomic design: atoms → molecules → organisms
- TypeScript strict mode, SvelteKit routing
- API integration with ky v1.14
- D3-force graph and Three.js 3D visualization
- Vitest unit tests + Playwright E2E

**Anti-patterns:**
- ❌ Svelte 4 `writable()`/`readable()` stores — use `$state` runes
- ❌ Business logic inside components — use `lib/` services
- ❌ Direct fetch() calls — use ky client from `lib/`

**Example Prompts:**
- "Create a Svelte 5 component for the notes list"
- "Add $state store for graph nodes"
- "Implement D3-force graph animation"

---

### `knowledge-graph-integration`
- **Focus:** API contracts, OpenAPI/Swagger, DTOs, webhooks, OAuth
- **Ideal for:** API specification, DTO generation, protocol compliance

**Key Responsibilities:**
- OpenAPI/Swagger annotations (swaggo/gin-swagger)
- Request/response DTO structs
- Nginx proxy routing (`/api/*` → backend:9000, `/graph-service/api/*` → graph-service:9091)
- API versioning conventions (`/api/v1/`)
- Webhook receivers
- Error response format standards

**Example Prompts:**
- "Create OpenAPI schema for new endpoint"
- "Generate DTO for note creation request"
- "Check nginx proxy routing for graph-service"

---

### `knowledge-graph-infrastructure`
- **Focus:** Docker, containerization, nginx, monitoring, backups, scaling
- **Ideal for:** Container setup, service configuration, availability

**Key Responsibilities:**
- Docker multi-stage builds (builder + runtime stages)
- Health checks on all services (`/health` endpoint required)
- Volume naming: `postgres_data`, `redis_data`, `huggingface_cache`
- Dev stack (docker-compose.yml) vs Personal stack (docker-compose.personal.yml)
- Nginx gateway configuration
- Environment variable management

**Example Prompts:**
- "Add health check to new Docker service"
- "Configure nginx proxy for new microservice"
- "Set up Yandex.Disk backup"

---

### `knowledge-graph-devops`
- **Focus:** CI/CD, deployment, logs, metrics, rollbacks, release automation
- **Ideal for:** Pipeline setup, deployment automation

**Key Responsibilities:**
- CI/CD pipelines and deployment automation
- Log monitoring: `docker compose logs -f backend`
- Health check verification after deployment
- Rollback procedures
- Dev vs personal stack deployment

**Example Prompts:**
- "Set up CI/CD pipeline for Go backend"
- "Check deployment health status"
- "Configure automated release process"

---

### `knowledge-graph-performance`
- **Focus:** Profiling, load testing, P95, response time, Redis caching, asynq
- **Ideal for:** Performance analysis, caching strategies, async offloading

**Key Responsibilities:**
- Redis caching with go-redis/v9
- P95 response time monitoring
- pgvector query optimization
- Async task offloading with asynq
- HuggingFace model caching (lazy load pattern)
- Frontend bundle optimization

**Example Prompts:**
- "Analyze API endpoint performance"
- "Add Redis caching for expensive query"
- "Move heavy processing to asynq background task"

---

### `knowledge-graph-security`
- **Focus:** Security audit, vulnerability scanning, Auth/AuthZ, encryption, compliance
- **Ideal for:** Security reviews, authentication, data protection

**Key Responsibilities:**
- Security audit and vulnerability scanning
- JWT validation middleware (golang-jwt/jwt/v5)
- CORS configuration
- Rate limiting for write operations (POST/PUT/DELETE)
- Secret management (never commit .env, tokens, OAuth keys)
- Input validation (go-playground/validator)

**Anti-patterns:**
- ❌ Secrets in code or git history
- ❌ JWT validation in handlers (must be in middleware)
- ❌ Missing rate limiting on write endpoints

**Example Prompts:**
- "Review JWT authentication implementation"
- "Add rate limiting to note creation endpoint"
- "Audit CORS configuration"

---

### `knowledge-graph-testing`
- **Focus:** Unit, integration, E2E, BDD tests, coverage, stability
- **Ideal for:** Test writing, coverage analysis, test automation

**Key Responsibilities:**
- Go table-driven tests with testify/require
- testcontainers-go for integration tests (PostgreSQL module)
- miniredis v2 for Redis unit tests
- Vitest + @testing-library/svelte for frontend
- Playwright E2E tests
- BDD/Cucumber tests (cucumber.mjs config)
- Coverage enforcement: `go test -coverprofile`, `vitest --coverage`

**Coverage requirement: >60% for all modules**

**Example Prompts:**
- "Write table-driven unit tests for note handler"
- "Add integration test with testcontainers-go"
- "Create Playwright E2E test for graph view"

---

### `knowledge-graph-nlp` *(NEW)*
- **Focus:** `nlp-service/`, Python FastAPI, sentence-transformers, HuggingFace embeddings
- **Ideal for:** NLP service development, embedding API, keyword extraction

**Key Responsibilities:**
- FastAPI endpoint development (`/health`, `/embed`, `/extract_keywords`)
- Lazy model loading via `ensure_model_loaded()` and `get_embedding_model()` from `nlp_utils`
- HF_HUB_OFFLINE=1 support for offline-first operation
- HF_HOME cache path configuration
- Uvicorn startup optimization (~1s startup, ~15s model load from cache)
- Pydantic request/response models
- pytest tests for NLP endpoints

**Anti-patterns:**
- ❌ Loading model at module import (breaks lazy loading)
- ❌ Not handling model-not-loaded state in endpoints
- ❌ Missing HF_HUB_OFFLINE support

**Example Prompts:**
- "Add new NLP endpoint for text summarization"
- "Optimize model loading for faster startup"
- "Write pytest tests for /embed endpoint"

---

### `knowledge-graph-data` *(NEW)*
- **Focus:** Database schemas, migrations, GORM models, pgvector, MongoDB collections
- **Ideal for:** Schema design, migration scripts, data layer architecture

**Key Responsibilities:**
- GORM model conventions (soft deletes with gorm.Model)
- PostgreSQL migration scripts (raw SQL in migrations/)
- pgvector column types and similarity search queries
- MongoDB collections: `audit_logs`, `drafts` (go.mongodb.org/mongo-driver)
- Redis key naming conventions and TTL strategies
- Database connection pooling configuration
- REDIS_FLUSH_ON_STARTUP flag for cache control
- Yandex.Disk backup configuration
- Transaction patterns with GORM

**Example Prompts:**
- "Create migration for new user_points table"
- "Add pgvector column for note embeddings"
- "Design MongoDB schema for audit log"
- "Optimize Redis key structure for user sessions"

---

## Agent Selection Matrix

| Task Type | Agent |
|-----------|-------|
| Go API endpoint creation | `knowledge-graph-backend-go` |
| PostgreSQL / Redis / MongoDB | `knowledge-graph-backend-go` |
| JWT / authentication | `knowledge-graph-security` |
| Svelte 5 component | `knowledge-graph-frontend-svelte` |
| D3-force / Three.js | `knowledge-graph-frontend-svelte` |
| State management ($state) | `knowledge-graph-frontend-svelte` |
| OpenAPI spec / DTOs | `knowledge-graph-integration` |
| Nginx proxy routing | `knowledge-graph-integration` |
| Docker / containers | `knowledge-graph-infrastructure` |
| Health checks | `knowledge-graph-infrastructure` |
| CI/CD pipelines | `knowledge-graph-devops` |
| Deployment automation | `knowledge-graph-devops` |
| Redis caching strategy | `knowledge-graph-performance` |
| asynq background tasks | `knowledge-graph-performance` |
| Security audit | `knowledge-graph-security` |
| Rate limiting | `knowledge-graph-security` |
| Unit/E2E/BDD tests | `knowledge-graph-testing` |
| Test coverage analysis | `knowledge-graph-testing` |
| Python NLP / FastAPI | `knowledge-graph-nlp` |
| HuggingFace embeddings | `knowledge-graph-nlp` |
| DB migrations / schema | `knowledge-graph-data` |
| pgvector / MongoDB schema | `knowledge-graph-data` |
| Task routing | `knowledge-graph-orchestrator` |

---

## Quick Start Commands

### Backend
```bash
cd backend
go run ./cmd/server                               # Start server
go test ./... -v                                  # Unit tests
go test -tags=integration ./...                   # Integration tests
go test -coverprofile=coverage.out ./...          # With coverage
go build ./cmd/server                             # Build
```

### Frontend
```bash
cd frontend
npm run dev                    # Dev server
npm run test:unit              # Vitest unit tests
npm run test                   # Playwright E2E tests
npm run test:bdd               # Cucumber BDD tests
npm run build                  # Production build
```

### NLP Service
```bash
cd nlp-service
uvicorn app.main:app --reload  # Dev server (Uvicorn ~1s)
pytest tests/ -v               # Tests
```

### Full Stack
```bash
docker compose up -d                                    # Dev stack
docker compose -f docker-compose.personal.yml up -d    # Personal stack
docker compose logs -f backend                         # Backend logs
```

### Health Checks
```bash
curl http://localhost:8080/health                       # Nginx gateway
curl http://localhost:9000/health                       # Backend
curl http://localhost:9091/health                       # Graph service
curl http://localhost:8085/health                       # Personal backend
```

---

## AI Tool Configuration

| Tool | Rules Location | Format |
|------|---------------|--------|
| Cursor AI | `.cursor/rules/*.md` | Markdown, always loaded |
| Koda VSCode (koda-base/koda-pro) | `.continue/rules/*.md` | Markdown with YAML frontmatter |
| Devin | `.devin/skills/knowledge-graph/SKILL.md` | Markdown skill file |
| Windsurf/Cascade | `.windsurfrules` | Markdown, auto-loaded |

---

## Language Policy

### English Required For:
- **UI strings** — buttons, labels, placeholders, error messages
- **Toast notifications** — GalacticLexicon messages
- **Note titles and content** — all user-created content
- **Code comments** — public API docs, README files
- **Commit messages** — clear, descriptive English
- **Documentation** — all files in `docs/`

### Not Required:
- **Internal code comments** — brief explanations in any language OK
- **Variable/function names** — follow project conventions

```typescript
// ✅ Good
toast.success("Note created successfully");
tooltip: "New star mapped to the galactic chart, captain!";

// ❌ Bad
toast.success("Заметка создана успешно");
```

---

## Agent Workflow

```
User Request
    ↓
knowledge-graph-orchestrator (analyzes + routes)
    ↓
Specialized Agent (executes within domain)
    ↓
knowledge-graph-testing (validates changes)
    ↓
Result with documentation update
```

---

## Best Practices

1. **Be specific** — specify exact files and task types in prompts
2. **Use the right agent** — ensures correct domain context
3. **Follow project patterns** — read existing code before writing new
4. **Verify with tests** — always run tests after changes
5. **English only** — all user-facing content must be in English
6. **Cross-reference agents** — some tasks need multiple agents (e.g., backend + security)
7. **Check anti-patterns** — each agent rule file lists what NOT to do

---

## Notes

- Agents are **not executable scripts** — they are metadata and instructions for AI tools
- All 11 agents are defined in both `.cursor/rules/` (Cursor) and `.continue/rules/` (Koda)
- `.windsurfrules` provides a unified overview for Windsurf/Cascade
- `.devin/skills/knowledge-graph/SKILL.md` provides context for Devin

---

*Last updated: July 2026 — Added knowledge-graph-nlp and knowledge-graph-data agents, Koda support, Windsurf rules.*
