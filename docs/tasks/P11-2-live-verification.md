# P11-2 — живая верификация и зачистка старой модели (отчёт для ревью)

- **Дата:** 2026-09-07
- **Агент:** Devin
- **Коммиты:** `42f0289`, `1397c39` (реализация P11-2), `a112560` (доработки по результатам верификации)
- **Связанные документы:** [`P11-2-multilingual-embeddings.md`](P11-2-multilingual-embeddings.md), [`../MANUAL_TEST_FEEDBACK.md`](../MANUAL_TEST_FEEDBACK.md)

## Что проверялось

1. Корректно ли повсюду убрана старая модель `all-MiniLM-L6-v2`.
2. Безопасность `scripts/cleanup/cleanup-docker.ps1` / `.sh` вокруг томов.
3. Полный цикл `run-full-test-cycle.ps1` на изолированном тест-стеке.
4. Содержимое данных Personal-стека (по запросу владельца).

## Аудит старой модели

Дефолт `all-MiniLM-L6-v2` оставался в исполняемых местах и заменён на
`paraphrase-multilingual-MiniLM-L12-v2`:

- `docker-compose.yml`, `docker-compose.personal.yml`, `docker-compose.test.yml`
- `backend/.env.example`, `nlp-service/.env.example` (NLP_MODEL_NAME)
- `backend/internal/config/config.go`, `services/graph-service/internal/config/config.go`
- `services/graph-service/internal/db/postgres_client.go` (fallback)
- `backend/internal/infrastructure/db/postgres/embedding_repo.go` (fallback)
- `nlp-service/Dockerfile` (ARG MODEL_NAME), `nlp-service/app/nlp_utils.py`, `nlp-service/entrypoint.sh`
- Документация, описывающая текущее поведение: `docs/architecture/glossary.md`,
  `scripts/backfill/README.md`, оба мастер-промпта.

Оставшиеся упоминания `all-MiniLM-L6-v2` — намеренные исторические: миграция `030`
(помечает старые строки), дизайн-доки P11-1/P11-2, тесты изоляции моделей.

## Исправления по ходу верификации

1. **Прогрев модели в NLP-образе.** `entrypoint.sh` и Dockerfile прогревали модель
   через `SentenceTransformer(name)` — это заполняло `~/.cache/torch`, а не HF
   hub-кэш (`$HF_HOME/hub`), который читает `nlp_utils.py` через
   `snapshot_download`. На чистом кэше модель качалась заново при каждом старте.
   Переведено на `snapshot_download(repo_id='sentence-transformers/<model>',
   cache_dir=$HF_HOME/hub)`; добавлено `HF_HUB_DISABLE_XET=1` (таймауты чтения CDN
   при скачивании 28 файлов модели, ~12 мин).

2. **cleanup-docker.ps1 / cleanup-docker.sh.** Раньше: default-режим делал
   `docker volume prune -f`, а `-Full`/`--full` — `docker system prune -af
   --volumes`. После остановки контейнеров Personal-тома выглядели «неиспользуемыми»
   и попадали под чистку. Теперь:
   - default и `-Full`/`--full` не трогают тома вообще;
   - тома удаляются только при явном `-RemoveVolumes`/`--remove-volumes`, и только
     anonymous dangling; имена с `personal` и метка `com.knowledgegraph.protected=true`
     пропускаются всегда;
   - `COMMANDS.md` и usage-тексты приведены в соответствие.

3. **run-full-test-cycle.ps1 — два дефекта:**
   - `SKIP_AUTH=true` выставлялся глобально и утекал в `go test` —
     `internal/config` падал («SKIP_AUTH is only allowed when APP_ENV=test»).
     Теперь переменная действует только на время `start-test.ps1` (и дальше по
     фазам, где она нужна явно).
   - Детект «dev/personal были запущены» шёл через `docker compose ps -q`, а
     тест-стек разделяет compose-проект `knowledge-graph` — `kg-test-*` давали
     ложное «было запущено», и скрипт после теста пытался поднять dev и personal
     стеки (в т.ч. Personal без явной просьбы — перехвачено и остановлено).
     Теперь детект по точным именам `kg-backend` / `kg-backend-personal`, а
     снапшоты контейнеров исключают `kg-test-*`.

## Результаты проверки

### Живой тест-стек (изолированный)

- NLP `/health` → `{"status":"healthy","model_loaded":true}`.
- `POST /embed` → 384-мерный вектор.
- `POST /similarity`: RU «кошка сидит на окне» ↔ EN "a cat sits on the window" →
  `0.9897`; несвязанная пара → `0.5447`.
