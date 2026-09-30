# CONFIG-AUDIT-1 — разбор исполнения (Devin, 2026-09-30)

## Реестр и сторож

- `docs/operations/CONFIG_REGISTRY.md` — все 189 ключей `config/*.json`: читатель (файл/поле),
  env-переопределение, compose-переопределение, наличие в CONFIGURATION_EN/RU, статус.
- Генератор `scripts/testing/generate-config-registry.mjs` разбирает `JSONConfig` по json-тегам
  и сопоставляет акцессоры `j.<Path>` в `resolveConfig`/`Load` с env-переменными; фронтенд-ключи
  — по фактическому использованию в `frontend/src` (включая алиасы `graphConfig2D` и перепривязки
  вида `this.performanceConfig.*`).
- Сторож `scripts/testing/check-config-registry.mjs`: дрейф файла относительно генератора,
  полнота (каждый ключ ↔ строка), у живых строк обязателен читатель. Подключён в `_core-checks.yml`.

## Порядок приоритета (подтверждён тестами)

| Сервис | Задумано | Тест |
|---|---|---|
| backend | env > JSON > дефолт | `TestEnvVarPriority` (уже был) |
| graph-service | env > JSON > дефолт | `TestFullLimitEnvOverridesJSONFile`, `TestFullLimitFromJSONWhenNoEnv`, `TestLimitHelperPrecedence` |
| nlp | env > дефолт | `test_env_precedence_over_defaults` (новый) + `test_hf_offline_enabled` |
| frontend | только `config/*.json`, вшито при сборке | `config.test.ts`: `config.frontend/ci_cd` зеркалят `config/frontend.json`/`ci_cd.json` |

Мутации: `getIntEnv` возвращает константу вместо env → `TestLimitHelperPrecedence` красный;
`resolveLimit` считает любой `limit` каноническим → `TestGetFullGraphHandler_LimitDoesNotPoisonCache` красный.

## Находки из постановки

1. `backend.pagination.max_limit=300` молча резал список и граф — **исправлено в NOTES-LIMIT-1**
   (`getNotes` читает страницы до `total`, `filterGraphData` без фильтров не теряет узлы).
2. Предел полного графа: `config/graph_service.json`=500, compose `GRAPH_FULL_LIMIT`=500, код=1000 —
   **дефолт в коде приведён к 500** (`config.go:97`). Корень расхождения «`limit=100` вернул 1000 узлов»:
   при попадании в кэш лейаут отдавался целиком, а `?limit=` и `nocache` на кэш не действовали;
   хуже — урезанный результат **перезаписывал** канонический кэш и снапшот дельт.
   Исправлено в трёх обработчиках (`GetFullGraphHandler`, `GetPublicGraphHandler`, gRPC
   `GetFullLayout` через `resolveLimit`/аналог): канонический размер = сконфигурированный предел;
   `limit` ниже предела — свежий ответ без чтения/записи кэша; выше предела — клэмп; `limit=0`
   больше не обходит предел. Тест `LimitDoesNotPoisonCache` покрывает все четыре случая.
3. `frontend.graph.2d.max_nodes` — **мёртвый ключ**, в списке владельцу ниже.
4. `frontend.api.default_limit=100` — уходил в `getFullGraphData` как `limit=100` и до правки
   резал и отравлял кэш. Дефолт изменён на `0` («решает сервер»); сам ключ стал мёртвым.

## Мёртвые ключи — решение владельца (подключить или убрать)

`backend.graph.max_nodes`, `backup.draft_ttl_hours`, `ci_cd.integration_test.migrate_all`,
`ci_cd.integration_test.truncate_list`, `frontend.language`, `frontend.test.debounce_timeout_ms`,
`frontend.test.max_retry_count`, `frontend.test.mock_goto_delay_ms`, `frontend.graph.2d.max_nodes`,
`frontend.api.default_limit`, `frontend.api.link_limit`, `nlp.model_name`, `nlp.max_text_length`,
`nlp.hf_home`, `nlp.hf_hub_disable_telemetry`, `nlp.hf_hub_offline` (nlp.json дублирует env —
python-сервис читает только окружение).

Замечено попутно (не исправлено — зафиксировано): `graph-service` при отсутствии ключа в JSON
получает ноль, а не дефолт (нет pre-seed как у backend `defaultJSONConfig`) — безопасно только
пока файл полный.

## Прогоны

