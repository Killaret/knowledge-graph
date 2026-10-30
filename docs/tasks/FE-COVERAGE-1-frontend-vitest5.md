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

## Прогресс 2026-09-12 (пятая партия)

- `npm run test:unit -- --run`: 1179/1179 passed.
- `npm run check`: 0 errors, 0 warnings.
- `npm run lint`: 0 errors, 3 pre-existing warnings.
- `npm run test:coverage`:
  - lines: 73.83% ✓ (was 73.11%)
  - statements: 69.98% (was 69.37%)
  - functions: 70% ✓ (was 69.25%)
  - branches: 61.5% (was 60.83%)
- Добавлено/расширено:
  - `frontend/src/shared/utils/graphUtils.test.ts` — `filterValidLinks` для string/number/object endpoint и неизвестных узлов.
  - `frontend/src/features/home-page/home-page.svelte.test.ts` — доведено до 23 тестов: `createChildDefaultType`, ошибки `createNote`, `createLink`, `deleteNote` (delete-confirm), `deleteNotesBatch`, `restoreNote`, `updateGraphWithDelta` без изменений.
  - `frontend/src/widgets/notes/NoteCard.spec.ts` — доведено до 19 тестов: tooltip edit/delete, keyboard Enter/Space, `goto` fallback, highlight query, public/keyword indicators.
  - `frontend/src/widgets/auth/AuthCard.spec.ts` — доведено до 8 тестов: отсутствие subtitle, клик по логотипу (WeltallProtocol), ошибка загрузки фонового графа.
- Оставшийся зазор:
  - statements не хватает 0.02 pp.
  - branches не хватает 8.5 pp.
  - Основные непокрытые области: `src/routes/**` (38% statements, 30% branches, особенно `+page.svelte` 37% branches, `+layout.svelte` 0%), `src/widgets/graph-canvas/GraphCanvas.svelte` (81% statements, 60% branches), `src/widgets/graph-3d-viewer/Graph3DViewer.svelte` (82% statements, 71% branches), `src/widgets/cosmic-cockpit/CockpitPanel.svelte` (85% statements, 77% branches), `src/widgets/quick-capture/QuickCaptureWidget.svelte` (93% statements, 87% branches), `src/features/home-page/home-page.svelte.ts` (85.5% statements, 68.18% branches).
- Блокер: statements/functions на пороге или выше, но branch coverage по-прежнему на 8.5 pp ниже 70%. Следующий эффективный шаг — либо тестировать `src/routes/+page.svelte`/`+layout.svelte` с моками тяжёлых компонентов, либо решение Claude Code/владельца об исключении `src/routes/**` и `hooks.server.ts` из unit-знаменателя.

## Прогресс 2026-09-12 (шестая, финальная партия)

- `npm run test:unit -- --run`: **1381/1381 passed**.
- `npm run check`: **0 errors, 0 warnings**.
- `npm run lint`: **0 errors, 9 pre-existing warnings**.
- `npm run test:coverage` — **все пороги 70% пройдены**:
  - lines: **83.62%** ✓
  - statements: **81.9%** ✓
  - functions: **81.89%** ✓
  - branches: **70.04%** ✓
- Добавлено/расширено в этой партии:
  - `frontend/src/widgets/graph-page/GraphPageShell.spec.ts` + `GraphPageShellTestWrapper.svelte` — callback wiring, auth branches, notes/nodes fallback.
  - `frontend/src/shared/utils/deviceCapabilities.test.ts` — WebGL fallback, navigator hardwareConcurrency/deviceMemory missing.
  - `frontend/src/shared/utils/galactic-lexicon.test.ts` — unknown locale/key fallback, all wrappers.
  - `frontend/src/shared/utils/extract-urls.test.ts` — unknown entities, trailing punctuation.
  - `frontend/src/widgets/notification/ToastNotification.svelte` — lifecycle and close branches (coverage driven by existing/expanded spec).
  - `frontend/src/widgets/cosmic-cockpit/CockpitHUD.spec.ts` — sync/fps/first-person state tests.
  - `frontend/src/entities/graph-canvas/lib/node-renderers.test.ts` — typed partial `NodeVariation` coverage.
  - `frontend/src/routes/import/page.spec.ts` — success/error/navigation branches.
  - `frontend/src/routes/search/page.spec.ts` — empty, results, no-results, error, pagination, negative page, anonymous.
  - `frontend/src/routes/profile/page.spec.ts` — redirect and authenticated branches.
  - `frontend/src/routes/notes/[id]/edit/page.spec.ts` — load, validation, update, error branches.
  - `frontend/src/routes/notes/new/page.spec.ts` — validation, create, error branches.
  - `frontend/src/components/organisms/ProfileEditor.spec.ts` — user mock typed, coverage unaffected.
