# Knowledge Graph Configuration

System parameters can be configured via:
1. **`config/*.json` source files** — Editable sections that are merged into `knowledge-graph.config.json`
2. **`knowledge-graph.config.json`** — Generated runtime configuration file (recommended)
3. **Environment variables** — Override specific values
4. **`.env` file** — For local development

For production, pass variables through `environment` in `docker-compose.yml` or `ConfigMap`/`Secrets` in Kubernetes.

🌐 **Language:** English

---

## Unified Configuration File

The `knowledge-graph.config.json` file in the project root is the **generated runtime artifact** and the single source of truth for all structural parameters. It is created from editable source files in `config/*.json`, and contains settings for backend, frontend, NLP service, backup, and CI/CD.

To regenerate the runtime config after editing source files, run:
```bash
npm run build-config
```

### Priority (highest to lowest)

1. **Environment variables** — Override any value from JSON
2. **`knowledge-graph.config.json`** — Generated runtime config shared across all components
3. **Hardcoded defaults** — Fallback in Go/TypeScript code

Note: edit the source files in `config/*.json` and regenerate `knowledge-graph.config.json` with `npm run build-config`.

### File Structure

```json
{
  "backend": {
    "server": { "rate_limit": { ... } },
    "database": { ... },
    "search": { ... },
    "recommendation": { ... },
    "pagination": { ... },
    "graph": { ... },
    "embedding": { ... },
    "asynq": { ... }
  },
  "frontend": {
    "test": { ... },
    "graph": { "2d": { ... }, "3d": { ... } },
    "api": { ... },
    "achievements": { ... }
  },
  "ci_cd": {
    "integration_test": { ... }
  },
  "nlp": { ... },
  "backup": { ... }
}
```

### Frontend Usage (TypeScript)

```typescript
import { graphConfig2D, apiConfig, testConfig, ACHIEVEMENT_POLL_INTERVAL_MS } from '$lib/config';

// Use centralized config
const enableShadows = nodes.length < graphConfig2D.shadows_threshold;
const limit = apiConfig.default_limit;
const pollInterval = ACHIEVEMENT_POLL_INTERVAL_MS;
```

### Backend Usage (Go)

```go
import "knowledge-graph/internal/config"

cfg := config.Load()
// cfg.ServerRateLimitEnabled
// cfg.RecommendationDepth
// cfg.GraphDefaultLimit
```

---

## Galactic Lexicon & Achievements Configuration

### User Settings (Database)

The Galactic Lexicon and Achievements system uses user-specific settings stored in the `user_settings` table:

| Setting Key | Type | Default | Description |
|--------------|------|---------|-------------|
| `galactic_mode` | boolean | `false` | Enable galactic (space-themed) messaging mode |
| `show_achievement_notifications` | boolean | `true` | Show toast notifications for unlocked achievements |
| `preferred_language` | string | `ru` | User's preferred language (ru/en) |

These settings can be updated via the user settings API and are respected by the frontend components.

### Frontend Configuration (`frontend.achievements`)

```json
{
  "frontend": {
    "achievements": {
      "poll_interval_ms": 7000
    }
  }
}
```

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `poll_interval_ms` | integer | `7000` | Polling interval for checking new achievements (milliseconds) |

### Environment Variable Overrides

| Variable | Description | Default |
|----------|-------------|---------|
| `FRONTEND_ACHIEVEMENTS_POLL_INTERVAL_MS` | Achievement polling interval | `7000` |

### Achievement Condition Types

Achievements use JSON-based conditions stored in the `condition_json` field:

**Count-based conditions:**
```json
{
  "type": "count",
  "entity": "note",
  "action": "create",
  "filter": { "type": "star" },
  "threshold": 10
}
```

