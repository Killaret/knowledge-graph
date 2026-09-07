# AI Tooling and Agent Workflow

## Authority

[`.windsurfrules`](../.windsurfrules) is the single normative source for AI-assisted development in Knowledge Graph. If this document differs from it, `.windsurfrules` wins.

## Active Roles

### Windsurf SWE 1.7 Max

The primary implementation agent for:

- code and refactoring;
- repository navigation;
- unit, integration, E2E, and BDD tests;
- documentation updates required by behavior changes;
- commits requested by the user.

### DeepSeek

The strategic advisor for:

- architecture alternatives;
- Roadmap planning;
- prompt design;
- high-level design review.

Its recommendations must be checked against the repository implementation and `.windsurfrules` before adoption.

### Devin

The CLI repository agent. Its project-specific entry point is:

- [`.devin/skills/knowledge-graph/SKILL.md`](../.devin/skills/knowledge-graph/SKILL.md)

Devin reads `.windsurfrules`, inspects the implementation, applies changes when requested, and performs relevant verification.

### Runtime NLP Service

The Python FastAPI NLP service is an application runtime component rather than a development agent. It provides embeddings, keyword extraction, similarity, and recommendation inputs. Its embedding model is preloaded during application startup and checked by `/health`.

## Inactive Tooling

The repository does not maintain project-specific configuration for:

- Cursor;
- Continue/Koda;
- GitHub Copilot;
- GitHub custom agents.

Do not reintroduce those configurations without an explicit project decision.

## Shared Prompts

The `.devin/prompts/` directory contains reusable prompts for all AI agents:

- [`.devin/prompts/MASTER_PROMPT.md`](../.devin/prompts/MASTER_PROMPT.md) — paste at the start of implementation or review chats.
- [`.devin/prompts/ANALYSIS_PROMPT.md`](../.devin/prompts/ANALYSIS_PROMPT.md) — use for strategic architecture and roadmap discussions.
- [`.devin/prompts/MASTER_PROMPT_RU.md`](../.devin/prompts/MASTER_PROMPT_RU.md) — Russian-language master prompt.

## Task Routing

| Task | Primary context |
|------|-----------------|
| Go API, domain, repositories | Backend Clean Architecture section in `.windsurfrules` |
| PostgreSQL, Redis, MongoDB, pgvector | Backend and data rules plus adjacent implementation |
| Svelte components and graph UI | Frontend FSD and Svelte 5 sections |
| API contracts | Handler DTOs, frontend API clients, OpenAPI documentation |
| NLP and embeddings | NLP service section and `nlp-service/` implementation |
| Docker and nginx | Docker section and all affected Compose variants |
| Tests | Testing section and `docs/TESTING.md` |
| Security | Security section; security takes precedence over all other concerns |

## Verification

- Backend: `cd backend && go test ./...`
- Frontend: `cd frontend && npm run test:unit`
- NLP: `cd nlp-service && pytest tests/ -v`
- E2E/BDD: use only the isolated test stack described in `docs/TESTING.md`.
- Full regression: use the canonical script and sequence documented in `docs/REGRESSION_TEST_PLAN.md`.

## Language Policy

- The committed runtime locale is English by default, with Russian support through the same i18n keys.
- Code identifiers, commit messages, API error codes, and authoritative documentation are in English.
- User-created note content may use any language.

## Priority

1. Security
2. Correctness and data integrity
3. Performance
4. Convenience
