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
│  │  HTTP API    │  │  SSE/WS (¹)  │  │  Static      │  │  E2E Tests         │ │
│  │  (Gin)       │  │              │  │  (frontend)  │  │  (Cucumber)        │ │
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

(¹) Server-push channel — not implemented (DOC-AUDIT-2, 2026-09-26). ADR 014 prescribes SSE;
tracked in SYNC-1 stage C. Updates currently reach the client via polling. [see ADR 014](decisions/014-event-driven-cache-invalidation.md)
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
| `cli` | `cmd/cli/main.go` | Recommendation precomputation CLI | — |

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

##### Achievement Domain (`domain/achievement/`) [see ADR 015](decisions/015-galactic-lexicon-and-achievements.md)

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
- `entity.go` — Achievement and UserAchievement entities + `Repository` interface
- `engine.go` — Condition evaluation logic
- `model.go` — Condition/trigger model types

##### Draft Domain [see ADR 011](decisions/011-drafts-autosave-mongodb.md)

Drafts live inside the note domain, not a separate `domain/draft/`:

- `domain/note/draft.go` — `Draft` entity (`noteID`, `userID`, `content`, `title`, `DraftState`,
  timestamps; TTL-based cleanup of abandoned drafts)
- `domain/note/draft_repository.go` — repository interface
- `application/draft/service.go` — autosave coordination service
- `infrastructure/mongo/draft_repo.go` — MongoDB implementation
- `interfaces/api/handlers/draft/` — HTTP handlers (`/notes/:id/draft`, `/drafts/:id/*`)

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
- `graph_extra_test.go` — Additional unit tests (matcher wiring, weights)

##### Keyword Similarity Architecture

**Code location:**
- `backend/internal/application/recommendation/keyword_similarity.go` — Similarity strategies [see ADR 016](decisions/016-keyword-similarity-strategies.md)
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
| `keyword_similarity.go` | Keyword similarity strategies (Jaccard, Overlap, Tversky, Weighted Jaccard, Cosine) [see ADR 016](decisions/016-keyword-similarity-strategies.md) |
| `keyword_matcher_impl.go` | KeywordMatcher implementation using KeywordSimilarity |
| `*_test.go` | Unit tests |

##### Achievement Application (`application/achievement/`) [see ADR 015](decisions/015-galactic-lexicon-and-achievements.md)

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
- `../../migrations/` — 68 SQL migration files

##### NLP Client (`infrastructure/nlp/`)

- `client.go` — HTTP client for NLP Service
- `client_test.go` — Tests with mocks

**Endpoints:**
- `POST /extract_keywords` — KeyBERT-hybrid keyword extraction with lemmatization (NLP-2)
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

**REST API Endpoints** (registered in `backend/cmd/server/router.go`; the full surface also
covers auth, users, settings, drafts, share, import, tags, backup, refetch, quality):