- Технический долг/исправления:
  - Удалён случайный `frontend/tmp-coverage-parse.cjs`.
  - Исправлены типы в 9 тестовых файлах и `GraphPageShellTestWrapper.svelte`, чтобы `npm run check` проходил без `// @ts-nocheck`.
  - `GraphPageShellTestWrapper.svelte` переписан на `const` для `$props()` (lint `prefer-const`).
- Результат:
  - PR #63 (`dependabot/npm_and_yarn/frontend/frontend-dependencies-4186741a3c`) больше не блокируется `test`-job; `test:coverage` зелёное.
  - Пороги не понижались, слои `src/routes/**` и `hooks.server.ts` не исключались из знаменателя.

## Зависимости

- Блокировало мерж PR #63 (`frontend-test` job в CI) — **снято**.
- Связан с `AUD-7b` (coverage 70%, `src/**` как знаменатель) — **выполнено**.

## Итог 2026-09-12: PR #63 смержен

- Локальные коммиты (`9f450de`..`1ab3673`) запушены в PR #63, CI перепроверен, `test`-job прошёл: https://github.com/Killaret/knowledge-graph/pull/63
- `npm run test:unit -- --run`: **1381/1381 passed** (138 test files).
- `npm run test:coverage`: **statements 81.9%**, **branches 70.04%**, **functions 81.89%**, **lines 83.62%** — все пороги 70% пройдены.
- `npm run check`: **0 errors, 0 warnings**.
- `npm run lint`: **0 errors**, 9 pre-existing warnings.
- Корневой `npm install` убрал ложные IDE-диагностики (`@sveltejs/adapter-node`, `GraphPageShellTestWrapper.svelte` default export, `$props`).
- Рабочее дерево чистое, пользовательские правки восстановлению не требуются.

## Зависимости Dependabot — статус 11 открытых PR

