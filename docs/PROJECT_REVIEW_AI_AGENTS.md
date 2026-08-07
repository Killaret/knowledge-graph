# Knowledge Graph — полный обзор проекта для AI-агентов

> **Живой артефакт передачи знаний.**
> Этот файл — обновляемый артефакт (living artifact) для передачи контекста между AI-агентами и human-разработчиками. Он содержит актуальный срез архитектуры, стека, тестовой стратегии, найденных рисков, применённых исправлений и roadmap. При любых существенных изменениях проекта этот документ следует актуализировать.

Документ подготовлен на основе read-only аудита архитектуры, DDD-слоёв, стиля кода, инфраструктуры и тестовой пирамиды, а также последующих исправлений рисков и полного регрессионного цикла.

---

## 1. Общее описание проекта

**Knowledge Graph** — мультитенантное SaaS-приложение для управления заметками с графовой структурой, NLP-рекомендациями и 3D-визуализацией (`docs/ARCHITECTURE_SUMMARY.md`, `README.md`).

Основные возможности:
- 3D-визуализация заметок как небесных тел.
- Графовые связи между заметками.
- Семантический поиск на pgvector.
- NLP-рекомендации (sentence-transformers).
- Система черновиков в MongoDB.
- JWT/OAuth2-аутентификация, RBAC.
- Геймификация (achievements).
- i18n: русский по умолчанию, английский — через ключи.
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
- **GORM v1.25** — ORM для PostgreSQL.
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
- **Vitest v3** — юнит-тесты.
- **Playwright v1.59** — E2E.
- **@cucumber/cucumber v12** — BDD.
- **FSD + Atomic Design** — структура `frontend/src/{shared,components,entities,features,widgets,routes}` (`.windsurfrules`).

### NLP
- **Python 3.11**, **FastAPI**.
- **sentence-transformers**, **transformers**, **torch**, **yake**, **nltk**.
- Модель `all-MiniLM-L6-v2`.
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
- UI по умолчанию на русском, но все строки через i18n-ключи.
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
- **Test**: frontend 3002, backend 18083, graph-service gRPC 19090 / HTTP 19091, postgres 15434, redis 16381, mongo 27019, nlp 15002.

### Volumes
- Dev: `postgres_data`, `redis_data`, `huggingface_cache`.
- Personal: `pgdata_personal`, `redisdata_personal`, `mongodbdata_personal`.
- Test: `test_postgres_data`, `test_mongodb_data`.

### Nginx
- `nginx.conf` и `nginx.personal.conf` — gateway с проксированием `/api` и `/graph-service/api`.

---

## 6. Тестовая пирамида

| Уровень | Команда | Инструмент | Покрытие |
|---|---|---|---|
| Go unit | `cd backend && go test ./...` | testify | min 60%, target 70% |
| Go integration | `cd backend && go test -tags=integration ./...` | testcontainers-go, miniredis | — |
| Frontend unit | `cd frontend && npm run test:unit` | Vitest | target 70% |
| E2E | `cd frontend && npm run test` | Playwright | — |
| BDD | `cd frontend && npm run test:bdd` | Cucumber | — |
| NLP | `cd nlp-service && pytest` | pytest | — |
| Full regression | `.\scripts\testing\run-full-test-cycle.ps1` | PowerShell + Docker + Playwright | — |
| Stacks identity | `.\scripts\ci\check-stacks-identity.ps1` | PowerShell | — |

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
8. **Coverage thresholds**: в `vitest.config.ts` стояли 60% (target — 70%).
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
- Устаревшие ссылки на `src/shared/three/` актуализированы (Three.js-логика перенесена в `docs/3d-archive/frontend/src/lib/three/`).

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
- `cd frontend && npm run test:coverage` — проходит при thresholds 70% (lines 80.36%, branch 80.69%, functions 79.97%).
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
   Основные ссылки в `.windsurfrules` и документах исправлены, но в `docs/3d-archive/` остаётся старая иерархия (архив, не production).

---

## 12. Ключевые файлы для быстрого старта

