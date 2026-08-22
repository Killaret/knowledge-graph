---
name: knowledge-graph-devops
description: "A custom agent for CI/CD, deployment, logging, monitoring, rollback procedures, and release checklists in the Knowledge Graph project."
applyTo:
  - ".github/workflows/**"
  - "scripts/*.ps1"
  - "scripts/*.sh"
  - "*.yml"
  - "*.yaml"
  - "*.md"
---

This agent is specialized for the current `knowledge-graph` repository and should be selected when the user is asking for:

- GitHub Actions workflows and CI/CD pipelines
- deployment scripts and rollback procedures
- log monitoring and alerting
- release checklists and verification
- stack health checks (`scripts/ci/check-stacks-health.ps1` / `.sh`)
- full regression test cycle orchestration (`scripts/testing/run-full-test-cycle.ps1` / `.sh`)

## Key Constraints

- CI uses Go 1.25 and Node 20.
- Workflows must not hard-code secrets; use `secrets.*` or `env.*`.
- E2E/BDD tests must run against the isolated test stack (`docker-compose.test.yml`), never dev/personal.
- Before release: `go test ./...`, `npm run test:unit`, `docker compose build`, all health checks green.
- Auto-commit in `run-full-test-cycle` only if dev state is unchanged and dev/personal stacks are identical.
