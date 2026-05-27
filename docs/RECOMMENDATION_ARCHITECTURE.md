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

## Keyword Similarity Component

### Overview

The recommendation system includes a flexible keyword similarity component that can be configured to use different similarity strategies when calculating recommendations. This component is integrated into the `TraversalService` and contributes to the final recommendation score based on the `gamma` weight parameter.

### Architecture

**Domain Layer:**
- `backend/internal/domain/graph/keyword_matcher.go` — Interface for keyword-based similarity matching

**Application Layer:**
- `backend/internal/application/recommendation/keyword_similarity.go` — Similarity strategy implementations
- `backend/internal/application/recommendation/keyword_matcher_impl.go` — KeywordMatcher implementation using similarity strategies

**Integration Point:**
- `backend/cmd/worker/main.go` — Worker sets up KeywordMatcher in TraversalService

### Configuration

Configuration is done through `knowledge-graph.config.json` in the `backend.recommendation` section:

```json
{
  "backend": {
    "recommendation": {
      "keyword_similarity_method": "jaccard",
      "keyword_tversky_alpha": 0.5,
      "keyword_tversky_beta": 0.5,
      "gamma": 0.2
    }
  }
}
```

**Parameters:**
- `keyword_similarity_method` — Similarity strategy: `jaccard`, `overlap`, `tversky`, `weighted_jaccard`, `cosine` (default: `jaccard`)
- `keyword_tversky_alpha` — Alpha parameter for Tversky index (default: 0.5)
- `keyword_tversky_beta` — Beta parameter for Tversky index (default: 0.5)
- `gamma` — Weight of keyword component in final score (set > 0 to enable keyword similarity)

### Available Strategies

| Strategy | Formula | Description | Requires Weights |
|----------|---------|-------------|------------------|
| **Jaccard** | |A ∩ B| / |A ∪ B| | Classic Jaccard coefficient | No |
| **Overlap** | |A ∩ B| / min(|A|, |B|) | Overlap coefficient | No |
| **Tversky** | |A ∩ B| / (|A ∩ B| + α|AB| + β|BA|) | Tversky index with parameters | No |
| **Weighted Jaccard** | sum(min(w1, w2)) / sum(max(w1, w2)) | Weighted Jaccard using keyword weights | Yes |
| **Cosine** | (A · B) / (|A| × |B|) | Cosine similarity of weight vectors | Yes |

**Notes:**
- When `gamma = 0`, the keyword component is disabled (NoOpKeywordMatcher is used)
- When weights are not available for weighted strategies, they fall back to unit weights (cosine) or regular Jaccard (weighted_jaccard)
- Tversky with α = β = 0.5 is equivalent to symmetric Dice coefficient
- Tversky with α = β = 1 is equivalent to Jaccard coefficient

### Integration Flow

```
1. Worker Initialization (cmd/worker/main.go)
   ↓
   Load config (keyword_similarity_method, tversky_alpha, tversky_beta)
   ↓
   Create KeywordSimilarity strategy via NewKeywordSimilarity()
   ↓
   Create KeywordMatcherImpl with KeywordRepository and strategy
   ↓
   Create TraversalServiceWithWeights(alpha, beta, gamma)
   ↓
   If gamma > 0: SetKeywordMatcher(keywordMatcher)

2. Recommendation Calculation (TraversalService.GetSuggestions)
   ↓
   Run BFS to get graph component scores
   ↓
   If gamma > 0: Call keywordMatcher.Match(sourceID, candidateIDs)
   ↓
   KeywordMatcher fetches keywords via GetKeywordsBatchWithWeights()
   ↓
   For each candidate: Calculate similarity using configured strategy
   ↓
   Aggregate: score = α×graph + β×semantic + γ×keyword
   ↓
   Return top N suggestions
```

### Component Weights

The recommendation score is a weighted combination of three components:

```
total_score = α × graph_score + β × semantic_score + γ × keyword_score
```

Where:
- `α` (alpha) — Graph traversal component weight
- `β` (beta) — Semantic similarity component weight
- `γ` (gamma) — Keyword similarity component weight

Weights are automatically normalized to sum to 1.0. Set `gamma > 0` to enable keyword similarity in recommendations.


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