- Главные правила: [`.windsurfrules`](../.windsurfrules).
- Архитектура: [`docs/ARCHITECTURE_SUMMARY.md`](ARCHITECTURE_SUMMARY.md).
- Команды: [`COMMANDS.md`](../COMMANDS.md).
- Backend wiring: [`backend/cmd/server/main.go`](../backend/cmd/server/main.go).
- Frontend i18n: [`frontend/src/shared/utils/i18n.ts`](../frontend/src/shared/utils/i18n.ts), [`frontend/src/shared/utils/i18n/messages/`](../frontend/src/shared/utils/i18n/messages/).
- Frontend entry: [`frontend/src/routes/+page.svelte`](../frontend/src/routes/+page.svelte), [`frontend/src/features/home-page/home-page.svelte.ts`](../frontend/src/features/home-page/home-page.svelte.ts).
- NLP: [`nlp-service/Dockerfile`](../nlp-service/Dockerfile), [`nlp-service/app/main.py`](../nlp-service/app/main.py).
- Backup: [`docs/BACKUP.md`](BACKUP.md), [`knowledge-graph.config.json`](../knowledge-graph.config.json).
- Regression: [`scripts/testing/run-full-test-cycle.ps1`](../scripts/testing/run-full-test-cycle.ps1).
- Regression plan: [`docs/REGRESSION_TEST_PLAN.md`](REGRESSION_TEST_PLAN.md).
- Roadmap: [`ROADMAP.ru.md`](../ROADMAP.ru.md).

---

## 13. Дополнение: статус фич и roadmap

### Общий статус проекта
- **Фаза:** Alpha → Beta.
- **Стабильность:** критических проблем нет.
- **Регрессионное тестирование:** 11/14 частей пройдено.
- **Покрытие тестами:** 923 frontend unit-тестов (+57 по 2D-рендереру: fog, LOD, search outline, offscreen-cache/throttling, renderer-orchestrator, renderer-utils, variation, animation), backend unit-тесты — все проходят.
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

### В процессе / запланировано (UI/UX)
- Исправление моргания графа при дельта-обновлениях.
- Селектор типов заметок — выпадающий список.
- Документация и UI для типов связей.
- Кнопка «Отмена/Назад» на странице входа.
- Исправление загрузки 3D-графа.
- Система рекомендаций Pure Precomputed (убрать fallback).
- Публичные сиды для тестирования real-auth.
- Аудит ресурсов (память, CPU, bundle size).

### Инструменты импорта/экспорта
| Фича | Статус |
|---|---|
| Букмарклет | Реализован |
| Массовый импорт URL | Реализован + извлечение контента по URL (`internal/infrastructure/web.ImportFetcher`) |
| Браузерное расширение | Запланирован |
| JSON/Markdown/CSV импорт-экспорт | Запланирован |

> В `ROADMAP.md`/`ROADMAP.ru.md` недавно отмечено, что content extraction для массового импорта URL реализован через `internal/infrastructure/web.ImportFetcher`.

### Геймификация и Obsidian
- Система кастомизации, очки, достижения, бейджи, лидерборды — в планах.
- Импорт и синхронизация с Obsidian — в планах.

### PWA и внешние интеграции
- PWA Capture (быстрые заметки, оффлайн, push) — запланировано.
- Интеграции Pocket / Readwise / Twitter — запланированы.

### 3D и кластеризация
- Базовый 3D-рендеринг (Three.js) — разморожен и перенесён в `features/graph-3d`.
- Orbital / Solar System 3D, Honeycomb Stellaris 3D, серверная кластеризация Louvain, кэширование кластеров — в планах.
- Галактические кластеры и LOD — в бэклоге.

### Связи, быстрое редактирование, onboarding
- Тултипы связей, анимация, gamma-кодирование, автосоздание связей — запланировано.
- Dust Inbox, inline-редактирование, автодополнение, исправление NLP-ключевых слов — запланировано.
- Сценарный onboarding с T1-T6, хлебные крошки, spotlight — запланировано.

### Экспериментальные идеи
- **Factory Line** — визуализация графа как производственная цепочка.
- **Semantic Guardians** — семантические стражи.