**Streak-based conditions:**
```json
{
  "type": "streak",
  "days": 7
}
```

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/achievements` | GET | List all available achievements |
| `/api/v1/users/me/achievements` | GET | Get current user's achievements |
| `/api/v1/users/me/achievements/:id/mark-seen` | POST | Mark achievement notification as seen |

---

## Backup Configuration

### JSON Configuration (`backup`)

```json
{
  "backup": {
    "local_path": "./backups",
    "cloud": {
      "enabled": false,
      "provider": "r2",
      "r2": {
        "account_id": "",
        "access_key_id": "",
        "secret_access_key": "",
        "bucket": "",
        "region": "auto"
      }
    },
    "schedule": "0 2 * * *",
    "retention_days": 7,
    "draft_ttl_hours": 168
  }
}
```

### Environment Variable Overrides

| Variable | Description | Default |
|----------|-------------|---------|
| `BACKUP_LOCAL_PATH` | Local backup directory | `./backups` |
| `BACKUP_CLOUD_ENABLED` | Enable cloud backup | `false` |
| `BACKUP_CLOUD_PROVIDER` | Cloud provider (yandex) | `yandex` |
| `BACKUP_YANDEX_OAUTH_TOKEN` | Yandex.Disk OAuth token | - |
| `BACKUP_YANDEX_FOLDER` | Yandex.Disk backup folder | `/KnowledgeGraphBackups` |
| `BACKUP_YANDEX_MAX_BACKUPS` | Maximum number of backups to keep | `10` |
| `BACKUP_SCHEDULE` | Cron schedule for backups | `0 2 * * *` |
| `BACKUP_RETENTION_DAYS` | Backup retention period | `7` |
| `BACKUP_DRAFT_TTL_HOURS` | Draft TTL in MongoDB | `168` |

### Backup Scripts

**Linux/Mac:** `scripts/utility/backup-personal.sh`
- Performs `pg_dump` to `backups/backup-personal-YYYY-MM-DD.sql.gz`
- Uploads to Yandex.Disk via WebDAV if cloud backup enabled
- Cleans up old backups older than retention days

**Windows:** `scripts/utility/backup-personal.ps1`
- Same functionality for Windows

---

**Backup Service (Docker):** `backup_scheduler` in `docker-compose.personal.yml`
- Runs backup scripts automatically every 24 hours
- Supports both local and cloud (Yandex.Disk) backup

**Go Service:** `backend/internal/infrastructure/cloud/yandex_backup.go`
- YandexBackupService for Yandex.Disk WebDAV integration
- Methods: UploadBackup, DownloadBackup, ListBackups, DeleteBackup
- Retry logic (3 attempts) for failed uploads
- Automatic cleanup of old cloud backups (max_backups, default: 10)

**For detailed backup setup guide, see:** [`docs/BACKUP.md`](docs/BACKUP.md)

---

## MongoDB Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MONGO_URL` | MongoDB connection string | `mongodb://localhost:27017` |
| `MONGO_DATABASE` | MongoDB database name | `knowledge_graph` |

### MongoDB Usage

MongoDB is used for storing note drafts with the following features:
- **State Pattern**: Active, Publishing, Published, Conflict states
- **TTL Index**: Automatic cleanup of expired drafts
- **Draft Sync**: Asynchronous synchronization with PostgreSQL notes

---

## Required Environment Variables

These must be set via environment variables (not in JSON):

| Variable | Component | Description | Example |
|----------|-----------|-------------|---------|
| `DATABASE_URL` | backend, worker | PostgreSQL connection string | `postgresql://kb_user:kb_pass@postgres:5432/knowledge_base?sslmode=disable` |

All other parameters can be configured via `knowledge-graph.config.json` or overridden via environment variables.

---

## Infrastructure Parameters

| Variable | Component | Description | Default |
|----------|-----------|-------------|---------|
| `SERVER_PORT` | backend | HTTP server port (Gin) | `8080` |
| `REDIS_URL` | backend, worker | Redis address for asynq queues and recommendation cache | `localhost:6379` |
| `NLP_SERVICE_URL` | backend, worker | Python NLP service URL | `http://localhost:5000` |

