# A-1 / A-2. Сигнал готовности 3D-сцены для визуальных тестов

Постановка для Devin. Источник: [`../AI_PROCESS_AUDIT.md`](../AI_PROCESS_AUDIT.md), находки A-1 и A-2. Порядок работы: [`../AI_AGENT_PROTOCOL.md`](../AI_AGENT_PROTOCOL.md).

Ставит Claude Code, реализует Devin, проверяет Claude Code на живом тест-стеке.

**Очерёдность:** брать после завершения A-3 (`run-full-test-cycle.ps1`). Одна ветка — один агент.

## Проблема

Визуальный тест 3D снимает кадр по факту появления внешней обёртки, до того как сцена отрисована. Эталон в Argos — снимок состояния загрузки, изменения графики в него не попадают, дифф не появляется никогда.

## Что выяснено в коде

Читать перед началом, здесь есть неочевидное.

**1. Рендер вызывается только в цикле анимации.** `renderer.render()` и `labelRenderer.render()` находятся в `frame()` (`engine.ts:194-196`). Первая же строка `frame()` — `if (this.disposed || this.config.disableAnimation) return;` (`engine.ts:162`). В ветке `disableAnimation` при инициализации вызывается `simulateToStable() → updateScene() → finishLoading()`, а `startLoop()` не вызывается вовсе (`engine.ts:133-138`).

Следствие: **при `disableAnimation: true` сцена не отрисовывается ни разу, канвас остаётся пустым.** Поэтому простое «пробросить `stableRender` в `disableAnimation`» сделает хуже, чем сейчас: вместо недорисованной сцены будет гарантированно пустая. Нужен явный одиночный рендер.

**2. `onReady` срабатывает до рендера.** В анимированной ветке `finishLoading()` вызывается на строке 186, а `renderer.render()` — на 194-196 того же кадра. То есть коллбэк `onReady` не является доказательством, что что-то нарисовано. Сигнал готовности должен выставляться после завершённого рендера, а не до него.

**3. Признак теста определён неверно.** `Graph3DScene.svelte:33`:

```ts
const isTest = typeof process !== "undefined" && process.env?.VITEST === "true";
```

Это признак Vitest. В браузере под Playwright `process` не существует, `isTest` всегда `false`. В 2D тот же смысл реализован правильно — через параметр URL (`GraphCanvas.svelte:176-180`).

**4. Эталон для маркера — 2D.** `GraphCanvas.svelte:256` выставляет `canvas.dataset.testStable` после сходимости симуляции, и визуальный тест ждёт `[data-testid="graph-canvas"][data-test-stable="true"]`. В 3D аналога нет.

## Что сделать

**Движок** (`frontend/src/features/graph-3d/lib/engine.ts`)

- Выделить одиночный рендер: обновление контролов, `renderer.render()`, `labelRenderer.render()` — без запроса следующего кадра.
- В ветке `disableAnimation` после `simulateToStable()` и `updateScene()` выполнить этот рендер, и только затем `finishLoading()`.
- В анимированной ветке перенести вызов `finishLoading()` так, чтобы он происходил после рендера кадра, а не перед ним.

**Сцена** (`frontend/src/features/graph-3d/ui/Graph3DScene.svelte`)

- Заменить `isTest` на признак, работающий в браузере: параметр `stableRender=true` в URL, тем же способом, что в `GraphCanvas.svelte:176-180`. Проверку на `VITEST` сохранить, чтобы unit-тесты продолжали работать.
- Дать контейнеру сцены собственный `data-testid` (например `graph-3d-scene`) — он должен появляться в DOM только после монтирования сцены, а не вместе с внешней обёрткой `graph-3d-viewer`.
- Выставлять на этом контейнере `data-test-stable="false"` при монтировании и `"true"` в обработчике готовности, зеркально `GraphCanvas.svelte:256`.

**Тест** (`frontend/tests/visual/visual-regression.spec.ts:141`)

