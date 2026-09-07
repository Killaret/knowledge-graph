# Product Backlog

Detailed plans behind the short [roadmap](../ROADMAP.md): what each item means, what it depends
on, and what is considered done. The roadmap says *where the project is going*; this file says
*what exactly is planned*.

Shipped work is not tracked here — it lives in [CHANGELOG.md](../CHANGELOG.md). Untested
hypotheses live in [IDEAS.md](IDEAS.md).

## 🎯 NOW: Current Focus

### 🔄 Manual Testing & Stabilization

| Task                                  | Status  | Priority    |
| ------------------------------------- | ------- | ----------- |
| Manual testing of all features        | ✅ Done | 🔴 Critical |
| Bug fixes from manual testing         | ✅ Done | 🔴 Critical |
| Critical verifications for production | ✅ Done | 🔴 Critical |

**Subtasks:**

- [x] Frontend E2E smoke tests (`cd frontend && npm run test:smoke`)
- [x] Backend integration tests (`cd backend && go test -tags=integration ./...`)
- [x] CI/CD workflows verification
- [x] NLP API testing
- [x] Backend auth API testing
- [x] Public graph verification

### 🎨 UI/UX: Cosmic Cockpit

| Task                                                                                          | Status  | Priority |
| --------------------------------------------------------------------------------------------- | ------- | -------- |
| Cosmic Cockpit UI — starship cockpit layout with slide-out panels, HUD, and first-person mode | ✅ Done | 🟠 High  |
| 2D node color variation — optional custom color + deterministic palette fallback              | ✅ Done | 🟠 High  |

**Scope:**

- Four slide-out panels (top, bottom, left, right) emulating a starship cockpit dashboard inspired by "Space Rangers".
- Hover with 200–300 ms delay; drag-to-open by panel edge/corner; anchor to pin panels.
- "First-person" (fullscreen) mode button hides all panels, showing only the graph.
- Left panel: cluster navigation, note tree, graph mode switcher, type filters with multi-select and top-down expand animation.
- Right panel: selected note details, link mini-graph, contextual actions (edit, add child note, add link, publish).
- Top panel: search, global filters, 2D/3D toggle, sync status.
- Bottom panel: HUD — current cluster, note/link count, graph health, FPS, delta sync indicator, first-person button.
- UI settings: sensitivity, auto-collapse, reduced motion.
- Graph node context menu for creating child notes.
- Singularity drop zone for archive/delete; Black Hole type is strictly an "unresolved problem" note type and no longer implies deletion.

**MVP:**

- Static layout with four panels.
- Bottom HUD bar.
- Type filters in left panel with multi-select.
- First-person mode.
- Basic slide-out animation.

**Dependencies:** FSD widget refactoring, graph store.

**Note:** Cosmic notifications/toasts are not part of the Cosmic Cockpit MVP; they will be handled by a separate optional notification system.

**Related analysis:**

- [UI Duplication and Note Creation Analysis](UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md) — describes the legacy UI duplication this redesign addresses

### ⚡ Performance & Resource Optimization

| Task                                                             | Status     | Priority |
| ---------------------------------------------------------------- | ---------- | -------- |
| Resource usage audit and optimization (memory, CPU, bundle size) | ⏳ Planned | 🟠 High  |

**Audit findings:**

- **Frontend bundle:** `three` is imported as `import * as THREE from "three"` in 7+ files — full library may end up in the bundle. Bundle analyzer is not configured.
- **Frontend runtime:** `frontend/src/features/home-page/home-page.svelte.ts` uses `setInterval` for delta polling; the graph animation loop is implemented in `frontend/src/entities/graph-canvas/lib/animation.ts` and consumed by `GraphCanvas.svelte` — both need lifecycle cleanup checks.
- **Frontend limits:** `max_nodes: 500` on the client, but `graph_service.full_limit: 1000` on the backend — mismatch may cause out-of-memory or timeouts.
- **Backend search:** `note_repo.go` uses `plainto_tsquery` + `ts_rank` full-text search — indexes and query cost need verification.
- **Backend cache:** `graph_service` caches full layout with 300s TTL; large graphs may pressure Redis memory.
- **Docker dev stack:** no memory limits for backend/frontend services (test stack has 512M–2G limits).
- **Testing:** performance and memory-usage tests are marked as `⏳` in `API_TEST_COVERAGE_PLAN.md`; no CI bundle-size check.

