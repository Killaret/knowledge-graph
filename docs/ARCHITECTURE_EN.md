# Knowledge Graph — Architecture Documentation

> **Relevance:** Auto-generated via @codemaps:  
> **Date:** 2026-04-27  
> **Stack:** Go + SvelteKit + Python (FastAPI) + PostgreSQL + Redis

---

## 📊 Overall Architecture

The system is built on **Clean Architecture** with 4 layers:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              PRESENTATION LAYER                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐ │
│  │  HTTP API    │  │  WebSocket   │  │  Static      │  │  E2E Tests         │ │
│  │  (Gin)       │  │  (optional)  │  │  (frontend)  │  │  (Cucumber)        │ │
│  └──────────────┘  └──────────────┘  └──────────────┘  └────────────────────┘ │
│                              ↕ interfaces/api/                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                            APPLICATION LAYER                                │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Use Cases (Application Services)                                       │ │
│  │  • CreateNote / UpdateNote / DeleteNote                               │ │
│  │  • CreateLink / UpdateLinkWeight                                      │ │
│  │  • GraphBuilding / RecommendationEngine                               │ │
│  │  • SearchOrchestrator                                                   │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                              ↕ application/                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                              DOMAIN LAYER                                     │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Entities & Business Rules                                              │ │
│  │  • Note (aggregate root)                                                │ │
│  │  • Link (value object)                                                  │ │
│  │  • Graph (traversal algorithms)                                         │ │
│  │  • Repository Interfaces                                                │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                              ↕ domain/                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                           INFRASTRUCTURE LAYER                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │  PostgreSQL  │  │    Redis     │  │   NLP        │  │   Asynq      │   │
│  │  (pgvector)  │  │   (cache)    │  │  Service     │  │  (queue)     │   │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘   │
│       db/               cache/              nlp/              queue/         │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Redis and Private Graph Cache [see ADR 014](architecture/decisions/014-event-driven-cache-invalidation.md)
- Redis is used for caching user's private graph alongside existing tokens, sessions, and queue data.
- Route `/api/v1/me/graph/cached` returns instant graph from Redis.
- Route `/api/v1/me/graph/fresh` computes fresh graph and can return `delta` for incremental UI updates.
- Frontend guarantees instant display via cache and smooth Canvas updates without full redraw.

---

## 🎯 System Components

### 1. Backend (Go)

**Location:** `backend/`

#### 1.1 Entry Points (`cmd/`)

| Command | File | Purpose | Port |
|---------|------|---------|------|
| `server` | `cmd/server/main.go` | HTTP API server | 8080 |
| `worker` | `cmd/worker/main.go` | Background job processor | — |
| `cli` | `cmd/cli/main.go` | Admin CLI | — |

#### 1.2 Domain Layer (`internal/domain/`)

##### Note Domain (`domain/note/`)

```go
type Note struct {
    id        uuid.UUID    // Aggregate ID
    title     Title        // Value Object
    content   Content      // Value Object
    type_     string       // "star", "planet", "moon", etc.
    metadata  Metadata     // JSONB metadata
    createdAt time.Time
    updatedAt time.Time
}
```

**Files:**
- `entity.go` — Aggregate root with business logic
- `value_objects.go` — Title, Content, Metadata
- `repository.go` — Repository interface
- `entity_test.go`, `value_objects_test.go` — Unit tests

##### Link Domain (`domain/link/`)

```go
type Link struct {
    id           uuid.UUID
    sourceNoteID uuid.UUID    // FK → Note
    targetNoteID uuid.UUID    // FK → Note
    linkType     LinkType     // Value Object
    weight       Weight       // Value Object [0..1]
    metadata     Metadata
    createdAt    time.Time
}
```

##### Achievement Domain (`domain/achievement/`) [see ADR 015](architecture/decisions/015-galactic-lexicon-and-achievements.md)

```go
type Achievement struct {
    id          uuid.UUID
    code        string       // Unique identifier (e.g., "first_note")
    title       string       // Display title
    description string       // Description
    icon        string       // Emoji icon
    condition   Condition    // Unlock condition
    points      int          // Achievement points
    isHidden    bool         // Hide until unlocked
}

type Condition struct {
    Type    string                 // "count" or "streak"
    Entity  string                 // "note", "link", "search", "share"
    Action  string                 // "create", "update", "delete"
    Filter  map[string]interface{} // Additional filters (e.g., note type)
    Threshold int                  // Required count
    Days    int                    // Required streak days
}
```

