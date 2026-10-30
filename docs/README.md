# Документация Knowledge Graph

Указатель по каталогу. Точка входа в проект — [README](../README.md), куда идёт проект —
[ROADMAP](../ROADMAP.md), что уже выпущено — [CHANGELOG](../CHANGELOG.md).

Пометка «архивный» означает, что документ описывает состояние на момент написания и не
поддерживается. Такие файлы сохраняются ради истории решений, но опираться на них нельзя.

## Структура каталога

С 2026-09-22 (DOC-REORG-1) документы разложены по подкаталогам. В корне `docs/` остаются
только входные и протокольные файлы: этот указатель, `AGENTS.md`, протокол, доска, журнал,
решения, обзор проекта и `LICENSES.md`.

| Каталог | Что лежит |
|---|---|
| [`architecture/`](architecture/README.md) | Архитектура, C4, UML, ADR, паттерны, схемы данных |
| [`api/`](api) | Контракты API и формат ошибок |
| [`operations/`](operations) | Docker, развёртывание, конфигурация, бэкапы, тестирование, регрессия |
| [`product/`](product) | Типы связей, визуализация, фичи, UX, бэклог продукта |
| [`agents/`](agents) | Обвязка AI-агентов: установка, аудиты процесса, журналы ручных проверок |
| [`tasks/`](tasks) | Постановки задач и разборы ревью |
| [`archive/`](archive) | Снятые с эксплуатации документы, включая `archive/3d/` и `archive/gordon/` |

## Состояние документации

**Сверено 2026-09-22.** Все документы перечислены в этом указателе, все локальные ссылки разрешаются, все упомянутые команды `npm run` существуют. Проверено машинно — фаза `Documentation links` в [`check-all`](../COMMANDS.md) и в CI, скрипт [`check-docs-links.mjs`](../scripts/testing/check-docs-links.mjs).

**Что эта отметка не означает.** Проверка следит за ссылками и командами, а не за смыслом: она не подтверждает, что описанное в тексте поведение совпадает с кодом. Фактические утверждения — числа, эндпоинты, имена моделей — сверяются при правке соответствующей задачи, как того требует раздел Documentation в [`.windsurfrules`](../.windsurfrules).

**Исторические документы под проверку не подпадают** и не должны: журнал, доска, датированные аудиты и `docs/tasks/` описывают состояние на момент написания. Их правка задним числом стёрла бы историю решений. Список исключений — в шапке скрипта.

## Архитектура

| Документ | О чём |
|---|---|
| [architecture/](architecture/README.md) | C4-модель, UML, 18 ADR — принятые архитектурные решения |
| [ARCHITECTURE_SUMMARY.md](architecture/ARCHITECTURE_SUMMARY.md) | Краткий срез системы, с него удобно начинать |
| [ARCHITECTURE_EN.md](architecture/ARCHITECTURE_EN.md) | Подробное описание слоёв и потоков |
| [FRONTEND_ARCHITECTURE_EN.md](architecture/FRONTEND_ARCHITECTURE_EN.md) | FSD и Atomic Design на фронтенде |
| [ARCHITECTURE_PATTERNS.md](architecture/ARCHITECTURE_PATTERNS.md) | Паттерны, применяемые в коде |
| [SaaS_DATABASE_SCHEMA.md](architecture/SaaS_DATABASE_SCHEMA.md) | Схема базы данных |
| [GRAPH_SERVICE_AUTH.md](architecture/GRAPH_SERVICE_AUTH.md) | Внутренняя авторизация graph-service |
| [GRAPH3D.md](architecture/GRAPH3D.md) | Устройство 3D-визуализации |
| [RECOMMENDATION_ARCHITECTURE.md](architecture/RECOMMENDATION_ARCHITECTURE.md) | Как устроены рекомендации |
| [RECOMMENDATION_API.md](api/RECOMMENDATION_API.md) | Контракт API рекомендаций |
| [RECOMMENDATION_TROUBLESHOOTING.md](architecture/RECOMMENDATION_TROUBLESHOOTING.md) | Разбор типовых сбоев рекомендаций |
| [API_ERRORS_EN.md](api/API_ERRORS_EN.md) | Формат ошибок API |
| [API_EN.md](api/API_EN.md) | Контракт API: Swagger UI, openAPI.yaml, импорт в Postman, авторизация для проб |
| [ARCHITECTURE_ROADMAP.md](archive/ARCHITECTURE_ROADMAP.md) | **Архивный.** Фазы 1–6 от апреля 2026; их нумерация не совпадает с текущим планом, Phase 6 отложена |

