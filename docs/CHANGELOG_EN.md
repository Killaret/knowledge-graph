# Changelog - Knowledge Graph

> Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
>
> Versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html)

🌐 **Language:** English

---

## [Unreleased]

### 🚀 Graph Service Configuration

- **Unified Configuration for graph-service**: All parameters now sourced from `knowledge-graph.config.json`
  - New `graph_service` section with layout engine parameters
  - Cache TTLs: note/full layouts (5 min), deltas (1 min)
  - Layout constants: 2D radius (100), 3D radius (120), 3D Z-step (5)
  - Streaming: chunk size (100 nodes per chunk)
  - Event tracking: TTL (24h), retry interval (5 min)
- **Environment Variable Overrides**: Full support for runtime configuration
  - `GRPC_PORT`, `HTTP_PORT`, `GRAPH_FULL_LIMIT`, `GRAPH_DEFAULT_DEPTH`
  - `CACHE_NOTE_TTL_SECONDS`, `CACHE_FULL_TTL_SECONDS`, `CACHE_DELTA_TTL_SECONDS`
- **Docker Integration**: Updated `Dockerfile` and `docker-compose.yml` to mount config
  - Config loaded from `/app/knowledge-graph.config.json` in container
  - Build context changed to project root for config access
- **Documentation**: Added Graph Service section to `docs/CONFIGURATION_EN.md` and `docs/CONFIGURATION_RU.md`

### 🎮 Galactic Lexicon & Achievements System
- **Galactic Lexicon**: Themed messaging system with two modes
  - `standard` mode: Technical, straightforward messages
  - `galactic` mode: Space-themed metaphors (e.g., "Ignite New Star" instead of "Create Note")
  - Categories: success, error, info, warning, achievement
  - Locales: Russian (ru) and English (en)
  - User-controlled via `galactic_mode` setting in user_settings table
  - Integrated into all UI components: CreateNoteModal, EditNoteModal, ConfirmModal, ShareModal, ApiErrorDisplay, ToastNotification
- **Achievements System**: Gamification with unlockable achievements
  - Backend domain layer with Achievement and UserAchievement entities
  - AchievementEngine for condition evaluation (count, streak types)
  - AchievementService with CheckTrigger (sync) and CheckStreaks (async)
  - API endpoints: GET /api/v1/achievements, GET /api/v1/users/me/achievements, POST /api/v1/users/me/achievements/:id/mark-seen
  - Achievement triggers integrated with note creation and link creation
  - ToastNotification integration for achievement unlocks
  - Respects user setting `show_achievement_notifications`
  - Frontend polling for new achievements (configurable via `frontend.achievements.poll_interval_ms`)
  - Database migrations: 017_create_user_settings, 018_create_achievements, 021_create_achievements_tables
- **Configuration**: Achievement polling interval moved to `knowledge-graph.config.json` under `frontend.achievements.poll_interval_ms` (default: 7000ms)

### 🚫 Feature Freezing
- **3D Graph Frozen**: 3D graph functionality temporarily frozen for v1.0
  - **Reason**: Improve stability and reduce maintenance overhead
  - **Impact**: All 3D components disabled, routes redirected to 2D
  - **Tests**: 3D tests marked as skipped with "3D feature frozen for v1" annotation
  - **Components Affected**: `Graph3D.svelte`, `LazyGraph3D.svelte`, `SmartGraph.svelte`
  - **Routes**: `/graph/3d` and `/graph/3d/:id` redirect to `/graph` and `/graph/:id`
  - **UI**: 3D toggle buttons hidden in FloatingControls component
  - **Bundle Size**: Three.js and 3D dependencies excluded from build
  - **Future**: 3D functionality preserved in codebase for potential reactivation in future versions

### ⚙️ Unified Configuration System
- **Single Source of Truth**: New `knowledge-graph.config.json` at project root
  - All structural parameters in one place
  - Shared between backend, frontend, and NLP service
  - Priority: ENV vars > JSON config > hardcoded defaults
- **Frontend Config Module**: `frontend/src/shared/config/config.ts` imports JSON directly
  - Exports: `graphConfig2D`, `graphConfig3D`, `apiConfig`, `testConfig`, `ciCdConfig`
  - Type-safe TypeScript interfaces
- **Backend Config Updates**: `backend/internal/config/config.go`
  - JSON loading with fallback paths
  - New sections: `server`, `database`, `search`, `pagination`
  - Extended: `graph` (limits), `recommendation` (batch_rate_limit)

