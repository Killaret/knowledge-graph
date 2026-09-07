---
name: kg-graph-3d
description: Карта подсистемы 3D-графа и её ловушки — рендер, сигнал готовности, детерминированный снимок, параметры тумана.
triggers:
  - model
---

# 3D-граф: где что лежит и на чём тут спотыкаются

Использовать при любой правке `frontend/src/features/graph-3d/`, при работе с визуальным тестом 3D и при разборе жалоб «граф пустой», «ничего не меняется на снимке», «мигает».

Выведено из: `frontend/src/features/graph-3d/lib/engine.ts`, `ui/Graph3DScene.svelte`, `frontend/tests/visual/visual-regression.spec.ts`, `docs/tasks/A-1-3d-readiness-signal.md`. При их изменении скилл проверить.

## Карта

| Файл | За что отвечает |
|---|---|
| `lib/engine.ts` | Жизненный цикл: `setData`, цикл `frame()`, `renderOnce()`, `finishLoading()`, адаптация производительности |
| `lib/scene.ts` | Сборка сцены: `WebGLRenderer`, камера, `CSS2DRenderer` для подписей |
| `lib/simulation.ts` | Раскладка d3-force в трёх измерениях |
| `lib/nodes.ts`, `links.ts`, `labels.ts` | Менеджеры геометрии, связей и подписей |
| `lib/camera.ts`, `fog.ts` | Автофокус камеры и пресеты тумана |
| `ui/Graph3DScene.svelte` | Монтирование движка, признак `stableRender`, маркер `data-test-stable` |
| `widgets/graph-3d-viewer/Graph3DViewer.svelte` | Внешняя обёртка: проверка WebGL, ленивый импорт сцены, оверлей ошибки |

Параметры тумана лежат в двух местах: `config/frontend.json` (строки 20 и 37) и `knowledge-graph.config.json` (строки 151 и 168). Собранный артефакт получается из `config/`, поэтому правка только в корневом файле до приложения не доедет.

## Ловушки

**1. Рендер вызывается только внутри цикла анимации.** `renderer.render()` живёт в `renderOnce()` (`engine.ts:242`), а вызывается из `frame()` (`engine.ts:221`). Первая строка `frame()` — `if (this.disposed || this.config.disableAnimation) return;` (`engine.ts:188`).

Отсюда: при `disableAnimation: true` цикл не запускается вовсе, и без явного вызова `renderOnce()` канвас останется **пустым**. В ветке инициализации это учтено — `engine.ts:148-154` делает `simulateToStable() → updateScene() → renderOnce() → finishLoading()`. Любой новый путь, отключающий анимацию, обязан рисовать кадр сам.

**2. Готовность означает «кадр нарисован», а не «симуляция сошлась».** `finishLoading()` (`engine.ts:337`) выставляет `isReady` и дёргает `onReady`, и вызывается строго после `renderOnce()` — в обеих ветках. Порядок закреплён тестом в `lib/engine.performance.test.ts` через `invocationCallOrder`. Не переставлять.

**3. Внешняя обёртка появляется в DOM раньше сцены.** `data-testid="graph-3d-viewer"` монтируется до проверки WebGL и до импорта сцены. Ждать в тестах нужно `[data-testid="graph-3d-scene"][data-test-stable="true"]` — это контейнер самой сцены. Ожидание видимости обёртки даёт снимок экрана загрузки; именно так была сломана визуальная регрессия 3D.

**4. `setData` сбрасывает готовность движка, но не обязательно маркер в компоненте.** `engine.ts:93` ставит `isReady = false` при каждом вызове. Если компонент выставляет `data-test-stable` только по `onReady` и не возвращает его в `false` перед новым `setData`, маркер будет утверждать готовность во время перекладки. Образец правильного поведения — 2D: `widgets/graph-canvas/GraphCanvas.svelte` пишет атрибут в обе стороны.

**5. Детерминизм включается признаком `stableRender`.** `ui/Graph3DScene.svelte:39-42`: значение берётся из пропа, иначе из `process.env.VITEST`, иначе из URL-параметра `stableRender=true`. Он же выключает анимацию и фиксирует зерно генератора (`:60-61`). Признак вида `process.env.X` в браузере под Playwright не работает — `process` там не существует.

## Проверки

```powershell
cd frontend; npx vitest run src/features/graph-3d
cd frontend; npm run test:unit
```

Визуальный тест 3D — `frontend/tests/visual/visual-regression.spec.ts`, проект Playwright `visual`, только на изолированном тест-стеке. Порядок прогона и порты — скилл `kg-regression`.

Признак того, что снимок действительно измеряет сцену: изменение плотности тумана в `config/frontend.json` с пересборкой артефакта даёт видимо другой снимок. Если снимок не меняется — сначала проверяйте ловушки 1 и 3, а не графику.
