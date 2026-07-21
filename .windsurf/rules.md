# Windsurf Agent Context

This file contains unified context for Windsurf AI and the full agent list.
Do not delete.

## Agent List (11 agents)
1. knowledge-graph-orchestrator — coordination and delegation
2. knowledge-graph-backend-go — Go backend, API, DB, auth
3. knowledge-graph-frontend-svelte — Svelte UI, components and state
4. knowledge-graph-integration — OpenAPI, contracts, webhooks
5. knowledge-graph-infrastructure — Docker, K8s, monitoring
6. knowledge-graph-devops — CI/CD, deploy, logging
7. knowledge-graph-performance — profiling, caching, P95
8. knowledge-graph-security — audit, auth, compliance
9. knowledge-graph-testing — unit/integration/E2E/BDD, coverage
10. knowledge-graph-nlp — Python FastAPI, NLP, HuggingFace
11. knowledge-graph-data — DB migrations, pgvector, schemas

## Commands and Permissions
- Use Windsurf as a unified context file
- If the question is about backend — apply backend-go rules
- If the question is about frontend — apply frontend-svelte rules
- If the question is about integration — apply integration rules
- If the question is about infrastructure, CI/CD, or deploy — apply infrastructure/devops
- If performance complexity — apply performance
- If security — apply security
- If testing — apply testing

## Tools
- `AI_RULES.md` — unified AI rules (single source of truth)
- `docs/AGENTS_EN.md` — full agent descriptions
- `docs/ROADMAP.md` — project roadmap

## Instructions
- Read this file first
- Do not delete `.windsurf/rules.md`
- Use it as the main context for Windsurf
- Redirect requests to the appropriate agents

## Protection
This file and the `.windsurf/` directory are important.
Do not delete.
