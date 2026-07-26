# Knowledge Graph Project - Devin Skill

## Project Overview
Knowledge Graph — note management system with graph relationships and NLP analysis.
Stack: Go 1.25, Svelte 5, Python FastAPI, PostgreSQL, Redis, MongoDB, Docker.

## Quick Start (PRIORITY)
**Read:** `AI_RULES.md` (single source of truth)
**Agents:** See `docs/AGENTS_EN.md` for all 11 agent descriptions

## Tech Stack (Actual Versions)

### Backend (Go 1.25)
- Framework: Gin v1.12
- ORM: GORM v1.25 + pgx/v5
- Redis: go-redis/v9 (use ConnMaxIdleTime/ConnMaxLifetime — NOT MaxConnAge/IdleTimeout)
- MongoDB: go.mongodb.org/mongo-driver v1.17 (audit logs, drafts)
- Task queue: asynq v0.23
- Auth: golang-jwt/jwt/v5
- Vector search: pgvector-go v0.2
- Testing: testify v1.11, testcontainers-go v0.40, miniredis v2.38

### Frontend (Svelte 5)
- Framework: SvelteKit + @sveltejs/adapter-node
- Language: TypeScript strict mode
- HTTP: ky v1.14
- Testing: Vitest v3, Playwright v1.59, @testing-library/svelte v5
- Visualization: D3-force v3, Three.js v0.184
- BDD: @cucumber/cucumber v12

### NLP Service (Python)
- Framework: FastAPI + uvicorn
- Models: sentence-transformers (HuggingFace)
- Lazy loading: ensure_model_loaded() / get_embedding_model() from nlp_utils
- Offline: HF_HUB_OFFLINE=1, HF_HOME=/path/to/cache
- Startup: Uvicorn ~1s, model load ~15s from cache

### Infrastructure
- Docker multi-stage builds
- Dev stack: docker-compose.yml (nginx:8080, backend:9000, graph-service:9091)
- Personal stack: docker-compose.personal.yml (backend:8085, gateway:8082, graph-service:8092)
- **Test stack: docker-compose.test.yml (frontend:3002, backend:18083, postgres:15434, redis:16381, mongo:27019, nlp:15002, graph-service:19091/19090)**
- All services must have /health endpoint

## Key Rules

### Testing (MANDATORY)
- Always run `go test ./...` after backend changes
- Always run `npm run test:unit` after frontend changes
- Coverage MUST be >60%
- **ALWAYS use test stack for E2E/BDD testing:** `.\scripts\testing\start-test.ps1`
- **Stop dev and personal stacks before E2E/BDD/regression** — running all stacks simultaneously causes Docker instability, port/resource conflicts, and Windows `localhost` → `::1` Playwright failures
- On Windows, prefer `http://127.0.0.1:3002` / `http://127.0.0.1:18083` or rebuild the test frontend with `VITE_API_URL=http://127.0.0.1:18083`
- Never run E2E/BDD tests against dev or personal stacks
- Use table-driven tests in Go with testify/require (not assert)

### Regression Testing (MANDATORY before production deployment)
- **ALWAYS run full regression cycle before production deployment**
- Run: `.\scripts\testing\run-full-test-cycle.ps1` (Windows) or `./scripts/testing/run-full-test-cycle.sh` (Linux/Mac)
- **Isolated Testing Model:** Dev and personal stacks are stopped during testing; only test stack runs
- Ensure `.env` contains `JWT_SECRET` and DB passwords matching existing `postgres_data` / `pgdata_personal` volumes
- Regression plan: 24-step comprehensive testing in `docs/REGRESSION_TEST_PLAN.md`
- Steps: State snapshot, stack isolation, test execution, state comparison, auto-commit
- Frequency: Full (release), Quick (PR), Smoke (daily), Identity (manual testing)
- Stacks identity check: `.\scripts\ci\check-stacks-identity.ps1` (verifies dev/personal/test consistency)
- Stack health check: `.\scripts\ci\check-stacks-health.ps1 -Stack <dev|personal|test|all>`
- Exit criteria: Dev state unchanged, dev/personal identical, all stacks healthy, all tests pass
- Auto-commit: Only if all checks pass (dev unchanged, dev/personal identical, stacks healthy)
- See `docs/REGRESSION_TEST_PLAN.md` for complete regression testing procedures