## Эксплуатация

| Документ | О чём |
|---|---|
| [DEPLOYMENT_EN.md](operations/DEPLOYMENT_EN.md) | Развёртывание self-hosted |
| [DOCKER.md](operations/DOCKER.md) | Стеки, контейнеры, порты |
| [CONFIGURATION_EN.md](operations/CONFIGURATION_EN.md) | Конфигурация системы, англоязычная версия |
| [CONFIGURATION_RU.md](operations/CONFIGURATION_RU.md) | То же по-русски; версии разошлись по объёму, сверять по английской |
| [STACK_CONFIGURATION_COMPARISON.md](operations/STACK_CONFIGURATION_COMPARISON.md) | Чем отличаются dev, personal и test стеки |
| [BACKUP.md](operations/BACKUP.md) | Резервное копирование: что, куда и как восстанавливать |
| [CLOUD_BACKUP_SETUP.md](archive/CLOUD_BACKUP_SETUP.md) | **Архивный.** Настройка Cloudflare R2 — устаревший путь |
| [YANDEX_DISK_BACKUP.md](archive/YANDEX_DISK_BACKUP.md) / [_EN](archive/YANDEX_DISK_BACKUP_EN.md) | **Частично устаревший.** REST API Яндекс.Диска — опциональный путь за `BACKUP_CLOUD_ENABLED`; основной — синхронизируемая папка (решение 31, см. BACKUP.md) |
| [LICENSES.md](LICENSES.md) | Инвентарь лицензий зависимостей и сторож `check-licenses.mjs` |

## Тестирование

| Документ | О чём |
|---|---|
| [TESTING.md](operations/TESTING.md) | Тестовая инфраструктура и текущее покрытие |
| [TESTING_COMMANDS.md](operations/TESTING_COMMANDS.md) | Команды прогонов |
| [REGRESSION_TEST_PLAN.md](operations/REGRESSION_TEST_PLAN.md) | Канонический регрессионный цикл |
| [ARGOS.md](operations/ARGOS.md) | Визуальная регрессия через Argos |
| [API_TEST_COVERAGE_PLAN.md](archive/API_TEST_COVERAGE_PLAN.md) | План покрытия API тестами |
| [MANUAL_TEST_FEEDBACK.md](agents/MANUAL_TEST_FEEDBACK.md) | Журнал находок ручного тестирования — заполняется по ходу |
| [MANUAL_TEST_CHECKLIST_COCKPIT.md](operations/MANUAL_TEST_CHECKLIST_COCKPIT.md) | Актуальный чек-лист ручной проверки |
| [MANUAL_TEST_CHECKLIST_MINIMAL.md](operations/MANUAL_TEST_CHECKLIST_MINIMAL.md) | Короткий чек-лист |
| [MANUAL_TEST_CHECKLISTS_RU.md](archive/MANUAL_TEST_CHECKLISTS_RU.md) | Полный набор чек-листов |
| [MANUAL_TEST_CHECKLIST_AI_AGENTS_3D_REFACTOR.md](archive/MANUAL_TEST_CHECKLIST_AI_AGENTS_3D_REFACTOR.md) | **Архивный.** Чек-лист под конкретный рефакторинг 3D, июль 2026 |
| [REGRESSION_TEST_PLAN_SUMMARY.md](archive/REGRESSION_TEST_PLAN_SUMMARY.md) | **Архивный.** Сокращённая копия плана регрессии |
| [MASS_IMPORT_TEST_PLAN.md](archive/MASS_IMPORT_TEST_PLAN.md) | **Архивный.** План тестирования массового импорта, функция выпущена |
| [TEST_PLAN_VALIDATION_AUTOMATION.md](archive/TEST_PLAN_VALIDATION_AUTOMATION.md) | **Архивный.** План автоматизации проверки валидации |

## Продукт и функции

