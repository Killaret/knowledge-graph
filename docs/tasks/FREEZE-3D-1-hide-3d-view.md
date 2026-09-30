# FREEZE-3D-1. 3D заморожен до готовности 2D

Статус: **бэклог — 1.0 · Devin; маленькая, после порядка владельца.** Постановка — Claude Code,
2026-09-28, по решению 82. Решение и его причина — [`RELEASE-1-scope-1.0.md`](RELEASE-1-scope-1.0.md),
раздел «3D заморожен».

## Что сделать

1. Настройка `frontend.graph.3d.enabled` в `knowledge-graph.config.json`, по умолчанию `false`, рядом
   с остальными настройками `frontend.graph.3d`.
2. При `false`:
   - в переключателе видов нет кнопки «3D»;
   - маршрут `/graph/3d` (`frontend/src/routes/graph/3d/+page.svelte`) показывает 2D-граф;
   - вид `3d`, если он где-то запомнен (`graphStore.currentView`), открывается как `graph`.
3. Код `features/graph-3d`, `widgets/graph-3d-viewer` и их тесты остаются и гоняются в CI. Ничего не
   удаляется: в мае 3D переписали и потеряли загрузку из тумана — её не держал ни один тест.
4. Документы: в `GRAPH3D.md` — пометка «заморожен до готовности 2D, решение 82»; настройка — в
   `CONFIGURATION_EN.md` и `CONFIGURATION_RU.md`.

## Критерии приёмки

1. Тест: при `enabled = false` кнопки «3D» нет, при `true` есть. Мутация «кнопка без проверки
   настройки» — красная.
2. Тест: переход на `/graph/3d` при `false` показывает 2D; запомненный вид `3d` открывается как `graph`.
3. Тесты 3D в CI выполняются и зелёные — не пропущены и не удалены.
4. Снимок верхней панели без «3D» — в `docs/agents/MANUAL_TEST_FEEDBACK.md`.

## Связи

- [`RELEASE-TEST-1-manual-run-1.0.md`](RELEASE-TEST-1-manual-run-1.0.md), пункт 5.3.
- [`UI-DESIGN-1-app-design-review.md`](UI-DESIGN-1-app-design-review.md) — история потерянной загрузки из тумана.
