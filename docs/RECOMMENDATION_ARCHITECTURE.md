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
// - UniqueKey("rec:{note_id}")    // Deduplication
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
2. **Task deduplication** — `UniqueKey` prevents duplicates
3. **Cascade limiting** — `reverseCascadeDepth = 1` prevents queue explosion
4. **Transactionality** — atomic update via `SaveBatch` + `DeleteNotInBatch`

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
