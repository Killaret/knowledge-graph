# Copilot Instructions for Knowledge Graph

🎯 **Quick Start:** Read `.koda/compact-rules.md` first (116 lines vs 11000+ lines of documentation)

---

## Overview
**Knowledge Graph** is a domain-driven design (DDD) backend system written in Go with PostgreSQL + Redis, implementing a note-taking system with graph-based relationships and semantic recommendations using embeddings (pgvector).

---

## Build, Test, and Lint

### Backend (Go)

#### Build
```bash
# Build binary
cd backend
go build -o server ./cmd/server

# Run server
./server

# Docker build
docker-compose up --build
```

#### Testing
```bash
# Unit tests only
cd backend
go test ./...

# Unit tests with verbose output
go test -v ./...

# Single test
go test -v -run TestCreateNote ./...

# Integration tests (requires running PostgreSQL)
go test -tags=integration ./...

# Coverage report
go test -cover ./... -coverprofile=coverage.out
```

**Test Files**: Colocated with source code (e.g., `entity_test.go`, `*_handler_test.go`). Tests use table-driven patterns and in-memory mocks.

#### Environment Setup
Create `.env` in project root (or use `.env` in backend/):
```
DATABASE_URL=postgresql://kb_user:kb_password@localhost:5432/knowledge_base?sslmode=disable
REDIS_URL=localhost:6379
NLP_SERVICE_URL=http://localhost:5000
JWT_SECRET=change-me-in-production
```

### Frontend (Svelte 5 + Vitest)

#### Build
```bash
cd frontend
npm run build
```

#### Testing
```bash
# Unit tests
npm run test:unit

# E2E tests (Playwright)
npm run test
```

---

## Architecture

### Backend Structure
```
backend/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── domain/                 # Business entities, value objects
│   ├── application/            # Use cases, services
│   ├── infrastructure/         # DB, Redis, config
│   └── interfaces/api/         # HTTP handlers
```

### Key Patterns
- **Clean Architecture**: Layers are isolated with interfaces
- **Domain-Driven Design**: Entities, Value Objects, Aggregates
- **Repository Pattern**: Access to data through interfaces
- **CQRS-Lite**: Separation of commands and queries

### Frontend Structure
```
frontend/src/
├── lib/                        # Business logic, API clients
├── components/                 # UI components
└── routes/                     # SvelteKit pages
```

### Key Patterns
- **Atomic Design**: Atoms → Molecules → Organisms
- **Svelte 5**: Runes for reactivity
- **TypeScript**: Strict typing throughout
- **Stores**: State management with Svelte stores

---

## AI Guidelines

### For Backend Tasks
- Follow Clean Architecture principles
- Use Repository Pattern for data access
- Keep domain logic pure (no external dependencies)
- Add unit tests for all business logic

### For Frontend Tasks
- Component-based architecture
- Use TypeScript for type safety
- Test with Vitest (unit) and Playwright (E2E)
- Follow atomic design patterns

### For Infrastructure Tasks
- Use Docker multi-stage builds
- Health checks for all services
- Graceful shutdown handling
- Volume mounts for persistence

---

## Common Commands

### Backend
```bash
cd backend
go test ./...                    # Run all tests
go build ./cmd/server            # Build server
go run ./cmd/server              # Run server
```

### Frontend
```bash
cd frontend
npm run test:unit                # Unit tests
npm run test                     # E2E tests
npm run build                    # Build
```

### Docker
```bash
docker compose up -d              # Start services
docker compose logs -f backend    # View logs
docker compose down              # Stop services
```

---

## Detailed Documentation
- Full AI rules: `.koda/compact-rules.md` (recommended) or `.cursor/rules/*.md`
- AGENTS documentation: `docs/AGENTS.md`
- Configuration: `docs/CONFIGURATION.md`
- Testing: `docs/TESTING.md`