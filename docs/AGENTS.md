# AI-Assisted Development in Knowledge Graph

## Source of Truth

The authoritative rules for AI-assisted development are in [`.windsurfrules`](../.windsurfrules). This document is a short operational reference and must not override those rules.

## Active Tools and Roles

| Role | Tool | Scope |
|------|------|-------|
| Primary implementation agent | Windsurf SWE 1.7 Max | Code, refactoring, tests, repository navigation, commits |
| Strategic advisor | DeepSeek | Architecture, Roadmap planning, prompt design, high-level review |
| CLI repository agent | Devin | Audits, implementation, verification, and automation through the project skill |
| Runtime semantic analysis | Python NLP service | Embeddings, keyword extraction, similarity, recommendations |

Cursor, Continue/Koda, GitHub Copilot, and GitHub custom-agent configurations are not used in this repository.

## Instruction Files

| File | Purpose |
|------|---------|
| [`.windsurfrules`](../.windsurfrules) | Normative project rules |
| [`.devin/skills/knowledge-graph/SKILL.md`](../.devin/skills/knowledge-graph/SKILL.md) | Devin-specific navigation and workflow |
| [`.windsurf/rules.md`](../.windsurf/rules.md) | Pointer to the normative rules |
| [`ARCHITECTURE_SUMMARY.md`](ARCHITECTURE_SUMMARY.md) | Architecture reference |
| [`TESTING.md`](TESTING.md) | Test environments and commands |
| [`REGRESSION_TEST_PLAN.md`](REGRESSION_TEST_PLAN.md) | Canonical regression sequence |
| [`../ROADMAP.md`](../ROADMAP.md) | Current roadmap |

## Required Workflow

1. Read `.windsurfrules` and the relevant subsystem documentation.
2. Inspect existing implementation and nearby tests before changing code.
3. Follow Clean Architecture on the backend and FSD/Svelte 5 runes on the frontend.
4. Add regression coverage for discovered defects.
5. Run relevant tests after changes.
6. Use only the isolated test stack for E2E and BDD.
7. Never start the personal stack unless explicitly requested.
8. Never delete personal stack volumes without explicit approval and the required backup.

## Priority

1. Security
2. Correctness and data integrity
3. Performance
4. Convenience