### Component Details

**SERVER_PORT** — Port where backend listens for HTTP requests
- In Docker Compose: `8080` (inside container)
- Exposed outside: `8080:8080`

**REDIS_URL** — Storage for:
- Asynq task queues (worker reads from here)
- Recommendation caching (TTL configured separately)

**NLP_SERVICE_URL** — Address for HTTP calls:
- `/extract_keywords` — keyword extraction
- `/embeddings` — vector embeddings generation
- `/health` — service health check

---

## Server & Rate Limiting

### JSON Configuration (`backend.server`)

```json
{
  "backend": {
    "server": {
      "rate_limit": {
        "enabled": true,
        "requests": 100,
        "window_seconds": 60,
        "endpoints": {
          "notes_create": 30,
          "links_create": 50,
          "notes_update": 20
        },
        "fallback_ports": ["8081", "8082"]
      }
    }
  }
}
```

### Environment Variable Overrides

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_RATE_LIMIT_ENABLED` | Enable rate limiting | `true` |
| `SERVER_RATE_LIMIT_REQUESTS` | General request limit | `100` |
| `SERVER_RATE_LIMIT_WINDOW_SECONDS` | Time window | `60` |
| `SERVER_PORT` | HTTP server port | `8080` |
| `SERVER_FALLBACK_PORTS` | Backup ports (comma-separated) | `8081,8082` |

### Rate Limiting Behavior

- **General requests**: All GET requests and unspecified endpoints
- **Write operations**: Stricter limits for POST/PUT/DELETE to prevent abuse
- **IP-based**: Limits are applied per client IP address
- **Response**: When exceeded, API returns `429 Too Many Requests`

---

## Recommendation Algorithm

### JSON Configuration (`backend.recommendation`)

```json
{
  "backend": {
    "recommendation": {
      "depth": 3,
      "decay": 0.5,
      "top_n": 20,
      "alpha": 0.5,
      "beta": 0.5,
      "gamma": 0.2,
      "cache_ttl_seconds": 300,
      "task_delay_seconds": 5,
      "batch_rate_limit": 10,
      "fallback_enabled": true,
      "fallback_ttl_seconds": 3600,
      "fallback_semantic_enabled": true,
      "keyword_enabled": true,
      "bfs_aggregation": "max",
      "bfs_normalize": true
    }
  }
}
```

### Environment Variable Overrides

| Variable | Description | Default | Range |
|----------|-------------|---------|-------|
| `RECOMMENDATION_ALPHA` | Weight of explicit links | `0.5` | 0.0 - 1.0 |
| `RECOMMENDATION_BETA` | Weight of semantic similarity | `0.5` | 0.0 - 1.0 |
| `RECOMMENDATION_DEPTH` | BFS traversal depth | `3` | 1 - 5 |
| `RECOMMENDATION_DECAY` | Weight decay for indirect links | `0.5` | 0.0 - 1.0 |
| `RECOMMENDATION_CACHE_TTL_SECONDS` | Cache TTL | `300` | 60 - 3600 |
| `EMBEDDING_SIMILARITY_LIMIT` | pgvector candidates limit | `30` | 10 - 100 |
| `RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED` | Enable semantic fallback | `true` | - |
| `RECOMMENDATION_KEYWORD_ENABLED` | Enable keyword component (gamma) | `true` | - |

### Detailed Description

**ALPHA + BETA** — Combination formula:
```
score = α × explicit_score + β × semantic_score
```
- `α + β` doesn't have to equal 1, but recommended for consistent scale
- `α = 1.0, β = 0.0` — explicit links only (ignore semantics)
- `α = 0.0, β = 1.0` — semantics only (ignore links)
- `α = 0.5, β = 0.5` — balanced: equal weight for links and semantics (default)

**DEPTH** — BFS Depth:
- `1` — direct links only (fast)
- `3` — links up to 3rd level (optimal)
- `5` — deep search (slow, lots of data)

**DECAY** — Weight decay:
- Applied to links starting from level 2
- Formula: `weight × decay^(depth-1)`
- `0.5` means: level 2 = 50%, level 3 = 25%

**CACHE_TTL** — Caching time:
- Recommendations cached in Redis for speed
- Cache key: `suggestions:{note_id}:{limit}`

**EMBEDDING_SIMILARITY_LIMIT** — Semantic candidates limit:
- How many notes to fetch from pgvector by embedding similarity
- Higher = more accurate, but slower

**RECOMMENDATION_FALLBACK_ENABLED** — Synchronous fallback toggle:
- `false` (default): Use only precomputed recommendations from `note_recommendations` table. Fastest, but new notes may temporarily have no recommendations.
- `true`: Enable synchronous pgvector and Redis cache fallbacks. Slower, but always returns results.

When `false`:
- API reads only from `note_recommendations` (single indexed SELECT)
- No pgvector queries on request path
- No Redis cache lookups for recommendations
- New notes return empty list until async worker completes

Recommended: Set `false` for production after verifying worker reliability. Use `true` only as emergency rollback if queue issues occur.

See also: [RECOMMENDATION_ARCHITECTURE.md](RECOMMENDATION_ARCHITECTURE.md#migration-to-pure-precomputed-scores)

---

## Graph Visualization & API

### JSON Configuration (`backend.graph`, `frontend.graph`)

```json
{
  "backend": {
    "graph": {
      "load_depth": 2,
      "max_nodes": 500,
      "default_limit": 100,
      "max_limit": 1000,
      "link_default_limit": 500,
      "link_max_limit": 5000
    }
  },
  "frontend": {
    "graph": {
      "2d": { "max_nodes": 500, "shadows_threshold": 100 },
      "3d": { "max_nodes": 500 }
    }
  }
}
```

### Environment Variable Overrides

| Variable | Description | Default |
|----------|-------------|---------|
| `GRAPH_LOAD_DEPTH` | Graph loading depth | `2` |
| `GRAPH_DEFAULT_LIMIT` | Default `/graph/all` limit | `100` |
| `GRAPH_MAX_LIMIT` | Max `/graph/all` limit | `1000` |
| `GRAPH_LINK_DEFAULT_LIMIT` | Default link limit | `500` |
| `GRAPH_LINK_MAX_LIMIT` | Max link limit | `5000` |

---

## Pagination

### JSON Configuration (`backend.pagination`)

```json
{
  "backend": {
    "pagination": {
      "default_limit": 20,
      "max_limit": 100
    }
  }
}
```

Used by `List` and `Search` endpoints for note pagination.

## Database

### JSON Configuration (`backend.database`)

```json
{
  "backend": {
    "database": {
      "retry_max_attempts": 3,
      "retry_delay_seconds": 5,
      "migrations_fail_on_error": false
    }
  }
}
```

## Search

### JSON Configuration (`backend.search`)

```json
{
  "backend": {
    "search": {
      "fulltext_languages": ["russian", "simple"],
      "ranking_weights": { "russian": 1.0, "simple": 1.0 },
      "fallback_to_ilike": true
    }
  }
}
```

## Advanced Parameters (BFS + Asynq)

These parameters are now fully integrated and loaded from `knowledge-graph.config.json`:

| JSON Path | Description | Default |
|-----------|-------------|---------|
| `backend.recommendation.gamma` | Coefficient for third component | `0.2` |
| `backend.recommendation.bfs_aggregation` | Weight aggregation: `max`, `sum`, `avg` | `max` |
| `backend.recommendation.bfs_normalize` | Normalize link weights | `true` |
| `backend.asynq.concurrency` | Asynq concurrency level | `10` |
| `backend.asynq.queue_default` | Default queue priority | `1` |
| `backend.asynq.queue_max_len` | Max queue length | `10000` |
| `backend.recommendation.keyword_enabled` | Enable keyword component (gamma) | `true` |

### Reserved Parameters Description

**RECOMMENDATION_GAMMA** — Coefficient for third component:
- Allows adding third factor to calculation (e.g., time factor or popularity)
- `0.2` — small contribution from additional factor
- Requires `keyword_enabled: true` to be active

**RECOMMENDATION_KEYWORD_ENABLED** — Enable keyword component (gamma):
- `true` — keyword component is included in scoring
- `false` — keyword component disabled (fallback to alpha/beta only)
- Allows gradual rollout or emergency disabling of keyword features

**BFS_AGGREGATION** — How to aggregate weights during graph traversal:
- `max` — use maximum path weight (recommended, noise-resistant)
- `sum` — sum weights of all paths (more aggressive scoring)
- `avg` — average value (smooths outliers)

**BFS_NORMALIZE** — Weight normalization:
- `true` — weights normalized to [0, 1] range
- `false` — use raw weights

**ASYNQ_CONCURRENCY** — Worker parallelism:
- How many tasks processed simultaneously
- Higher = faster, but higher CPU load

**ASYNQ_QUEUE_DEFAULT** — Queue priority:
- Processing priority for default task queue
- Higher value = higher priority

---

## Minimal `.env` File (Recommended)

Only database URL and optional overrides:

```env
# Required
DATABASE_URL=postgresql://kb_user:kb_password@postgres:5432/knowledge_base?sslmode=disable