### Architecture
- **Backend:** Clean Architecture layers: domain → application → infrastructure → interfaces/api
- **Frontend:** Atomic design (`frontend/src/components/`) + shared layer (`frontend/src/shared/`) + feature slices (`frontend/src/features/`)
- **State:** Svelte 5 runes ONLY ($state, `$derived`, $effect, $props) — no Svelte 4 writable/readable
- **DI:** No global variables — explicit dependency injection everywhere
- **NLP:** Lazy loading — model loads on first /health request, not at import

### Security
- Never commit secrets (.env, tokens, OAuth keys)
- Use environment variables for all secrets
- Rate limiting required for write operations (POST/PUT/DELETE)
- CORS configured correctly
- JWT validation in middleware, not in handlers

### Language Policy (RU UI via i18n, English for code/API/commits/public docs)
**User-facing runtime content uses Russian by default (`ru` locale), with English (`en`) support through i18n keys:**
- UI strings (buttons, labels, placeholders, error messages)
- Toast messages and tooltips (GalacticLexicon)
- Note titles and content (where the app renders labels; user-created content may be in any language)

**MUST be in English:**
- Code identifiers and variable names
- API error codes
- Commit messages
- Public code comments and authoritative documentation

**User content (note titles/bodies) may be in any language.**
**Exceptions:** Internal code comments (brief explanations in any language OK)

```typescript
// ✅ Good — Russian UI string from galactic-lexicon / i18n
toast.success(formatMessage("note.created", "ru", { title: note.title }));
placeholder="Поиск по графу знаний...";

// ✅ Good — English variant is available through the same i18n key
const en = formatMessage("note.created", "en", { title: note.title });

// ❌ Bad — hardcoded string without i18n key
toast.success("Note created successfully");
```

## Common Commands

```bash
# Backend
cd backend && go test ./...                           # Unit tests
cd backend && go test -tags=integration ./...         # Integration tests
cd backend && go test -coverprofile=coverage.out ./... # With coverage
cd backend && go build ./cmd/server                   # Build server

# Frontend
cd frontend && npm run test:unit                      # Vitest unit tests
cd frontend && npx playwright test                    # Playwright E2E tests
cd frontend && npx playwright test --grep="@visual"   # Visual regression tests
cd frontend && npm run test:bdd                       # Cucumber BDD tests
cd frontend && npm run build                          # Build frontend

# NLP Service
cd nlp-service && pytest tests/ -v                    # Unit tests
curl http://localhost:5000/health                    # Health check

# Test Stack Management
.\scripts\testing\start-test.ps1                             # Start test stack
.\scripts\testing\seed-test-data.ps1                          # Seed test data
.\scripts\testing\stop-test.ps1                              # Destroy test stack
docker compose -f docker-compose.test.yml up -d --build  # Manual start
docker compose -f docker-compose.test.yml down -v            # Manual cleanup

# Regression Testing (Isolated Model)
.\scripts\testing\run-full-test-cycle.ps1                    # Full regression cycle (24 steps)
.\scripts\ci\check-stacks-identity.ps1                   # Stacks identity check
.\scripts\ci\check-stacks-health.ps1 -Stack dev          # Check dev stack health
.\scripts\ci\check-stacks-health.ps1 -Stack personal     # Check personal stack health
.\scripts\ci\check-stacks-health.ps1 -Stack test         # Check test stack health

# Health Checks
curl http://localhost:8080/health                    # Dev stack nginx
curl http://localhost:9000/health                    # Dev stack backend
curl http://localhost:8082/health                    # Personal stack
curl http://localhost:18083/health                   # Test stack backend
curl http://localhost:3002                           # Test stack frontend
```