**Files:**
- `entity.go` — Achievement and UserAchievement entities
- `engine.go` — Condition evaluation logic
- `repository.go` — Repository interface

##### Draft Domain (`domain/draft/`) [see ADR 011](architecture/decisions/011-drafts-autosave-mongodb.md)

```go
type Draft struct {
    id        uuid.UUID
    noteID    *uuid.UUID   // Nullable for new notes
    userID    uuid.UUID
    tenantID  uuid.UUID
    content   DraftContent // Rich text content
    metadata  Metadata
    createdAt time.Time
    updatedAt time.Time
    expiresAt time.Time    // TTL for automatic cleanup
}

type DraftContent struct {
    title   string
    body    string
    format  string // "markdown", "richtext"
}
```

**Files:**
- `entity.go` — Draft entity with TTL logic
- `repository.go` — Repository interface (MongoDB implementation)
- `service.go` — Autosave coordination service

##### Graph Domain (`domain/graph/`)

**Graph Traversal Algorithms:**
- `bfs.go` — Breadth-First Search
- `neighbor_loader.go` — Node neighbor loader
- `keyword_matcher.go` — Keyword matching
- `traversal_service.go` — Traversal service with repository integration
- `normalizer.go` — Link weight normalization
- `aggregation.go` — Traversal result aggregation

**Tests:**
- `traversal_test.go` — Unit tests
- `traversal_integration_test.go` — Integration tests

##### Keyword Similarity Architecture

**Code location:**
- `backend/internal/application/recommendation/keyword_similarity.go` — Similarity strategies [see ADR 016](architecture/decisions/016-keyword-similarity-strategies.md)
- `backend/internal/application/recommendation/keyword_matcher_impl.go` — Matcher implementation
- `backend/internal/domain/graph/keyword_matcher.go` — Interface in domain layer

**Configuration** (`knowledge-graph.config.json`):
```json
{
  "backend": {
    "recommendation": {
      "keyword_similarity_method": "jaccard",  // jaccard, overlap, tversky, weighted_jaccard, cosine
      "keyword_tversky_alpha": 0.5,             // Alpha parameter for Tversky
      "keyword_tversky_beta": 0.5,              // Beta parameter for Tversky
      "gamma": 0.2                             // Weight of keyword component (enables function when > 0)
    }
  }
}
```

**Integration** (`backend/cmd/worker/main.go` → `TraversalService`):
```go
// 1. Create strategy from config
keywordSimilarity, err := recommendation.NewKeywordSimilarity(
    cfg.RecommendationKeywordSimilarityMethod,
    cfg.RecommendationKeywordTverskyAlpha,
    cfg.RecommendationKeywordTverskyBeta,
)

// 2. Create matcher with keyword repository
keywordMatcher := recommendation.NewKeywordMatcherImpl(keywordRepo, keywordSimilarity)

// 3. Configure TraversalService
traversalSvc := graphDomain.NewTraversalServiceWithWeights(...)

// 4. Set matcher if gamma > 0
if cfg.RecommendationGamma > 0 {
    traversalSvc.SetKeywordMatcher(keywordMatcher)
}
```

**Available Strategies:**
| Strategy | Description | Requires Weight |
|----------|-------------|-----------------|
| `jaccard` | Classic Jaccard coefficient: |A ∩ B| / |A ∪ B| | No |
| `overlap` | Overlap coefficient: |A ∩ B| / min(|A|, |B|) | No |
| `tversky` | Tversky index with alpha/beta parameters: |A ∩ B| / (|A ∩ B| + α|A\B| + β|B\A|) | No |
| `weighted_jaccard` | Weighted Jaccard: sum(min(w1, w2)) / sum(max(w1, w2)) | Yes |
| `cosine` | Cosine similarity of weight vectors | Yes |

