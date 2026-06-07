# Recommendation API Specification

## GET /notes/{id}/suggestions

Returns precomputed recommendations for a specific note.

### Request

```http
GET /notes/{note_id}/suggestions?limit=20
Accept: application/json
```

**Query parameters:**

| Param | Default | Notes |
|-------|---------|-------|
| `limit` | `RECOMMENDATION_TOP_N` (20) | Clamped to 1..100; invalid values are ignored and the default is used |

### Response Logic

1. **Precomputed recommendations** — fast read from `note_recommendations`
2. **Freshness check** — compare `updated_at` of recommendations vs note
3. **Fallback to semantic neighbors** — if precomputed unavailable
4. **Fallback to Redis** — if semantic disabled
5. **HTTP 202 Accepted** — if nothing available, triggers background calculation

### Success Response (200 OK)

The response body is a flat list under `suggestions`. Each item has `note_id`, `title`, and `score`. Source/staleness are communicated via headers (see below), **not** body fields.

```json
{
  "suggestions": [
    {
      "note_id": "a0000000-0000-0000-0000-000000000002",
      "title": "Related Note Title",
      "score": 0.95
    }
  ],
  "generated_at": "2024-01-15T10:30:00Z"
}
```

> Source shape: `backend/internal/interfaces/api/notehandler/note_handler.go` (`SuggestionsResponse` / `Suggestion`).

### Response Headers

```http
X-Recommendations-Stale: true                 # Precomputed data is stale; a refresh was enqueued
X-Recommendations-Source: semantic-fallback   # Result came from the pgvector semantic fallback
X-Recommendations-Source: redis-fallback      # Result came from the Redis cache fallback
```

No `X-Recommendations-Source` header is set when the result comes from the precomputed `note_recommendations` table.

### Pending Response (202 Accepted)

Returned when no recommendations are available yet. A background calculation is enqueued and the body is an **empty suggestions list** (same schema as the 200 response):

```json
{
  "suggestions": []
}
```

`X-Recommendations-Stale: true` is also set. Clients should retry shortly after.

> Planned change (Roadmap P0-04, not yet implemented): add a `status: "ready" | "pending"` field to the body. Until then, distinguish pending from ready by the **202 vs 200** status code.

### Error Responses

#### 400 Bad Request

Returned when the note ID is not a valid UUID:

```json
{
  "error": "invalid id"
}
```

> Note: this endpoint does not return `404` for a non-existent note — an unknown ID falls through to the `202` pending path with an empty list.

## Configuration

### Environment Variables

> Canonical reference for all environment variables: [`CONFIGURATION_EN.md`](CONFIGURATION_EN.md). The subset below is the recommendation-specific config.

```bash
# Number of recommendations to return
RECOMMENDATION_TOP_N=20

# Task delay for deduplication (seconds)
RECOMMENDATION_TASK_DELAY_SECONDS=5

# Redis fallback
RECOMMENDATION_FALLBACK_ENABLED=true
RECOMMENDATION_FALLBACK_TTL_SECONDS=3600

# Semantic neighbors fallback
RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED=true

# Asynq settings
ASYNQ_CONCURRENCY=10
ASYNQ_QUEUE_DEFAULT=1
ASYNQ_QUEUE_MAX_LEN=10000
```

## Initial Population

### CLI Command

**File:** `backend/cmd/cli/main.go`

```bash
# Build
cd backend
go build -o bin/recommendation-cli ./cmd/cli

# Run (dry-run for testing)
./bin/recommendation-cli --dry-run

# Actual run
./bin/recommendation-cli

# For large databases (>1000 notes), increase delay
./bin/recommendation-cli --batch-delay=60
```

## Migration from Existing System

1. Apply migration:
   ```bash
   psql -d knowledge_base -f backend/migrations/013_create_note_recommendations.up.sql
   ```

2. Run CLI for initial population:
   ```bash
   ./bin/recommendation-cli
   ```

3. Wait for queue completion (monitor via asynqmon)

4. Switch API to new handler (already done in code)

5. Remove old Redis cache if needed:
   ```bash
   redis-cli --scan --pattern "recommendations:*" | xargs redis-cli del
   ```
