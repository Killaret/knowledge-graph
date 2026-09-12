# CI-1. Циклическая зависимость в `shared`, из-за которой CI красный шесть недель

Постановка для Devin. Порядок работы: [`../AI_AGENT_PROTOCOL.md`](../AI_AGENT_PROTOCOL.md).

Ставит Claude Code, реализует Devin, проверяет Claude Code.

**Очерёдность: перед AUD-5.** Задача маленькая, но она блокирует всю джобу `Frontend Checks`, а с ней — единственный шанс увидеть остальные проверки фронтенда.

## Проблема

```
Found circular dependencies:
  shared/utils/variation.ts > shared/lib/graph/helpers.ts
```

Взаимный импорт:

- `shared/utils/variation.ts:6` берёт `darkenColor`, `lightenColor` из `$shared/lib/graph/helpers`;
- `shared/lib/graph/helpers.ts:4` берёт `applyHueShift` из `$shared/utils/variation`.

Воспроизводится локально: `node frontend/scripts/check-circular.mjs`.

## Что выяснено

**Это не следствие слияния в `main`.** Цикл присутствует в `6d0ac43`, `27adbf5` и `e06ef6b` — то есть был в `ai-agents` до вливания.

**Хронология важнее самого цикла:**

| Дата | Событие |
|---|---|
| 2026-07-18 | `b75478c` вносит импорт в `helpers.ts` — цикл появляется |
| 2026-07-25 | `be50b3a` добавляет шаг `check:circular` в CI — проверка начинает падать |
| 2026-09-07 | замечено, потому что владелец посмотрел на CI после merge |

**CI на `ai-agents` красный примерно шесть недель**, и никто не смотрел. Джоба `Frontend Checks` падает на первом же шаге, поэтому `lint`, `format:check`, `svelte-check`, юнит-тесты и проверка типизации BDD **в CI не выполнялись всё это время**. Что за ними скрыто — неизвестно до починки.

Остальные джобы того же прогона зелёные: `Backend Checks`, `Backend Integration Tests`, `Graph Service Checks`, `NLP Service Checks`, `Config & Migration Validation`, `Docker Compose Validation`.

## Как чинить

`applyHueShift` определена в `variation.ts:156`, а её родственница `applyHueShiftToRGBA` — уже в `helpers.ts:57`. То есть функция лежит не там, где её место.

Предлагаемое направление: **перенести `applyHueShift` в `helpers.ts`**, рядом с `applyHueShiftToRGBA`. Тогда `helpers.ts` становится листовым модулем цветовых примитивов, а `variation.ts` импортирует из него в одну сторону.

`applyHueShift` упоминается в пяти файлах — импорты придётся обновить. Реэкспорт из `variation.ts` ради совместимости не нужен: потребители внутренние.

Если увидишь причину сделать иначе — сделай иначе и объясни; направление здесь предложение, а не требование. Важен результат: односторонняя зависимость и осмысленное место для функции.

## Ограничения

- Не отключать и не ослаблять `check:circular` — проверка права, неправ код.
- Не трогать `services/graph-service/` — там твоя работа по AUD-5.
- Логику цветовых функций не менять, это перенос.

## Критерии приёмки

1. `node frontend/scripts/check-circular.mjs` — циклов нет.
2. `npm run lint`, `npm run format:check`, `npm run check`, `npm run test:unit` — зелёные. **Их давно не гоняли в CI**, поэтому за первым шагом могут всплыть другие отказы; если всплывут, это отдельная находка, о ней сообщить, а не чинить молча вместе с этой.
3. Джоба `Frontend Checks` на `ai-agents` проходит целиком.

## Что приложить к результату

Вывод `check-circular.mjs` до и после, и список того, что вскрылось за упавшим шагом, если вскрылось.
