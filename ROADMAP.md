# Knowledge Graph Roadmap

**Updated:** July 26, 2026  
**Status:** Graph API unification and automation-robustness completed; awaiting final verification  
**Version:** v2.6

---

## 📊 Project Status

- **Development Phase:** Alpha → Beta transition
- **Regression Testing:** ✅ Partially complete (11/14 parts passed)
- **System Stability:** ✅ Stable (no critical issues)
- **Test Coverage:** ✅ Frontend unit tests (866 tests), Backend unit tests (all passed)
- **Production Readiness:** ⏳ Pending critical verifications (E2E, integration, CI/CD)

---

## 🎯 NOW: Current Focus

### 🔄 Manual Testing & Stabilization

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Manual testing of all features | ✅ Done | 🔴 Critical | - |
| Bug fixes from manual testing | ✅ Done | 🔴 Critical | - |
| Critical verifications for production | ✅ Done | 🔴 Critical | - |

**Subtasks:**
- [x] Frontend E2E smoke tests (`cd frontend && npm run test:smoke`)
- [x] Backend integration tests (`cd backend && go test -tags=integration ./...`)
- [x] CI/CD workflows verification
- [x] NLP API testing
- [x] Backend auth API testing
- [x] Public graph verification

### 🎨 UI/UX: Cosmic Cockpit

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Cosmic Cockpit UI — starship cockpit layout with slide-out panels, HUD, and first-person mode | ✅ Done | 🟠 High | 📝 Yes |

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
- [UI Duplication and Note Creation Analysis](docs/UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md) — describes the legacy UI duplication this redesign addresses

### ⚡ Performance & Resource Optimization

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Resource usage audit and optimization (memory, CPU, bundle size) | ⏳ Planned | 🟠 High | 📝 No |

**Audit findings:**
- **Frontend bundle:** `three` is imported as `import * as THREE from "three"` in 7+ files — full library may end up in the bundle. Bundle analyzer is not configured.
- **Frontend runtime:** `+page.svelte` uses `setInterval` for delta polling; `CockpitViewport` uses `requestAnimationFrame` loop — both need lifecycle cleanup checks.
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

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Event-driven invalidation & fallback model | ✅ Done | 🔴 Critical | 📝 Yes |
| Auth & user-scoped filtering in graph-service | ✅ Done | 🔴 Critical | 📝 Yes |
| Graph API unification & double-load removal | ✅ Done | 🟠 High | 📝 Yes |
| Graph analytics API (Neighbors, Path, Recommendations) | ✅ Done | 🟡 Medium | 📝 Yes |
| Materialized view / graph index | ✅ Done | 🟡 Medium | 📝 Yes |
| gRPC-web / SSE full-graph streaming | ⏳ Planned | 🟡 Medium | 📝 No |
| Improved layout algorithms (Honeycomb, force-directed, Cosmic Navigator) | 🔄 In Progress | 🟠 High | 📝 Yes |

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

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Create comprehensive deployment guide | ⏳ Planned | 🟠 High | 📝 Yes |

**Scope:**
- Docker deployment instructions
- Environment configuration
- Database setup and migrations
- NLP service setup
- SSL/TLS configuration
- Backup procedures
- Monitoring setup

**Related docs:**
- [Docker Deployment Guide](docs/DOCKER.md)

### 🌐 Public Note Pool (Publish/Unpublish)

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Implement publish/unpublish functionality | ✅ Done | 🟠 High | 📝 Yes |

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

### 🔗 Link Improvements

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Add tooltips for links | ⏳ Planned | 🟡 Medium | 📝 Yes |
| Improve link animations | ⏳ Planned | 🟡 Medium | 📝 Yes |
| Implement gamma-coding for link strength | ⏳ Planned | 🟡 Medium | 📝 Yes |
| Auto-link creation based on NLP | ⏳ Planned | 🟡 Medium | 📝 Yes |

