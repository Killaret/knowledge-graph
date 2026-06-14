# AI Agents in Knowledge Graph

This document describes the AI agent roles used in the repository to support work with frontend, backend, testing, integration, infrastructure, and documentation.

## Agent List

### `knowledge-graph-orchestrator`
- **Focus:** Main delegation logic for Cursor AI
- **Ideal for:** Routing tasks to appropriate agents, analyzing user requests

**Key Responsibilities:**
- Always active on startup
- Analyzes user requests
- Delegates tasks to appropriate agents (backend, frontend, integration, infrastructure, devops, performance, security, testing)
- Does not execute work without coordination

**Example Prompts:**
- "Choose the best agent for this task"
- "Delegate task to backend"
- "Collect security report"
- "Determine required tests"

---

### `knowledge-graph-backend-go`
- **Focus:** `backend/`, Go 1.23+, API, PostgreSQL, Redis, RabbitMQ, JWT authorization
- **Ideal for:** Backend development, server logic, database optimization, API endpoints

**Key Responsibilities:**
- Go API and server logic
- PostgreSQL, Redis, RabbitMQ integration
- Authorization and JWT management
- Build, testing, and backend documentation
- Backend performance optimization

**Example Prompts:**
- "Write a new endpoint"
- "Add integration test for API"
- "Optimize database query"
- "Implement authentication flow"

---

### `knowledge-graph-frontend-svelte`
- **Focus:** `frontend/`, Svelte 5, TypeScript, UI/UX, components, state management
- **Ideal for:** Frontend development, UI components, forms, routing, validation

**Key Responsibilities:**
- Svelte 5 components and TypeScript
- UI/UX, state management, routes, and forms
- API integration, validation, and tests
- Bundle optimization and loading performance

**Example Prompts:**
- "Create a component for notes list"
- "Add note creation form"
- "Optimize page rendering"
- "Implement responsive UI"

---

### `knowledge-graph-integration`
- **Focus:** API mapping, OpenAPI, webhooks, OAuth, integration contracts
- **Ideal for:** API contracts, DTO generation, protocol specifications

**Key Responsibilities:**
- API mapping and OpenAPI specification
- Protocols, webhooks, OAuth, and integration contracts
- DTO generation and specification compliance

**Example Prompts:**
- "Create OpenAPI schema for new endpoint"
- "Write webhook receiver"
- "Check API contracts compliance"
- "Generate DTO for new feature"

---

### `knowledge-graph-infrastructure`
- **Focus:** Docker, containerization, orchestration, monitoring, backups, scaling
- **Ideal for:** Infrastructure setup, container configuration, monitoring, backups

**Key Responsibilities:**
- Docker, containerization, and orchestration
- Monitoring, backups, and scaling
- Environment configuration and availability

**Example Prompts:**
- "Describe Docker Compose for service"
- "Add monitoring and backup"
- "Optimize infrastructure for fault tolerance"
- "Configure environment variables"

---

### `knowledge-graph-devops`
- **Focus:** CI/CD, deployment, logs, metrics, errors, rollbacks, release automation
- **Ideal for:** Pipeline setup, deployment automation, release management

**Key Responsibilities:**
- CI/CD and deployment
- Logs, metrics, errors, and rollbacks
- Build and release automation

**Example Prompts:**
- "Set up pipeline for build"
- "Check logs and deployment status"
- "Optimize release process"
- "Configure CI/CD workflow"

---

### `knowledge-graph-performance`
- **Focus:** Profiling, load testing, optimization, P95, response time, caching
- **Ideal for:** Performance analysis, optimization, caching strategies

**Key Responsibilities:**
- Profiling, load testing, and optimization
- P95, response time, CPU and memory usage
- Caching and scaling

**Example Prompts:**
- "Analyze performance bottlenecks"
- "Suggest query optimization"
- "Reduce API response time"
- "Configure caching strategy"

---

### `knowledge-graph-security`
- **Focus:** Security audit, vulnerability scanning, Auth/AuthZ, encryption, compliance
- **Ideal for:** Security reviews, authentication, authorization, data protection

**Key Responsibilities:**
- Security audit and vulnerability scanning
- Auth/AuthZ, encryption, compliance
- Data protection and breach handling

**Example Prompts:**
- "Conduct security audit"
- "Check authentication and authorization"
- "Suggest security improvements"
- "Review data encryption"

