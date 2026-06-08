# AI Tools for Knowledge Graph

## Single Source of Truth

**Read:** AI_RULES.md (116 lines)

This replaces the outdated 11000+ line .koda/ system.

## IDE Integration

### Cursor IDE
- Automatically loads `.cursorrules`
- Points to AI_RULES.md
- Works out of the box

### GitHub Copilot
- Automatically loads `.github/copilot/copilot-instructions.md`
- Points to AI_RULES.md
- Works out of the box

### Other AI Tools
- Read AI_RULES.md directly
- No configuration needed
- Single file for all tools

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

## Configuration

- `.editorconfig` - consistent code formatting
- `.gitignore` - excludes .koda/, .devin/ (local configs)