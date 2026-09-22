# Knowledge Graph — полный обзор проекта для AI-агентов

> **Живой артефакт передачи знаний.**
> Этот файл — обновляемый артефакт (living artifact) для передачи контекста между AI-агентами и human-разработчиками. Он содержит актуальный срез архитектуры, стека, тестовой стратегии, найденных рисков, применённых исправлений и roadmap. При любых существенных изменениях проекта этот документ следует актуализировать.

Документ подготовлен на основе read-only аудита архитектуры, DDD-слоёв, стиля кода, инфраструктуры и тестовой пирамиды, а также последующих исправлений рисков и полного регрессионного цикла.

---

## 1. Общее описание проекта

**Knowledge Graph** — мультитенантное SaaS-приложение для управления заметками с графовой структурой, NLP-рекомендациями и 3D-визуализацией (`docs/architecture/ARCHITECTURE_SUMMARY.md`, `README.md`).

Основные возможности:

- 3D-визуализация заметок как небесных тел.
- Графовые связи между заметками.
- Семантический поиск на pgvector.
- NLP-рекомендации (sentence-transformers).
- Система черновиков в MongoDB.
- JWT/OAuth2-аутентификация, RBAC.
- Геймификация (achievements).
- i18n: английский по умолчанию (`en`), русский (`ru`) — через те же i18n-ключи.
- Резервное копирование на Яндекс.Диск.

Среды:

- **Dev** — `docker-compose.yml`.
- **Personal** — `docker-compose.personal.yml`.
- **Test** — `docker-compose.test.yml` (изолированный, с `kg-test-*` контейнерами, описан в `.windsurfrules`).

---

## 2. Технологический стек

### Backend (Go)

- **Go 1.25** — основной язык (`backend/go.mod`).
- **Gin v1.12** — HTTP-роутер/фреймворк.
- **GORM v1.31.2** — ORM для PostgreSQL.
- **pgx/v5** — драйвер PostgreSQL.
- **go-redis/v9** — клиент Redis (запрещён v8 API, `.windsurfrules`).
- **asynq v0.26.0** — очереди задач на Redis (обновлён с v0.23.0, `backend/go.mod`).
- **mongo-driver v1.17.9** — MongoDB.
- **pgvector-go v0.2.0** — векторный поиск.
- **golang-jwt/jwt/v5** — JWT.
- **testify** — тесты, **testcontainers-go** — интеграционные тесты.
- **miniredis/v2** — in-memory Redis для тестов.
- **swaggo/swag** — OpenAPI/Swagger документация.

### Frontend (SvelteKit)

