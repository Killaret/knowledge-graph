# DB-POOL-1. Параметризовать пул соединений PostgreSQL

Из разбора пакета Gordon (GORDON-2): настройки пула БД были захардкожены в `db.go` — изменение требовало пересборки и передеплоя. Имена `POSTGRES_*` уже были задокументированы в `.env.example` и `DEPLOY.md`/`DEPLOY.ru.md`, но никогда не реализованы.

## Что сделано

- `config/backend.json` — новая секция `database.pool`: `max_open_conns: 25`, `max_idle_conns: 5`, `conn_max_lifetime_seconds: 300`, `conn_max_idle_time_seconds: 60`, `stats_interval_seconds: 300`. Дефолты = прежние захардкоженные значения, поведение не меняется.
- `GET /api/v1/metrics/database` — новый endpoint со статистикой пула (`db.Stats()` через `GetPoolStats`), только для роли `admin` (`middleware.RequireAdmin()`); статический API-ключ даёт роль `admin`. Описан в `openAPI.yaml` (контрактный тест).
- Периодическое логирование пула в `cmd/server` — интервал тикера теперь из конфига (`stats_interval_seconds`, был захардкожен 5 мин); неположительное значение → дефолт 5 мин.
- `backend/internal/config/config.go` — поля `DatabasePool*` в `Config`, загрузка по проектной схеме `env → JSON → дефолт` (`getIntEnv` + `getJSONIntOrDefault`).
- `backend/internal/infrastructure/db/db.go` — `PoolConfig`, `DefaultPoolConfig()`, `ConnectWithPool(dsn, pool)` с нормализацией неположительных значений к дефолтам; `Connect(dsn)` остался обёрткой — cli/seed/embed-recompute/rotate-api-keys и существующие тесты не тронуты.
- `cmd/server` (оба вызова в `connectDatabaseWithRetry`) и `cmd/worker` передают значения из `cfg`.
- `knowledge-graph.config.json` перегенерирован (`npm run build-config`).

## Env-оверрайды

| Переменная | Дефолт |
|---|---|
| `POSTGRES_MAX_OPEN_CONNS` | 25 |
| `POSTGRES_MAX_IDLE_CONNS` | 5 |
| `POSTGRES_CONN_MAX_LIFETIME_SECONDS` | 300 |
| `POSTGRES_CONN_MAX_IDLE_TIME_SECONDS` | 60 |
| `POSTGRES_POOL_STATS_INTERVAL_SECONDS` | 300 |

В `.env.example`, `DEPLOY.md`, `DEPLOY.ru.md` имена приведены в соответствие (было `*_MINUTES`, реализовано в секундах по конвенции проекта); `docs/CONFIGURATION_EN.md` — секция `database.pool` в JSON-примере и блок переменных.

## Проверка

- `go build ./...` — OK; `go test ./internal/infrastructure/db ./internal/config ./cmd/...` — PASS, включая новые unit-тесты `normalizePoolConfig`, `ConnectWithPool` на пустом DSN и `TestMetricsHandler_ReturnsPoolStats`; контрактный тест `TestRouterMatchesOpenAPISpec` PASS (роут описан в `openAPI.yaml`).
- `check-tasks-index`, `check-decisions`, `check-docs-links` — OK.

## Статус

Выполнено Devin 2026-09-18, на ревью у Claude Code. Источник — открытый вопрос №1 разбора GORDON-2.
