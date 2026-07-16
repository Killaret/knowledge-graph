# Agents in Knowledge Graph

**Updated:** July 2026
**Status:** See [AGENTS_EN.md](AGENTS_EN.md) for full documentation (11 agents)

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

**All user-facing content MUST be in English:**
- UI strings (buttons, labels, placeholders, errors)
- Toast messages and tooltips (GalacticLexicon)
- Note titles and content
- Commit messages

**Exceptions:** Internal code comments (any language OK)

```typescript
// ✅ Good
toast.success("Note created successfully");

// ❌ Bad
toast.success("Заметка создана успешно");
```

---

## Regression Testing Plan

**Comprehensive 24-step isolated regression testing plan for production readiness:**

### Regression Test Cycle (Isolated Model)
- **Document:** `docs/REGRESSION_TEST_PLAN.md`
- **Script:** `scripts/run-full-test-cycle.ps1` (Windows) or `scripts/run-full-test-cycle.sh` (Linux/Mac)
- **Identity Check:** `scripts/check-stacks-identity.ps1` (verifies dev/personal/test consistency)
- **Health Check:** `scripts/check-stacks-health.ps1 -Stack <dev|personal|test|all>`

### Isolated Testing Model
**⚠️ IMPORTANT:** The test cycle uses an isolated model where dev and personal stacks are stopped during testing to prevent resource conflicts and ensure accurate test results.

### Test Steps (24 total)
1. **Step 0:** Capture dev stack state snapshot (containers, health, API)
2. **Step 1:** Stop dev stack (`docker compose down`)
3. **Step 2:** Stop personal stack (`docker compose -f docker-compose.personal.yml down`)
4. **Step 3:** Check stacks identity (dev/personal/test consistency)
5. **Step 4:** Start test stack (`start-test.ps1`)
6. **Step 5:** Seed test data (`seed-test-data.ps1`)
7. **Step 6:** Docker build verification
8. **Step 7:** NLP service tests
9. **Step 8:** Backend unit tests
10. **Step 9:** Backend API verification
11. **Step 10:** Asynchronous tasks verification
12. **Step 11:** PGVECTOR verification
13. **Step 12:** Redis & MongoDB verification
14. **Step 13:** Frontend unit tests
15. **Step 14:** Manual testing instructions
16. **Step 15:** Public graph verification
17. **Step 16:** CI/CD verification
18. **Step 17:** Stop test stack (`stop-test.ps1`)
19. **Step 18:** Start dev stack (`docker compose up -d --wait`)
20. **Step 19:** Start personal stack (`docker compose -f docker-compose.personal.yml up -d --wait`)
21. **Step 20:** Compare dev stack state with pre-test snapshot
22. **Step 21:** Compare dev and personal stacks for identity
23. **Step 22:** Check dev and personal stacks health
24. **Step 23:** Auto-commit with test success marker (if all checks pass)

### Automatic State Verification
- **Pre-test snapshot:** Captures dev stack state before testing
- **Post-test comparison:** Compares dev stack state after testing
- **Dev/Personal identity:** Verifies dev and personal stacks are identical
- **Auto-commit:** Only if dev state unchanged and dev/personal identical
- **Failure handling:** Stops with exit code 1 if differences found

### Frequency
- **Full Regression:** Before each production deployment
- **Quick Regression:** Before each major feature release (Steps 0-6, 8, 13)
- **Smoke Regression:** After each minor feature release (Steps 0-2, 8-9, 15)
- **Identity Check:** Before each manual testing session

### Exit Criteria
- **PASS:** Dev state unchanged, dev/personal identical, all stacks healthy, all tests pass
- **FAIL:** Dev state changed, dev/personal not identical, stacks not healthy, test failure, infrastructure failure, data leakage

### See Also
- [REGRESSION_TEST_PLAN.md](REGRESSION_TEST_PLAN.md) — Complete regression testing procedures
- [TESTING_EN.md](TESTING_EN.md) — Testing infrastructure and procedures
- [FINAL_TEST_REPORT.md](archive/FINAL_TEST_REPORT.md) — Latest test results

---

## Testing Commands & Procedures

**Comprehensive testing commands for AI agents and developers:**

### Frontend Testing
```bash
# Unit tests (Vitest)
cd frontend && npm run test:unit

# E2E tests (Playwright)
cd frontend && npx playwright test

# Visual regression tests
cd frontend && npx playwright test --grep="@visual"

# BDD tests (Cucumber)
cd frontend && npm run test:bdd

# Build verification
cd frontend && npm run build
```

