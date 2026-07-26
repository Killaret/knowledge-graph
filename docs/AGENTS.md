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
- `internal/application/cache/graph_cache.go`: confirmed as an application-layer graph cache orchestration service that depends only on the `domain/cache.CacheClient` port. It stays in `application/cache` and the placement debt is closed.

### Remaining architecture debt to address in future iterations

The major Clean Architecture violations listed in previous audits have been resolved. The current status and any remaining minor debt are tracked in the **Clean Architecture Refactor Notes** section below.

Key remaining items:
- Continue i18n coverage for user-facing strings (active pages and new components are covered; `WeltallProtocol.svelte` body and table headers remain hardcoded English placeholders).

### Frontend architecture notes

- `frontend/src/routes/auth/login/+page.svelte` — `AuthCard` title now uses the `app.title` i18n key instead of the hardcoded "Knowledge Graph"; the card now also renders `PreloadIndicator` below `LoginForm`.
- `frontend/src/components/organisms/ProfileEditor.svelte` — heading and email label now use `profile.editTitle` and `profile.emailLabel` i18n keys.
- `frontend/src/components/organisms/AuthCard.svelte` — the `GalaxyIcon` logo is now a focusable button that opens `WeltallProtocol`; `WeltallProtocol` uses a bindable `show` prop and i18n keys for `close` and `protocol.message`.
- `frontend/src/components/organisms/PreloadIndicator.svelte` — new component that polls `PreloadService` and shows `preload.loading` / `preload.ready` i18n messages.
- `frontend/src/shared/services/PreloadService.ts` — exported `isPreloadingData` and `getStats` helpers for reactive UI consumption.
- Production frontend code is now free of explicit `any` type annotations in `src/` (excluding `.test.ts`/`.spec.ts` and mocks).
- Auth state was split into `src/shared/stores/auth-session.svelte.ts` (low-level reactive state, no API client imports) and `src/shared/stores/auth.svelte.ts` (auth flows). This removes the previous circular dependency between the API client and the auth store.
- Madge is configured for circular-dependency detection via `npm run check:circular` (`frontend/scripts/check-circular.mjs` + `frontend/tsconfig.madge.json`).

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
curl http://localhost:18080/health           # Dev nginx gateway
curl http://localhost:9000/health           # Backend
curl http://localhost:9091/health           # Graph service
curl http://localhost:18080/api/v1/notes     # Notes API
```

### Personal Stack (docker-compose.personal.yml)

```bash
curl http://localhost:18085/health           # Personal backend
curl http://localhost:18082/health           # Personal API gateway
curl http://localhost:18084/health           # Personal frontend gateway
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
- `internal/interfaces/api/handlers/auth/handler.go` — now depends on `domainuser.Repository`, `auth.RefreshTokenRepository`, `auth.TokenStore`, `auth.EmailSender`, and `auth.OAuthProvider` ports; no direct `infrastructure/email`, `infrastructure/oauth`, or `postgres` usage.
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
- `internal/auth/interfaces.go` — added `EmailSender` and `OAuthProvider`/`OAuthUserInfo`/`OAuthProviderFactory` ports; `internal/infrastructure/email` and `internal/infrastructure/oauth` now implement these auth-level ports.
- `internal/interfaces/api/handlers/auth/handler_unit_test.go` — no longer imports `internal/infrastructure/oauth`; uses the `auth.OAuthProvider` port with a test double.
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
- `internal/application/graph/traversal_integration_test.go` moved to `internal/tests/integration/graph`; domain tests no longer import infrastructure packages.
- `frontend/src/shared/stores/achievements.svelte.ts` — converted from Svelte 4 `writable` store to Svelte 5 `$state` runes.
- `frontend/src/shared/api/graph.ts` — graph loading error messages now use `formatMessage` i18n keys instead of hardcoded Russian strings.
- `frontend/src/components/organisms/LoginForm.svelte` and `RegisterForm.svelte` — UI strings now use `formatMessage` i18n keys from `shared/utils/i18n.ts`.
- `backend/internal/infrastructure/db/postgres/embedding_repo_test.go` — uses `testutil.SetupTestVectorDB` (pgvector container) instead of a hardcoded `localhost:5432` DSN.
- `internal/domain/cache/cachetest/fake.go` — new in-memory `cache.CacheClient` test double.
- Application-layer unit tests (`achievement`, `cache`, `queries/graph`, `user`) no longer import `internal/infrastructure/cache` or spin up Redis/miniredis.
- `internal/interfaces/api/handlers/{user,auth,share}/handler_test.go` are now tagged `//go:build integration` because they spin up real `postgres` repositories via testcontainers.
- `frontend/src/routes/+page.svelte` no longer uses `any` types for graph data transformation; `RawNode`/`RawLink` interfaces and `normalizeNode`/`normalizeLink` helpers are used.
- `frontend/src/routes/+page.svelte` — filter labels, sort options, and the generic load-error message now use `formatMessage` / i18n keys.
- `frontend/src/components/organisms/CreateNoteModal.svelte` and `EditNoteModal.svelte` — all modal labels, placeholders, buttons, and error messages are now i18n keys (with standard + galactic variants).
- `frontend/src/routes/+page.svelte` — alerts, empty states, selection controls, bulk actions, sort label, confirm modal, and undo toast are now i18n keys; no `any` types remain.
- `frontend/src/components/molecules/SearchBar.svelte` — placeholder, aria-label, and button text now use i18n keys.
- `frontend/src/components/organisms/FloatingControls.svelte` — all titles, aria-labels, placeholders, and menu items now use i18n keys.
- `frontend/src/components/molecules/NoteCard.svelte` — tooltip and card labels (links, edit/delete, indicators, dates) now use i18n keys.
- `frontend/src/entities/celestial-body.ts` — `CelestialBody.label` now resolves through i18n keys (`celestialBody.type.*`), removing hardcoded English user-facing labels.
- `frontend/src/components/organisms/NoteSidePanel.svelte` — panel labels, dates, links section, and delete-links modal now use i18n keys.
- `frontend/src/entities/link-type.ts` — `LinkType.label` now resolves through i18n keys (`linkType.*`).
- `frontend/src/components/organisms/LinkCreator.svelte` and `LinkTooltip.svelte` — link creation form and tooltip labels now use i18n keys.
- `frontend/src/components/organisms/ConfirmModal.svelte` — default/galactic confirm/cancel/title labels now use i18n keys.
- `frontend/src/components/molecules/TypeSelector.svelte` — aria-label now uses i18n key.
- `frontend/src/components/atoms/ToastNotification.svelte` and `ApiErrorDisplay.svelte` — close/error labels now use i18n keys.
- `frontend/src/components/organisms/Sidebar.svelte` — placeholder text now uses i18n keys.
- `frontend/src/entities/graph-mode.ts` — mode labels (`label`/`focusLabel`) now resolve through i18n keys.
- `frontend/src/features/graph-ui/controls.svelte`, `modals.svelte`, and `overlay.svelte` — button titles, modal titles/placeholders, search placeholder, hotkey tooltip, undo toast, and graph stats now use i18n keys.
- `frontend/src/features/graph-canvas/canvas-state.svelte.ts` — hotkey help list and duplicate-link warning now use i18n keys.
- `frontend/src/components/molecules/GraphTooltip.svelte` — node type and link type labels now resolve through `CelestialBody`/`LinkType` i18n keys; tooltip weight label uses an i18n key.
- `frontend/src/routes/+layout.svelte` — skip-auth badge title now uses i18n keys.
- `frontend/src/routes/+page.svelte` — loading and retry labels now use i18n keys.
- `frontend/src/routes/auth/{login,register,forgot-password,reset-password,yandex/callback}/+page.svelte` — page titles, subtitles, error messages, and button labels now use i18n keys.
- `frontend/src/routes/graph/{+page,[id],3d,3d/[id]}/+page.svelte` — titles, hints, loading/error states, toggle labels, and 3D frozen messages now use i18n keys.
- `frontend/src/routes/notes/{new,[id],[id]/edit}/+page.svelte` — form labels, placeholders, buttons, error messages, and note detail labels now use i18n keys.
- `frontend/src/routes/search/+page.svelte` — search title, placeholder, loading/error/empty states, pagination, and pluralized result counts now use i18n keys.
- `frontend/src/routes/profile/+page.svelte` — profile title, subtitle, and loading state now use i18n keys.

