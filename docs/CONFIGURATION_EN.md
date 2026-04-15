# Knowledge Graph Configuration

All system parameters are set via environment variables. For local development, define them in `.env` file in the project root. In production, pass variables through `environment` in `docker-compose.yml` or `ConfigMap`/`Secrets` in Kubernetes.

🌐 **Languages**: [English](CONFIGURATION_EN.md) | [Русский](CONFIGURATION.md)

---

## Required Variables

| Variable | Component | Description | Example |
|----------|-----------|-------------|---------|
| `DATABASE_URL` | backend, worker | PostgreSQL connection string. Format: `postgresql://user:password@host:port/dbname?sslmode=disable` | `postgresql://kb_user:kb_pass@postgres:5432/knowledge_base?sslmode=disable` |

**Used in:**
- `backend` — database connection for API requests
- `worker` — connection for saving async task results

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

## Recommendation Parameters (Graph + Embeddings)

| Variable | Where | Description | Default | Range |
|----------|-------|-------------|---------|-------|
| `RECOMMENDATION_ALPHA` | backend | Weight of explicit links (BFS) in final score | `0.5` | 0.0 - 1.0 |
| `RECOMMENDATION_BETA` | backend | Weight of semantic similarity (embeddings) | `0.5` | 0.0 - 1.0 |
| `RECOMMENDATION_DEPTH` | backend | Maximum BFS traversal depth | `3` | 1 - 5 |
| `RECOMMENDATION_DECAY` | backend | Weight decay for indirect links (depth > 1) | `0.5` | 0.0 - 1.0 |
| `RECOMMENDATION_CACHE_TTL_SECONDS` | backend | Recommendation cache TTL in Redis | `300` | 60 - 3600 |
| `EMBEDDING_SIMILARITY_LIMIT` | backend | Limit for pgvector similarity candidates | `30` | 10 - 100 |

### Detailed Description

**ALPHA + BETA** — Combination formula:
```
score = α × explicit_score + β × semantic_score
```
- `α + β` doesn't have to equal 1, but recommended for consistent scale
- `α = 1.0, β = 0.0` — explicit links only (ignore semantics)
- `α = 0.0, β = 1.0` — semantics only (ignore links)
- `α = 0.7, β = 0.3` — balanced: links matter more, but semantics considered

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

---

## Graph Visualization Parameters

| Variable | Where | Description | Default | Range |
|----------|-------|-------------|---------|-------|
| `GRAPH_LOAD_DEPTH` | backend, frontend | Graph loading depth for 3D visualization | `2` | 1 - 4 |

---

## Advanced Recommendation Parameters (BFS + Asynq)

> ⚠️ These parameters are declared in `config.go` but **not yet used** in code. Reserved for future algorithm improvements.

| Variable | Where | Description | Default | Status |
|----------|-------|-------------|---------|--------|
| `RECOMMENDATION_GAMMA` | backend | Coefficient for third component | `0.2` | 🚧 Reserved |
| `BFS_AGGREGATION` | backend | Weight aggregation method: `max`, `sum`, `avg` | `max` | 🚧 Reserved |
| `BFS_NORMALIZE` | backend | Normalize link weights | `true` | 🚧 Reserved |
| `ASYNQ_CONCURRENCY` | worker | Asynq concurrency level | `10` | 🚧 Reserved |
| `ASYNQ_QUEUE_DEFAULT` | worker | Default queue priority | `1` | 🚧 Reserved |

### Reserved Parameters Description

**RECOMMENDATION_GAMMA** — Coefficient for third component:
- Allows adding third factor to calculation (e.g., time factor or popularity)
- `0.2` — small contribution from additional factor

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

## Full `.env` File Example

```env
# Required
DATABASE_URL=postgresql://kb_user:kb_password@postgres:5432/knowledge_base?sslmode=disable

# Optional
SERVER_PORT=8080
REDIS_URL=redis:6379
NLP_SERVICE_URL=http://nlp:5000

# Recommendations
RECOMMENDATION_ALPHA=0.5
RECOMMENDATION_BETA=0.5
RECOMMENDATION_DEPTH=3
RECOMMENDATION_DECAY=0.5
RECOMMENDATION_CACHE_TTL_SECONDS=300
EMBEDDING_SIMILARITY_LIMIT=30

# Graph visualization
GRAPH_LOAD_DEPTH=2

# Advanced recommendation parameters
RECOMMENDATION_GAMMA=0.2
BFS_AGGREGATION=max
BFS_NORMALIZE=true

# Asynq worker parameters
ASYNQ_CONCURRENCY=10
ASYNQ_QUEUE_DEFAULT=1
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
| [`architecture/adr.md`](architecture/adr.md) | Architectural Decision Records (ADR) |
| [`DEPLOYMENT_EN.md`](DEPLOYMENT_EN.md) | Deployment, configuration verification |
| [`TESTING.md`](TESTING.md) | Testing, parameter verification |
| [`backend/openAPI.yaml`](../backend/openAPI.yaml) | API specification |
