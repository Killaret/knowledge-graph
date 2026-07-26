# Cursor Rule: knowledge-graph-orchestrator

Always active. Routes every task to the correct specialized agent and enforces cross-cutting constraints.

---

## Agent Decision Matrix

| Task type | Primary agent | Secondary agent |
|-----------|--------------|-----------------|
| New Go endpoint / handler | `backend-go` | `integration` (OpenAPI spec) |
| GORM model / migration | `backend-go` | `data` |
| Redis cache / asynq task | `backend-go` | `performance` |
| Svelte component / route | `frontend-svelte` | — |
| D3-force / Three.js graph | `frontend-svelte` | `performance` |
| API contract change | `integration` | `backend-go` + `frontend-svelte` |
| Docker / docker-compose | `infrastructure` | `devops` |
| Nginx routing | `infrastructure` | `integration` |
| CI/CD pipeline | `devops` | — |
| Logs / rollback | `devops` | — |
| JWT / CORS / rate-limit | `security` | `backend-go` |
| pgvector / embedding query | `data` | `performance` |
| Redis key naming / pooling | `data` | `performance` |
| Go table-driven tests | `testing` | `backend-go` |
| Vitest / Playwright tests | `testing` | `frontend-svelte` |
| NLP FastAPI endpoint | `nlp` | `integration` |
| HuggingFace model loading | `nlp` | `infrastructure` |
| P95 / caching strategy | `performance` | `backend-go` |
| Secret management | `security` | `infrastructure` |
| Yandex OAuth | `security` | `backend-go` |

---

## Escalation Rules

Invoke **multiple agents sequentially** when a task crosses boundaries:

1. **New feature end-to-end** (e.g., "Add note sharing"):
   - `backend-go` → domain entity + repository + handler
   - `integration` → DTO + OpenAPI annotation
   - `frontend-svelte` → component + API client call
   - `testing` → unit + integration + E2E tests
   - `security` → verify auth guard on new endpoint

2. **Schema migration**:
   - `data` → write raw SQL migration file
   - `backend-go` → update GORM model + repository
   - `testing` → testcontainers integration test

3. **Performance regression**:
   - `performance` → identify hot path
   - `backend-go` → fix query / add index
   - `data` → confirm pgvector index strategy
   - `devops` → verify with `docker compose logs -f backend`

4. **Security audit**:
   - `security` → JWT middleware + CORS + rate-limit review
   - `backend-go` → confirm no global state, proper error propagation
   - `devops` → confirm no secrets in image layers

---

## Common Task Routing Scenarios

### "Write a new endpoint for note recommendations"
```
1. backend-go   → domain service in internal/application/recommendation/
2. backend-go   → Gin handler in internal/interfaces/api/notehandler/
3. integration  → @Summary/@Param Swagger comment + DTO struct
4. frontend-svelte → API client function in frontend/src/shared/api/notes.ts
5. testing      → table-driven unit test + testcontainers integration test
```

### "GraphCanvas is slow with 500 nodes"
```
1. performance  → profile D3-force tick count, canvas draw calls
2. frontend-svelte → reduce simulation alphaDecay, memoize link data
3. performance  → check backend graph query (graph-service HTTP_PORT 9091)
4. data         → verify pgvector IVFFlat index on note_embeddings
```

### "Add Yandex OAuth login"
```
1. security     → verify PKCE flow, state parameter, token expiry
2. backend-go   → handler in internal/interfaces/api/handlers/auth/
3. backend-go   → Redis PKCE store via internal/auth/redis_store.go StorePKCE()
4. frontend-svelte → YandexLoginButton.svelte + auth/yandex/callback route
5. testing      → mock OAuth callback in Playwright E2E
```

### "NLP service returns 503 on startup"
```
1. nlp          → ensure_model_loaded() in nlp-service/app/nlp_utils.py
2. infrastructure → healthcheck start_period: 600s in docker-compose.yml (nlp service)
3. devops       → docker compose logs -f nlp
```

---

## Cross-Agent Coordination Patterns

### Pattern: Backend + Frontend contract sync
When `backend-go` changes a response DTO, `integration` must update the OpenAPI
spec, and `frontend-svelte` must update the TypeScript type in
`frontend/src/shared/types/`. Never let these drift.

### Pattern: Migration → model → test atomicity
`data` writes the migration file → `backend-go` updates the GORM model in the
same PR → `testing` adds a testcontainers test that runs the migration. All
three must land together.

### Pattern: Redis key ownership
All Redis keys are documented in `knowledge-graph-data.md`. Before adding a new
key, `backend-go` must register it in the naming table; `performance` reviews
the TTL; `security` confirms no PII in key names.

---

## Hard Rules (enforce on every task)

- User-facing runtime text uses Russian via i18n (`ru` locale) with English (`en`) fallback; code identifiers, API error codes, commit messages, and public docs MUST be in **English**.
- **No global variables** — pass dependencies explicitly.
- Coverage target: **>60%** for every new package.
- Repositories return **domain entities only** (`*note.Note`, not `NoteModel`).
- Svelte components use **Svelte 5 runes only** — never `writable`/`readable` from Svelte 4.
- API routes follow `/api/v1/` prefix (see `backend/cmd/server/router.go`).
- Every new Docker service must have a `/health` endpoint.