**Action plan:**

1. Add `rollup-plugin-visualizer` or `vite-bundle-analyzer` to `frontend/package.json`.
2. Profile 3D graph load with 500+ nodes and measure FPS against `knowledge-graph.config.json` thresholds.
3. Audit and add `GIN`/`tsvector` indexes for full-text search.
4. Set memory limits for dev stack backend/frontend in `docker-compose.yml`.
5. Add baseline performance smoke tests (large graph load, bundle size budget).

---

## 🧩 Graph Service: Stabilization & Evolution Plan

> **Context:** per ADR 013, the main backend should publish `Note*`/`Link*` events and stop being the primary graph source. `graph-service` owns layout computation, caching, and analytics. The main backend remains a **fallback** when `graph-service` is unavailable.

| Task                                                                     | Status         | Priority    |
| ------------------------------------------------------------------------ | -------------- | ----------- |
| Event-driven invalidation & fallback model                               | ✅ Done        | 🔴 Critical |
| Auth & user-scoped filtering in graph-service                            | ✅ Done        | 🔴 Critical |
| Graph API unification & double-load removal                              | ✅ Done        | 🟠 High     |
| Graph analytics API (Neighbors, Path, Recommendations)                   | ✅ Done        | 🟡 Medium   |
| Materialized view / graph index                                          | ✅ Done        | 🟡 Medium   |
| gRPC-web / SSE full-graph streaming                                      | ⏳ Planned     | 🟡 Medium   |
| Improved layout algorithms (Honeycomb, force-directed, Cosmic Navigator) | 🔄 In Progress | 🟠 High     |

### 1. Event-driven invalidation & fallback model

- `note_handler` and `link_handler` publish `NoteCreated`/`Updated`/`Deleted` and `LinkCreated`/`Updated`/`Deleted` events via `internal/infrastructure/events/publisher.go` to the Redis `graph:events` channel.
- `graph-service` (`internal/subscriber/pubsub.go`) listens and invalidates keys `graph-service:full:*`, `graph-service:note:*`, `graph-service:delta:*`.
- Main backend endpoints `/graph/all`, `/me/graph/fresh`, `/me/graph/cached` become **fallback**: frontend/proxy checks `graph-service` health and switches to backend only on unavailability.
- This is critical because `events.Publisher` is currently not wired to handlers, so `graph-service` cache only expires by TTL.

### 2. Auth & user-scoped filtering in graph-service ✅

- Add JWT middleware to `graph-service` HTTP and gRPC.
- Derive `user_id` from token and remove public `user_id` query parameter.
- Add `creator_id` filter in `services/graph-service/internal/db/postgres_client.go` for all `notes`/`links` queries.
- Anonymous/unauthenticated requests default to `is_public = true` to prevent leaking private notes.
- `frontend/src/hooks.server.ts` proxy must forward `authorization` or a signed `x-internal-auth` header to `graph-service`.
- Public graph moves to a separate endpoint (`/api/v1/graph/public`) filtering `is_public = true`.

### 3. Graph API unification & double-load removal ✅

- Standardize graph-service response fields: `id`, `title`, `type`, `source`, `target`, `weight`, `link_type`.
- Remove fallback normalization (`id/Id/ID`, `source/source_note_id`) in `frontend/src/routes/+page.svelte`.
- `+page.svelte` calls `getGraphWithPreload()` once (which loads `getFullGraphData()` when no cached data) and uses `Promise.all([getNotes(), getGraphWithPreload()])` for authenticated users; the sequential `getFreshGraph()` + repeated `loadGraphData()` flow is removed.
- Use `/api/v1/graph/delta?last_hash=` for incremental updates via `PreloadService`.
- `PreloadService` exposes `seedGraph(graphData)` so any full-graph response is cached with `lastHash` and can drive delta updates.

### 4. Graph analytics API

