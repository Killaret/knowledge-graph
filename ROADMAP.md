# Knowledge Graph Roadmap

**Updated:** July 25, 2026  
**Status:** Graph API unification completed; awaiting verification  
**Version:** v2.5

---

## 📊 Project Status

- **Development Phase:** Alpha → Beta transition
- **Regression Testing:** ✅ Partially complete (11/14 parts passed)
- **System Stability:** ✅ Stable (no critical issues)
- **Test Coverage:** ✅ Frontend unit tests (526 tests), Backend unit tests (all passed)
- **Production Readiness:** ⏳ Pending critical verifications (E2E, integration, CI/CD)

---

## 🎯 NOW: Current Focus

### 🔄 Manual Testing & Stabilization

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Manual testing of all features | 🔄 In Progress | 🔴 Critical | - |
| Bug fixes from manual testing | 🔄 In Progress | 🔴 Critical | - |
| Critical verifications for production | ⏳ Planned | 🔴 Critical | - |

**Subtasks:**
- [ ] Frontend E2E tests (`cd frontend && npx playwright test`)
- [x] Backend integration tests (`cd backend && go test -tags=integration ./...`)
- [x] CI/CD workflows verification
- [ ] NLP API testing
- [ ] Backend auth API testing
- [x] Public graph verification

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

### 2. Auth & user-scoped filtering in graph-service

- Add JWT middleware to `graph-service` HTTP and gRPC.
- Derive `user_id` from token and remove public `user_id` query parameter.
- Add `creator_id` filter in `services/graph-service/internal/db/postgres_client.go` for all `notes`/`links` queries.
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

### 🚀 Spaceship Navigator (Cosmic Navigator)

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Implement cosmic navigation interface | ✅ Implemented | 🟡 Medium | 📝 Yes |

**Scope:**
- 3D navigation metaphor
- Galaxy/constellation view
- Zoom levels and navigation controls
- Search and filtering in navigator
- Integration with existing graph view
- [x] **Hybrid 3D Layout via Graph Service** — runtime layout provider switching via `createLayoutProvider` / `toRuntimeConfig`.
- [x] **3D Fog Presets** — configurable `birth` / `nebula` / `deep-space` fog presets with smooth initial→final transitions.
- [x] **3D Performance Presets** — FPS-based auto-downgrade/upgrade of fog quality, starfield count, and auto-rotate.

### 🔗 Link Improvements

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Add tooltips for links | ⏳ Planned | 🟡 Medium | 📝 Yes |
| Improve link animations | ⏳ Planned | 🟡 Medium | 📝 Yes |
| Implement gamma-coding for link strength | ⏳ Planned | 🟡 Medium | 📝 Yes |

**Scope:**
- Hover tooltips showing link metadata
- Animated link creation/deletion
- Visual strength indicators (gamma-coding)
- Link type differentiation
- Bidirectional link visualization

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

---

### 🍯 Honeycomb View (Graph Visualization Mode)

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Implement radial node placement by link weight | ⏳ Planned | 🟡 Medium | 📝 Yes |
| Add view mode switcher (Free / Honeycomb / Clusters) | ⏳ Planned | 🟡 Medium | 📝 Yes |
| Implement hover-reveal link visualization | ⏳ Planned | 🟡 Medium | 📝 Yes |

**Scope:**
- Center node selected by max `sum(weight)` across all links
- Radial placement on concentric rings by link weight thresholds (>0.7, 0.3-0.7, <0.3)
- Line thickness, opacity, and color based on weight and link type
- Hidden links by default in Honeycomb mode; reveal on node hover with 300ms fade-out
- Mode switcher in FloatingControls with Free / Honeycomb / Clusters options
- Persist selected mode in `graphSettings` (localStorage / user settings)
- Use D3 force.find or spatial hash for fast cursor-to-node lookup

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

### 🌌 Galactic Clusters (Semantic Clustering)

| Task | Status | Priority | Prompt Ready |
|------|--------|----------|-------------|
| Backend clustering service for semantic grouping | ⏳ Planned | 🟢 Low | 📝 No |
| Background cluster recalculation | ⏳ Planned | 🟢 Low | 📝 No |
| Canvas visualization for clusters | ⏳ Planned | 🟢 Low | 📝 No |

**Scope:**
- Semantic clustering of notes based on link relationships and embeddings
- Dedicated backend clustering service with background recalculation (ASYNQ)
- Cluster rendering on canvas
- Integration with view mode switcher (Clusters mode)

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

---

## 📈 Progress Metrics

### Development Progress
- **Total Tasks:** 53+
- **Completed:** 30+ (57%)
- **In Progress:** 3 (6%)
- **Planned:** 20+ (38%)

### Testing Coverage
- **Frontend Unit Tests:** 803 tests ✅
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