# Changelog - Knowledge Graph

> Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
> 
> Versioning follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html)

🌐 **Languages**: [English](CHANGELOG_EN.md) | [Русский](CHANGELOG.md)

---

## [Unreleased]

### Added
- Planned: Graph export (image/video)
- Planned: Node grouping in visualization
- Planned: Keyboard shortcuts for graph

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
- **Architecture**: C4 Model, UML diagrams, 14 ADRs
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
- **ADR-014**: Added Progressive Rendering decision description

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
- [README_EN.md](../README_EN.md) — central navigation
- [CONFIGURATION_EN.md](CONFIGURATION_EN.md) — configuration
- [DEPLOYMENT_EN.md](DEPLOYMENT_EN.md) — deployment
- [TESTING.md](TESTING.md) — testing
