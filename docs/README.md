# Документация Knowledge Graph

Указатель по каталогу. Точка входа в проект — [README](../README.md), куда идёт проект —
[ROADMAP](../ROADMAP.md), что уже выпущено — [CHANGELOG](../CHANGELOG.md).

Пометка «архивный» означает, что документ описывает состояние на момент написания и не
поддерживается. Такие файлы сохраняются ради истории решений, но опираться на них нельзя.

## Архитектура

| Документ | О чём |
|---|---|
| [architecture/](architecture/README.md) | C4-модель, UML, 17 ADR — принятые архитектурные решения |
| [ARCHITECTURE_SUMMARY.md](ARCHITECTURE_SUMMARY.md) | Краткий срез системы, с него удобно начинать |
| [ARCHITECTURE_EN.md](ARCHITECTURE_EN.md) | Подробное описание слоёв и потоков |
| [FRONTEND_ARCHITECTURE_EN.md](FRONTEND_ARCHITECTURE_EN.md) | FSD и Atomic Design на фронтенде |
| [ARCHITECTURE_PATTERNS.md](ARCHITECTURE_PATTERNS.md) | Паттерны, применяемые в коде |
| [SaaS_DATABASE_SCHEMA.md](SaaS_DATABASE_SCHEMA.md) | Схема базы данных |
| [GRAPH_SERVICE_AUTH.md](GRAPH_SERVICE_AUTH.md) | Внутренняя авторизация graph-service |
| [GRAPH3D.md](GRAPH3D.md) | Устройство 3D-визуализации |
| [RECOMMENDATION_ARCHITECTURE.md](RECOMMENDATION_ARCHITECTURE.md) | Как устроены рекомендации |
| [RECOMMENDATION_API.md](RECOMMENDATION_API.md) | Контракт API рекомендаций |
| [RECOMMENDATION_TROUBLESHOOTING.md](RECOMMENDATION_TROUBLESHOOTING.md) | Разбор типовых сбоев рекомендаций |
| [API_ERRORS_EN.md](API_ERRORS_EN.md) | Формат ошибок API |
| [ARCHITECTURE_ROADMAP.md](ARCHITECTURE_ROADMAP.md) | **Архивный.** Фазы 1–6 от апреля 2026; их нумерация не совпадает с текущим планом, Phase 6 отложена |

## Эксплуатация

| Документ | О чём |
|---|---|
| [DEPLOYMENT_EN.md](DEPLOYMENT_EN.md) | Развёртывание self-hosted |
| [DOCKER.md](DOCKER.md) | Стеки, контейнеры, порты |
| [CONFIGURATION_EN.md](CONFIGURATION_EN.md) | Конфигурация системы, англоязычная версия |
| [CONFIGURATION_RU.md](CONFIGURATION_RU.md) | То же по-русски; версии разошлись по объёму, сверять по английской |
| [STACK_CONFIGURATION_COMPARISON.md](STACK_CONFIGURATION_COMPARISON.md) | Чем отличаются dev, personal и test стеки |
| [BACKUP.md](BACKUP.md) | Резервное копирование: что, куда и как восстанавливать |
| [CLOUD_BACKUP_SETUP.md](CLOUD_BACKUP_SETUP.md) | Настройка облачного бэкапа |
| [YANDEX_DISK_BACKUP.md](YANDEX_DISK_BACKUP.md) / [_EN](YANDEX_DISK_BACKUP_EN.md) | Бэкап на Яндекс.Диск |

## Тестирование

| Документ | О чём |
|---|---|
| [TESTING.md](TESTING.md) | Тестовая инфраструктура и текущее покрытие |
| [TESTING_COMMANDS.md](TESTING_COMMANDS.md) | Команды прогонов |
| [REGRESSION_TEST_PLAN.md](REGRESSION_TEST_PLAN.md) | Канонический регрессионный цикл |
| [ARGOS.md](ARGOS.md) | Визуальная регрессия через Argos |
| [API_TEST_COVERAGE_PLAN.md](API_TEST_COVERAGE_PLAN.md) | План покрытия API тестами |
| [MANUAL_TEST_FEEDBACK.md](MANUAL_TEST_FEEDBACK.md) | Журнал находок ручного тестирования — заполняется по ходу |
| [MANUAL_TEST_CHECKLIST_COCKPIT.md](MANUAL_TEST_CHECKLIST_COCKPIT.md) | Актуальный чек-лист ручной проверки |
| [MANUAL_TEST_CHECKLIST_MINIMAL.md](MANUAL_TEST_CHECKLIST_MINIMAL.md) | Короткий чек-лист |
| [MANUAL_TEST_CHECKLISTS_RU.md](MANUAL_TEST_CHECKLISTS_RU.md) | Полный набор чек-листов |
| [MANUAL_TEST_CHECKLIST_AI_AGENTS_3D_REFACTOR.md](MANUAL_TEST_CHECKLIST_AI_AGENTS_3D_REFACTOR.md) | **Архивный.** Чек-лист под конкретный рефакторинг 3D, июль 2026 |
| [REGRESSION_TEST_PLAN_SUMMARY.md](REGRESSION_TEST_PLAN_SUMMARY.md) | **Архивный.** Сокращённая копия плана регрессии |
| [MASS_IMPORT_TEST_PLAN.md](MASS_IMPORT_TEST_PLAN.md) | **Архивный.** План тестирования массового импорта, функция выпущена |
| [TEST_PLAN_VALIDATION_AUTOMATION.md](TEST_PLAN_VALIDATION_AUTOMATION.md) | **Архивный.** План автоматизации проверки валидации |

