# E2E-CANVAS-1 — Real-auth публичный граф: топ-бар и зум канваса

## Статус

**в работе у Devin** — требуется локализация корня и регрессионные тесты.

## Что падает

Два теста в `frontend/tests/cockpit-canvas-controls.spec.ts` (проект `chromium-real-auth`, изолированный тест-стек):

1. `public graph top bar exposes canvas controls and fog toggle`
2. `canvas zoom changes transform and keeps the canvas visible`

Последний полный регрессионный прогон (`run-full-test-cycle.ps1 -SkipManual`) показал:

- `chromium-real-auth` — 2 failed из этого файла.
- Все остальные уровни (unit, integration, `chromium-skip-auth`, BDD, visual/Argos) — зелёные.

## Фиксация дефекта

| Позиция | Значение |
|---|---|
| Файл | `frontend/tests/cockpit-canvas-controls.spec.ts` |
| Тег | `@auth-real` |
| URL теста | `http://127.0.0.1:3002/graph?full=1&nocache=1` |
| Канвас | `[data-testid="graph-canvas"]` |
| Топ-бар | `[data-testid="graph-top-bar"]` |
| Контролы | `top-bar-reset`, `top-bar-open-search`, `top-bar-focus`, `top-bar-fog` |
| Контракт зума | `window.__graphCanvas.transform.k` |

Ожидания теста:

- Туман по умолчанию включён: `aria-pressed="true"`, клик переключает в `"false"`, второй клик — обратно.
- Масштаб после `wheel(0, -10)` увеличивается (`zoomedK > initialK`).
- Масштаб после `wheel(0, 40)` уменьшается (`zoomedOutK < zoomedK`).
- Канвас остаётся видимым.

## Гипотезы (по убыванию вероятности)

### 1. Режим `community` отключает интерактивность

`frontend/src/shared/stores/graph-view.svelte.ts` возвращает `graphView.mode = "community"` для анонимного пользователя. `GraphCanvas` передаёт `readonly={graphView.mode === "community"}`.

В `event-bridge.ts` `onZoom` проверяет `if (context.readonly) return;`. Поэтому на публичном графе `wheel` не меняет `transform.k`, и зум остаётся на `1.0`.

**Что проверить:**

- Должен ли публичный граф вообще быть `readonly`?
- Если должен, то тест `canvas zoom` не имеет права ожидать зума без авторизации.
- Если не должен, `readonly` не должен блокировать масштабирование (только редактирование/создание связей).

### 2. `canvasController` не успевает появиться / публичный граф пуст

Если `loadGraph` для анонима возвращает `0` узлов, `GraphCanvas` не монтируется (`graphData.nodes.length > 0`), `canvasController` остаётся `undefined`, контролы в `GraphTopBar` не рендерятся (условие `{#if canvasController}`), а `window.__graphCanvas` не определяется.

**Что проверить:**

- Логи `loadGraph` в `frontend/src/shared/services/graphLoader.ts`.
- Seed на тест-стеке: создаёт ли `seed-test-data.ps1` публичные заметки в `real-auth` режиме?
- Endpoint публичного графа (`/graph/public` или `/graph?full=1`) не требует JWT для анонима.

### 3. Туман по умолчанию выключен в `community`

Хотя `config/frontend.json` и `knowledge-graph.config.json` задают `fog.enabled: true`, путь к настоящему `graphConfig2D` может брать `fog.enabled` из `knowledge-graph.config.json`, который может быть переопределён переменной среды или `build-config`.

**Что проверить:**

- Распечатать `graphConfig2D.fog.enabled` в dev-консоли на `http://127.0.0.1:3002/graph?full=1&nocache=1`.
- Проверить, что `createFogState()` инициализирует `enabled` именно из конфига, а не из `false` по умолчанию.

### 4. Проблема тайминга / `data-test-stable`

Канвас может быть видим, но `GraphCanvas.onMount` ещё не выставил `window.__graphCanvas` и не подключил `attachEvents`. Тест ждёт `__graphCanvas`, но не ждёт `data-test-stable="true"`.

**Что проверить:**

- Добавить ожидание `canvas.locator('[data-test-stable="true"]')` перед `wheel`.
- Проверить порядок: `onMount` → `window.__graphCanvas = { ... }` → `attachEvents(...)`.

