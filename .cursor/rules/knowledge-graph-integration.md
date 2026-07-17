# Cursor Rule: knowledge-graph-integration

API contracts, OpenAPI, DTOs, and service routing. Every public endpoint must
have a Swagger annotation, a DTO struct, and a corresponding TypeScript type.

---

## API Versioning

All new routes live under `/api/v1/`. Legacy non-versioned routes (e.g. `/notes`)
are kept for backward compatibility only — do NOT add new ones there.

```go
// backend/cmd/server/router.go
v1 := r.Group("/api/v1")
{
    v1.POST("/notes", writeLimiter, noteHandler.Create)
    v1.GET("/notes/:id", cacheControlMiddleware(60), noteHandler.Get)
    v1.PUT("/notes/:id", writeLimiter, noteHandler.Update)
    v1.DELETE("/notes/:id", writeLimiter, noteHandler.Delete)
    v1.GET("/graph/all", cacheControlMiddleware(300), graphHandler.GetFullGraph)
    // auth, tags, achievements, backup ...
}
```

---

## Swagger / OpenAPI Annotations (swaggo/gin-swagger)

Every handler function must have a full Swagger comment block:

```go
// @Summary      Create a new note
// @Description  Creates a note and enqueues NLP processing (keywords + embedding)
// @Tags         notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body   body      notehandler.CreateNoteRequest  true  "Note payload"
// @Success      201    {object}  notehandler.NoteResponse
// @Failure      400    {object}  notehandler.ErrorResponse  "Validation error"
// @Failure      401    {object}  notehandler.ErrorResponse  "Unauthorized"
// @Failure      500    {object}  notehandler.ErrorResponse  "Internal error"
// @Router       /api/v1/notes [post]
func (h *Handler) Create(c *gin.Context) { ... }
```

Generate the spec: `cd backend && swag init -g cmd/server/main.go`.
The spec is served at `/openapi.yaml`; Swagger UI at `/swagger/index.html`.

---

## DTO Patterns

DTOs live in the handler package (`interfaces/api/notehandler/`), NOT in domain.

```go
// Request DTO — validated by go-playground/validator tags
type CreateNoteRequest struct {
    Title    string                 `json:"title"    validate:"required,max=200"`
    Content  string                 `json:"content"  validate:"max=10000"`
    Type     string                 `json:"type"     validate:"omitempty,oneof=star highlight link"`
    Metadata map[string]interface{} `json:"metadata"`
}

// Response DTO — mirrors API contract, not the domain entity
type NoteResponse struct {
    ID        string                 `json:"id"`
    Title     string                 `json:"title"`
    Content   string                 `json:"content"`
    Type      string                 `json:"type"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
}

// Standard error shape — used across ALL handlers
type ErrorResponse struct {
    Error   string `json:"error"`
    Details string `json:"details,omitempty"`
}
```

---

## Standard Error Response Format

All error responses use the same JSON envelope:

```json
{ "error": "note not found" }
{ "error": "validation failed", "details": "title: required" }
```

HTTP status codes:
| Situation | Status |
|-----------|--------|
| Validation / bad JSON | 400 |
| Not authenticated | 401 |
| Not authorized (ownership) | 403 |
| Resource not found | 404 |
| Rate limited | 429 |
| Unexpected server error | 500 |

---

## Nginx Proxy Routing

`nginx.conf` (port 8080) routes traffic:

```
/api/*              → backend:8080         (Go Gin server)
/graph-service/*    → graph-service:9091   (Go graph service, strip prefix)
/health             → 200 OK inline        (nginx itself)
```

Port 8081 proxies everything to `frontend:3000` (SvelteKit Node server).

```nginx
location /api/ {
    set $backend_host backend:8080;
    proxy_pass http://$backend_host;
}
location /graph-service/ {
    set $graph_service_host graph-service:9091;
    rewrite ^/graph-service(/.*)$ $1 break;
    proxy_pass http://$graph_service_host;
}
```

**Never hard-code service IPs** — always use Docker Compose service names
resolved at runtime via `resolver 127.0.0.11`.

---

## Dev vs Docker Environment Configuration

| Variable | Dev (local) | Docker compose |
|----------|-------------|----------------|
| `VITE_API_URL` | `http://localhost:8080` | `http://localhost:8080` |
| `VITE_API_TARGET` (SSR) | `http://localhost:9000` | `http://backend:8080` |
| `GRAPH_SERVICE_URL` (SSR) | `http://localhost:9091` | `http://graph-service:9091` |
| `DATABASE_URL` | `postgresql://...@localhost:15432/...` | `postgresql://...@postgres:5432/...` |
| `REDIS_URL` | `localhost:6379` | `redis:6379` |

Frontend browser code must always route through nginx (`VITE_API_URL`).
Frontend SSR hooks (`hooks.server.ts`) use `VITE_API_TARGET` to bypass nginx.

---

## TypeScript Type Sync

Every DTO change in Go must be mirrored in `frontend/src/shared/types/`:

```typescript
// frontend/src/shared/types/note.ts
export interface Note {
  id: string;
  title: string;
  content: string;
  type: 'star' | 'planet' | 'comet' | 'galaxy' | 'asteroid';
  metadata?: Record<string, unknown>;
  created_at: string;   // ISO-8601 string from JSON
  updated_at: string;
}

export interface CreateNoteRequest {
  title: string;
  content?: string;
  type?: 'star' | 'planet' | 'comet' | 'galaxy' | 'asteroid';
  metadata?: Record<string, unknown>;
}
```

---

## Graph Service Integration

The graph-service (`services/graph-service/`) exposes its own HTTP API on port
9091 (gRPC on 9090). It is accessed by:
- Frontend SSR: `GRAPH_SERVICE_URL` env var (direct Docker network)
- Browser client: via nginx `/graph-service/` prefix

```typescript
// frontend: VITE_GRAPH_SERVICE_URL = http://localhost:8080/graph-service
// SSR:      GRAPH_SERVICE_URL      = http://graph-service:9091
```

---

## Anti-Patterns

```go
// ❌ Domain entity as JSON response (leaks internal structure)
c.JSON(200, domainNote)  // note.Note has unexported fields — returns {}

// ❌ Missing Swagger annotation
func (h *Handler) Create(c *gin.Context) { ... }  // not documented

// ❌ Inconsistent error format
c.JSON(400, "bad request")          // plain string, not ErrorResponse
c.JSON(400, gin.H{"msg": "error"})  // wrong key name

// ❌ Bypassing nginx in browser code
const API = 'http://backend:9000';  // not reachable from browser

// ❌ Hardcoded port in frontend
const url = 'http://localhost:9000/notes';  // must use VITE_API_URL env
```