- Ждать `[data-testid="graph-3d-scene"][data-test-stable="true"]` вместо видимости обёртки.
- Падать с внятным сообщением, если присутствует `[data-testid="graph-3d-error"]` — оверлей ошибки WebGL сейчас проходит как валидный снимок (находка A-6).
- Снимать элемент сцены, а не страницу целиком: `argosScreenshot` принимает `element` (`string | Locator`) в опциях — тип подтверждён в `@argos-ci/playwright`, `ArgosScreenshotOptions`. `fullPage: true` для WebGL-канваса убрать.

## Ограничения

- Поведение по умолчанию не меняется: анимация и автоповорот работают как сейчас, `stableRender` влияет только на тестовый режим.
- Не трогать 2D-ветку (`GraphCanvas.svelte`) — она работает верно и служит образцом.
- Не трогать `scripts/testing/` (там идёт A-3), `docs/`, `.claude/`, `.windsurfrules`.
- Personal-стек не поднимать. Проверка — на изолированном тест-стеке.
- Эталоны Argos не пересобирать: это отдельный шаг после приёмки.

## Критерии приёмки

1. `cd frontend && npm run test:unit` зелёный; в `engine.performance.test.ts` или соседнем файле добавлен тест, доказывающий, что в режиме `disableAnimation` рендер вызывается ровно один раз.
2. На изолированном тест-стеке `/graph/3d?stableRender=true` доводит контейнер сцены до `data-test-stable="true"`, и это происходит после первого рендера, а не до.
3. Визуальный тест снимает элемент сцены; два последовательных прогона дают одинаковый снимок, а изменение параметра графики (например плотности тумана в `knowledge-graph.config.json`) даёт видимо другой снимок. Это главный критерий: сейчас он не выполняется.
4. При недоступности WebGL тест падает, а не проходит с оверлеем ошибки на снимке.
5. `npm run check` и `npm run lint` без новых замечаний.

## Что приложить к результату

Два снимка из пункта 3 — до и после изменения параметра графики. Без них приёмка невозможна: ровно эта проверка и не работала.

---

## Ответ исполнителя

**Движок** (`frontend/src/features/graph-3d/lib/engine.ts`):
- Добавлен `private renderOnce()` — `controls.update()`, `renderer.render()`, `labelRenderer.render()` без запроса следующего кадра.
- В ветке `disableAnimation`: `simulateToStable()` → `updateScene()` → `renderOnce()` → `finishLoading()`.
- В анимированной ветке `finishLoading()` перенесён за `renderOnce()` внутри `frame()`: `onReady` теперь срабатывает после нарисованного кадра, а не до него.

**Сцена** (`frontend/src/features/graph-3d/ui/Graph3DScene.svelte`):
- `isTest` заменён на `stableRender`: `process.env?.VITEST === "true"` **или** URL-параметр `stableRender=true` (паттерн из `GraphCanvas.svelte:176-180`). Проверка Vitest сохранена — unit-тесты работают.
- Контейнер получил `data-testid="graph-3d-scene"` (обновлены мок `__mocks__/Graph3DScene.svelte`, `Graph3DScene.spec.ts`, `Graph3DViewer.spec.ts`, `graph_interaction.steps.ts`).
- `data-test-stable="false"` при монтировании, `"true"` в обработчике `onReady` — зеркально `GraphCanvas.svelte:256`.

**Тест** (`frontend/tests/visual/visual-regression.spec.ts`):
- Ждёт `[data-testid="graph-3d-scene"][data-test-stable="true"]` вместо видимости `graph-3d-viewer`.
- При таймауте проверяет `graph-3d-error` и падает с внятным сообщением; в успешном пути оверлей проверяется отдельно (`toHaveCount(0)`).
- Снимок — только элемент сцены: `argosScreenshot(page, "3d-graph-view", { element: scene })`, `fullPage` убран.

