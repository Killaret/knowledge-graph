# Architecture Patterns — Knowledge Graph (Summary)

Краткое и точное собрание архитектурных паттернов и принципов, используемых в проекте (frontend + backend).

## 1. Ключевые парадигмы
- Domain-Driven Design (DDD) — разделение на доменные сущности, value-objects, репозитории.
- Clean Architecture / Layered: Domain → Application → Interfaces → Infrastructure.
- SOLID — применяется при проектировании сервисов, менеджеров и репозиториев.

## 2. Часто используемые паттерны
- Factory Method: `New<Type>` / `Reconstruct<Type>` для создания и восстановления агрегатов.
- Repository (Adapter): интерфейсы в домене + реализации в infrastructure (Postgres, MongoDB).
- Singleton/Manager: конфиги и менеджеры (JWTManager, clients) — инициализируются в startup и передаются через DI.
- Strategy / Policy: fallback-стратегии (search fallback, recommendation fallback), конфигурируемые через `config`.
- Observer/Event: очередь/worker модель (asynq) для асинхронных задач и side-effects.
- Template Method / Middleware: unified response middleware, auth middleware, error mapping.

## 3. Технические практики
- Безопасность: RLS, JWT с refresh/blacklist, role/permissions в claims.
- Тестируемость: table-driven tests, interface mocks, integration tests for repo layer.
- Resilience: circuit breakers (sony/gobreaker) и retry policies.
- Observability: мониторинг очередей, CB, и метрик производительности.

## 4. Соглашения по кодовой базе
- Именования конструктора: `NewX` и `ReconstructX` для восстановления из DB.
- Mocks: `mock_<interface>.go` и helpers в `internal/testutil`.
- Handlers: `internal/interfaces/api/*` используют `common` helpers для унифицированных ответов.
- Конфигурация: централизованная `knowledge-graph.config.json` + `config/*.json`.

## 5. Грубый набор правил для архитекторов и разработчиков
- Доменные сущности не должны импортировать инфраструктурные пакеты.
- Вся персистенция через доменные интерфейсы.
- Компоненты должны быть проверяемы и мокируемы (избегать глобальных состояний).
- Любое расширение функциональности через новые адаптеры/репозитории должно проходить через интерфейс домена.

## 6. Как пользовались этими паттернами в проекте
- Backend: value objects, repository pattern, services для use-cases, Postgres/Mongo/Redis adapters, workers.
- Frontend: component-based, centralized config, preload services, test mocks for browser APIs.

## 7. Следующие шаги
- Сформировать набор ADR (если нужно) для ключевых паттернов.
- Синхронизировать `FRONTEND_PATTERNS.md` и `BACKEND_PATTERNS.md` в единый `ARCHITECTURE_PATTERNS.md` (этот файл).
- Подготовить компактный `ARCHITECTURE_GUIDELINES.md` для новых участников команды.