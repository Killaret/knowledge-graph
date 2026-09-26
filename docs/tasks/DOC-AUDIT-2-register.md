# DOC-AUDIT-2. Реестр сверки документации с кодом

Постановка — [`DOC-AUDIT-2-docs-vs-code.md`](DOC-AUDIT-2-docs-vs-code.md). Метод: каждое проверяемое
утверждение о текущей системе сверено с кодом по норме «Verifying a Finding»; поведение — исполнением
где возможно. Правки документов входят в те же коммиты, что и строки реестра о них.

Вердикты: **верно** · **расхождение** (код делает иначе) · **нет в коде** (описано, но отсутствует —
в тексте ставится пометка, строка уходит владельцу) · **устарело** (в плане то, что сделано).

## Нет в коде — на решение владельца

| # | Утверждение | Где | История | Предложение |
|---|---|---|---|---|
| A-1 | SSE/WS-push канал до клиента | ARCHITECTURE_EN.md (диаграмма, сноска ¹) | Не было никогда: SSE по ADR 014 — постановка SYNC-1 этап C | В работе по SYNC-1; пометка стоит |
| A-2 | gRPC API помечен «Primary» | ARCHITECTURE_EN.md, раздел Graph Service | Сервер и контракт есть (`proto/graph_layout.proto`, :9090), но клиентов нет — backend/frontend ходят по HTTP :9091. Замысел владельца (2026-09-26): gRPC как фолбэк при перегрузке HTTP, а не постоянный транспорт — см. `DOC-AUDIT-2-docs-vs-code.md`, «Замысел владельца» | Решение владельца: реализовать пороговый фолбэк или снять gRPC до потребителя |
| A-17 | PostgreSQL RLS + `tenant_id` (мультитенантность) | ARCHITECTURE_SUMMARY.md (exec summary, C4, JWT, security layers) | Нет ни одной миграции с `CREATE POLICY`/`tenant`; JWT без `tenant_id`. 1.0 — локально, один пользователь (решение 67) | Пометки проставлены; скорее всего «вычеркнуть до SaaS» |
| A-18 | Circuit breaker `sony/gobreaker` | ARCHITECTURE_SUMMARY.md (tech stack, «CB protected», мониторинг) | Нет в `go.mod`; NLP-клиент — прямой HTTP | Владельцу: нужен ли CB до 1.0 |
| A-19 | Audit logging «90-day retention», MongoDB audit logs | ARCHITECTURE_SUMMARY.md (exec summary, C4, backup) | Таблица `audit_log` (миграция 016) и `AuditLogModel` есть, **писателей нет** — ни в PG, ни в Mongo | Владельцу: задел есть, механизм не заведён |
| A-20 | TLS 1.3 «for all connections» | ARCHITECTURE_SUMMARY.md (security layers) | Dev/personal стеки — plain HTTP, `sslmode=disable` | Пометка; решение — деплойная тема |
| A-26 | ADR 004 «soft delete» для заметок | ADR 004 + `note_repo.go` | `NoteModel.DeletedAt` — `*time.Time` (не `gorm.DeletedAt`), поэтому `NoteRepository.Delete`/`DeleteBatch` делают **жёсткий DELETE**; `Restore` и индекс `idx_notes_deleted_at` — мёртвый код. Ссылки при этом действительно мягко удаляются | **Кандидат-дефект**: либо владелец подтверждает hard-delete для заметок (тогда убрать Restore/колонку), либо чинить модель |
| A-27 | Вертикаль ADR 003/006/007/008/009/010 как «принятое решение» | `decisions/003…010` | Решения приняты для SaaS-эволюции, но текст читается как действующая архитектура; RLS/tenancy/audit/CB в коде отсутствуют | Проставлены строки «Implementation status» (006 — частично: `role` claim + `RequireNoteAccess`) |
| A-28 | Каналы `cache:invalidate:*` и payload с `tenant_id` | ADR 014 | Скетч показывает per-entity каналы; реально — один канал `graph:events` с типизированными событиями; снимки `snapshot:*` событиями не инвалидируются (SYNC-1) | Дописка о реальной схеме проставлена |
| A-29 | Sidebar/CCC «stub, width:0, hidden» | FRONTEND_ARCHITECTURE_EN.md | Компонента нет в `frontend/src` вообще | Помечено «not implemented» |
| A-30 | Сторож `check-decisions.mjs` краснел на RELEASE-1 | `scripts/testing/check-decisions.mjs` | Правило 2 помечало строки `used` при любом совпадении id: маркер `MODEL-2` (дата 09-24) забирал строку 67 (RELEASE-1, 09-26), потому что строка 67 упоминает MODEL-2 как составную часть; у маркера MODEL-2 отдельной строки нет — это цитата решения 60 | Сторож исправлен: id-совпадение сначала предпочитает строку с той же датой, затем разрешён общий id даже на занятой строке (несколько файлов могут цитировать одно решение) |