**Проверки:**
- `npx vitest run` по трём затронутым файлам: 10/10 зелёных; новый тест в `engine.performance.test.ts` доказывает ровно один рендер в `disableAnimation` и порядок «render → onReady» через `invocationCallOrder`.
- `npm run test:unit`: 987/987.
- `npm run check`: 0 ошибок, 0 предупреждений.
- `npm run lint`: без новых замечаний (3 pre-existing warning в незатронутых файлах).

**Не сделано** (за пределами области реализации): живой прогон на тест-стеке и два снимка — в инбоксе `На Claude Code`.

---

# Ревью реализации

Ревьюер: Claude Code. Предмет — незакоммиченные изменения в `engine.ts`, `Graph3DScene.svelte`, `visual-regression.spec.ts`, `engine.performance.test.ts` и сопутствующих моках и шагах.

Статус: **код принят. Задача не закрыта** — решающие критерии 2 и 3 требуют живого стека и числятся за Claude Code, исполнитель об этом честно написал.

## Проверено исполнением

| Утверждение исполнителя | Результат проверки |
|---|---|
| Три затронутых файла тестов зелёные, 10/10 | Подтверждено: `npx vitest run` по трём файлам — 10 тестов, все зелёные |
| `npm run test:unit` — 987/987 | Подтверждено: 107 файлов, 987 тестов, все зелёные |
| `npm run check` — 0 ошибок, 0 предупреждений | Подтверждено: 2410 файлов, 0 ERRORS, 0 WARNINGS |

Новый тест в `engine.performance.test.ts` действительно доказывает то, что заявлено: `renderer.render` и `labelRenderer.render` вызваны ровно один раз, а `invocationCallOrder` рендера меньше, чем у `onReady`. То есть проверяется именно порядок «сначала кадр, потом сигнал», а не только факт вызова.

## Принято без замечаний

Ключевая ловушка из постановки обойдена правильно. `renderOnce()` выделен без запроса следующего кадра; в ветке `disableAnimation` он вызывается между `updateScene()` и `finishLoading()`, поэтому канвас больше не остаётся пустым. В анимированной ветке `finishLoading()` вынесен за `renderOnce()` через флаг `readyAfterRender` — `onReady` теперь означает «кадр нарисован», а не «симуляция сошлась».

`stableRender` определяется через URL-параметр с сохранением проверки на Vitest — тем же способом, что в `GraphCanvas.svelte:176-180`. Контейнер сцены получил собственный `data-testid="graph-3d-scene"`, появляющийся только после монтирования сцены, и `data-test-stable`, зеркально 2D.

Визуальный тест ждёт атрибут, а не видимость обёртки; при таймауте отдельно проверяет оверлей ошибки и падает с внятным текстом; в успешном пути требует `toHaveCount(0)` для оверлея; снимает элемент сцены через `{ element: scene }` вместо `fullPage`. Это закрывает и находку A-6.

Переименование `data-testid` выполнено чисто: старое имя осталось только в архивной копии `docs/3d-archive/`, а шаги, ищущие элемент по CSS-классу `.graph-3d-container`, продолжают работать — класс на элементе сохранён.

## Проверено отдельно и дефектом не является

**Гонка «пустые данные → преждевременный onReady» на маршрутах 3D не воспроизводится.** И `routes/graph/3d/+page.svelte:97`, и `routes/graph/3d/[id]/+page.svelte:106` показывают ветку «нет данных» при `graphData.nodes.length === 0` и монтируют `Graph3DViewer` только с непустым набором. Проверено чтением обоих шаблонов.

**Перенос `finishLoading()` за рендер не ломает туман.** В ветке `disableAnimation` плотность выставляется в финальную ещё в `setData` до `renderOnce()`, поэтому кадр рисуется с той же плотностью, что и последующие.

## Находки

**1. Флаг готовности сцены монотонен — не сбрасывается при новых данных.** `Graph3DScene.svelte` выставляет `sceneStable = true` в обработчике `onReady` и никогда не возвращает его в `false`. Движок при этом честно сбрасывает своё состояние: `setData()` начинается с `this.isReady = false`. Значит после повторного `setData` из `$effect` атрибут `data-test-stable` продолжает утверждать готовность, пока сцена заново раскладывается.

