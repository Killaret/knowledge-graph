# DOC-REORG-1 — заметки исполнителя для ревью

**Исполнитель:** Devin · **Дата:** 2026-09-22 · **Основание:** решение владельца 52
(«делай reorg, ничего не потерять, не сломать»), постановкой служит разбор
[`DOC-AUDIT-1-documentation-inspection.md`](DOC-AUDIT-1-documentation-inspection.md) и
принятый [`DOC-AUDIT-1-review-findings.md`](DOC-AUDIT-1-review-findings.md).

## Что сделано

58 файлов и 2 каталога перемещены `git mv` по иерархии из разбора:

- `docs/agents/` — AI_AGENT_SETUP, AI_PROCESS_AUDIT, MANUAL_TEST_FEEDBACK
- `docs/architecture/` — ARCHITECTURE_SUMMARY/_EN/PATTERNS, FRONTEND_ARCHITECTURE_EN,
  SaaS_DATABASE_SCHEMA, GRAPH_SERVICE_AUTH, GRAPH3D, RECOMMENDATION_ARCHITECTURE,
  RECOMMENDATION_TROUBLESHOOTING
- `docs/api/` — API_EN, API_ERRORS_EN, RECOMMENDATION_API
- `docs/operations/` — DOCKER, DEPLOYMENT_EN, STACK_CONFIGURATION_COMPARISON,
  CONFIGURATION_EN/RU, BACKUP, TESTING, TESTING_COMMANDS, REGRESSION_TEST_PLAN, ARGOS,
  MANUAL_TEST_CHECKLIST_COCKPIT/MINIMAL
- `docs/product/` — LINK_TYPES (+RU), LINKS_CHEATSHEET, GRAPH_LINKS_VISUALIZATION,
  CELESTIAL_BODY_SEMANTICS, ANOMALY_TYPES, BOOKMARKLET + bookmarklet.js,
  OBSIDIAN_IMPORT_SPEC, UX_GUIDELINES_EN, IDEAS, BACKLOG, FRONTEND_FEATURES,
  UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS, NOTE_ERROR_CORRECTION_PLAN
- `docs/archive/` — ARCHITECTURE_ROADMAP (уже был помечен архивным), EXTERNAL_AUDIT_2026-09,
  AUTO_LINK_CREATION_PLAN (реализован LINKS-1), API_TEST_COVERAGE_PLAN,
  UI_MODERNIZATION_ROADMAP, MASS_IMPORT_TEST_PLAN, TEST_PLAN_VALIDATION_AUTOMATION,
  CRITICAL_FIXES, MANUAL_TEST_CHECKLISTS_RU (стейл), MANUAL_TEST_CHECKLIST_AI_AGENTS_3D_REFACTOR
  (ветка слита), YANDEX_DISK_BACKUP×2, CLOUD_BACKUP_SETUP, REGRESSION_TEST_PLAN_SUMMARY;
  `gordon/` → `archive/gordon/`, `3d-archive/` → `archive/3d/`

В корне `docs/` остаются: `README.md`, `AGENTS.md`, `AGENTS_EN.md` (тонкий указатель),
`AI_AGENT_PROTOCOL.md`, `AI_HANDOFF.md`, `AI_LOG.md`, `DECISIONS.md`,
`PROJECT_REVIEW_AI_AGENTS.md`, `LICENSES.md`.

Ничего не удалено: снятые файлы лежат в `archive/`, AGENTS_EN сведён в указатель
(старое содержимое в истории git).

## Как переписывались ссылки

Механически, скриптом [`scripts/devops/migrate-docs-reorg.mjs`](../../scripts/devops/migrate-docs-reorg.mjs)
(оставлен в репозитории как запись трансформации):

1. Markdown-ссылки `[t](href)` и reference-определения во всех `.md` и `.windsurfrules`:
   цель резолвится от старого положения файла, мапится через таблицу переездов, переписывается
   относительным путём от нового положения. Покрывает и смену глубины (`../X.md` →
   `../../X.md`), и переезд цели.
2. Литеральные строки `docs/<старое>` — во всех текстовых файлах: скрипты, workflow,
   промпты, скиллы, inline-код.

