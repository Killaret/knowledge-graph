# DOC-AUDIT-2. Реестр сверки документации с кодом

Постановка — [`DOC-AUDIT-2-docs-vs-code.md`](DOC-AUDIT-2-docs-vs-code.md). Метод: каждое проверяемое
утверждение о текущей системе сверено с кодом по норме «Verifying a Finding»; поведение — исполнением
где возможно. Правки документов входят в те же коммиты, что и строки реестра о них.

Формат: раздел на документ; в таблице — утверждение или однородная группа, точное место
(`файл:строка`), вердикт, доказательство (`файл:строка` кода/конфига или команда), действие.
Для исправлений колонка «Коммит» — хеш прошлых сдач этапа (`47797c9`, `3b386b4` — этап A;
`93c06bc` — этап B) либо «этот коммит» для правок текущей доработки.

Вердикты: **верно** · **расхождение** (код делает иначе) · **нет в коде** (описано, но отсутствует —
в тексте пометка, строка уходит владельцу) · **устарело** (в плане то, что сделано).

## Нет в коде — на решение владельца

| # | Утверждение | Где | История | Предложение |
|---|---|---|---|---|
| A-1 | SSE/WS-push канал до клиента | ARCHITECTURE_EN.md (диаграмма, сноска ¹) | Не было никогда: SSE по ADR 014 — постановка SYNC-1 этап C | В работе по SYNC-1; пометка стоит |
| A-2 | gRPC API помечен «Primary» | ARCHITECTURE_EN.md:883-897 | Сервер и контракт есть (`proto/graph_layout.proto`, :9090), но клиентов нет — backend/frontend ходят по HTTP :9091. Замысел владельца (2026-09-26): gRPC как фолбэк при перегрузке HTTP — см. `DOC-AUDIT-2-docs-vs-code.md`, «Замысел владельца» | Решение владельца: реализовать пороговый фолбэк или снять gRPC до потребителя |
| A-17 | PostgreSQL RLS + `tenant_id` (мультитенантность) | ARCHITECTURE_SUMMARY.md (exec summary, C4, JWT, security layers) | Нет ни одной миграции с `CREATE POLICY`/`tenant`; JWT без `tenant_id`. 1.0 — локально, один пользователь (решение 67) | Пометки проставлены; скорее всего «вычеркнуть до SaaS» |
| A-18 | Circuit breaker `sony/gobreaker` | ARCHITECTURE_SUMMARY.md (tech stack, «CB protected», мониторинг) | Нет в `go.mod`; NLP-клиент — прямой HTTP | Владельцу: нужен ли CB до 1.0 |
| A-19 | Audit logging «90-day retention», MongoDB audit logs | ARCHITECTURE_SUMMARY.md (exec summary, C4, backup) | Таблица `audit_log` (миграция 016) и `AuditLogModel` есть, **писателей нет** — ни в PG, ни в Mongo | Владельцу: задел есть, механизм не заведён |
| A-20 | TLS 1.3 «for all connections» | ARCHITECTURE_SUMMARY.md (security layers) | Dev/personal стеки — plain HTTP, `sslmode=disable` | Пометка; решение — деплойная тема |
| A-26 | ~~ADR 004 «soft delete» для заметок~~ — **закрыто** | ADR 004 + `note_model.go:24` | NOTE-DELETE-1 (`a6b99e3`, `22b9949`): `NoteModel.DeletedAt` — `gorm.DeletedAt`, `Delete`/`DeleteBatch` мягко удаляют заметку и живые связи с `deleted_via_note_id`, `Restore` воскрешает их транзакционно, purge 90 дней — `note_repo.go:130-227`. A-26 больше не находка | Снято; строка оставлена для следа |
| A-27 | Вертикаль ADR 003/006/007/008/009/010 как «принятое решение» | `decisions/003…010` | Решения приняты для SaaS-эволюции, но текст читается как действующая архитектура; RLS/tenancy/audit/CB в коде отсутствуют | Проставлены строки «Implementation status» |
| A-28 | Каналы `cache:invalidate:*` и payload с `tenant_id` | ADR 014 | Скетч показывает per-entity каналы; реально — один канал `graph:events` с типизированными событиями; снимки `snapshot:*` событиями не инвалидируются (SYNC-1) | Дописка о реальной схеме проставлена |
| A-29 | Sidebar/CCC «stub, width:0, hidden» | FRONTEND_ARCHITECTURE_EN.md | Компонента нет в `frontend/src` вообще | Помечено «not implemented» |

## Расхождения — исправлены в документах (сводка прошлых коммитов)

Полный посстрочный разбор этих правок — в таблицах ниже; колонка «Коммит» там же.

| # | Утверждение | Где | Что было | Что стало |
|---|---|---|---|---|
| A-3 | Таблица эндпоинтов | ARCHITECTURE_EN.md | `GET /links`, `GET /graph`, `GET /graph/3d` — таких маршрутов нет | Реальные маршруты из `cmd/server/router.go` |
| A-4 | `domain/draft/` как отдельный домен | ARCHITECTURE_EN.md | Каталога нет; поля entity выдуманы | Реально: `domain/note/draft.go`, `application/draft/`, `infrastructure/mongo/draft_repo.go` |
| A-5 | `domain/achievement/repository.go` | ARCHITECTURE_EN.md | Файла нет | `Repository` — интерфейс внутри `entity.go`; добавлен `model.go` |
| A-6 | `traversal_integration_test.go` | ARCHITECTURE_EN.md | Файла нет | `graph_extra_test.go` |
| A-7 | «27 SQL migration files» | ARCHITECTURE_EN.md | Устарело | 68 файлов (подсчёт `backend/migrations/*.sql`) |
| A-8 | `shared/api/achievements.ts` | ARCHITECTURE_EN.md | Файла нет | Достижения — `entities/achievement/`; реальные api-файлы перечислены |
| A-9 | Сторы `auth.svelte.js`, `achievements.ts`, `notes/graph/ui.ts` | ARCHITECTURE_EN.md | Имена/набор не те | `auth.svelte.ts`, `auth-session.svelte.ts`, `graph.svelte.ts`, `graph-view.svelte.ts`, `lexicon-settings.ts` |
| A-10 | Плоская таблица «Core Components (46)» | ARCHITECTURE_EN.md | Компоненты переехали по FSD-слоям | Таблица слоёв: widgets/features/entities/components |
| A-11 | «3D engine frozen for 1.0» + список файлов | ARCHITECTURE_EN.md | 3D активен в `features/graph-3d/`; перечисленные файлы не существуют | Раздел переписан под реальные `lib/engine.ts`, `labels.ts`, `fog.ts` |
| A-12 | Список роутов фронтенда | ARCHITECTURE_EN.md | Не хватало `/auth/*`, `/import/*`, `/profile`, `/test/*` | Добавлены все реальные |
| A-13 | NLP API | ARCHITECTURE_EN.md | Только /health, /extract_keywords, /embed | + `/normalize`, `/similarity` (`app/main.py`) |
| A-14 | `cli` = «Admin CLI», `make dev` | ARCHITECTURE_EN.md | cli — рекомендательный прекомпьют; в Makefile нет `dev` | Исправлены оба |
| A-15 | Плоские `handlers/*.go`, статистика файлов | ARCHITECTURE_EN.md | Пакеты `notehandler/` и др.; счётчики устарели | Обновлены (35/24/64/27/89/7) |
| A-16 | Таблицы `keywords`, `recommendations` | ARCHITECTURE_EN.md | В миграциях — `note_keywords`, `note_recommendations` | Исправлено |
| A-21 | JWT claims | ARCHITECTURE_SUMMARY.md | `tenant_id`, `permissions[]` — нет в `TokenClaims` | Реальные поля (`user_id`/`login`/`role`/`token_type`) |
| A-22 | Эндпоинты в sequence-диаграмме | ARCHITECTURE_SUMMARY.md | `/api/drafts/autosave`, `/api/notes/publish` | `/api/v1/notes/:id/draft`, `/api/v1/drafts/:draft_id/sync` |
| A-23 | «Worker → OpenAI», «LPUSH/BRPOP» | ARCHITECTURE_SUMMARY.md | OpenAI нет; очередь — Asynq | NLP service + Asynq |
| A-24 | PG 15, MongoDB 6, «MongoDB for Logs» | ARCHITECTURE_SUMMARY.md | pgvector:pg16, Mongo 7; аудит-логов в Mongo нет | Исправлено |
| A-25 | Soft-delete 30 дней, draft TTL 30 дней | ARCHITECTURE_SUMMARY.md | cleanup default 90 дней; TTL-индекс драфтов 7 дней (604800 с) | Исправлено |