### 🎯 Flexible Keyword Similarity in Recommendations
- **5 Similarity Strategies**: Jaccard, Overlap, Tversky, Weighted Jaccard, Cosine
  - Configurable via `knowledge-graph.config.json`:
    - `keyword_similarity_method` — similarity strategy (default: "jaccard")
    - `keyword_tversky_alpha` — alpha parameter for Tversky index (default: 0.5)
    - `keyword_tversky_beta` — beta parameter for Tversky index (default: 0.5)
  - Strategy implementations in `backend/internal/application/recommendation/keyword_similarity.go`
  - Factory pattern: `NewKeywordSimilarity(method, alpha, beta)` for strategy creation
- **Integration with TraversalService**:
  - `KeywordMatcher` interface in domain layer (`backend/internal/domain/graph/keyword_matcher.go`)
  - `KeywordMatcherImpl` implementation with KeywordRepository and similarity strategy
  - Worker initialization in `backend/cmd/worker/main.go` sets up matcher based on config
  - Keyword component enabled when `gamma > 0` in recommendation weights
- **Weighted Scoring**:
  - Supports both weighted and unweighted similarity strategies
  - Weighted strategies use keyword weights from `KeywordRepositoryWithWeights`
  - Automatic fallback to unit weights when weights unavailable
- **Documentation Updates**:
  - README.md: Added "Гибкое сходство ключевых слов в рекомендациях" section
  - docs/ARCHITECTURE_EN.md: Added keyword similarity architecture subsection
  - docs/RECOMMENDATION_ARCHITECTURE.md: Added comprehensive keyword similarity component documentation


### � Security
- **SQL Injection Fix**: Parameterized queries in note search ORDER BY clause (`note_repo.go`)
- **Rate Limiting**: Token bucket middleware with config-driven limits
  - General: 100 req/min per IP (configurable)
  - Per-endpoint limits for write operations
  - Conditional enable/disable via config

### 🚀 Added
- **Yandex.Disk Backup System**: Automatic backup system for personal instance
  - Local backup scripts: `scripts/devops/backup-personal.sh` (Linux/Mac) and `scripts/devops/backup-personal.ps1` (Windows)
  - Go service: `backend/internal/infrastructure/cloud/yandex_backup.go` for Yandex.Disk WebDAV integration
  - Asynq task: `TypeBackupToCloud` for async cloud backup uploads
  - Docker service: `backup_scheduler` in `docker-compose.personal.yml` for daily scheduled backups
  - Configuration: `knowledge-graph.config.json` section `backup` with `provider=yandex`
  - Automatic cleanup of old backups (7 days local, 10 backups in cloud)
  - Retry logic for failed uploads (3 attempts)
  - Documentation: [`BACKUP.md`](BACKUP.md) with comprehensive setup guide
- **Comprehensive Health Check**: `/health` endpoint now checks all dependencies
- **Graph API Pagination**: `/graph/all` supports DB-level pagination
- **FindAllPaginated**: New repository methods with metadata
- **Redis Error Handling**: Graceful degradation
- **NLP Health Check**: Model verification endpoint
- **Unknown Node Type**: Fallback visualization for nodes without type
  - Displays as question mark (?) in dashed circle
  - Represents conditional/indeterminate object of any shape
  - Falls back to 'unknown' when node.type is null/undefined/empty

### 🔧 Changed
- **Configuration Architecture**: Centralized all parameters in `knowledge-graph.config.json`
  - Removed hardcoded constants from `main.go` (rate limits, retry delays)
  - `graph_handler.go` uses `cfg.GraphDefaultLimit/MaxLimit`
  - `note_service.go` uses `cfg.PaginationMaxLimit/DefaultLimit`
  - Frontend uses `graphConfig2D.shadows_threshold`, `apiConfig.default_limit`
- **Error Handling**: Graceful degradation
  - Database retry with configurable delay from JSON
  - Server restart on fallback ports from config
  - Migration failures log warning but don't crash

### 🧪 Fixed
- **Integration Tests**: Added missing models to test migrations
- **SearchBar Tests**: Enabled previously skipped tests
- **API Documentation**: Updated OpenAPI spec with pagination schema
- **Configuration Consistency**: All hardcoded numbers now sourced from config
  - Graph limits: 500/1000 nodes, 500/5000 links
  - Pagination: 20 default, 100 max
  - Shadows threshold: 100 nodes (2D graph)
  - `GraphNode` and `GraphLink` schemas
- **Input Validation**: Structured validation with human-readable errors:
  - **Notes**: title (required, max 200), content (max 50000), type (oneof: star, planet, moon, comet, galaxy, nebula, asteroid, satellite, blackhole, unknown)
  - **Links**: source/target UUID validation, link_type (oneof: reference, dependency, related, custom), weight (0-1)
  - Validation errors return structured JSON: `{"error": "validation_failed", "message": "..."}`