### Current backend coverage

- `go test ./...` passes.
- `go vet ./...` passes.
- Aggregated backend coverage: **71.3%** (target 70%; enforced minimum 60%).

### Frontend coverage snapshot

- Frontend unit + coverage tests pass: 803 passed, 34 skipped.
- Statements: **85.51%**, Branches: **85.72%**, Functions: **88.84%**, Lines: **85.51%**.
- Biggest gaps: `shared/utils/deviceCapabilities.ts` (29.29% stmts), `components/organisms/GraphCanvas/delta.ts` (51.16% stmts), `components/organisms/GraphCanvas.svelte` (18.19% funcs), `components/atoms/SpaceBackground.svelte` (0% stmts).

### Full regression test results (latest run)

Run: 2026-07-26. Test stack was started in isolation (dev/personal stopped). E2E/BDD was run in two phases: a clean `SKIP_AUTH=true` stack for skip-auth tests, then a clean `SKIP_AUTH=false` stack for `@auth-real` tests. On Windows, Playwright/Node resolves `localhost` to `::1` while Docker binds published ports to `127.0.0.1`, so tests used `FRONTEND_URL=http://127.0.0.1:<port>` and `BACKEND_URL=http://127.0.0.1:18083`. On some Windows hosts port `3002` is inside a Hyper-V excluded range; use `FRONTEND_PORT` (e.g. `50070`) when starting the test stack and matching `FRONTEND_URL`.

