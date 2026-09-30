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
