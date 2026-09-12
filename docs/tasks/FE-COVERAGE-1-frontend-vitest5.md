# FE-COVERAGE-1: Поднять frontend unit-coverage до 70% после обновления Vitest 5

## Контекст

- PR #63 (frontend grouped Dependabot update) разблокирован: Node обновлён до `v22.22.2`, `npm ci`, `npm run lint`, `npm run check`, `npm run test:unit` проходят.
- `npm run test:coverage` падает на порогах 70%:
  - lines: 68.27%
  - statements: 64.55%
  - functions: 63.53%
  - branches: 54.6%
- Падение связано с тем, что Vitest 5 + `@vitest/coverage-v8` 5 начали учитывать больше файлов/ветвей, чем Vitest 3.

## Что уже сделано

- `frontend/package.json`: откачены `eslint` (`^10.10.0` → `^9.39.5`), `@eslint/js` (`^10.0.1` → `^9.22.0`), `typescript` (`^7.0.2` → `^5.9.3`).
- `frontend/src/shared/api/client.ts`, `graph.ts`: адаптированы под `ky` 1.7+ (хуки принимают `state`, `prefixUrl` → `prefix`).
- `frontend/eslint.config.js`: `no-useless-assignment` отключён для `*.svelte`.
- `frontend/src/shared/stores/auth.svelte.ts`, `src/widgets/graph-canvas/GraphCanvas.svelte`, `tests/visual/visual-authenticated.spec.ts`: исправлены lint-ошибки ESLint 9.
- `frontend/vitest-setup.ts` и 5 `GraphCanvas.*.spec.ts`: `vi.fn()`-моки переписаны с arrow-функций на обычные `function`, потому что Vitest 5 не разрешает `new` для arrow-реализаций.

## Прогресс 2026-09-12

- `npm run test:unit -- --run`: 1083/1083 passed.
- `npm run check`: 0 errors, 0 warnings.
- `npm run lint`: 0 errors, 3 pre-existing warnings.
- `npm run test:coverage`:
  - lines: 69.5% (was 68.27%)
  - statements: 65.58% (was 64.55%)
  - functions: 64.52% (was 63.53%)
  - branches: 56.97% (was 54.6%)
- Добавлено/расширено:
  - `frontend/src/shared/api/graph.test.ts` — нормализация, `getGraphData`, `getFullGraphData`, `getGraphDelta`, `getCachedGraph`, `getFreshGraph`, public endpoint, обработка ошибок/401.
  - `frontend/src/shared/api/client.test.ts` — `Authorization`/`X-API-Key` заголовки, анонимный 401, параллельный refresh-токен.
  - `frontend/src/shared/utils/deviceCapabilities.test.ts` — WebGL-ветки, tiers, mobile detection.
  - `frontend/src/shared/stores/graph.svelte.test.ts` — полный обход `graphStore` (select, toggle, reset).
  - `frontend/src/entities/graph-canvas/lib/delta.test.ts` — уже существовал, покрытие подтверждено.

## Прогресс 2026-09-12 (вторая партия)

- `npm run test:unit -- --run`: 1101/1101 passed.
- `npm run check`: 0 errors, 0 warnings.
- `npm run lint`: 0 errors, 3 pre-existing warnings.
- `npm run test:coverage`:
  - lines: 69.85% (was 69.5%)
  - statements: 65.98% (was 65.58%)
  - functions: 64.95% (was 64.52%)
  - branches: 57.47% (was 56.97%)
- Добавлено/расширено:
  - `frontend/src/shared/api/import.test.ts` — bookmarklet, preview, batch create, import status.
  - `frontend/src/features/graph-ui/overlay.test.ts` — рендер overlay в разных состояниях (duplicate warning, focus mode, fog danger/recovery, undo toast, search box, help tooltip).
  - `frontend/src/features/graph-interaction/zoom-pan.test.ts` — покрытие приведено к 100% lines.
- Оставшийся зазор:
  - lines не хватает 0.15 pp.
  - statements не хватает ~4.0 pp.
  - functions не хватает ~5.1 pp.
  - branches не хватает ~12.5 pp.
  - Основные непокрытые области: `src/features/home-page/home-page.svelte.ts`, Svelte-компоненты (`CockpitPanel.svelte`, `CockpitNoteDetails.svelte`), `src/routes/**`, `src/shared/stores/auth.svelte.ts` (initAuth/login ветки), `src/shared/api/notes.ts`, `src/shared/api/links.ts`, `src/shared/api/sharing.ts`, `src/shared/utils/extract-urls.ts`.
- Блокер: без стратегии по Svelte-компонентам и/или route-файлам 70% global по functions/branches не достигается за счёт чистой TS-логики. Требуется решение Claude Code / владельца.

## Прогресс 2026-09-12 (третья партия)

- `npm run test:unit -- --run`: 1113/1113 passed.
- `npm run check`: 0 errors, 0 warnings.
- `npm run lint`: 0 errors, 3 pre-existing warnings.
- `npm run test:coverage`:
  - lines: 70.00% ✓ (was 69.85%)
  - statements: 66.1% (was 65.98%)
  - functions: 65.12% (was 64.95%)
  - branches: 57.57% (was 57.47%)
- Добавлено/расширено:
  - `frontend/src/shared/api/notes.test.ts` — `restoreNote`, `publishNote`, `unpublishNote`.
  - `frontend/src/shared/api/links.test.ts` — `updateLink`, `deleteAllNoteLinks`.
  - `frontend/src/shared/api/sharing.test.ts` — `createShareLink` с `expires_at` и `max_uses`.
  - `frontend/src/shared/stores/auth.svelte.test.ts` — `updateUserInfo` failure, `login` under SKIP_AUTH.
