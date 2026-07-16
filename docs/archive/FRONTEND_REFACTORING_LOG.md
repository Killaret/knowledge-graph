# Frontend Refactoring Log — FSD + Atomic Design

This log tracks the batch-by-batch reorganization of the frontend from the `src/lib` monolith to FSD/Atomic Design layers.

## Phase 5 Plan

- Batch 5.1: Move `src/lib/api` → `src/shared/api` and update imports.
- Batch 5.2: Move `src/lib/stores` → `src/shared/stores` and update imports.
- Batch 5.3: Move `src/lib/utils` → `src/shared/utils` and update imports.
- Batch 5.4: Move `src/lib/hooks` → `src/shared/hooks` and update imports.
- Batch 5.5: Move `src/lib/services` → `src/shared/services` and update imports.
- Batch 5.6: Move `src/lib/types` → `src/shared/types`, `src/lib/mocks` → `src/shared/mocks`, `src/lib/test-utils` → `src/shared/test-utils`, `src/lib/styles` → `src/shared/styles`.
- Batch 5.7: Move `src/lib/components` → `src/components` with `atoms/`, `molecules/`, `organisms/` and update aliases/imports.
- Batch 5.8: Verify build and run unit tests.

## Completed

### Phases 0-4

- Phase 0: WSL Docker cleaned (Docker Desktop WSL data reset, `docker_data.vhdx` recreated, 32+ GB free).
- Phase 1: Backend dead code removed (`internal/services/note_service.go`, `internal/services/cleanup.go`, `internal/application/draft/handler.go`).
- Phase 2: Config unified (`src/shared/config/config.ts` is canonical, `src/lib/config.ts` removed).
- Phase 3: Graph rendering consolidated (`features/graph-rendering` removed, `shared/lib/graph/renderer` reduced to anomalies, `animation-utils.js` and `animations` removed).
- Phase 4: Unused FSD aliases `$entities` and `$widgets` removed from `svelte.config.js` and `vitest.config.ts`.

### Phase 5

- Batch 5.1: `src/lib/api` → `src/shared/api` completed. All imports updated to `$shared/api` (and `../api` references in `src/lib/stores/auth.svelte.test.ts`). `npm run test:unit` passed: 52 files passed, 2 skipped, 526 tests passed, 37 skipped.
- Batch 5.2-5.6: All remaining shared logic (`stores`, `utils`, `hooks`, `services`, `mocks`, `test-utils`, `types`, `styles`) moved to `src/shared`. `src/lib/assets` removed. `src/shared/types/index.ts` now re-exports `errors`. `vitest.config.ts` `$app` mock aliases updated to `src/shared/mocks`.
- Batch 5.7-5.8: `src/lib/components` moved to `src/components` and organized into `atoms`, `molecules`, `organisms` (Atomic Design). `src/lib` removed. `index.ts` barrel removed and modal imports converted to direct alias imports. `src/features/graph-interaction` imports updated to `$components/organisms/GraphCanvas/...`. `svelte.config.js` and `vitest.config.ts` updated with `$components` alias and `src/lib` alias removed. `npx svelte-kit sync` completed.
- Verification: `npm run test:unit` passed (52 files, 526 tests), `npm run build` passed, `cd backend && go test ./...` passed.
- Next: Phase 6/7b — E2E (Playwright) and BDD require the isolated test stack (`scripts/start-test.ps1`) and seed data, not yet run.