## Docker

```bash
# Dev stack
docker compose up -d                                 # Start dev stack
docker compose -f docker-compose.personal.yml up -d  # Start personal stack
docker compose logs -f backend                       # Backend logs

# Test Stack (for E2E/BDD testing - Isolated Model)
.\scripts\testing\run-full-test-cycle.ps1                    # Full test cycle (isolated model)
.\scripts\testing\start-test.ps1                            # Start isolated test stack
.\scripts\testing\stop-test.ps1                             # Destroy test stack (removes all data)
.\scripts\testing\seed-test-data.ps1                        # Seed test database
.\scripts\ci\check-stacks-health.ps1 -Stack test         # Check test stack health
.\scripts\ci\check-stacks-identity.ps1                 # Check stacks identity (dev/personal/test)
docker compose -f docker-compose.test.yml up -d --build  # Manual start
docker compose -f docker-compose.test.yml down -v            # Manual cleanup
```

## File Structure

```
backend/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── domain/                 # Entities, Value Objects, Aggregates
│   ├── application/            # Use cases
│   ├── infrastructure/         # DB, Redis, MongoDB, external services
│   └── interfaces/api/         # Gin handlers, middleware
frontend/src/
├── shared/                     # API clients, stores, services, utils, types, mocks, test-utils, styles, config
├── components/                 # atoms/ molecules/ organisms/
├── features/                   # graph-interaction/, graph-forms/
└── routes/                     # SvelteKit pages
nlp-service/
├── app/main.py                 # FastAPI app
├── app/nlp_utils.py            # ensure_model_loaded, get_embedding_model
└── app/models.py               # Pydantic request/response models
```

## AI Agents (11 total)

| Agent | Domain |
|-------|--------|
| knowledge-graph-orchestrator | Task routing & delegation |
| knowledge-graph-backend-go | Go API, PostgreSQL, Redis, MongoDB, JWT |
| knowledge-graph-frontend-svelte | Svelte 5, TypeScript, UI/UX |
| knowledge-graph-integration | OpenAPI, DTOs, API contracts |
| knowledge-graph-infrastructure | Docker, monitoring, backups |
| knowledge-graph-devops | CI/CD, deployment |
| knowledge-graph-performance | Profiling, caching, P95 |
| knowledge-graph-security | Audit, Auth/AuthZ, encryption |
| knowledge-graph-testing | Unit/integration/E2E tests |
| knowledge-graph-nlp | Python FastAPI, NLP, HuggingFace embeddings |
| knowledge-graph-data | DB migrations, GORM, pgvector, schemas |

**Agent rules:** `.cursor/rules/` (Cursor AI)
**Koda rules:** `.continue/rules/` (Koda VSCode extension — based on Continue.dev)
**Devin skill:** `.devin/skills/knowledge-graph/SKILL.md` (this file)

## Orchestrator Routing Matrix

When a task touches multiple layers, route to the **primary** agent first, then the **secondary** agent.

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
| Logs / rollback | `devops` | — |
| JWT / CORS / rate-limit | `security` | `backend-go` |
| pgvector / embedding query | `data` | `performance` |
| Redis key naming / pooling | `data` | `performance` |
| Go table-driven tests | `testing` | `backend-go` |
| Vitest / Playwright tests | `testing` | `frontend-svelte` |
| NLP FastAPI endpoint | `nlp` | `integration` |
| HuggingFace model loading | `nlp` | `infrastructure` |
| P95 / caching strategy | `performance` | `backend-go` |
| Secret management | `security` | `infrastructure` |
| Yandex OAuth | `security` | `backend-go` |

## Escalation Scenarios

Invoke multiple agents sequentially when a task crosses boundaries:

1. **New feature end-to-end** (e.g., "Add note sharing"):
   - `backend-go` → domain entity + repository + handler
   - `integration` → DTO + OpenAPI annotation
   - `frontend-svelte` → component + API client call
   - `testing` → unit + integration + E2E tests
   - `security` → verify auth guard on new endpoint