**Scope:**
- Hover tooltips showing link metadata
- Animated link creation/deletion
- Visual strength indicators (gamma-coding)
- Link type differentiation
- Bidirectional link visualization
- Auto-create links on note creation when recommendation score is above threshold

**Related docs:**
- [Auto Link Creation Plan](docs/AUTO_LINK_CREATION_PLAN.md)

### 🧹 Dust Processor (Quick Notes Handler)

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Implement dust note processing | ⏳ Planned | 🟡 Medium | 📝 Yes |

**Scope:**
- Automatic dust note categorization
- Smart dust note suggestions
- Dust note to regular note conversion
- Batch dust note processing
- Dust note analytics
- Dust inbox panel and dust → note enrichment

**Related docs:**
- [Note Error Correction Plan](docs/NOTE_ERROR_CORRECTION_PLAN.md) (Quick Edit, autocompletion, dust inbox)

### 📝 Quick Edit & Note Autocompletion

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Quick inline note editing (Ctrl+E) | ⏳ Planned | 🟡 Medium | 📝 Yes |
| Note title/content autocompletion | ⏳ Planned | 🟡 Medium | 📝 Yes |
| NLP keyword/embedding correction | ⏳ Planned | 🟡 Medium | 📝 Yes |

**Scope:**
- Inline editing of title and content without opening full note page
- Autosave on blur with optimistic update and rollback on error
- Autocomplete suggestions for title, type, tags, and related notes
- User correction of NLP keywords and manual recomputation of embeddings

**Related docs:**
- [Note Error Correction Plan](docs/NOTE_ERROR_CORRECTION_PLAN.md)

---

### Phase 10: Honeycomb View 🟡 Medium

Status: ⏳ Planned
Description: Представление графа в виде сот (гексагональной сетки).

□ 2D Honeycomb — узлы привязаны к ячейкам сетки, связи — прямые линии. Веса связей влияют на близость ячеек.
□ 3D Honeycomb (Stellaris) — плоская гексагональная сетка в 3D-пространстве, связи — изогнутые трубки с неоновым свечением.
□ Переключение в режим сот из тулбара (2D и 3D).

### 🎤 Interactive Onboarding & Breadcrumbs

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Scenario-based narrator (T1-T6) | ⏳ Planned | 🟠 High | 📝 Yes |
| Breadcrumb progress (scenario + step) | ⏳ Planned | 🟠 High | 📝 Yes |
| Spotlight and element highlighting | ⏳ Planned | 🟡 Medium | 📝 Yes |
| Tour state, skip, and localStorage | ⏳ Planned | 🟡 Medium | 📝 Yes |

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

### 📥 Import/Export Tools

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| JSON import/export | ⏳ Planned | 🟢 Low | 📝 Yes |
| Markdown import/export | ⏳ Planned | 🟢 Low | 📝 Yes |
| CSV import/export | ⏳ Planned | 🟢 Low | 📝 Yes |

**Scope:**
- Bulk import/export functionality
- Format validation
- Data mapping
- Conflict resolution
- Import history

### 🎮 Gamification (Customization & Points)

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Customization system | ⏳ Planned | 🟢 Low | 📝 Yes |
| Points and achievements | ⏳ Planned | 🟢 Low | 📝 Yes |

**Scope:**
- User profile customization
- Point system for activities
- Achievement badges
- Leaderboards
- Progress tracking

### 🔌 Obsidian Integration

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Obsidian import | ⏳ Planned | 🟢 Low | 📝 Yes |
| Obsidian sync | ⏳ Planned | 🟢 Low | 📝 Yes |

**Scope:**
- Markdown file import
- Wiki-link conversion
- Tag mapping
- Metadata preservation
- Two-way sync

**Related docs:**
- [Obsidian Import Spec](docs/OBSIDIAN_IMPORT_SPEC.md)

