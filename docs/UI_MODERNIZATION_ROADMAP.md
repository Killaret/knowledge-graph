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
- Add `deleteNotesBatch(ids: string[])` to `frontend/src/shared/api/notes.ts`.
- Add unit test for batch delete API call.

**NoteCard redesign**
- Rewrite `frontend/src/components/molecules/NoteCard.svelte` with:
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
- Russian UI strings by default (`ru` locale); English available via i18n.

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
- Russian UI strings by default (`ru` locale); English available via i18n

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
- User-facing content defaults to Russian (`ru` locale); English (`en`) available via i18n
- Two-stage undo toast: "Done" (1.5s) → "Restore" (5s) — use i18n keys
- Visual feedback for all interactions (hover, selection, drag)
- Accessibility: proper ARIA labels, keyboard navigation
- Responsive design with mobile support

**Pending:**
- None - Iteration 5 complete.

### Iteration 6 — FSD Refactoring (Feature-Sliced Design + Atomic Design)

**Completed:**
- Configured SvelteKit aliases (`svelte.config.js`) and TypeScript paths (`tsconfig.json`) for the new layout
- `frontend/src/lib/` removed; code split between:
  - `frontend/src/shared/` — api, stores, services, utils, types, mocks, test-utils, styles, config
  - `frontend/src/components/` — Atomic Design: atoms, molecules, organisms
  - `frontend/src/features/` — feature slices:
    - `features/graph-interaction/` — drag-and-drop, hotkeys, zoom-pan
    - `features/graph-forms/` — note form, link form
  - `frontend/src/shared/lib/graph/` — renderer helpers, anomaly drawers, graph utilities
- Refactored `GraphCanvas.svelte` internals into `frontend/src/components/organisms/GraphCanvas/` submodules (ghost-node, gravity-system, renderer, etc.)
- Extracted graph interaction logic to `frontend/src/features/graph-interaction/`
  - `drag-and-drop.ts`, `hotkeys.ts`, `zoom-pan.ts`
- Extracted graph form logic to `frontend/src/features/graph-forms/`
  - `note-form.ts`, `link-form.ts`
- Updated all imports to use current aliases (`$shared/*`, `$components/*`, `$features/*`); legacy `$lib` alias removed
- Unit tests, build, and Svelte check pass successfully

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

