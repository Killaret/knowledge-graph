# DOC-AUDIT-1. Инспекция документации — находки и исправления

Запрошена владельцем 2026-09-21: «провести инспекцию документации, исправления
сразу, разночтения — на ревью». Исполнитель: Devin. Проверяет: Claude Code.

## Метод

Прочитаны шапки и ключевые разделы всех 60 документов верхнего уровня `docs/`,
корневых `.md`, подпапок `archive/`, `3d-archive/`, `gordon/`, `architecture/`,
`assets/`, `tasks/` (130 файлов — по индексу и шапкам). Сличение индекса
`docs/README.md` с фактическим списком файлов — машинное; содержательные
проверки — чтением и сверкой с кодом/решениями.

## Исправлено в этом заходе

| Находка | Что сделано |
|---|---|
| `YANDEX_DISK_BACKUP.md` описывал REST API как основной путь | баннер устаревания: решение 31 — синхронизируемая папка, API за `BACKUP_CLOUD_ENABLED` |
| `CLOUD_BACKUP_SETUP.md` ссылался на устаревший YANDEX-документ как на «актуальный» | указатель переведён на `BACKUP.md` |
| `AUTO_LINK_CREATION_PLAN.md` — статус «⏳ Запланировано» при реализованных gamma-связях | статус: частично реализовано LINKS-1, ссылки на постановку и находки |
| `MANUAL_TEST_CHECKLISTS_RU.md` отправлял на архивный 3D-чек-лист | указатель на актуальный `MANUAL_TEST_CHECKLIST_COCKPIT.md` |
| `LICENSES.md` отсутствовал в индексе | добавлен в `docs/README.md` |
| `gordon/` (11 файлов, Gordon AI, январь 2026) не был в индексе | помечен архивным в индексе с пояснением |
| `TZ-Java-source-text-handler` в корне — место неочевидное | пояснение в индексе: живое ТЗ, на него ссылаются IMP-4/NOTE-QUALITY-1 |
| `getJSONFloatOrDefault`: при существующем JSON отсутствующий ключ даёт 0/false вместо Go-дефолта | нюанс записан в `CONFIGURATION_EN.md` и `CONFIGURATION_RU.md` |
| `ARCHITECTURE_SUMMARY.md` не объяснял, откуда 0.6/степень 2 | добавлена ссылка на кривую автолинка W-1 |

## Найдено, но не трогал — на решение

1. **`TESTING_COMMANDS.md` vs `COMMANDS.md` vs разделы `TESTING.md`** — три
   пересекающихся списка команд. Кандидат на слияние или явное разделение
   ролей; решение владельца, куда смотреть первым.
2. **Три чек-листа ручного тестирования** (`_RU`, `COCKPIT`, `MINIMAL`) —
   пересечение по содержанию; RU-версия несёт устаревший «последнее
   обновление» про Allotropic-carbon. Свести или пометить роли.
3. **`AGENTS.md` / `AGENTS_EN.md`** — индекс сам отмечает «версии расходятся,
   сведение запланировано»; задача не заведена.
4. **`docs/gordon/`** — физически не в `archive/`, хотя по сути архив.
   Перемещение ломает ничего (ссылок нет), но это реструктуризация — на
   владельца/Claude.
5. **`docs/bookmarklet.js`** — исполняемый код внутри `docs/`; на него
   ссылается только `BOOKMARKLET.md`. Работает, но необычно.
6. **`NOTE_ERROR_CORRECTION_PLAN.md`, `UI_MODERNIZATION_ROADMAP.md`,
   `OBSIDIAN_IMPORT_SPEC.md`, `API_TEST_COVERAGE_PLAN.md`** — статусы
   «⏳ Запланировано» с июля; не сверено, какие части уже сделаны. Массовая
   сверка планов с кодом — отдельная задача.
7. **`docs/3d-archive/frontend`** — снятый с эксплуатации код рядом с доками;
   индекс помечает, физическое место ок.

## NLP-4 — процессная находка

Спека `docs/tasks/NLP-4-note-logical-form-normalization.md` написана Devin
(коммиты `6cc47d5`, `417fc03`, `af9ac3c`, `fd54376`) по решениям владельца.
По протоколу постановки ставит Claude Code — эта спека его формального ревью
как постановки не проходила. Реализацию начинать после того, как Claude
посмотрит спеку и подтвердит постановку (как RECO-1), либо по явному слову
владельца.

## Второй заход — предложение по объединению и иерархии

Дополнительно исправлено: `API_ERRORS_EN.md` начинался со stray-байта «да»
перед заголовком — убран.

### Кластеры дублирования