### Backend Testing
```bash
# Unit tests
cd backend && go test ./...

# Integration tests
cd backend && go test -tags=integration ./...

# Race detection (requires CGO_ENABLED=1)
cd backend && CGO_ENABLED=1 go test -race ./...

# Build verification
cd backend && go build ./cmd/server
```

### NLP Service Testing
```bash
# Unit tests
cd nlp-service && pytest tests/ -v

# Health check
curl http://localhost:5000/health

# API tests
curl -X POST http://localhost:5000/extract_keywords -H "Content-Type: application/json" -d '{"text":"test","top_n":3}'
curl -X POST http://localhost:5000/embed -H "Content-Type: application/json" -d '{"text":"test"}'
```

### Test Stack Management
```bash
# Full test cycle (isolated model - stops dev/personal stacks)
.\scripts\run-full-test-cycle.ps1      # Windows
./scripts/run-full-test-cycle.sh       # Linux/Mac

# Start test stack
.\scripts\start-test.ps1              # Windows
./scripts\start-test.sh               # Linux/Mac

# Seed test data
.\scripts\seed-test-data.ps1          # Windows
./scripts\seed-test-data.sh           # Linux/Mac

# Stop and destroy test stack
.\scripts\stop-test.ps1               # Windows
./scripts\stop-test.sh                # Linux/Mac

# Manual test stack management
docker compose -f docker-compose.test.yml up -d --build
docker compose -f docker-compose.test.yml down -v
```

### Stack Health Checks
```bash
# Check all stacks (default)
.\scripts\check-stacks-health.ps1              # Windows
./scripts/check-stacks-health.sh               # Linux/Mac

# Check specific stack
.\scripts\check-stacks-health.ps1 -Stack dev   # Windows
.\scripts\check-stacks-health.ps1 -Stack personal
.\scripts\check-stacks-health.ps1 -Stack test
./scripts/check-stacks-health.sh --stack dev    # Linux/Mac
./scripts/check-stacks-health.sh --stack personal
./scripts/check-stacks-health.sh --stack test
```

### Regression Testing
```bash
# Full regression cycle (isolated model - 24 steps)
.\scripts\run-full-test-cycle.ps1      # Windows
./scripts/run-full-test-cycle.sh       # Linux/Mac

# Stacks identity check
.\scripts\check-stacks-identity.ps1    # Windows
./scripts\check-stacks-identity.sh     # Linux/Mac

# Individual regression steps
# Step 0: Capture dev stack state snapshot
# Step 1-2: Stop dev and personal stacks
# Step 3: Check stacks identity
# Step 4: Start test stack
# Step 5: Seed test data
# Step 6-16: Run tests and verifications
# Step 17: Stop test stack
# Step 18-19: Restore dev and personal stacks
# Step 20-22: Compare states and verify identity
# Step 23: Auto-commit if all checks pass
```

### Database Verification
```bash
# PostgreSQL (test stack)
docker exec kg-test-postgres psql -U kb_user -d knowledge_test -c "SELECT extname FROM pg_extension WHERE extname = 'vector';"
docker exec kg-test-postgres psql -U kb_user -d knowledge_test -c "SELECT COUNT(*) FROM note_embeddings;"

# Redis (test stack)
docker exec kg-test-redis redis-cli PING
docker exec kg-test-redis redis-cli KEYS "*"

# MongoDB (test stack)
docker exec kg-test-mongo mongosh --eval "db.adminCommand('ping')"
docker exec kg-test-mongo mongosh --eval "db.getCollectionNames()"
```

### Health Checks
```bash
# Dev stack
curl http://localhost:8080/health           # Nginx gateway
curl http://localhost:9000/health           # Backend
curl http://localhost:9091/health           # Graph service
curl http://localhost:8080/api/v1/notes     # Notes API

# Personal stack
curl http://localhost:8085/health           # Personal backend
curl http://localhost:8082/health           # Personal API gateway
curl http://localhost:8092/health           # Personal graph service

# Test stack
curl http://localhost:8083/health           # Test backend
curl http://localhost:3002                  # Test frontend
curl http://localhost:15002/health          # Test NLP service
```