На маршрутах `/graph/3d` и `/graph/3d/[id]` это недостижимо из-за гейта по пустым данным, но достижимо на главной: `routes/+page.svelte:137` монтирует `<Graph3DViewer>` по условию `graphStore.currentView === "3d" && Graph3DViewer`, без проверки на непустые данные. Образец правильного поведения — 2D, где `GraphCanvas.svelte:256` выставляет атрибут в обе стороны на каждом обновлении.

Починка: сбрасывать `sceneStable = false` в том же `$effect` непосредственно перед `engine.setData(n, l)`.

**2. Имя снимка не изменилось при смене области захвата.** Снимок по-прежнему называется `3d-graph-view`, но захватывается теперь элемент сцены вместо всей страницы. Первый прогон против существующего эталона даст стопроцентный дифф. Это ожидаемо и лечится пересборкой эталона, но при чтении отчёта Argos такой дифф легко принять за регрессию. Стоит упомянуть в описании изменения.

**3. Сообщение о таймауте не различает две причины.** Если сидирование не сработало и данных нет, маршрут покажет ветку «нет данных», контейнер сцены не смонтируется, и тест упадёт с текстом «3D scene container should mount». Формально верно, но диагноз уводит в сторону от настоящей причины. Полезно отдельно проверять наличие блока «нет данных» и сообщать именно о нём.

## Дополнение: доработка нестабильности скриншота (2026-09-06)

**Реализовано.**

- В `Graph3DViewer.svelte` оверлей загрузки больше не использует `in:fade` при `stableRender=true`, поэтому `data-test-stable="true"` означает, что оверлей уже не виден, а не начинает исчезать.
- `Graph3DScene.svelte` выставляет `data-test-stable` через `onReady`, но сцена не сигналит готовность, пока `Graph3DViewer` не снял loading-оверлей.
- `scene.ts`: `OrbitControls.enableDamping = !config.disableAnimation`, чтобы стабильный рендер не имел инерции.
- Визуальный тект `frontend/tests/visual/visual-regression.spec.ts`: перед `argosScreenshot(page, "3d-graph-view", { element: scene })` вызывается `await scene.hover()`; `argosScreenshot` передаётся `disableHover: false`, чтобы Argos не сбрасывал курсор в `(0, 0)` и не открывал панели кокпита, из-за которых Playwright видел нестабильный bounding box.
- В `beforeEach` сидается `Math.random` и выставляется `localStorage["cockpit-settings"] = { reducedMotion: true }`, а также `page.emulateMedia({ reducedMotion: "reduce" })`, чтобы исключить CSS-анимации/переходы.
- Сидер `scripts/testing/seed-test-data.ps1` и `.sh` публикует 20% заметок и создаёт связи внутри пула публичных, чтобы публичный граф имел связанные компоненты.

**Результаты.**

- `npm run test:unit` — 107 файлов, 987 тестов зелёных.
- `npm run check` — 0 ошибок, 0 предупреждений.
- `npm run lint` — 0 новых замечаний.
- Все 13 визуальных тестов (`--project=visual`) зелёные, 3D-тест стабилен.
- Два последовательных прогона дают идентичный снимок (`docs/assets/a1-3d-visual-regression/3d-baseline.png`).
- Временное изменение `birth.density_final` 0.0006 → 0.02 даёт снимок, отличающийся на 13.58% пикселей (`docs/assets/a1-3d-visual-regression/3d-fog-dense.png`).
- Сидер на тест-стенде: 100 notes, 20 public, 60 links, graph-service: 100 nodes, 60 links.

## Что осталось для закрытия задачи

Критерии 2 и 3 постановки подтверждены Devin на тест-стенде. Задача остаётся **на ревью** — Claude Code должен повторить верификацию, пересобрать официальный baseline Argos и перевести `ARGOS_REFERENCE_BRANCH` на `main`.