- **Svelte 5 + runes** (`$state`, `$derived`, `$effect`, `$props`, `.windsurfrules`).
- **TypeScript strict**.
- **SvelteKit** — meta-фреймворк.
- **ky v1.14** — HTTP-клиент.
- **D3-force v3** / **Three.js v0.184** — граф и 3D.
- **Node v22.22.2** — runtime для фронтенда.
- **Vitest v5** — юнит-тесты (обновлено с v3 в ходе PR #63).
- **Playwright v1.59** — E2E.
- **@cucumber/cucumber v12** — BDD.
- **FSD + Atomic Design** — структура `frontend/src/{shared,components,entities,features,widgets,routes}` (`.windsurfrules`).

### NLP

- **Python 3.11**, **FastAPI**.
- **sentence-transformers**, **transformers**, **torch**, **keybert**, **pymorphy3**, **nltk**.
- Модель по умолчанию — `paraphrase-multilingual-MiniLM-L12-v2` (384-мерный, единое пространство для русского и английского).
- `NLP_MODEL_NAME` — единая переменная окружения для сервисов: `nlp-service` предзагружает её на стадии сборки, `backend` и `graph-service` фильтруют векторы по ней.
- `HF_HUB_OFFLINE=1` — offline-first режим (dev/personal); в тестовом стеке `HF_HUB_OFFLINE=0`, чтобы можно было докачать модель при пустом host-cache (`nlp-service/Dockerfile`, `docker-compose.test.yml`).

### Инфраструктура

- **Docker multi-stage** для всех сервисов.
- **nginx** — gateway.
- **PostgreSQL 16 + pgvector**, **Redis 7**, **MongoDB 7**.
- **Java 17 source-text-handler** (Spring Boot).

---

## 3. Архитектура и DDD-слои

Проект построен по **Clean Architecture** с четырьмя слоями (внутренние → внешние):

```
domain/          → сущности, value objects, aggregates (чистый Go)
application/     → use cases, сервисы, query handlers, кэш
infrastructure/  → GORM-репозитории, Redis, Mongo, внешние API
interfaces/api/  → Gin handlers, middleware, DTOs
```

### Domain layer

- `internal/domain/note` — `Note`, `Title`, `Content`, `Metadata`.
- `internal/domain/link` — `Link`, `LinkType`, value objects.
- `internal/domain/tag` — `Tag`.
- `internal/domain/user` — `User`, ошибки.
- `internal/domain/share` — sharing-логика.
- `internal/domain/graph` — BFS, нормализация, keyword matching, traversal.
- `internal/domain/achievement` — achievements engine.
- `internal/domain/permission` — интерфейсы репозитория.

### Application layer

- `internal/application/{achievement, cache, draft, graph, import, recommendation, user}`.
- `internal/application/queries/graph` — CQRS-lite query handlers (`GetSuggestionsHandler`).
- `internal/application/events` — публикация событий.
- `internal/application/common` — `TaskQueue` абстракция.

### Infrastructure layer

- `internal/infrastructure/db/postgres` — GORM-репозитории.
- `internal/infrastructure/db` — пулы подключений.
- `internal/infrastructure/cache` — Redis cache client.
- `internal/infrastructure/queue` — `asynq`-очередь и worker.
- `internal/infrastructure/nlp` — клиент NLP-сервиса.
- `internal/infrastructure/mongo` — MongoDB-репозитории.
- `internal/infrastructure/cloud` — Яндекс.Диск backup.
- `internal/infrastructure/email` — почта.
- `internal/infrastructure/web` — `ImportFetcher` для content extraction.

### Interfaces layer

- `internal/interfaces/api/handlers/*` — HTTP-хендлеры.
- `internal/interfaces/api/notehandler`, `taghandler`, `linkhandler` и др.
- `internal/interfaces/api/middleware` — auth, recovery, CORS, rate limiting.
- `internal/interfaces/api/common` — DTOs и валидация.
- `cmd/server/main.go` — wiring зависимостей.

### Мультитенантность и безопасность

- PostgreSQL **Row-Level Security (RLS)** для tenant isolation.
- JWT + RBAC.
- Rate limiting на write-эндпоинтах.
- Аудит-логи и черновики в MongoDB (TTL, high-volume writes).

### Graph Service

- Отдельный Go-микросервис: `services/graph-service/`.
- gRPC-порт 9090, HTTP-порт 9091.
- Своя БД (PostgreSQL).
- Используется для аналитики графа.

---

## 4. Frontend-архитектура

### FSD + Atomic Design

- `src/shared/` — утилиты, API, типы, stores, сервисы, конфиг.
- `src/components/{atoms,molecules,organisms}/` — UI-компоненты.
- `src/entities/` — domain entities (note, user, tag, achievement).
- `src/features/` — пользовательские сценарии (graph-interaction, graph-forms, preload, home-page).
- `src/widgets/` — сложные секции (`SidebarWidget`, `GraphCanvas`, `CockpitLayout`).
- `src/routes/` — SvelteKit-страницы.

### Правила импортов (MANDATORY)

- `shared/` не импортирует `entities/features/widgets/routes`.
- `components/atoms` не импортирует `molecules/organisms`.
- `entities/` импортируют только `shared/`.
- `widgets/` могут импортовать все нижележащие слои.

### Svelte 5 runes

- Запрещены Svelte 4 `writable`/дёривативы.
- `$state`, `$derived`, `$effect`, `$props` — единственный допустимый способ.
- Типизация strict, никаких `any` в production-коде (после фикса убраны 3 `as any` в auth-сторах).

### i18n

- `src/shared/utils/i18n.ts` — barrel, реэкспортирует `formatMessage` и типы `Locale`/`MessageParams`.
- `src/shared/utils/i18n/messages/*.ts` — ключи по доменам (`auth`, `common`, `graph`, `import`, `notes`, `profile`, `ui`) для `en` и `ru`.
- `formatMessage(key, locale, params)`.
- UI по умолчанию на английском, но все строки через i18n-ключи; Russian supported through the same keys.
- `SidebarWidget.svelte` переведена на i18n-ключи.

---

## 5. Инфраструктура и Docker

### Многоступенчатые Dockerfile

- Все сервисы обязаны быть multi-stage.
- `backend/Dockerfile`, `frontend/Dockerfile`, `nlp-service/Dockerfile`, `services/graph-service/Dockerfile`, `source-text-handler/Dockerfile`.
- Все production-образы содержат `HEALTHCHECK`.

### HEALTHCHECK endpoints

- backend `/health`.
- frontend `/health` (SvelteKit endpoint).
- nlp `/health`.
- graph-service `/health` (gRPC 9090, HTTP 9091).
- source-text-handler `/health`.

### Стеки

- **Dev**: backend 9000, nginx API 18080, nginx frontend 18081, graph-service 9091.
- **Personal**: backend direct 18085, nginx API 18082, nginx frontend 18084, graph-service 9092.
- **Test**: frontend 3002, backend 18083, nginx public perimeter 18086, graph-service gRPC 19090 / HTTP 19091, postgres 15434, redis 16381, mongo 27019, nlp 15002.

### Volumes

- Dev: `postgres_data`, `redis_data`, `huggingface_cache`.
- Personal: `pgdata_personal`, `redisdata_personal`, `mongodbdata_personal`.
- Test: `test_postgres_data`, `test_mongodb_data`.

### Nginx

- `nginx.conf` и `nginx.personal.conf` — gateway с проксированием `/api` и `/graph-service/api`.
- Публичный graph-service proxy обнуляет `X-Internal-Auth` и `X-User-Id`; делегирование пользователя по внутреннему токену включается отдельно через `GRAPH_SERVICE_TRUST_USER_HEADER` только внутри Docker-сети.
- Оба nginx ограничивают тело запроса 10 MiB, скрывают версию и выставляют `nosniff`, `SAMEORIGIN`, `strict-origin-when-cross-origin`.

---

## 6. Тестовая пирамида

| Уровень         | Команда                                         | Инструмент                       | Покрытие            |
| --------------- | ----------------------------------------------- | -------------------------------- | ------------------- |
| Go unit         | `cd backend && go test ./...`                   | testify                          | min 70%, target 70% |
| Go integration  | `cd backend && go test -tags=integration ./...` | testcontainers-go, miniredis     | —                   |
| Frontend unit   | `cd frontend && npm run test:unit`              | Vitest                           | target 70%          |
| E2E             | `cd frontend && npm run test`                   | Playwright                       | —                   |
| BDD             | `cd frontend && npm run test:bdd`               | Cucumber                         | —                   |
| NLP             | `cd nlp-service && pytest`                      | pytest                           | —                   |
| Local core checks | `.\scripts\testing\check-all.ps1 [-Quick]`    | PowerShell/Bash + shared manifest | CI-equivalent phases |
| Full regression | `.\scripts\testing\run-full-test-cycle.ps1`     | PowerShell + Docker + Playwright | —                   |
| Stacks identity | `.\scripts\ci\check-stacks-identity.ps1`        | PowerShell                       | —                   |

### Regression

- `run-full-test-cycle.ps1` включает:
  1. Snapshot состояния dev/personal.
  2. Остановку dev/personal.
  3. Проверку `check-stacks-identity`.
  4. Подъём test stack.
  5. Seed тестовых данных.
  6. Backend unit и integration.
  7. Frontend unit.
  8. Playwright E2E в `skip-auth` и `real-auth` режимах.
  9. Восстановление dev/personal.

---

## 7. Что проверялось в рамках read-only аудита

- **DDD-слои**: направления импортов, отсутствие обратных зависимостей `infrastructure → interfaces`, глобальных переменных, `panic` в бизнес-логике.
- **Code style Go**: идиоматичность, обработка ошибок, конструкторы с DI, rate limiting на write endpoint-ах.
- **Redis API**: использование `go-redis/v9` и отсутствие `go-redis/redis/v8`.
- **Frontend**: Svelte 5 runes, строгая типизация, FSD-границы, i18n, отсутствие `any`.
- **Docker**: multi-stage, `HEALTHCHECK`, версии базовых образов.
- **Compose**: консистентность dev/personal/test, порты, volumes, `SKIP_AUTH`, `HF_HUB_OFFLINE`.
- **Тесты**: покрытие, threshold, прохождение юнит, интеграционных, E2E, BDD.
- **Миграции**: отсутствие пропусков в нумерации.
- **CORS**: конфигурация origins, methods, headers.
- **Зависимости**: транзитивные устаревшие пакеты (`go-redis/v8` через asynq).

---

## 8. Найденные риски и несоответствия (изначально)

1. **Отслеживаемые бинарники**: `backend/bin/server` и `backend/bin/cli.exe` были закоммичены.
2. **Отсутствовали HEALTHCHECK** в `graph-service/Dockerfile` и `source-text-handler/Dockerfile`.
3. **Go version mismatch**: `graph-service` использовал Go 1.24 вместо 1.25.
4. **NLP Dockerfile**: одностадийный, без multi-stage.
5. **Несоответствия в compose/документации**: неправильные порты graph-service, отсутствовал `redis_data` volume, устаревшие ссылки на `src/shared/three/`.
6. **Frontend i18n и `any`**: `SidebarWidget.svelte` содержал хардкодный русский; `auth-session` и `auth` использовали `as any`.
7. **Formatter/ESLint**: `npm run format:check` и `npx eslint .` сообщали о проблемах.
8. **Coverage thresholds**: frontend и backend enforced min подняты до 70% (было 60% в `vitest.config.ts` и 64.8% в backend CI).
9. **go-redis v8 transitive**: устаревший `github.com/go-redis/redis/v8` тянулся через `asynq`.
10. **Пропуски миграций**: отсутствовали файлы `015` и `021`.
11. **CORS**: `CORS_ALLOWED_ORIGINS` был настроен, но methods/headers/max-age захардкожены в middleware.
12. **Backend `panic()`**: `internal/config/config.go:360` (`mustJSON` вызывает `panic(err)`).
13. **`lib/pq`**: импорт устаревшего драйвера в `cmd/seed/main.go` и `internal/infrastructure/db/db_connection_test.go` при использовании `pgx/v5`.
14. **Компиляция интеграционных тестов `notehandler`**: `New(...)` вызывался с 11 аргументами вместо 12 (пропущен `importSvc`).

---

## 9. Что было исправлено

### 9.1. Бинарники

- Удалены `backend/bin/server`, `backend/bin/cli.exe`.
- Добавлен `backend/bin/` в `.gitignore`.

### 9.2. HEALTHCHECK

- `services/graph-service/Dockerfile` — `HEALTHCHECK` на HTTP 9091.
- `source-text-handler/Dockerfile` + `HealthCheckService.java` — добавлен эндпоинт `/health`.

### 9.3. Go version graph-service

- `services/graph-service/go.mod` — `go 1.25.0`.
- `services/graph-service/Dockerfile` — `FROM golang:1.25-alpine AS builder`.

### 9.4. NLP Dockerfile и compose

- Переделан в двухстадийный: builder + runtime, копируется venv, HuggingFace cache, NLTK data.
- `entrypoint.sh` использует `${HF_HOME:-/root/.cache/huggingface}`.
- `docker-compose.test.yml` — `HF_HUB_OFFLINE=0`, чтобы тестовый стек мог докачать модель при пустом host-cache.

### 9.5. Compose/документация

- `.windsurfrules` и `docker-compose.test.yml` — порт graph-service приведён к gRPC 19090 / HTTP 19091.
- `docker-compose.yml` — добавлен volume `redis_data` для dev Redis.
- Устаревшие ссылки на `src/shared/three/` актуализированы (Three.js-логика перенесена в `docs/archive/3d/frontend/src/lib/three/`).

### 9.6. Frontend i18n и строгая типизация

- Добавлены i18n-ключи `nav.*` в `en` и `ru` секции `frontend/src/shared/utils/i18n.ts`.
- `SidebarWidget.svelte` полностью переведена на `t("...")`.
- Добавлен `frontend/src/shared/types/window.d.ts` с `__SKIP_AUTH__` и `__ACCESS_TOKEN__`.
- Убраны `as any` в `auth.svelte.ts` и `auth-session.svelte.ts`.

### 9.7. Format / lint

- Выполнены `npm run format` и `npm run lint`.
- `npm run format:check` и `npx eslint .` — чисто.

### 9.8. Coverage

- `vitest.config.ts` — thresholds подняты до 70% по lines/functions/branches/statements.

### 9.9. go-redis v8

- `backend/go.mod` — `asynq v0.23.0 → v0.26.0`, `go-redis/v9 v9.5.5 → v9.14.1`.
- Проверено: `go-redis/redis/v8` больше не фигурирует в `go.mod`.

### 9.10. Миграции

- Добавлены `015_noop_schema_anchor.{up,down}.sql` и `021_noop_schema_anchor.{up,down}.sql` (no-op `SELECT 1;`).

### 9.11. Интеграционные тесты notehandler

- Исправлены вызовы `New(...)` во всех `*_integration_test.go`, добавлен последний аргумент `importSvc` (`nil` для тестов).

### 9.12. Event-driven backup

- Добавлен `BackupEnabled` (`BACKUP_ENABLED`) в `knowledge-graph.config.json`, `internal/config/config.go` и Docker Compose.
- `docker-compose.yml`/`docker-compose.test.yml` устанавливают `BACKUP_ENABLED=false`; `docker-compose.personal.yml` — `BACKUP_ENABLED=true`.
- `AsynqClient` не ставит `backup:database` в очередь, если `BackupEnabled=false`.
- `worker/main.go` регистрирует `BackupDatabaseHandler`/`BackupToCloudHandler` только при `BackupEnabled=true`.
- Исправлена дедупликация `NewDatabaseBackupTask`: убран timestamp из payload, `asynq.Unique(5m)` теперь работает корректно.

---

## 10. Результаты верификации

Все изменения прошли следующие проверки:

- `cd backend && go test -short -count=1 ./...` — 0 failures.
- `cd backend && go test -tags=integration -count=1 ./...` — 0 failures (включая `notehandler`).
- `cd backend && go vet -tags=integration ./...` — 0 warnings.
- `cd backend && go build ./cmd/server && go build ./cmd/worker && go build ./cmd/cli` — успешно.
- `cd frontend && npm run check` — 0 errors, 0 warnings.
- `cd frontend && npm run build` — успешно.
- `cd frontend && npm run test:coverage` — проходил при thresholds 70% (lines 80.36%, branch 80.69%, functions 79.97%) до обновления Vitest 5. После PR #63: lines 69.85%, statements 65.98%, functions 64.95%, branches 57.47% — ниже порога 70%, но сдвинулся ближе. Без стратегии по `home-page.svelte.ts`, Svelte-компонентам (`CockpitPanel`, `CockpitNoteDetails`) и `src/routes/**` 70% global не достижим.
- `cd frontend && npm run format:check` — чисто.
- `cd frontend && npx eslint .` — чисто.
- `.\scripts\testing\run-full-test-cycle.ps1 -SkipManual` — exit code 0, оба режима (`skip-auth` и `real-auth`) Playwright-E2E прошли, dev/personal стеки восстановлены.

---

## 11. Остаточные риски / замечания, требующие внимания

1. **`panic` в `internal/config/config.go:360`**
   `mustJSON(value)` вызывает `panic(err)` при ошибке сериализации. Рекомендуется вернуть ошибку.

2. **Использование `lib/pq`**
   Импорт `github.com/lib/pq` остаётся в:
   - `backend/cmd/seed/main.go`
   - `backend/internal/infrastructure/db/db_connection_test.go`
     Проект использует `pgx/v5`; `lib/pq` — лишняя устаревшая зависимость, может быть удалена.

3. **CORS middleware**
   `CORS_ALLOWED_ORIGINS` вынесен в env, но `methods`, `headers`, `max-age` захардкожены. Для полной конфигурируемости стоит вынести их в переменные окружения.

4. **Устаревшие ссылки на `src/shared/three/`**
   Основные ссылки в `.windsurfrules` и документах исправлены, но в `docs/archive/3d/` остаётся старая иерархия (архив, не production).

5. **Security alerts после включения Dependency graph / CodeQL**
   После включения Dependency graph, Dependabot alerts и CodeQL появился пул задач, вынесенных в отдельные issues:
   - **#41** — очередь на review Dependabot PR.
   - **#42** — review настроек безопасности GitHub (`main` branch protection, workflow permissions, CODEOWNERS, visibility).
   - **#43** — review `.github/dependabot.yml` (добавить `services/graph-service` и root npm, правила группировки/ignore).
   - **#44–#48** — Dependabot alerts по манифестам (`backend/go.mod`, `services/graph-service/go.mod`, `package-lock.json`, `frontend/package-lock.json`, `nlp-service/requirements.txt`). Всего 148 алертов: 16 critical, 74 high, 49 medium, 9 low.
   - **#49–#54** — CodeQL alerts (cookie Secure, SSRF в `import_fetcher.go`, weak hashing, `extract-urls.ts`, `check-core-workflow-sync.mjs`, workflow permissions).

   **Состояние 2026-09-12:** ruleset `main` активен (ID `23024343`), blanket-major-игноры убраны из `dependabot.yml`, `services/graph-service` и root `npm` добавлены, `permissions:` добавлены в `_core-checks.yml`, `frontend-tests.yml`, `ci.yml`, `security.yml`. #50 dismissed как `won't fix`. #51 (API-key Argon2id), #52 (`extract-urls.ts` sanitization) и #53 (`check-core-workflow-sync.mjs` escaping) реализованы в PR #55 (`security/findings`), тесты проходят. #49 (cookie Secure) dismissed как mitigated. PR #55 замёржен в `main` 2026-09-12. #54 (workflow permissions) закроется после повторного CodeQL-скана на `main`. Следующий шаг: пересоздать существующие API-ключи (breaking change) и обработать Dependabot PR.

6. **Frontend Dependabot group PR #63**
   - Node обновлён до `v22.22.2`, чтобы удовлетворить `jsdom@30`.
   - Из группы Dependabot исключены/откачены: `eslint` до `^9.39.5`, `@eslint/js` до `^9.22.0`, `typescript` до `^5.9.3` (ESLint 10 и TS 7 несовместимы с `eslint-plugin-jsx-a11y` и SvelteKit).
   - `ky` v1.7+ адаптирован: хуки принимают state-объект (`{ request }` / `{ request, response }`), `prefixUrl` заменён на `prefix`.
   - Моки Vitest 5 приведены к конструируемым `function`-реализациям (`ResizeObserver`, `THREE.WebGLRenderer` и др.).
   - `npm run lint`, `npm run check`, `npm run test:unit` — зелёные. `npm run test:coverage` — **зелёное**: lines 83.62%, statements 81.9%, functions 81.89%, branches 70.04%, все выше порога 70%.
   - FE-COVERAGE-1 завершён: покрыты API (`notes.ts`, `links.ts`, `sharing.ts`), `client.ts`, `graph.svelte.ts`, `auth.svelte.test.ts`, `overlay.svelte`, `CockpitNoteDetails`, `CockpitPanel`, `FloatingAuthPanel`, `QuickCaptureWidget`, `home-page.svelte.ts`, `NoteCard`, `AuthCard`, `graphUtils`, `GraphPageShell`, `deviceCapabilities`, `galactic-lexicon`, `extract-urls`, `ToastNotification`, `CockpitHUD`, `node-renderers` и route-спеки (`import`, `search`, `profile`, `notes/[id]`, `notes/[id]/edit`, `notes/new`). Также исправлены типы в 9 test-файлах и `GraphPageShellTestWrapper.svelte`, удалён `frontend/tmp-coverage-parse.cjs`. PR #63 готов к мержу.

---

## 12. Ключевые файлы для быстрого старта

- Главные правила: [`.windsurfrules`](../.windsurfrules).
- Архитектура: [`docs/architecture/ARCHITECTURE_SUMMARY.md`](architecture/ARCHITECTURE_SUMMARY.md).
- Команды: [`COMMANDS.md`](../COMMANDS.md).
- Backend wiring: [`backend/cmd/server/main.go`](../backend/cmd/server/main.go).
- Frontend i18n: [`frontend/src/shared/utils/i18n.ts`](../frontend/src/shared/utils/i18n.ts), [`frontend/src/shared/utils/i18n/messages/`](../frontend/src/shared/utils/i18n/messages).
- Frontend entry: [`frontend/src/routes/+page.svelte`](../frontend/src/routes/+page.svelte), [`frontend/src/features/home-page/home-page.svelte.ts`](../frontend/src/features/home-page/home-page.svelte.ts).
- NLP: [`nlp-service/Dockerfile`](../nlp-service/Dockerfile), [`nlp-service/app/main.py`](../nlp-service/app/main.py).
- Backup: [`docs/operations/BACKUP.md`](operations/BACKUP.md), [`knowledge-graph.config.json`](../knowledge-graph.config.json).
- Regression: [`scripts/testing/run-full-test-cycle.ps1`](../scripts/testing/run-full-test-cycle.ps1).
- Regression plan: [`docs/operations/REGRESSION_TEST_PLAN.md`](operations/REGRESSION_TEST_PLAN.md).
- Roadmap: [`ROADMAP.md`](../ROADMAP.md), детальные планы — [`BACKLOG.md`](product/BACKLOG.md).

---

## 13. Дополнение: статус фич и roadmap

### Общий статус проекта

- **Фаза:** Alpha → Beta.
- **Стабильность:** критических проблем нет.
- **Регрессионное тестирование:** 11/14 частей пройдено.
- **Покрытие тестами:** 1381 frontend unit-тестов проходят, покрытие выше 70% по всем четырём метрикам: lines 83.62%, statements 81.9%, functions 81.89%, branches 70.04% — FE-COVERAGE-1 **завершён**. PR #63 больше не блокируется `test`-job. Backend unit-тесты — все проходят.
- **Готовность к production:** ожидает финальных проверок (E2E, интеграция, CI/CD).

### Текущий фокус — уже выполнено

- Ручное тестирование всех функций.
- Исправление багов из ручного тестирования.
- Критические проверки перед production.
- E2E smoke, backend integration, CI/CD workflows, NLP API, auth API, публичный граф — всё пройдено.

### Реализованный UI: Cosmic Cockpit

- Космическая «кабина» с четырьмя выдвижными панелями, HUD, режим «от первого лица».
- Drag-to-open, якоря, 2D/3D-переключатель, фильтры типов, детали заметки, мини-граф связей, Singularity-зона архивирования.
- **2D Renderer Performance Pack (2026-08):** адаптивный туман войны, viewport culling, throttling до `idle_fps`, LOD по зуму, offscreen-кэш для стабильного состояния, кэширование `getVariation`/`isNewNode`, O(1) lookup эндпоинтов связей.
- **2D Renderer idle-throttling fix (2026-08):** removed per-frame debug logging in `GraphCanvas.svelte`; `startAnimationLoop` no longer fetches nodes on every rAF tick; heavy work (black hole/ghost updates, gravity, fog update, node-angle animation, redraw) is now skipped when the graph is stable and idle. Regression test added in `GraphCanvas.rendering.spec.ts`.
- **2D Renderer fog-warning toast (2026-08):** the low-FPS warning is now a debounced two-state toast (2 s hold, 1 s rearm debounce) controlled by `frontend/src/features/graph-canvas/fog-warning.ts`; it shows a red 'danger' banner when the load is high and a blue 'recovery' banner when the load drops, only while the user is actively interacting with the graph and without flickering. `GraphCanvas.svelte` now treats panning and node dragging as interactions (`dragState.dragging`) so the warning triggers on pan and drag, not only on hover and focus mode. `GraphCanvas.svelte` passes `fogWarningState.kind` to `GraphCanvasOverlay`. New i18n key `graphOverlay.fogRecovery`. Regression tests in `fog-warning.test.ts`.
- **2D Graph hover / neighbor highlight (2026-08):** `GraphCanvas.svelte` computes direct neighbor ids from `simLinks` and passes them to `fogState.update` and the renderer. `fog-state.svelte.ts` expands the clear radius to cover the hovered node and its direct neighbors. `drawAllNodes` now renders the hovered node at opacity 1, direct neighbors at 0.85, and unrelated nodes at 0.3; neighbors and the hovered node are not simplified when zoomed out. `link-renderers.ts` keeps links from the hovered node at full opacity and slightly boosts neighbor-neighbor links. New tests in `fog.test.ts`, `fog-state.svelte.test.ts`, and `renderer-orchestrator.test.ts`.
- **Graph top bar fog toggle (2026-08):** the fog button icon changed from a generic four-line icon to a cloud-with-lines icon. Added `graphOverlay.fogToggleTitle` i18n key and used it for `title`/`aria-label`/SVG `<title>` so the button has a clear tooltip. `GraphTopBar.spec.ts` still passes.
- **2D Node Color Variation (2026-08):** опциональная кастомизация цвета нод через `GraphNode.color`/`glowColor`; при отсутствии ручного значения `getVariation` выбирает детерминированный цвет из палитры `frontend/src/shared/lib/graph/color-schemes.ts`, вдохновлённой реальными космическими объектами.
- **3D Geometry Visibility Fix (2026-09):** nodes are now rendered as bright `MeshBasicMaterial` spheres using the celestial body glow color; node size scales with the graph bounding radius so spheres remain visible at the auto-zoom camera distance; fog density was tuned so it adds atmosphere without hiding geometry. 3D view now shows nodes, not only labels.
- **2D Layout Outlier Fix (2026-09):** tightened the bounding force radius and strength in `entities/graph-canvas/lib/simulation.ts` so isolated or weakly connected nodes no longer drift far from the cluster; the dense center remains a known limitation solvable only by semantic clustering.
- **Fog Toggle & Top Bar Spacing Fix (2026-09):** fog button toggle now schedules an immediate canvas redraw in 2D; top-bar controls were spaced so the fog button and language switcher no longer overlap.

### В процессе / запланировано (UI/UX)

- Исправление моргания графа при дельта-обновлениях.
- Селектор типов заметок — выпадающий список.
- Документация и UI для типов связей.
- Кнопка «Отмена/Назад» на странице входа.
- Мобильная адаптивность графа / touch-events (ручная проверка, запланировано, не реализовано).
- Система рекомендаций Pure Precomputed (убрать fallback).
- Публичные сиды для тестирования real-auth.
- Аудит ресурсов (память, CPU, bundle size).
- **Пассивный режим при переключении вкладки:** при `document.hidden` останавливать `requestAnimationFrame` и d3-симуляцию, при `visibilitychange` возобновлять, чтобы фоновая вкладка не ела CPU/GPU.

### Инструменты импорта/экспорта

| Фича                             | Статус                                                                                |
| -------------------------------- | ------------------------------------------------------------------------------------- |
| Букмарклет                       | Реализован                                                                            |
| Массовый импорт URL              | Реализован + извлечение контента по URL (`internal/infrastructure/web.ImportFetcher`) |
| Браузерное расширение            | Запланирован                                                                          |
| JSON/Markdown/CSV импорт-экспорт | Запланирован                                                                          |

> В роадмапе отмечено, что content extraction для массового импорта URL реализован через `internal/infrastructure/web.ImportFetcher`.

### Геймификация и Obsidian

- Система кастомизации, очки, достижения, бейджи, лидерборды — в планах.
- Импорт и синхронизация с Obsidian — в планах.

### PWA и внешние интеграции

- PWA Capture (быстрые заметки, оффлайн, push) — запланировано.
- Интеграции Pocket / Readwise / Twitter — запланированы.

### 3D и кластеризация

- Базовый 3D-рендеринг (Three.js) — разморожен и перенесён в `features/graph-3d`.
- Видимость 3D-нод/связей исправлена: сферы видны, туман не скрывает граф, 2D ↔ 3D не ломает связи.
- Orbital / Solar System 3D, Honeycomb Stellaris 3D, серверная кластеризация Louvain, кэширование кластеров — в планах.
- Галактические кластеры и LOD — в бэклоге.

### Связи, быстрое редактирование, onboarding

- Тултипы связей, анимация, gamma-кодирование, автосоздание связей — запланировано.
- Dust Inbox, inline-редактирование, автодополнение, исправление NLP-ключевых слов — запланировано.
- Сценарный onboarding с T1-T6, хлебные крошки, spotlight — запланировано.

### Экспериментальные идеи

- **Factory Line** — визуализация графа как производственная цепочка.
- **Semantic Guardians** — семантические стражи.

---

## 14. Рефакторинг августа 2026: единый State-driven UI для графа

### Цель

Убрать дублирование логики, зависящей от состояния авторизации, и выделить единую точку принятия решений по auth-conditional рендерингу/поведению.

### Внесённые изменения

- **GraphPageShell** (`frontend/src/widgets/graph-page/GraphPageShell.svelte`) — единый обёртковый компонент для всех граф-страниц (`/`, `/graph`, `/graph/3d`, `/graph/:id`, `/graph/3d/:id`).
  - Считает `typeCounts` и пробрасывает в `GraphTopBar`.
  - Условно передаёт `onNoteCreate`/`onNoteDelete`/`onNoteEdit`/`onCreateChildNote` при авторизации и `onSignIn`/`onRegister` при публичном режиме.
- **GraphTopBar + CockpitLeftPanel** — фильтры по типу, поиск, reset/focus и link-type controls остались только в `GraphTopBar`; `CockpitLeftPanel` сокращена до навигации, списка заметок и импорта/экспорта.
- **NoteCard readonly** — `readonly` prop скрывает edit/delete в tooltip; страница `/notes/[id]` скрывает соответствующие кнопки для неаутентифицированных пользователей.
- **Shared auth guards** (`frontend/src/shared/composables/auth.ts`):
  - `useAnonymousGuard()` — защида auth-страниц от залогиненных пользователей.
  - `useRequireAuth()` — редирект незалогиненных со страниц, требующих авторизацию (`profile`, `import`, `import/bookmarks`).
- **GraphLoader** (`frontend/src/shared/services/graphLoader.ts`) — единый, auth-aware загрузчик графа:
  - Используется `home-page.svelte.ts` (full graph через `getGraphWithPreload`, fallback из notes) и `routes/graph/+page.svelte` (full/centered graph, Knowledge Core).
  - Только `shared/` импорты; FSD-граница соблюдена.
  - `fullGraphLoader` передаётся как callback из `features/preload`, чтобы `shared` не зависел от `features`.

### Покрытие и статус

- `npx svelte-check` — 0 errors, 0 warnings.
- `npm run test:unit` — 103 test files, 946 tests passed.
- Test stack пересобран и поднят.
- Playwright (test stack): `smoke-real-auth`, `cockpit-canvas-controls`, `floating-auth-panel`, `public-graph` — 10/10 passed.
- Ручные сценарии: см. `docs/archive/MANUAL_TEST_CHECKLISTS_RU.md` раздел `0.6` и `docs/operations/TESTING.md` раздел `Manual Regression Scenarios`.

---

## 15. Дополнения и правки по итогам тестирования (август 2026)

### 15.1. Пассивный режим при переключении вкладки

**Требование:** при уходе пользователя на другую вкладку граф должен переходить в пассивный режим и не потреблять ресурсы системы.

**Почему это важно:** сейчас анимационный цикл `GraphCanvas.svelte` продолжает вызывать `requestAnimationFrame` и d3-симуляцию даже когда вкладка неактивна. Браузер обычно троттлит rAF до ~1 fps на фоновой вкладке, но вычисления в `onUpdate` и силам симуляции всё равно выполняются, что расходует CPU и батарею.

**Рекомендуемый подход:**
- В `GraphCanvas.svelte` добавить `document.addEventListener('visibilitychange', ...)`.  
  При `document.hidden === true`:
  - `cancelAnimationFrame(animationFrameId)`;
  - `simulation?.stop()` (или `simulation?.alphaTarget(0)` с `simulation?.tick(0)`);
  - прекратить обновление FPS/тумана.
  При `document.hidden === false`:
  - `simulation?.restart()`;
  - запустить `startAnimationLoop` заново;
  - пометить `needsRedraw = true`, чтобы восстановить картинку.
- Учесть, что `graph-service` и WebSocket/SSE-подключения (если есть) тоже можно заморозить, но это опционально.
- Добавить unit-тест / Playwright-тест, проверяющий, что на скрытой вкладке rAF не вызывается.

**Файлы:**
- `frontend/src/widgets/graph-canvas/GraphCanvas.svelte`
- `frontend/src/widgets/graph-canvas/GraphCanvas.rendering.spec.ts` (регресс-тест)

### 15.2. Fog-warning / recovery — итоги ручного теста

**Сделано:**
- Контроллер предупреждения переписан на two-state (`danger` красный / `recovery` синий), 2 s hold, 1 s rearm debounce.
- В `isInteracting` добавлен `dragState.dragging`, поэтому пан и drag-нод теперь считаются взаимодействием.

**Проблема, обнаруженная на тестовом стеке:**
- Предупреждение не удаётся поймать на публичном графе с 50 нодами, потому что FPS не падает ниже `warning_threshold` (18).

**Следующие шаги / договорённости:**
- Проверить поведение на **личном стеке** с реальными/большими данными.
- Если и там сложно поймать — пересидировать тестовый стек с большим public-графом (300–500 нод, 500–1000 связей) и/или временно поднять `warning_threshold` только для ручного теста.
- Не рекомендуется оставлять `warning_threshold` высоким в продакшене — это приведёт к ложным предупреждениям.

### 15.3. Оставшиеся ручные кейсы

- **Case 1.4** — hover, подсветка соседей и туман: в процессе ручной проверки.
  - 2026-08-23: реализована подсветка прямых соседей (opacity 1 / 0.85 / 0.3), расширение радиуса тумана на соседей и усиление связей от hovered-узла. Требуется повторная ручная проверка.

---

## 16. AI-агент prompt-экосистема

Для единообразной работы всех AI-моделей в проекте создана общая prompt-экосистема в `.devin/prompts/`.

### Роли и инструменты

- **Windsurf SWE 1.7 Max** — основной implementation-агент.
- **Devin** — CLI-аудит, автоматизация, верификация (`SKILL.md`).
- **DeepSeek** — стратегическая архитектура, roadmap, prompt design.
- **Claude / Claude Code** — общий мастер-промпт и аналитический промпт для обсуждений.
- **Python NLP service** — runtime embeddings, keywords, similarity (не development-агент).

### Файлы промптов

- `.devin/prompts/MASTER_PROMPT.md` — единый мастер-промпт для имплементации, ревью и работы с кодом (английский).
- `.devin/prompts/MASTER_PROMPT_RU.md` — русская версия мастер-промпта.
- `.devin/prompts/ANALYSIS_PROMPT.md` — промпт для стратегического анализа, архитектуры и trade-off.
- `.devin/prompts/README.md` — инструкция по использованию промптов.

### Правила использования

- В начале каждого чата с AI-моделью вставляйте соответствующий промпт.
- Мастер-промпт обязывает модель читать `.windsurfrules`, `PROJECT_REVIEW_AI_AGENTS.md`, `SKILL.md` и `ARCHITECTURE_SUMMARY.md`.
- Аналитический промпт используется для roadmap, архитектурных альтернатив и обзора рисков.
- Все промпты являются living documents: обновляются при изменении `.windsurfrules`, стека, архитектуры или ролей AI-инструментов.

### Ссылки

- [`.devin/prompts/MASTER_PROMPT.md`](../.devin/prompts/MASTER_PROMPT.md)
- [`.devin/prompts/MASTER_PROMPT_RU.md`](../.devin/prompts/MASTER_PROMPT_RU.md)
- [`.devin/prompts/ANALYSIS_PROMPT.md`](../.devin/prompts/ANALYSIS_PROMPT.md)
- [`.devin/prompts/README.md`](../.devin/prompts/README.md)
- [`.devin/skills/knowledge-graph/SKILL.md`](../.devin/skills/knowledge-graph/SKILL.md)
- [`.windsurfrules`](../.windsurfrules)
- [`docs/AGENTS.md`](AGENTS.md)
- [`docs/AGENTS_EN.md`](AGENTS_EN.md)

## 17. A-1: детерминированная 3D-визуализация и связанный публичный сидер (2026-09-06)

**Что сделано.**

- В `Graph3DViewer.svelte` убран `in:fade` для loading-оверлея в режиме стабильного рендера — `data-test-stable=true` теперь означает, что оверлей уже не виден.
- В `Graph3DScene.svelte` и `engine.ts` стабильный режим (`stableRender=true`) заставляет движок синхронно сходить и отрисовать ровно один кадр до `finishLoading()`.
- `OrbitControls.enableDamping` отключается при `disableAnimation`.
- `Math.random` сидируется во визуальном тесте и в `addInitScript`, чтобы звёздное поле и раскладка не менялись между прогонами.
- Визуальный тест `frontend/tests/visual/visual-regression.spec.ts` переводит курсор в центр сцены (`scene.hover()`) и отключает Argos-ский `disableHover`, чтобы движение мыши в `(0, 0)` не открывало панели и не сбивало стабилизацию скриншота.
- Сидер `scripts/testing/seed-test-data.ps1`/`seed-test-data.sh` публикует 20% заметок и создаёт связи преимущественно внутри пула публичных заметок, чтобы публичный граф содержал связанные компоненты.

**Результаты верификации.**

- `npm run test:unit`: 107 файлов, 987 тестов — зелёные.
- `npm run check`: 0 ошибок, 0 предупреждений.
- `npm run lint`: 0 новых замечаний (3 pre-existing warning).
- Все 13 визуальных тестов Playwright (`--project=visual`) прошли.
- 3D-тест отработал дважды с идентичным снимком (`3d-baseline.png`).
- Временное изменение `birth.density_final` с 0.0006 на 0.02 дало снимок, отличающийся на 13.58% пикселей (`3d-fog-dense.png`), что доказывает чувствительность эталона к графическим параметрам.
- Сидер на тест-стенде: 100 notes, 20 public, 60 links, 100 embeddings, 100 keyword-processed; graph-service: 100 nodes, 60 links.

**Артефакты.**

- Базовый и «туманный» снимки: [`docs/assets/a1-3d-visual-regression/`](assets/a1-3d-visual-regression)
- Скрипт сравнения: удалён после использования.

**Осталось.**

- Повторная верификация Claude Code на живом стенде и пересборка официальных baseline Argos (`ARGOS_REFERENCE_BRANCH=main`).

## 19. Ручной пересчёт рекомендаций и обнаруженные риски (2026-09-11)

### 19.1 Почему не работали рекомендации и граф

- `note_recommendations` и `links` были пусты после массового импорта.
- `embed-recompute -post` ставит задачи `RefreshRecommendations`/`RecalculateLinkWeights`, но `RefreshService.RefreshRecommendations` делает BFS по `links`. Без `links` кандидатов нет.
- `note_links_closure` — materialized view на основе `links`. Без `links` view пуст, graph-service тоже не находит связей.
- Семантический fallback (`FindSimilarNotes` по `note_embeddings`) работает, но не сохраняет результаты в `note_recommendations` и не создаёт `links`.

### 19.2 Ручное семантическое заполнение

- SQL `seed_semantic_safe.sql` заполнил 108 `note_recommendations` (топ-6 по косинусному сходству) и 18 `links` (топ-1, `link_type='related'`, `source_type='gamma'`).
- `REFRESH MATERIALIZED VIEW note_links_closure` — 28 строк.
- Живое API вернуло рекомендации и граф.
- Обнаружен баг: precomputed-ветка `GetSuggestions` не подгружала `title`. Исправлен в `backend/internal/interfaces/api/notehandler/note_handler.go` и покрыт тестами.

### 19.3 Docker Desktop — авария и анализ

**Простое объяснение, почему упала БД (и Docker):**

- Я вручную вставил 108 связей (`links`) — по 6 исходящих на каждую заметку.
- `note_links_closure` — это materialized view с рекурсивным CTE, который перебирает **все простые пути** между заметками до глубины 10.
- Средняя исходящая степень 6 означает: на глубине 5 — `6^5` = 7776 путей, на глубине 10 — десятки миллионов.
- PostgreSQL не справился с перебором, `psql` завис, WSL VM завис. Я принудительно остановил `wsl -t docker-desktop`.
- После аварийного останова Docker VM пошёл в `read-only file system`, и Docker Desktop перестал стартовать (`docker ps` → `Docker Desktop is unable to start`).

**Вывод:** проблема не в общем количестве связей, а в **высокой исходящей степени одного узла** (6 исходящих) в рекурсивном view. Если ограничить исходящую степень 1–2 связями на заметку, `note_links_closure` строится за миллисекунды. Поэтому `GammaLinkGenerator` лимитирует исходящую степень `MaxGammaOutDegree` (по умолчанию 2).

**Детали:**

- 108 связей, топ-6 семантических соседей на заметку → ~5.7 исходящих на узел.
- `REFRESH MATERIALIZED VIEW note_links_closure` вызвал экспоненциальный взрыв числа простых путей.
- **Урок:** `note_links_closure` не безопасен для плотных графов. Нужно либо ограничивать исходящую степень гамма-связей (`links` с `source_type='gamma'`) до 2 или меньше, либо переделывать view.

### 19.4 Java source-text-handler / batch API — на ревью

- Java-сервис обрабатывает URL/документ, чанкует, чистит контент, извлекает язык и формирует `title`/`content`.
- Java-specific endpoint `POST /api/v1/import/java/batch` был реализован, а затем удалён по решению владельца. Сейчас на ревью: `docs/tasks/IMP-4-claude-review.md`.
- Вопросы ревью:
  - должен ли Java использовать существующий `POST /api/v1/import/bookmarks`;
  - или нужен generic `POST /api/v1/import/batch` (без `java` в имени);
  - какие операции доступны Java: batch create, single create, no delete/edit;
  - как реализовать ручное создание связей между заметками (backend `POST /api/v1/links` уже есть, но frontend UI, судя по `UX-1`, не позволяет);
  - dedup по `external_id` и индекс в `metadata`;
  - pipeline постобработки с `GammaLinkGenerator` и `note_links_closure`.
- Предполагаемое тело generic batch:
  ```json
  {
    "request_id": "uuid",
    "items": [
      {"title": "...", "content": "...", "type": "unknown", "source_url": "...", "external_id": "...", "metadata": {...}}
    ]
  }
  ```
- Гамма-связи (`GammaLinkGenerator`) будут интегрированы в worker после ревью; сейчас сервис готов и покрыт тестами.

### 19.5 Нужные тесты

- Юнит: `GammaLinkGenerator` не создаёт больше `MaxGammaOutDegree` связей, пропускает self-loops, не дублирует `(source, target, link_type)` — реализовано, `go test ./internal/application/recommendation` зелёное.
- Интеграция: `REFRESH MATERIALIZED VIEW note_links_closure` с плотным графом (6 связей на узел) отменяется по `statement_timeout` или не завершается в разумное время — доказательство уязвимости.
- Интеграция: `REFRESH MATERIALIZED VIEW` с разреженным графом (≤2 связи на узел) завершается <1s.
- E2E/контракт: `POST /notes/batch` возвращает `data[].id`, `import_task_id` и признак постобработки.

### 19.6 BATCH-1: batch-роуты notes/links — на ревью Claude Code / владельца (2026-09-13)

**Что сделано.**

- Реализованы `POST /api/v1/notes/batch/create`, `POST /api/v1/notes/batch/delete`, `POST /api/v1/import/batch`.
- Старый `POST /api/v1/notes/batch` удалён из роутера и OpenAPI; фронтенд `deleteNotesBatch` переехал на `v1/notes/batch/delete`.
- Покрытие велось в согласованном цикле: регрессионные тесты → намеренно падающие тесты → исправления.
- Найдено и исправлено:
  - `POST /api/v1/import/batch` не проверял `FindByID == nil` и позволял создавать связи на несуществующие заметки;
  - пустой `notes` в `/import/batch` возвращал 200 вместо 400;
  - доменный лимит `note.NewContent` (10 000 rune) расходился с API/OpenAPI (50 000 символов) — приведён к 50 000;
  - `metadata.type` не валидировался по `ValidCelestialBodyTypes` в `POST /notes`, `/notes/batch/create` и `/import/batch`;
  - `source_url` из импортного batch-айтема терялась и не сохранялась в метаданных.
- Прогоны: `go test ./...`, `go vet ./...`, `go test ./cmd/server/...`, `npm run test:unit -- --run`, `npm run check`, `npm run lint` зелёные (9 pre-existing warnings).

**Открытые риски / вопросы.**

- Контракт ссылок в `/import/batch`: внешний Java/source-text handler не имеет UUID новых заметок. Текущий механизм клиентских `id` работает, но неудобен. Нужно решить: индексы массива, `external_id` с маппингом в ответе или упорядоченные операции. Обсуждается с Claude Code / владельцем.
- **Adversarial-тестирование:** процесс зафиксирован в `docs/tasks/BATCH-TEST-1-strategy.md`; осталось договориться о маркировке и формализации "практического исчерпания".
- **DDD / Clean Architecture:** валидация `noteType` сейчас в `interfaces`, нужен перенос в `domain`. Варианты описаны в `docs/tasks/BATCH-DDD-1-validation.md`.
- **Таксономия типов заметок:** обсуждена в `docs/tasks/NOTE-TYPE-TAXONOMY.md`; нужно согласовать `scaleRank`, состав `UI_TYPES` и единый порядок во всех списках (backend, frontend, OpenAPI) перед реализацией `BATCH-DDD-1`.

## 18. AUD-4: контракт входа через Яндекс (2026-09-06)

**Что сделано.**

- В `backend/cmd/server/router.go` инициация OAuth перемещена с `/api/v1/auth/yandex` на `/api/v1/auth/yandex/login`; старый путь удалён.
- В `backend/internal/interfaces/api/handlers/auth/handler.go` `YandexLogin` теперь возвращает `200` с JSON `{"url": "<authorization URL>"}` вместо HTTP-редиректа.
- `backend/internal/interfaces/api/middleware/jwt.go` уже содержал `/api/v1/auth/yandex/login` в `SkipPaths` — проверено, дублирующих записей нет.
- `backend/openAPI.yaml` обновлён: путь `/api/v1/auth/yandex/login`, ответ `200` с телом `{ url: string }`, плюс `501` при отсутствии настройки; блок `302` удалён.
- `frontend/src/shared/api/auth.ts` не требовал изменений — клиент уже звал `/auth/yandex/login` и ожидал JSON.
- `YandexLoginButton.svelte` использует `window.location.href = result.url`, то есть переход происходит в браузере, а не через `fetch`.
- `README.md` очищен от пометки «OAuth2 через Яндекс сейчас не работает».
- `docs/operations/CONFIGURATION_EN.md` дополнен разделом **Authentication** с JSON `backend.auth` и таблицей переменных окружения, включая `YANDEX_CLIENT_ID` и `YANDEX_CLIENT_SECRET`.
- В `docker-compose.yml`, `docker-compose.personal.yml` и `docker-compose.test.yml` переменные `YANDEX_CLIENT_ID` и `YANDEX_CLIENT_SECRET` теперь прокидываются в backend-сервисы из окружения или `.env`.
- Юнит-тест `TestYandexLogin_S256` обновлён: проверяет статус `200`, парсит JSON, разбирает URL и верифицирует параметры `client_id`, `response_type`, `state`, `code_challenge` и `code_challenge_method`. Добавлен `TestYandexLogin_NotConfigured` для случая без `YandexClientID`.

**Результаты верификации.**

- `go test ./...` — зелёное.
- `go vet ./...` — чисто.
- `npm run test:unit` — 107 файлов, 987 тестов зелёных.
- `npm run check` — 0 ошибок, 0 предупреждений.
- `npm run lint` — 0 новых замечаний (3 pre-existing warning).
- Живой тест-стенд: `GET http://127.0.0.1:18083/api/v1/auth/yandex/login` до правки возвращал `404`, после — `200` с URL `https://oauth.yandex.com/authorize?client_id=test-yandex-client-id&code_challenge=...&code_challenge_method=S256&response_type=code&scope=login%3Aemail+login%3Ainfo&state=...`.
- Старый путь `/api/v1/auth/yandex` теперь возвращает `401` (маршрут удалён, запрос падает на JWT middleware).

**Осталось.**

- Повторная проверка Claude Code на живом тест-стеке и, при необходимости, реальным `YANDEX_CLIENT_ID`/`YANDEX_CLIENT_SECRET` (callback остаётся вне зоны задачи).

### 19.6 Docker Desktop / Personal recovery (2026-09-11)

**Что произошло.**

- Падение Docker/WSL из-за `note_links_closure` (см. §19.3) оставило Docker Desktop в состоянии `unable to start`.
- `wsl -t docker-desktop` и аварийные попытки привели к тому, что WSL-дистрибутив `docker-desktop` не стартовал с ошибкой `Wsl/Service/CreateInstance/E_FAIL`.
- Диск D: был заполнен: `D:\Docker\wsl\disk\docker_data.vhdx` (45.18 ГБ) — активный Docker WSL-диск, данные Personal-стека внутри. Docker не мог стартовать из-за `There is not enough space on the disk` при копировании main-дистрибутива.
- `C:\Users\...\AppData\Local\Docker\wsl` оказался junction → `D:\Docker\wsl`, поэтому все Docker-данные живут на D:.

**Что сделано.**

- Ручной бэкап Personal-данных:
  - Архив: `C:\Users\89209\Desktop\my items\kg-personal-volumes-2026-09-11.tar.gz` (25.8 МБ).
  - В архиве: `knowledge-graph_pgdata_personal/_data`, `knowledge-graph_redisdata_personal`, `knowledge-graph_mongodbdata_personal`.
  - Проверка: `tar -tzf` показал 1589 записей, `pgdata_personal/_data` на месте.
- `wsl --unregister docker-desktop` — убран битый системный дистрибутив (данные не тронуты).
- `e2fsck -fy /dev/sde` — исправлены мелкие ошибки файловой системы VHDX (free blocks / free inodes count).
- `fstrim` + `diskpart compact vdisk` — уменьшили `docker_data.vhdx` с 45.18 ГБ до 37.04 ГБ, освободив ~8.1 ГБ на D:.
- Docker Desktop запущен, Personal- и test-стеки поднялись.
- Проверка данных: `SELECT COUNT(*) FROM notes` в `knowledge_personal` вернуло 19 — заметки на месте.

**Осталось / риски.**

- `D:\Docker\wsl\disk\docker_data.vhdx.backup` (8.41 ГБ, 22.08) — старый бэкап VHDX. Можно удалить после проверки свежего архива.
- `D:\Docker\wsl\main\ext4.vhdx.old` (100 МБ) — старый системный дистрибутив; можно удалить.
- После старта стеков `kg-nlp-personal` и `kg-nlp-test` находятся в `health: starting` — нужно дождаться full healthy, прежде чем тестировать рекомендации.
- Для перестраховки стоит сделать `pg_dump`-бэкап через `backup-personal.ps1` после восстановления, чтобы свежий `.sql.gz` дополнил VHDX-архив.

## 20. UX-2: 500-страница и аудит обработки ошибок (2026-09-11)

**Что сделано.**

- Добавлен `frontend/src/routes/+error.svelte`: full-viewport (`position: fixed; inset: 0; z-index: 900`), Cosmic Cockpit фон, i18n (`error.500.*`, `error.unknown.*`), кнопки «Обновить» и «На главную», dev-only stack trace.
- Добавлен новый тип `server-error` в `frontend/src/components/atoms/StateIllustration.svelte`: космическая иллюстрация разъединённого удлинителя с искрой.
- `+error.svelte` выбирает иллюстрацию по статусу: `404` → `404`, `5xx` → `server-error`, остальное → `error`.
- Добавлен `frontend/src/routes/error-page.spec.ts` и `frontend/src/shared/utils/route-match.test.ts`.
- В `+layout.svelte` исправлен критический баг публичных маршрутов: `currentPath.startsWith("/")` делал публичными **все** пути. Вынесена функция `isPublicRoute` в `shared/utils/route-match.ts`.
- Обновлены `docs/tasks/UX-2-500-error-page.md` и `docs/AI_HANDOFF.md`; статус: на ревью у Claude Code.
|- Добавлен Playwright-сценарий `frontend/tests/error-500-page.spec.ts` и test-only route `src/routes/test/500/+page.server.ts` (триггер `?trigger=500`) для проверки full-viewport рендера в браузере.
|- `ApiErrorDisplay.svelte` теперь использует иллюстрацию `server-error` для API-ошибок с кодом `INTERNAL_ERROR`.
|- Исправлены `golangci-lint` замечания в `note_handler.go` и `import_fetcher_test.go`.
|- Восстановлен гейт форматирования: `npm ci` + `npm run format` привели 6 Svelte-файлов в соответствие с lock-версией `prettier-plugin-svelte`.
|- Обновлена языковая политика: UI по умолчанию — English (`en`), документация — Russian; правка в `.windsurfrules`, `MASTER_PROMPT.md`, `MASTER_PROMPT_RU.md`.

**Аудит обработки ошибок.**

- SvelteKit route-ошибки (`load`/SSR) перехватываются `+error.svelte` (full-screen, i18n).
- API-ошибки в страницах graph, import, notes, search, home показываются **локально** (`StateIllustration`, `ApiErrorDisplay`) — не перекрывают весь экран.
- `GraphPageShell` и `GraphCanvas` ловят ошибки загрузки/рендера и показывают внутри canvas-области.
- `hooks.server.ts` проксирует `/api/v1/*` на backend; 500 от backend возвращаются клиенту как JSON, а не как route-ошибка.

**Результаты верификации.**

- `npm run check` — 0 ошибок, 0 предупреждений.
- `npm run test:unit` — 112 файлов, 1014 тестов passed.
- `npx vitest run src/routes/error-page.spec.ts src/shared/utils/route-match.test.ts` — passed.

**Осталось / открытые вопросы.**

- Решение по языку: приложение — `en` по умолчанию, документация — Russian. Зафиксировано в `.windsurfrules`, `MASTER_PROMPT.md`, `MASTER_PROMPT_RU.md`.
- Playwright-регрессия full-viewport: реализована `frontend/tests/error-500-page.spec.ts` + `src/routes/test/500/+page.server.ts` (триггер `?trigger=500`).
- `ApiErrorDisplay.svelte` использует `server-error` для API-ошибок с кодом `INTERNAL_ERROR`.
- PR #36 с правками и доской: https://github.com/Killaret/knowledge-graph/pull/36.
- Ревью Claude Code: проверить все коммиты в окне 2026-09-11, включая `aff53f2`, `460e913`, `5c69aa3`, `bcf7b59`, `0d2655e`, `8808a1f`, `183521a`, `8d19daf`, `ff32ee0`, `21f5d8d`, `5acc40d`, `ff65307`, `af2f957`, `49c4e67` и все последующие до слияния.

## 21. Dependabot PR — итоговая разборка (#21–#32, #38–#40, #56–#63, #68, #70–#77), 2026-09-12

Все Dependabot-PR обработаны, кроме **#25**. Замёржены: #21–#23, #24, #26, #27, #28, #29, #30, #31, #38, #56, #57, #60, #61, #62, #63. Закрыты как дублирующие/устаревшие: #32 (дублирует #28), #39, #40, #58, #59 (вошли в консолидированный PR #68). После фикса CI (PR #78) пришла новая волна Dependabot-PR: #70–#77 и **#79** (`nltk`) — все смержены. **#85** (`go_modules`) — смержен. **#25** (`yake`) — отклонён в пользу замены на `keybert` (MIT) с лемматизацией.

| # | Область | Зависимость | С | По | Риск | Итог |
|---|---|---|---|---|---|---|
| #21 | CI Actions | `actions/setup-python` | 6 | 7 | Низкий | ✅ замёржен |
| #22 | CI Actions | `actions/setup-go` | 6 | 7 | Низкий | ✅ замёржен |
| #23 | CI Actions | `actions/checkout` | 5 | 7 | Низкий | ✅ замёржен |
| #62 | root npm | grouped (7 updates) | — | — | Средний | ✅ замёржен |
| #63 | frontend npm | grouped (17 updates) | — | — | Средний-высокий | ✅ замёржен после фиксов `eslint`/`jsdom`/покрытия; `test:unit` 1381/1381, `test:coverage` lines 83.62%, branches 70.04% |
| #24 | backend Go | `golang.org/x/net` | 0.52.0 | 0.58.0 | Средний | ✅ замёржен (merge-конфликт go.mod разрешён, `go test -p 1 ./...` pass) |
| #26 | backend Go | `go-redis/v9` | 9.14.1 | 9.22.0 | Средний | ✅ замёржен |
| #28 | backend Go | `testcontainers-go/modules/postgres` | 0.40.0 | 0.44.0 | Средний-высокий | ✅ замёржен (разрешён конфликт go.mod, `go test -p 1 ./...` pass) |
| #30 | backend Go | `pgvector-go` | 0.2.0 | 0.4.1 | Средний | ✅ замёржен (разрешён конфликт go.mod, `go test -p 1 ./...` pass) |
| #32 | backend Go | `testcontainers-go` | 0.40.0 | 0.44.0 | Средний-высокий | ❌ закрыт как дублирующий #28 |
| #56 | NLP Python | `httpx` | 0.25.2 | 0.28.1 | Низкий-средний | ✅ замёржен |
| #25 | NLP Python | `yake` | 0.4.8 | — | Средний | ❌ отклонён — вместо обновления до 0.7.3 будет замена на `keybert` (MIT) с лемматизацией (рус/англ); см. [`tasks/NLP-2-yake-replace-keybert-lemmatization.md`](tasks/NLP-2-yake-replace-keybert-lemmatization.md) |
| #27 | NLP Python | `python-dotenv` | 1.0.0 | 1.2.3 | Низкий | ✅ замёржен |
| #29 | NLP Python | `pydantic` | 2.5.2 | 2.13.5 | Средний | ✅ замёржен |
| #31 | NLP Python | `sentence-transformers` | 2.2.2 | 2.7.0 | Высокий | ✅ замёржен; пересчёт embeddings не потребовался |
| #57 | graph-service | `go-redis/v9` | 9.5.5 | 9.22.0 | Средний | ✅ замёржен |
| #60 | graph-service | `testify` | 1.11.1 | 1.12.1 | Низкий | ✅ замёржен |
| #61 | graph-service | `protobuf` | 1.34.2 | 1.36.12 | Средний | ✅ замёржен |
| #38 | graph-service | `pgx/v5` | 5.7.2 | 5.9.2 | Средний | ✅ замёржен |
| #68 | graph-service | консолидированный PR (#39/#40/#58/#59) | — | — | Средний-высокий | ✅ замёржен; обновлены `grpc`, `containerd`, `testcontainers-go`, `moby/go-archive v0.3.3` (fix GHSA-hfg8-hc9c-6c3h / CVE-2026-17106); в `allow-licenses` добавлен `LicenseRef-scancode-google-patent-license-golang` |
| #39 | graph-service | `grpc` | 1.67.0 | 1.83.2 | Средний-высокий | ❌ закрыт — вошёл в #68 |
| #40 | graph-service | `containerd` | 1.7.18 | 1.7.35 | Средний | ❌ закрыт — вошёл в #68 |
| #58 | graph-service | `testcontainers-go/modules/postgres` | 0.35.0 | 0.44.0 | Средний-высокий | ❌ закрыт — вошёл в #68 |
| #59 | graph-service | `testcontainers-go` | 0.35.0 | 0.44.0 | Средний-высокий | ❌ закрыт — вошёл в #68 |
|| #70 | frontend npm | `@humanfs/node` | 0.16.7 | 0.16.8 | Низкий | ✅ замёржен |
|| #71 | frontend npm | `@sveltejs/kit` | 2.59.0 | 2.70.3 | Средний | ✅ замёржен |
|| #72 | frontend npm | `brace-expansion` | 1.1.14 | 1.1.18 | Низкий | ✅ замёржен |
|| #73 | frontend npm | `vite` | 8.0.10 | 8.3.0 | Средний-высокий | ✅ замёржен |
|| #74 | frontend npm | `js-yaml` | 4.1.1 | 4.3.2 | Низкий | ✅ замёржен |
|| #75 | frontend npm | `svelte` | 5.55.5 | 5.57.0 | Средний | ✅ замёржен |
|| #76 | root npm | `brace-expansion` | 1.1.14 | 1.1.18 | Низкий | ✅ замёржен |
|| #77 | root npm | `postcss` | 8.5.14 | 8.5.28 | Низкий | ✅ замёржен |
|| #85 | backend Go | `go_modules` (`edwards25519`, `moby/go-archive`, `quic-go`) | — | — | Средний-высокий | ✅ замёржен — закрыты алерты GHSA-hfg8-hc9c-6c3h, GHSA-vvgj-x9jq-8cj9, GHSA-fw7p-63qq-7hpr |
|| #79 | NLP Python | `nltk` | 3.8.1 | 3.10.3 | Средний | ✅ замёржен — `allow-ghsas: GHSA-8mgp-746c-j5xp` принят и задокументирован; см. [`tasks/DEPENDABOT-79-nltk-vulnerability.md`](tasks/DEPENDABOT-79-nltk-vulnerability.md) |

**Порядок действий (выполнен):**

1. ✅ CI Actions (#21–#23) — squash-merge.
2. ✅ Root npm (#62) — squash-merge.
3. ✅ Frontend npm (#63) — fixed `eslint`/`jsdom`/coverage, squash-merge.
4. ✅ Backend Go (#24, #26, #28, #30); #32 закрыт.
5. ✅ NLP (#56, #27, #29, #31); #79 смержен; #25 отклонён — будет замена на `keybert` с лемматизацией.
6. ✅ Graph-service (#38, #57, #60, #61) — squash-merge; конфликтующие #39/#40/#58/#59 объединены в PR #68 и смержены.
7. ✅ CI fix (PR #78) — починен запуск Core Checks (`permissions:`, `environment:` для `run-name`, `smoke` env, `needs` для smoke).
8. ✅ Новая волна Dependabot (#70–#77), #79 (`nltk`) и #85 (`go_modules`) — все смержены после фикса CI.

**Следующий шаг:**
- **#25** (`yake`): отклонён — вместо обновления `yake` до 0.7.3 будет замена на `keybert` (MIT) с лемматизацией (рус/англ). PR #25 закрыт, `yake` остаётся 0.4.8. Подробности: [`tasks/DEPENDABOT-25-yake-license.md`](tasks/DEPENDABOT-25-yake-license.md), [`tasks/NLP-2-yake-replace-keybert-lemmatization.md`](tasks/NLP-2-yake-replace-keybert-lemmatization.md).
- **#79** (`nltk` 3.8.1 → 3.10.3): ✅ замёржен — добавлен `allow-ghsas: GHSA-8mgp-746c-j5xp`; остаточный риск принят и задокументирован. Подробности: [`tasks/DEPENDABOT-79-nltk-vulnerability.md`](tasks/DEPENDABOT-79-nltk-vulnerability.md).

## 22. Правки CI под PR #36, 2026-09-11

**NLP: таймаут из-за nvidia-колёс.**

- Симптом: `Core Checks / NLP Service Checks` падала по таймауту 10 минут, скачивая `torch==2.14.0` + `nvidia_cudnn` 553 МБ + `nvidia_cusparselt` 170 МБ + `nvidia_nccl` 216 МБ и др.
- Причина: `pip install -r requirements.txt` в `_core-checks.yml` брал последний `torch` с PyPI (CUDA-версия), тогда как `Dockerfile` всегда использует CPU-индекс.
- Исправление (`ff32ee0`): заменить `pip install -r requirements.txt` на `pip install --index-url https://download.pytorch.org/whl/cpu --extra-index-url https://pypi.org/simple -r requirements.txt` и увеличить `timeout-minutes` до 20.
- Результат: `NLP Service Checks` стала проходить за ~1 минуту.

**Smoke Tests: `SKIP_AUTH` без `APP_ENV=test`.**

- Симптом: `Smoke Tests` падает на шаге `Start backend for smoke tests` с `FATAL: SKIP_AUTH=true is only allowed when APP_ENV=test; current APP_ENV=development`.
- Причина: в `ci.yml` smoke-тесты стартуют backend и graph-service с `SKIP_AUTH=true`, но без `APP_ENV=test`.
- Исправление (`21f5d8d`): добавить `APP_ENV: test` в env для шагов `Start backend for smoke tests` и `Start graph-service for smoke tests`.

**Smoke Tests: graph-service REDIS_URL в формате URL.**

- Симптом: `Start graph-service for smoke tests` падает с `failed to connect to redis: dial tcp: address redis://localhost:6379: too many colons in address`.
- Причина: graph-service ожидает `RedisURL` как `host:port`, а `ci.yml` передавал `redis://localhost:6379`.
- Исправление (`5acc40d`): `REDIS_URL: localhost:6379` для graph-service.

**Smoke Tests: Go module cache.**

- Симптом: `Start backend for smoke tests` провисел более 5 минут на `go mod download`, потому что в smoke-джобе не было `actions/setup-go` и кеша.
- Исправление (`ff65307`): добавлен `actions/setup-go@v6` с `cache-dependency-path: '**/go.sum'` в `smoke-tests`.

**Smoke Tests: мало времени на компиляцию сервисов.**

- Симптом: сервис падает по таймауту опроса `health` (30 попыток × 2 сек = 60 сек), потому что `go run` компилирует из исходников дольше минуты.
- Исправление (`af2f957`): увеличить цикл ожидания backend и graph-service до 90 попыток (до 3 минут).

**Smoke Tests: конфликт миграций, Redis URL и frontend URL.**

- Симптом: backend при `SKIP_AUTH=true` не мог поднять Redis (`redis://localhost:6379: too many colons in address`), не применял `019_add_test_user` из-за ручного `migrate up` + собственного `RunMigrations`, регистрация возвращала 500, а 49 тестов палили с `ERR_CONNECTION_REFUSED` на `http://127.0.0.1:5173`.
- Исправление (`49c4e67`):
  - убрать ручной `Apply database migrations` из smoke-тестов, чтобы backend сам применил SQL-миграции;
  - `REDIS_URL: localhost:6379` для backend smoke;
  - `FRONTEND_URL`, `VITE_API_TARGET`, `VITE_GRAPH_SERVICE_URL` и BDD-URL переключены на `localhost`.

- Статус: все вспомогательные правки в `ci.yml`; следующий прогон CI подтверждает.

**Smoke Tests: отсутствует тестовый пользователь после миграции 029.**

|- Симптом: после исправления миграций Playwright-создание заметок падает с `Failed to save note`; в логе Postgres `insert or update on table "notes" violates foreign key constraint "notes_creator_id_fkey"` для `creator_id=00000000-0000-0000-0000-000000000000`.
|- Причина: миграция `029_remove_test_user.up.sql` удаляет zero-UUID пользователя, созданного `019_add_test_user.up.sql`; `skip_auth` middleware всё ещё использует этот ID, а сидер `backend/cmd/seed` не запускался в smoke-джобе.
|- Исправление (`5f0f6b8`): после подъёма backend запустить `go run ./cmd/seed` с `APP_ENV=test` и `SEED_TEST_USER_PASSWORD` до начала тестов.
|- Результат: 49 Playwright smoke-тестов и 43 BDD-шага (5 сценариев) проходят.

**Smoke Tests: BDD-сценарии проходят, но шаг зависает на выходе.**

|- Симптом: `Run smoke BDD tests` отрапортовал `5 scenarios (5 passed), 43 steps (43 passed), 0m49.839s`, но процесс не завершился и джоба ушла в timeout/cancel.
|- Причина: `frontend/tests/features/support/hooks.ts` запускает Vite dev-сервер сам, а `devServer.kill` не убивает дочерний процесс Vite; процесс остаётся висеть, GitHub Actions не переходит к следующему шагу.
|- Исправление (`c7f4786`):
  - выделить отдельный шаг `Start frontend dev server for smoke tests` с `npm run dev &` и дождаться `http://localhost:5173`;
  - `PLAYWRIGHT_DEV_SERVER=true` + `webServer.reuseExistingServer: true` заставляют Playwright переиспользовать уже поднятый Vite, а не стартовать новый;
  - `test:cucumber` видит готовый сервер и не стартует собственный, поэтому завершается сразу после отчёта;
  - `SKIP_AUTH: "true"` добавлен в env `Run smoke BDD tests`, чтобы `hooks.ts` инжектировал `__SKIP_AUTH__`.
|- Результат: `Smoke Tests` проходит за ~5m35s; полный CI run `34642092163` — success (все 9 джоб).

**Итог CI-4:**

|- PR #36 run `34642092163` — `conclusion: success`, все Core Checks и Smoke Tests зелёные; Playwright 49 passed/2 skipped, BDD 5 scenarios/43 steps passed.

## 23. CI-5: починка запуска Core Checks и волна Dependabot PR #70–#77, 2026-09-12

**Контекст.** После мёрджа PR #63 и ряда Dependabot-обновлений в `main` новые PR стали застревать на стадии `Core Checks`: джоба не могла запустить ни одного шага из-за ошибки `Error when evaluating 'strategy' for job 'build-matrix'` / `Error parsing called workflow` и аналогичных синтаксических/контекстных ошибок. Параллельно пришла новая волна Dependabot-PR (#70–#77), которую нельзя было проверить и смержить, пока CI не работал.

**PR #78 — `ci: fix Core Checks startup and pre-existing failures`.**

- Исправлен запуск `Core Checks` (`_core-checks.yml`):
  - `run-name` вынесен на уровень `workflow` вместо `job`, убрана ссылка на `inputs` в `run-name`.
  - `permissions:` добавлены/исправлены: `contents: read` и `checks: read` для `reusable_workflow_call`.
  - `environment:` убран из `job`-level, оставлен для шагов, которым он действительно нужен.
  - `needs:` у `Smoke Tests` (`ci.yml`) теперь корректно ссылается на джобы `Core Checks` / `Security Audit`.
- Исправлена совместимость NLP после `httpx 0.28.1`:
  - `frontend/package-lock.json` обновлён (`ky` hooks state, `prefixUrl` → `prefix`).
  - `nlp-service/requirements.txt`: `fastapi` поднят до совместимой с `httpx 0.28` версии, `uvicorn` синхронизирован.
- Форматирование: Prettier применён к 20 frontend test/spec файлам.
- Локальная верификация:
  - `npm run format:check` — чисто.
  - `npm run check` — 0 errors, 0 warnings.
  - `npm run lint` — 0 errors, 9 pre-existing warnings.
  - `npm run test:unit -- --run` — 1381/1381 passed.
  - `npm run test:coverage` — lines 83.62%, statements 81.9%, functions 81.89%, branches 70.04% — все выше порога 70%.
- CI run `34710120173` / `34710117628` — `conclusion: success`: CodeQL, Dependency Review, Frontend Security Audit, Core Checks (Backend, Frontend, Graph, NLP, Integration), Docker Compose, Smoke Tests, `test` — все зелёные.
- PR #78 замёржен в `main`.

**Новая волна Dependabot PR #70–#77.**

После фикса CI Dependabot поднял свежие PR:

|| # | Область | Зависимость | С | По | Итог |
|---|---|---|---|---|---|---|
|| #70 | frontend npm | `@humanfs/node` | 0.16.7 | 0.16.8 | ✅ замёржен |
|| #71 | frontend npm | `@sveltejs/kit` | 2.59.0 | 2.70.3 | ✅ замёржен |
|| #72 | frontend npm | `brace-expansion` | 1.1.14 | 1.1.18 | ✅ замёржен |
|| #73 | frontend npm | `vite` | 8.0.10 | 8.3.0 | ✅ замёржен |
|| #74 | frontend npm | `js-yaml` | 4.1.1 | 4.3.2 | ✅ замёржен |
|| #75 | frontend npm | `svelte` | 5.55.5 | 5.57.0 | ✅ замёржен |
|| #76 | root npm | `brace-expansion` | 1.1.14 | 1.1.18 | ✅ замёржен |
|| #77 | root npm | `postcss` | 8.5.14 | 8.5.28 | ✅ замёржен |
|| #85 | backend Go | `go_modules` (`edwards25519`, `moby/go-archive`, `quic-go`) | — | — | ✅ замёржен — закрыты GHSA-hfg8-hc9c-6c3h, GHSA-vvgj-x9jq-8cj9, GHSA-fw7p-63qq-7hpr |
|| #79 | NLP Python | `nltk` | 3.8.1 | 3.10.3 | ✅ замёржен — `allow-ghsas: GHSA-8mgp-746c-j5xp` |

- Каждый PR был обновлён до актуального `main` и прогнан с фиксом CI.
- Все checks (`Analyze`, `Core Checks`, `Dependency Review`, `Frontend Security Audit`, `test`, `Smoke Tests`) — зелёные.
- Все PR смержены через squash.

**Состояние на 2026-09-12.**

- Открытые Dependabot-PR: **#25** (`yake`) — отклонён в пользу замены на `keybert` (MIT) с лемматизацией; **#79** (`nltk` 3.8.1 → 3.10.3) — смержен с `allow-ghsas: GHSA-8mgp-746c-j5xp`.
- Frontend coverage после всех обновлений: **1381/1381 unit-тестов passed**, lines 83.62%, statements 81.9%, functions 81.89%, branches 70.04% — выше 70%.
- `npm run check` — 0 errors, 0 warnings; `npm run lint` — 0 errors, 9 pre-existing warnings.

## 24. NOTE-TYPE-TAXONOMY и DDD `NoteType` value object (2026-09-14)

**Контекст.** Типы заметок (`galaxy`, `nebula`, `blackhole`, `star`, `planet`, `moon`, `comet`, `satellite`, `asteroid`, `dust`, `debris`, плюс системные/аномалии) были рассогласованы между backend, frontend и OpenAPI. `moon` был известен домену, но отсутствовал в пользовательском селекторе; порядок в списках шёл не по космической иерархии; `blackhole` спорно располагался ниже `star`; дефолтный тип зависел от `types[0]`.

**Решения владельца:**
- Единая шкала `scaleRank` от `galaxy` (100) к `debris` (5).
- `blackhole` выше `star` (массивнее и иное смысловое наполнение).
- `moon` включается в пользовательский UI.
- `debris` ниже `dust`; `dust` остаётся для быстрых захватов/инбокса.
- Дефолтный тип при создании заметки — `star`.

**Реализация Devin (ветка `devin/batch-api-37`):**
- `backend/internal/domain/note/type.go` — `NoteType` value object с `scaleRank`, `IsUserSelectable`, валидацией, `DefaultNoteType()`.
- `backend/internal/domain/note/entity.go` — `Note` хранит `NoteType`; конструкторы и `SetType` принимают `NoteType`.
- `backend/internal/interfaces/api/notehandler/note_handler.go` — `resolveNoteType` через `note.NewType`; `validateResolvedNoteType` удалён.
- `backend/internal/interfaces/api/common/validation/validators.go` — `IsValidCelestialBodyType` делегирует `note.NewType`.
- `backend/internal/infrastructure/db/postgres/note_repo.go` — `toDomainNote` преобразует строки БД в `NoteType`; пустые legacy-значения мапятся в `star`.
- `backend/internal/application/import/service.go` — импорт преобразует типы через `note.NewType`; дефолт `asteroid` сохранён.
- `backend/openAPI.yaml` — все note-type enum приведены к единому каноническому порядку.
- `frontend/src/entities/shared/model/celestial-body.ts` — `scaleRank` для всех типов; `UI_TYPES` включает `moon` и сортируется по `scaleRank`; `ALL` в каноническом порядке.
- `frontend/src/components/molecules/TypeSelector.svelte` — `defaultSelected` ищет `star`, а не `types[0]`.
- `CreateNoteModal`, `NoteForm`, graph-формы, импорт закладок, фильтры графа/home page — все используют `CelestialBody.UI_TYPES`.
- Тесты: `note/type_test.go`, `celestial-body.test.ts`, `CreateNoteModal.spec.ts`, `EditNoteModal.spec.ts`, `GraphCanvas.events.spec.ts`, `home-page.svelte.test.ts`.

**Верификация:**
- `cd backend && go test ./...` — зелёное.
- `cd backend && go vet ./...` — чисто.
- `cd backend && go test ./cmd/server/...` — контрактный тест проходит.
- `cd frontend && npm run test:unit -- --run` — 1381/1381 passed.
- `cd frontend && npm run build` — успешно.
- `cd frontend && npm run check` — 0 errors, 0 warnings.

**Статус:** реализация готова, передана на ревью Claude Code. Не мержить без ревью.

## 25. URL-заголовки, quality loop, события/напоминалки, архив, CI/DEPLOY — передача Claude (2026-09-14)

**Контекст.** За 13–14 сентября владелец и Devin обсуждали и частично прототипировали пять крупных тем. Результаты зафиксированы в task-файлах, `docs/AI_HANDOFF.md` и `docs/AI_LOG.md`; вся очередь вынесена на ревью/обсуждение Claude Code.

### 25.1 URL-HEADING-1: извлечение `title`/`content` из h1–h6

- Постановка: `docs/tasks/URL-HEADING-1-heading-extraction.md`; полный отчёт пробного прогона: `docs/tasks/URL-HEADING-1-findings-probe.md`.
- Прототип на 22 URL из `bookmarks_11.09.2026.html` (tech_doc, course, russian, complex) подтвердил, что `h1` внутри `<main>`/`<article>` обычно лучше `<title>`, а `h2`–`h6` дают осмысленный outline.
- Главные открытые вопросы:
  - title-кандидаты: `<title>`, один/несколько `h1`, fallback по URL;
  - фильтрация шума по class/id (`promo`, `news`, `related`, `subscribe`, `comments`) и layout/grid;
  - обработка `mw-parser-output`/`documentation`/`content`, где нет `<main>`;
  - `<title>` часто содержит сайтовые суффиксы (`— Википедия`, `| OneLogin Developers`);
  - 401/404 ответы требуют graceful fallback.
- Результат: передано на ревью Claude Code; production-код `ImportFetcher.Extract` пока не менялся.

### 25.2 NOTE-QUALITY-1: цикл качества, health, duplicate review

- Постановка: `docs/tasks/NOTE-QUALITY-1-quality-loop.md`.
- Решения владельца:
  - без cron; обработка при создании/импорте и по ручной кнопке "улучшить заметку";
  - пользовательские метки могут пускать заметку в тот же цикл;
  - не ходить по сети к source URL;
  - не удалять и не мержить дубликаты автоматически;
  - reclassification под контролем пользователя;
  - шкала low/medium/high, динамические веса, stopping criteria.
- Открытые вопросы: конкретные критерии, веса, векторное сравнение title и контента, duplicate-review UI.

### 25.3 COMET-1: события и напоминалки

- Постановка: `docs/tasks/COMET-1-event-reminder-fields.md`.
- Нужно выбрать тип: `comet`, `satellite` или новый `reminder`.
- Кандидатные поля: `event_at`, `timezone`, `reminder_at`, `recurrence_rule`, `duration`, `location`, `attendees`.
- Возможна delayed-задача `reminder:send` через `asynq`.

### 25.4 ARCHIVE-1: сохранение обсуждений и постановок

- Постановка: `docs/tasks/ARCHIVE-1-discussion-history.md`.
- Требование владельца: постановки и обсуждения — это история проекта и доказательство работы с агентами.
- Варианты: git-история, `docs/archive/`, отдельный репо/ветка, GitHub Releases.
- Нужно решить: какие документы включать, как санировать персональные URL/данные, периодичность.

### 25.5 CI-MAIN-1 / DEPLOY-1: починка Main Branch CI/CD и Production Deployment

- Workflow `main.yml` и `deploy.yml` починены:
  - `backend/Dockerfile` — `--target server` и `--target worker`;
  - `127.0.0.1`, `APP_ENV=test`, `JWT_SECRET` ≥32 chars, `.env` для Docker Compose;
  - Cucumber `BeforeAll` hang устранён через таймауты и `127.0.0.1`;
  - seed test user, health checks и service discovery.
- Run IDs: Main Branch CI/CD `34758935365`, Production Deployment `34758935316` (main) и `34758941321` (ai-agents) — зелёные.
- Статус: Devin реализовал и верифицировал; передано на финальное ревью/подпись Claude Code.

### 25.6 Прочее, требующее внимания Claude

- **WSL-SWAP:** обсудить перенос/отключение `D:\wsl-swap\swap.vhdx` (309 МБ, max 8 ГБ) и риски OOM.
- **HOUSEKEEPING-1:** ветки `devin/*` и `security/findings` смержены/удалены; оставлены `main`, `ai-agents`, `java-source-text-handler`; `main` и `ai-agents` синхронизированы.
- **PLAYWRIGHT-ARTIFACTS-1:** в корне 6 untracked директорий `*-chromium-skip-auth-retry*/` и `debug-note-page.png` (удалён) — нужно либо удалить, либо добавить в `.gitignore`.
- **GITHUB-SECURITY-1:** GitHub Dependabot показывает 3 новые находки на `main` (1 high, 2 low); нужен triage и план.

**Статус:** всё передано на ревью/обсуждение Claude Code.

## 26. Test stack JWT, PowerShell check runner, publish API documentation (2026-09-14)

### 26.1 ENV-1 — `JWT_SECRET` for clean-machine test stack

- `docker-compose.test.yml`: `${JWT_SECRET:-test-jwt-secret-32-characters-long}` in backend, worker and e2e.
- `scripts/testing/start-test.ps1` and `start-test.sh` load `.env.test` without overwriting exported process variables, then set the default if `JWT_SECRET` is still empty.
- `.env.test.example` and `.gitignore` added so a real `.env.test` stays local.
- Verified: test stack starts on a machine with no manual `.env`, services become healthy, only expected Yandex OAuth warnings remain.

### 26.2 CHECK-ALL-1 — PowerShell 5.1 phase tracking

- `scripts/testing/lib/phase-tracking.ps1` converted to ASCII-safe output (removed UTF-8 em-dash without BOM).
- `check-all.ps1` now fails loudly when `phase-tracking.ps1` cannot be dot-sourced and when required functions are missing.
- Added a non-ASCII/BOM guard for `scripts/**/*.ps1`.
- Verified by mutation: a deliberately broken phase prints `[FAIL]` and the final exit code is non-zero; a clean full run of `check-all.ps1` (without `-Quick`) is PASS, only `golangci-lint` skipped because the tool is not installed.

### 26.3 NOTE-PUBLISH-DOC — publish path in API guide

- `docs/api/API_EN.md` explicitly documents `POST /api/v1/notes/{id}/publish` and `POST /api/v1/notes/{id}/unpublish`.
- Notes are created private; `PUT /notes/{id}` does not accept `is_public` or `source_url` (the DTO and `UpdateNoteRequest` schema no longer include them).
- `node scripts/testing/check-docs-links.mjs .` passes.

**Статус:** реализация выполнена, передана на ревью Claude Code.

## 27. DECISIONS-1 — сторож указателя решений владельца (2026-09-14)

### 27.1 Указатель

- `docs/DECISIONS.md` дополнен до 29 записей; добавлены 7 ранее неиндексированных решений: `BOARD-1`, `DEPENDABOT-25`, `DEPENDABOT-79`, `NOTE-QUALITY-1`, `NOTE-TYPE-TAXONOMY`, `P11-1`, `SECURITY-1`.
- Решения без кода (`NOTE-QUALITY-1`, `P11-1`) помечены `(кода не требует)`.
- Исправлен парсинг идентификаторов, начинающихся с цифр (`P11-1`).

### 27.2 Сторож

- `scripts/testing/check-decisions.mjs` реализует 4 правила: резолв ссылок, соответствие маркеров указателю, след кода по идентификатору в коммите, архив терминальных строк старше трёх дней.
- Интегрирован в `check-all.ps1` через `core-checks.tsv` (`decisions`) и в CI `frontend-checks` (`Check decision index`).
- `check-core-workflow-sync.mjs`: 18 local phases match 18 CI steps.

### 27.3 Протокол

- В `docs/AI_AGENT_PROTOCOL.md` добавлено правило: коммит реализации называет идентификатор задачи в заголовке или теле; исключения — `(кода не требует)`.

### 27.4 Мутации

- Все 4 мутации пройдены: решение без записи в указателе, битая ссылка, решение без коммита, устаревшая терминальная строка. Выводы приложены в `docs/tasks/DECISIONS-1-decision-index-and-guard.md`.

### 27.5 Верификация

- `check-all.ps1` без `-Quick`: 17 PASS, 1 SKIP (`golangci-lint`), exit 0.
- `node scripts/testing/check-decisions.mjs .` PASS.

**Статус:** реализация выполнена, передана на ревью Claude Code.

## 28. BOARD-1 — ретенция доски и сторож размера (2026-09-14)

### 28.1 Правила

- Правило ретенции распространено и на реплики: `docs/AI_HANDOFF.md` (шапка), `docs/AI_AGENT_PROTOCOL.md` (таблица обмена), `.claude/commands/kg-work.md`, `.devin/skills/kg-work/SKILL.md`.
- Описан формат реплики: 3–4 предложения, указатель, ссылка на `docs/tasks/<id>-review-findings.md`.
- Раздел «Решения владельца» из `docs/AI_HANDOFF.md` перенесён в `docs/DECISIONS.md` без изменений; в доске оставлена ссылка. `docs/DECISIONS.md` добавлен в список обязательного чтения `CLAUDE.md`.

### 28.2 Сторож

- `scripts/testing/check-board-size.mjs` проверяет, что `docs/AI_HANDOFF.md` не превышает 120 КБ (решение владельца 2026-09-14; было 40 КБ).
- Интегрирован в `core-checks.tsv`, `_core-checks.yml` и `check-all.ps1`/`check-all.sh` через `frontend-checks` (`Check board size`).
- `check-core-workflow-sync.mjs`: 19 local phases match 19 CI steps.

### 28.3 Верификация

- `node scripts/testing/check-board-size.mjs .` — FAIL: `AI_HANDOFF.md` ~208.0 КБ, порог 40 КБ. Это ожидаемое состояние: содержимое доски не чистил, первая чистка по постановке за владельцем.
- `check-all.ps1` без `-Quick`: 17 PASS, 1 SKIP, 1 FAIL (board size); остальные фазы зелёные, включая backend и graph integration. В сообщении сторожа исправлена опечатка `replics` → `replies`.

## 29. PUB-3 — переименование публичного эндпоинта графа (`all` → `public`) (2026-09-14)

### 29.1 Что изменено

- `backend/cmd/server/router.go`: маршрут переименован из `all` в `public`.
- `backend/internal/interfaces/api/middleware/jwt.go`: `SkipPaths` обновлён на `/api/v1/graph/public`.
- `backend/openAPI.yaml`: путь и summary (`Get public graph`).
- Go-интеграционные тесты (`graphhandler/*_test.go`) используют `/graph/public`.
- Фронтенд/E2E/Playwright тесты (`frontend/src/shared/api/graph.test.ts`, `frontend/tests/preload-full-cycle.spec.ts`, `frontend/tests/public-graph-real-auth.spec.ts`, `tests/e2e/api-contract.spec.ts`) обновлены на `/graph/public`.
- Скрипты проверки стеков и регресса (`scripts/ci/check-stacks-health.*`, `scripts/testing/run-full-test-cycle.ps1`) обновлены.
- Активная документация (`docs/api/API_EN.md`, `docs/api/API_ERRORS_EN.md`, `docs/operations/CONFIGURATION_EN.md`, `docs/operations/DOCKER.md`, `docs/architecture/GRAPH3D.md`, `docs/LINK_TYPES*.md`, `docs/archive/MANUAL_TEST_CHECKLISTS_RU.md`, `docs/archive/API_TEST_COVERAGE_PLAN.md`, `docs/product/BACKLOG.md`, `docs/assets/graph-loading-flow.*`, `docs/DECISIONS.md`, `CHANGELOG.md`) приведена в соответствие.
- Постановки и review-findings (`docs/tasks/PUB-3-rename-graph-endpoints.md`, `PUB-2-graph-view-mode.md`, `PUB-2-review-findings.md`, `API-1-openapi-contract-and-handover.md`, `SPECS-1-review-findings.md`, `AUD-2-*`) обновлены.

### 29.2 Живая верификация

- Тест-стек поднят, данные засеяны (20 публичных заметок).
- `curl -s -D - http://127.0.0.1:18083/api/v1/graph/public?limit=1` → `HTTP/1.1 200 OK` (анонимно, `Cache-Control: private, max-age=300`).
- `GET /api/v1/graph/{old-public}?limit=1` → `HTTP/1.1 404 Not Found`.

### 29.3 Поиск остатков

- Активный код, тесты, скрипты и документация больше не содержат действующих ссылок на старый публичный путь.
- Оставшиеся совпадения ограничены историческими/спецификационными документами: `docs/architecture/decisions/018-public-graph-access-model.md` (ADR), `docs/archive/EXTERNAL_AUDIT_2026-09.md` (снапшот аудита), `docs/archive/TEST_EXECUTION_REPORT.md` (архив).

### 29.4 Верификация

- `go test ./internal/interfaces/api/graphhandler/... -v` — PASS.
- `go test ./internal/interfaces/api/graphhandler/... -v -tags=integration` — PASS.
- `npm run test:unit -- src/shared/api/graph.test.ts` — 41/41 PASS.
- `check-all.ps1` без `-Quick`: 19 PASS, 1 SKIP (`golangci-lint` не установлен), exit 0.

**Статус:** реализация выполнена, передана на ревью Claude Code.

## 30. AUTHOR-1 — сторож авторства коммитов (2026-09-14)

### 30.1 Что изменено

- `scripts/testing/check-commit-authorship.mjs`: читает `docs/AI_AGENT_PROTOCOL.md`, проверяет `main..HEAD` (или CI range), ловит коммит, у которого `Co-Authored-By:` — агент, а автор — другой.
- `scripts/testing/core-checks.tsv` и `.github/workflows/_core-checks.yml`: добавлена фаза `commit-authorship` (`Check commit authorship`) в `frontend-checks`.
- `scripts/testing/check-agent-session-tree.mjs`: не позволяет начать сессию агента, если `git status --porcelain` не пуст. Это второй сигнал AUTHOR-1: убирает условие, при котором чужая работа попадает в чужой коммит.
- `.devin/skills/kg-work/SKILL.md` и `docs/AI_AGENT_PROTOCOL.md`: шаг 0 `/kg-work` теперь требует чистого дерева.
- `docs/AI_LOG.md` и `docs/tasks/AUTHOR-1-commit-authorship-guard.md` дополнены исправлением атрибуции и итогом.

### 30.2 Атрибуция

- Первая реализация сторожа трейлеров (`check-commit-authorship.mjs`, строка в `core-checks.tsv`, шаг в `_core-checks.yml`) была написана Devin, но попала в коммит `d3f3e17`, автором в git указан Claude Code, потому что Claude Code закоммитил широким захватом поверх незакоммиченного дерева Devin. Это иллюстрация дыры, описанной в дополнении постановки. История не переписывается: коммит в `origin/ai-agents`. Исправление записано в `docs/AI_LOG.md`.

### 30.3 Мутации

1. Claude Opus 5 + `Co-Authored-By: Devin` → FAIL (нарушение).
2. Devin + `Co-Authored-By: Devin` → OK.
3. Человек без трейлера → OK.

Все три мутации отработали и откачены `git reset --hard`; история не изменилась.

### 30.4 Верификация

- `node scripts/testing/check-commit-authorship.mjs .` — `Commit authorship OK for 33 commit(s) in main..HEAD.`
- `node scripts/testing/check-core-workflow-sync.mjs` — `Workflow sync OK: 20 local phases match 20 CI steps.`
- `check-all.ps1` без `-Quick` — 20 PASS, 1 SKIP (`golangci-lint`), exit 0. В том числе новая фаза `Commit authorship guard`.

**Статус:** реализация выполнена, передана на ревью Claude Code.

## 31. BACKUP-2 — canonical backup directory

**Коммит:** `b7070c8`.

### 31.1 Задача

Единый источник каталога бэкапа Personal-стека: скрипты, сторож и compose должны смотреть в `Desktop\\my items`, а не в `<repo>/backups`.

### 31.2 Что изменилось

- `scripts/devops/backup-policy.env` теперь содержит `KG_BACKUP_DIR=Desktop/my items` и `KG_BACKUP_MAX_AGE_HOURS=24`.
- `backup-personal.{ps1,sh}`, `check-personal-backup.{ps1,sh}`, `guard-personal-data.py` и `docker-compose.personal.yml` читают `KG_BACKUP_DIR`.
- Относительный путь разворачивается от домашнего каталога; абсолютный (`C:/...`, `/...`, `\\\\...`) остаётся без изменений.
- `check-personal-backup.sh` починен: убран сломанный `REPO_ROOT` (`A || B && C`), исправлена тильда-развёртка, которая дублировала `$HOME` на путях вида `/c/Users/...`.
- `guard-personal-data.py` теперь использует `resolve_backup_dir()` вместо хардкода `<repo>/backups`.
- Тесты сторожа (`test_guard_personal_data.py`) расширены с 36 до 39 assertions.
- Подключены в `core-checks.tsv` и `_core-checks.yml` как фаза `Personal data guard tests`.

### 31.3 Мутации

1. Свежий бэкап → `[PASS]` (PowerShell, shell, guard).
2. Состаренный тот же файл → `[ERROR] ... h old` (PowerShell, shell, guard deny).
3. Пустой каталог → `[ERROR] No backup directory` (PowerShell, shell, guard deny).
4. Нулевой свежий файл → `[ERROR] No non-empty backup` (PowerShell, guard deny; shell — то же).
5. Shell до/после: до — «No backup directory» при полном каталоге; после — тот же `[ERROR]`, что и PowerShell.

### 31.4 Верификация

- `python scripts/devops/test_guard_personal_data.py` — `All checks passed (39 assertions)`.
- `bash -n scripts/devops/backup-personal.sh` / `check-personal-backup.sh` — зелёно.
- `node scripts/testing/check-core-workflow-sync.mjs` — `21 local phases match 21 CI steps`.
- `check-all.ps1 -Quick` — 21 PASS, 3 SKIP (`golangci-lint`, backend integration, graph integration), exit 0.

**Статус:** реализация выполнена, передана на ревью Claude Code.

## 32. DOC-SYNC-2 — Adversarial Phase norm mirrored

**Коммит:** `5376f9c`.

Раздел `.windsurfrules` **Adversarial Phase (MANDATORY for new surfaces)** зеркалирован в:
- `.devin/skills/knowledge-graph/SKILL.md`
- `.devin/prompts/MASTER_PROMPT.md`
- `.devin/prompts/MASTER_PROMPT_RU.md`

Также убран терминальный ряд доски `CI-4` (дата 2026-09-11), который превысил 3 дня и вызывал `check-decisions` FAIL.

**Статус:** на ревью Claude Code.

## 33. REG-2 — интеграционный тест пакетной близости

### 33.1 Проблема

`FindSimilarNotesBatch` используется в `embedding_loader.go` и `gamma_link_generator.go`, но до правки ни один тест не выполнял его SQL на настоящей pgvector-базе: все существующие тесты мокали репозиторий.

### 33.2 Что изменено

- `backend/internal/infrastructure/db/postgres/embedding_repo.go`: параметр `[]uuid.UUID` теперь передаётся через `pq.Array([]string{...})` в оператор `ANY(?)`. Сам SQL не менялся; сырой срез GORM разворачивал в несколько плейсхолдеров, что приводило к `ERROR: syntax error at or near ","`.
- `backend/internal/infrastructure/db/postgres/embedding_repo_test.go`: новый интеграционный тест `TestEmbeddingRepository_FindSimilarNotesBatch` (`//go:build integration`) проверяет:
  - оценку в `[0, 1]` и хотя бы одну ненулевую;
  - один результат на каждый запрошенный `id`, у которого есть соседи;
  - пустой результат для заметки без эмбеддинга;
  - лимит именно на каждый `id`, а не на выдачу целиком;
  - фильтрацию по `model_name` (заметка со старой моделью не просачивается).

### 33.3 Данные теста

- Источники: `6000...`, `7000...`.
- Цели: `1000...` (близкая), `2000...` (далёкая, `cosine distance ≈ 2`), `3000...` (близкая, не влезает в лимит), `4000...` (старая модель), `5000...` (без эмбеддинга).
- `limit = 2`, кандидатов 3: каждый источник получает ровно 2 результата.

### 33.4 Мутации

1. `as score` → `as similarity`: `FindSimilarNotesBatch` падает с `ERROR: column "score" does not exist` (потому что `ORDER BY score DESC` ссылается на алиас; поле `Score` структуры не заполняется).
2. Убрать `GREATEST/LEAST`: тест падает с `score -1 out of [0, 1]` и `expected far score 0.0 after clamping, got -1`.
3. Глобальный `LIMIT ?` в SQL: тест падает с `source 7000...: missing from batch results` и `expected 2 results, got 0`.

Все три мутации откачены.

### 33.5 Верификация

- `go test -count=1 -tags=integration -run TestEmbeddingRepository_FindSimilarNotesBatch ./internal/infrastructure/db/postgres/...` — PASS (`TEST_DATABASE_URL` на тест-стек, testcontainers не используется).
- `go test ./...` (без тега `integration`) — PASS.
- Документация по поведению `FindSimilarNotesBatch` в `docs/` отсутствует; дополнительных документов не требовалось.

**Статус:** на ревью Claude Code.

## 34. TASKS-INDEX-1 — генератор указателя постановок

### 34.1 Что изменено

- `scripts/testing/generate-tasks-index.mjs`: генерирует `docs/tasks/README.md` из `docs/tasks/*.md`, `docs/AI_HANDOFF.md` и `docs/AI_LOG.md`; идентификатор выбирается по самому длинному известному доске/журналу префиксу файла; поддерживает суффиксы `AUD-7a`/`AUD-7b`; ссылки статуса нормализуются от `tasks/X.md` к `X.md`.
- `scripts/testing/check-tasks-index.mjs`: сторож дрейфа — ловит файлы без известного идентификатора, расхождение между каталогом и `docs/tasks/README.md`, ссылки доски/журнала на отсутствующие файлы.
- `scripts/testing/core-checks.tsv` и `.github/workflows/_core-checks.yml`: добавлена фаза `tasks-index` в `frontend-checks`.
- Переименованы 10 файлов, чьи имена не начинались с доски/журнала: `BATCH-1-api-design.md`, `BATCH-DDD-1-validation.md`, `BATCH-TEST-1-strategy.md`, `AUD-1-review-findings.md`, `DOCS-LINKS-1-review-findings.md`, `SPECS-1-review-findings.md`, `VERIFY-FINDING-MIRROR-1-review-findings.md`, `NLP-2-yake-replace-keybert-lemmatization.md`, `PROJECT-SKILLS-1-review-findings.md`. Три файла из исходного списка 13 (`AUD-7a-enforce-boundaries.md`, `AUD-7b-lint-tests-and-coverage-denominator.md`, `NOTE-TYPE-TAXONOMY.md`) после исправления парсера оказались корректными.
- Установлены идентификаторы в `docs/AI_LOG.md` и `docs/AI_HANDOFF.md` для `PROJECT-SKILLS-1`, `DOCS-LINKS-1`, `SPECS-1`, `VERIFY-FINDING-MIRROR-1`, `DEPENDABOT-79`, `MONGO-1`.

### 34.2 Мутации

1. `MUTATION-1-temp.md` без обновления указателя → `check-tasks-index.mjs` FAIL: `Unknown identifier: MUTATION-1-temp.md` + drift.
2. `mutation-no-id.md` без идентификатора → FAIL: `Unknown identifier: mutation-no-id.md`.
3. Строка доски со ссылкой `tasks/NONEXISTENT-999.md` → FAIL: `docs\AI_HANDOFF.md links to missing task file: NONEXISTENT-999.md`.
4. Удаление всех мутаций и повторная генерация → `Task index OK: 89 entries, no drift, no broken board links.`

### 34.3 Верификация

- `node scripts/testing/generate-tasks-index.mjs .` — `Generated 89 task index entries at docs\tasks\README.md.`
- `node scripts/testing/check-tasks-index.mjs .` — `Task index OK: 89 entries, no drift, no broken board links.`
- `node scripts/testing/check-docs-links.mjs .` — `Docs OK`.
- `node scripts/testing/check-core-workflow-sync.mjs` — число фаз сходится с CI.

**Статус:** на ревью Claude Code.

## 35. GORDON-1 — внешние документы Gordon и полный локальный прогон

### 35.1 Что изменено

- Три внешних AI-документа, ранее лежавших в корне (`DEPLOYMENT_OPTIMIZATION_MAP.md`, `PROJECT_REVIEW_COMPREHENSIVE.md`, `docs/GORDON_ANALYSIS.md`), перенесены в `docs/archive/gordon/`.
- Создана постановка [`tasks/GORDON-1-gordon-documents-review.md`](tasks/GORDON-1-gordon-documents-review.md) — ожидает обзора и вердикта владельца.
- `docs/tasks/README.md` перегенерирован: 90 записей, `check-tasks-index.mjs` зелёный.

### 35.2 Верификация

- Полный `scripts/testing/check-all.ps1` (не `-Quick`) прошёл: **21 PASS**, **1 SKIP** (`golangci-lint` не установлен), **exit 0**.
- Frontend unit tests: 139 файлов / 1430 тестов PASS.
- Backend coverage: 66.4% (порог на тот момент 64.8%).
- Task index, decision index, board size, commit authorship, documentation links — все зелёные.

**Статус:** GORDON-1 ждёт человека.

## 36. DEPLOY-2 — публикация Docker-образов из CI

### 36.1 Что изменено

- `.github/workflows/deploy.yml` публикует 5 образов на Docker Hub после успешной проверки deploy-стека:
  - тег `YYYY-MM-DD-<short-sha>`;
  - тег `main`, который двигается на каждый зелёный push в `main`.
- Публикация только из `main`; `ai-agents` проверяется, но не публикует.
- При отсутствии `DOCKER_USERNAME`/`DOCKER_PASSWORD` шаг публикации пропускается с сообщением, workflow остаётся зелёным.
- `deploy` job перестал быть `echo`: теперь он верифицирует манифесты опубликованных образов и имеет `environment: production`.
- `docker-compose.deploy.yml` по умолчанию использует `main` вместо застывшего `2026-09-08`, с комментарием про семантику тегов.
- `docs/api/API_EN.md` и `docs/operations/DEPLOYMENT_EN.md` описывают: `main` = последний зелёный `main`, датированный тег = заморозка.

### 36.2 Верификация

- `docker compose -f docker-compose.deploy.yml config` валиден.
- `.github/workflows/deploy.yml` валиден (`python -c "import yaml; ..."`).
- `check-all -Quick` — 18 PASS, 3 SKIP, exit 0.

**Статус:** на ревью у Claude Code.

## 37. COVERAGE-1 — пороги покрытия unit-тестов подняты до 70%

### 37.1 Решение владельца

- Все пороги unit coverage — **70% как цель, так и enforced min** (frontend и backend).
- Backend измеряется по unit-testable пакетам; из знаменателя исключены CLI main, generated gRPC client (`internal/infrastructure/graph`), test helpers (`internal/testutil`, `internal/domain/cache/cachetest`) и `scripts`.

### 37.2 Изменения

- `.windsurfrules`: обновлена таблица покрытия.
- `docs/operations/TESTING.md`: актуальные цифры и пояснение по знаменателю backend.
- `docs/PROJECT_REVIEW_AI_AGENTS.md` §6: Go unit min 70%, frontend unit target/min 70%.
- `.devin/prompts/MASTER_PROMPT.md` и `MASTER_PROMPT_RU.md`: Go backend min 70%, frontend target/min 70%.
- `scripts/testing/core-checks.tsv` и `.github/workflows/_core-checks.yml`: backend coverage threshold 70%.
- `scripts/testing/check-all.ps1` и `check-all.sh`: используют `backend-coverage-total.py`.
- `scripts/testing/backend-coverage-total.py` + `backend-coverage-excludes.txt`: вычисляет backend unit coverage по `cover.out` с фильтром.
- `docs/DECISIONS.md`: решение #38.
- `docs/tasks/COVERAGE-1-align-backend-coverage-threshold.md` и `docs/AI_HANDOFF.md`/`docs/AI_LOG.md` обновлены.

### 37.3 Верификация

- `python scripts/testing/backend-coverage-total.py backend/cover.out 70` → **72.2%** [PASS].
- Frontend `npm run test:coverage` → statements **82.21%**, branches **70.7%**, functions **82.38%**, lines **84.01%** — все выше 70%.

**Статус:** на ревью у Claude Code.

---

## 38. FE-DEPS-106, GORM-COMBINED-1 и E2E-CANVAS-1 (2026-09-16)

### 38.1 FE-DEPS-106

Совместимое обновление frontend-инструментария (PR #110, смерджен в main): `vite` 8.3, `happy-dom` 20.14.3, `@types/node` 26.5.1. TypeScript 7.0.2 и ESLint 10.10.0 отложены из-за peer-конфликтов. Перед мержом починены: битая относительная ссылка в `docs/PROJECT_REVIEW_AI_AGENTS.md` и smoke test host 127.0.0.1 в `.github/workflows/ci.yml`.

### 38.2 GORM-COMBINED-1

Комбинированное обновление backend в `ai-agents`: `gorm.io/gorm` 1.31.2, `gorm.io/driver/postgres` 1.6.3, `gorm.io/datatypes` 1.2.7. Прогнаны `go mod tidy`, `go mod verify`, `go build ./...`, `go vet ./...`, `go test ./...` и `go test -tags=integration ./...` — PASS. `check-all.ps1 -Quick` — PASS (кроме отсутствующего `node_modules`, который CI установит).

### 38.3 E2E-CANVAS-1

Заведена постановка по двум падающим real-auth тестам `cockpit-canvas-controls.spec.ts` (`fog toggle`, `zoom transform`). Зафиксированы селекторы, URL, ожидания по туману и масштабу, гипотезы и план локализации.

2026-09-17 — исполнено Devin, на ревью у Claude Code. Корень: `readonly` (режим `community` для анонима) в `event-bridge.ts` ранним `return` отсекал не только редактирование, но и все view-взаимодействия — публичный граф нельзя было зумить и панорамировать. Семантика исправлена: readonly = «не редактируется, но интерактивен» — pan инициализируется сразу в `handleMouseDown` до детекта узлов, zoom/dblclick/touch больше не отсекаются; защита сохранена для выбора узла, контекстного меню, клавиатуры, drag узлов и ghost-ноды. `GraphTopBar` получил вариант `floating`: анониму доступны view-контролы (reset/search-open/focus/fog) и `top-bar-sign-in`/`top-bar-register`; поиск, фильтры типов/связей и переключатель personal/community — только авторизованным. В `GraphCanvas` guard `dataKey === lastDataKey && simState.isRunning` упрощён до `dataKey === lastDataKey` — эффект пересоздавал остановленную симуляцию и сбрасывал состояние.

Верификация: 1436 unit-тестов PASS (139 файлов), svelte-check PASS, lint 0 ошибок; Playwright `chromium-real-auth` `tests/cockpit-canvas-controls.spec.ts` — 7/7 PASS на изолированном стеке `SKIP_AUTH=false`. Ограничение среды: `nlp-test` не собирался — диск D: почти полон, модель ~4.4 ГБ рушила containerd (SIGBUS); для canvas-тестов NLP не нужен, DNS-имя закрыто stub-контейнером.