| Layer | Command | Result |
|-------|---------|--------|
| Backend unit | `go test -p=1 -count=1 ./...` | **PASS** |
| Backend integration | `go test -tags=integration -p=1 -count=1 ./...` | **PASS** |
| Frontend unit | `npm run test:unit` | **837 passed, 0 failed** |
| E2E skip-auth (clean stack) | `FRONTEND_URL=http://127.0.0.1:50070 BACKEND_URL=http://127.0.0.1:18083 SKIP_AUTH=true npx playwright test --project=chromium-skip-auth` | **74 passed, 10 skipped, 0 failed** |
| BDD skip-auth (clean stack) | `FRONTEND_URL=http://127.0.0.1:50070 BACKEND_URL=http://127.0.0.1:18083 SKIP_AUTH=true node scripts/run-bdd.cjs` | **5 scenarios, 43 steps passed** |
| E2E real auth (clean stack) | `FRONTEND_URL=http://127.0.0.1:50070 BACKEND_URL=http://127.0.0.1:18083 SKIP_AUTH=false npx playwright test --project=chromium-real-auth` | **12 passed, 0 failed** |
| BDD real auth (clean stack) | `FRONTEND_URL=http://127.0.0.1:50070 BACKEND_URL=http://127.0.0.1:18083 SKIP_AUTH=false node scripts/run-bdd.cjs` | **9 scenarios, 65 steps passed** |

### Environment isolation findings

- **Do not run dev/personal/test stacks simultaneously.** Concurrent stacks cause Docker instability, port/resource conflicts, and Windows `localhost` → `::1` Playwright connection failures.
- **Use only the test stack during E2E/BDD/regression.** Stop dev and personal stacks first.
- **`.env` must be present** with `JWT_SECRET`, `POSTGRES_PASSWORD`, and `PERSONAL_POSTGRES_PASSWORD` matching the existing `postgres_data` / `pgdata_personal` volumes; otherwise dev/personal backend fails to connect.
- **For real-auth tests the test stack must be started with `SKIP_AUTH=false`.** Backend, graph-service, worker and frontend all read this flag; if it is `true`, `/api/v1/users/me` returns 500 and `/api/v1/auth/refresh` returns 400.
- **If `127.0.0.1:3002` is blocked by Hyper-V**, start the stack with `$env:FRONTEND_PORT="50070"` and point Playwright at `FRONTEND_URL=http://127.0.0.1:50070`.
- **Seed test data with `SKIP_AUTH=false` for real-auth runs.** The `seed-test-data.ps1` script authenticates as `testuser`; if the backend is in skip-auth mode, notes are created with the anonymous `00000000-0000-0000-0000-000000000000` owner and real-auth graph requests return an empty graph.

### Remaining debt

- BDD real-auth mode is implemented in `frontend/tests/features/login.feature` and `auth.steps.ts`; run it with `npm run test:bdd:realauth` against a `SKIP_AUTH=false` stack.
- Full `run-full-test-cycle.ps1` 25-step script has manual verification steps (public graph, CI/CD, docs) that are not automated.

### Verification checklist

Before committing backend changes:

1. `cd backend && go test ./...`
2. `cd backend && go vet ./...`
3. Remove generated artifacts: `coverage.out`, `*.cov`, `*.tmp`, `*.log`, `frontend/coverage`, `backend/.coverage_tmp`.
4. Update this document if architecture boundaries change.
