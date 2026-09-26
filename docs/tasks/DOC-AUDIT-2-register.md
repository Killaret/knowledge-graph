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
| decisions/001–018 | ~60 | ~50 | — | 6 (статусы «Implementation status» у 003/006/007/008/009/010/013/014; 004 — частично, 011 — реализовано; **RBAC богаче, чем казалось:** `user_roles`+`role_permissions`+`permission_repo`+middleware существуют — нет только tenant-скоупа) | 1 (список в README — легендарная нумерация) |
| c4/context.puml | ~8 | 4 | 4 (нет graph-service/worker/mongo) | — | — |
| c4/container.puml | ~10 | 5 | 4 (нет worker/graph-service/mongo; «HTTPS/WebSocket») | — | — |
| c4/component.puml | ~12 | 6 | 4 (Command/Query Bus нет в коде; Event Bus = Redis publisher) | — | — |
| uml/er-diagram.puml | ~15 | ~12 | — | — | добавлен список всех 21 таблиц как комментарий |
| uml/sequence-create-note.puml | ~15 | ~8 | 5 (шины, имена задач, `POST /api/v1/notes`) | — | — |
| uml/sequence-suggestions.puml | ~15 | ~9 | 4 (фактическая 5-уровневая цепочка, ключ `recommendations:*`, α=0.5/β=0.5/γ=0.2) | — | — |
| uml/deployment-local.puml | ~12 | ~7 | 4 (worker — отдельный контейнер kg-worker; mongo/graph-service; порт backend 9000→8080) | — | — |
| uml/class-domain.puml | ~20 | ~16 | 2 (LinkType +parent/+child; сигнатура NewNote и методы) | — | — |
| operations/DOCKER.md | ~25 | ~20 | 5 (frontend нет host-порта — только nginx:18081; Redis dev=16379/personal=16380, не 6379/6380; graph-service строка «gRPC» → 9091 HTTP + 9090 gRPC без клиентов) | — | — |
| operations/STACK_CONFIGURATION_COMPARISON.md | ~20 | ~16 | 4 (те же redis-порты; frontend dev не опубликован) | — | — |
| operations/DEPLOYMENT_EN.md | ~40 | ~25 | 10 (`migrate` CLI не существует ×4 — миграции только авто при старте `server`; `./seed` → `./test-seed` и только APP_ENV=test; `health-check.sh` ×2 несуществует; `/db-check` нет; `docker-compose.monitoring.yml` нет; порт 18086 — тест-стек, не dev) | 2 (k8s/ раздел + monitoring — помечены «target, not implemented») | — |
| operations/CONFIGURATION_EN.md | ~50 | ~40 | 6 (env→JSON-only: `SERVER_FALLBACK_PORTS`, `GRAPH_DEFAULT_DEPTH`, `GRAPH_STREAM_CHUNK_SIZE`, `FRONTEND_ACHIEVEMENTS_POLL_INTERVAL_MS`, `BACKUP_DRAFT_TTL_HOURS`; `BACKUP_CLOUD_PROVIDER` дефолт `r2`, а не `yandex`) | 1 (`backup.draft_ttl_hours` — парсится, но никем не читается — мёртвый ключ) | — |
| operations/CONFIGURATION_RU.md | ~30 | ~25 | 5 (те же env→JSON; `GRAPH_MAX_NODES` — выдуманная переменная, удалена) | — | — |
| operations/TESTING.md | ~40 | ~40 | — | — | — (порты/команды тест-стека точны; `run-bdd.cjs`, recompute-команды верны) |
| operations/TESTING_COMMANDS.md | ~20 | ~20 | — | — | — |
| operations/REGRESSION_TEST_PLAN.md | ~30 | ~30 | — | — | — (`cleanup-test-artifacts.py`, `/graph/3d/[id]` — существуют) |
| operations/BACKUP.md | ~40 | ~40 | — | — | — (актуален: KG_BACKUP_DIR, backup_scheduler, ретенции) |
| operations/ARGOS.md | ~15 | ~15 | — | — | — (проекты visual/visual-real-auth, spec-файлы — на месте) |
| operations/MANUAL_TEST_CHECKLIST_*.md | ~15 | ~15 | — | — | — |
| api/API_EN.md | ~50 | ~50 | — | — | — (свежий, проверен против стека) |
| api/RECOMMENDATION_API.md | ~25 | ~18 | 6 (в цепочке пропущен live graph-service/BFS шаг; реальные значения `X-Recommendations-Source`: `table`/`graph-service`/`semantic`/`redis`/`empty`, а не `*-fallback`; header ставится всегда) | — | — |
| api/API_ERRORS_EN.md | ~30 | ~27 | 3 (+`DUPLICATE_LINK`/`INVALID_UUID`/`INVALID_REQUEST`; 429 без поля `code` — `error`+`retry_after`; `RATE_LIMIT_EXCEEDED` не существует) | — | — |
| backend/openAPI.yaml | ~3400 | покрыт `router_contract_test.go` — дрейф невозможен | — | — | — |
| ROADMAP.md | ~15 | ~12 | 3 («English-only модель» — реально multilingual MiniLM, план = e5 (MODEL-2); «word forms as raw strings» — `/normalize` уже в worker-пайплайне; статусы блокеров Now: заголовки снятия на gateway сделано, 019 test user и Cache-Control — ещё открыты; строки 2–3 Now фактически выполнены) | — | — |
| ROADMAP.ru.md | ~5 | ~5 | — | — | — (указатель на EN + фазы 21/22 — план, не факты) |
| CHANGELOG.md | ~30 | ~30 | +1 (добавлена запись SYNC-1 A — дельта от снимка, resync, события на путях записи, сторож) | — | — |
| product/BACKLOG.md | ~50 | ~35 | 9 (убраны ✅-секции: manual testing, cockpit UI, graph-service 1–5, publish/unpublish, bookmarklet/mass-import; строка «events.Publisher не проведён» — устарела после SYNC-1 A; ссылка на `API_TEST_COVERAGE_PLAN.md` → docs/archive/; дата обновления) | — | — |
| product/FRONTEND_FEATURES.md | ~10 | ~9 | 1 (NoteSidePanel → CockpitNoteDetails.svelte) | — | — |
| product/UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md | ~15 | ~15 | 1 (баннер актуализации: FloatingControls→GraphTopBar, NoteSidePanel→CockpitNoteDetails) | — | — |
| product/LINK_TYPES.md + _RU | ~30 | ~30 | — | — | — (все 6 типов, parent/child — есть) |
| product/LINKS_CHEATSHEET.md | ~15 | ~14 | 1 (путь renderer.ts → entities/graph-canvas/lib/) | — | — |
| product/NOTE_ERROR_CORRECTION_PLAN.md | ~10 | ~9 | 1 (QuickCaptureWidget → widgets/quick-capture/) | — | — |
| product/ANOMALY_TYPES.md, CELESTIAL_BODY_SEMANTICS.md, IDEAS.md, BOOKMARKLET.md, OBSIDIAN_IMPORT_SPEC.md, UX_GUIDELINES_EN.md, GRAPH_LINKS_VISUALIZATION.md | ~60 | ~58 | — | — | — (пути/эндпоинты выборочно подтверждены; спеки-планы помечены статусами) |
| README.md (корень) | ~10 | ~10 | — | — | — (таблица портов верна) |
| DEPLOY.md / DEPLOY.ru.md | ~60 | ~55 | — | — | — (PERSONAL_REDIS_URL — compose-уровень, верен; `migrate/migrate` внешний образ — валидно; внутренние порты контейнеров корректны) |
| BOOTSTRAP.md | ~10 | ~10 | — | — | — (check-disk-layout.ps1, start-test.ps1 — существуют) |
| COMMANDS.md | ~40 | ~38 | 2 (health-порты: backend 8080→9000, NLP 8000→5000 — контейнерные порты вместо хостовых) | — | — |
| docs/AGENTS.md | ~10 | ~10 | — | — | — (роли совпадают с AI_AGENT_PROTOCOL) |
| docs/LICENSES.md | — | — | — | — | — (самопроверяется `check-licenses.mjs` — зелёный, 87 депенденси) |
| docs/PROJECT_REVIEW_AI_AGENTS.md | ~30 | ~27 | 3 («мультитенантное SaaS» → локальное однопользовательское, мультитенантность — цель; pgvector-go v0.2.0→v0.4.1; GORM-версия в .windsurfrules отстаёт от go.mod v1.31.2 — зафиксировано как мягкое) | — | — |
| docs/agents/*.md | ~15 | ~15 | — | — | — (процессные документы) |
| TZ-Java-source-text-handler-2026-08-30.md | — | — | — | — | — (статус «ожидает реализации» — точен, сервиса нет) |
| backend/README.md | ~10 | ~9 | 1 («Russian messages» → English — реальные сообщения в response.go на английском) | — | — |
| backend/internal/domain/**/README.md | ~15 | ~15 | — | — | — (пакеты совпадают с деревом) |
| frontend/README.md | — | — | 1 (был стоковый шаблон `sv create` — заменён на реальный: стек, команды, порты прокси, FSD-карта) | — | — |
| frontend/src/shared/services/README.md | ~15 | ~15 | — | — | — (API PreloadService совпадает) |
| .windsurfrules | ~40 | ~39 | 1 (`frontend/src/app/` — такого каталога нет; entry points живут в корне src/) | — | — (тест-порты, команды, правила — сверены) |
| .devin/skills/*, .devin/prompts/* | ~25 | ~25 | — | — | — (пути существуют; kg-graph-3d signal/clip-факты соответствуют engine.ts) |