**Data Flow:**
```
TraversalService.GetSuggestions()
    ↓
keywordMatcher.Match(sourceID, candidateIDs)
    ↓
keywordSimilarity.Similarity(sourceKeywords, targetKeywords, weights)
    ↓
AggregateWeighted(graphScore, semanticScore, keywordScore, alpha, beta, gamma)
```

#### 1.3 Application Layer (`internal/application/`)

##### Graph Application (`application/graph/`)

| File | Purpose |
|------|---------|
| `composite_loader.go` | Composite loader (keywords + embeddings) |
| `embedding_loader.go` | Vector proximity loading |
| `neighbor_loader.go` | Neighbor loading via links |
| `composite_loader_test.go` | Tests |

##### Recommendation Application (`application/recommendation/`)

| File | Purpose |
|------|---------|
| `refresh_service.go` | Recommendation refresh service |
| `affected_notes.go` | Identify affected notes |
| `keyword_similarity.go` | Keyword similarity strategies (Jaccard, Overlap, Tversky, Weighted Jaccard, Cosine) [see ADR 016](architecture/decisions/016-keyword-similarity-strategies.md) |
| `keyword_matcher_impl.go` | KeywordMatcher implementation using KeywordSimilarity |
| `*_test.go` | Unit tests |

##### Achievement Application (`application/achievement/`) [see ADR 015](architecture/decisions/015-galactic-lexicon-and-achievements.md)

| File | Purpose |
|------|---------|
| `service.go` | Achievement service with trigger checking |
| `engine.go` | Achievement engine for condition evaluation |
| `engine_test.go` | Unit tests for engine |

**Features:**
- `CheckTrigger` — Synchronous check for achievement unlocking on user actions
- `CheckStreaks` — Asynchronous streak-based achievement checking
- `TrackLogin` — Login streak tracking with Redis
- Notification integration with user settings

##### Common (`application/common/`)

- `task_queue.go` — Abstraction over task queue

##### Queries (`application/queries/graph/`)

- `get_suggestions.go` — Query handler for getting suggestions

#### 1.4 Infrastructure Layer (`internal/infrastructure/`)

##### Database (`infrastructure/db/`)

**PostgreSQL Repositories (`db/postgres/`):**

| File | Entity | CRUD | Specifics |
|------|--------|------|-----------|
| `note_repo.go` | Note | ✅ | Full-text search, pagination |
| `link_repo.go` | Link | ✅ | Cascade delete by source |
| `embedding_repo.go` | Embedding | ✅ | pgvector similarity search |
| `tag_repo.go` | Tag | ✅ | Many-to-many with notes |
| `user_repo.go` | User | ✅ | Auth data |
| `recommendation_repo.go` | Recommendation | ✅ | Suggestions storage |
| `achievement_repo.go` | Achievement | ✅ | User achievement tracking |
| `user_settings_model.go` | UserSettings | ✅ | User preferences (galactic_mode, etc.) |

**Models (`db/postgres/*_model.go`):**
- `note_model.go` — GORM model for Note
- `note_embedding_model.go` — Vector embeddings (pgvector)
- `note_keyword_model.go` — Extracted keywords
- `link_model.go` — GORM model for Link
- `tag_model.go`, `note_tag_model.go` — Tagging system
- `recommendation_model.go` — Precomputed recommendations

**Migrations:**
- `migrations.go` — Migration runner
- `../../migrations/` — 27 SQL migration files

##### NLP Client (`infrastructure/nlp/`)

- `client.go` — HTTP client for NLP Service
- `client_test.go` — Tests with mocks

**Endpoints:**
- `POST /extract_keywords` — YAKE keyword extraction
- `POST /embed` — SentenceTransformers embeddings

##### Queue (`infrastructure/queue/`)

**Asynq (Redis-based task queue):**

| File | Purpose |
|------|---------|
| `asynq_client.go` | Client for enqueuing tasks |
| `worker.go` | Worker processor |
| `tasks.go` | Task definitions |
| `tasks/recommendation.go` | Recommendation refresh task |

**Task Types:**
- `recommendation:refresh` — Refresh recommendations for note
- `backup:cloud` — Upload backup to cloud storage

##### Cloud (`infrastructure/cloud/`)

**Yandex.Disk Backup Service:**