---

### `knowledge-graph-testing`
- **Focus:** Unit, integration, and E2E tests, coverage, stability, regression
- **Ideal for:** Test writing, coverage analysis, test automation

**Key Responsibilities:**
- Unit, integration, and E2E tests
- Coverage, stability, and regression
- Test automation and reporting

**Example Prompts:**
- "Write unit tests for new handler"
- "Add integration tests"
- "Check coverage and stability"
- "Create E2E test scenario"

---

## Agent Selection Matrix

| Task Type | Agent |
|-----------|-------|
| Backend Go development | `knowledge-graph-backend-go` |
| API endpoint creation | `knowledge-graph-backend-go` |
| Database optimization | `knowledge-graph-backend-go` |
| Frontend/UI components | `knowledge-graph-frontend-svelte` |
| Svelte 5 development | `knowledge-graph-frontend-svelte` |
| State management | `knowledge-graph-frontend-svelte` |
| API contracts & DTOs | `knowledge-graph-integration` |
| OpenAPI specification | `knowledge-graph-integration` |
| Webhooks & OAuth | `knowledge-graph-integration` |
| Docker & containers | `knowledge-graph-infrastructure` |
| Monitoring & backups | `knowledge-graph-infrastructure` |
| CI/CD pipelines | `knowledge-graph-devops` |
| Deployment automation | `knowledge-graph-devops` |
| Performance profiling | `knowledge-graph-performance` |
| Caching strategies | `knowledge-graph-performance` |
| Security audit | `knowledge-graph-security` |
| Auth/AuthZ review | `knowledge-graph-security` |
| Unit/E2E tests | `knowledge-graph-testing` |
| Test coverage analysis | `knowledge-graph-testing` |
| Task routing | `knowledge-graph-orchestrator` |

---

## Quick Start Commands

### Backend
```bash
cd backend
go run ./cmd/server            # Start server
go test ./... -v               # Unit tests
go test -tags=integration ./... # Integration tests
go test -coverprofile=coverage.out ./... # With coverage
```

### Frontend
```bash
cd frontend
npm run dev                    # Dev server
npm run test:unit              # Vitest unit tests
npm run test                   # Playwright E2E tests
npm run test:bdd               # Cucumber BDD tests
npm run build                  # Production build
```

### NLP Service
```bash
cd nlp-service
uvicorn app.main:app --reload  # Dev server
pytest tests/ -v               # Tests
```

### Full Stack
```bash
docker-compose up              # All services
docker-compose up postgres redis # Only DB and cache
```

---

## How to Use Agents

1. **Identify the task or problem**
2. **Select the appropriate agent by area:**
   - Frontend/UI → `knowledge-graph-frontend-svelte`
   - Backend/Infrastructure → `knowledge-graph-backend-go`
   - API Integration → `knowledge-graph-integration`
   - Infrastructure/Containers → `knowledge-graph-infrastructure`
   - CI/CD/Deployment → `knowledge-graph-devops`
   - Performance → `knowledge-graph-performance`
   - Security → `knowledge-graph-security`
   - Testing → `knowledge-graph-testing`
   - Task routing → `knowledge-graph-orchestrator`
3. **Follow agent recommendations** when updating files

---

## Agent Workflow

```
User Request
    ↓
knowledge-graph-orchestrator
    ↓
┌─────────────────────────────────────┐
│  Analyze task type                  │
│  Delegate to appropriate agent      │
└─────────────────────────────────────┘
    ↓
Specialized Agent
    ↓
Execute task within domain
    ↓
Return results with documentation
```

---

## Best Practices

1. **Be specific in prompts** — specify exact files and task types
2. **Use the appropriate agent** — ensures correct context
3. **Follow project patterns** — agents know architectural conventions
4. **Verify commands** — all commands in documentation are tested
5. **Update documentation** — when making changes, update relevant sections
6. **Cross-reference agents** — some tasks may need multiple agents (e.g., backend + integration)

---

## Notes

- Agents are **not executable scripts** — they are metadata for task routing
- Each agent has specific domain expertise
- Agents maintain consistency across the codebase
- Documentation should be kept up-to-date with agent responsibilities

---

## Maintenance

This document is maintained by the `knowledge-graph-docs-maintenance` agent (not listed as active agent but responsible for documentation updates).

When adding new agents or changing responsibilities, update this file accordingly.