### 📱 PWA Capture (Mobile Quick Notes)

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| PWA development | ⏳ Planned | 🟢 Low | 📝 Yes |
| Quick capture interface | ⏳ Planned | 🟢 Low | 📝 Yes |

**Scope:**
- Progressive Web App
- Offline support
- Quick note capture
- Mobile-optimized UI
- Push notifications

### 🌐 External API Integrations

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Pocket integration | ⏳ Planned | 🟢 Low | 📝 Yes |
| Readwise integration | ⏳ Planned | 🟢 Low | 📝 Yes |
| Twitter integration | ⏳ Planned | 🟢 Low | 📝 Yes |

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

---

## 🧪 Experimental & Ideas (Гипотезы и эксперименты)

Идеи, требующие проработки, прототипирования и проверки гипотез. Могут быть переведены в активные фазы после validation.

### Phase 13: Factory Line (Производственная линия) 🟢 Low — Hypothesis
**Priority:** 🟡 Medium (после кластеризации и сот)
**Status:** 💡 Idea
**Description:** Режим визуализации графа как производственной цепочки (Factorio / Shapez).
- Заметки отображаются как «станки» с входами и выходами.
- Связи — конвейеры с анимацией движения (анимация только для связей в текущем viewport, fallback — статические стрелки).
- Типы заметок мапятся на роли: `raw` (сырьё / черновик), `processor` (обработчик / анализ), `product` (продукт / вывод). Источник роли — существующее поле `note.type` или новый тег `factoryRole`.
- Обрыв цепи подсвечивается как «брак».
- Метрика продуктивности: число завершённых цепочек.
- **Гипотеза:** метафора завода мотивирует замыкать цепи и вести гигиену заметок.
- **MVP:** режим отображения с прямоугольными узлами и анимированными стрелками.
- **Зависимости:** нет (отдельный режим визуализации, но требуется ролевая классификация заметок).
- **Validation:** >30% новых связей замыкаются в завершённые цепочки за месяц; режим используется ≥1 раз в неделю активными пользователями.

### Phase 14: Semantic Guardians (Семантические стражи) 🟢 Low — Hypothesis
**Priority:** 🟢 Low (после кластеризации и архива)
**Status:** 💡 Idea
**Description:** Пассивная геймификация гигиены знаний через персонифицированного стража кластера.
- Каждый кластер получает стража — семантическую фигурку, отражающую суть кластера (герой, бог, автор, предмет). Пользователь называет стража сам.
- Страж привязан к кластеру и может быть перемещён между кластерами (drag-and-drop).
- Сила стража зависит от свежести заметок в кластере, общего количества связей, завершённости кластера и наличия архивных заметок.
- **Волны забвения** атакуют кластер, если он не обновлялся M дней. Страж автоматически отбивает атаку (авто-анимация), если кластер активен.
- Перемещение стража = переприоритизация: кластер-донор тускнеет, кластер-акцептор усиливается.
- **Семантическая обратная связь:** страж, названный пользователем («Достоевский», «Проект Альфа»), служит напоминанием о заброшенной теме.
- **Гипотеза:** персонифицированные стражи создают эмоциональную связь с заметками и мотивируют возвращаться к заброшенным темам.
- **MVP:** ручное именование стража кластера, drag-and-drop между кластерами, одна фигурка-силуэт, анимация волны забвения.
- **Зависимости:** Phase 11 (Galactic Clusters), Phase 15 (Archive & Note Hygiene), NLP-сервис для будущей авто-генерации имён.
- **Validation:** рост возвратов к «забытым» кластерам на ≥20% в течение месяца; ≥30% активных пользователей дают имя хотя бы одному стражу.