| File | Purpose |
|------|---------|
| `yandex_backup.go` | YandexBackupService for Yandex.Disk via WebDAV |

**Methods:**
- `UploadBackup()` — Upload backup with retry logic
- `DownloadBackup()` — Download backup
- `ListBackups()` — List backups in cloud
- `DeleteBackup()` — Delete backup

#### 1.5 Interfaces (HTTP Handlers) (`internal/interfaces/api/`)

**REST API Endpoints:**

```
GET    /health              → Health check
GET    /notes               → List notes (paginated)
POST   /notes               → Create note
GET    /notes/:id           → Get note
PUT    /notes/:id           → Update note
DELETE /notes/:id           → Delete note
GET    /notes/:id/suggestions → Get recommendations
GET    /notes/search        → Full-text search

GET    /links               → List links
POST   /links               → Create link
GET    /links/:id           → Get link
PUT    /links/:id           → Update link
DELETE /links/:id           → Delete link

GET    /graph               → Get graph data (nodes + edges)
GET    /graph/3d            → 3D graph data (hierarchical)

GET    /achievements        → List all achievements
GET    /users/me/achievements → Get user's achievements
POST   /users/me/achievements/:id/mark-seen → Mark achievement notification as seen
```

---

### 2. Frontend (SvelteKit)

**Location:** `frontend/`

#### 2.1 Routes (`src/routes/`)

| Route | File | Purpose |
|-------|------|---------|
| `/` | `+page.svelte` | Main page with note list |
| `/graph` | `graph/+page.svelte` | 2D interactive graph (D3.js) |
| `/graph/3d` | `graph/3d/+page.svelte` | 3D graph (Three.js) |
| `/graph/3d/:id` | `graph/3d/[id]/+page.svelte` | 3D graph focused on note |
| `/graph/:id` | `graph/[id]/+page.svelte` | 2D graph focused |
| `/notes/:id` | `notes/[id]/+page.svelte` | View note |
| `/notes/:id/edit` | `notes/[id]/edit/+page.svelte` | Edit note |
| `/notes/new` | `notes/new/+page.svelte` | Create note |
| `/search` | `search/+page.svelte` | Full-text search |

#### 2.2 Components (`src/lib/components/`)

**Core Components (46 total):**

| Component | Technology | Purpose | Status |
|-----------|------------|---------|--------|
| `GraphCanvas.svelte` | D3.js | 2D force-directed graph | Active |
| `Graph3D.svelte` | Three.js | 3D celestial visualization [see ADR 017](architecture/decisions/017-color-palette-redesign.md) | **Frozen** |
| `LazyGraph3D.svelte` | dynamic import | Lazy loading 3D | **Frozen** |
| `SmartGraph.svelte` | D3+Svelte | Adaptive graph (2D/3D) | 2D only |
| `NoteCard.svelte` | Svelte | Note card | Active |
| `NoteEditor.svelte` | Svelte | WYSIWYG editor | Active |
| `NoteSidePanel.svelte` | Svelte | Side panel details | Active |
| `CreateNoteModal.svelte` | Svelte | Create modal | Active |
| `EditNoteModal.svelte` | Svelte | Edit modal | Active |
| `SearchBar.svelte` | Svelte | Search bar | Active |
| `Sidebar.svelte` | Svelte | Navigation | Active |
| `FloatingControls.svelte` | Svelte | Floating control buttons | Active |
| `BackButton.svelte` | Svelte | Back navigation | Active |
| `ConfirmModal.svelte` | Svelte | Action confirmation | Active |
| `Modal.svelte` | Svelte | Base modal | Active |
| `Button.svelte` | Svelte | UI Button | Active |
| `TypeSelector.svelte` | Svelte | Node type selector (star/planet/moon) | Active |
| `ToastNotification.svelte` | Svelte | Ephemeral notifications with galactic mode support | Active |
| `ApiErrorDisplay.svelte` | Svelte | Error display with lexicon integration | Active |
| `ShareModal.svelte` | Svelte | Share modal with lexicon integration | Active |

#### 2.3 API Client (`src/lib/api/`)

