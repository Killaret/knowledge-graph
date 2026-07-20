# Agents in Knowledge Graph

**Updated:** July 2026
**Status:** See [AGENTS_EN.md](AGENTS_EN.md) for full documentation (11 agents)

---

## Operations & Roadmap

- **[Regression Testing Plan Summary](REGRESSION_TEST_PLAN_SUMMARY.md)** — 24-step isolated regression cycle
- **[Testing Commands & Procedures](TESTING_COMMANDS.md)** — Commands for unit, integration, E2E, BDD, and stack management
- **[UI Modernization Roadmap](UI_MODERNIZATION_ROADMAP.md)** — Canvas, NoteCard, multilingual lexicon, and UX improvements

---

## Quick Reference

The project uses **11 specialized AI agents** defined across multiple AI tools:

| Agent | Focus |
|-------|-------|
| knowledge-graph-orchestrator | Task routing & delegation |
| knowledge-graph-backend-go | Go API, PostgreSQL, Redis, MongoDB, JWT |
| knowledge-graph-frontend-svelte | Svelte 5, TypeScript, UI/UX |
| knowledge-graph-integration | OpenAPI, DTOs, API contracts |
| knowledge-graph-infrastructure | Docker, nginx, monitoring |
| knowledge-graph-devops | CI/CD, deployment |
| knowledge-graph-performance | Profiling, caching, P95 |
| knowledge-graph-security | Auth/AuthZ, audit, encryption |
| knowledge-graph-testing | Unit/integration/E2E/BDD |
| knowledge-graph-nlp | Python FastAPI, NLP, HuggingFace *(NEW)* |
| knowledge-graph-data | DB migrations, pgvector, schemas *(NEW)* |

---

## AI Tool Configuration

| Tool | Rules Location |
|------|---------------|
| Cursor AI | `.cursor/rules/*.md` |
| Koda VSCode (koda-base/koda-pro) | `.continue/rules/*.md` |
| Windsurf/Cascade | `.windsurfrules` |
| Devin | `.devin/skills/knowledge-graph/SKILL.md` |

---

## Backend Architecture Audit Notes

Date: 2026-07-20

### Completed Clean Architecture / DDD cleanups

- `cmd/server/main.go`: extracted `run(...)` with explicit dependency injection; covered with `cmd/server/main_test.go`.
- `internal/interfaces/api/middleware/apikey.go`: introduced `APIKeyRepository` interface so the middleware no longer performs raw GORM queries directly inside `APIKey()`.
- `internal/interfaces/api/handlers/backup/handler.go`: removed the unused `*cloud.YandexBackupService` concrete dependency; handler now only depends on `config.Config` and `common.TaskQueue`.
- `internal/infrastructure/db/postgres/link_repo.go`: moved `ErrDuplicateLink` to `internal/domain/link/errors.go` so the link handler depends on the domain error instead of an infrastructure package.

### Remaining architecture debt to address in future iterations

1. **Handlers with direct `*gorm.DB` usage (biggest Clean Architecture violations):**
   - `internal/interfaces/api/handlers/auth/handler.go` — uses `*gorm.DB` and `postgres.UserModel`/`postgres.RefreshTokenModel`.
   - `internal/interfaces/api/handlers/user/handler.go` — uses `*gorm.DB`, `postgres.UserModel`, and `postgres.APIKeyModel`.
   - `internal/interfaces/api/handlers/share/handler.go` — uses `*gorm.DB` and several `postgres` models.
   - Recommended fix: introduce domain repositories (`user.Repository`, `auth.TokenRepository`, `apikey.Repository`, `share.Repository`) in the `domain` layer and move persistence queries to `infrastructure/db/postgres`.

2. **Middleware with persistence coupling:**
   - `internal/interfaces/api/middleware/permissions.go` still takes `*gorm.DB` and queries role/permissions directly. Extract a `PermissionRepository` interface and inject it from `cmd/server`.

3. **Application-layer Redis/GORM leaks:**
   - `internal/application/cache/graph_cache.go` and `internal/application/user/settings_service.go` import `github.com/redis/go-redis/v9` directly. Replace with a cache port interface (e.g., `CacheClient`) implemented in `infrastructure/cache`.
   - `internal/application/queries/graph/get_suggestions.go` imports `go-redis` for caching. Same cache-port abstraction applies.
   - `internal/application/recommendation/refresh_service.go` imports `gorm` and `go-redis`. It should depend on `note.Repository`, `link.Repository`, and a cache port only.

