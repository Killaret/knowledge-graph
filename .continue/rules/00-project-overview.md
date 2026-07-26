---
name: Project Overview
alwaysApply: true
description: Knowledge Graph project overview, tech stack, architecture, language policy, and key commands
---

# Knowledge Graph — Project Overview

## What is this project?

Knowledge Graph is a note management system with graph-based relationships,
NLP-powered semantic analysis (embeddings, keyword extraction), and interactive
2D/3D graph visualization. Notes are connected via manual links and automatic
similarity-based recommendations.

## Tech Stack

- **Backend:** Go 1.25, Gin, GORM, PostgreSQL (pgvector), Redis (go-redis/v9), MongoDB, asynq, JWT
- **Frontend:** Svelte 5 (runes), TypeScript strict, SvelteKit, D3-force, Three.js
- **NLP Service:** Python FastAPI, sentence-transformers, HuggingFace (offline-first)
- **Infrastructure:** Docker multi-stage, docker-compose, nginx gateway

## Architecture — Clean Architecture + DDD

```
backend/internal/
├── domain/          # Entities, Value Objects, Repository interfaces (NO external deps)
├── application/     # Use cases — orchestrate domain logic
├── infrastructure/  # Implementations: PostgreSQL repos, Redis, NLP client, queue
└── interfaces/      # HTTP handlers (Gin), middleware, request/response DTOs
```

**Rules:**
- Dependencies flow INWARD: interfaces → application → domain ← infrastructure
- Domain layer has ZERO external imports (no gorm, no gin, no redis)
- Repositories return domain entities, never DB models
- No global variables — explicit dependency injection everywhere

## Language Policy

**User-facing runtime content uses Russian by default (`ru` locale), with English (`en`) support through i18n keys:**

- UI strings, buttons, labels, placeholders, errors
- Toast messages, error messages, tooltips
- Note titles and content

**MUST be in English:**
- Code identifiers and variable names
- Commit messages
- API error codes
- Public code comments, README, authoritative docs

**User content (note titles/bodies) may be in any language.**
**Exceptions:** Internal code comments — any language OK.

```typescript
// ✅ Good — Russian UI string from galactic-lexicon / i18n
toast.success(formatMessage("note.created", "ru", { title: note.title }));
placeholder="Поиск по графу знаний...";

// ✅ Good — English variant via same i18n key
const en = formatMessage("note.created", "en", { title: note.title });

// ❌ Bad — hardcoded string without i18n key
toast.success("Note created successfully");
```

## Key Commands

```bash
# Backend
cd backend && go test ./...          # Unit tests
cd backend && go build ./cmd/server  # Build

# Frontend
cd frontend && npm run test:unit     # Vitest
cd frontend && npm run build         # Production build

# Docker
docker compose up -d                 # Start stack
docker compose -f docker-compose.personal.yml up -d  # Personal
```

## Testing Requirements

- Coverage MUST be >60% for all modules
- Go: table-driven tests with `testify` (assert/require)
- Frontend: Vitest + @testing-library/svelte
- E2E: Playwright
- NLP: pytest
- Always run tests before committing

## Agents (11)

1. **knowledge-graph-orchestrator** — task routing
2. **knowledge-graph-backend-go** — Go API, PostgreSQL, Redis, MongoDB, JWT
3. **knowledge-graph-frontend-svelte** — Svelte 5, TypeScript, UI/UX
4. **knowledge-graph-integration** — OpenAPI, DTOs, contracts
5. **knowledge-graph-infrastructure** — Docker, monitoring, backups
6. **knowledge-graph-devops** — CI/CD, deployment
7. **knowledge-graph-performance** — profiling, caching, P95
8. **knowledge-graph-security** — audit, Auth/AuthZ, encryption
9. **knowledge-graph-testing** — unit/integration/E2E tests
10. **knowledge-graph-nlp** — Python FastAPI, NLP, embeddings
11. **knowledge-graph-data** — DB migrations, GORM, pgvector, schemas

## Critical Invariants

1. No global variables — dependency injection only
2. Repositories return domain entities (never `*NoteModel`)
3. Svelte 5 runes syntax only ($state, $derived, $effect, $props)
4. Never commit secrets (.env, tokens)
5. Russian UI via i18n; English for code/API/commits/public docs
6. Health checks on all Docker services
7. Context (`ctx context.Context`) passed through all Go methods
