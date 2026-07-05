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

**Pending:**
- Improve ghost node creation flow with tippy form.
- Improve drag-and-drop link creation UX.
- Add "Delete all links" button in note side panel link management.
- Extract shared `HelpHotkeysModal` component and reuse on graph + list pages.
- Update visual and integration tests.

### Iteration 4 — Multilingual lexicon

- Refactor `frontend/src/lib/utils/galactic-lexicon.ts` to `{ ru: string, en: string }` structure.
- Add `formatMessage(key, lang, params)` helper.
- Add `frontend.language` to `knowledge-graph.config.json` (default `en`).
- Add language toggle in user settings UI (saved to `user_settings`).
- Replace all UI strings with `formatMessage` calls.
- Add unit tests for `formatMessage`.

### Iteration 5 — Tests and documentation

- Update Playwright visual screenshots for NoteCard and canvas states.
- Add unit tests for new indicators, tooltip, selection.
- Add integration tests for ghost node creation, black hole deletion, drag-and-drop link creation.
- Add BDD scenarios if applicable.
- Update documentation with final hotkey list and UX guidelines.

### Current status
- Iteration 1 completed.
- Iteration 2 completed.
- Iteration 3 completed (keyboard shortcuts, bulk actions menu, visual feedback).
- Iteration 4 in progress (hotkeys added, two-stage undo toast implemented).