### Testing Best Practices
- **ALWAYS** use isolated test stack for E2E and BDD testing
- **NEVER** run E2E/BDD tests against dev or personal stacks
- **ALWAYS** verify stacks identity before regression testing
- **ALWAYS** destroy test stack with `down -v` after testing
- **ALWAYS** use English text patterns in frontend tests (language policy)
- **ALWAYS** run unit tests before integration tests
- **ALWAYS** verify health endpoints before API testing
- **ALWAYS** use the full test cycle script for regression testing (isolated model)
- **ALWAYS** check dev stack state before/after testing for data leakage
- **ALWAYS** verify dev/personal identity after testing

---

## UI Modernization Roadmap

Comprehensive plan for interactive canvas, NoteCard redesign, multilingual lexicon, and improved UX.

### Iteration 1 — Foundation (backend batch API + NoteCard redesign)

**Backend**
- Add `DeleteBatch(ctx, ids []uuid.UUID) error` to `internal/domain/note/repository.go`.
- Implement `DeleteBatch` in `internal/infrastructure/db/postgres/note_repo.go` using soft delete.
- Add `DeleteBatch` handler in `internal/interfaces/api/notehandler/note_handler.go`.
- Register route `POST /api/v1/notes/batch` in `backend/cmd/server/router.go` with rate limiter.
- Add stubs to mock repositories in tests (`draft/service_test.go`, `graph_handler_test.go`, `link_handler_test.go`, `notehandler/mock_repo.go`).
- Add unit test in `note_handler.go` test suite and repository tests.

**Frontend API**
- Add `deleteNotesBatch(ids: string[])` to `frontend/src/lib/api/notes.ts`.
- Add unit test for batch delete API call.

**NoteCard redesign**
- Rewrite `frontend/src/lib/components/NoteCard.svelte` with:
  - Nebula gradient background (`--color-surface-elevated`), 12px border radius.
  - 4px left type stripe (`--color-star`, `--color-planet`, `--color-comet`, `--color-galaxy`, `--color-asteroid`).
  - Type emoji in top-left corner.
  - Hover lift and type-colored glow.
  - "New" indicator: yellow pulsing dot for notes < 24h.
  - "Updated" indicator: turquoise dot for notes updated < 1h.
  - Special "dust" style for quick-capture notes.
- Add `tippy.js` tooltip on hover showing: up to 3 keywords, link count, Edit/Delete buttons.
- Add selection checkbox and `selectMode` / `selected` / `onSelect` props.
- Add staggered fade-in and fade-out animations.
- Improve accessibility: `role="article"`, `aria-label`, `aria-checked`.
- Update `NoteCard.spec.ts` with new tests for indicators, tooltip, selection.

**Acceptance criteria**
- Backend tests pass.
- Frontend build succeeds.
- NoteCard renders redesigned visuals, indicators, tooltip, and checkbox.
- English UI strings only.

### Iteration 2 — Page-level selection, sorting, undo

**Page integration**
- Update `frontend/src/routes/+page.svelte` list view with:
  - Sort dropdown: created, updated, type.
  - Select all / clear selection.
  - Floating action bar for batch delete with counter.
  - Batch delete confirmation via `ConfirmModal`.
  - Two-stage undo toast: first "Done" (1.5s), then "Restore" (5s).
  - Restore uses existing `restoreNote` API.
- Update empty state with English galactic lexicon text.
- Add integration tests for batch delete and undo.

### Iteration 3 — List view enhancements

**Keyboard shortcuts**
- Add `Ctrl+A` to select all notes in list view
- Add `Delete`/`Backspace` to batch delete selected notes
- Add `Escape` to clear selection or close side panel

**Bulk actions menu**
- Add "Actions" button to floating batch panel
- Add dropdown menu with options:
  - Move to type (placeholder)
  - Add tags (placeholder)
  - Export notes (placeholder)

**Visual feedback**
- Add selection count badge in floating panel
- Add confirmation dialog for batch delete
- Improve visual styling for bulk actions menu

**Acceptance criteria**
- Keyboard shortcuts work in list view
- Bulk actions menu displays correctly
- Confirmation dialog appears before batch delete
- English UI strings only

### Iteration 4 — Canvas improvements

**Completed:**
- Add hotkeys: `N` (ghost node), `Del`/`Backspace` (delete selected), `Ctrl+Z` (undo placeholder).
- Improve black hole deletion animation and two-stage undo toast (Done → Restore stages).
- Add "Delete all links" button in note side panel link management with confirmation modal.
- Extract shared `HelpHotkeysModal` component and integrate into GraphCanvas.
- Reuse `HelpHotkeysModal` on list page with menu integration.
- Improve ghost node creation flow with enhanced modal design (backdrop blur, gradients, improved styling).
- Improve drag-and-drop link creation UX with visual preview link and enhanced link form styling.
- Update visual and integration tests (added keyboard shortcut tests, visual regression tests for new UI states).

