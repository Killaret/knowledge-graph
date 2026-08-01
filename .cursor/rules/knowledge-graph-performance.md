# Cursor Rule: knowledge-graph-performance

Performance targets, caching strategies, and optimization patterns for all
services in the Knowledge Graph stack.

## Manual Found → Automated Covered

If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue. Choose the test level by severity and scope:

- **unit** — pure logic, validators, or utilities (e.g. `errorMessage.ts`, email validation).
- **integration** — handlers, repositories, or routes (e.g. `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
- **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g. public graph, achievements, SSE fallback).

The test should fail before the fix (where safe) and pass after the fix. If the defect depends on manual data or config setup, fix the seed or config script — not only the instructions.

---

## P95 Targets

| Endpoint | P95 target | Notes |
|----------|-----------|-------|
| `GET /api/v1/notes` | < 100ms | Redis-cached list |
| `GET /api/v1/notes/:id` | < 50ms | Single DB read |
| `GET /api/v1/graph/all` | < 300ms | pgvector + graph traversal |
| `GET /me/graph/cached` | < 30ms | Redis cache hit |
| `POST /embed` (NLP) | < 200ms | Model already warm |
| D3-force render (200 nodes) | < 16ms/frame | Canvas, not SVG |

---

## Redis Caching Patterns (go-redis/v9)

### Cache-Aside with invalidation
```go
// backend/internal/infrastructure/db/postgres/note_repo.go
const (
    notesCacheKey = "notes:all"
    notesCacheTTL = 5 * time.Minute
)

func (r *NoteRepository) FindAll(ctx context.Context) ([]*note.Note, error) {
    // 1. Check cache
    if r.redis != nil {
        cached, err := r.redis.Get(ctx, notesCacheKey).Bytes()
        if err == nil {
            var models []NoteModel
            if json.Unmarshal(cached, &models) == nil {
                return toDomainNotes(models), nil
            }
        }
    }
    // 2. Load from DB
    var models []NoteModel
    r.db.WithContext(ctx).Order("created_at DESC").Find(&models)
    // 3. Populate cache (store models, not domain entities — exported fields)
    if data, err := json.Marshal(models); err == nil {
        r.redis.Set(ctx, notesCacheKey, data, notesCacheTTL)
    }
    return toDomainNotes(models), nil
}

// Invalidate on mutation
func (r *NoteRepository) invalidateCache(ctx context.Context) {
    if r.redis != nil { r.redis.Del(ctx, notesCacheKey) }
}
```

### Graph cache with per-user key
```go
// backend/internal/application/cache/graph_cache.go
func (c *GraphCache) key(userID string) string { return "graph:" + userID }
// TTL: 5 minutes; invalidated on note/link mutation
```

### Redis key patterns (all prefixed):
```
notes:all              ← full notes list cache
graph:{userID}         ← user's graph data
auth:blacklist:{hash}  ← JWT blacklist
auth:refresh:{hash}    ← refresh token store
auth:pkce:{state}      ← OAuth PKCE parameters
auth:perm:{uid}:{res}:{action}  ← permission cache
```

---

## pgvector Query Optimization

```sql
-- Create IVFFlat index (run once, tune lists count = sqrt(row_count))
CREATE INDEX note_embeddings_ivfflat_idx
  ON note_embeddings
  USING ivfflat (embedding vector_cosine_ops)
  WITH (lists = 100);

-- Set probes for recall/speed tradeoff (session-level or per-query)
SET ivfflat.probes = 10;
```

```go
// Efficient cosine similarity search — uses index
r.db.WithContext(ctx).Raw(`
    SELECT e2.note_id,
           (1 - (e1.embedding <=> e2.embedding)) / 2.0 AS similarity
    FROM note_embeddings e1
    JOIN note_embeddings e2 ON e1.note_id != e2.note_id
    WHERE e1.note_id = ?
    ORDER BY similarity DESC
    LIMIT ?
`, noteID, limit).Scan(&results)
// See: backend/internal/infrastructure/db/postgres/embedding_repo.go
```

---

## Gin Request Timing Middleware

```go
// Add to backend/cmd/server/router.go
func timingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        latency := time.Since(start)
        // Log P95 violations
        if latency > 300*time.Millisecond {
            log.Printf("[SLOW] %s %s took %v", c.Request.Method, c.Request.URL.Path, latency)
        }
        c.Header("X-Response-Time", latency.String())
    }
}
```

---

## Connection Pool Configuration

```go
// backend/internal/infrastructure/db/db.go
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
sqlDB.SetConnMaxIdleTime(1 * time.Minute)
```

Redis pool:
```go
redis.NewClient(&redis.Options{
    Addr:            cfg.RedisURL,
    PoolSize:        10,
    ConnMaxIdleTime: time.Minute,      // v9 API
    ConnMaxLifetime: 5 * time.Minute,  // v9 API
})
```

---

## Async Task Offloading (asynq)

Heavy NLP operations are always offloaded to the worker to keep HTTP P95 low:

```go
// After creating a note in the handler — do NOT block the HTTP response
if err := q.EnqueueComputeEmbedding(ctx, note.ID().String()); err != nil {
    log.Printf("failed to enqueue embedding for %s: %v", note.ID(), err)
    // Non-fatal: note is created, embedding will retry
}
if err := q.EnqueueExtractKeywords(ctx, note.ID().String(), 10); err != nil {
    log.Printf("failed to enqueue keywords for %s: %v", note.ID(), err)
}
// Return 201 immediately — NLP processes asynchronously
```

See `backend/internal/application/common/task_queue.go` for the interface.
Worker entrypoint: `backend/cmd/worker/main.go`.

---

## HuggingFace Model Caching (NLP Service)

```
Timeline from container start:
  t=0s   uvicorn starts, FastAPI app loads (~1s)
  t=1s   container is "started" (healthcheck begins)
  t=1s   first /health call triggers ensure_model_loaded()
  t=15s  model loads from ./huggingface_cache (~15s if cache exists)
  t=15s  /health returns 200 {"status": "healthy", "model_loaded": true}
  t=600s healthcheck start_period expires (retries: 30 @ 30s interval)
```

Model is loaded once globally — subsequent `/embed` calls take ~5ms.
Never restart the NLP container in production without warming the model first.

---

## Frontend Bundle Optimization

```typescript
// vite.config.ts — code splitting
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          d3:     ['d3-force', 'd3-selection', 'd3-zoom'],
          three:  ['three'],
          vendor: ['ky']
        }
      }
    }
  }
});
```

- D3-force and Three.js are heavy — always in separate chunks.
- `GraphCanvas.svelte` imports from `./GraphCanvas/` modules (lazy tree-shaking).
- Canvas 2D for graph (not SVG) — handles 500+ nodes at 60fps.
- Sourcemaps stripped from production build (see `frontend/Dockerfile`).

---

## Anti-Patterns

```go
// ❌ N+1 query — loading each note's links separately
for _, note := range notes {
    links, _ := linkRepo.FindByNote(ctx, note.ID())  // one query per note
}
// ✅ Use batch query or eager load with GORM Preload

// ❌ Caching domain entities (unexported fields serialize as {})
json.Marshal(domainNote)  // returns "{}" — cache models, not entities

// ❌ Synchronous embedding generation in HTTP handler
embedding := nlpClient.Embed(ctx, note.Content())  // blocks for ~200ms
// ✅ Use asynq: EnqueueComputeEmbedding()

// ❌ Scanning all Redis keys with KEYS (blocks Redis)
r.redis.Do(ctx, "KEYS", "graph:*")
// ✅ Use SCAN with Iterator (see graph_cache.go InvalidateAll)
```
