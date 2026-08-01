# Knowledge Graph Project - AI Rules

## Project Overview
Knowledge Graph - система управления заметками с графовыми связями и NLP-анализом.

## Tech Stack
- **Backend:** Go 1.25+, Gin, GORM, PostgreSQL (pgx/v5), Redis (go-redis/v9), MongoDB, gRPC
- **Frontend:** Svelte 5 (runes only), SvelteKit, TypeScript strict, Vitest, Playwright, ky v1.14
- **NLP:** Python FastAPI, sentence-transformers, HuggingFace
- **Infrastructure:** Docker, Docker Compose, nginx gateway

## Key Rules for AI

### 1. Tests are Mandatory
- Always run `go test ./...` after backend changes
- Frontend: `npm run test:unit` for unit tests
- E2E: `npm run test` for Playwright
- Coverage must be >60%

### 1.1 Manual Found → Automated Covered
- If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue.
- Choose the test level by severity:
  - **unit** — pure logic, validators, or utilities (e.g., `errorMessage.ts`, email validation).
  - **integration** — handlers, repositories, or routes (e.g., `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
  - **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g., public graph, achievements, SSE fallback).
- The test should actually fail before the fix (where safe) and pass after the fix.
- If the defect is related to manual data preparation (seed scripts, config files), fix the script or config — not just the instruction.
- If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue.
- Choose the test level by severity:
  - **unit** — pure logic, validators, or utilities (e.g., `errorMessage.ts`, email validation).
  - **integration** — handlers, repositories, or routes (e.g., `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
  - **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g., public graph, achievements, SSE fallback).
- The test should actually fail before the fix (where safe) and pass after the fix.
- If the defect is related to manual data preparation (seed scripts, config files), fix the script or config — not just the instruction.

### 2. Go Conventions
- Clean Architecture: domain/application/infrastructure/interfaces layers
- Private fields in entities + factory functions
- Repository Pattern for data access
- DDD: Entities, Value Objects, Aggregates
- No global variables - explicit dependency injection

### 3. Frontend Conventions (Svelte 5)
- Components: `frontend/src/components/` with Atomic Design (`atoms/` → `molecules/` → `organisms/`)
- Shared logic: `frontend/src/shared/` (`api/`, `stores/`, `services/`, `utils/`, `types/`, `mocks/`, `test-utils/`, `styles/`, `config/`)
- Feature slices: `frontend/src/features/` (`graph-interaction/`, `graph-forms/`)
- State via runes-based `.svelte.ts` modules (no Svelte 4 `writable`/`readable`)
- TypeScript for all types
- SvelteKit for routing
- Import aliases: `$shared/*`, `$components/*`, `$features/*` (legacy `$lib` alias removed)

### 4. Docker & Infrastructure
- Multi-stage builds for optimization
- Volumes for data persistence (postgres_data, redis_data, huggingface_cache)
- Health checks for all services
- Graceful shutdown

### 4.1 Test Stack (Isolated Testing Environment)
- **ALWAYS** use `docker-compose.test.yml` for automated testing
- Test stack is completely isolated from dev/personal environments
- Unique ports: Frontend 3002, Backend 18083, PostgreSQL 15434, Redis 16381, MongoDB 27019, NLP 15002, Graph service 9095
- Separate volumes: `pgdata_test`, `mongodbdata_test` (destroyed after use)
- Test database: `knowledge_test` with test credentials
- SKIP_AUTH enabled for testing
- Use scripts: `.\scripts\testing\start-test.ps1`, `.\scripts\testing\stop-test.ps1`, `.\scripts\testing\seed-test-data.ps1`
- Never run E2E/BDD tests against dev or personal stacks
- Test stack destroyed with `down -v` to remove all test data

### 5. Security
- Never commit secrets (.env, tokens)
- Use environment variables
- Rate limiting for write operations
- CORS configured correctly

