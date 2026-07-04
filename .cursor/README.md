# Cursor Agent Configuration

This directory contains rules for Cursor AI agents.

## Important
- Cursor reads all files in `.cursor/rules/` automatically
- Orchestrator analyzes and routes tasks to specialized agents
- Each agent has its own domain, rules, and code examples
- Do not delete this directory

## Agent Files (11 total)

| File | Agent | Domain |
|------|-------|--------|
| `knowledge-graph-orchestrator.md` | Orchestrator | Task routing |
| `knowledge-graph-backend-go.md` | Backend Go | Go, PostgreSQL, Redis, MongoDB |
| `knowledge-graph-frontend-svelte.md` | Frontend Svelte | Svelte 5, TypeScript, UI/UX |
| `knowledge-graph-integration.md` | Integration | OpenAPI, DTOs, nginx |
| `knowledge-graph-infrastructure.md` | Infrastructure | Docker, monitoring |
| `knowledge-graph-devops.md` | DevOps | CI/CD, deployment |
| `knowledge-graph-performance.md` | Performance | Profiling, caching |
| `knowledge-graph-security.md` | Security | Auth, audit |
| `knowledge-graph-testing.md` | Testing | Unit/E2E/BDD |
| `knowledge-graph-nlp.md` | NLP | Python FastAPI, HuggingFace |
| `knowledge-graph-data.md` | Data | Migrations, pgvector, schemas |

## Other AI Tool Configurations

- **Koda VSCode (koda-base/koda-pro):** `.continue/rules/*.md`
- **Windsurf/Cascade:** `.windsurfrules`
- **Devin:** `.devin/skills/knowledge-graph/SKILL.md`

## Full Documentation

See `docs/AGENTS_EN.md` for complete agent descriptions, selection matrix, and examples.

## Language Policy

All user-facing content MUST be in English:
- UI strings (buttons, labels, placeholders, errors)
- Toast messages and tooltips
- Note titles and content
- Commit messages

**Exceptions:** Internal code comments (brief explanations in any language OK)