### Phase 15: Archive & Note Hygiene (Архив и гигиена заметок) 🟡 Medium — Planned
**Priority:** 🟡 Medium (перед Semantic Guardians)
**Status:** ⏳ Planned
**Description:** Механика автоматической архивации неиспользуемых заметок.
- Заметка, не обновлявшаяся N дней, с низкой активностью связей и не помеченная как важная, автоматически помечается как «забытая».
- Забытые заметки тускнеют в графе, связи прерываются.
- При достижении критического порога заметка уходит в **Архив** (скрывается из основного графа).
- Архив — отдельный слой графа, не влияющий на основную навигацию. Режим просмотра архива.
- Пользователь может вручную архивировать/разархивировать заметки.
- После Phase 14 страж кластера может защищать заметки в кластере от архивации.
- Метрика «здоровья графа»: процент активных заметок к общему числу.
- **Гипотеза:** автоматическая архивация снижает когнитивную нагрузку и поддерживает граф в актуальном состоянии.
- **MVP:** ASYNQ scheduled task для пометки «забытых» заметок, UI-переключатель в архив, endpoint для ручной архивации.
- **Зависимости:** нет (базовая механика над существующей моделью заметок); интеграция со стражами (Phase 14) — опционально.
- **Validation:** ≥70% автоматически архивированных заметок не разархивируются вручную в течение месяца; пользователи отмечают, что граф стал «чище» (опрос/фидбек).

### 💡 DB-backed Runtime Configuration 🟢 Low — Idea
**Priority:** 🟢 Low  
**Status:** 💡 Idea  
**Description:** Store rarely-changed runtime tunables in a `system_config` table and expose a protected admin API for CRUD + manual reload.

- **Scope:**
  - Runtime tunables only: rate limits, graph/recommendation/pagination limits, password policy, UI thresholds.
  - No secrets (JWT_SECRET, SMTP passwords, OAuth credentials remain env/JSON only).
- **API:**
  - `GET /api/v1/admin/system-config` — list all entries.
  - `POST /api/v1/admin/system-config` — create a key/value pair.
  - `PATCH /api/v1/admin/system-config/:key` — update value.
  - Protected by `STATIC_API_KEY` (or `RequireAdmin()` later).
- **Runtime effect:** load from DB at startup; manual reload endpoint to refresh in-memory `config.Config`.
- **Validation:** deferred to implementation; consider a strict key/type registry.
- **Risks:** drift between `knowledge-graph.config.json` and DB; caching/reload semantics across backend instances and worker.
- **MVP:** migration + repository + service + handler + reload endpoint + integration tests.
- **Dependencies:** existing `user_settings` pattern, `RequireAdmin` middleware, `CacheClient`.
- **Validation criteria:** changing a rate limit via API and reload affects live requests without restart; no invalid keys accepted after validation is implemented.

*Add future ideas above the next section break.*

---

## ✅ DONE: Completed Tasks

### 🏗️ Infrastructure & Stability

| Task | Status | Priority | Completion Date |
|------|--------|----------|-----------------|
| Stabilize dev/personal/test stacks | ✅ Done | 🔴 Critical | July 2026 |
| Fix 502 error on dev stack | ✅ Done | 🔴 Critical | July 2026 |
| Full regression testing plan (24 steps) | ✅ Done | 🔴 Critical | July 2026 |
| Automatic stacks identity check | ✅ Done | 🔴 Critical | July 2026 |
| Docker build verification | ✅ Done | 🔴 Critical | July 2026 |
| Worker for test stack configuration | ✅ Done | 🔴 Critical | July 2026 |
| CORS configuration via environment variables | ✅ Done | 🔴 Critical | July 2026 |
| Healthcheck verification in regression testing | ✅ Done | 🔴 Critical | July 2026 |
| Isolated testing model implementation | ✅ Done | 🔴 Critical | July 2026 |
| Automatic state verification (dev pre/post-test) | ✅ Done | 🔴 Critical | July 2026 |
| Auto-commit on successful regression cycle | ✅ Done | 🔴 Critical | July 2026 |
| Dev/Personal identity verification | ✅ Done | 🔴 Critical | July 2026 |
| CI/CD service containers, timeouts and Docker health checks | ✅ Done | 🔴 Critical | July 2026 |