| File | Purpose |
|------|---------|
| `client.ts` | ky instance configuration |
| `notes.ts` | Notes API (CRUD + search + suggestions) |
| `links.ts` | Links API |
| `graph.ts` | Graph data API |
| `achievements.ts` | Achievements API |

#### 2.4 Utilities (`src/lib/utils/`)

| File | Purpose |
|------|---------|
| `galactic-lexicon.ts` | Galactic Lexicon - themed messaging system |
| `galactic-lexicon.test.ts` | Unit tests for Galactic Lexicon |

**Galactic Lexicon:** [see ADR 015](architecture/decisions/015-galactic-lexicon-and-achievements.md)
- Supports two modes: `standard` (technical) and `galactic` (space-themed metaphors)
- Categories: success, error, info, warning, achievement
- Locales: Russian (ru) and English (en)
- User-controlled via `galactic_mode` setting in user_settings table
- Integrated into all UI components (modals, toasts, error displays)

#### 2.5 Stores (`src/lib/stores/`)

| File | Purpose |
|------|---------|
| `auth.svelte.js` | Authentication state |
| `lexicon-settings.ts` | Lexicon locale and mode settings |
| `achievements.ts` | Achievement polling and notification state |

#### 2.6 3D Engine (`src/lib/three/`) - **FROZEN for v1.0**

> **🚫 FROZEN FEATURE:** 3D graph functionality has been temporarily frozen for version 1.0 to improve stability and reduce maintenance overhead. See CHANGELOG.md for details.

| File | Purpose | Status |
|------|---------|--------|
| `scene.ts` | Three.js scene setup | Frozen |
| `camera.ts` | Camera controls | Frozen |
| `renderer.ts` | WebGL renderer | Frozen |
| `graph3d.ts` | 3D graph visualization logic | Frozen |
| `celestial.ts` | Celestial body rendering (stars, planets) [see ADR 017](architecture/decisions/017-color-palette-redesign.md) | Frozen |
| `controls.ts` | OrbitControls wrapper | Frozen |
| `animation.ts` | Animation loop | Frozen |
| `types.ts` | TypeScript types | Frozen |

#### 2.5 State Management (`src/lib/stores/`)

- `notes.ts` — Svelte store for notes
- `graph.ts` — Graph state store
- `ui.ts` — UI state (modals, selection)

---

### 3. NLP Service (Python)

**Location:** `nlp-service/`

**Stack:** FastAPI + spaCy + sentence-transformers + YAKE + NLTK

#### 3.1 API Endpoints (`app/main.py`)

```python
GET  /health              → {status, model_loaded, version}
POST /extract_keywords    → ExtractKeywordsResponse
POST /embed               → EmbedResponse
```

#### 3.2 Models (`app/models.py`)

```python
ExtractKeywordsRequest:  {text: str, top_n: int}
ExtractKeywordsResponse: {keywords: [{keyword, weight}]}
EmbedRequest:          {text: str}
EmbedResponse:         {embedding: float[]}
```

#### 3.3 NLP Utils (`app/nlp_utils.py`)

| Function | Library | Purpose |
|----------|---------|---------|
| `extract_keywords()` | YAKE | Keyword extraction (RU/EN) |
| `embedding_model.encode()` | sentence-transformers | Text vectorization |

**Model:** `all-MiniLM-L6-v2` (384 dimensions)

#### 3.4 Tests (`tests/`)

- `test_api.py` — FastAPI endpoint tests
- `test_nlp_utils.py` — NLP function tests

---

### 4. Infrastructure Services

#### 4.1 PostgreSQL (pgvector)

**Docker:** `pgvector/pgvector:pg16`

**Database:** `knowledge_base`
**User:** `kb_user`

**Extensions:**
- `pgvector` — Vector operations
- `pg_trgm` — Trigram search

**Tables:**
```sql
notes          — Notes
links          — Links between notes
note_embeddings — Vectors (384 dim)
keywords       — Keywords
tags           — Tags
note_tags      — Many-to-many
recommendations — Precomputed recommendations
users          — Users
```

#### 4.2 Redis

**Purpose:**
- Task queue backend (Asynq)
- Cache layer (optional)

#### 4.3 Docker Compose