# Optional overrides (rest is in knowledge-graph.config.json)
SERVER_PORT=8080
REDIS_URL=redis:6379
NLP_SERVICE_URL=http://nlp:5000
```

## Full Environment Variable Reference

For cases where you need to override JSON values via environment:

```env
# Server
SERVER_PORT=8080
SERVER_RATE_LIMIT_ENABLED=true
SERVER_RATE_LIMIT_REQUESTS=100
SERVER_RATE_LIMIT_WINDOW_SECONDS=60
SERVER_FALLBACK_PORTS=8081,8082

# Database
DATABASE_RETRY_MAX_ATTEMPTS=3
DATABASE_RETRY_DELAY_SECONDS=5
MIGRATIONS_FAIL_ON_ERROR=false

# Search
SEARCH_FALLBACK_TO_ILIKE=true

# Recommendations
RECOMMENDATION_ALPHA=0.5
RECOMMENDATION_BETA=0.5
RECOMMENDATION_GAMMA=0.2
RECOMMENDATION_DEPTH=3
RECOMMENDATION_DECAY=0.5
RECOMMENDATION_TOP_N=20
RECOMMENDATION_CACHE_TTL_SECONDS=300
RECOMMENDATION_TASK_DELAY_SECONDS=5
RECOMMENDATION_BATCH_RATE_LIMIT=10
RECOMMENDATION_FALLBACK_ENABLED=true
RECOMMENDATION_FALLBACK_TTL_SECONDS=3600
RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED=true
RECOMMENDATION_KEYWORD_ENABLED=true
BFS_AGGREGATION=max
BFS_NORMALIZE=true

