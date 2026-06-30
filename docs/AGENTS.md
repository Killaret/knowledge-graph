# Agents in Knowledge Graph

**Updated:** June 2026  
**Status:** See [AGENTS_EN.md](AGENTS_EN.md) for current agent documentation in English

---

## Quick Reference

The project now uses **9 specialized AI agents** located in .cursor/rules/:

| Agent | Focus |
|-------|-------|
| knowledge-graph-orchestrator | Task routing & delegation |
| knowledge-graph-backend-go | Backend API & Go development |
| knowledge-graph-frontend-svelte | Frontend UI & Svelte 5 |
| knowledge-graph-integration | API contracts & DTOs |
| knowledge-graph-infrastructure | Docker & containers |
| knowledge-graph-devops | CI/CD & deployment |
| knowledge-graph-performance | Profiling & optimization |
| knowledge-graph-security | Security audit & Auth |
| knowledge-graph-testing | Unit/E2E tests |

---

## Full Documentation

For complete agent descriptions, responsibilities, and examples, see:
- **[AGENTS_EN.md](AGENTS_EN.md)** � Full English documentation

---

## Service Health Checks

### Dev Stack (docker-compose.yml)

**Health Check Script:**
```bash
# Check all containers
docker ps

# Check nginx gateway
curl http://localhost:8080/health

# Check backend health
curl http://localhost:9000/health

# Check graph service health
curl http://localhost:9091/health

# Check API endpoints through nginx
curl http://localhost:8080/api/v1/notes
curl http://localhost:8080/graph-service/api/v1/graph/full
```

### Personal Stack (docker-compose.personal.yml)

**Health Check Script:**
```bash
# Check personal backend
curl http://localhost:8085/health

# Check personal API gateway
curl http://localhost:8082/health

# Check personal graph service
curl http://localhost:8092/health
```

### Proxy Architecture

**Docker Environment:**
- Nginx (8080): Single API gateway
  - `/api/*` → Backend (8080)
  - `/graph-service/api/*` → Graph Service (9091)
- Frontend (5173): Production build with adapter-node
  - Uses nginx for microservice communication

**Dev Environment:**
- Vite Proxy (vite.config.ts): Local development
  - `/api/v1` → Backend (9000)
  - `/graph-service/api` → Graph Service (9091)

---

## Maintenance

This file is kept as a redirect to AGENTS_EN.md for historical reference.

---

## 🌐 Language Policy

**All user-facing content MUST be in English:**

- ✅ **Заметки** — заголовки и содержимое
- ✅ **Аннотации** — любые текстовые поля для пользователя
- ✅ **UI-строки** — кнопки, лейблы, плейсхолдеры, сообщения об ошибках
- ✅ **Тултипы и тосты** — сообщения GalacticLexicon
- ✅ **Commit-сообщения** — понятные, на английском

**Исключения:**
- Внутренние комментарии в коде — краткие пояснения на любом языке
- Имена переменных/функций — следуйте соглашениям проекта

### Examples:

```typescript
// ✅ Good
toast.success("Note created successfully");

// ❌ Bad
toast.success("Заметка создана успешно");
```

