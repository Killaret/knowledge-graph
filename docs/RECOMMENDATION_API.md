# Recommendation API Specification

## GET /api/v1/notes/{id}/suggestions

Returns precomputed recommendations for a specific note.

### Request

```http
GET /api/v1/notes/{note_id}/suggestions
Accept: application/json
```

### Response Logic

1. **Precomputed recommendations** — fast read from `note_recommendations`
2. **Freshness check** — compare `updated_at` of recommendations vs note
3. **Fallback to semantic neighbors** — if precomputed unavailable
4. **Fallback to Redis** — if semantic disabled
5. **HTTP 202 Accepted** — if nothing available, triggers background calculation

### Success Response (200 OK)

```json
{
  "note_id": "a0000000-0000-0000-0000-000000000001",
  "recommendations": [
    {
      "note_id": "a0000000-0000-0000-0000-000000000002",
      "title": "Related Note Title",
      "score": 0.95,
      "reason": "Common neighbors: 3"
    }
  ],
  "source": "precomputed",
  "generated_at": "2024-01-15T10:30:00Z"
}
```

### Response Headers

```http
X-Recommendations-Stale: true           # Data is stale
X-Recommendations-Source: semantic-fallback
X-Recommendations-Source: redis-fallback
X-Recommendations-Count: 5
```

### Fallback Response (202 Accepted)

Returned when no recommendations are available and background calculation is triggered:

```json
{
  "status": "calculating",
  "message": "Recommendations are being calculated in the background",
  "retry_after_seconds": 5
}
```

### Error Responses

#### 404 Not Found
```json
{
  "error": "note_not_found",
  "message": "Note with specified ID does not exist"
}
```

#### 500 Internal Server Error
```json
{
  "error": "recommendation_service_error",
  "message": "Failed to retrieve recommendations"
}
```

## Configuration

### Environment Variables

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