- **Test Fixes**: Updated `graph_handler_test.go` for pagination support:
  - `setupGraphRouter()` now creates complete config with pagination limits
  - Mock expectations updated for `FindAllPaginated` with proper `int64` total counts
  - All 11 handler tests now passing

### Planned
- Graph export (image/video)
- Node grouping in visualization
- Keyboard shortcuts for graph

### 🚀 Phase 4: Explorer Update (feature/explorer-update)

#### 🚀 Spaceship Navigator
- **Interactive starship** (~40x40px SVG) that follows cursor, points at nodes, drifts idly
- **3 modes**: FOLLOW_CURSOR (0.3s delay), POINT_AT_NODE (near active node), IDLE_DRIFT (random movement)
- **GalacticLexicon tooltips**: "Новая звезда нанесена на карту, капитан!" (note), "Гравитационный луч установлен!" (link), "Кодекс исследователя пополнен!" (achievement)
- **Loading animation**: Patrol orbit on empty space, camera zoom 0.5→1.0
- **Segmented node loading**: Waves (0-200ms, 200-500ms, 500-800ms) with ship flying between appearing nodes

#### 🏆 Points & Cosmetics System
- **Points for actions**: CreateNote(+1), CreateLink(+2), QuickCapture(+1), PublishNote(+5)
- **Daily bonus streaks**: 2-day +2, 7-day +10 (Asynq task)
- **Cosmetics shop**: Spaceship skins, engines, trails, satellites
- **Database**: `user_points`, `point_transactions`, `user_cosmetics` tables
- **API**: `GET /users/me/points`, `GET /cosmetics`, `POST /cosmetics/buy`, `GET /users/me/cosmetics`

#### 📥 Import/Export Notes
- **ImportFromJSON**: Parse JSON, create notes and links
- **ExportToJSON**: All user notes in JSON format
- **ExportToMarkdown**: Obsidian-compatible .zip (YAML frontmatter + [[wikilinks]])
- **ExportToCSV**: CSV with id, title, type, created_at, links_count
- **API**: `POST /import`, `GET /export?format=json|markdown|csv`, `GET /notes/{id}/export?format=md`

---

## [1.0.0] - 2026-04-15

### ✨ Highlights

**3D Progressive Rendering with Fog of War**
- Incremental node loading with animation
- Three.js modular architecture (core/simulation/rendering/camera)
- Smooth camera transitions (lerpCamera)
- Auto-zoom on graph load

**Complete Testing Stack**
- Backend: 25+ unit tests (~1100 lines of Go)
- Frontend: 48 E2E tests with Playwright
- 3D Graph: WebGL, fog animation, progressive loading tests
- **BDD: 5 feature files with Cucumber/Gherkin** and comprehensive documentation

**Bilingual Documentation**
- English versions of key documents (README, CONFIGURATION, DEPLOYMENT)
- Language switchers in all Russian documents

### 🚀 Added

#### Backend
- **Domain Layer**: Entities (Note, Link), Value Objects (Title, Content), Graph Traversal with MAX strategy
- **Application Layer**: Composite Neighbor Loader with weighted aggregation (α=0.5, β=0.5)
- **Infrastructure Layer**: PostgreSQL repositories, Redis recommendation caching
- **Interface Layer**: HTTP handlers (notes, links, graph) with Gin
- **Worker**: asynq task processor for NLP embeddings
- **Graph Algorithms**: BFS traversal (depth=3, decay=0.5), semantic search via pgvector

#### Frontend
- **Note Management**: CRUD operations, search, note types (star, planet, comet, galaxy)
- **2D Graph**: D3-force visualization, Canvas rendering
- **3D Graph**: Three.js + d3-force-3d, progressive rendering, fog of war
- **Progressive Loading**: Batches of 5 nodes, opacity animation, auto-zoom
- **API Client**: ky-based HTTP client with Vite proxy

#### NLP Service
- Keyword extraction using KeyBERT
- Text embeddings via sentence-transformers (all-MiniLM-L6-v2, 384 dim)
- FastAPI endpoints: `/extract_keywords`, `/embeddings`, `/health`
- Model caching in `/app/cache`

#### Infrastructure
- PostgreSQL 16 + pgvector extension
- Redis 7 (asynq queues + cache)
- Docker Compose for local development
- Kubernetes manifests (production-ready)

#### Documentation
- **Architecture**: C4 Model, UML diagrams, 1 ADR (layered architecture)
- **Configuration**: Complete env variable description with code verification
- **API**: OpenAPI 3.1 specification (fixed to match implementation)
- **Frontend Arch**: Three.js modules, Progressive Rendering
- **Deployment**: Complete deployment guide (RU + EN)
- **Testing**: Testing strategy, BDD, execution
- **API Errors**: Error reference guide
- **BDD Features**: 5 feature files with detailed documentation:
  - `graph_navigation.feature` — graph navigation
  - `note_management.feature` — note management
  - `search_and_discovery.feature` — search and discovery
  - `graph_view.feature` — 2D/3D modes
  - `import_export.feature` — import/export

