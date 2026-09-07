---
name: kg-regression
description: Прогон изолированного тест-стека и полного регрессионного цикла — порядок, порты, ловушки Windows и ручного запуска Playwright, чтение отчёта.
triggers:
  - model
---

# Регрессия и тест-стек

Использовать при запуске E2E, BDD, визуальных тестов, полного цикла, а также при правках `scripts/testing/`.

Выведено из: `scripts/testing/run-full-test-cycle.ps1`, `scripts/testing/check-all.ps1`, `scripts/testing/core-checks.tsv`, `scripts/testing/lib/phase-tracking.ps1`, `.github/workflows/_core-checks.yml`, `.windsurfrules` (раздел Testing Requirements), `docs/TESTING.md`, `docs/tasks/A-3-review-findings.md`. При их изменении скилл проверить.

## Железное правило

E2E, BDD и визуальные тесты гоняются **только** на изолированном тест-стеке. Dev и personal стеки перед этим останавливаются: одновременная работа трёх стеков даёт конфликты портов и нестабильность Docker.

Personal-стек не поднимается без явной просьбы владельца.

## Порядок

```powershell
.\scripts\testing\start-test.ps1
.\scripts\testing\seed-test-data.ps1 -NoteCount 20 -LinkCount 10 -Seed 42
cd frontend; npx playwright test --project=visual
.\scripts\testing\stop-test.ps1
```

Полный цикл одной командой — `.\scripts\testing\run-full-test-cycle.ps1`. Он сам останавливает и восстанавливает стеки, снимает снимки состояния в `scripts/testing/temp/snapshots/` и печатает пофазный отчёт.

## Порты тест-стека

| Служба | Адрес |
|---|---|
| Frontend | 3002 |
| Backend | 18083 |
| PostgreSQL | 15434 |
| Redis | 16381 |
| MongoDB | 27019 |
| NLP | 15002 |
| Graph service | 19090 gRPC, 19091 HTTP |
| nginx public perimeter | 18086 |

База — `knowledge_test`, контейнеры с префиксом `kg-test-`.

**На Windows в Playwright и BDD использовать `http://127.0.0.1:3002`, не `localhost`.** Node резолвит `localhost` в `::1`, и соединение уходит в никуда. Прямой адрес бэкенда `http://127.0.0.1:18083` нужен только для настройки и health-проверок; приложение ходит через прокси `/api` на том же origin.

## Ловушки запуска Playwright вручную

Штатные `npm run test:*` подставляют окружение сами. При прямом вызове `npx playwright test` его нет, и обе ловушки срабатывают молча.

**`FRONTEND_URL` обязателен.** `playwright.config.ts` берёт `baseURL` из него, а без переменной откатывается на `http://localhost:5173` — адрес vite-dev, которого при работе через стек не существует. Симптом: `net::ERR_CONNECTION_REFUSED at http://localhost:5173/...` в тесте, который вчера проходил. Запускать так:

```bash
FRONTEND_URL=http://127.0.0.1:3002 BACKEND_URL=http://127.0.0.1:18083 npx playwright test --project=<имя>
```

**Относительные пути в конфиге разрешаются от каталога процесса, а не от каталога конфигурации.** `tests/setup/.auth/testuser.json` в `auth.setup.ts` и в проекте `visual-real-auth` — относительный. Запуск из `frontend/` пишет и читает верный файл; запуск из корня репозитория создаёт `<корень>/tests/setup/.auth/` — вне `frontend/`, вне его `.gitignore`, и проект этого файла не увидит.

Симптом коварный: прогон проходит, а тест работает со **старым** состоянием, потому что новое ушло в другое место. Проверять по времени изменения файла, а не по факту `1 passed`.

## Как читать отчёт цикла

Фазы регистрируются через общую библиотеку `scripts/testing/lib/phase-tracking.ps1` (для шелла — `phase-tracking.sh`). В отчёте три состояния: `[PASS]`, `[FAIL]`, `[SKIP]`. Пропуск не роняет цикл, любой ненулевой код выхода — роняет.

Автокоммита в цикле нет и быть не должно: он был удалён по решению владельца после того, как рапортовал об успехе при упавших тестах.

При жёсткой остановке причина печатается строкой `Error: <сообщение>` перед итоговым блоком. Если её нет, а фазы все зелёные — смотрите места, где `throw` не сопровождается регистрацией фазы.

## Правка самих скриптов

Логика фаз общая для боевого скрипта и регрессионного теста. Тест `scripts/testing/test-a3-exit-codes.ps1` подключает ту же библиотеку — не копию.

Проверять правки фазовой логики мутацией: временно испортить агрегацию в `lib/phase-tracking.ps1`, прогнать `pwsh -NoProfile -File scripts/testing/test-a3-exit-codes.ps1` и убедиться, что тест упал. Если не упал — тест не измеряет то, что должен.

Подставная упавшая фаза должна отдавать код выхода **2**, а не 1: ошибка, которую этот тест ловит, как раз в том, что проверялась только единица.

## Локальная копия Core Checks

`check-all.ps1` / `check-all.sh` повторяют пять джоб `_core-checks.yml` без управления стеками. Полный режим запускает интеграционные тесты при доступном Docker; `-Quick` / `--quick` регистрирует их как `[SKIP]`. Любой недоступный инструмент также даёт `[SKIP]` с причиной; это не выводится как полностью зелёный прогон. Соответствие команд workflow проверяет `check-core-workflow-sync.mjs` по общему `core-checks.tsv`.

```powershell
.\scripts\testing\check-all.ps1
.\scripts\testing\check-all.ps1 -Quick
```

```bash
./scripts/testing/check-all.sh
./scripts/testing/check-all.sh --quick
```

## Пирамида и команды

| Уровень | Команда |
|---|---|
| Go unit | `cd backend; go test ./...` |
| Go integration | `cd backend; go test -tags=integration ./...` |
| Frontend unit | `cd frontend; npm run test:unit` |
| Frontend покрытие | `cd frontend; npm run test:coverage` |
| E2E | `cd frontend; npx playwright test --project=chromium-skip-auth` |
| BDD | `cd frontend; npm run test:bdd` |
| Визуальные | `cd frontend; npx playwright test --project=visual` |
| NLP | `cd nlp-service; pytest tests/ -v` |

Найден дефект — регрессионный тест обязателен до закрытия задачи. Уровень выбирается по охвату, тест должен падать до правки. Норма — `.windsurfrules`, блок «Manual Found → Automated Covered».