- Implement gRPC/HTTP methods `GetNeighbors`, `GetPath`, `GetRecommendations` in `graph-service`.
- HTTP endpoints:
  - `GET /api/v1/graph/note/:id/neighbors?depth=`
  - `GET /api/v1/graph/path?from=&to=`
  - `GET /api/v1/graph/recommendations?note_id=&limit=`
- Replace backend BFS (`backend/internal/domain/graph/bfs.go`) and recommendation traversal with calls to `graph-service`.

### 5. Materialized view / graph index

- Create materialized view `note_links_closure` for transitive link closure.
- Refresh the view on `LinkCreated`/`LinkDeleted` events (async via trigger or ASYNQ).
- Use the view for `GetPath`, degree computation, semantic clustering, and Honeycomb layout.

### 6. gRPC-web / SSE full-graph streaming

- For graphs >500 nodes, stream `GetFullLayout` in chunks via gRPC streaming or Server-Sent Events.
- Frontend progressively renders nodes and links.
- Alternative for JSON API: NDJSON streaming.

### 7. Improved layout algorithms

- **Honeycomb View:** radial layout centered at the node with maximum `sum(weight)`, with concentric circles by weight thresholds.
- **Force-directed 2D/3D:** replace/augment current circle/helix with server-side physical simulation.
- **Cosmic Navigator:** hybrid server-side 3D layout (`layout=3d`) plus client-side `d3-force-3d` relaxation.

---

## 🚀 NEXT: Immediate Goals

### 📝 Self-Hosted Deployment Documentation

| Task                                  | Status     | Priority |
| ------------------------------------- | ---------- | -------- |
| Create comprehensive deployment guide | ⏳ Planned | 🟠 High  |

**Scope:**

- Docker deployment instructions
- Environment configuration
- Database setup and migrations
- NLP service setup
- SSL/TLS configuration
- Backup procedures
- Monitoring setup

**Related docs:**

- [Docker Deployment Guide](DOCKER.md)

### 🌐 Public Note Pool (Publish/Unpublish)

| Task                                      | Status  | Priority |
| ----------------------------------------- | ------- | -------- |
| Implement publish/unpublish functionality | ✅ Done | 🟠 High  |

**Scope:**

- Backend API for publish/unpublish
- Frontend UI controls
- Public graph filtering
- Privacy controls
- Public sharing links
- Unpublish workflow

---

## 🔜 SOON: Medium-Term Goals

### Phase 4: Cosmic Navigator (Spaceship) 🔴 Critical

Status: ⏳ Planned
Description: 3D-визуализация графа знаний в виде космоса.

☑ Базовый 3D-рендеринг (Three.js) — разморожен и отрефакторен в features/graph-3d.
□ Orbital / Solar System 3D-режим — узлы распределены по орбитам, удалённым от центральной заметки или центра кластера. Орбиты = смысловые слои. Связи опциональны.
□ Honeycomb Stellaris 3D-режим — гексагональная сетка в 3D с изогнутыми светящимися трубками (гиперпространственные маршруты), как в Stellaris.
□ Динамическое переключение между режимами (2D/3D).
□ Серверная кластеризация (Louvain) для автоматического выделения смысловых слоёв и центров.
□ Кэширование кластеров в Redis, инвалидация по событиям.
□ Knowledge Voyager (Полёт по знаниям) — автопилотный полёт по кластерам по сплайн-траектории (`features/graph-3d/lib/autopilot.ts`), с подсветкой узлов, не открывавшихся N дней. Вау-эффект, требует 3D-режимов и кластеризации.
□ Ghost Notes (Призрачные заметки) — публичные заметки с похожими эмбеддингами появляются как полупрозрачные узлы в пространстве между кластерами (`GET /api/v1/graph/ghost-notes?context=clusterId`); пользователь может «захватить» узел в свой граф. Зависит от Knowledge Voyager и публичных заметок.

### 🔗 Link Improvements

| Task                                     | Status     | Priority  |
| ---------------------------------------- | ---------- | --------- |
| Add tooltips for links                   | ⏳ Planned | 🟡 Medium |
| Improve link animations                  | ⏳ Planned | 🟡 Medium |
| Implement gamma-coding for link strength | ⏳ Planned | 🟡 Medium |
| Auto-link creation based on NLP          | ⏳ Planned | 🟡 Medium |

