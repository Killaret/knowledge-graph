# Agents in Knowledge Graph

**Updated:** July 2026
**Status:** See [AGENTS_EN.md](AGENTS_EN.md) for full documentation (11 agents)

---

## Operations & Roadmap

- **[Regression Testing Plan Summary](REGRESSION_TEST_PLAN_SUMMARY.md)** — 25-step isolated regression cycle
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

The major Clean Architecture violations listed in previous audits have been resolved. The current status and any remaining minor debt are tracked in the **Clean Architecture Refactor Notes** section below.

Key remaining items:
- `internal/application/cache/graph_cache.go` uses `cache.CacheClient`, but consider whether graph-cache orchestration belongs in `application` or a specialized service.
- `cmd/worker/main.go` still constructs an asynq server directly; consider wrapping it in the `internal/infrastructure/queue` package.
- Some backend unit tests still import concrete infrastructure (`*redis.Client`, `postgres` repositories) and should use ports or test doubles.
- Frontend still has hardcoded user-facing strings and `any` types in several components/pages that need i18n / strict typing.

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

// ✅ Good — English variant is available through the same i18n key
const en = formatMessage("note.created", "en", { title: note.title });

// ❌ Bad — hardcoded string without i18n key
toast.success("Note created successfully");
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
- `internal/interfaces/api/notehandler/note_handler.go` — now depends on `recommendation.Repository` and `recommendation.EmbeddingRepository` ports, plus `cache.CacheClient`; no direct `*postgres.RecommendationRepository`, `*postgres.EmbeddingRepository`, or `*redis.Client`.
- `internal/interfaces/api/middleware/permissions.go` — now depends on `permission.Repository` port implemented in `postgres.PermissionRepository`; no `*gorm.DB`.
- `internal/application/cache/graph_cache.go`, `internal/application/queries/graph/get_suggestions.go`, `internal/application/user/settings_service.go`, and `internal/application/achievement/service.go` — now depend on `cache.CacheClient` port implemented in `infrastructure/cache.RedisCacheClient`; no direct `*redis.Client`.
- `internal/infrastructure/db/postgres/note_repo.go` — now depends on the domain `cache.CacheClient` port for the `FindAll` cache; no direct `*redis.Client`.
- `internal/application/common/task_queue.go` — `TaskQueue` port no longer imports `asynq`; typed enqueue methods hide the concrete task queue implementation.
- `internal/application/achievement/service.go` — no longer imports `asynq`; uses the `common.TaskQueue` port's `EnqueueNotification` method.
- `internal/interfaces/api/handlers/backup/handler.go`, `internal/interfaces/api/notehandler/note_handler.go`, and `internal/interfaces/api/linkhandler/link_handler.go` — no longer import `internal/infrastructure/queue/tasks`; task enqueueing goes through the `common.TaskQueue` port.
- `internal/infrastructure/db/postgres/user_repo.go` — role lookup extracted to `domain/user.RoleRepository` port implemented in `postgres.RoleRepository`; no direct `UserRoleModel` queries from `UserRepository`.
- New domain packages:
  - `internal/domain/user` — `User`, `APIKey` aggregates and repository ports.
  - `internal/domain/share` — `NoteShare`, `ShareLink` aggregates and repository port.
- New infrastructure implementations:
  - `internal/infrastructure/db/postgres/user_repo.go`
  - `internal/infrastructure/db/postgres/apikey_repo.go`
  - `internal/infrastructure/db/postgres/refresh_token_repo.go`
  - `internal/infrastructure/db/postgres/share_repo.go`
- `internal/auth/interfaces.go` — defines `TokenStore` and `RefreshTokenRepository` ports.
- `cmd/worker/main.go` — no longer imports asynq; server creation and ServeMux wiring live in `internal/infrastructure/queue`.
- `internal/interfaces/api/middleware/apikey.go` — now depends on `domain/user.APIKeyRepository`; no `*gorm.DB`, no local `APIKeyModel`.
- `internal/interfaces/api/middleware/jwt.go` and `internal/interfaces/api/middleware/permissions.go` — now depend on `auth.TokenStore` interface; no `*auth.RedisTokenStore`.
- `cmd/server/health.go` — now depends on small `DBPinger`, `RedisPinger`, and `NLPHealthChecker` interfaces; no `*gorm.DB` or `*redis.Client`.
- `cmd/server/router.go` — no longer receives `*gorm.DB` or `*redis.Client`; the health handler is injected as `gin.HandlerFunc`.
- `cmd/cli/main.go` — enqueues recommendation tasks through the `common.TaskQueue` port (`queue.NewAsynqClient`) instead of using `asynq` directly.
- `internal/application/draft/service.go` — no longer holds a concrete `*http.Client`.
- `internal/domain/graph/traversal_integration_test.go` moved to `internal/application/graph`; domain tests no longer import infrastructure packages.
- `frontend/src/shared/stores/achievements.svelte.ts` — converted from Svelte 4 `writable` store to Svelte 5 `$state` runes.
- `frontend/src/shared/api/graph.ts` — graph loading error messages now use `formatMessage` i18n keys instead of hardcoded Russian strings.
- `frontend/src/components/organisms/LoginForm.svelte` and `RegisterForm.svelte` — UI strings now use `formatMessage` i18n keys from `shared/utils/i18n.ts`.
- `backend/internal/infrastructure/db/postgres/embedding_repo_test.go` — uses `testutil.SetupTestVectorDB` (pgvector container) instead of a hardcoded `localhost:5432` DSN.

### Current backend coverage

- `go test ./...` passes.
- `go vet ./...` passes.
- Aggregated backend coverage: **60.6%** (target 70%; enforced minimum 60%).

### Frontend coverage snapshot

- Frontend unit tests pass: 580 passed, 37 skipped.
- Statements: **63.63%**, Branches: **78.74%**, Functions: **56.91%**, Lines: **63.63%**.
- Biggest gaps: `features/graph-interaction` (~28%), `features/graph-forms` (~22%), `features/graph-canvas` (~41%), `shared/stores` (~43%).

### Remaining debt

- `internal/application/cache/graph_cache.go` uses `cache.CacheClient`, but the `GraphCache` service itself is application-layer; consider whether graph-cache orchestration belongs in `application` or a specialized service.
- Some backend unit/integration tests still import concrete infrastructure (`*redis.Client`, `postgres` repositories) and should use ports or test doubles.
- Frontend still has hardcoded user-facing strings and `any` types in several components/pages that need i18n / strict typing.
- Frontend coverage and E2E stack not covered by these notes.

### Verification checklist

Before committing backend changes:

1. `cd backend && go test ./...`
2. `cd backend && go vet ./...`
3. Remove generated artifacts: `coverage.out`, `*.cov`, `*.tmp`, `*.log`, `frontend/coverage`, `backend/.coverage_tmp`.
4. Update this document if architecture boundaries change.
