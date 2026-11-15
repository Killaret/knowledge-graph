# DEPENDABOT-2 — разбор исполнения

Запись — Devin, 2026-09-30. Постановка: `DEPENDABOT-2-open-prs-2026-09.md`.
Проверка локальная: каждый безопасный PR смержен во временную ветку от `ai-agents`
(`b59a870`), прогнан затронутый сервис, ветки удалены. Рабочее дерево после прогонов
чистое. CI на самих PR красный, но прогоны идут на старом общем состоянии репозитория
(Config & Migration Validation, Core Checks) — не относятся к изменениям зависимостей,
поэтому вердикт строится на локальных прогонах (постановка допускает это при сомнении).

## Вердикты по безопасным PR

| PR | Изменение | Проверка | Вердикт |
|---|---|---|---|
| #115 | `golang.org/x/text` 0.41→0.42, backend | `go test ./...` в `backend/` — зелёный, кроме `TestNLP4Config` (находка 1, не от PR) | можно сливать |
| #117 | `golang.org/x/net` 0.58→0.59, backend | `go test ./...` в `backend/` — зелёный | можно сливать |
| #118 | `google.golang.org/grpc` 1.83.2→1.84.0, graph-service | `go test ./...` в `services/graph-service/` — зелёный (все пакеты, `proto` без тестов) | можно сливать |
| #119 | `go.mongodb.org/mongo-driver` 1.17.9→1.17.10, backend | `go test ./...` в `backend/` — зелёный | можно сливать |
| #120 | `golang.org/x/crypto` 0.55→0.57, backend | прогон auth/web/middleware пакетов — зелёный | можно сливать |
| #121 | `github.com/jackc/pgx/v5` 5.10→5.11, backend | прогон db/events/handlers/application/api — зелёный | можно сливать |
| #124 | `actions/checkout` v5→v7 | дифф — только теги версий в `.github/workflows/`; конфликтов с `ai-agents` нет | можно сливать |
| #125 | `actions/setup-go` v6→v7 | то же | можно сливать |
| #129 | корень: `jsdom` 30.0.1→30.1.1, `tsx` 4.23.13→4.23.15, `thesvg` 3.3.4→3.3.8 | `npm install` чисто; `frontend` lint 0 ошибок, `svelte-check` 0, полный `vitest` — только находка 2 | можно сливать |

**Список «можно сливать» для владельца:** #115, #117, #118, #119, #120, #121, #124, #125, #129 —
все девять. Конфликтов с `ai-agents` не выявлено: все ветки смержились локально без конфликтов.
Слияние — за владельцем, PR идут в `main`.

## «После 1.0» — не закрывать

#128 (`sentence-transformers` 2.7→6.1), #126 (`huggingface-hub` 0.23→2.0), #96 (`fastapi` 0.115→0.141),
#127 (`uvicorn` 0.34→0.54), #130 (группа фронтенда с TypeScript 7 и ESLint 10) ждут 1.0 —
по постановке не закрываются, Dependabot пересоздаст при необходимости.

## Находки

1. **`TestNLP4Config` ждал старый дефолт.** Тест утверждал «pipeline must default to off»,
   MODEL-2 включил конвейер по умолчанию (`nlp.pipeline.enabled: true` в
   `knowledge-graph.config.json:282`) — тест краснел на любом прогоне `backend`, включая
   чистый `ai-agents`. Не связан с зависимостями. Исправлено ожидание и комментарий
   (`backend/internal/config/config_test.go`).

2. **`event-bridge.test.ts`: два теста link-hover красные после UX-3.** Наведение на
   (25,12) попадало в радиус хита 30px обоих концов связи (узлы на 50px друг от друга),
   и `findLinkAtPosition` правильно отдавал узлу приоритет — связь не подсвечивалась.
   Узел `n2` разнесён с (50,10) на (200,10), наведение на связь ведётся в (100,10) —
   вне обоих радиусов, на сегменте. 26/26 зелёные
   (`frontend/src/features/graph-interaction/event-bridge.test.ts`).

3. **У корня нет скрипта `lint:fix`.** Постановочная проверка «локальный прогон» для #129
   упиралась в `npm run lint:fix` в корне — скрипта нет (`Missing script`). Линт живёт в
   `frontend/`; прогнан отдельно — 0 ошибок, 9 предупреждений в чужих файлах. Дефект
   сборочной обвязки, не PR; записан сюда как исключение, критерий не правится.

4. **`npm audit` на корневых dev-deps:** после установки #129 отчёт показывает 4
   уязвимости (3 low, 1 high). Информационно для владельца — зависимости инструментов
   разработки, не блокер слияния, но стоит посмотреть при удобном случае.

## Слияние в main (2026-09-30, Devin — по просьбе владельца)

Все девять PR обработаны:

| PR | Итог |
|---|---|
| #115 x/text, #117 x/net, #118 grpc, #119 mongo, #121 pgx, #124 checkout, #125 setup-go, #129 root-deps | **смержены** в `main` (merge commit) |
| #120 x/crypto | Dependabot закрыл сам: x/crypto 0.57 подтянулся транзитивно через x/net 0.59 |

### Выход из строя main и починка

x/text 0.42, x/net 0.59, x/crypto 0.57 декларируют `go 1.26.0` и подняли `go`
в `backend/go.mod` до 1.26.0; CI держит `go-version: '1.25'` + `GOTOOLCHAIN: local` →
все Go-джобы на main красные. Локально проверка не ловила — toolchain 1.26
скачивался автоматически.

Фикс: PR #133 — `go 1.25.0`, x/crypto 0.55.0, x/net 0.58.0, x/text 0.41.0 +
транзитивные x/mod 0.38, x/sync 0.22, x/sys 0.47, x/tools 0.48;
pgx 5.11.0 и mongo-driver 1.17.10 сохранены (совместимы с Go 1.25).

Дополнительно PR #134 — регенерация `docs/tasks/README.md` на main
(дрейф даты LINKS-2, красный «Check task index»).

### Итог CI на main после #134

Core Checks зелёные целиком (frontend, backend, integration, graph-service,
NLP, migration drift, config, stacks identity). Осталось два **флага, не от
мержей**:

- `Full Test Suite`: 2 красных `auth-functional.spec.ts` (SKIP_AUTH) — те же
  падали 24.09 (run 36011499757), до всех мержей.
- `Production Deployment`: `container kg-nlp is unhealthy` на старте deploy-стека —
  красный с ~23.09 (30 ранов подряд), до мержей.

### Рекомендация владельцу

Пока проект на Go 1.25 — добавить в `dependabot.yml` `ignore` для
`golang.org/x/*` (или поднимать тулчейн до 1.26 отдельной задачей: Dockerfile,
CI `go-version`, `.windsurfrules`, локальные установки). Иначе Dependabot
пришлёт те же бампы заново.