- Оставшийся зазор:
  - statements не хватает ~3.9 pp.
  - functions не хватает ~4.9 pp.
  - branches не хватает ~12.4 pp.
  - Основные непокрытые области: `src/features/home-page/home-page.svelte.ts`, Svelte-компоненты (`CockpitPanel.svelte`, `CockpitNoteDetails.svelte`), `src/routes/**`.
- Блокер: `lines` достиг 70% за счёт чистой TS-логики и Svelte-компонента `overlay.svelte`. Для statements/functions/branches нужно тестировать крупные Svelte/runes-модули (`home-page.svelte.ts`, `CockpitPanel.svelte`, `CockpitNoteDetails.svelte`) или исключать `src/routes/**` из unit-знаменателя.

## Цель

Довести `npm run test:coverage` до 70% по всем четырём метрикам, чтобы `test`-job в `frontend-tests.yml` проходил и PR #63 можно было мержить.

## Стратегия

1. Исключить из unit-coverage слой, который не покрывается юнитами:
   - `src/routes/**` — страницы покрываются Playwright E2E/BDD.
   - `src/hooks.server.ts` — серверный хук SvelteKit, не юнит.
2. Добавить юнит-тесты для логики с наибольшим количеством непокрытых ветвей:
   - `src/features/cosmic-cockpit/lib/panel-geometry.ts` — 32 непокрытые ветви, 0%.
   - `src/entities/graph-canvas/lib/delta.ts` — 66 непокрытых ветвей.
   - `src/shared/api/graph.ts` — 61 непокрытая ветвь.
   - `src/shared/api/client.ts` — 13 непокрытых ветвей.
   - `src/shared/utils/deviceCapabilities.ts` — 27 непокрытых ветвей.
   - `src/features/graph-3d/lib/nodes.ts` — 34 непокрытые ветви.
   - `src/features/graph-3d/lib/camera.ts` — 12 непокрытых ветвей.
   - `src/features/graph-interaction/zoom-pan.ts` — 12 непокрытых ветвей.
   - `src/features/cosmic-cockpit/model/cockpit.svelte.ts` — 14 непокрытых ветвей.
3. При необходимости — добавить/дополнить тесты для Svelte-компонентов с наибольшими пробелами (`CockpitPanel.svelte`, `GraphCanvas.svelte`, `CockpitNoteDetails.svelte`), но приоритет за чистой TS-логикой.

## Команды для проверки

```powershell
cd D:\knowledge-graph\frontend
npm run test:unit -- --run
npm run test:coverage
```

## Прогресс 2026-09-12 (четвёртая партия)

- `npm run test:unit -- --run`: 1172/1172 passed.
- `npm run check`: 0 errors, 0 warnings.
- `npm run lint`: 0 errors, 3 pre-existing warnings.
- `npm run test:coverage`:
  - lines: 73.11% ✓
  - statements: 69.37% (was 66.1%)
  - functions: 69.25% (was 65.12%)
  - branches: 60.83% (was 57.57%)
- Добавлено/расширено:
  - `frontend/src/widgets/cosmic-cockpit/CockpitNoteDetails.test.ts` — 15+ тестов на загрузку, редактирование, удаление, ссылки, время, навигацию.
  - `frontend/src/widgets/cosmic-cockpit/CockpitPanel.spec.ts` — 24 теста на collapse/expand, drag, hover, pin, keyboard, reduced-motion.
  - `frontend/src/widgets/floating-auth-panel/FloatingAuthPanel.spec.ts` — drag, close, tab, pointer, cleanup.
  - `frontend/src/widgets/quick-capture/QuickCaptureWidget.spec.ts` — расширен до 18 тестов: docked-режим, submit, пустой ввод, ошибки, Ctrl/Meta+Enter, Escape, backdrop, readonly-ответ, success.
  - `frontend/src/features/home-page/home-page.svelte.test.ts` — 16 тестов на `createHomePageState`: публичный граф, ошибки загрузки, delta-обновление, toggle layout, ошибка layout, batch-delete, child-note; + `__tests__/TestHomePageHost.svelte`.
- Оставшийся зазор:
  - statements не хватает 0.63 pp.
  - functions не хватает 0.75 pp.
  - branches не хватает 9.17 pp.
  - Основные непокрытые области: `src/routes/**` (38% statements, 30% branches), `src/features/home-page/home-page.svelte.ts` (81.5% statements, 55.5% branches), `src/widgets/notes/NoteCard.svelte` (71% statements, 56% branches), `src/widgets/quick-capture/QuickCaptureWidget.svelte` (93% statements, 87% branches, 4 оставшиеся ветки), `src/widgets/graph-canvas/GraphCanvas.svelte` (81% statements, 60% branches), `src/widgets/auth/AuthCard.svelte` (81% statements, 50% branches).
- Блокер: statements/functions на расстоянии < 1 pp от порога, но branch coverage остаётся значительно ниже 70%. Без тестирования оставшихся Svelte-компонентов и/или `src/routes/**` (или исключения route-файлов из unit-знаменателя) `npm run test:coverage` не пройдёт.

## Зависимости

- Блокирует мерж PR #63 (`frontend-test` job в CI).
- Связан с `AUD-7b` (coverage 70%, `src/**` как знаменатель).