# Pagination
PAGINATION_DEFAULT_LIMIT=20
PAGINATION_MAX_LIMIT=100

# Graph API
GRAPH_LOAD_DEPTH=2
GRAPH_DEFAULT_LIMIT=100
GRAPH_MAX_LIMIT=1000
GRAPH_LINK_DEFAULT_LIMIT=500
GRAPH_LINK_MAX_LIMIT=5000

# Embedding
EMBEDDING_SIMILARITY_LIMIT=30

# Asynq
ASYNQ_CONCURRENCY=10
ASYNQ_QUEUE_DEFAULT=1
ASYNQ_QUEUE_MAX_LEN=10000
```

## Configuration Examples

### 1. Explicit Links Only (no embeddings)
```env
RECOMMENDATION_ALPHA=1.0
RECOMMENDATION_BETA=0.0
RECOMMENDATION_DEPTH=3
RECOMMENDATION_DECAY=0.5
```

### 2. Semantic Similarity Only (ignore explicit links)
```env
RECOMMENDATION_ALPHA=0.0
RECOMMENDATION_BETA=1.0
```

### 3. Balanced Mode (50/50)
```env
RECOMMENDATION_ALPHA=0.5
RECOMMENDATION_BETA=0.5
```

### 4. Deep Graph Traversal (careful, may be slow)
```env
RECOMMENDATION_DEPTH=5
RECOMMENDATION_DECAY=0.3
```

### 5. Extended Recommendation Cache (10 minutes)
```env
RECOMMENDATION_CACHE_TTL_SECONDS=600
```

### 6. More Embedding Candidates
```env
EMBEDDING_SIMILARITY_LIMIT=50
```

### 7. High Sensitivity to Embeddings (α=0.3, β=0.7)
```env
RECOMMENDATION_ALPHA=0.3
RECOMMENDATION_BETA=0.7
```

### 8. Deep Graph Visualization (more neighbor levels)
```env
GRAPH_LOAD_DEPTH=4
```

### 9. Minimal Graph Visualization (direct links only)
```env
GRAPH_LOAD_DEPTH=1
```

### 10. Aggressive BFS (sum aggregation, no normalization)
```env
BFS_AGGREGATION=sum
BFS_NORMALIZE=false
```

### 11. Conservative BFS (avg aggregation, with normalization)
```env
BFS_AGGREGATION=avg
BFS_NORMALIZE=true
```

### 12. High-Performance Worker
```env
ASYNQ_CONCURRENCY=20
ASYNQ_QUEUE_DEFAULT=2
```

### 13. Resource-Efficient Worker (minimum resources)
```env
ASYNQ_CONCURRENCY=2
ASYNQ_QUEUE_DEFAULT=1
```

### 14. Pure Precomputed Mode (fastest, no fallbacks)
```env
RECOMMENDATION_FALLBACK_ENABLED=false
```

### 15. Fallback Mode (slower, always returns results)
```env
RECOMMENDATION_FALLBACK_ENABLED=true
```

### 16. Disable Keyword Component (gamma)
```env
RECOMMENDATION_KEYWORD_ENABLED=false
```

### How to Apply Changes

#### After editing `.env` or variables in `docker-compose.yml`, restart backend:

```bash
docker-compose restart backend
```

#### Check logs — you should see configuration loaded message:

```bash
docker logs kg-backend --tail 30 | grep "Config loaded"
```

#### Example output:

```
Config loaded: alpha=0.50, beta=0.50, depth=3, decay=0.50, cacheTTL=5m0s, embeddingLimit=30, graphLoadDepth=2
```

### Final Recommendation Score Formula

#### For source note A and candidate C:

```
score = α * explicit_weight(A, C) + β * content_sim(A, C)
explicit_weight(A, C) — sum of weights of all explicit links from A to C (with decay for indirect paths starting from level 2).

