# Backend Patterns — Knowledge Graph

Документ фиксирует архитектурные и проектные паттерны, реально используемые в backend-части проекта.

## 1. Общая архитектурная парадигма
- Clean Architecture / DDD: слои Domain → Application → Interfaces → Infrastructure.
- Domain содержит сущности, value objects, интерфейсы репозиториев и минимальную бизнес-логику (без зависимостей на инфраструктуру).
- Infrastructure реализует адаптеры: Postgres (GORM), MongoDB, Redis, очереди (asynq), внешние API (NLP/embedding).
- Interfaces предоставляет HTTP handlers (Gin), unified response helpers и middleware.

## 2. DDD-паттерны в коде
- Aggregates: `Note` и `Link` — отдельные агрегаты с собственными репозиториями.
- Value objects: `Title`, `Content`, `Metadata` — immutable объекты с валидацией в конструкторе (`NewTitle(...) (Title, error)`).
- Factory / Reconstruction:
  - `NewNote(...)` — создание нового агрегата.
  - `ReconstructNote(...)` / `ReconstructNoteWithCreator(...)` — восстановление из хранилища.
- Repository pattern:
  - Интерфейс `Repository` объявлен в `internal/domain/<aggregate>/repository.go`.
  - Реализация находится в `internal/infrastructure/db/postgres/*_repo.go`.
  - Поток зависимостей: handlers/services → domain interfaces → infrastructure implementations.

## 3. Application / Services
- `internal/services` содержит сервисы (use-cases) вроде `NoteService`:
  - Валидация входных параметров (pagination, query normalizing).
  - Делегирование на репозитории.
  - Служит фасадом для HTTP handlers и воркеров.

## 4. Infrastructure & интеграция
- Postgres с GORM — модели `*_model.go` + конвертация domain ↔ model (`toGormNote`, `toDomainNote`).
- Redis используется для кэширования (например, кеш списка заметок), token store и permission cache.
- MongoDB для черновиков/audit-логов (TTL, write-heavy).
- Очереди: `asynq` / Redis для фоновых задач (embedding, recompute recommendations).
- JWT: `internal/auth/JWTManager` для генерации/валидации пар токенов (access/refresh).
- Token storage: `RedisTokenStore` — хранение refresh tokens, blacklist.
- Circuit breaker: `sony/gobreaker` используется для защиты внешних сервисов (OpenAI, NLP).

## 5. Тесты и методики тестирования
- Unit tests: table-driven tests для value objects и domain-логики (`*_test.go`).
- Repo tests: unit + integration tests (`note_repo_test.go`, `note_repo_integration_test.go`).
- Mocks: интерфейсные моки (`mock_*`), тест-утилиты в `internal/testutil`.
- CI: запускаются unit и integration тесты; репозитории содержат примеры создания временных БД/транзакций.

## 6. GoF и структурные паттерны, применяемые в проекте
- Factory method: `New<Type>` / `Reconstruct<Type>` для создания сущностей и менеджеров.
- Repository: интерфейс + реализация (Adapter pattern).
- Singleton-ish: конфигурация и менеджеры (например, JWTManager создаётся в startup и передаётся дальше — контролируемая одна инстанция).
- Strategy / Policy: реализация fallback-стратегий (fallback to Redis, search fallback to ILIKE) и конфигурируемое поведение через `config`.
- Template / Hook: middleware для unified responses, error mapping.

## 7. SOLID-практики и проектные соглашения
- Single Responsibility: доменные сущности не зависят от инфраструктуры; репозитории — только про персистенс.
- Open/Closed: поведение расширяется через внедрение новых репозиториев/адаптеров, без изменений домена.
- Dependency Inversion: зависимости на интерфейсы домена, реализуемые инфраструктурой.

## 8. Error handling и API contract
- Все HTTP handlers используют `internal/interfaces/api/common` helpers (JSON, ErrorWithDetails).
- Domain errors переводятся в семантические HTTP-коды (NOT_FOUND, VALIDATION_ERROR, CONFLICT, INTERNAL_ERROR).

## 9. Operational patterns
- RLS (Row-Level Security) на уровне Postgres для multi-tenant безопасности — концепция отражена в документации и коде (tenant_id propagation).
- Backups: Postgres snapshots + WAL; MongoDB TTL для логов; Redis snapshotting.
- Monitoring: circuit breaker state, queue depth, TTL expiries

## 10. Рекомендации и будущие шаги
- Формализовать интерфейсы use-cases (`application/` слой) для упрощённого внедрения CQRS в будущем.
- Документировать контракты сообщений фоновых задач (payload schema для очередей).
- Добавить ADR для ключевых GoF-паттернов, если требуются строгие гарантии (например, fallback policy).