### 🌌 Graph Service, Analytics & Public Graph

| Task | Status | Priority | Completion Date |
|------|--------|----------|-----------------|
| Event-driven invalidation & fallback model | ✅ Done | 🔴 Critical | July 2026 |
| Auth & user-scoped filtering in graph-service | ✅ Done | 🔴 Critical | July 2026 |
| Graph analytics API (Neighbors, Path, Recommendations) | ✅ Done | 🟡 Medium | July 2026 |
| Materialized view / graph index (`note_links_closure`) | ✅ Done | 🟡 Medium | July 2026 |
| Public Note Pool (publish/unpublish) | ✅ Done | 🟠 High | July 2026 |
| 3D graph via graph-service (`GraphServiceLayoutProvider`) | ✅ Done | 🟠 High | July 2026 |

### 🎨 Frontend Improvements

| Task | Status | Priority | Completion Date |
|------|--------|----------|-----------------|
| GraphCanvas FSD refactoring | ✅ Done | 🟠 High | July 2026 |
| NoteCard redesign (nebula gradient, indicators) | ✅ Done | 🟠 High | July 2026 |
| Galactic lexicon (multilingual support) | ✅ Done | 🟠 High | July 2026 |
| Interactive canvas controls (ghost node, drag-drop) | ✅ Done | 🟠 High | July 2026 |
| Black hole deletion with animation | ✅ Done | 🟠 High | July 2026 |
| Two-stage undo toast (Done → Restore) | ✅ Done | 🟠 High | July 2026 |
| HelpHotkeysModal component | ✅ Done | 🟠 High | July 2026 |
| Enhanced ghost node creation modal | ✅ Done | 🟠 High | July 2026 |
| Drag-and-drop link creation UX | ✅ Done | 🟠 High | July 2026 |
| Frontend unit tests (526 tests passing) | ✅ Done | 🔴 Critical | July 2026 |

### 🔧 Backend Improvements

| Task | Status | Priority | Completion Date |
|------|--------|----------|-----------------|
| Backend unit tests (all packages passing) | ✅ Done | 🔴 Critical | July 2026 |
| Clean Architecture implementation | ✅ Done | 🔴 Critical | Earlier |
| JWT authentication middleware | ✅ Done | 🔴 Critical | Earlier |
| Rate limiting on write endpoints | ✅ Done | 🔴 Critical | Earlier |
| NLP service lazy loading | ✅ Done | 🔴 Critical | Earlier |
| PGVECTOR extension setup | ✅ Done | 🔴 Critical | July 2026 |
| Asynchronous task processing (ASYNQ) | ✅ Done | 🔴 Critical | July 2026 |
| Redis caching integration | ✅ Done | 🔴 Critical | Earlier |
| MongoDB audit logs | ✅ Done | 🔴 Critical | Earlier |

### 🎯 Features

| Task | Status | Priority | Completion Date |
|------|--------|----------|-----------------|
| Achievements system | ✅ Done | 🟠 High | Earlier |
| Keyboard shortcuts (N, Del, Ctrl+Z, ?) | ✅ Done | 🟠 High | July 2026 |
| List view with batch operations | ✅ Done | 🟠 High | July 2026 |
| Selection mode with checkboxes | ✅ Done | 🟠 High | July 2026 |
| Bulk actions menu | ✅ Done | 🟠 High | July 2026 |
| Sort dropdown (created, updated, type) | ✅ Done | 🟠 High | July 2026 |
| Type filtering in list view | ✅ Done | 🟠 High | July 2026 |
| Note indicators (new, updated) | ✅ Done | 🟠 High | July 2026 |
| NoteCard tooltips with metadata | ✅ Done | 🟠 High | July 2026 |
| Accessibility improvements (ARIA labels) | ✅ Done | 🟠 High | July 2026 |

### 📚 Documentation