content_sim(A, C) — cosine similarity of embeddings, normalized to [0,1] range.

α + β doesn't have to equal 1, but sum = 1 is recommended for consistent scale.
```

### Notes

#### Embeddings are computed asynchronously by worker. Ensure worker (kg-worker) is running and has processed tasks for notes, otherwise content_sim will be 0.

#### If beta > 0, note_embeddings table must have vectors for all notes. If not, recommendations will rely only on explicit links.

#### Changing RECOMMENDATION_DEPTH above 3 may significantly increase database load and response time.

#### Recommendation cache invalidates by TTL. For immediate clearing, restart backend or manually delete keys:

```bash
docker exec kg-redis redis-cli KEYS "suggestions:*" | xargs docker exec kg-redis redis-cli DEL
```

---

## Related Documentation

## Graph Service (gRPC Microservice)

The `graph-service` is a separate gRPC microservice for layout computation and graph streaming. It has its own configuration section in `knowledge-graph.config.json`.

### JSON Configuration (`graph_service`)

```json
{
  "graph_service": {
    "grpc_port": "9090",
    "http_port": "9091",
    "full_limit": 1000,
    "default_depth": 2,
    "event_channel": "graph:events",
    "cache": {
      "note_layout_ttl_seconds": 300,
      "full_layout_ttl_seconds": 300,
      "delta_ttl_seconds": 60
    },
    "layout": {
      "2d_radius": 100.0,
      "3d_radius": 120.0,
      "3d_z_step": 5.0,
      "default_node_size": 1.0
    },
    "stream_chunk_size": 100,
    "event_tracking_ttl_hours": 24,
    "unprocessed_event_check_interval_minutes": 5
  }
}
```

### Environment Variable Overrides

| Variable | Description | Default |
|----------|-------------|---------|
| `GRPC_PORT` | gRPC server port | `9090` |
| `HTTP_PORT` | HTTP fallback server port | `9091` |
| `POSTGRES_URL` | PostgreSQL connection string | `postgresql://postgres:postgres@postgres:5432/knowledge_base?sslmode=disable` |
| `REDIS_URL` | Redis address | `redis:6379` |
| `EVENT_CHANNEL` | Redis Pub/Sub channel | `graph:events` |
| `GRAPH_FULL_LIMIT` | Default full graph limit | `1000` |
| `GRAPH_DEFAULT_DEPTH` | Default depth for note layout | `2` |
| `GRAPH_STREAM_CHUNK_SIZE` | Nodes per chunk in streaming | `100` |
| `CACHE_NOTE_TTL_SECONDS` | Note layout cache TTL | `300` (5 min) |
| `CACHE_FULL_TTL_SECONDS` | Full layout cache TTL | `300` (5 min) |
| `CACHE_DELTA_TTL_SECONDS` | Delta cache TTL | `60` (1 min) |