## Расхождения — исправлены в документах

| # | Утверждение | Где | Что было | Что стало | Коммит |
|---|---|---|---|---|---|
| A-3 | Таблица эндпоинтов | ARCHITECTURE_EN.md | `GET /links`, `GET /graph`, `GET /graph/3d` — таких маршрутов нет | Реальные маршруты из `cmd/server/router.go` (+`/notes/:id/links`, `/graph/public`, `/graph/analytics`, `/me/graph/*`) | — |
| A-4 | `domain/draft/` как отдельный домен | ARCHITECTURE_EN.md | Каталога нет; поля entity выдуманы (nullable noteID, DraftContent) | Реально: `domain/note/draft.go`, `application/draft/`, `infrastructure/mongo/draft_repo.go` | — |
| A-5 | `domain/achievement/repository.go` | ARCHITECTURE_EN.md | Файла нет | `Repository` — интерфейс внутри `entity.go`; добавлен `model.go` | — |
| A-6 | `traversal_integration_test.go` | ARCHITECTURE_EN.md | Файла нет | `graph_extra_test.go` | — |
| A-7 | «27 SQL migration files» | ARCHITECTURE_EN.md | Устарело | 68 файлов (подсчёт `backend/migrations/*.sql`) | — |
| A-8 | `shared/api/achievements.ts` | ARCHITECTURE_EN.md | Файла нет | Достижения — `entities/achievement/`; добавлены реальные api-файлы (auth, import, sharing, users, quality, errorMessage) | — |
| A-9 | Сторы `auth.svelte.js`, `achievements.ts`, `notes/graph/ui.ts` | ARCHITECTURE_EN.md | Имена/набор не те | `auth.svelte.ts`, `auth-session.svelte.ts`, `graph.svelte.ts`, `graph-view.svelte.ts`, `lexicon-settings.ts` | — |
| A-10 | Плоская таблица «Core Components (46)» | ARCHITECTURE_EN.md | Компоненты переехали по FSD-слоям; Graph3D/LazyGraph3D/NoteSidePanel/FloatingControls с такими именами отсутствуют | Таблица слоёв: widgets/features/entities/components (atoms/molecules/organisms) | — |
| A-11 | «3D engine frozen for 1.0» + список файлов | ARCHITECTURE_EN.md | Устарело: 3D активен в `features/graph-3d/`; перечисленные файлы (graph3d.ts, celestial.ts, controls.ts…) не существуют | Раздел переписан под реальные `lib/engine.ts`, `labels.ts`, `fog.ts`… | — |
| A-12 | Список роутов фронтенда | ARCHITECTURE_EN.md | Не хватало /auth/*, /import/*, /profile, /test/* | Добавлены все реальные | — |
| A-13 | NLP API | ARCHITECTURE_EN.md | Только /health, /extract_keywords, /embed | + /normalize, /similarity (`app/main.py`) | — |
| A-14 | `cli` = «Admin CLI», `make dev` | ARCHITECTURE_EN.md | cli — рекомендательный прекомпьют; в Makefile нет `dev` | Исправлены оба | — |
| A-15 | Плоские `handlers/*.go` в графе зависимостей; статистика файлов | ARCHITECTURE_EN.md | Пакеты `notehandler/` и др.; счётчики устарели | Обновлены (35/24/64/27/89/7) | — |
| A-16 | Таблицы `keywords`, `recommendations` | ARCHITECTURE_EN.md | В миграциях — `note_keywords`, `note_recommendations` | Исправлено | — |
| A-21 | JWT claims | ARCHITECTURE_SUMMARY.md | `tenant_id`, `permissions[]` — нет в `TokenClaims` | Реальные поля (`user_id`/`login`/`role`/`token_type`) | — |
| A-22 | Эндпоинты в sequence-диаграмме | ARCHITECTURE_SUMMARY.md | `/api/drafts/autosave`, `/api/notes/publish` | `/api/v1/notes/:id/draft`, `/api/v1/notes/:id/publish` | — |
| A-23 | «Worker → OpenAI», «LPUSH/BRPOP» | ARCHITECTURE_SUMMARY.md | OpenAI нет; очередь — Asynq | NLP service + Asynq | — |
| A-24 | PG 15, MongoDB 6, «MongoDB for Logs» | ARCHITECTURE_SUMMARY.md | pgvector:pg16, Mongo 7; аудит-логов в Mongo нет | Исправлено | — |
| A-25 | Soft-delete 30 дней, draft TTL 30 дней | ARCHITECTURE_SUMMARY.md | cleanup default 90 дней; TTL-индекс драфтов 7 дней (604800 с) | Исправлено | — |

## Реестр по документам — этап A (docs/architecture/, 32 документа)

| Документ | Утверждений | Верно | Расхождение | Нет в коде | Устарело |
|---|---|---|---|---|---|
| ARCHITECTURE_EN.md | ~45 | ~24 | 14 | 2 (SSE/WS, gRPC «primary») | 1 (статистика/таблицы — попутно) |
| ARCHITECTURE_SUMMARY.md | ~40 | ~25 (pipeline-секции gamma/CHUNK-1/NLP-4/NOTE-QUALITY-1 в основном верны) | 10 | 4 (RLS/tenancy, CB, audit writer, TLS) | — |
| ARCHITECTURE_PATTERNS.md | ~12 | 8 | — | 3 (RLS, permissions-claims, gobreaker — помечены «не реализовано») | 1 (устаревший список-«roadmap» в §7) |
| README.md (architecture) | ~18 | ~15 | 1 (старый порядковый список ADR помечен легендарным) | — | 1 |
| FRONTEND_ARCHITECTURE_EN.md | ~35 | ~18 | 9 (FloatingControls/NoteSidePanel/Sidebar-stub/Graph3D.svelte/graphStore/маршруты/BDD-файлы/perf-тезисы) | 2 (Sidebar-компонент, virtual scrolling) | 3 (3D «frozen», /notes/create, stores) |
| GRAPH3D.md | ~30 | ~28 | 2 (FloatingControls→GraphTopBar, NoteSidePanel→CockpitNoteDetails) | — | — |
| GRAPH_SERVICE_AUTH.md | ~20 | ~19 | 1 (порт 29091 уточнён как test-stack; dev — 9091) | — | 1 (находка №1 «JWT из query» — уже исправлена в коде) |
| RECOMMENDATION_ARCHITECTURE.md | ~30 | ~27 | 3 (5-й уровень фолбэка `graph-service`, раздельные флаги fallback-ов, header-таблица) | — | — |
| RECOMMENDATION_TROUBLESHOOTING.md | ~15 | ~13 | 2 (контейнер `kg-worker`, флаг `--batch-size` отсутствует у CLI) | — | — |
| SaaS_DATABASE_SCHEMA.md | — (целиком целевое) | — | — | документ=задел (баннер «target-state, not implemented» + ada-002≠MiniLM) | — |
| WEIGHTS_CALCULATION.md | ~8 | ~8 | — | — | — (λ с 2-го уровня и нормализация `(1-d)/2` подтверждены `bfs.go`) |
| atam.md | ~10 | ~7 | 2 (BFS — итеративный Go, не CTE; лимита «50 первого уровня» нет — только `topN`) | 1 (лимит первого уровня) | — |
| glossary.md | ~25 | ~24 | 1 (Link Type + `parent`/`child`) | — | — |
| clustering.md | — | — | — | — | — (корректно помечена «отложена», с реальными причинами) |
| decisions/001–018 | ~60 | ~50 | — | 6 (статусы «Implementation status» у 003/006/007/008/009/010/013/014; 004 — частично, 011 — реализовано) | 1 (список в README — легендарная нумерация) |
