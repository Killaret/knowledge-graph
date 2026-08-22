# Copilot Instructions for Knowledge Graph

**Read:** `.windsurfrules` and `docs/PROJECT_REVIEW_AI_AGENTS.md` (single source of truth)

This file contains all project rules in 116 lines instead of 11000+ lines of outdated documentation.

**Quick Reference:**
- Backend: Go 1.25+, Clean Architecture, DDD
- Frontend: Svelte 5, Vitest, Playwright
- Tests: Always run after changes, coverage >60%
- Docker: Multi-stage builds, health checks, volumes
- 🌐 **Russian UI by default** with English i18n support

## Language Policy

**User-facing runtime content uses Russian by default (`ru` locale), with English (`en`) support through i18n keys:**
- UI strings, buttons, labels, placeholders, errors
- Toast messages, tooltips, GalacticLexicon
- Note titles and content

**MUST be in English:**
- Code identifiers and variable names
- Commit messages
- API error codes
- Public code comments, README, authoritative docs

**User content (note titles/bodies) may be in any language.**
**Exceptions:** Internal code comments — any language OK.

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