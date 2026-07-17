# Agents in Knowledge Graph

**Updated:** July 2026
**Status:** See [AGENTS_EN.md](AGENTS_EN.md) for full documentation (11 agents)

---

## Operations & Roadmap

- **[Regression Testing Plan Summary](REGRESSION_TEST_PLAN_SUMMARY.md)** — 24-step isolated regression cycle
- **[Testing Commands & Procedures](TESTING_COMMANDS.md)** — Commands for unit, integration, E2E, BDD, and stack management
- **[UI Modernization Roadmap](UI_MODERNIZATION_ROADMAP.md)** — Canvas, NoteCard, multilingual lexicon, and UX improvements

---

## Quick Reference

The project uses **11 specialized AI agents** defined across multiple AI tools:

| Agent | Focus |
|-------|-------|
| knowledge-graph-orchestrator | Task routing & delegation |
| knowledge-graph-backend-go | Go API, PostgreSQL, Redis, MongoDB, JWT |
| knowledge-graph-frontend-svelte | Svelte 5, TypeScript, UI/UX |
| knowledge-graph-integration | OpenAPI, DTOs, API contracts |
| knowledge-graph-infrastructure | Docker, nginx, monitoring |
| knowledge-graph-devops | CI/CD, deployment |
| knowledge-graph-performance | Profiling, caching, P95 |
| knowledge-graph-security | Auth/AuthZ, audit, encryption |
| knowledge-graph-testing | Unit/integration/E2E/BDD |
| knowledge-graph-nlp | Python FastAPI, NLP, HuggingFace *(NEW)* |
| knowledge-graph-data | DB migrations, pgvector, schemas *(NEW)* |

---

## AI Tool Configuration

| Tool | Rules Location |
|------|---------------|
| Cursor AI | `.cursor/rules/*.md` |
| Koda VSCode (koda-base/koda-pro) | `.continue/rules/*.md` |
| Windsurf/Cascade | `.windsurfrules` |
| Devin | `.devin/skills/knowledge-graph/SKILL.md` |

---

## Full Documentation

- **[AGENTS_EN.md](AGENTS_EN.md)** — Full English documentation, agent descriptions, selection matrix

---

## Service Health Checks

### Dev Stack (docker-compose.yml)

```bash
curl http://localhost:8080/health           # Nginx gateway
curl http://localhost:9000/health           # Backend
curl http://localhost:9091/health           # Graph service
curl http://localhost:8080/api/v1/notes     # Notes API
```

### Personal Stack (docker-compose.personal.yml)

```bash
curl http://localhost:8085/health           # Personal backend
curl http://localhost:8082/health           # Personal API gateway
curl http://localhost:8092/health           # Personal graph service
```

---

## Language Policy

**All user-facing content MUST be in English:**
- UI strings (buttons, labels, placeholders, errors)
- Toast messages and tooltips (GalacticLexicon)
- Note titles and content
- Commit messages

**Exceptions:** Internal code comments (any language OK)

```typescript
// ✅ Good
toast.success("Note created successfully");

// ❌ Bad
toast.success("Заметка создана успешно");
```

---