**Scope:**

- Hover tooltips showing link metadata
- Animated link creation/deletion
- Visual strength indicators (gamma-coding)
- Link type differentiation
- Bidirectional link visualization
- Auto-create links on note creation when recommendation score is above threshold

**Related docs:**

- [Auto Link Creation Plan](AUTO_LINK_CREATION_PLAN.md)

### 🧹 Dust Processor (Quick Notes Handler)

| Task                           | Status     | Priority  |
| ------------------------------ | ---------- | --------- |
| Implement dust note processing | ⏳ Planned | 🟡 Medium |

**Scope:**

- Automatic dust note categorization
- Smart dust note suggestions
- Dust note to regular note conversion
- Batch dust note processing
- Dust note analytics
- Dust inbox panel and dust → note enrichment

**Related docs:**

- [Note Error Correction Plan](NOTE_ERROR_CORRECTION_PLAN.md) (Quick Edit, autocompletion, dust inbox)

### 📝 Quick Edit & Note Autocompletion

| Task                               | Status     | Priority  |
| ---------------------------------- | ---------- | --------- |
| Quick inline note editing (Ctrl+E) | ⏳ Planned | 🟡 Medium |
| Note title/content autocompletion  | ⏳ Planned | 🟡 Medium |
| NLP keyword/embedding correction   | ⏳ Planned | 🟡 Medium |

**Scope:**

- Inline editing of title and content without opening full note page
- Autosave on blur with optimistic update and rollback on error
- Autocomplete suggestions for title, type, tags, and related notes
- User correction of NLP keywords and manual recomputation of embeddings

**Related docs:**

- [Note Error Correction Plan](NOTE_ERROR_CORRECTION_PLAN.md)

---

### Phase 10: Honeycomb View 🟡 Medium

Status: ⏳ Planned
Description: Представление графа в виде сот (гексагональной сетки).

□ 2D Honeycomb — узлы привязаны к ячейкам сетки, связи — прямые линии. Веса связей влияют на близость ячеек.
□ 3D Honeycomb (Stellaris) — плоская гексагональная сетка в 3D-пространстве, связи — изогнутые трубки с неоновым свечением.
□ Переключение в режим сот из тулбара (2D и 3D).

### 🎤 Interactive Onboarding & Breadcrumbs

| Task                                  | Status     | Priority  |
| ------------------------------------- | ---------- | --------- |
| Scenario-based narrator (T1-T6)       | ⏳ Planned | 🟠 High   |
| Breadcrumb progress (scenario + step) | ⏳ Planned | 🟠 High   |
| Spotlight and element highlighting    | ⏳ Planned | 🟡 Medium |
| Tour state, skip, and localStorage    | ⏳ Planned | 🟡 Medium |

**Scope:**

- Guided tours for first graph, first note, first link, search, 3D exploration, public sharing
- Breadcrumbs showing scenario and current step
- Narrator panel with Next/Prev/Skip/Done
- Spotlight overlay highlighting UI elements
- Integration with collapsed cockpit panels
- i18n keys for all tour strings

**Related docs:**

- `.devin/plans/plan-18f8373d959afb70.md` — cockpit redesign plan with onboarding epic

---

## 📋 LATER: Backlog

### 🎓 Coach Marks System

| Task | Status | Priority |
|------|--------|----------|
| Coach Mark Manager (statuses, queue, localStorage) | ⏳ Planned | 🟢 Low |
| CoachMarkOverlay component (spotlight + tooltip) | ⏳ Planned | 🟢 Low |
| Feature definitions: welcome, createNote, createLink, clusters, cosmicCockpit, periodicNotes | ⏳ Planned | 🟢 Low |
| Integration with CosmicCockpitLayout and `?` restart button | ⏳ Planned | 🟢 Low |
| i18n keys and test coverage | ⏳ Planned | 🟢 Low |

