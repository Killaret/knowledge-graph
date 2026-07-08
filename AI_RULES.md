# Knowledge Graph Project - AI Rules

## Project Overview
Knowledge Graph - система управления заметками с графовыми связями и NLP-анализом.

## Tech Stack
- **Backend:** Go 1.23+, PostgreSQL, Redis, gRPC
- **Frontend:** Svelte 5, Vitest, Playwright
- **NLP:** Python FastAPI, sentence-transformers, HuggingFace
- **Infrastructure:** Docker, Docker Compose

## Key Rules for AI

### 1. Tests are Mandatory
- Always run `go test ./...` after backend changes
- Frontend: `npm run test:unit` for unit tests
- E2E: `npm run test` for Playwright
- Coverage must be >60%

### 2. Go Conventions
- Clean Architecture: domain/application/infrastructure/interfaces layers
- Private fields in entities + factory functions
- Repository Pattern for data access
- DDD: Entities, Value Objects, Aggregates
- No global variables - explicit dependency injection

### 3. Frontend Conventions (Svelte 5)
- Components: atoms → molecules → organisms
- Stores for state, stores/lib/services for business logic
- TypeScript for all types
- SvelteKit for routing

### 4. Docker & Infrastructure
- Multi-stage builds for optimization
- Volumes for data persistence (postgres_data, redis_data, huggingface_cache)
- Health checks for all services
- Graceful shutdown

### 4.1 Test Stack (Isolated Testing Environment)
- **ALWAYS** use `docker-compose.test.yml` for automated testing
- Test stack is completely isolated from dev/personal environments
- Unique ports: Frontend 3002, Backend 8083, PostgreSQL 5434, Redis 6381
- Separate volumes: `pgdata_test`, `mongodbdata_test` (destroyed after use)
- Test database: `knowledge_test` with test credentials
- SKIP_AUTH enabled for testing
- Use scripts: `.\scripts\start-test.ps1`, `.\scripts\stop-test.ps1`, `.\scripts\seed-test-data.ps1`
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
.\scripts\start-test.ps1             # Start isolated test stack
.\scripts\stop-test.ps1              # Destroy test stack (removes all data)
.\scripts\seed-test-data.ps1         # Populate test database
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
│   ├── lib/            # Business logic
│   ├── components/     # UI components
│   └── routes/          # Pages
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

### 🌐 Language Policy (ENGLISH ONLY)

**All user-facing content MUST be in English:**

- ✅ **Note titles and content** — all user-created notes
- ✅ **Annotations and descriptions** — any text fields for users
- ✅ **UI strings** — buttons, labels, placeholders, error messages
- ✅ **Toast messages and tooltips** — GalacticLexicon messages
- ✅ **Comments in code** — public API docs, README files
- ✅ **Commit messages** — clear, descriptive English

**Exceptions:**
- Internal code comments — brief explanations in any language OK
- Variable/function names — follow project conventions

```typescript
// ✅ Good
toast.success("Note created successfully");

// ❌ Bad
toast.success("Заметка создана успешно");
```

## AI-Specific Hints

For backend tasks:
- Use Clean Architecture
- Repositories return domain entities
- Use case layers in application
- Don't mix layers

For frontend tasks:
- Components should be reactive
- Stores only for cross-component state
- Components receive data via props/stores

For infrastructure:
- Docker multi-stage builds for optimization
- Health checks are mandatory
- Use volumes for persistence