| Документ | О чём |
|---|---|
| [BACKLOG.md](product/BACKLOG.md) | Детальные планы: что именно запланировано и от чего зависит |
| [IDEAS.md](product/IDEAS.md) | Гипотезы, не ставшие планами |
| [FRONTEND_FEATURES.md](product/FRONTEND_FEATURES.md) | Возможности интерфейса |
| [LINK_TYPES.md](product/LINK_TYPES.md) / [RU](product/LINK_TYPES_RU.md) | Типы связей и их веса |
| [LINKS_CHEATSHEET.md](product/LINKS_CHEATSHEET.md) | Шпаргалка по связям |
| [GRAPH_LINKS_VISUALIZATION.md](product/GRAPH_LINKS_VISUALIZATION.md) | Как связи отображаются на графе |
| [CELESTIAL_BODY_SEMANTICS.md](product/CELESTIAL_BODY_SEMANTICS.md) | Смысл типов небесных тел |
| [ANOMALY_TYPES.md](product/ANOMALY_TYPES.md) | Типы аномалий |
| [BOOKMARKLET.md](product/BOOKMARKLET.md) | Букмарклет для быстрого захвата |
| [OBSIDIAN_IMPORT_SPEC.md](product/OBSIDIAN_IMPORT_SPEC.md) | Спецификация импорта из Obsidian |
| [UX_GUIDELINES_EN.md](product/UX_GUIDELINES_EN.md) | Принципы UX проекта |
| [UI_MODERNIZATION_ROADMAP.md](archive/UI_MODERNIZATION_ROADMAP.md) | План модернизации интерфейса |
| [AUTO_LINK_CREATION_PLAN.md](archive/AUTO_LINK_CREATION_PLAN.md) | План автосоздания связей по рекомендациям |
| [NOTE_ERROR_CORRECTION_PLAN.md](product/NOTE_ERROR_CORRECTION_PLAN.md) | План исправления ошибок в заметках |
| [UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md](product/UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md) | Разбор дублирования в UI |
| [CRITICAL_FIXES.md](archive/CRITICAL_FIXES.md) | **Архивный.** Отчёт о критических исправлениях, август 2026 |

## Работа агентов и аудиты

Рабочая кухня проекта: над кодом здесь работают человек и два AI-агента, и порядок передачи
работы описан явно.

| Документ | О чём |
|---|---|
| [AI_AGENT_PROTOCOL.md](AI_AGENT_PROTOCOL.md) | Кто ставит задачу, кто исполняет, кто проверяет |
| [DECISIONS.md](DECISIONS.md) | Решения владельца: что решили, когда и где разбор |
| [AI_AGENT_SETUP.md](agents/AI_AGENT_SETUP.md) | Из чего состоит обвязка агентов и как развернуть её на другой машине |
| [AI_HANDOFF.md](AI_HANDOFF.md) | Доска: что за кем числится прямо сейчас |
| [AI_LOG.md](AI_LOG.md) | Журнал переходов задач |
| [tasks/](tasks) | Постановки задач и замечания ревью |
| [AGENTS.md](AGENTS.md) / [AGENTS_EN.md](AGENTS_EN.md) | Описание агентов; AGENTS_EN — тонкий указатель на AGENTS.md (сведено при DOC-REORG-1) |
| [PROJECT_REVIEW_AI_AGENTS.md](PROJECT_REVIEW_AI_AGENTS.md) | Долгая история состояния проекта |
| [AI_PROCESS_AUDIT.md](agents/AI_PROCESS_AUDIT.md) | Аудит верификации и агентной обвязки, сентябрь 2026 |
| [EXTERNAL_AUDIT_2026-09.md](archive/EXTERNAL_AUDIT_2026-09.md) | Внешний аудит кода и безопасности, 22 находки |

## Архив

[archive/](archive) — отчёты и планы, отработавшие своё. [3d-archive/](archive/3d) — снятый с
эксплуатации код 3D-визуализации. [gordon/](archive/gordon) — **архивный** пакет разборов и гайдов по
деплою от Gordon (Docker AI Assistant), январь 2026; Kubernetes-ориентирован, к текущему
self-hosted пути отношения не имеет.

Корневой `TZ-Java-source-text-handler-2026-08-30.md` — согласованное ТЗ внешнего Java-сервиса
(нет в этом репозитории), лежит в корне потому, что на него ссылаются постановки IMP-4 и
NOTE-QUALITY-1.