**Pending:**
- None - Iteration 4 complete.

### Iteration 4 — Multilingual lexicon

**Completed:**
- Created new i18n.ts with formatMessage helper
- Added en/ru message dictionary with common UI strings
- Added formatMessage, getCurrentLocale, setLocale functions
- Added frontend.language to knowledge-graph.config.json (default: en)
- Added language toggle in ProfileEditor component
- Replaced UI strings in ProfileEditor with formatMessage calls
- Added unit tests for i18n functions (16 tests passing)

**Pending:**
- None - Iteration 4 Multilingual complete.

### Iteration 5 — Tests and documentation

**Completed:**
- Updated Playwright visual screenshots for NoteCard and canvas states
- Added unit tests for new indicators, tooltip, selection (12 tests passing)
- Integration tests for ghost node creation, black hole deletion, drag-and-drop link creation already exist
- BDD scenarios skipped (not applicable for current scope)

**Documentation - Final Hotkey List:**

**Graph Canvas:**
- `N` - Create ghost node at center
- `Delete`/`Backspace` - Delete selected node
- `Ctrl+Z` - Undo (placeholder)
- `?` - Open help modal

**List View:**
- `Ctrl+A` - Select all notes
- `Delete`/`Backspace` - Batch delete selected notes
- `Escape` - Clear selection or close side panel

**UX Guidelines:**
- All user-facing content in English only
- Two-stage undo toast: "Done" (1.5s) → "Restore" (5s)
- Visual feedback for all interactions (hover, selection, drag)
- Accessibility: proper ARIA labels, keyboard navigation
- Responsive design with mobile support

**Pending:**
- None - Iteration 5 complete.

### Iteration 6 — FSD Refactoring (Feature-Sliced Design)

**Completed:**
- Configured SvelteKit alias (svelte.config.js) and TypeScript paths (tsconfig.json) for FSD
- Created FSD folder structure:
  - `features/graph-rendering/nodes/` - Node rendering logic
  - `features/graph-rendering/anomalies/` - Black holes, ghost nodes
  - `features/graph-interaction/` - Drag-and-drop, hotkeys, zoom-pan
  - `features/graph-forms/` - Note form, link form
  - `shared/lib/graph/` - Renderer, animations, types
  - `widgets/graph-canvas/` - GraphCanvas component
- Refactored renderer.ts from 1540 KB to 35 KB (97% reduction)
  - Extracted node drawing functions to `features/graph-rendering/nodes/`
  - Extracted anomaly functions to `features/graph-rendering/anomalies/`
  - Extracted helper functions to `shared/lib/graph/`
  - Removed duplicate functions
  - Updated imports to use FSD path aliases
- Refactored GraphCanvas.svelte from 45 KB to 41.7 KB (7.3% reduction)
  - Extracted drag-and-drop logic to `features/graph-interaction/drag-and-drop.ts`
  - Extracted hotkeys logic to `features/graph-interaction/hotkeys.ts`
  - Extracted zoom-pan logic to `features/graph-interaction/zoom-pan.ts`
  - Extracted form logic to `features/graph-forms/note-form.ts` and `link-form.ts`
  - Updated imports to use FSD path aliases
  - Fixed all lint errors
- Legacy `GraphCanvas/` structure preserved for backward compatibility
- All builds pass successfully
- All changes committed and pushed (commits 985d32f, 656b0c0, 5b7f0a8, 2237904)

**Pending:**
- None - Iteration 6 complete.

### Current status
- Iteration 1 completed.
- Iteration 2 completed.
- Iteration 3 completed (keyboard shortcuts, bulk actions menu, visual feedback).
- Iteration 4 completed (hotkeys added, two-stage undo toast implemented, Delete all links button added, HelpHotkeysModal component extracted, ghost node form enhanced, drag-and-drop link creation UX improved, tests updated).
- Iteration 4 Multilingual completed (i18n infrastructure created, language toggle added, ProfileEditor internationalized).
- Iteration 5 completed (visual screenshots updated, unit tests added, integration tests verified, documentation updated with hotkey list and UX guidelines).
- Iteration 6 completed (FSD refactoring: renderer.ts reduced 97%, GraphCanvas.svelte reduced 7.3%, FSD structure created, build passes).

