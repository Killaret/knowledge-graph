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

| Уровень         | Команда                                         | Инструмент                       | Покрытие            |
| --------------- | ----------------------------------------------- | -------------------------------- | ------------------- |
| Go unit         | `cd backend && go test ./...`                   | testify                          | min 60%, target 70% |
| Go integration  | `cd backend && go test -tags=integration ./...` | testcontainers-go, miniredis     | —                   |
| Frontend unit   | `cd frontend && npm run test:unit`              | Vitest                           | target 70%          |
| E2E             | `cd frontend && npm run test`                   | Playwright                       | —                   |
| BDD             | `cd frontend && npm run test:bdd`               | Cucumber                         | —                   |
| NLP             | `cd nlp-service && pytest`                      | pytest                           | —                   |
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
- Roadmap: [`ROADMAP.md`](../ROADMAP.md), детальные планы — [`BACKLOG.md`](BACKLOG.md).

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
- Ручные сценарии: см. `docs/MANUAL_TEST_CHECKLISTS_RU.md` раздел `0.6` и `docs/TESTING.md` раздел `Manual Regression Scenarios`.

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

- Базовый и «туманный» снимки: [`docs/assets/a1-3d-visual-regression/`](assets/a1-3d-visual-regression/)
- Скрипт сравнения: удалён после использования.

**Осталось.**

- Повторная верификация Claude Code на живом стенде и пересборка официальных baseline Argos (`ARGOS_REFERENCE_BRANCH=main`).

## 18. AUD-4: контракт входа через Яндекс (2026-09-06)

**Что сделано.**

- В `backend/cmd/server/router.go` инициация OAuth перемещена с `/api/v1/auth/yandex` на `/api/v1/auth/yandex/login`; старый путь удалён.
- В `backend/internal/interfaces/api/handlers/auth/handler.go` `YandexLogin` теперь возвращает `200` с JSON `{"url": "<authorization URL>"}` вместо HTTP-редиректа.
- `backend/internal/interfaces/api/middleware/jwt.go` уже содержал `/api/v1/auth/yandex/login` в `SkipPaths` — проверено, дублирующих записей нет.
- `backend/openAPI.yaml` обновлён: путь `/api/v1/auth/yandex/login`, ответ `200` с телом `{ url: string }`, плюс `501` при отсутствии настройки; блок `302` удалён.
- `frontend/src/shared/api/auth.ts` не требовал изменений — клиент уже звал `/auth/yandex/login` и ожидал JSON.
- `YandexLoginButton.svelte` использует `window.location.href = result.url`, то есть переход происходит в браузере, а не через `fetch`.
- `README.md` очищен от пометки «OAuth2 через Яндекс сейчас не работает».
- `docs/CONFIGURATION_EN.md` дополнен разделом **Authentication** с JSON `backend.auth` и таблицей переменных окружения, включая `YANDEX_CLIENT_ID` и `YANDEX_CLIENT_SECRET`.
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