```
GET    /health              → Health check
GET    /notes               → List notes (paginated)
POST   /notes               → Create note
GET    /notes/:id           → Get note
PUT    /notes/:id           → Update note
DELETE /notes/:id           → Delete note
GET    /notes/:id/suggestions → Get recommendations
GET    /notes/search        → Full-text search

POST   /links               → Create link
GET    /links/:id           → Get link
PUT    /links/:id           → Update link
DELETE /links/:id           → Delete link
GET    /notes/:id/links     → Links of a note
DELETE /notes/:id/links     → Delete links of a note

GET    /notes/:id/graph     → Graph data rooted at a note (depth-limited)
GET    /graph/public        → Public graph (no auth)
GET    /graph/analytics     → Graph analytics
GET    /me/graph/cached     → Private graph from Redis cache
GET    /me/graph/fresh      → Private graph computed fresh (delta-capable)

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
| `/` | `+page.svelte` | Main page: note list + graph cockpit |
| `/auth/login` | `auth/login/+page.svelte` | Login |
| `/auth/register` | `auth/register/+page.svelte` | Registration |
| `/auth/forgot-password` | `auth/forgot-password/+page.svelte` | Password reset request |
| `/auth/reset-password` | `auth/reset-password/+page.svelte` | Password reset |
| `/auth/yandex/callback` | `auth/yandex/callback/+page.svelte` | Yandex OAuth callback |
| `/graph` | `graph/+page.svelte` | 2D interactive graph (D3-force) |
| `/graph/3d` | `graph/3d/+page.svelte` | 3D graph (Three.js) |
| `/graph/3d/:id` | `graph/3d/[id]/+page.svelte` | 3D graph focused on note |
| `/graph/:id` | `graph/[id]/+page.svelte` | 2D graph focused |
| `/import` | `import/+page.svelte` | Import hub |
| `/import/bookmarks` | `import/bookmarks/+page.svelte` | Bookmark import |
| `/notes/:id` | `notes/[id]/+page.svelte` | View note |
| `/notes/:id/edit` | `notes/[id]/edit/+page.svelte` | Edit note |
| `/notes/new` | `notes/new/+page.svelte` | Create note |
| `/profile` | `profile/+page.svelte` | User profile and settings |
| `/search` | `search/+page.svelte` | Full-text search |
| `/test/*` | `test/**` | Isolated component test pages |

#### 2.2 UI Structure (FSD layers)

The frontend follows Feature-Sliced Design; components are distributed across layers, not in a
single flat folder:

| Layer | Location | Holds |
|-------|----------|-------|
| `widgets/` | `src/widgets/` | `graph-canvas/` (`GraphCanvas.svelte`, `SmartGraph.svelte`), `graph-3d-viewer/`, `notes/` (`NoteCard.svelte`), `cosmic-cockpit/`, `search/`, `quick-capture/`, `notification/`, `auth/`, `floating-auth-panel/`, `confirm/` |
| `features/` | `src/features/` | `graph-3d/` (Three.js engine — see 2.6), `home-page/`, `graph-interaction/`, `graph-forms/`, `graph-canvas/`, `graph-ui/`, `cosmic-cockpit/`, `cosmic-ui/`, `preload/` |
| `entities/` | `src/entities/` | `note/`, `achievement/` (model + UI), `link/`, `user/`, `tag/` |
| `components/` | `src/components/` | Atomic design: `atoms/` (Button, Modal, IconButton, SplashScreen, ApiErrorDisplay…), `molecules/` (TypeSelector, TagSelector, NoteForm, GraphNodeContextMenu…), `organisms/` (NoteEditor, LoginForm, RegisterForm, ProfileEditor, PreloadIndicator…) |

#### 2.3 API Client (`src/shared/api/`)

| File | Purpose |
|------|---------|
| `client.ts` | ky instance configuration |
| `auth.ts` | Auth API (login, register, OAuth) |
| `notes.ts` | Notes API (CRUD + search + suggestions) |
| `links.ts` | Links API |
| `graph.ts` | Graph data API (full, delta, resync flag) |
| `import.ts` | Bookmark/batch import API |
| `sharing.ts` | Share links and shared-note access |
| `users.ts` | User profile/settings API |
| `quality.ts` | Note quality API |
| `errorMessage.ts` | API error → message mapping |

Achievement data lives in `entities/achievement/model/achievement.ts`, not `shared/api/`.

#### 2.4 Utilities (`src/shared/utils/`)

| File | Purpose |
|------|---------|
| `galactic-lexicon.ts` | Galactic Lexicon - themed messaging system |
| `galactic-lexicon.test.ts` | Unit tests for Galactic Lexicon |

**Galactic Lexicon:** [see ADR 015](decisions/015-galactic-lexicon-and-achievements.md)
- Supports two modes: `standard` (technical) and `galactic` (space-themed metaphors)
- Categories: success, error, info, warning, achievement
- Locales: Russian (ru) and English (en)
- User-controlled via `galactic_mode` setting in user_settings table
- Integrated into all UI components (modals, toasts, error displays)

#### 2.5 Stores (`src/shared/stores/`)

| File | Purpose |
|------|---------|
| `auth.svelte.ts` | Authentication state |
| `auth-session.svelte.ts` | Auth session handling |
| `graph.svelte.ts` | Graph UI state (auto-links toggle and related flags) |
| `graph-view.svelte.ts` | Graph view mode (personal / community) |
| `lexicon-settings.ts` | Lexicon locale and mode settings |

Achievement state lives in `entities/achievement/`; note list state in `entities/note/`.

#### 2.6 3D Engine (`src/features/graph-3d/`)

The 3D graph is **active** — rendered by `features/graph-3d/` behind the
`widgets/graph-3d-viewer/` widget on `/graph/3d` and `/graph/3d/:id`. (This section earlier
described an older removed module as "frozen"; the current engine was built separately.)

| File | Purpose |
|------|---------|
| `lib/engine.ts` | Render engine: scene lifecycle, apply data, selection |
| `lib/scene.ts` | Three.js scene setup |
| `lib/camera.ts` | Camera controls |
| `lib/labels.ts` | `LabelManager` — selective node labels |
| `lib/links.ts`, `lib/nodes.ts` | Link and node rendering |
| `lib/fog.ts` | Fog parameters |
| `lib/simulation.ts` | Force layout in 3D space |
| `model/layout-provider.ts` | Layout data provider |
| `ui/Graph3DScene.svelte` | Scene component |

---

### 3. NLP Service (Python)

**Location:** `nlp-service/`

**Stack:** FastAPI + sentence-transformers + KeyBERT + pymorphy3 + NLTK

#### 3.1 API Endpoints (`app/main.py`)

```python
GET  /health              → {status, model_loaded, version}
POST /extract_keywords    → ExtractKeywordsResponse
POST /embed               → EmbedResponse
POST /normalize           → NormalizeResponse (keyword normalization)
POST /similarity          → SimilarityResponse (pairwise similarity)
```

#### 3.2 Models (`app/models.py`)

```python
ExtractKeywordsRequest:  {text: str, top_n: int, title: str}
ExtractKeywordsResponse: {extractor, keywords: [{keyword, surface, weight}]}
EmbedRequest:          {text: str, title: str}
EmbedResponse:         {embedding: float[], chunks?: int, no_content?: bool}
```

`chunks`/`no_content` are returned only when `EMBED_CHUNKING=on` (CHUNK-1):
`app/core/chunking.py` splits the note into structure-aware chunks, the service
encodes them in one batched call, and the document vector is the L2-normalized
mean. The note title is injected into every chunk's model input. Off is the
default and preserves the legacy single-encode behavior.

#### 3.3 NLP Utils (`app/nlp_utils.py`)

| Function | Library | Purpose |
|----------|---------|---------|
| `extract_keywords()` | KeyBERT-hybrid + pymorphy3/WordNet | Keyword extraction + lemmatization (RU/EN) |
| `embedding_model.encode()` | sentence-transformers | Text vectorization |

**Model:** `paraphrase-multilingual-MiniLM-L12-v2` (384 dimensions) via `NLP_MODEL_NAME`

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
note_keywords  — Keywords
tags           — Tags
note_tags      — Many-to-many
note_recommendations — Precomputed recommendations
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
3. `mongo_personal` — MongoDB 7 (port 27018) [see ADR 011](decisions/011-drafts-autosave-mongodb.md) — drafts
4. `nlp` — Python service (port 5001)
5. `graph-service-personal` — Graph service (port 9092) [see ADR 013](decisions/013-graph-service-isolation.md)
6. `backend_personal` — Go API (port 8080)
7. `worker_personal` — Background worker
8. `nginx_personal` — Reverse proxy (ports 18082, 18084)
9. `frontend_personal` — SvelteKit (port 3001)
10. `backup_scheduler` — Automatic backup service

#### 4.4 Backup Service

**Purpose:** Automatic PostgreSQL database backup with local storage and Yandex.Disk cloud backup support.

**Components:**

**Backup scripts:**
- `scripts/devops/backup-personal.sh` — Bash script for Linux/Mac
- `scripts/devops/backup-personal.ps1` — PowerShell script for Windows

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
   - Storage in the synced host folder (`~/Desktop/my items`; inside containers it is `/backups`, see `KG_BACKUP_DIR`)
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

**More details:** [`docs/operations/BACKUP.md`](../operations/BACKUP.md)

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
keybert                         # Keywords (semantic ranking)
pymorphy3                       # Russian lemmatization
nltk                            # English lemmatization + stopwords
```

---

## 🚀 Deployment

### Local Development

```bash
docker-compose up -d
```

(The `Makefile` has no `dev` target — only cleanup helpers; see `COMMANDS.md` for the real commands.)

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

- `docs/api/API_ERRORS_EN.md` — API errors and codes
- `ROADMAP.md` — Development plans (moved to project root)
- `WEIGHTS_CALCULATION.md` — Link weight calculation logic
- `docs/architecture/c4/` — C4 Model diagrams
- `docs/architecture/decisions/` — ADR (Architecture Decision Records)

### Key ADR (Architecture Decision Records)

- **[ADR 011: Drafts Autosave in MongoDB](decisions/011-drafts-autosave-mongodb.md)** — Autosave drafts in MongoDB with eventual sync to PostgreSQL
- **[ADR 013: Graph Service Isolation](decisions/013-graph-service-isolation.md)** — Extract Graph Service as separate microservice with gRPC and direct DB access
- **[ADR 014: Event-Driven Cache Invalidation](decisions/014-event-driven-cache-invalidation.md)** — Use Redis Pub/Sub for cache invalidation
- **[ADR 015: Galactic Lexicon and Achievements](decisions/015-galactic-lexicon-and-achievements.md)** — Unified galactic lexicon, i18n, SSE for achievement notifications
- **[ADR 016: Keyword Similarity Strategies](decisions/016-keyword-similarity-strategies.md)** — Strategy pattern for keyword similarity metrics with weight support
- **[ADR 017: Color Palette Redesign](decisions/017-color-palette-redesign.md)** — Black-purple-red color palette for space theme

---

## 🗺️ Module Dependency Graph

```
cmd/
├── server → interfaces/api → application/* → domain/* → infrastructure/*
├── worker → application/common (task_queue) → infrastructure/queue
└── cli    → infrastructure/db (migrations)

interfaces/api/
├── notehandler/    → application/* (note use cases), domain/note
├── linkhandler/    → domain/link
├── graphhandler/   → application/graph
├── taghandler/     → domain/note (tags)
└── handlers/*      → auth, user, settings, draft, share, backup, achievement

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
| Domain | 35 | Low (business logic) |
| Application | 24 | Medium (orchestration) |
| Infrastructure | 64 | High (technical details) |
| Interfaces | 27 | Medium (HTTP) |
| Frontend | 89 | Medium (UI) |
| NLP | 7 | Low (models) |

*(Non-test `.go` files per layer; frontend counted as `.svelte` files; NLP as `.py` under `app/` — 2026-09-26.)*

---

## 📈 Graph Service

### Overview

Graph Service is an independent microservice responsible for computing 2D/3D graph layouts for the Knowledge Graph frontend. It provides high-performance graph visualization with caching, incremental updates, and event-driven invalidation [see ADR 014](decisions/014-event-driven-cache-invalidation.md).

### Architecture

The Graph Service consists of:

- **API Layer**: HTTP API (port 9091 — the consumed interface) and gRPC server (port 9090 — implemented, currently without callers) [see ADR 013](decisions/013-graph-service-isolation.md)
- **Layout Engine**: 2D circular and 3D spiral layout algorithms with delta computation
- **Cache Layer**: Redis-backed caching with configurable TTL
- **Data Layer**: Direct PostgreSQL read access (notes, links, embeddings)
- **Event Subscriber**: Redis Pub/Sub for cache invalidation with acknowledgment tracking

### API Contracts

#### gRPC API [see ADR 013](decisions/013-graph-service-isolation.md)

```protobuf
service GraphService {
  rpc GetNoteLayout(NoteLayoutRequest) returns (LayoutResponse);
  rpc GetFullLayout(FullLayoutRequest) returns (stream LayoutChunk);
  rpc GetDelta(DeltaRequest) returns (DeltaResponse);
}
```

> Note (DOC-AUDIT-2, 2026-09-26): the gRPC server listens on :9090 and honors this contract,
> but **no client in the repository calls it** — backend and frontend use HTTP. The owner's
> intent is gRPC as an *overload fallback*, not a parallel primary transport; recorded in
> [`DOC-AUDIT-2`](../tasks/DOC-AUDIT-2-docs-vs-code.md) for decision.

#### HTTP API (the interface actually consumed)

```
GET /api/v1/graph/note/:id?depth=2&user_id={userId}
GET /api/v1/graph/full?limit=1000&user_id={userId}
GET /api/v1/graph/delta?last_hash={hash}&user_id={userId}
GET /api/v1/graph/public
GET /api/v1/graph/path
GET /api/v1/graph/recommendations
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

All keys live under the `graph-service:` prefix:

- `graph-service:full:{userId}` - Current full-layout pointer (5 min TTL). Events delete it.
- `graph-service:snapshot:{userId}:{dataHash}` - Immutable snapshot of the layout served under a data hash (15 min TTL, `CACHE_SNAPSHOT_TTL_SECONDS`). Events must NOT delete snapshots: a snapshot is the baseline the delta is computed against — the version the client actually holds, not whatever is cached as current.
- `graph-service:delta:{userId}:{lastHash}` - Cached delta response (1 min TTL).
- `graph-service:note:{userId}:{noteId}:depth-{d}` - Note neighborhood layout (5 min TTL).

### Delta Contract (SYNC-1)

- The graph version (`meta.hash`, `X-Layout-Hash`, `current_hash`) is a hash of the **served data** — sorted note/link fields — never of the computed layout, so a layout recalculation alone cannot roll the version forward.
- `GET /api/v1/graph/delta?last_hash={hash}` diffs the current data against the snapshot stored under `{hash}`. If no snapshot exists, the response is `{"resync": true}` and the client reloads the graph wholesale — it must never pretend the whole graph was added.
- Removals are explicit (`removed_nodes`, `removed_links`); a link that keeps its key but changes fields (weight, source_type, gamma_origin) is re-sent in `added_links` and replaces the old entry on the client.
- On gRPC the same contract applies; a missing snapshot is answered with `NotFound` instead of a JSON flag.
- Event publication is transactional (SYNC-1 stage A2): `infrastructure/outbox` decorators wrap the note/link repositories and insert a `graph_outbox` row in the same transaction as the write, so a write and its event commit or roll back together. A relayer in `cmd/worker` drains unsent rows to `graph:events` with at-least-once delivery (`FOR UPDATE SKIP LOCKED`, batched); sent rows are purged after `OUTBOX_SENT_RETENTION_DAYS` (30 days). Manual publishing outside the relay is forbidden — `scripts/testing/check-graph-write-paths.mjs` fails the build on a manual `Publish*` call or a repository constructed without the outbox decorator.

### Direct PostgreSQL Reading

Graph Service reads directly from PostgreSQL as the single source of truth, and delivery of change events is guaranteed by the transactional outbox above — direct reads do not replace the outbox because a read answers "what is" but cannot answer "what changed since the client's version" [see ADR 014](decisions/014-event-driven-cache-invalidation.md).

---

*Generated with ❤️ by Cascade*