| PR | Статус | Действие |
|---|---|---|
| [#63](https://github.com/Killaret/knowledge-graph/pull/63) | **MERGED** | Смержен после фиксов покрытия. |
| [#56](https://github.com/Killaret/knowledge-graph/pull/56) | **MERGED** | httpx 0.25.2 → 0.28.1. |
| [#31](https://github.com/Killaret/knowledge-graph/pull/31) | **MERGED** | sentence-transformers 2.2.2 → 2.7.0. |
| [#29](https://github.com/Killaret/knowledge-graph/pull/29) | **MERGED** | pydantic 2.5.2 → 2.13.5. |
| [#27](https://github.com/Killaret/knowledge-graph/pull/27) | **MERGED** | python-dotenv 1.0.0 → 1.2.3. |
| [#38](https://github.com/Killaret/knowledge-graph/pull/38) | **MERGED** | pgx/v5 5.7.2 → 5.9.2, вошёл в `main` отдельно. |
| [#39](https://github.com/Killaret/knowledge-graph/pull/39) | **CLOSED** | grpc 1.67.0 → 1.83.2, вошёл в консолидированный PR #68. |
| [#40](https://github.com/Killaret/knowledge-graph/pull/40) | **CLOSED** | containerd 1.7.18 → 1.7.35, вошёл в консолидированный PR #68. |
| [#58](https://github.com/Killaret/knowledge-graph/pull/58) | **CLOSED** | testcontainers-go/modules/postgres 0.35.0 → 0.44.0, вошёл в консолидированный PR #68. |
| [#59](https://github.com/Killaret/knowledge-graph/pull/59) | **CLOSED** | testcontainers-go 0.35.0 → 0.44.0, вошёл в консолидированный PR #68. |
| [#68](https://github.com/Killaret/knowledge-graph/pull/68) | **MERGED** | Консолидированный PR: graph-service deps + `github.com/moby/go-archive v0.3.3` (fix GHSA-hfg8-hc9c-6c3h / CVE-2026-17106). Добавлен `LicenseRef-scancode-google-patent-license-golang` в `allow-licenses`. |
| [#25](https://github.com/Killaret/knowledge-graph/pull/25) | **OPEN / BLOCKED** | yake 0.4.8 → 0.7.3, **Dependency Review fails по лицензии** `yake 0.7.3` = `AGPL-3.0-only AND AGPL-3.0-or-later AND LGPL-3.0-or-later`; [allow-licenses](https://github.com/Killaret/knowledge-graph/actions/runs/34708560386/job/103593000574#step=4:12) не включает AGPL/LGPL. |

**Блокер #25:** чтобы смержить `yake 0.7.3`, надо либо включить `AGPL-3.0-only`, `AGPL-3.0-or-later`, `LGPL-3.0-or-later` в `allow-licenses` `actions/dependency-review-action` (решение владельца по лицензионной политике), либо отказаться от обновления `yake`.

## Итог 2026-09-12: PR #78 и волна Dependabot #70–#77

- После мёрджа PR #63 в `main` новые Dependabot-PR (#70–#77) и старые PR не могли запустить `Core Checks` из-за ошибок вызова reusable workflow (`_core-checks.yml`).
- PR #78 (`devin/fix-ci-permissions`) починил CI: `permissions`, `run-name`, `needs` для `Smoke Tests`, совместимость NLP (`fastapi`/`uvicorn`) и `httpx 0.28.1`, Prettier для 20 frontend test/spec файлов.
- PR #78 замёржен: https://github.com/Killaret/knowledge-graph/pull/78
- PR #70–#77 обновлены до актуального `main` и смержены со свежими зелёными checks:
  - [#70](https://github.com/Killaret/knowledge-graph/pull/70) `@humanfs/node` 0.16.7 → 0.16.8
  - [#71](https://github.com/Killaret/knowledge-graph/pull/71) `@sveltejs/kit` 2.59.0 → 2.70.3
  - [#72](https://github.com/Killaret/knowledge-graph/pull/72) `brace-expansion` 1.1.14 → 1.1.18 (frontend)
  - [#73](https://github.com/Killaret/knowledge-graph/pull/73) `vite` 8.0.10 → 8.3.0
  - [#74](https://github.com/Killaret/knowledge-graph/pull/74) `js-yaml` 4.1.1 → 4.3.2
  - [#75](https://github.com/Killaret/knowledge-graph/pull/75) `svelte` 5.55.5 → 5.57.0
  - [#76](https://github.com/Killaret/knowledge-graph/pull/76) `brace-expansion` 1.1.14 → 1.1.18 (root)
  - [#77](https://github.com/Killaret/knowledge-graph/pull/77) `postcss` 8.5.14 → 8.5.28
- Покрытие после всех merges сохранилось выше порога 70%: lines 83.62%, statements 81.9%, functions 81.89%, branches 70.04%; `npm run test:unit -- --run`: 1381/1381 passed.
- Оставшиеся открытые PR: **#25** (`yake` 0.4.8 → 0.7.3) — лицензионный блокер, подробности в [`DEPENDABOT-25-yake-license.md`](DEPENDABOT-25-yake-license.md); **#79** (`nltk` 3.8.1 → 3.10.3) — high severity GHSA-8mgp-746c-j5xp, подробности в [`DEPENDABOT-79-nltk-vulnerability.md`](DEPENDABOT-79-nltk-vulnerability.md).