**Services (dev stack):**
1. `postgres` — pgvector (port 5432)
2. `redis` — Redis 7 (port 6379)
3. `nlp` — Python service (port 5000)
4. `backend` — Go API (port 8080)
5. `worker` — Background worker
6. `frontend` — SvelteKit (port 3000)

**Services (personal stack):**
1. `postgres_personal` — pgvector (port 5433)
2. `redis_personal` — Redis 7 (port 6380)
3. `mongo_personal` — MongoDB 7 (port 27018) [see ADR 011](architecture/decisions/011-drafts-autosave-mongodb.md) — drafts
4. `nlp` — Python service (port 5001)
5. `graph-service-personal` — Graph service (port 9092) [see ADR 013](architecture/decisions/013-graph-service-isolation.md)
6. `backend_personal` — Go API (port 8081)
7. `worker_personal` — Background worker
8. `nginx_personal` — Reverse proxy (ports 8082, 8083)
9. `frontend_personal` — SvelteKit (port 3001)
10. `backup_scheduler` — Automatic backup service

#### 4.4 Backup Service

**Purpose:** Automatic PostgreSQL database backup with local storage and Yandex.Disk cloud backup support.

**Components:**

**Backup scripts:**
- `scripts/utility/backup-personal.sh` — Bash script for Linux/Mac
- `scripts/utility/backup-personal.ps1` — PowerShell script for Windows

**Go-service:**
- `backend/internal/infrastructure/cloud/yandex_backup.go` — YandexBackupService for Yandex.Disk via WebDAV API

**Asynq task:**
- `TypeBackupToCloud` — Async task for uploading backups to cloud

**Docker service:**
- `backup_scheduler` — Automatic backup execution every 24 hours (in docker-compose.personal.yml)

**Functionality:**
1. **Local backup:**
   - pg_dump PostgreSQL database
   - gzip compression
   - Storage in `./backups/`
   - Automatic cleanup of old backups (default 7 days)

2. **Cloud backup (Yandex.Disk):**
   - Upload via WebDAV API
   - OAuth authentication
   - Storage in `/KnowledgeGraphBackups/`
   - Automatic cleanup (max_backups, default 10)
   - Retry logic (3 attempts)

3. **Configuration:**
   ```json
   {
     "backup": {
       "local_path": "./backups",
       "cloud": {
         "enabled": true,
         "provider": "yandex",
         "yandex": {
           "oauth_token": "token",
           "backup_folder": "/KnowledgeGraphBackups",
           "max_backups": 10
         }
       },
       "schedule": "0 2 * * *",
       "retention_days": 7
     }
   }
   ```

**YandexBackupService Methods:**
- `UploadBackup(ctx, localPath, remoteKey)` — Upload backup with retry logic
- `DownloadBackup(ctx, remoteKey, localPath)` — Download backup
- `ListBackups(ctx, prefix)` — List backups in cloud
- `DeleteBackup(ctx, remoteKey)` — Delete backup
- `ensureFolder(ctx, folderURL)` — Create folder on Yandex.Disk
- `cleanupOldBackups(ctx)` — Cleanup old backups

**Environment Variables:**
- `BACKUP_CLOUD_ENABLED` — Enable cloud backup
- `BACKUP_YANDEX_TOKEN` — Yandex.Disk OAuth token
- `BACKUP_YANDEX_FOLDER` — Folder on Yandex.Disk
- `BACKUP_DIR` — Local backup folder
- `CLEANUP_OLD_BACKUPS` — Cleanup old backups

**More details:** [`docs/BACKUP.md`](docs/BACKUP.md)

---

## 🔄 Data Flow

### Creating note with recommendations

```
1. User creates note (Frontend)
   ↓ POST /notes
2. HTTP Handler receives request
   ↓
3. Application Service: CreateNote
   ├─ Validate input
   ├─ Create Note aggregate
   ├─ Save to PostgreSQL (note_repo)
   └─ Enqueue task: recommendation:refresh
   ↓
4. Response: 201 Created
   ↓
5. Worker picks up task (async)
   ├─ Call NLP: /extract_keywords
   ├─ Call NLP: /embed
   ├─ Save embedding to pgvector
   ├─ Calculate recommendations
   │  ├─ Vector similarity search
   │  ├─ Keyword matching
   │  └─ Graph traversal (BFS)
   └─ Save recommendations
```