### 6. NLP Service
- Lazy load: embedding model loads on first request (not at import)
- Offline-first: HF_HUB_OFFLINE=1 for local-only operation
- Models cached in huggingface_cache volume via HF_HOME
- Health check triggers model load, ensures service readiness
- Use get_embedding_model() and ensure_model_loaded() API
- Uvicorn starts in ~1s, model loads in ~15s from cache

### 7. Configuration
- knowledge-graph.config.json - main configuration
- .env - secrets (do not commit)
- ENV variables override JSON config

## Common Commands

### Backend
```bash
cd backend
go test ./... -v                    # Unit tests
go build ./cmd/server               # Build server
go run ./cmd/server                 # Run server
```

### Frontend
```bash
cd frontend
npm run test:unit                   # Unit tests
npm run test                         # E2E tests
npm run build                         # Build
```

### Docker
```bash
docker compose up -d                 # Start all services
docker compose -f docker-compose.personal.yml up -d  # Personal instance
docker compose logs -f backend         # Backend logs

# Test Stack (for automated testing)
.\scripts\testing\start-test.ps1             # Start isolated test stack
.\scripts\testing\stop-test.ps1              # Destroy test stack (removes all data)
.\scripts\testing\seed-test-data.ps1         # Populate test database
docker compose -f docker-compose.test.yml up -d --build  # Manual start
docker compose -f docker-compose.test.yml down -v        # Manual cleanup
```

## Project Structure
```
backend/
├── cmd/                 # Entry points
├── internal/
│   ├── domain/         # Business logic (entities, value objects)
│   ├── application/    # Use cases
│   ├── infrastructure/  # DB, Redis, external services
│   └── interfaces/     # HTTP handlers, middleware
frontend/
├── src/
│   ├── shared/          # API clients, stores, services, utils, types, mocks, test-utils, styles, config
│   ├── components/      # UI components (atoms / molecules / organisms)
│   ├── features/        # Feature slices (graph-interaction, graph-forms)
│   └── routes/          # SvelteKit pages
nlp-service/            # Python FastAPI NLP
services/
docker-compose.yml      # Main stack
docker-compose.personal.yml  # Personal instance
docker-compose.test.yml      # Test stack (isolated)
scripts/
├── start-test.ps1           # Start test stack
├── stop-test.ps1            # Destroy test stack
└── seed-test-data.ps1       # Seed test data
```

## AI-Specific Hints

### 🌐 Language Policy

**User-facing runtime content uses Russian by default (`ru` locale), with English (`en`) support through i18n keys:**

- UI strings (buttons, labels, placeholders, errors)
- Toast messages and tooltips (GalacticLexicon)
- Note titles and content

**MUST be in English:**
- Code identifiers and variable names
- Commit messages
- API error codes

**User content (note titles/bodies) may be in any language.**
**Documentation:** Authoritative docs are in English; Russian translations may be added alongside.
**Exceptions:** Internal code comments (any language OK).

```typescript
// ✅ Good — Russian UI string from galactic-lexicon / i18n
toast.success(formatMessage("note.created", "ru", { title: note.title }));
placeholder="Поиск по графу знаний...";

// ✅ Good — English variant is available through the same i18n key
const en = formatMessage("note.created", "en", { title: note.title });

// ❌ Bad — hardcoded string without i18n key
toast.success("Note created successfully");
```

## AI-Specific Hints

For backend tasks:
- Use Clean Architecture
- Repositories return domain entities
- Use case layers in application
- Don't mix layers

For frontend tasks:
- Use Svelte 5 runes (`$state`, `$derived`, `$effect`, `$props`) — no Svelte 4 stores
- Keep business logic out of components; use `$shared/services/` or `$features/*`
- Components receive data via props; global state via `$shared/stores/`
- Use ky-based API clients from `$shared/api/*`

For infrastructure:
- Docker multi-stage builds for optimization
- Health checks are mandatory
- Use volumes for persistence