## Продукт и функции

| Документ | О чём |
|---|---|
| [BACKLOG.md](BACKLOG.md) | Детальные планы: что именно запланировано и от чего зависит |
| [IDEAS.md](IDEAS.md) | Гипотезы, не ставшие планами |
| [FRONTEND_FEATURES.md](FRONTEND_FEATURES.md) | Возможности интерфейса |
| [LINK_TYPES.md](LINK_TYPES.md) / [RU](LINK_TYPES_RU.md) | Типы связей и их веса |
| [LINKS_CHEATSHEET.md](LINKS_CHEATSHEET.md) | Шпаргалка по связям |
| [GRAPH_LINKS_VISUALIZATION.md](GRAPH_LINKS_VISUALIZATION.md) | Как связи отображаются на графе |
| [CELESTIAL_BODY_SEMANTICS.md](CELESTIAL_BODY_SEMANTICS.md) | Смысл типов небесных тел |
| [ANOMALY_TYPES.md](ANOMALY_TYPES.md) | Типы аномалий |
| [BOOKMARKLET.md](BOOKMARKLET.md) | Букмарклет для быстрого захвата |
| [OBSIDIAN_IMPORT_SPEC.md](OBSIDIAN_IMPORT_SPEC.md) | Спецификация импорта из Obsidian |
| [UX_GUIDELINES_EN.md](UX_GUIDELINES_EN.md) | Принципы UX проекта |
| [UI_MODERNIZATION_ROADMAP.md](UI_MODERNIZATION_ROADMAP.md) | План модернизации интерфейса |
| [AUTO_LINK_CREATION_PLAN.md](AUTO_LINK_CREATION_PLAN.md) | План автосоздания связей по рекомендациям |
| [NOTE_ERROR_CORRECTION_PLAN.md](NOTE_ERROR_CORRECTION_PLAN.md) | План исправления ошибок в заметках |
| [UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md](UI_DUPLICATION_AND_NOTE_CREATION_ANALYSIS.md) | Разбор дублирования в UI |
| [CRITICAL_FIXES.md](CRITICAL_FIXES.md) | **Архивный.** Отчёт о критических исправлениях, август 2026 |

## Работа агентов и аудиты

Рабочая кухня проекта: над кодом здесь работают человек и два AI-агента, и порядок передачи
работы описан явно.

| Документ | О чём |
|---|---|
| [AI_AGENT_PROTOCOL.md](AI_AGENT_PROTOCOL.md) | Кто ставит задачу, кто исполняет, кто проверяет |
| [AI_HANDOFF.md](AI_HANDOFF.md) | Доска: что за кем числится прямо сейчас |
| [AI_LOG.md](AI_LOG.md) | Журнал переходов задач |
| [tasks/](tasks/) | Постановки задач и замечания ревью |
| [AGENTS.md](AGENTS.md) / [AGENTS_EN.md](AGENTS_EN.md) | Описание агентов; версии расходятся, сведение запланировано |
| [PROJECT_REVIEW_AI_AGENTS.md](PROJECT_REVIEW_AI_AGENTS.md) | Долгая история состояния проекта |
| [AI_PROCESS_AUDIT.md](AI_PROCESS_AUDIT.md) | Аудит верификации и агентной обвязки, сентябрь 2026 |
| [EXTERNAL_AUDIT_2026-09.md](EXTERNAL_AUDIT_2026-09.md) | Внешний аудит кода и безопасности, 22 находки |

## Архив

[archive/](archive/) — отчёты и планы, отработавшие своё. [3d-archive/](3d-archive/) — снятый с
эксплуатации код 3D-визуализации.