### Requesting recommendations

```
GET /notes/:id/suggestions
   ↓
Query Handler: GetSuggestions
   ├─ Check precomputed recommendations
   ├─ If stale/empty → trigger refresh
   └─ Return top-N suggestions
```

### Graph search

```
GET /graph?center=:id&depth=2
   ↓
Graph Application Service
   ├─ Load center node
   ├─ BFS traversal (depth-limited)
   ├─ Load neighbors via links
   ├─ Enrich with metadata
   └─ Return: {nodes, edges}
```

---

## 🧪 Testing Strategy

### Backend Tests

| Type | Location | Framework |
|------|----------|-----------|
| Unit | `*_test.go` (next to code) | Go testing |
| Integration | `*_integration_test.go` | Go testing + testcontainers |
| E2E | `tests/features/` | Cucumber + Playwright |

### Frontend Tests

| Type | Location | Framework |
|------|----------|-----------|
| Unit | `*.test.ts` | Vitest |
| Component | `*.spec.ts` | Testing Library |
| E2E | `tests/` | Playwright |

### NLP Tests

| Type | Location | Framework |
|------|----------|-----------|
| Unit | `tests/test_nlp_utils.py` | pytest |
| API | `tests/test_api.py` | pytest + FastAPI TestClient |

---

## 📦 Dependencies

### Backend (go.mod)

```
github.com/gin-gonic/gin        # HTTP router
github.com/jackc/pgx/v5         # PostgreSQL driver
github.com/hibiken/asynq        # Task queue
github.com/redis/go-redis/v9    # Redis client
github.com/google/uuid          # UUID generation
```

### Frontend (package.json)

```
svelte                          # Framework
@threlte/core                   # Three.js for Svelte
three                           # 3D engine
d3                              # 2D graph visualization
ky                              # HTTP client
```

### NLP (requirements.txt)

```
fastapi                         # Web framework
sentence-transformers           # Embeddings
yake                            # Keywords
spacy                           # NLP
nltk                            # Text processing
```

---

## 🚀 Deployment

### Local Development

```bash
docker-compose up -d
# Or:
make dev
```

### Production Considerations

- **Database:** Connection pooling (pgx pool)
- **Cache:** Redis cluster for high availability
- **Queue:** Horizontal scaling workers
- **NLP:** GPU instances for embeddings
- **Frontend:** CDN for static assets

---

## 🔐 Security

- **CORS:** Configured for frontend origin
- **SQL Injection:** GORM + parameterized queries
- **XSS:** Svelte auto-escaping
- **Input Validation:** Validator at all layers

---

## 📚 Additional Documentation

- `API_ERRORS.md` — API errors and codes
- `ROADMAP.md` — Development plans (moved to project root)
- `WEIGHTS_CALCULATION.md` — Link weight calculation logic
- `docs/architecture/c4/` — C4 Model diagrams
- `docs/architecture/decisions/` — ADR (Architecture Decision Records)

### Key ADR (Architecture Decision Records)

- **[ADR 011: Drafts Autosave in MongoDB](architecture/decisions/011-drafts-autosave-mongodb.md)** — Autosave drafts in MongoDB with eventual sync to PostgreSQL
- **[ADR 013: Graph Service Isolation](architecture/decisions/013-graph-service-isolation.md)** — Extract Graph Service as separate microservice with gRPC and direct DB access
- **[ADR 014: Event-Driven Cache Invalidation](architecture/decisions/014-event-driven-cache-invalidation.md)** — Use Redis Pub/Sub for cache invalidation
- **[ADR 015: Galactic Lexicon and Achievements](architecture/decisions/015-galactic-lexicon-and-achievements.md)** — Unified galactic lexicon, i18n, SSE for achievement notifications
- **[ADR 016: Keyword Similarity Strategies](architecture/decisions/016-keyword-similarity-strategies.md)** — Strategy pattern for keyword similarity metrics with weight support
- **[ADR 017: Color Palette Redesign](architecture/decisions/017-color-palette-redesign.md)** — Black-purple-red color palette for space theme

---

## 🗺️ Module Dependency Graph

