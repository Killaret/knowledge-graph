# Copilot Instructions for Knowledge Graph

**Read:** AI_RULES.md (single source of truth)

This file contains all project rules in 116 lines instead of 11000+ lines of outdated documentation.

**Quick Reference:**
- Backend: Go 1.23+, Clean Architecture, DDD
- Frontend: Svelte 5, Vitest, Playwright
- Tests: Always run after changes, coverage >60%
- Docker: Multi-stage builds, health checks, volumes

## Quick Commands

```bash
# Backend
cd backend && go test ./...
cd backend && go build ./cmd/server

# Frontend
cd frontend && npm run test:unit
cd frontend && npm run test

# Docker
docker compose up -d
docker compose -f docker-compose.personal.yml up -d
```