- `schema_migrations`: применено до `030` включительно.
- `note_embeddings.model_name` NOT NULL; все строки —
  `paraphrase-multilingual-MiniLM-L12-v2`, `vector_dims=384`.
- `embed-recompute -dry-run`: 0 missing; после ручной пометки одной строки
  `all-MiniLM-L6-v2` — ровно 1 missing; после возврата — 0.
- Сидер: 100 заметок / 60 связей / 100 эмбеддингов; graph-service: 100 узлов.

### Полный регрессионный цикл (`-SkipManual`)

PASS: старт стека, сидер, backend unit + integration, pgvector, Redis+Mongo,
frontend unit + coverage (987/987), E2E SKIP_AUTH (75 пройдено), BDD (5 сценариев
/ 43 шага), real-auth стек и его сидер, визуальные тесты (13/13, Argos build #21),
остановка и состояние стеков.

**FAIL: `chromium-real-auth` — 20 passed / 7 failed**, стабильно в двух прогонах
(полный цикл + изолированный перезапуск на том же образе):

- `cockpit-canvas-controls.spec.ts` — клик по `top-bar-fog` перехватывается
  `.right-cluster` (таймаут 60с); тест zoom/transform падает.
- `floating-auth-panel.spec.ts` — после логина через floating panel
  `graph-stats` не обновляется 20с (граф не перезагружается).
- `note-creation-flows.spec.ts` (×3) и `child-note-flows.spec.ts` — заметка,
  созданная через кнопку `+`, ghost-форму по `N` или панель child-note, не
  появляется в list view (`[data-testid="note-title"]` не находится за 20с).

С P11-2 не связано (фронтенд/auth не менялись). Зафиксировано в
[`../MANUAL_TEST_FEEDBACK.md`](../MANUAL_TEST_FEEDBACK.md) → «Urgent fixes»,
требует триажа.

## Данные Personal-стека (инвентаризация)

Для отчёта поднимались **только** `postgres_personal`, `mongo_personal`,
`redis_personal` (read-only запросы), после чего контейнеры остановлены и удалены;
тома не трогались. Backend/frontend Personal-стека не запускались.

- **PostgreSQL `knowledge_personal`** (~10.2 МБ, 21 таблица):
  - `users`: 1 — `test_user` (создан 2026-08-08);
  - `notes`: 1 — системная «Knowledge Core» (`id 00000000-…-0001`, 579 символов);
  - `links`, `note_embeddings`, `note_keywords`, `tags`, `note_tags`, `note_shares`,
    `note_likes`, `note_recommendations`, `audit_log`, `refresh_tokens`: **0 строк**;
  - `schema_migrations`: применено до **028** — отстаёт на две миграции
    (`029_remove_test_user`, `030_add_embedding_model_name`). Применятся при
    следующем старте personal-backend: `029` удалит только legacy-пользователя
    `id 00000000-…-0000` (если есть), `030` добавит `note_embeddings.model_name`
    (таблица пуста — помечать нечего).
- **MongoDB:** только системные БД (`admin`, `config`, `local`), прикладных данных нет.
- **Redis:** 2 ключа — `asynq:servers`, `asynq:workers` (реестр воркеров, задач нет).

**Вывод:** реальных пользовательских данных в Personal-стеке сейчас нет — только
системная стартовая заметка и тестовый пользователь.

### Бэкап перед чисткой

Сделан сырой архив трёх Personal-томов (логический `pg_dump` недоступен на хосте —
нет `pg_dump` в PATH; скрипт `backup-personal.ps1` требует его):

- `backups/personal-volumes-raw-2026-09-06-215805-pgdata_personal.tar.gz` (7.6 МБ)
- `backups/personal-volumes-raw-2026-09-06-215805-mongodbdata_personal.tar.gz` (4.1 МБ)
- `backups/personal-volumes-raw-2026-09-06-215805-redisdata_personal.tar.gz` (623 Б)

## Текущее состояние

- Все стеки остановлены (тест-стек уничтожен с `down -v`, dev/personal не были
  запущены до проверки и не восстанавливались — ложный детект исправлен).
- Personal-тома целы, данные не менялись.
- Образы пересобраны: `knowledge-graph-*-test` (включая NLP с прогретой
  мультиязычной моделью в кэше образа).
- Коммиты `42f0289`, `1397c39`, `a112560` запушены в `origin/feat/2d-adaptive-fog`.
- Ветка ждёт ревью Claude Code; открытый вопрос — 7 падений `chromium-real-auth`.
