# 012 — Key Architecture & Design Patterns

Date: 2026-05-22

Status: Accepted

## Context
The Knowledge Graph project spans frontend (Svelte 5) and backend (Go). Multiple architectural and design patterns are applied across the codebase to ensure maintainability, testability and operational resilience. A compact, agreed ADR is required to keep decisions discoverable for new contributors and for future reviews.

## Decision
We formally adopt and document the following key patterns and practices as canonical for the project:

- Domain-Driven Design (DDD) in backend: aggregates, value objects, repository interfaces in `internal/domain`, infrastructure adapters in `internal/infrastructure`.
- Clean Architecture / Layered separation: Domain → Application (services/use-cases) → Interfaces (HTTP handlers) → Infrastructure (DB, cache, queues).
- Repository pattern with explicit interfaces in domain and GORM implementations in `internal/infrastructure/db/postgres`.
- Factory / Reconstruction constructors: `New<Type>` and `Reconstruct<Type>` for entity creation and rehydration.
- Singleton / Manager pattern for long-lived clients and services (JWTManager, PreloadService), initialized in startup and injected via DI.
- Strategy/Fallback policies for external dependencies and search (FTS → ILIKE fallback) and recommendation pipelines.
- Event / Worker model via Redis/asynq for side-effects and background processing; messages must include explicit schema version.
- Testing patterns: table-driven unit tests for domain logic, integration tests for repositories, MSW and jsdom mocks for frontend unit tests, Playwright for E2E.
- Frontend patterns: component composition, centralized config (`knowledge-graph.config.json`), `PreloadService` singleton, stores for global state, and test mocks for browser APIs.

## Consequences
- Code must avoid importing infrastructure packages from domain; any cross-layer call should be via interfaces.
- New persistence adapters must implement existing repository interfaces; migration should be incremental and tested with integration suite.
- Background task payloads must be versioned and validated to enable rolling upgrades.
- Frontend components should keep business logic in `src/lib/services/*` and stores; UI files (`.svelte`) contain primarily rendering logic.
- Test stability is a priority — introduce mocks for browser APIs and provide fixtures for repository integration tests.

## Implementation Notes & References
- Backend examples:
  - Aggregates and value objects: `backend/internal/domain/note/entity.go`, `backend/internal/domain/note/value_objects.go`
  - Repository interface: `backend/internal/domain/note/repository.go`
  - Postgres implementation: `backend/internal/infrastructure/db/postgres/note_repo.go`
  - Services/use-cases: `backend/internal/services/note_service.go`

- Frontend examples:
  - Patterns doc: `frontend/FRONTEND_PATTERNS.md`
  - Preload service: `frontend/src/lib/services/PreloadService.ts`
  - Stores: `frontend/src/lib/stores`
  - Test setup: `frontend/vitest-setup.ts`, `frontend/vitest.config.ts`, `frontend/playwright.config.ts`

## Next Steps
- Keep ADR updated when introducing major alternative patterns (e.g., switching to CQRS, adding a new message queue).
- For each pattern change, create an ADR referencing this document and the specific rationale.

Authored-by: automation (dev-assistant)