### 5. Зум чувствителен к минимальному `deltaY`

`frontend/src/features/graph-interaction/zoom-pan.ts`: `const newScale = Math.min(Math.max(transform.k * (1 + -e.deltaY * 0.001), 0.1), 5);`

Для `deltaY = -10` новый масштаб `1.01`; для `deltaY = 40` — `0.96`. Этого достаточно, но если событие не достигает `handleZoom` (см. гипотезу 1), масштаб не меняется.

### 6. Условный рендеринг `GraphTopBar` в `community`

`GraphTopBar.svelte` скрывает всю группу canvas-контролов под `{#if canvasController}`. Если `canvasController` не пробрасывается в публичный граф (например, `GraphPageShell` не получает `bind:controller` из `GraphCanvas` из-за `{#key}` или `loading`), контролы не появятся.

**Что проверить:**

- В dev-режиме убедиться, что `[data-testid="top-bar-fog"]` присутствует в DOM анонимного `/graph`.

### 7. Playwright-специфика: `page.mouse.wheel` и пассивные слушатели

`attachEvents` регистрирует `wheel` с `{ passive: false }` и вызывает `e.preventDefault()`. Playwright `page.mouse.wheel` генерирует событие, которое может не пройти, если canvas не в фокусе или если `pointer-events`/`touch-action` перехватывают.

**Что проверить:**

- Вручную зайти через Playwright trace и посмотреть, дошло ли `wheel` до `<canvas>`.
- Попробовать `page.mouse.click` в центр canvas перед `wheel`.

## План расследования

1. **Сначала убедиться, что дефект воспроизводится локально.**
   - Поднять изолированный тест-стек.
   - Запустить только этот файл:
     ```powershell
     $env:FRONTEND_URL='http://127.0.0.1:3002'
     $env:BACKEND_URL='http://127.0.0.1:18083'
     cd frontend
     npx playwright test --project=chromium-real-auth tests/cockpit-canvas-controls.spec.ts --reporter=line
     ```
   - Собрать trace и логи консоли.

2. **Определить, пустой ли публичный граф в `real-auth`.**
   - В Playwright `page.evaluate` вызвать `fetch('/api/graph?full=1').then(r => r.json())` анонимно.
   - Сравнить с `SKIP_AUTH`.

3. **Проверить `readonly` как корень зума.**
   - В dev-инструментах открыть `/graph?full=1&nocache=1` без авторизации.
   - Проверить `graphView.mode`, `readonly`, наличие `wheel`-слушателя и изменение `window.__graphCanvas.transform.k`.

4. **Проверить `canvasController` и `fogEnabled`.**
   - Вывести в dev-консоль `canvasController` из `GraphTopBar` или `GraphPageShell`.
   - Проверить, что `aria-pressed` действительно `"true"` до первого клика.

5. **Установить настоящую причину и записать в review-findings.**

## Критерии приёмки

- [ ] Корень двух падений установлен и задокументирован.
- [ ] Исправление либо в production-коде, либо в тесте (если ожидание теста неправильно).
- [ ] Добавлен регрессионный тест, который падает до исправления и проходит после.
- [ ] Adversarial-тесты на граничные случаи: пустой граф, отключённый туман, `readonly`, быстрый `wheel`.
- [ ] Полный `run-full-test-cycle.ps1 -SkipManual` зелёный по E2E real-auth.
- [ ] `docs/AI_HANDOFF.md` и `docs/AI_LOG.md` обновлены.

## Связанные файлы

- `frontend/tests/cockpit-canvas-controls.spec.ts`
- `frontend/src/routes/graph/+page.svelte`
- `frontend/src/widgets/graph-canvas/GraphCanvas.svelte`
- `frontend/src/features/graph-interaction/event-bridge.ts`
- `frontend/src/features/graph-interaction/zoom-pan.ts`
- `frontend/src/features/graph-canvas/fog-state.svelte.ts`
- `frontend/src/shared/stores/graph-view.svelte.ts`
- `frontend/src/widgets/graph-page/GraphPageShell.svelte`
- `frontend/src/features/graph-ui/GraphTopBar.svelte`
- `frontend/src/shared/services/graphLoader.ts`
- `scripts/testing/seed-test-data.ps1`

## Назначение

- **Исполнитель:** Devin (расследование + исправление + тесты).
- **Ревью:** Claude Code.