### Layout Engine Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `2d_radius` | Circle radius for 2D layout | `100.0` |
| `3d_radius` | Helix radius for 3D layout | `120.0` |
| `3d_z_step` | Vertical step between nodes in 3D | `5.0` |
| `default_node_size` | Base node size | `1.0` |

### Cache TTL Behavior

- **Note Layout Cache** (`CACHE_NOTE_TTL_SECONDS`): Per-note cached layouts (default: 5 min)
- **Full Layout Cache** (`CACHE_FULL_TTL_SECONDS`): Full graph cached layouts (default: 5 min)
- **Delta Cache** (`CACHE_DELTA_TTL_SECONDS`): Delta responses (default: 1 min)

Cache invalidation happens automatically via Redis Pub/Sub when notes or links are updated.

### Event Tracking Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `event_tracking_ttl_hours` | How long to keep event acknowledgment records | `24` |
| `unprocessed_event_check_interval_minutes` | Retry worker interval | `5` |

### gRPC Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `GetNoteLayout` | Unary | Get layout for a specific note |
| `GetFullLayout` | Server Stream | Stream full graph in chunks |
| `GetDelta` | Unary | Get changes since last hash |

### HTTP Fallback Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/api/v1/graph/note/:id` | GET | Note layout (JSON) |
| `/api/v1/graph/full` | GET | Full graph (JSON) |
| `/api/v1/graph/delta` | GET | Delta (JSON) |

### Usage in docker-compose.yml

```yaml
graph-service:
  build:
    context: .
    dockerfile: ./services/graph-service/Dockerfile
  environment:
    GRPC_PORT: 9090
    HTTP_PORT: 9091
    POSTGRES_URL: postgresql://user:pass@postgres:5432/knowledge_base
    REDIS_URL: redis:6379
    GRAPH_FULL_LIMIT: 1000
    CACHE_NOTE_TTL_SECONDS: 300
```

### How to Apply Changes

#### After editing variables in `docker-compose.yml`, restart graph-service:

```bash
docker-compose restart graph-service
```

#### Check logs — you should see configuration loaded message:

```bash
docker logs kg-graph-service --tail 30
```

### Related Documentation

| Document | Description |
|----------|-------------|
| [`architecture/decisions/013-graph-service-isolation.md`](architecture/decisions/013-graph-service-isolation.md) | Graph Service architecture decision |
| [`architecture/decisions/014-event-driven-cache-invalidation.md`](architecture/decisions/014-event-driven-cache-invalidation.md) | Event-driven cache invalidation |

### How to Apply Changes

#### After editing `.env` or variables in `docker-compose.yml`, restart backend:

```bash
docker-compose restart backend
```

#### Check logs — you should see configuration loaded message:

```bash
docker logs kg-backend --tail 30 | grep "Config loaded"
```

#### Example output:

```
Config loaded: alpha=0.50, beta=0.50, depth=3, decay=0.50, cacheTTL=5m0s, embeddingLimit=30, graphLoadDepth=2
```

### Final Recommendation Score Formula

#### For source note A and candidate C:

```
score = α * explicit_weight(A, C) + β * content_sim(A, C)
explicit_weight(A, C) — sum of weights of all explicit links from A to C (with decay for indirect paths starting from level 2).

content_sim(A, C) — cosine similarity of embeddings, normalized to [0,1] range.

α + β doesn't have to equal 1, but sum = 1 is recommended for consistent scale.
```

### Notes

#### Embeddings are computed asynchronously by worker. Ensure worker (kg-worker) is running and has processed tasks for notes, otherwise content_sim will be 0.

#### If beta > 0, note_embeddings table must have vectors for all notes. If not, recommendations will rely only on explicit links.

#### Changing RECOMMENDATION_DEPTH above 3 may significantly increase database load and response time.

#### Recommendation cache invalidates by TTL. For immediate clearing, restart backend or manually delete keys:

```bash
docker exec kg-redis redis-cli KEYS "suggestions:*" | xargs docker exec kg-redis redis-cli DEL
```

---

## Related Documentation

| Document | Description |
|----------|-------------|
| [`architecture/README.md`](architecture/README.md) | Backend architecture, DDD layers |
| [`architecture/decisions/`](architecture/decisions/) | Architectural Decision Records (ADR) |
| [`DEPLOYMENT_EN.md`](DEPLOYMENT_EN.md) | Deployment, configuration verification |
| [`TESTING.md`](TESTING.md) | Testing, parameter verification |
| [`backend/openAPI.yaml`](../backend/openAPI.yaml) | API specification |


---

## Personal Environment Authentication

### Overview

The personal environment (\docker-compose.personal.yml\) is designed for local development without authentication.

### Configuration

**Environment Variable (.env):**
\\\ash
SKIP_AUTH=true  # Disables authentication for personal environment
\\\

**Behavior:**
- All requests are allowed without authentication
- Notes are created with \creator_id = 00000000-0000-0000-0000-000000000000\ (test user)
- Test user is created by migration \ 19_add_test_user.up.sql\

### Authentication Fix (2026-06-15)

**Issues Fixed:**
1. Corrected \main.go\ to use \DefaultSkipAuthConfig\ instead of manual configuration
2. Changed default user ID from \ 0000000-0000-0000-0000-000000000001\ to \ 0000000-0000-0000-0000-000000000000\ (matches migration)
3. Fixed \pikey.go\ admin UUID to match migration

**Current Status:**
- ✅ \id_creator\ is properly assigned: \ 0000000-0000-0000-0000-000000000000\
- ✅ All notes have proper owner (test user)
- ✅ Consistent with migration 019

### Differences Between Environments

| Environment | Authentication | creator_id | Purpose |
|-------------|----------------|------------|---------|
| **Dev** | Enabled | Real user ID | Development with auth |
| **Personal** | Disabled (SKIP_AUTH=true) | Test user ID (all zeros) | Local testing without auth |

### Security Note

This is intentional for local development. For production or shared environments, authentication should always be enabled.