| Кластер | Файлы | Предложение |
|---|---|---|
| Команды и прогоны | `COMMANDS.md` (корень), `TESTING_COMMANDS.md`, разделы `TESTING.md`, `REGRESSION_TEST_PLAN.md` | `COMMANDS.md` — канонический список; `TESTING_COMMANDS.md` слить в `TESTING.md` и снять файл; в `REGRESSION_TEST_PLAN` оставить только порядок фаз |
| Чек-листы ручного тестирования | `MANUAL_TEST_CHECKLISTS_RU` (32 КБ), `MANUAL_TEST_CHECKLIST_COCKPIT`, `MANUAL_TEST_CHECKLIST_MINIMAL` | один документ с двумя уровнями (полный/короткий); RU-файл несёт устаревшее «последнее обновление» |
| Бэкап | `BACKUP.md`, `YANDEX_DISK_BACKUP.md`+`_EN`, `CLOUD_BACKUP_SETUP.md` | `BACKUP.md` — единственный актуальный; API-раздел Яндекса слить в него, оба файла → `archive/` |
| Архитектура верхнего уровня | `ARCHITECTURE_SUMMARY`, `ARCHITECTURE_EN`, `ARCHITECTURE_PATTERNS`, `ARCHITECTURE_ROADMAP`(архив), `FRONTEND_ARCHITECTURE_EN`, `SaaS_DATABASE_SCHEMA`, `GRAPH_SERVICE_AUTH`, `GRAPH3D`, `RECOMMENDATION_ARCHITECTURE` | роль у каждого есть, но они плоские — вынести в `docs/architecture/` рядом с C4/ADR |
| Деплой/стеки | `DEPLOY.md`+`.ru` (корень), `DEPLOYMENT_EN`, `DOCKER.md`, `STACK_CONFIGURATION_COMPARISON` | `DEPLOY*` — точка входа, `DOCKER`+`STACK_CONFIGURATION_COMPARISON` слить в один «стеки и порты»; `DEPLOYMENT_EN` (апрель, «Production Ready») проверить на стейл или в архив |
| Агенты | `AGENTS.md`/`AGENTS_EN`, `CLAUDE.md`, `AI_AGENT_SETUP`, `AI_AGENT_PROTOCOL` | `AGENTS_EN` дублирует `AGENTS` и уже разошёлся — оставить RU каноническим (по правилу языка) или сделать EN тонким указателем |
| Планы с июльским статусом | `AUTO_LINK_CREATION_PLAN` (исправлен), `NOTE_ERROR_CORRECTION_PLAN`, `UI_MODERNIZATION_ROADMAP`, `API_TEST_COVERAGE_PLAN`, `OBSIDIAN_IMPORT_SPEC` | сверить с кодом: сделанное → archive, живое → в BACKLOG |
| Типы связей | `LINK_TYPES`+`_RU`, `LINKS_CHEATSHEET`, `GRAPH_LINKS_VISUALIZATION` | один документ «связи»: типы + отображение + `source_type` (user/gamma) — сейчас gamma в трёх местах не упомянут явно |
| `docs/gordon/` | 11 файлов, январь 2026 | переместить в `archive/gordon/` — физически туда, где уже лежит архив |
| `docs/3d-archive/` | снятый код | слить в `archive/3d/` — одна точка архива вместо двух |

### Предлагаемая иерархия

```
docs/
  README.md                  — индекс (уже есть, машинно проверяется)
  architecture/              — C4/ADR/uml + SUMMARY, _EN, PATTERNS, FRONTEND,
                               SaaS_SCHEMA, GRAPH_SERVICE_AUTH, GRAPH3D, RECOMMENDATION_*
  api/                       — API_EN, API_ERRORS_EN, RECOMMENDATION_API
  operations/                — DOCKER+стеки, DEPLOYMENT, CONFIGURATION_*, BACKUP,
                               TESTING, REGRESSION, ARGOS
  product/                   — LINK_TYPES(+cheatsheet+viz), CELESTIAL, ANOMALY,
                               BOOKMARKLET, OBSIDIAN, UX, IDEAS, BACKLOG, FEATURES
  agents/                    — AI_*, AGENTS, DECISIONS, PROJECT_REVIEW, аудиты,
                               MANUAL_TEST_FEEDBACK
  tasks/                     — как есть (130 постановок/ревью)
  archive/                   — archive/ + gordon/ + 3d-archive/ + снятые планы
```

Плоские файлы верхнего уровня останутся только входные: `README`, `AGENTS`,
`AI_HANDOFF`, `AI_LOG`, `DECISIONS` — то, что читается каждую сессию.

### Что это ломает и почему решение не моё

- Перемещения рвут ~все относительные ссылки — `check-docs-links.mjs` и
  индекс задач ловят это, но правка ссылок объёмная;
- 130 файлов `docs/tasks/` содержат ссылки `../FILE.md` на верхний уровень —
  переезд в подпапки меняет глубину относительных путей;
- соглашение об именах (UPPER_CASE vs kebab-case в `gordon/`, `architecture/`)
  и языковые пары (EN может отставать — зафиксировано самим индексом) —
  нормативный вопрос для `.windsurfrules`.

Рекомендация: вынести как отдельную задачу DOC-REORG-1 — спека от Claude Code,
исполнение Devin, прогон `check-docs-links` + регенерация индексов после
каждого перемещения.