| Task | Status | Priority | Completion Date |
|------|--------|----------|-----------------|
| AGENTS.md with 11 specialized agents | ✅ Done | 🔴 Critical | July 2026 |
| REGRESSION_TEST_PLAN.md (20-part plan) | ✅ Done | 🔴 Critical | July 2026 |
| TESTING.md (testing infrastructure) | ✅ Done | 🔴 Critical | July 2026 |
| TESTING.md (Russian translation) | ✅ Done | 🔴 Critical | July 2026 |
| MANUAL_TEST_CHECKLISTS_RU.md | ✅ Done | 🔴 Critical | July 2026 |
| docs/archive/REGRESSION_TEST_REPORT.md | ✅ Done | 🔴 Critical | July 2026 |
| CORS configuration documentation | ✅ Done | 🔴 Critical | July 2026 |
| Healthcheck verification documentation | ✅ Done | 🔴 Critical | July 2026 |
| Testing commands in AGENTS.md | ✅ Done | 🔴 Critical | July 2026 |
| Testing commands in Devin skill | ✅ Done | 🔴 Critical | July 2026 |

### 🆕 Recently Completed (August 2026)

| Task | Status | Priority | Completion Date |
|------|--------|----------|-----------------|
| Cosmic Cockpit UI integration (4 panels, HUD, first-person mode) | ✅ Done | 🟠 High | August 2026 |
| Floating auth panel (draggable, non-blocking) | ✅ Done | 🟠 High | August 2026 |
| PUT /links/:id — link update with validation | ✅ Done | 🟠 High | August 2026 |
| Link weight recalculation (ASYNQ task + migration 028) | ✅ Done | 🟠 High | August 2026 |
| Backend graph analytics (/graph/analytics: PageRank, clusters, top centers) | ✅ Done | 🟡 Medium | August 2026 |
| NoteForm / TypeSelector decoupled from CelestialBody | ✅ Done | 🟠 High | August 2026 |
| FSD migration: components → widgets and entities | ✅ Done | 🟠 High | August 2026 |

---

## 📈 Progress Metrics

### Development Progress
- **Total Tasks:** 60+
- **Completed:** 37+ (62%)
- **In Progress:** 3 (5%)
- **Planned:** 20+ (33%)

### Testing Coverage
- **Frontend Unit Tests:** 866 tests ✅
- **Backend Unit Tests:** All packages ✅
- **Frontend E2E Tests:** ⏳ Pending
- **Backend Integration Tests:** ✅ Done (2026-07-23)
- **Regression Testing:** 11/14 parts ✅

### System Stability
- **Dev Stack:** ✅ Stable
- **Personal Stack:** ✅ Stable
- **Test Stack:** ✅ Stable
- **Critical Issues:** 0
- **Known Bugs:** 0 (pending manual testing)

---

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

## 🔗 Related Documentation

- [README.md](README.md) - Project overview and quick start
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture and design
- [AGENTS.md](docs/AGENTS.md) - AI agents and development workflows
- [REGRESSION_TEST_PLAN.md](docs/REGRESSION_TEST_PLAN.md) - Testing procedures
- [REGRESSION_TEST_REPORT.md](docs/archive/REGRESSION_TEST_REPORT.md) - Latest test results
- [TESTING.md](docs/TESTING.md) - Testing infrastructure guide

---

## 🎛️ Visualization Mode Switcher

| Mode | Name | Status |
|------|------|--------|
| Free | Free Flight (current D3-force) | ✅ Working |
| Honeycomb | Honeycomb View | ⏳ Ready for implementation |
| Clusters | Galactic Clusters | 🌙 Backlog |

---

## 🧭 Visualization Approach Comparison

| Aspect | Galactic Clusters | Honeycomb View |
|--------|-------------------|----------------|
| Implementation complexity | High (backend + frontend + ASYNQ) | Low (frontend only) |
| Requires backend? | Yes | No |
| Chaos reduction | Full (grouping + hidden lines) | Partial (radial layout + visual encoding) |
| When to use | When semantic grouping is needed | When fast visual order is needed |

