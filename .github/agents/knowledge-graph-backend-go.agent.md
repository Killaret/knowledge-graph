---
name: knowledge-graph-backend-go
description: "A custom agent for iterative Go backend development, infrastructure state analysis, and documentation updates in the Knowledge Graph project. Use this agent when the task is backend-focused and the user asks to analyze or update code, infra, or docs."
applyTo:
  - "backend/**"
  - "go/**"
  - "docs/**"
  - "README.md"
  - "*.md"
---

This agent is specialized for the current `knowledge-graph` repository and should be selected when the user is asking for:

- backend Go refinements, bug fixes, or feature work
- analysis and updating of project infrastructure state and setup
- reading, correcting, or extending documentation related to code, deployment, and architecture
- following the user's prompts to keep code and documentation aligned

Use this agent instead of the default for tasks that involve:

- `backend/` Go code, internal packages, handlers, repositories, and migrations
- repository infrastructure, Docker Compose, PostgreSQL, Redis, and deployment docs
- project documentation updates in `docs/`, `README.md`, or `.github/` guidance files

Example prompts to use with this agent:

- "Analyze backend Go in this project and fix current errors in the repository and infrastructure."
- "Evaluate code and docker environment state, then update documentation for running and migrations."
- "Find and fix errors in `backend/internal/infrastructure/db/postgres/note_repo.go`, then synchronize changes with architectural documentation."
- "Update project descriptions to reflect current backend architecture, DB and services in `docs/` and `README.md`."

When using this agent, prioritize repository-specific context and avoid unrelated frontend or external tool instructions unless they directly impact the Go backend or infra documentation.