4. **Handler concrete infrastructure dependencies:**
   - `internal/interfaces/api/notehandler/note_handler.go` depends on `*postgres.RecommendationRepository`, `*postgres.EmbeddingRepository`, and `*redis.Client`. Define handler-local or domain-driven interfaces so persistence and cache can be mocked in tests.
   - `internal/interfaces/api/handlers/backup/handler.go` still creates `*asynq.Task` through `internal/infrastructure/queue/tasks`. Consider moving task payload builders into `application` or `domain` and converting `common.TaskQueue` to operate on an application-level `Task` type, with the `infrastructure/queue` adapter translating to `asynq`.

### Conventions used during audit

- Keep constructors explicit (`func New...`) and avoid global state.
- Domain packages define errors and repository interfaces; infrastructure packages implement them.
- Interface consumers should not know concrete `gorm`, `go-redis`, `mongo`, or `asynq` types.

---

## Full Documentation

- **[AGENTS_EN.md](AGENTS_EN.md)** — Full English documentation, agent descriptions, selection matrix

---

## Service Health Checks

### Dev Stack (docker-compose.yml)

```bash
curl http://localhost:8080/health           # Nginx gateway
curl http://localhost:9000/health           # Backend
curl http://localhost:9091/health           # Graph service
curl http://localhost:8080/api/v1/notes     # Notes API
```

### Personal Stack (docker-compose.personal.yml)

```bash
curl http://localhost:8085/health           # Personal backend
curl http://localhost:8082/health           # Personal API gateway
curl http://localhost:8092/health           # Personal graph service
```

---

## Language Policy

**User-facing runtime content MUST be in English:**
- UI strings (buttons, labels, placeholders, errors)
- Toast messages and tooltips (GalacticLexicon)
- Note titles and content
- Commit messages

**May be bilingual:**
- Documentation — the authoritative English version must be present; Russian translations may be added alongside.

**Exceptions:** Internal code comments (any language OK)

```typescript
// ✅ Good
toast.success("Note created successfully");

// ❌ Bad
toast.success("Заметка создана успешно");
```

---

## Clean Architecture Refactor Notes

This section tracks the ongoing migration from direct `*gorm.DB` usage in handlers to Clean Architecture layers.

### Completed decoupling

- `cmd/server/main.go` — `run()` extracted with dependency injection; wired through constructors.
- `internal/interfaces/api/handlers/user/handler.go` — now depends on `domainuser.Repository` and `domainuser.APIKeyRepository`; no `*gorm.DB` or `postgres` model usage.
- `internal/interfaces/api/handlers/auth/handler.go` — now depends on `domainuser.Repository`, `auth.RefreshTokenRepository`, and `auth.TokenStore`; no direct `postgres` usage.
- `internal/interfaces/api/handlers/share/handler.go` — now depends on `domainnote.Repository`, `domainuser.Repository`, and `domainshare.Repository`; no `*gorm.DB` or `postgres` model usage.
- `internal/application/achievement/engine.go` — now depends on `domain/achievement.Counter` port implemented in `postgres.AchievementCounter`; no `*gorm.DB`.
- `internal/application/recommendation/refresh_service.go` and `affected_notes.go` — now depend on `note.Repository` and `recommendation.Repository` ports; no `*gorm.DB` or direct `postgres` usage.
- New domain packages:
  - `internal/domain/user` — `User`, `APIKey` aggregates and repository ports.
  - `internal/domain/share` — `NoteShare`, `ShareLink` aggregates and repository port.
- New infrastructure implementations:
  - `internal/infrastructure/db/postgres/user_repo.go`
  - `internal/infrastructure/db/postgres/apikey_repo.go`
  - `internal/infrastructure/db/postgres/refresh_token_repo.go`
  - `internal/infrastructure/db/postgres/share_repo.go`
- `internal/auth/interfaces.go` — defines `TokenStore` and `RefreshTokenRepository` ports.

### Current backend coverage

- `go test ./...` passes.
- `go vet ./...` passes.
- Aggregated backend coverage: **61.9%** (target >60%).

### Remaining debt

- `internal/interfaces/api/notehandler/note_handler.go` still depends on concrete `*postgres.RecommendationRepository` and `*postgres.EmbeddingRepository`.
- `internal/interfaces/api/middleware/permissions.go` still queries role/permissions directly via `*gorm.DB`; introduce a `PermissionRepository` port.
- `internal/infrastructure/db/postgres` still contains business logic that should migrate to domain or application services (e.g., role lookup during user creation could live in an application service).
- `internal/interfaces/api/handlers/backup` is decoupled but minimal.
- Frontend coverage and E2E stack not covered by these notes.

### Verification checklist

Before committing backend changes:

1. `cd backend && go test ./...`
2. `cd backend && go vet ./...`
3. Remove generated artifacts: `coverage.out`, `*.cov`, `*.tmp`, `*.log`.
4. Update this document if architecture boundaries change.
