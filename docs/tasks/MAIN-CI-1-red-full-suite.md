# MAIN-CI-1 — красный Full Test Suite на main после мержа ai-agents

Состояние зафиксировано при доставке `ai-agents` → `main` 2026-11-15 (Devin, по поручению владельца).

## Что сделано при доставке

- PR #180 (`merge/ai-agents-2026-11-15` → `main`) влит — снапшот `ai-agents` по `a26b66c`
  (всё, прошедшее ревью-цикл; 44 коммита `ai-agents` новее — на ревью, на main не готовы).
- После мержа два тестовых разрыва чинены отдельным PR #181 (`faaa7be`):
  - `graph-service` integration-фикстура: `SELECT due_at/done_at` приехал в COMET-1 stage E,
    `CREATE TABLE notes` в тесте — нет (фикстура чинена на `ai-agents` в `2fe5c94`, сюда перенесены
    только строки схемы). `ERROR: column "due_at" does not exist` во всех `GetNotes_*` → зелёный.
  - `home-page.svelte.test.ts` «switches views and toggles layout provider»: приведён к контракту
    FREEZE-3D-1 — заглушка `isGraph3DEnabled() → true` для включённого кейса + отдельный `it`
    «выключено → "3d" читается как "graph"» (как в `graph.svelte.test.ts`, предписано ревью №8).
  - Локальные прогоны до пуша: vitest 37/37, `go test -tags=integration` 7/7.

## Текущее состояние CI на main (`faaa7be`)

Зелёное: Frontend Tests, все Core Checks (Frontend/Backend/Graph Service/NLP/Integration),
CodeQL, Dependency Graph, Config & Migration, Stacks Identity.

Красное:

1. **Production Deployment** — шаг «Publish verified images to Docker Hub»: `docker login`
   возвращает `unauthorized: incorrect username or password`. Секреты `DOCKER_USERNAME`/
   `DOCKER_PASSWORD` отклонены Docker Hub. Образы собираются и верифицируются — падение только
   на публикации. Фикс — за владельцем (креды в секретах репозитория).
2. **Full Test Suite → Run all Playwright tests** — 14 failed из 109 (`[chromium-skip-auth]`):
   - `auth-functional.spec.ts` ×2 (test_user UI, доступ без аутентификации)
   - `freeze-3d-1.spec.ts` (топ-бар без 3D-тогла, `/graph/3d` отдаёт 2D)
   - `ghost-form-types.spec.ts` (debris кликабельна при 720px)
   - `comet-1.spec.ts` (reminder, .ics, archive suggestion)
   - `note-health-1.spec.ts` (нижний HUD без декоративного блока)
   - `notes.spec.ts` «open 3D graph for a note with links»
   - `mass-import.spec.ts` (preview+import batch URLs)
   - `promises.spec.ts` ×5 (P-01, P-04, P-05/P-11, P-06, P-03/P-08)
   - `ux3-blackhole.spec.ts`
   Типичная ошибка: `expect(locator).toBeVisible()` / `toContainText` — `graph-canvas` и др.

## Важно: сьют был красным и до мержа

«Main Branch CI/CD → Full Test Suite» падал на `21694e5` 2026-09-30 — последнем коммите main
до сегодняшнего дня (все 7 прогонов того дня красные). Мерж не создал красноту: набор упавших
тестов мог измениться, но сьют не был зелёным. Отдельные провалы (`freeze-3d-1`, `notes` 3D)
похожи на протухшие под FREEZE-3D-1 спеки; `promises` ×5 и `auth` ×2 требуют разбора —
не исключены реальные регрессии из 288 коммитов.

## Стороннее наблюдение

- Security Audit на main: 23 npm-уязвимости (7 high) — зона DEPENDABOT-2/-3.
- `notes.spec.ts` 3D и `freeze-3d-1.spec` частично пересекаются с доработкой FREEZE-3D-1
  (бэклог Devin) — при доработке сверить.

## Что нужно

Триаж 14 упавших E2E на тест-стеке: для каждого — протухший контракт или дефект продукта;
дефект → отдельная строка/доработка по обычному циклу.