2. **Schema migration**:
   - `data` → write raw SQL migration file
   - `backend-go` → update GORM model + repository
   - `testing` → testcontainers integration test

3. **Performance regression**:
   - `performance` → identify hot path
   - `backend-go` → fix query / add index
   - `data` → confirm pgvector index strategy
   - `devops` → verify with `docker compose logs -f backend`

4. **Security audit**:
   - `security` → JWT middleware + CORS + rate-limit review
   - `backend-go` → confirm no global state, proper error propagation
   - `devops` → confirm no secrets in image layers

## Agent Priority Hierarchy

When agents conflict, apply this order:

1. **Security** — never weaken auth, validation, rate limiting, or secret handling for performance or convenience.
2. **Correctness** — domain invariants, data integrity, and API contracts come before optimization.
3. **Performance** — caching, indexing, async offloading, bundle size.
4. **Convenience** — developer experience, helper scripts, optional shortcuts.

**Default resolution:** if `performance-agent` wants a change that `security-agent` rejects, keep the security requirement unless the user explicitly overrides it.

## Cross-Agent Handoff

When delegating from one subagent to another, include a brief handoff block:

```
Handoff to: <agent-name>
Scope: <one-line summary>
Changed files:
- <path>
- <path>
Affected contracts:
- OpenAPI: <endpoint>
- DTO: <struct/path>
- TS type: <path>
Tests added/updated:
- <test path>
Open questions:
- <anything the next agent must decide>
Next step:
- <concrete instruction>
```

The receiving agent MUST read the listed files before acting.

## Workflows

### Backend Tasks
1. Read existing code patterns in `backend/internal/`
2. Follow Clean Architecture layers (no cross-layer imports)
3. Add unit tests with table-driven approach
4. Run `go test ./...` to verify
5. Build with `go build ./cmd/server`

### Frontend Tasks
1. Read existing components in `frontend/src/components/`
2. Follow atomic design (place in atoms/molecules/organisms)
3. Use Svelte 5 runes ($state, $derived, $effect)
4. Add Vitest unit tests
5. Run `npm run test:unit` to verify

### NLP Tasks
1. Check `nlp-service/app/nlp_utils.py` for existing utilities
2. Use lazy loading pattern — never load model at module import
3. Test with `pytest tests/ -v`
4. Verify /health endpoint triggers model load correctly

### Infrastructure Tasks
1. Read existing docker-compose files first
2. Use multi-stage builds (builder + runtime stages)
3. Add /health endpoints to new services
4. Use named volumes for persistence
5. Test with `docker compose up -d`

### Testing Tasks (E2E/BDD)
1. **ALWAYS start test stack:** `.\scripts\testing\start-test.ps1`
2. Seed test data if needed: `.\scripts\testing\seed-test-data.ps1`
3. Run E2E/BDD tests against test stack (ports 3002/18083)
4. **ALWAYS destroy test stack after testing:** `.\scripts\testing\stop-test.ps1`
5. Never test against dev/personal stacks (data contamination risk)

## Error Handling
- Go: Return errors explicitly, never panic in business logic
- Frontend: Handle API errors in stores/services, show user-friendly messages
- NLP: Log exceptions with logger.exception(), return 500 with detail
- Add context to errors: fmt.Errorf("creating note: %w", err)

## Recent Changes
- Go version updated to 1.25 (go.mod)
- Redis: go-redis/v9 API (ConnMaxIdleTime/ConnMaxLifetime, not MaxConnAge/IdleTimeout)
- NLP service: lazy loading via ensure_model_loaded(), HF_HUB_OFFLINE support
- MongoDB added for audit logs and drafts (go.mongodb.org/mongo-driver)
- pgvector added for semantic search (pgvector-go v0.2)
- asynq added for async task processing
- **Test stack added:** docker-compose.test.yml for isolated E2E/BDD testing
- **Test scripts:** start-test.ps1, stop-test.ps1, seed-test-data.ps1
