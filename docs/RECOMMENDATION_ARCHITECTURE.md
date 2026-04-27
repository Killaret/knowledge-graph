# Recommendation System Architecture

## Overview

The Knowledge Graph recommendation system uses **asynchronous event-driven precomputation**. Instead of synchronous calculation on every request, recommendations are computed in the background and stored in the `note_recommendations` table.

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Create/Update  │────▶│  Asynq Queue     │────▶│  Worker         │
│  Note/Link      │     │  (Redis)         │     │  (RefreshSvc)   │
└─────────────────┘     └──────────────────┘     └────────┬────────┘
                                                         │
                                                         ▼
                                              ┌────────────────────┐
                                              │  note_recommendations│
                                              │  (PostgreSQL)        │
                                              └──────────┬─────────┘
                                                         │
                                                         ▼
                                              ┌────────────────────┐
                                              │  GET /suggestions   │
                                              │  (fast read)        │
                                              └────────────────────┘
```

## Core Components

### 1. Precomputed Recommendations Table

**Migration:** `backend/migrations/013_create_note_recommendations.up.sql`

```sql
CREATE TABLE note_recommendations (
    note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    recommended_note_id UUID NOT NULL REFERENCES notes(id) ON DELETE CASCADE,
    score REAL NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (note_id, recommended_note_id)
);
```

### 2. Refresh Service

**File:** `backend/internal/application/recommendation/refresh_service.go`

Atomically updates recommendations in a transaction:
- Gets recommendations via `TraversalService.GetSuggestions`
- Saves via `SaveBatch` (UPSERT)
- Removes stale via `DeleteNotInBatch`

### 3. Asynq Tasks

**File:** `backend/internal/infrastructure/queue/tasks/recommendation.go`

```go
const TypeRefreshRecommendations = "recommendation:refresh"

// Task options:
// - MaxRetry(3)                    // 3 retries
// - Timeout(30s)                   // Timeout
// - ProcessIn(delay)              // Delay (dedup)
// - TaskID("rec:{note_id}")       // Deduplication (asynq.TaskID)
```

### 4. Event Logic

**Files:**
- `backend/internal/interfaces/api/notehandler/note_handler.go`
- `backend/internal/interfaces/api/linkhandler/link_handler.go`

Automatically queues tasks on note/link create/update/delete.

### 5. Affected Notes Determination

**File:** `backend/internal/application/recommendation/affected_notes.go`

```go
const reverseCascadeDepth = 1  // Cascade limit

func GetAffectedNotes(targetNoteID) []uuid.UUID {
    // 1. The note itself
    // 2. Notes that recommend it (direct only)
}
```

## Performance Comparison

| Approach | Latency | DB Load | Scalability |
|----------|---------|---------|-------------|
| Sync BFS | 50-200ms | High | Limited |
| Async (current) | 5-10ms | Low | Excellent |

## Optimizations

1. **Batch neighbor loading** — `GetNeighborsBatch` reduces SQL queries
2. **Task deduplication** — `TaskID` prevents duplicates (unique task ID)
3. **Cascade limiting** — `reverseCascadeDepth = 1` prevents queue explosion
4. **Transactionality** — atomic update via `SaveBatch` + `DeleteNotInBatch`

## API Response Headers

The `/suggestions` endpoint returns `X-Recommendations-*` headers to indicate data source and freshness:

| Header | Value | Meaning |
|--------|-------|---------|
| `X-Recommendations-Source` | `table` | Data from `note_recommendations` table (precomputed) |
| `X-Recommendations-Source` | `semantic` | Fallback to pgvector semantic similarity |
| `X-Recommendations-Source` | `redis` | Fallback to Redis cache |
| `X-Recommendations-Source` | `empty` | No data available, background task triggered |
| `X-Recommendations-Stale` | `true` | Data may be outdated (fallback sources always stale) |

**Note:** `X-Recommendations-Stale` is only set when data is stale or from fallback sources. Fresh precomputed data has no `Stale` header.

## Migration to Pure Precomputed Scores

### Current State (Transition Period)

The `/suggestions` API currently has 4 fallback levels:

1. **Table `note_recommendations`** — precomputed data (target state)
2. **Semantic fallback** — fast pgvector query (to be removed)
3. **Redis cache** — old synchronous results (to be removed)
4. **Empty list + 202 Accepted** — background computation triggered

### Target Architecture (Pure Precomputed)

```
┌─────────────────┐     ┌─────────────────────┐
│  GET /suggestions    │  SELECT FROM note_  │
│                      │  recommendations    │
│  Response:           │  (indexed, fast)    │
│  - suggestions[]     │                     │
│  - status: "ready"   │                     │
│    | "pending"       │                     │
└─────────────────┘     └─────────────────────┘
```

### Configuration

```bash
# .env file
RECOMMENDATION_FALLBACK_ENABLED=false  # Disable all synchronous fallbacks
```

When `false`:
- API reads **only** from `note_recommendations` table
- No pgvector queries on request path
- No Redis cache lookups for recommendations
- New notes return empty list until worker completes

### Pros and Cons

| Pros | Cons |
|------|------|
| Maximum performance: single indexed SELECT | New notes temporarily without recommendations |
| Code simplicity: no fallback branches | Requires reliable worker operation |
| Predictable behavior: always consistent with last computation | Queue backlog delays updates |
| No embedding repo calls on API path | |

### Handling New Notes

When no recommendations exist (new note):

```json
{
  "suggestions": [],
  "status": "pending",
  "message": "Рекомендации скоро появятся"
}
```

Frontend should show: **"Рекомендации рассчитываются..."**

Background worker is triggered immediately via Asynq task.

### Rollback Strategy

Keep `RECOMMENDATION_FALLBACK_ENABLED` toggle for emergency rollback:

```bash
# If queue is stuck or recommendations stale
RECOMMENDATION_FALLBACK_ENABLED=true  # Re-enable fallbacks
```

Future cleanup: After system proves reliability (30+ days stable), fallback code can be permanently removed.

### Pre-Migration Checklist

- [ ] CLI filled `note_recommendations` for all existing notes
- [ ] Asynq queue monitoring configured (`asynqmon` deployed)
- [ ] Worker error alerting enabled
- [ ] Frontend handles `status: "pending"` gracefully
- [ ] Fallback toggle tested in staging

## Monitoring

### Metrics to Track

1. **Asynq queue length:**
   ```bash
   redis-cli LLEN asynq:{default}
   ```

2. **Pending tasks count:**
   ```bash
   redis-cli ZCARD asynq:scheduled
   ```

3. **Worker errors:**
   ```bash
   grep "failed to refresh" /var/log/worker.log
   ```

### asynqmon Web Interface

Recommended to deploy `asynqmon` for visual monitoring:

```bash
docker run -p 8080:8080 hibiken/asynqmon --redis-addr=localhost:6379
```