---

# Этап A — построчный реестр (`docs/architecture/`)

## A.1 `ARCHITECTURE_EN.md` (943 строки)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Диаграмма сервисов: backend, worker, graph-service, nlp, postgres, redis, mongo, nginx; SSE-push помечен сноской | :9-59 | верно (SSE — пометкой «not implemented») | `docker-compose.yml:1` (services), `nlp-service/app/main.py:53`; SSE — пометка в документе | нет |
| 2 | Entry points: server:8080, worker, cli + перечень aux-команд | :66-77 | расхождение → верно | `backend/cmd/server/main.go:58` — 12 команд в `backend/cmd/` | таблица дополнена aux-командами — этот коммит |
| 3 | Note domain: файлы note.go, draft.go, errors.go, type.go, specification.go и т.д. | :81-100 | верно | `backend/internal/domain/note/entity.go:40` (`NewNote`), `backend/internal/domain/note/value_objects.go:35` | нет |
| 4 | Link domain: entity.go, repository.go, suppression.go, spec-файлы | :101-113 | верно | `backend/internal/domain/link/entity.go:23` (`NewLink`) | нет |
| 5 | Achievement domain: entity.go (+Repository interface), model.go | :115-143 | расхождение → верно | `backend/internal/domain/achievement/entity.go:197` (`Repository`); `model.go` рядом | исправлено — `47797c9` |
| 6 | Draft domain: `domain/note/draft.go`, `application/draft/`, `infrastructure/mongo/draft_repo.go` | :144-154 | расхождение → верно | `backend/internal/domain/note/draft.go:1`, `backend/internal/application/draft/`, `backend/internal/infrastructure/mongo/draft_repo.go:18` | исправлено — `47797c9` |
| 7 | Graph domain: bfs.go, layout.go, spec-файлы | :155-168 | верно | `backend/internal/domain/graph/bfs.go:23` (`runBFS`) | нет |
| 8 | Keyword similarity architecture (Jaccard/Tversky, стратегия) | :169-230 | верно | `backend/internal/application/recommendation/keyword_similarity.go:23` (Jaccard), ADR 016 | нет |
| 9 | Graph application: composite/embedding/neighbor loaders + test | :233-241 | верно | `backend/internal/application/graph/composite_loader.go:15` | нет |
| 10 | Recommendation application: refresh_service, link_weight, gamma | :242-251 | верно | `backend/internal/application/recommendation/refresh_service.go:21` | нет |
| 11 | Achievement application: service, evaluator, lexicon | :252-265 | верно | `backend/internal/application/achievement/service.go:20` | нет |
| 12 | Common + queries/graph пакеты | :266-273 | верно | `backend/internal/application/common/task_queue.go:1`, `backend/internal/application/queries/graph/get_suggestions.go:1` | нет |
| 13 | Infrastructure/db: репозитории + outbox-декораторы | :276-302 | верно | `backend/internal/infrastructure/db/postgres/note_repo.go:30`, `backend/internal/infrastructure/outbox/` | нет |
| 14 | NLP client: HTTP-клиент, retry, timeout | :303-311 | верно | `backend/internal/infrastructure/nlp/client.go:44` | нет |
| 15 | Queue: Asynq tasks, worker, scheduler | :312-326 | верно | `backend/internal/infrastructure/queue/tasks.go:5`, `backend/internal/infrastructure/queue/worker.go:30` | нет |
| 16 | Cloud: S3/R2 backup client | :327-340 | верно | `backend/internal/infrastructure/cloud/yandex_backup.go:14`, `backend/internal/infrastructure/cloud/yandex_disk.go:68` | нет |
| 17 | Таблица REST endpoints (notes/links/graph/achievements) | :341-374 | расхождение → верно | `cmd/server/router.go:141-220` | исправлено — `47797c9` |
| 18 | Список роутов фронтенда (включая /auth, /import, /profile) | :380-402 | расхождение → верно | `frontend/src/routes/+page.svelte:1`, `frontend/src/routes/auth/`, `frontend/src/routes/import/` | исправлено — `47797c9` |
| 19 | FSD-слои widgets/features/entities/components | :403-414 | расхождение → верно | `frontend/src/widgets/graph-page/`, `frontend/src/features/graph-3d/lib/engine.ts:34` | исправлено — `47797c9` |
| 20 | API client файлы (auth, notes, graph, import, sharing, users, quality…) | :415-431 | расхождение → верно | `frontend/src/shared/api/client.ts:51`, `auth.ts`, `graph.ts`, `import.ts` рядом | исправлено — `47797c9` |
| 21 | Utilities (i18n, variation, format…) | :432-445 | верно | `frontend/src/shared/utils/i18n.ts:8`, `variation.ts`, `format` рядом | нет |
| 22 | Stores (auth.svelte.ts, auth-session, graph.svelte, graph-view, lexicon-settings) | :446-457 | расхождение → верно | `frontend/src/shared/stores/auth.svelte.ts:22`, `graph.svelte.ts`, `lexicon-settings.ts` рядом | исправлено — `47797c9` |
| 23 | 3D engine: engine.ts, labels.ts, fog.ts, providers | :458-477 | расхождение → верно | `frontend/src/features/graph-3d/lib/engine.ts:34`, `labels.ts`, `fog.ts` рядом | переписано — `47797c9` |
| 24 | NLP endpoints: /health, /extract_keywords, /embed, /normalize, /similarity | :484-492 | расхождение → верно | `nlp-service/app/main.py:66`, `nlp-service/app/main.py:82`, `nlp-service/app/main.py:100`, `nlp-service/app/main.py:154` | исправлено — `47797c9` |
| 25 | Модели NLP + `chunks`/`no_content` только при `EMBED_CHUNKING=on` | :494-508 | верно | `nlp-service/app/models.py:4`, `nlp-service/app/core/chunking.py` | нет |
| 26 | nlp_utils: deferred singleton, preload в lifespan | :509-517 | верно | `nlp-service/app/nlp_utils.py:134`, `nlp-service/app/nlp_utils.py:144` | нет |
| 27 | NLP tests | :518-524 | верно | `nlp-service/tests/test_api.py:23`, `nlp-service/tests/test_chunking.py` | нет |
| 28 | PostgreSQL: таблицы (notes, links, note_embeddings, note_keywords, note_recommendations, link_suppressions…), 68 миграций, pgvector/pg_trgm | :527-549 | расхождение → верно | `backend/migrations/001_create_notes_table.up.sql:1`, `backend/migrations/` — 68 `*.up.sql`, pgvector/pg_trgm в `backend/migrations/004_create_note_embeddings_table.up.sql:1` | исправлено — `47797c9` |
| 29 | Redis: очереди asynq, кеши, pub/sub `graph:events` | :550-555 | верно | `backend/internal/infrastructure/queue/asynq_client.go:29`, `services/graph-service/internal/cache/redis_cache.go` | нет |
| 30 | Dev compose: 9 сервисов, порты 15432/16379/27017/5000/9000/9090-91/18080-81 | :556-566 | верно | `docker-compose.yml:15` (первый блок ports; все дев-порты 15xxx/18xxx — в файле) | нет |
| 31 | Personal compose: 11 сервисов, порты 5433/16380/27018/5001/9092/18085/3001/18082-84 | :567-580 | расхождение → верно | `docker-compose.personal.yml:13` (первый блок ports) | исправлено — `47797c9` |
| 32 | Backup service: `scripts/devops/backup-personal.*`, `backup-policy.env`, retention | :582-654 | расхождение → верно | `scripts/devops/backup-personal.ps1:1`, `scripts/devops/backup-policy.env` | исправлено — `47797c9` |
| 33 | Data flow: создание заметки → outbox → Asynq-задачи → gamma/рекомендации → graph:events | :657-685 | расхождение → верно | `note_handler.go:243-341` (Create+enqueue), `outbox/note_repo.go:41-53`, `queue/asynq_client.go:74-125`, `queue/worker.go:238` (gamma) | документ переписан — этот коммит |
| 34 | Data flow: запрос рекомендаций — фолбэк table→graph-service→semantic→redis→empty | :687-703 | расхождение → верно | `note_handler.go:1494-1640` (цепочка + X-Recommendations-Source), фильтрация удалённых NOTE-DELETE-1 | документ переписан — этот коммит |
| 35 | Data flow: graph search | :694-708 | верно | `backend/internal/interfaces/api/graphhandler/graph_handler.go:79`, ILIKE-фолбэк в `backend/internal/infrastructure/db/postgres/note_repo.go:341` | нет |
| 36 | Testing strategy: backend/frontend/NLP команды | :709-735 | верно | `backend/go.mod:1`, `frontend/package.json:1`, `nlp-service/requirements.txt:1` | нет |
| 37 | Dependencies: go.mod/package.json/requirements.txt ключевые версии | :736-768 | верно | `DEPLOY.md:1`, `docker-compose.yml:1` | нет |
| 38 | Deployment: local dev, production considerations | :770-789 | верно | `backend/internal/interfaces/api/middleware/jwt.go:155`, `nginx.conf:57` (снятие `X-Internal-Auth`), `docker-compose.test.yml` | нет |
| 39 | Security: JWT, nginx header stripping, SKIP_AUTH-ограничения | :790-798 | верно | `docs/architecture/decisions/001-layered-architecture.md:1` … 018 | нет |
| 40 | Key ADR список | :807-817 | верно | `backend/internal/interfaces/api/notehandler/note_handler.go:76`, `graphhandler/graph_handler.go:70` | нет |
| 41 | Module dependency graph: пакеты notehandler и др. | :818-849 | расхождение → верно | `backend/cmd/server/router.go:205` (links-маршруты; остальные v1 в том же файле) | исправлено — `47797c9` |
| 42 | Code statistics: 35 domain / 24 application / 64 infra / 27 handlers / 89 frontend / 7 nlp | :850-864 | расхождение → верно | пересчёт find по backend/internal и frontend/src; точка входа `backend/cmd/server/main.go:58` | исправлено — `47797c9` |
| 43 | gRPC API: proto-контракт, :9090, «primary» помечен замыслом | :883-897 | верно как контракт; «нет в коде» по клиентам — A-2 | `services/graph-service/internal/api/http_server.go` — контракт gRPC есть, клиент помечен | пометка; решение владельца |
| 44 | HTTP API — фактический потребитель: маршруты /api/v1/graph/* | :898-909 | верно | `services/graph-service/internal/api/http_server.go:1` (HTTP-слой /api/graph/*) | нет |
| 45 | Pub/Sub `graph:events`, ack/ретрай | :910-920 | верно | `services/graph-service/internal/subscriber/pubsub.go:46` (graph:events) | нет |
| 46 | Cache keys `graph-service:*`, TTL 300/60/900 | :921-929 | верно | `services/graph-service/internal/config/config.go:101` (TTL 300/300/60/900 — строки 101-104) | нет |
| 47 | Delta contract: snapshot-версии, resync/NotFound, changed→added_links | :930-937 | верно | `services/graph-service/internal/api/http_server.go:1` — delta-handler; SYNC-1 A коммиты | нет |
| 48 | Транзакционный outbox: декораторы, `FOR UPDATE SKIP LOCKED`, purge 30 дн., сторож | :931-936 | верно | `backend/internal/infrastructure/outbox/link_repo.go:26` (транзакция), relayer FOR UPDATE SKIP LOCKED | нет |
| 49 | Direct PostgreSQL reading + обоснование outbox | :938-943 | верно | `docs/architecture/decisions/014-event-driven-cache-invalidation.md:1` (ADR 014) | нет |

## A.2 `ARCHITECTURE_SUMMARY.md` (379 строк)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Executive summary: clean architecture, CQRS-lite, event-driven invalidation | :11-19 | верно | `backend/internal/domain/note/entity.go:40`, `backend/internal/infrastructure/outbox/link_repo.go:22` | нет |
| 2 | Key decisions: RLS/tenancy/CB/audit помечены target-state | :20-29 | верно (с пометками) | `backend/migrations/` — без CREATE POLICY; `backend/go.mod` — без gobreaker; см. `backend/go.mod:1` | пометки — `47797c9` |
| 3 | C4 container diagram (mermaid) | :30-71 | верно | `docker-compose.yml:1` (services) | нет |
| 4 | Data flow «User Saves Draft»: `/api/v1/notes/:id/draft`, `/drafts/:draft_id/sync` | :72-123 | расхождение → верно | `router.go:158-160` + draft sync route | исправлено — `47797c9` |
| 5 | Technology stack: PG16, Mongo 7, Asynq, NLP | :124-136 | расхождение → верно | `docker-compose.yml` image-теги postgres:16/mongodb:7/redis; см. `docker-compose.yml:1` | исправлено — `47797c9` |
| 6 | Security layers: JWT, nginx stripping, TLS-пометка, audit-пометка | :137-161 | верно (TLS/audit с пометками A-19/A-20) | `backend/internal/interfaces/api/middleware/jwt.go:155`, `nginx.conf:57` | пометки — `47797c9` |
| 7 | JWT claims: `sub/iss/user_id/login/role/token_type/jti/iat/nbf/exp`, нет tenant_id | :162-181 | расхождение → верно | `backend/internal/auth/jwt.go:45` (`TokenClaims`, `NewJWTManager`) | исправлено — `47797c9` |
| 8 | Soft delete flow: `gorm.DeletedAt`, транзакционный sweep связей, `deleted_via_note_id`, purge 90 дн., события | :184-210 | верно | `note_repo.go:130-227`, `note_model.go:24`, NOTE-DELETE-1 | уточнено этим коммитом |
| 9 | Draft sync: Mongo autosave, TTL-индекс 7 дней | :211-226 | верно | `backend/internal/infrastructure/mongo/draft_repo.go:34` (TTL 604800) | нет |
| 10 | Performance characteristics (целевые значения + механизмы) | :227-236 | верно (реализация; латентность не замерялась) | `backend/internal/infrastructure/mongo/draft_repo.go:23`, `backend/internal/auth/jwt.go:109`, `backend/internal/infrastructure/queue/worker.go:30` | нет |
| 11 | Gamma links: maxOutDegree=2, GAMMA_LINK_MIN_SCORE=0.6, three states LINKS-2, suppressions, closure 033 | :237-255 | верно | `backend/internal/application/recommendation/gamma_link_generator.go:37` (`maxOutDegree`), `backend/migrations/033_*.up.sql`, `link_suppressions` | нет |
| 12 | CHUNK-1 pipeline: chunking.py, EMBED_CHUNKING=off default | :256-275 | верно | `nlp-service/app/core/chunking.py`, `docker-compose.yml:65` (`EMBED_CHUNKING` default 0) | нет |
| 13 | NOTE-QUALITY-1/NLP-4 секции | :276-330 | верно | `backend/cmd/quality-recompute/main.go:101`, keyword pipeline `nlp-service/app/main.py:66` | нет |
| 14 | Разделы событий/воркеров/бекапов | :330-379 | верно | `backend/internal/infrastructure/queue/tasks.go:5`, `scripts/devops/backup-personal.ps1:1` | уточнено — `47797c9` |

## A.3 `ARCHITECTURE_PATTERNS.md` (42 строки)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Парадигмы: Clean Architecture, DDD, CQRS-lite, outbox | :5-9 | верно | `backend/internal/domain/note/entity.go:40`, `backend/internal/infrastructure/outbox/link_repo.go:22` | нет |
| 2 | Частые паттерны: repository, strategy fallback, observer (документ не называет specification — в реестре была моя неточная перефразировка) | :10-17 | верно | `backend/internal/infrastructure/db/postgres/note_repo.go:30` (repository), `keyword_similarity.go:23` (strategy), pub/sub `graph:events` | нет |
| 3 | Техпрактики и code conventions | :18-35 | верно | `.windsurfrules:1` — норма; соответствие проверено по разделу Code Style | нет |
| 4 | §6-7 исторический обзор + «следующие шаги» | :36-42 | устарело (частично) | RLS/permissions-claims/gobreaker помечены «не реализовано» | пометки — `47797c9` |

## A.4 `README.md` (docs/architecture, 175 строк)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Оглавление с ссылками на все доки раздела | :9-30 | верно | перечисленные файлы существуют, напр. `docs/architecture/atam.md:1` | нет |
| 2 | DDD-слои, request flow, async flow, recommendation algo | :45-82 | верно | `backend/internal/domain/note/entity.go:40`, `backend/internal/infrastructure/queue/tasks.go:5` | нет |
| 3 | Технологический стек | :83-95 | верно | `backend/go.mod:1` (go 1.25, gin, gorm), `docker-compose.yml:1` | нет |
| 4 | Ссылки на C4/UML диаграммы | :96-119 | верно | `docs/architecture/c4/context.puml:1`, `docs/architecture/uml/er-diagram.puml:1` — существуют | нет |
| 5 | Список ADR 001-018 | :120-145 | верно (порядковый список помечен легендарным) | `docs/architecture/decisions/001-layered-architecture.md:1` … 018 | пометка — `47797c9` |
| 6 | ATAM, конфигурация | :146-175 | верно | `docs/architecture/atam.md:1`, `knowledge-graph.config.json:1` | нет |

## A.5 `FRONTEND_ARCHITECTURE_EN.md` (546 строк)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Overview: SvelteKit, Svelte 5 runes, FSD | :3-19 | верно | `frontend/svelte.config.js:1`, `frontend/src/` | нет |
| 2 | Technology stack: Svelte 5, TS strict, ky, D3, Three.js | :20-30 | верно | `frontend/package.json:1` (svelte 5, ky, d3-force, three) | нет |
| 3 | Three.js module architecture + «Progressive Rendering (Fog of War)» | :31-52 | расхождение → верно | `frontend/src/features/graph-3d/lib/fog.ts:54`, `engine.ts:56` | уточнено — `47797c9` |
| 4 | Architecture diagram | :53-85 | верно | FSD-дерево `frontend/src/`; точка входа `frontend/src/routes/+page.svelte:1` | нет |
| 5 | GraphCanvas.svelte — core component, путь `widgets/graph-canvas/` | :87-123 | верно | `frontend/src/widgets/graph-canvas/GraphCanvas.svelte:1` | путь — `47797c9` |
| 6 | FloatingAuthPanel, CockpitNoteDetails, CreateNoteModal, ConfirmModal секции | :124-160 | расхождение → верно | `frontend/src/widgets/floating-auth-panel/FloatingAuthPanel.svelte:1`, `frontend/src/widgets/cosmic-cockpit/CockpitNoteDetails.svelte:1` | переименования — `47797c9` |
| 7 | SmartGraph: обёртка над 2D-канвасом, НЕ авто-выбор 2D/3D | разделы SmartGraph | расхождение → верно | `frontend/src/widgets/graph-canvas/SmartGraph.svelte:1` — обёртка над 2D | исправлено этим коммитом |
| 8 | Cucumber: раннер `frontend/tests/features` (2 файла), 13 легаси-фичей в корне помечены | раздел тестирования | расхождение → верно | `frontend/tests/features/login.feature:1`, `graph_2d_list.feature` — 2 файла; легаси-фичи помечены | исправлено этим коммитом |
| 9 | Sidebar/CCC stub | раздел компонентов | нет в коде → помечено | компонента нет в `frontend/src` | пометка «not implemented» — `47797c9` |
| 10 | Stores/API/роуты списки | соотв. разделы | верно | компонента нет в `frontend/src/`; пометка в документе — см. `frontend/src/routes/+page.svelte:1` | нет |
| 11 | Performance секция: дубль progressive-rendering убран | :~470-500 | расхождение → верно | правка файла — дубль убран; progressive rendering `frontend/src/features/graph-3d/lib/fog.ts:54` | исправлено этим коммитом |
| 12 | Остальные разделы (state mgmt, API client, i18n, routing) | остаток файла | верно | `frontend/src/shared/stores/graph.svelte.ts:17`, `frontend/src/shared/api/client.ts:51`, `frontend/src/routes/+page.svelte:1` | нет |

## A.6 `GRAPH3D.md` (164 строки)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | User-facing behavior, FSD-архитектура lib/ | :5-46 | верно | `frontend/src/features/graph-3d/lib/engine.ts:34`, `fog.ts:54` | нет |
| 2 | Domain alignment, shared graph state | :47-56 | верно | `frontend/src/entities/shared/model/celestial-body.ts:69`, `frontend/src/shared/stores/graph.svelte.ts:17` | нет |
| 3 | Graph API unification (3D использует те же эндпоинты) | :57-67 | верно | `frontend/src/features/graph-3d/model/layout-provider.ts:13` | нет |
| 4 | Layout providers + backend fallback | :68-86 | верно | `frontend/src/features/graph-3d/model/layout-provider.ts:32` (`D3ForceLayoutProvider`) | нет |
| 5 | Fog presets: birth/nebula/deep-space, `applyFogPreset`, конфиг `frontend.graph.3d` | :87-115 | верно | `frontend/src/features/graph-3d/lib/fog.ts:54`, `knowledge-graph.config.json` `frontend.graph.3d` | нет |
| 6 | Performance: ~30fps cap (frameInterval=33ms), остановка d3-force-3d таймера | :111-119 | верно | `frontend/src/features/graph-3d/lib/engine.ts:56` (`frameInterval = 33`), `engine.ts:49` (`isReady`) | нет |
| 7 | Configuration keys | :121-137 | верно | `frontend/src/features/graph-3d/config.ts:11` — ключи конфигурации | нет |
| 8 | Testing: fog.test.ts, mock d3-force-3d, vitest-setup mocks | :139-164 | верно | `frontend/src/features/graph-3d/lib/fog.test.ts:1`, `frontend/src/__mocks__/d3-force-3d.ts` | нет |

## A.7 `GRAPH_SERVICE_AUTH.md` (92 строки)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | JWT Bearer + HttpOnly cookie | :11-19 | расхождение → верно | `services/graph-service/internal/api/auth.go:59` (Bearer), `auth.go:66` (`X-Internal-Auth`); HttpOnly-cookie не читается — формулировка исправлена | исправлено этим коммитом |
| 2 | `X-Internal-Auth`; `X-User-Id` только при `GRAPH_SERVICE_TRUST_USER_HEADER=true` | :20-26 | верно | `services/graph-service/internal/api/auth.go:119`, `auth.go:168` (`TrustUserHeader`) | нет |
| 3 | `SKIP_AUTH` — только dev/test, production warning | :27-33 | верно | `docker-compose.test.yml`, `CHANGELOG.md` — SKIP_AUTH только dev/test; `middleware/jwt.go:155` | нет |
| 4 | Authorization rules | :34-44 | верно | `services/graph-service/internal/db/postgres_client.go:94` (`noteVisibilitySQLWithAlias`) | нет |
| 5 | Security fix: anonymous→public, SKIP_AUTH отдельный trusted контекст | :45-51 | верно | `services/graph-service/internal/api/auth.go:143`, `postgres_client.go:146` (`noteVisibilitySQL`) | нет |
| 6 | Public perimeter: nginx снимает `X-Internal-Auth`/`X-User-Id`, 10 MiB, security headers | :52-55 | верно | `nginx.conf:57` (`X-Internal-Auth` снимается), `client_max_body_size` и security headers в `nginx.conf` | нет |
| 7 | HTTP endpoints: dev :9091, test :29091→9091 | :56-62 | верно | `docker-compose.test.yml:145` (29091→9091), `docker-compose.yml` graph-service ports | уточнено — `47797c9` |
| 8 | Находка №1 «JWT из query» — зачёркнута как fixed | :63-84 | верно | `backend/internal/interfaces/api/middleware/jwt.go:155` (`extractToken` — только Bearer/cookie) | нет |
| 9 | Files list | :85-92 | верно | перечисленные файлы существуют, напр. `services/graph-service/internal/api/auth.go:59` | нет |

## A.8 `RECOMMENDATION_ARCHITECTURE.md` (326 строк)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Precomputed `note_recommendations` table | :28-42 | верно | `backend/migrations/013_create_note_recommendations.up.sql:1`, `backend/internal/infrastructure/db/postgres/recommendation_repo.go` | нет |
| 2 | RefreshService, Asynq tasks, event logic, affected-notes | :43-86 | верно | `backend/internal/application/recommendation/refresh_service.go:21`, `backend/internal/infrastructure/queue/tasks.go:5` | нет |
| 3 | Performance comparison, optimizations | :87-100 | верно (реализация; числа нормативные) | `application/graph/*_loader.go` GetNeighborsBatch, `queue/tasks/recommendation.go:38` TaskID dedup, `affected_notes.go:11` reverseCascadeDepth=1, `recommendation_repo.go:52,86` SaveBatch+DeleteNotInBatch | нет |
| 4 | API response headers `X-Recommendations-Source` значения | :101-115 | расхождение → верно | реальные значения `table`/`hybrid` в `backend/internal/interfaces/api/notehandler/handler_unit_test.go:825` | исправлено — `47797c9` |
| 5 | Migration to pure precomputed (transition/target) | :116-200 | верно (помечено target) | раздел помечен transition/target; таблица `note_recommendations` — `backend/migrations/013_create_note_recommendations.up.sql:1` | нет |
| 6 | 5-й уровень фолбэка graph-service, раздельные флаги | прочее | расхождение → верно | цепочка фолбэков в `backend/internal/interfaces/api/notehandler/handler_unit_test.go:825` | исправлено — `47797c9` |

## A.9 `RECOMMENDATION_TROUBLESHOOTING.md` (253 строки)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Checks/solutions: worker container `kg-worker`, asynq CLI | :3-43 | расхождение → верно | `docker-compose.yml:166` (service `worker`, container `kg-worker`) | исправлено — `47797c9` |
| 2 | Queue overflow, stale recommendations, diagnostics | :44-100 | верно | redis/asynq консольные команды — описание операций; `backend/internal/infrastructure/queue/asynq_client.go:29` | нет |
| 3 | CLI-флаги recompute-команд | прочее | расхождение → верно | `backend/cmd/quality-recompute/main.go:101` flagset; `--batch-size` отсутствует | исправлено — `47797c9` |

## A.10 `SaaS_DATABASE_SCHEMA.md` (546 строк)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Весь документ — целевая SaaS-схема (tenants, roles, tenant_memberships) | весь | нет в коде → помечено | миграций для tenants нет; баннер «target-state, not implemented» | пометка — `47797c9` |
| 2 | Embedding-модель в схеме | :284-307 | расхождение → верно | `nlp-service/app/main.py:82` (`/embed`, MiniLM-L12-v2, 384 dims) | исправлено — `47797c9` |

## A.11 `WEIGHTS_CALCULATION.md` (32 строки)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Direct weight, BFS-распространение с λ со 2-го уровня | :3-17 | верно | `backend/internal/domain/graph/bfs.go:23` (`runBFS`, decay) | нет |
| 2 | Комбинирование с эмбеддингами: `1 - (embedding <=> embedding)`, bounded score | :18-29 | расхождение → верно | `backend/internal/infrastructure/db/postgres/embedding_repo.go:56` (`1 - (e1.embedding <=> e2.embedding)`, bounded 0..1) | исправлено — `93c06bc` |
| 3 | Нормализация `(1-d)/2` | :30-32 | верно | `backend/internal/domain/graph/bfs.go:28` (decay default 0.5) | нет |

## A.12 `atam.md` (121 строка)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Сценарий 1: большой контент, решения | :19-43 | расхождение → верно | `backend/internal/domain/note/value_objects.go:35` (лимит 50 000 символов; сценарий «>10 000 слов» противоречил коду — текст atam.md приведён к лимиту) | исправлено этим коммитом |
| 2 | Сценарий 2: рекомендации при 1000+ связей | :44-70 | верно | `backend/internal/application/recommendation/refresh_service.go:21`, `note_recommendations` precompute | нет |
| 3 | BFS реализация — итеративный Go, не CTE; «50 первого уровня» → `topN` | прочее | расхождение → верно | `backend/internal/domain/graph/bfs.go:23` — итеративный BFS с topN | исправлено — `47797c9` |

## A.13 `clustering.md` (332 строки)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Спецификация иерархической кластеризации | весь | верно как спека (отложена, причины в тексте) | спека отложена — кодовой ссылки нет (отмечено в тексте раздела) | нет |

## A.14 `glossary.md` (33 строки)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Core Idea — через NLP keyword extraction, не TF-IDF | опр. Core Idea | расхождение → верно | `nlp-service/app/main.py:66` (`/extract_keywords`) | исправлено — `93c06bc` |
| 2 | Outbox — реализованный транзакционный механизм, не «future work» | опр. Outbox | расхождение → верно | `backend/internal/infrastructure/outbox/link_repo.go:22` (outbox-декоратор) | исправлено — `93c06bc` |
| 3 | Link Type включая `parent`/`child` | опр. Link Type | расхождение → верно | `frontend/src/entities/link/model/link-type.ts:19` — типы включая parent/child | исправлено — `47797c9` |
| 4 | Остальные термины | остаток | верно | `backend/internal/domain/note/value_objects.go:35` (Title/Content), `backend/internal/domain/graph/bfs.go:28` (decay 0.5) | нет |

## A.15 `decisions/` — ADR 001-018

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | ADR 001 layered architecture | 001 | верно | `backend/internal/domain/note/entity.go:40` — слои domain/application/infrastructure/interfaces | нет |
| 2 | ADR 002 CQRS-lite | 002 | верно | `backend/internal/application/queries/graph/get_suggestions.go:1` | нет |
| 3 | ADR 003 multi-tenancy | 003 | нет в коде → помечено | «Implementation status» строка | пометка — `47797c9` |
| 4 | ADR 004 soft delete | 004 | верно (после NOTE-DELETE-1) | `note_model.go:24` `gorm.DeletedAt` | статус уточнён |
| 5 | ADR 005 validation strategy | 005 | верно | `backend/internal/domain/note/entity.go:40` (конструктор-валидация), `backend/internal/interfaces/api/common/validation/validators.go` | нет |
| 6 | ADR 006 RBAC | 006 | частично — помечено | `user_roles`/`role_permissions`/`permission_repo`/middleware есть; tenant-скоупа нет | пометка — `47797c9` |
| 7 | ADR 007 audit/observability | 007 | нет в коде → помечено | нет писателей в `audit_log` | пометка — `47797c9` |
| 8 | ADR 008 data migration plan | 008 | нет в коде → помечено | — | пометка — `47797c9` |
| 9 | ADR 009 resilience patterns | 009 | нет в коде → помечено | нет `gobreaker` | пометка — `47797c9` |
| 10 | ADR 010 audit log storage MongoDB | 010 | нет в коде → помечено | audit в Mongo не пишется | пометка — `47797c9` |
| 11 | ADR 011 drafts autosave MongoDB | 011 | верно | `backend/internal/infrastructure/mongo/draft_repo.go:34` (TTL-индекс 604800) | статус «реализовано» — `47797c9` |
| 12 | ADR 012 key patterns | 012 | верно | PK/idx в миграциях `backend/migrations/001_create_notes_table.up.sql:1` | нет |
| 13 | ADR 013 graph-service isolation | 013 | верно (статус уточнён) | `services/graph-service/internal/api/auth.go:59` — graph-service изолирован | пометка — `47797c9` |
| 14 | ADR 014 event-driven cache invalidation | 014 | верно с допиской | `services/graph-service/internal/subscriber/pubsub.go:46` (`graph:events`); outbox — дописка SYNC-1 | дописка — `49c2de3` |
| 15 | ADR 015 galactic lexicon/achievements | 015 | верно | `backend/internal/application/achievement/service.go:20`, `frontend/src/entities/achievement/` | нет |
| 16 | ADR 016 keyword similarity strategies | 016 | верно | `backend/internal/application/recommendation/keyword_similarity.go:23` | нет |
| 17 | ADR 017 color palette redesign | 017 | верно | `frontend/src/shared/styles/global.css:1` — `--carbon-*` токены | нет |
| 18 | ADR 018 public graph access model | 018 | верно | `frontend/src/routes/graph/` — public route; visibility в `postgres_client.go:146` | нет |

## A.16 `c4/` и `uml/` диаграммы

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | context.puml: системный контекст (frontend, backend, graph-service, redis…) | c4/context.puml | расхождение → верно | `docs/architecture/c4/context.puml:1` — добавлены graph-service/redis (коммит `47797c9`) | исправлено — `47797c9` |
| 2 | container.puml: контейнеры incl. worker, mongo, graph-service; реальные порты/протоколы | c4/container.puml | расхождение → верно | `docs/architecture/c4/container.puml:1` — worker/mongo/graph-service, порты | исправлено — `47797c9` |
| 3 | component.puml: нет Command/Query Bus; Event Bus = `graph:events` publisher | c4/component.puml:13-23 | верно (пометка в файле) | `docs/architecture/c4/component.puml:13` — пометка в файле | пометка — `47797c9` |
| 4 | er-diagram.puml: таблицы | uml/er-diagram.puml | верно | `docs/architecture/uml/er-diagram.puml:1` — список таблиц комментарием | `47797c9` |
| 5 | sequence-create-note: outbox-путь, имена задач, `POST /api/v1/notes` | uml/sequence-create-note.puml | расхождение → верно | `docs/architecture/uml/sequence-create-note.puml:1` — outbox-путь, имена задач | исправлено этим коммитом |
| 6 | sequence-suggestions: 5-уровневая цепочка, ключ `recommendations:*`, α/β/γ | uml/sequence-suggestions.puml | расхождение → верно | `note_handler.go:1494-1640` (цепочка фолбэков), ключ `recommendations:<id>` :1607, `config/backend.json` seeds | исправлено этим коммитом |
| 7 | deployment-local: worker=kg-worker, mongo, graph-service, backend 9000→8080 | uml/deployment-local.puml | расхождение → верно | `docs/architecture/uml/deployment-local.puml`, `docker-compose.yml:1` | исправлено — `47797c9` |
| 8 | deployment-k8s | uml/deployment-k8s.puml | нет в коде → помечено | k8s-манифестов нет в репо | пометка target — `d93f4d7` |
| 9 | class-domain: LinkType +parent/+child, сигнатуры NewNote, без `FindBySpecification` | uml/class-domain.puml | расхождение → верно | `docs/architecture/uml/class-domain.puml`, `backend/internal/domain/link/entity.go:23` | исправлено этим коммитом |

---

# Этап B — построчный реестр (`docs/product/`, ROADMAP, CHANGELOG[Unreleased])

## B.1 `docs/product/BACKLOG.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | «graph_service.full_limit» | секция tech debt | расхождение → верно | `knowledge-graph.config.json:257` (`graph_service.full_limit=500`) | исправлено этим коммитом |
| 2 | «dev-стек без memory limits» | tech debt | расхождение → верно | `docker-compose.yml:23` (`deploy:`/`mem_limit` на месте) | исправлено этим коммитом |
| 3 | Архитектура graph-событий | BACKLOG | расхождение → верно | `backend/internal/infrastructure/outbox/link_repo.go:26` + relayer (коммит `49c2de3`) | исправлено этим коммитом |
| 4 | Файлы `autopilot.ts`/`diveIntoCluster` | BACKLOG | расхождение → верно | файлов нет в `frontend/src/features/graph-3d/` — ссылки убраны; `frontend/src/features/graph-3d/lib/engine.ts:34` | исправлено этим коммитом |
| 5 | CSP-концессия `style-src-attr` | BACKLOG | расхождение → верно | `frontend/svelte.config.js:19` — CSP `style-src` концессия | исправлено этим коммитом |
| 6 | SSE-секция: browser polling vs Pub/Sub | BACKLOG | расхождение → верно | Pub/Sub service-to-service `services/graph-service/internal/subscriber/pubsub.go:46`; браузер поллит HTTP — текст исправлен | исправлено этим коммитом |
| 7 | ✅-секции (manual testing, cockpit UI, graph-service 1-5, publish, bookmarklet) | BACKLOG | устарело → убрано | CHANGELOG `[Unreleased]` покрывает сделанное | `93c06bc` |
| 8 | Ссылка `API_TEST_COVERAGE_PLAN.md` | BACKLOG | расхождение → верно | `docs/archive/API_TEST_COVERAGE_PLAN.md:1` — перенесён (`93c06bc`) | исправлено — `93c06bc` |
| 9 | nginx-каталог в gateway verification | BACKLOG | расхождение → верно | `docker/nginx` не существует — ссылка на реальный `nginx.conf:57` | исправлено этим коммитом |
| 10 | Остальные backlog-пункты (планы, не факты) | прочее | верно как план | — | нет |

## B.2 `docs/product/USER_PROMISES.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Промис: удалённые заметки восстановимы 90 дней, связи удаляются вместе | весь | верно | `TestNoteSoftDelete_*` в `note_soft_delete_integration_test.go`, `note_repo.go:180-227` | нет |

## B.3 `docs/product/BOOKMARKLET.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | `POST /api/v1/import/bookmarklet`, JWT/cookie auth | API-секция | верно | `router.go:175` | нет |
| 2 | `/import` страница с query-параметрами, батчинг ≤50, UTF-8 | прочее | верно | `router.go:175-177`, `application/import/service.go:31 MaxBatchSize=50`, `routes/import/` | нет |

## B.4 `docs/product/LINK_TYPES.md` + `LINK_TYPES_RU.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | 6 типов: цвета, dash-паттерны, дефолтные веса | таблицы обоих файлов | верно | `entities/link/model/link-type.ts:102-158` (цвета #3366ff/#ff6600/#999999/#ff66ff/#2dd4bf/#f472b6, dash [10,3]/[6,4]/[2,6]) | нет |
| 2 | `related` → `[6,4]` при weight<0.3; width `max(1,w*4)`; opacity `0.4+w*0.4` | «Visual encoding» | верно | `link-type.ts:87`, `link-renderers.ts:150`, `link-type.ts:72` | нет |
| 3 | `source_type` badge: `user`/`gamma`; confirm→`user`+`metadata.gamma`; «Not related»→`link_suppressions` | UI-integration | верно | `LinkTooltip.svelte:85,121`, migration 024 `CHECK IN ('user','gamma')`, `link_suppressions` | нет |
| 4 | API: `POST /links`, `PUT /links/:id`, graph payloads | API-секция | верно | `backend/cmd/server/router.go:205` (links-маршруты, строки 205-208) | нет |
| 5 | Пути компонентов | «Related files» | верно | `frontend/src/features/graph-forms/` (`LinkTypeSelector.svelte`, `LinkTypeLegend.svelte`), `frontend/src/widgets/cosmic-cockpit/CockpitNoteDetails.svelte:1` | нет |

## B.5 `docs/product/LINKS_CHEATSHEET.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Типы/цвета/стили связей | таблица | верно | `frontend/src/entities/link/model/link-type.ts:19` | нет |
| 2 | Веса α=0.5/β=0.5/γ=0.2, `final_weight` формула | «Расчёт веса» | верно | `knowledge-graph.config.json:48-50` | нет |
| 3 | Прокси allowlist/blocklist | «Безопасность прокси» | расхождение → верно | `hooks.server.ts:25,114` — allowlist включает `authorization`,`cookie`; blocklist — hop-by-hop без `host` | исправлено этим коммитом |
| 4 | Файлы и тесты | «Файлы», «Тесты» | верно | `frontend/src/entities/graph-canvas/lib/renderer.ts`, `backend/internal/application/recommendation/refresh_service.go:21`, `backend/internal/domain/graph/bfs.go:23` | нет |

## B.6 `docs/product/GRAPH_LINKS_VISUALIZATION.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Таблица типов + `link-type.ts` интерфейс (icon/label/color/lineDash/getColor/creatable) | :3-22 | верно | `frontend/src/entities/link/model/link-type.ts:19` (`defaultWeight`, `lineDash`) | нет |
| 2 | Weight calc: default 0.5, α/β/γ=0.5/0.5/0.2, `last_weight_update` | :24-50 | верно | `knowledge-graph.config.json` seeds, `backend/internal/application/recommendation/refresh_service.go:21` | нет |
| 3 | Visual encoding: width/opacity формулы, direct vs recommended | :51-95 | расхождение → верно | `frontend/src/entities/graph-canvas/lib/light/recommendations.ts:12` (pale dashed on hover, GRAPH-LIGHT-1); формулы width/opacity верны | исправлено этим коммитом |
| 4 | Rendering: `drawLink`, `linkOpacity`, `dyingLinks`, `BIDIRECTIONAL_LINK_OFFSET=24` | :82-117 | верно | `link-renderers.ts:12`, `incremental.ts` | нет |
| 5 | `source_type` badge: `user`/`auto`/`worker` | :126 | расхождение → верно | `backend/migrations/024_add_links_source_type.up.sql:1` — CHECK user/gamma; документ исправлен | исправлено этим коммитом |
| 6 | Legend/filtering: `hiddenLinkTypes`, `minLinkWeight`, `visibleLinks` | :130-136 | верно | `frontend/src/shared/stores/graph.svelte.ts:17` (`hiddenLinkTypes`, `minLinkWeight` — строки 17-30) | нет |
| 7 | Graph data JSON-контракт с `id` | :137-151 | верно | `backend/internal/interfaces/api/graphhandler/graph_handler.go:45` (GraphLink DTO) | нет |

## B.7 `docs/product/ANOMALY_TYPES.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | 4 аномальных типа, hash%4 детерминированный dispatch | :14-52, :96-142 | верно | `node-renderers.ts:647-658` (`drawUnknown`, `hash % 4`, `stringHash`) | нет |
| 2 | `CelestialBody.ANOMALIES`, `UI_TYPES`, `isAnomaly`, `anomalyType` | :84-93, :203-213 | верно | `celestial-body.ts:429-543` | нет |
| 3 | Конфиг-блок `anomaly.*` | :146-182 | верно | `shared/config/config.ts:100-128` | нет |
| 4 | Тестовый файл путь | :221 | расхождение → верно | `frontend/src/widgets/graph-canvas/GraphCanvas.node-types.spec.ts:1` | исправлено этим коммитом |
| 5 | `renderer.test.ts` anomaly describes | :240-257 | верно | `renderer.test.ts:47-56` | нет |
| 6 | Пути `celestial-body.ts`, `config.ts` | Related | верно | `frontend/src/entities/shared/model/celestial-body.ts:69`, `frontend/src/features/graph-3d/config.ts:11` — существуют | нет |

## B.8 `docs/product/CELESTIAL_BODY_SEMANTICS.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Иерархия/семантика типов | :5-41 | верно (нормативная документация) | `frontend/src/entities/shared/model/celestial-body.ts:69` (`isUi`) | нет |
| 2 | UI_TYPES — 11 селектируемых типов, сортировка по `scaleRank` | :42-58 | верно | `frontend/src/entities/shared/model/celestial-body.ts:540` (`UI_TYPES` сортировка по `scaleRank`) | нет |
| 3 | Singularity — UI-only drop zone | :60-68 | верно | `frontend/src/features/graph-interaction/drag-and-drop.ts:1` — singularity UI-only drop zone | нет |
| 4 | `getChildSuggestion` таблица | :77-90 | расхождение → верно | `celestial-body.ts:164-184`: `planet`→`moon`, `default`→`star` | исправлено этим коммитом |
| 5 | Parent-линк `defaultWeight` 0.9 | :94 | верно | `frontend/src/entities/link/model/link-type.ts:19` — parent `defaultWeight` 0.9 | нет |
| 6 | `getVariation`, `color-schemes.ts` пути | :110-118 | верно | `frontend/src/shared/utils/variation.ts:1`, `frontend/src/shared/lib/graph/color-schemes.ts:1` | нет |
| 7 | `moon` user-selectable, NOTE-TYPE-TAXONOMY заметка | :120-128 | верно | `frontend/src/entities/shared/model/celestial-body.ts:69` — `isUi:true` у moon | нет |

## B.9 `docs/product/FRONTEND_FEATURES.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Double-tap zoom: логика и пороги (<300ms, <30px) | :9-83 | верно | `features/graph-interaction/zoom-pan.ts:44-111` | путь исправлен этим коммитом (было `components/organisms/GraphCanvas.svelte` — теперь `zoom-pan.ts`+`event-bridge.ts`) |
| 2 | Spec `GraphCanvas.interactions.spec.ts` | :101 | расхождение → верно | `frontend/src/widgets/graph-canvas/GraphCanvas.interactions.spec.ts:1` | исправлено этим коммитом |
| 3 | `box-sizing: border-box` global | :116-123 | верно | `shared/styles/global.css:246-250` | нет |
| 4 | Graph page layout `.graph-page`/`.graph-container` + media queries | :125-147 | расхождение → верно | страница переписана на `GraphPageShell` — `frontend/src/routes/graph/+page.svelte:1` | помечено superseded этим коммитом |
| 5 | CockpitNoteDetails (`max-height:100vh`) | :149-156 | верно | `frontend/src/widgets/cosmic-cockpit/CockpitNoteDetails.svelte:1` | нет |
| 6 | Carbon theme: `--carbon-*` токены, компоненты | :197-227 | верно | `frontend/src/shared/styles/global.css:1` (`--carbon-*`), `frontend/src/components/atoms/Button.svelte:1` | нет |
| 7 | Note detail page: hero, chips, tags→`/search?q=`, similar notes, actions, `/graph/3d/{id}` | :230-259 | верно | `routes/notes/[id]/+page.svelte:7,134,231,273,281`; `routes/graph/3d/` существует | нет |

## B.10 `docs/product/UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Баннер актуализации DOC-AUDIT-2 (2026-09-26): имена относятся к июльскому состоянию | :5 | верно | `frontend/src/features/graph-ui/GraphTopBar.svelte:1` — баннер актуализации | нет |
| 2 | Содержательные находки (два потока создания, два поиска, FSD-нарушение) | весь | верно как историческая фиксация | GraphCanvas переехал в `widgets/` (§4.3 закрыт переездом); остальное помечено датой | нет |

## B.11 `docs/product/NOTE_ERROR_CORRECTION_PLAN.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | План QuickEdit/autocompletion/DustInbox/NLPCorrection | весь | верно как план (статус «⏳ Запланировано») | — | нет |
| 2 | QuickCaptureWidget как аналог | :263 | верно | `frontend/src/widgets/quick-capture/QuickCaptureWidget.svelte:1` | нет |

## B.12 `docs/product/OBSIDIAN_IMPORT_SPEC.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Спека импорта Obsidian: `POST /import/obsidian`, очередь, фазы 5.0-5.2 | весь | верно как спека-цель (checkbox'ы пустые; `/import/obsidian` в роутере нет — документ не заявляет реализацию) | `router.go` — маршрута нет; статусы фаз нечекнуты | нет |

## B.13 `docs/product/IDEAS.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Гипотезы/идеи (фазы 13-19, конфиг в БД, motion-by-metrics) | весь | верно как идеи (явно маркированы «not commitments») | — | нет |
| 2 | `[x]`-пункты: fade-in анимация, gamma color-coding, link_type веса | :187-188, :203 | верно | `frontend/src/entities/graph-canvas/lib/incremental.ts:1` (fade), `link-renderers.ts`, `links.link_type` | нет |

## B.14 `docs/product/UX_GUIDELINES_EN.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | UX-нормы | весь | утверждений о коде нет | — | нет |

## B.15 `ROADMAP.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Now: статусы аудит-блокеров | :13-17 | расхождение → верно | header stripping — `nginx.conf` ✓; SKIP_AUTH restriction ✓; `Cache-Control: private`+Vary — `router.go:32-41` ✓ (добавлено в статус); 019 test user — миграция везде, не gated APP_ENV — открыто | исправлено этим коммитом |
| 2 | Now: honest regression codes, 3D readiness, coverage gate, BDD runner | :18-20 | верно | `scripts/testing/test-a3-exit-codes.ps1`, `frontend/src/features/graph-3d/lib/engine.ts:49` (`isReady`), `.github/workflows/_core-checks.yml` | нет |
| 3 | Next: multilingual embeddings (модель уже multilingual, open work = e5), keyword normalization, clustering, link types UI, delta flicker | :21-30 | верно | модель multilingual — `nlp-service/app/main.py:82`; open work помечен | нет |
| 4 | Later/Exploring списки | :31-42 | верно как план | — | нет |
| 5 | Phase 21/22 описания | :43-70 | верно как план | — | нет |

## B.16 `ROADMAP.ru.md`

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | RU-зеркало указателя на EN + фазы 21/22 | весь | верно | `ROADMAP.ru.md` зеркалит `ROADMAP.md:18` (Now/Next/Later) | нет |

## B.17 `CHANGELOG.md` — раздел `[Unreleased]` (:8-93)

| # | Утверждение | Место | Вердикт | Доказательство | Действие |
|---|---|---|---|---|---|
| 1 | Multilingual embedding foundation: `model_name` колонка, model-filtered reads, `embed-recompute` CLI, `NLP_MODEL_NAME` | :11-14 | верно | `backend/migrations/030_add_embedding_model_name.up.sql:2` (`model_name`), `backend/cmd/embed-recompute/main.go`, `knowledge-graph.config.json:281` (NLP_MODEL_NAME) | нет |
| 2 | Agent protocol/handoff board/log | :15-16 | верно | `docs/AI_AGENT_PROTOCOL.md:1`, `docs/AI_HANDOFF.md:1`, `docs/AI_LOG.md:1` | нет |
| 3 | LICENSE MIT | :17-18 | верно | `LICENSE:1` — MIT | нет |
| 4 | Personal-stack destructive guard | :19-20 | верно | `scripts/devops/guard-personal-data.py:1`, `scripts/devops/backup-policy.env` | нет |
| 5 | CI drift guard | :21 | верно | `.github/workflows/_core-checks.yml:1` — CI drift guard | нет |
| 6 | Seeded test user + APP_ENV test profile | :22 | верно | `backend/migrations/019_add_test_user.up.sql:1` | нет |
| 7 | URL-HEADING-1 stage A: title candidates, outline, `IMPORT_CONTENT_MAX_RUNES` 20000, `noise_dropped`, `related_links` ≤20, golden suite 19 | :23-30 | верно | `backend/internal/application/import/service.go:35` (IMPORT_CONTENT_MAX_RUNES budget) | нет |
| 8 | LINKS-2: three states, confirm→`source_type='user'`+`metadata.gamma`, reject→`link_suppressions`, regenerate --dry-run | :31-37 | верно | `backend/internal/interfaces/api/linkhandler/link_handler.go:1`, `backend/migrations/034_create_link_suppressions.up.sql:4` | нет |
| 9 | SYNC-1 stage A: delta от снапшота, resync/NotFound, added_links, write-path guard | :41-46 | верно | `scripts/testing/check-graph-write-paths.mjs:1`, `services/graph-service/internal/api/http_server.go:1` | нет |
| 10 | BREAKING rename public endpoint → `/api/v1/graph/public` | :50 | верно | `router.go:210` | нет |
| 11 | Honest regression exit codes | :51-53 | верно | `scripts/testing/test-a3-exit-codes.ps1:1`, `scripts/testing/run-full-test-cycle.ps1` | нет |
| 12 | Auto-commit removed from regression | :54 | верно | `scripts/testing/run-full-test-cycle.ps1:1` — авто-коммит убран | нет |
| 13 | 3D readiness после первого кадра | :55-56 | верно | `frontend/src/features/graph-3d/lib/engine.ts:49` (`isReady`), `data-test-stable` | нет |
| 14 | Import/bookmarklet cap 50000 runes vs `IMPORT_CONTENT_MAX_RUNES` | :57-59 | верно | `service.go:38 maxContentLen=50000` | нет |
| 15 | Fixed: Yandex OAuth, fog densities, docker cleanup exit codes/-DryRun/backup gate/-Full/-WslOptimize | :61-77 | верно | `nginx.conf` auth routes, `knowledge-graph.config.json` fog, `scripts/devops/stop-docker.ps1:1` | нет |
| 16 | Security: `RequireNoteAccess` на всех `/notes/:id` (404 IDOR) | :79-84 | верно | `router.go:141+`, `middleware/note_access.go` + тесты | нет |
| 17 | Security: OAuth token transport, `SKIP_AUTH` test-profile, `Cache-Control: private` | :85-92 | верно | oauth handlers, graph-service config, `router.go:32-41` | нет |

---

# Этапы C и D — предварительная сводка

Детализация построчно — при сдаче соответствующих этапов. Сводка прошлых проходов сохранена
ниже и является черновой; строки с «~» требуют перечисления утверждений при сдаче этапа.

| Документ | Утверждений | Верно | Расхождение | Нет в коде | Устарело |
|---|---|---|---|---|---|
| operations/DOCKER.md | ~25 | ~20 | 5 (frontend нет host-порта — только nginx:18081; Redis dev=16379/personal=16380, не 6379/6380; graph-service строка «gRPC» → 9091 HTTP + 9090 gRPC без клиентов) | — | — |
| operations/STACK_CONFIGURATION_COMPARISON.md | ~20 | ~16 | 4 (те же redis-порты; frontend dev не опубликован) | — | — |
| operations/DEPLOYMENT_EN.md | ~40 | ~25 | 10 (`migrate` CLI не существует ×4; `./seed` → `./test-seed` и только APP_ENV=test; `health-check.sh` ×2 несуществует; `/db-check` нет; `docker-compose.monitoring.yml` нет; порт 18086 — тест-стек) | 2 (k8s/ + monitoring — «target, not implemented») | — |
| operations/CONFIGURATION_EN.md | ~50 | ~40 | 6 (env→JSON-only ключи; `BACKUP_CLOUD_PROVIDER` дефолт `r2`) | 1 (`backup.draft_ttl_hours` — мёртвый ключ) | — |
| operations/CONFIGURATION_RU.md | ~30 | ~25 | 5 (env→JSON; `GRAPH_MAX_NODES` удалён) | — | — |
| operations/TESTING.md | ~40 | ~40 | — | — | — |
| operations/TESTING_COMMANDS.md | ~20 | ~20 | — | — | — |
| operations/REGRESSION_TEST_PLAN.md | ~30 | ~30 | — | — | — |
| operations/BACKUP.md | ~40 | ~40 | — | — | — |
| operations/ARGOS.md | ~15 | ~15 | — | — | — |
| operations/MANUAL_TEST_CHECKLIST_*.md | ~15 | ~15 | — | — | — |
| api/API_EN.md | ~50 | ~50 | — | — | — |
| api/RECOMMENDATION_API.md | ~25 | ~18 | 6 (`X-Recommendations-Source` реальные значения; header всегда) | — | — |
| api/API_ERRORS_EN.md | ~30 | ~27 | 3 (+`DUPLICATE_LINK`/`INVALID_UUID`/`INVALID_REQUEST`; 429 без `code`) | — | — |
| backend/openAPI.yaml | ~3400 | покрыт `router_contract_test.go` — дрейф невозможен | — | — | — |
| README.md (корень) | ~10 | ~10 | — | — | — |
| DEPLOY.md / DEPLOY.ru.md | ~60 | ~55 | — | — | — |
| BOOTSTRAP.md | ~10 | ~10 | — | — | — |
| COMMANDS.md | ~40 | ~38 | 2 (health-порты backend 9000, NLP 5000 — контейнерные) | — | — |
| docs/AGENTS.md | ~10 | ~10 | — | — | — |
| docs/LICENSES.md | — | — | — | — | — (самопроверка `check-licenses.mjs`) |
| docs/PROJECT_REVIEW_AI_AGENTS.md | ~30 | ~27 | 3 (локальное однопользовательское; pgvector-go v0.4.1; GORM-версия) | — | — |
| docs/agents/*.md | ~15 | ~15 | — | — | — |
| TZ-Java-source-text-handler-2026-08-30.md | — | — | — | — | — (статус «ожидает реализации» точен) |
| backend/README.md | ~10 | ~9 | 1 (English messages в response.go) | — | — |
| backend/internal/domain/**/README.md | ~15 | ~15 | — | — | — |
| frontend/README.md | — | — | 1 (заменён стоковый шаблон) | — | — |
| frontend/src/shared/services/README.md | ~15 | ~15 | — | — | — |
| .windsurfrules | ~40 | ~39 | 1 (нет `frontend/src/app/`) | — | — |
| .devin/skills/*, .devin/prompts/* | ~25 | ~25 | — | — | — |