Итог: 327 ссылок + 317 литералов в 120 файлах. После миграции ни одного упоминания
старых путей не осталось (проверено `grep` по всем 19 переехавшим именам — единственное
вхождение старых путей — сама таблица в скрипте миграции).

## Отклонения от предложенной иерархии (на ревью)

1. **Протокольные файлы остались в корне `docs/`**, а не в `agents/`: AI_AGENT_PROTOCOL,
   AI_HANDOFF, AI_LOG, DECISIONS, PROJECT_REVIEW_AI_AGENTS. На них завязаны сторожа
   (`check-commit-authorship.mjs`, `check-decisions.mjs`, `check-board-*`) и список
   «читать перед работой» в CLAUDE.md — они и есть «входные».
2. **AGENTS.md остался на месте** — Windsurf грузит его как rule-файл с областью видимости
   на поддерево `docs/`; перенос в подпапку сузил бы область.
3. **TESTING_COMMANDS.md не слит в TESTING.md** — оба документа живые (команды из шпаргалки
   проверены: `run-bdd.cjs`, `test:realauth` и др. существуют). Размещены рядом в
   `operations/` с взаимными ссылками; редакционное слияние — отдельное решение.
4. **Чек-листы не слиты** в один двухуровневый: стейловый `MANUAL_TEST_CHECKLISTS_RU`
   ушёл в archive, живые COCKPIT и MINIMAL — в `operations/`.
5. **DEPLOYMENT_EN.md → `operations/`**, не в archive: это единственный документ про
   production/Kubernetes, на него ссылается корневой DEPLOY.md. Дата «апрель 2026»
   отмечена в указателе.
6. **`DEPLOY.md`, `DEPLOY.ru.md`, `TZ-Java-source-text-handler-2026-08-30.md` остались
   в корне репозитория** — за пределами скоупа разбора (он про `docs/`).
7. **NOTE_ERROR_CORRECTION_PLAN → `product/`**, не archive: статус «запланировано»,
   функционал не реализован — живой план.

## Синхронизированные производные

- `check-docs-links.mjs`: HISTORICAL переписан под новые пути (`docs/3d-archive` удалён —
  покрывается `docs/archive`; EXTERNAL_AUDIT и др. переехали).
- `docs/README.md`: ссылки переписаны скриптом, добавлена таблица структуры каталога,
  отметка «Сверено» обновлена.
- `docs/archive/README.md`: таблица пополнения DOC-REORG-1, дерево актуальной документации
  перерисовано под новую иерархию.
- `.windsurfrules`, `CLAUDE.md`, `.devin/prompts/*`, `.devin/skills/*`, `.claude/commands/*`,
  `run-full-test-cycle.*`, `.devin/config.json`, `frontend/svelte.config.js` — литеральные
  `docs/…` обновлены.
- `docs/tasks/README.md` регенерирован (140 строк, без дрифта).

## Верификация

| Проверка | Результат |
|---|---|
| `check-docs-links.mjs` | Docs OK — все локальные ссылки разрешаются |
| `check-tasks-index.mjs` | 140 записей, дрифта нет, ссылки доски целы |
| Остаточные упоминания старых путей | 0 вне таблицы миграции |
| Удалённые файлы | 0 — только `git mv` + AGENTS_EN → указатель |

## Что осталось за пределами (сознательно)

- Редакционные слияния содержимого (LINK_TYPES-кластер в один документ, DEPLOYMENT_* сверка
  с корневым DEPLOY.md) — документы размещены рядом и перекрёстно связаны; объединение
  текста — решение Claude Code.
- Бэклог-строка BACKUP-DIR-1: выравнивание дефолтов бэкапа под `~/Desktop/my items`
  (решение 52) — отдельная задача, после пуша.

## Послесловие (BACKUP-DIR-1 коммит)

При перепаковке коммита реорганизации (`29891a1` → `15809ec`, правка lint-staged)
переписанные ссылки в `docs/AI_LOG.md` затерлись — 8 исторических ссылок снова
указали на старые пути и упали на check-docs-links в следующем коммите.
Восстановлены (`agents/MANUAL_TEST_FEEDBACK.md`, `agents/AI_AGENT_SETUP.md`,
`archive/gordon/`). Урок: журнал переписывается скриптом как любой файл —
«историчность» освобождает его только от npm-проверки, ссылки обязаны резолвиться.
