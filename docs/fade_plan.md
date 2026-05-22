# fade_plan — Восстановление эффекта «Завесы Тумана» для 2D GraphCanvas

TL;DR — добавить opacity для узлов и связей в `GraphCanvas`, запускать анимацию 0→1 через `requestAnimationFrame`, связав прогресс с долей стабилизированных узлов (как в 3D). Добавить тест и убедиться, что существующие тесты проходят.

## Краткие шаги

1. Добавить `Map` для хранения opacity: `nodeOpacity = new Map()` и `linkOpacity = new Map()`, инициализировать при загрузке данных.

2. В `startSimulation` (`frontend/src/lib/components/GraphCanvas/simulation.ts`) при старте:
   - Запустить RAF-loop и подписаться на `simulation.on('tick')`.
   - Вычислять долю стабилизированных узлов → `progress`.
   - Вычислять `targetOpacity = ease(progress)` и интерполировать значения в `nodeOpacity`/`linkOpacity` к `targetOpacity`.
   - При завершении симуляции выполнить добивку до `1` за 2400 ms.

3. В рендерере (`frontend/src/lib/components/GraphCanvas/renderer.ts`) учитывать opacity из `Map` при рисовании (`ctx.globalAlpha` или `rgba(...)`) и вызывать перерисовку при изменениях.

4. Управлять RAF: хранить id и отменять через `cancelAnimationFrame` при остановке/анмаунте.

5. Тесты: добавить `GraphCanvas.fade.spec.ts` — мокать симуляцию, эмулировать тики, проверять рост opacity и финальную добивку до ~1.

6. Валидация: запустить тесты и сделать визуальную проверку.

## Relevant files

- `frontend/src/lib/components/GraphCanvas.svelte`
- `frontend/src/lib/components/GraphCanvas/simulation.ts`
- `frontend/src/lib/components/GraphCanvas/renderer.ts`
- `frontend/src/lib/components/Graph3D/fogManager.ts` (референс)
- `frontend/src/lib/components/GraphCanvas.rendering.spec.ts`
- `frontend/src/lib/components/GraphCanvas.fade.spec.ts`

## Status

- ✅ Реализовано: 2D fog veil effect теперь плавно прогрессирует от `0` до `1` вместе со стабилизацией узлов.
- ✅ Тесты: добавлен `frontend/src/lib/components/GraphCanvas.fade.spec.ts`, утверждающий начальную нулевую непрозрачность, прогрессию и финальное приближение к `1`.
- ✅ Валидация: целевой тестовый набор `GraphCanvas` прошёл успешно.