---

**Last Updated:** July 23, 2026  
**Next Review:** After manual testing completion

---

## 🧪 Experimental & Ideas

### Phase 13: Factory Line 🟢 Low — Hypothesis
**Priority:** 🟡 Medium (after clustering and honeycomb)
**Status:** 💡 Idea
**Description:** Graph visualization as a production chain (Factorio / Shapez).
- Notes displayed as "machines" with inputs and outputs.
- Links are conveyors with movement animations.
- Note types: raw material (draft), processor (analysis), product (conclusion).
- Broken chain highlighted as "defect".
- Productivity metric: number of completed chains.
- **Hypothesis:** factory metaphor motivates closing chains and maintaining note hygiene.
- **MVP:** display mode with rectangular nodes and animated arrows.
- **Dependencies:** none (separate visualization mode).

### Phase 14: Semantic Guardians 🟢 Low — Hypothesis
**Priority:** 🟢 Low (after archive and hygiene)
**Status:** 💡 Idea
**Target:** Mobile (PWA first, native later)
**Description:** Passive gamification of knowledge hygiene through personified cluster guardians. Main interface — mobile app (PWA or native).
- Each cluster gets one guardian — a semantic figure reflecting the cluster's essence (hero, god, author, subject). User names the guardian.
- Guardian is bound to the cluster and can be moved between clusters (drag-and-drop, touch-optimized).
- Guardian strength depends on note freshness, connection count, completeness.
- **Waves of oblivion** passively attack clusters (daily/weekly). Guardians automatically defend (auto-animation).
- Push notifications: "Wave of oblivion approaching cluster 'Philosophy'. Dostoevsky ready to defend."
- Moving guardian = reprioritization: donor cluster dims, acceptor cluster strengthens.
- **Semantic feedback:** guardian serves as reminder of abandoned topic.
- **Mobile-first:** main guardian interface — mobile companion app. Desktop — view state only.
- **Hypothesis:** mobile app with guardians increases retention and motivates returning to notes.
- **MVP:** PWA with cluster display, drag-and-drop guardians, wave animations, push notifications.
- **Dependencies:** none (independent feature).

### Phase 15: Archive & Note Hygiene 🟡 Medium — Planned
**Priority:** 🟡 Medium (before Semantic Guardians)
**Status:** ⏳ Planned
**Description:** Automatic archiving of unused notes.
- Note not updated for N days automatically marked as "forgotten".
- Forgotten notes dim in graph, connections break.
- At critical threshold, note moves to **Archive** (hidden from main graph).
- Archive — separate graph layer, not affecting main navigation. Archive view mode.
- User can manually archive/unarchive notes.
- "Graph health" metric: percentage of active notes to total.
- **Hypothesis:** automatic archiving reduces cognitive load and keeps graph current.
- **MVP:** scheduled task on backend to mark "forgotten" notes, UI toggle for archive, endpoint for manual archiving.
- **Dependencies:** none (basic mechanic over existing note model).

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
  - >5 connections → raise type one level (Planet → Star)
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

### TD-3: Full SSE Implementation to Replace HTTP Polling
**Priority:** 🟡 Medium (fixes a real performance/load problem)
**Status:** ⏳ Planned
**Context:** The application currently has **no** push channel to the browser at all. What
looks like "events" is actually plain HTTP polling:
- `GET /graph-service/api/v1/graph/delta?last_hash=...` — polled by the frontend every 30s
  plus on window focus (`refreshAfterMutation` in `frontend/src/routes/+page.svelte`).
- `GET /api/v1/users/me/achievements` — previously polled at `poll_interval_ms`
  (bug: a config value of `0` was treated by JS as "poll as fast as possible" rather than
  "disabled" — `setInterval(fn, 0)` fired ~125 times/sec instead of never firing; fixed by
  switching to event-driven refresh after user actions instead of an interval).
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