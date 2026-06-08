# AI Project Instructions

**Knowledge Graph** - система управления заметками с графовыми связями и NLP-анализом.

## 🎯 Quick Start (PRIORITY 1)

**Read this first:** `.koda/compact-rules.md` (116 lines of practical rules)

This contains everything you need:
- Project stack (Go, Svelte 5, Python FastAPI, Docker)
- Key AI rules (tests, conventions, security)
- Common commands
- Project structure
- AI-specific hints

## 📁 Where to Find What

| Purpose | File | Size | Priority |
|---------|------|------|----------|
| **Compact rules** | `.koda/compact-rules.md` | 116 lines | 🔥 PRIMARY |
| **AI tools guide** | `AI_TOOLS.md` | 21 lines | HIGH |
| **Work report** | `WORK_REPORT.md` | 75 lines | REFERENCE |
| **Detailed agents** | `.cursor/rules/*.md` | 11000+ lines | DETAILED |
| **Cursor rules** | `.cursorrules` | 23 lines | CURSOR |
| **Copilot rules** | `.github/copilot/copilot-instructions.md` | 166 lines | COPILOT |

## 🛠️ How Different AI Tools Load Rules

### Cursor IDE
- ✅ Automatically loads `.cursorrules` (main file)
- ✅ Automatically loads `.cursor/rules/*.md` (detailed rules)
- ✅ Updated to point to compact rules first

### GitHub Copilot
- ✅ Automatically loads `.github/copilot/copilot-instructions.md`
- ✅ Updated to point to compact rules first

### Other AI Tools
- Read `.koda/compact-rules.md` directly
- Read `AI_TOOLS.md` for navigation
- Ignore the 11000+ line `.koda/` system (it's outdated)

## 🚀 Quick Commands

```bash
# Backend
cd backend && go test ./...           # Run tests
cd backend && go build ./cmd/server   # Build server

# Frontend
cd frontend && npm run test:unit      # Unit tests
cd frontend && npm run test           # E2E tests

# Docker
docker compose up -d                  # Start services
docker compose -f docker-compose.personal.yml up -d  # Personal instance
```

## 📋 Project Stack

- **Backend:** Go 1.23+, PostgreSQL, Redis, gRPC
- **Frontend:** Svelte 5, Vitest, Playwright
- **NLP:** Python FastAPI, sentence-transformers
- **Infrastructure:** Docker, Docker Compose

## 🎯 Key Rules Summary

1. **Tests mandatory** - Always run `go test ./...` after changes
2. **Clean Architecture** - Domain/Application/Infrastructure layers
3. **No globals** - Use dependency injection
4. **TypeScript everywhere** - Frontend strict typing
5. **Docker multi-stage builds** - Optimized images

## 📖 For More Details

- Full documentation: `docs/` directory
- Testing guide: `docs/TESTING.md`
- Configuration: `docs/CONFIGURATION.md`
- API docs: `openAPI.yaml`

---

**Tip:** Start with `.koda/compact-rules.md` - it has everything you need in 116 lines. Only dive into the detailed `.cursor/rules/` if you need specific domain knowledge.