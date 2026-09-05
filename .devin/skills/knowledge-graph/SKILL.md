# Knowledge Graph Project — Devin Skill

## Authority

Read `.windsurfrules` first. It is the single normative source for architecture, security, testing, documentation, language, Docker, and data-preservation rules. This skill adds Devin-specific navigation and workflow only.

## Project

Knowledge Graph is a note-management system with graph relationships and NLP-powered analysis.

| Layer | Current implementation |
|-------|------------------------|
| Backend | Go 1.25, Gin v1.12.0, GORM v1.25, pgx/v5 |
| Data | PostgreSQL 16 + pgvector, Redis 7 with go-redis/v9.14.1, MongoDB 7 |
| Async | Asynq v0.26.0 |
| Frontend | SvelteKit, Svelte 5 runes, strict TypeScript, ky, D3-force, Three.js |
| NLP | Python 3.11, FastAPI, sentence-transformers, YAKE, NLTK |
| Testing | Testify, testcontainers, Vitest, Playwright, Cucumber, pytest |
| Infrastructure | Docker Compose, multi-stage images, nginx gateway |

## Primary Navigation

- Backend wiring: `backend/cmd/server/main.go`, `backend/cmd/worker/main.go`
- Backend layers: `backend/internal/{domain,application,infrastructure,interfaces}`
- Frontend: `frontend/src/{shared,components,entities,features,widgets,routes}`
- NLP: `nlp-service/app/{main.py,models.py,nlp_utils.py}`
- Runtime configuration: `knowledge-graph.config.json`, `config/`
- Architecture: `docs/ARCHITECTURE_SUMMARY.md`
- Testing: `docs/TESTING.md`, `docs/REGRESSION_TEST_PLAN.md`
- Agent handoff: `docs/AI_HANDOFF.md`
- Agent protocol: `docs/AI_AGENT_PROTOCOL.md`
- Current audit: `docs/AI_PROCESS_AUDIT.md`
- Roadmap: `ROADMAP.md`, `ROADMAP.ru.md`

## Devin Workflow

1. Read `.windsurfrules`, `docs/AI_HANDOFF.md`, and `docs/AI_AGENT_PROTOCOL.md` first; then read relevant subsystem documentation.
2. Update `docs/AI_HANDOFF.md` when handing off or taking over a task.
3. Search the codebase before deciding on an implementation.
4. Follow existing constructors, dependency injection, error handling, and test patterns.
5. Add a regression test for every discovered defect.
6. Run the narrowest relevant verification first, then the required subsystem checks.
7. Review the diff for unrelated files and secrets before committing.
8. Never start the personal stack unless the user explicitly requests it.

## Task Routing

| Task | Read first | Required verification |
|------|------------|-----------------------|
| Go backend/API | neighboring handler, application service, domain ports, repository | `cd backend && go test ./...` |
| Database/migration | adjacent migrations, domain repository, GORM implementation | relevant Go unit/integration tests |
| Svelte/UI | neighboring feature/widget/entity and tests | `cd frontend && npm run test:unit` |
| API contract | handler DTOs, frontend API client, OpenAPI docs | backend + frontend relevant tests |
| NLP | `nlp-service/app/main.py`, `nlp_utils.py`, tests | `cd nlp-service && pytest tests/ -v` |
| Docker/infrastructure | all affected Compose variants and health checks | config validation and relevant health checks |
| E2E/BDD | `docs/TESTING.md` and test scripts | isolated test stack only |
| Security | middleware, validation, rate limiting, secret flow | targeted tests plus security review |

## Current Runtime Facts

- The NLP embedding model is preloaded during FastAPI lifespan startup through `ensure_model_loaded()`; `/health` verifies readiness.
- Dev host gateway is `http://127.0.0.1:18080`; backend direct is `http://127.0.0.1:9000`.
- Test frontend is `http://127.0.0.1:3002`; backend is `http://127.0.0.1:18083`; NLP is `http://127.0.0.1:15002`.
- Personal volumes contain live user data and must never be deleted without explicit approval and the required backup procedure.
- E2E and BDD tests run only against `docker-compose.test.yml`; stop dev and personal stacks first.

## Common Commands

```powershell
# Backend
cd backend; go test ./...
cd backend; go test -tags=integration ./...
cd backend; go build ./cmd/server

# Frontend
cd frontend; npm run test:unit
cd frontend; npm run build

# NLP
cd nlp-service; pytest tests/ -v

# Isolated test stack
.\scripts\testing\start-test.ps1
.\scripts\testing\seed-test-data.ps1
.\scripts\testing\stop-test.ps1

# Canonical full regression
.\scripts\testing\run-full-test-cycle.ps1
```

## Commands

| Trigger | Action |
|---|---|
| `/kg-handoff` | Отдельный skill `.devin/skills/kg-handoff`. Прочитать `docs/AI_HANDOFF.md`, выполнить пункты, закреплённые за Devin, пропуская те, чьё условие не наступило. По окончании обновить `docs/AI_HANDOFF.md` и связанный `docs/tasks/<id>.md`. |
| `/kg-review` | Отдельный skill `.devin/skills/kg-review`. Проверить работу другого агента по `docs/AI_HANDOFF.md` и `docs/tasks/<id>-review-findings.md`. Смотреть дифф, а не описание автора; проверять исполнением; отдельно оценить, способен ли предложенный автором способ проверки поймать дефект. Записать замечания в `docs/tasks/<id>-review-findings.md` и обновить инбокс. |

## Priority

Resolve conflicts in this order:

1. Security
2. Correctness and data integrity
3. Performance
4. Convenience