```
cmd/
├── server → interfaces/api → application/* → domain/* → infrastructure/*
├── worker → application/common (task_queue) → infrastructure/queue
└── cli    → infrastructure/db (migrations)

interfaces/api/
├── handlers/note.go     → application/graph, domain/note
├── handlers/link.go     → domain/link
├── handlers/graph.go    → application/graph
└── handlers/search.go   → domain/note (Search)

application/graph/
├── composite_loader.go  → domain/graph, domain/note, domain/link
├── embedding_loader.go  → infrastructure/db/postgres (embedding_repo)
└── neighbor_loader.go   → domain/graph, domain/link

application/recommendation/
├── refresh_service.go   → domain/note, domain/link, infrastructure/nlp
└── affected_notes.go    → domain/link

infrastructure/
├── db/postgres/         → domain/* (implements Repository interfaces)
├── nlp/client.go        → (external HTTP calls)
└── queue/               → application/common (task_queue abstraction)
```

---

## 📊 Code Statistics

| Component | Files | Complexity |
|-----------|-------|------------|
| Domain | 19 | Low (business logic) |
| Application | 10 | Medium (orchestration) |
| Infrastructure | 38 | High (technical details) |
| Interfaces | 16 | Medium (HTTP) |
| Frontend | 46 | Medium (UI) |
| NLP | 4 | Low (models) |

---

## 📈 Graph Service

### Overview

Graph Service is an independent microservice responsible for computing 2D/3D graph layouts for the Knowledge Graph frontend. It provides high-performance graph visualization with caching, incremental updates, and event-driven invalidation [see ADR 014](architecture/decisions/014-event-driven-cache-invalidation.md).

### Architecture

The Graph Service consists of:

- **API Layer**: gRPC server (port 9090) and HTTP fallback (port 9091) [see ADR 013](architecture/decisions/013-graph-service-isolation.md)
- **Layout Engine**: 2D circular and 3D spiral layout algorithms with delta computation
- **Cache Layer**: Redis-backed caching with configurable TTL
- **Data Layer**: Direct PostgreSQL read access (notes, links, embeddings)
- **Event Subscriber**: Redis Pub/Sub for cache invalidation with acknowledgment tracking

### API Contracts

#### gRPC API (Primary) [see ADR 013](architecture/decisions/013-graph-service-isolation.md)

```protobuf
service GraphService {
  rpc GetNoteLayout(NoteLayoutRequest) returns (LayoutResponse);
  rpc GetFullLayout(FullLayoutRequest) returns (stream LayoutChunk);
  rpc GetDelta(DeltaRequest) returns (DeltaResponse);
}
```

#### HTTP Fallback (Secondary)

```
GET /api/v1/graph/note/:id?depth=2&user_id={userId}
GET /api/v1/graph/full?limit=1000&user_id={userId}
GET /api/v1/graph/delta?last_hash={hash}&user_id={userId}
GET /health
```

### Event-Driven Cache Invalidation

The Graph Service subscribes to Redis Pub/Sub channel `graph:events` and processes events with acknowledgment tracking:

1. **Event receipt**: Records `timestamp_received` and `is_acknowledged=false`
2. **Cache invalidation**: Invalidates affected cache keys based on event type
3. **Event acknowledgment**: Sets `is_acknowledged=true` and `timestamp_processed`
4. **Periodic worker**: Scans for unacknowledged events older than 5 minutes

Supported events: `NoteCreated`, `NoteUpdated`, `NoteDeleted`, `LinkCreated`, `LinkUpdated`, `LinkDeleted`

### Caching Strategy

- `layout:note:{noteId}:depth-{depth}` - Note neighborhood layout (30 min TTL)
- `layout:full:{userId}` - Full user graph layout (30 min TTL)
- `layout:delta:{userId}:{lastHash}` - Delta between versions (5 min TTL)

### Direct PostgreSQL Reading

Graph Service reads directly from PostgreSQL as the single source of truth, eliminating the need for an Outbox pattern while maintaining consistency through event-driven cache invalidation [see ADR 014](architecture/decisions/014-event-driven-cache-invalidation.md).

---

*Generated with ❤️ by Cascade*