- `go test ./internal/...` (graph-service) — зелёные, новый тест красный на мутации кэша.
- `vitest graph.test.ts graphLoader.test.ts` 52/52, `config.test.ts` 2/2, `svelte-check` 0 ошибок.
- `pytest -k env_precedence` — зелёный.
- `check-config-registry.mjs` — 189 ключей, дрейфа нет.

## Ревью — Claude Code, 2026-09-30

Реализация: Devin, `954acfc`. **Вердикт: отклонено — три блокера и решение владельца по мёртвым ключам.**
Основное сделано хорошо: реестр 189 ключей и сторож, тесты порядка приоритета во всех сервисах, граф-сервис
больше не отравляет кэш урезанным ответом.

### Мутации

| Мутация | Результат |
|---|---|
| любой явный `limit` читает и пишет канонический кэш | красная: `TestGetFullGraphHandler_LimitDoesNotPoisonCache` |
| `limit` выше предела не урезается | красная: тот же тест |
| `limit=0` обходит предел | красная: `TestGetPublicGraphHandlerNoCache/cached_empty` |
| `GRAPH_FULL_LIMIT` из окружения не читается | красная: `TestFullLimitEnvOverridesJSONFile`, `TestLimitHelperPrecedence` |
| ключ в `config/*.json` без строки реестра | красная: сторож — «missing from the registry», «drift» |
| **фронтенд снова просит полный граф на 100 узлов** | **зелёная** — блокер 1 |
| умолчание в коде граф-сервиса снова 1000 | зелёная — мелочь: работает только без файла конфига |

### Блокер 1: запрос полного графа с фронтенда не закреплён тестом

Находка 4 постановки — `frontend.api.default_limit = 100` уходил в `getFullGraphData` — исправлена умолчанием
`limit = 0` (`shared/api/graph.ts`). Но тест «should use default limit when called without parameter»
(`graph.test.ts:216`) проверяет лишь, что `limit` непустой, а это верно и для «0», и для «100». Вернуть 100 — значит
снова показывать на графе 100 заметок из всех, и ни один тест этого не заметит. Нужно: тест, что без параметра
уходит `limit=0` (или `limit` не уходит); мутация «100» — красная.

### Блокер 2: шаг CI без строки в `core-checks.tsv`

Шаг «Check CONFIG registry» добавлен в `_core-checks.yml`, строки в `scripts/testing/core-checks.tsv` нет:
`check-core-workflow-sync.mjs` — «Workflow-only steps: … Check CONFIG registry» (рядом та же беда у шага
SPEC-AUDIT-1). Первая фаза `check-all.ps1` красная.

### Блокер 3: интеграционный тест дельты граф-сервиса красный

`TestGRPCIntegrationTestSuite/TestGetDelta` (`grpc_integration_test.go:203`): тест из SYNC-1 запрашивает
`GetFullLayout` с `Limit: 10`, затем дельту от выданной версии — и получает «snapshot not found: resync
required». После правки узкий запрос снимок не сохраняет — так и задумано, чтобы урезанный ответ не отравлял базу
дельт. Тест нужно перевести на канонический размер и добавить проверку, что узкий запрос снимка не создаёт и дельта
от него отвечает `resync`. В отчёте прогон граф-сервиса — `go test ./internal/...` без тега `integration`, поэтому
падение не было видно.

### Решение владельца 102 — мёртвые ключи

Сервис NLP читает общий `knowledge-graph.config.json` в том же порядке, что бэкенд и граф-сервис: окружение, затем
файл, затем умолчание; ключи `nlp.*` оживают. Способ, который не трогает код библиотек: при старте сервис берёт
`nlp.*` из файла и подставляет в окружение то, чего там нет. Тест порядка: окружение важнее файла, файл важнее
умолчания. Файл попадает в образ NLP так же, как в образы бэкенда. Остальные 11 мёртвых ключей — убрать.

### Мелочь

- Граф-сервис при отсутствии ключа в JSON получает ноль, а не умолчание (записано Devin) — отдельной строкой в
  бэклог после 1.0, если не чинится здесь попутно.

### Принято без замечаний

- Реестр 189 ключей, генератор и сторож; сторож ловит ключ без строки и дрейф.
- Порядок приоритета подтверждён тестами в бэкенде, граф-сервисе, NLP и фронтенде.
- Кэш полного графа: канонический размер — сконфигурированный предел; меньший `limit` — свежий ответ без кэша;
  больший — урезается; `limit=0` предел не обходит. Умолчание в коде приведено к 500.
- Документация конфигурации дополнена.