**Scope:**
- Contextual hints shown on first use of a feature, not only at app start.
- Per-feature status: `seen`, `dismissed`, `postponed` (with timestamp and `postponeDays`).
- Manager API: `start`, `next`, `skip`, `postpone`, `dismiss`, `finish`, `isSeen`, `isDismissed`, `isPostponed`, `resetFeature`, `resetAll`.
- Queue: only one coach mark active at a time.
- Overlay: dark semi-transparent background, cyan neon border, gradient title, glow highlight.
- Buttons: Next, Done (last step), Skip, Postpone, Dismiss all.
- `localStorage` key: `kg_coach_marks_<variant>`.
- i18n keys: `coach.<featureId>.<stepIndex>.title` and `.text`.
- FSD placement: `entities/coach-mark/` for state/manager, `features/coach-mark/` for overlay and feature definitions.
- Tests: unit for `CoachMarkManager`, component for `CoachMarkOverlay`, E2E for first login and postpone/dismiss/restart.

**Related docs:**
- `docs/COACH_MARKS.md` (to be created)
- [SOON: Interactive Onboarding & Breadcrumbs](#interactive-onboarding--breadcrumbs)

### 📥 Import/Export Tools

| Task                   | Status     | Priority |
| ---------------------- | ---------- | -------- |
| JSON import/export     | ⏳ Planned | 🟢 Low   |
| Markdown import/export | ⏳ Planned | 🟢 Low   |
| CSV import/export      | ⏳ Planned | 🟢 Low   |
| Bookmarklet            | ✅ Done    | 🟠 High  |
| Mass URL import        | ✅ Done    | 🟠 High  |
| Browser extension      | ⏳ Planned | 🟠 High  |

#### Browser capture options

| Option                | Pros                                                            | Cons                                                                                  | Status         |
| --------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------------------------- | -------------- |
| **Bookmarklet**       | No installation, instant MVP                                    | URL length limit (~2000 chars), opens a new tab                                       | ✅ Implemented |
| **Mass URL import**   | Bulk processing, async deduplication, preview editing, progress | Content extraction from URL implemented (`internal/infrastructure/web.ImportFetcher`) | ✅ Implemented |
| **Browser Extension** | Full text access, background capture, notifications             | Requires installation, needs token setup                                              | ⏳ Planned     |

**Scope:**

- Bulk import/export functionality
- Format validation
- Data mapping
- Conflict resolution
- Import history
- Browser capture (bookmarklet, mass URL import, extension)

### 🎮 Gamification (Customization & Points)

| Task                    | Status     | Priority |
| ----------------------- | ---------- | -------- |
| Customization system    | ⏳ Planned | 🟢 Low   |
| Points and achievements | ⏳ Planned | 🟢 Low   |

**Scope:**

- User profile customization
- Point system for activities
- Achievement badges
- Leaderboards
- Progress tracking

### 🔌 Obsidian Integration

| Task            | Status     | Priority |
| --------------- | ---------- | -------- |
| Obsidian import | ⏳ Planned | 🟢 Low   |
| Obsidian sync   | ⏳ Planned | 🟢 Low   |

**Scope:**

- Markdown file import
- Wiki-link conversion
- Tag mapping
- Metadata preservation
- Two-way sync

**Related docs:**

- [Obsidian Import Spec](OBSIDIAN_IMPORT_SPEC.md)

### 📱 PWA Capture (Mobile Quick Notes)

| Task                    | Status     | Priority |
| ----------------------- | ---------- | -------- |
| PWA development         | ⏳ Planned | 🟢 Low   |
| Quick capture interface | ⏳ Planned | 🟢 Low   |

**Scope:**

- Progressive Web App
- Offline support
- Quick note capture
- Mobile-optimized UI
- Push notifications

### 🌐 External API Integrations

| Task                 | Status     | Priority |
| -------------------- | ---------- | -------- |
| Pocket integration   | ⏳ Planned | 🟢 Low   |
| Readwise integration | ⏳ Planned | 🟢 Low   |
| Twitter integration  | ⏳ Planned | 🟢 Low   |

**Scope:**

- OAuth authentication
- Content import
- Automatic syncing
- API rate limiting
- Error handling

---

### Phase 11: Galactic Clusters 🟢 Low

Status: ⏳ Backlog
Description: Кластеризация графа и визуализация галактических скоплений.

□ Серверная кластеризация (graph-service) — алгоритм Louvain, API /clusters, поле cluster_id в /full.
□ 3D Orbital-режим с кластерами — каждый кластер образует свою «солнечную систему», центральная заметка кластера — звезда, остальные — планеты на орбитах.
□ LOD для дальних кластеров (отображение как единое тело).
□ Цветовое кодирование и размер кластеров в зависимости от числа заметок и суммарного веса.
□ Zoomable User Interface (ZUI) — двойной тап/клик по кластеру погружает внутрь (3D: `engine.diveIntoCluster()` / `surfaceToGalacticView()`; 2D: D3 zoom + фильтрация узлов по `cluster_id`), отображая только заметки этого кластера. Зависит от серверной кластеризации.

---

## 🎛️ Visualization Mode Switcher

| Mode      | Name                           | Status                      |
| --------- | ------------------------------ | --------------------------- |
| Free      | Free Flight (current D3-force) | ✅ Working                  |
| Honeycomb | Honeycomb View                 | ⏳ Ready for implementation |
| Clusters  | Galactic Clusters              | 🌙 Backlog                  |

---

## 🧭 Visualization Approach Comparison

| Aspect                    | Galactic Clusters                 | Honeycomb View                            |
| ------------------------- | --------------------------------- | ----------------------------------------- |
| Implementation complexity | High (backend + frontend + ASYNQ) | Low (frontend only)                       |
| Requires backend?         | Yes                               | No                                        |
| Chaos reduction           | Full (grouping + hidden lines)    | Partial (radial layout + visual encoding) |
| When to use               | When semantic grouping is needed  | When fast visual order is needed          |

---

**Last Updated:** August 30, 2026
**Next Review:** After the next roadmap milestone

---
## ⚠️ Technical Debt & Research

### TD-1: CelestialBody Auto-Assignment (MVP)

**Priority:** 🟡 Medium
**Status:** ⏳ Planned
**Description:** Temporary automatic assignment of celestial body type based on note size and connection count. Problem: currently CelestialBody types are assigned manually and lack semantic binding. Automatic determination via clustering and NLP is complex, requiring real usage data. Need temporary rule so types have at least basic meaning.

- **Character count grading:**
  - 0–100 → Fragment (asteroid)
  - 100–500 → Comet
  - 500–2000 → Planet
  - 2000–10000 → Star
  - 10000+ → Galaxy
- **Modifiers:**
  - > 5 connections → raise type one level (Planet → Star)
  - No connections → lower type (Star → Comet)
  - Orphan and <100 characters → Cosmic dust (barely visible)
- Type recalculated on each note save.
- **Important:** this is temporary solution. After clustering implementation and user behavior data accumulation, logic will be revised.
- **Dependencies:** none (uses existing size and connections fields).

### TD-2: CelestialBody Semantics Research

**Priority:** 🟢 Low
**Status:** ⚠️ Requires Research
**Description:** Research of semantic binding of CelestialBody types to cluster context.

- How do users manually assign types?
- Can a note be a "Star" in one cluster and a "Planet" in another?
- Is batch edit tool needed for type reassignment?
- Defer until real usage data available.
- **MVP for a semantic model (once research above is validated):** descriptions and usage examples for each celestial body type, surfaced as UI hints when choosing a type.
- **Implementation sketch:** update `celestial-body.ts` (per-type description/example fields), extend i18n keys, add hints to `TypeSelector.svelte`; document the final mapping in `docs/CELESTIAL_BODY_SEMANTICS.md`.

### TD-3: Full SSE Implementation to Replace HTTP Polling

**Priority:** 🟡 Medium (fixes a real performance/load problem)
**Status:** ⏳ Planned
**Context:** The application currently has **no** push channel to the browser at all. What
looks like "events" is actually plain HTTP polling:

- `GET /graph-service/api/v1/graph/delta?last_hash=...` — polled by the frontend every 30s
  plus on window focus (`frontend/src/features/home-page/home-page.svelte.ts`).
- `GET /api/v1/users/me/achievements` — still uses guarded interval polling in
  `frontend/src/entities/achievement/model/store.svelte.ts`; a configured interval of `0`
  disables the timer, preventing the former tight-loop behavior.
- The only real pub/sub in the project is the Redis Pub/Sub between backend and graph-service
  (`backend/internal/infrastructure/events/publisher.go` → `services/graph-service/internal/subscriber/pubsub.go`),
  used solely for graph-service cache invalidation — it never reaches the browser.
- ADR-015 already flagged SSE for achievements as TODO at decision time, but it was never
  implemented — polling stuck around instead.

**Description:** Design and implement a unified SSE layer (`text/event-stream`,
`EventSource` on the frontend) for every scenario that currently relies on polling or that
needs real-time notifications, reusing the existing Redis Pub/Sub infrastructure as the
transport between service instances.

**Candidates to migrate from polling/refresh-after-action to SSE:**

1. **Achievements** (`/api/v1/users/me/achievements`) — push an `AchievementUnlocked` event
   right after `achievement.Service.CheckTrigger` saves a `UserAchievement` (currently run in
   a background goroutine after the create note/link response — an ideal place to publish from).
2. **Graph delta** (`graph-service/api/v1/graph/delta`) — a `GraphChanged` event on
   NoteCreated/NoteUpdated/NoteDeleted/LinkCreated/LinkDeleted instead of polling every 30s;
   removes both the polling itself and the related "graph flicker" from failed deltas.
3. **Login streak progress** (`GetStreak`, which doesn't even have its own backend route yet —
   see the missing `GET /api/v1/users/me/streak` found during the OpenAPI audit) — can be
   designed push-first from the start.
4. **Sharing notifications** — when a note is shared with a user (`ShareNote`), the recipient
   could get an instant notification instead of discovering it on the next list load.
5. **Draft conflicts** (`draft.SyncDraft` / `ResolveConflict`) — the client currently has to
   discover sync conflicts itself; SSE could proactively notify about draft state changes from
   another client/tab.
6. **Cross-tab achievement/streak toasts in real time** — each tab currently runs its own
   independent polling; a single SSE connection per user removes the duplication.

**Architecture (plan):**

- Backend: a single SSE hub (e.g. `internal/infrastructure/sse`) — a registry of active
  connections keyed by `user_id`, with graceful disconnect/reconnect handling.
- Cross-instance transport: the same Redis Pub/Sub (`internal/infrastructure/events`),
  extended with new event types (`AchievementUnlocked`, `GraphChanged`, `ShareCreated`,
  `DraftConflict`) published from the relevant application services.
- Endpoint: `GET /api/v1/events/stream` (a single multiplexed stream for all event types for
  one user, using the SSE `event:` field for client-side routing) — preferred over one stream
  per feature.
- Frontend: a single `EventSource` client in `shared/services` with auto-reconnect
  (retry/backoff) that components subscribe to for the event types they need; on disconnect,
  temporarily fall back to the existing polling/refresh-after-action logic so functionality
  isn't lost when SSE is unavailable (proxy/firewall/old browser).
- Both nginx gateways (dev/personal) need to support long-lived HTTP connections without
  buffering (`proxy_buffering off`, increased `proxy_read_timeout`) — verify `docker/nginx/*`
  configs.

**MVP:** a single SSE endpoint (`/api/v1/events/stream`) for achievements only
(`AchievementUnlocked`), reusing the existing Redis Pub/Sub channel; expand to `GraphChanged`
and the remaining candidates above after it stabilizes.

**Dependencies:** no new external services (reuses existing Redis); needs sign-off from
`knowledge-graph-security` on authenticating long-lived SSE connections (JWT via query
parameter or a separate short-lived token, since `EventSource` doesn't support custom headers).
## 🎯 Priority Legend

- 🔴 **Critical:** Must complete before production deployment
- 🟠 **High:** Important for user experience, complete soon
- 🟡 **Medium:** Nice to have, complete when time permits
- 🟢 **Low:** Future enhancements, backlog items

---

## 📝 Status Legend

- ✅ **Done:** Completed and tested
- 🔄 **In Progress:** Currently being worked on
- ⏳ **Planned:** Scheduled for implementation
- 🌙 **Backlog:** Future consideration

---