### 🔧 Changed

- **Note Type Handling**: Fixed `type` passing in `CreateNoteModal`
- **Link Schema**: Field `description` → `link_type` + added timestamps
- **Routing**: Note-centric navigation instead of graph-first
- **Default Values**: Fixed default values (Alpha=0.5, Beta=0.5 as in code)

### ⚡ Performance

- **Recommendation Caching**: Redis TTL 300 seconds for recommendations
- **Progressive Graph Loading**: Nodes load in batches without blocking UI
- **WebGL Optimization**: Device capability detection, performance mode toggle
- **Database**: IVFFlat indexes for pgvector similarity search

### 🐛 Fixed

- **OpenAPI**: Fixed discrepancies with implementation:
  - Link schema (added `link_type`, `created_at`, `updated_at`)
  - Error schema (fields `error`, `message`, `detail`, `code`)
  - Added endpoints: `/notes/search`, `/graph/all` with `limit` parameter
  - Added tags with descriptions
- **CONFIGURATION.md**: Complete verification and fixes:
  - Fixed default values (Alpha=0.5, Beta=0.5 as in code)
  - Parameters Gamma, BFS_AGGREGATION, ASYNQ_CONCURRENCY marked as reserved
  - Added "Component" column for all variables
  - Improved example formatting
- **Bilingual Documentation**: Created English versions of key documents:
  - `README_EN.md` — English version of main README
  - `docs/CONFIGURATION_EN.md` — English configuration guide
  - `docs/DEPLOYMENT_EN.md` — English deployment guide
  - Added language switchers to Russian versions
- **architecture/README.md**: Added Backend Architecture section with diagrams
- **Frontend Arch**: Updated documentation for 3D modules
- **ADR-001**: Layered architecture decision description

### 🧪 Testing

| Component | Coverage | Details |
|-----------|----------|---------|
| Backend Domain | 85% | 25+ tests |
| Backend Application | 75% | Composite loader |
| Backend Interface | 70% | HTTP handlers |
| Frontend E2E | 48 tests | Playwright |
| 3D Graph | Full | WebGL, fog, animation |
| **BDD Features** | 5 feature files | Cucumber + Gherkin |

**BDD Documentation:**
- `tests/features/README.md` — comprehensive BDD guide (English)
- Step definitions in TypeScript with Playwright
- Tag support: @smoke, @regression, @slow
- CI/CD integration

---

## [0.9.0] - 2026-03-15

### Added
- Basic 3D visualization via Three.js
- d3-force-3d integration for physics
- Initial progressive loading support

### Changed
- Frontend structure refactoring
- Upgrade to Svelte 5

---

## [0.5.0] - 2026-02-01

### Added
- Basic CRUD for notes
- 2D graph via D3.js
- PostgreSQL with pgvector
- Redis for caching

### Infrastructure
- Docker Compose setup
- Basic migrations

---

## Version History

| Version | Date | Key Changes |
|---------|------|-------------|
| 1.0.0 | 2026-04-15 | 🎉 Production ready, 3D rendering, bilingual docs |
| 0.9.0 | 2026-03-15 | 3D visualization foundation |
| 0.5.0 | 2026-02-01 | MVP with 2D graph |

---

## Roadmap Comparison

| Planned | Status | Notes |
|---------|--------|-------|
| 3D Progressive Rendering | ✅ Complete | Fog of War, incremental loading |
| BDD Testing | ✅ Complete | 5 feature files, full documentation |
| Bilingual Documentation | ✅ Complete | RU + EN for key documents |
| Graph export | 🚧 Planned | PNG/SVG/MP4 |
| Node grouping | 🚧 Planned | Node clustering |

---

## Upgrade Guide

### From 0.9.0 to 1.0.0

```bash
# Configuration update
# Check your .env file — default values changed:
# RECOMMENDATION_ALPHA=0.5 (was 0.7)
# RECOMMENDATION_BETA=0.5 (was 0.3)

# Apply migrations
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" up

# Restart
docker-compose up -d
```

---

## Acknowledgments

- Three.js community for excellent documentation
- Asynq for reliable task queue
- pgvector for vector search in PostgreSQL

---

**Full Documentation:**
- [README.md](../README.md) — central navigation
- [CONFIGURATION_EN.md](CONFIGURATION_EN.md) — configuration
- [DEPLOYMENT_EN.md](DEPLOYMENT_EN.md) — deployment
- [TESTING.md](TESTING.md) — testing
