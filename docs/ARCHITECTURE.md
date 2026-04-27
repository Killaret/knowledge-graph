# Knowledge Graph — Architecture Documentation

> **Актуальность:** Auto-generated via @codemaps:  
> **Дата:** 2026-04-27  
> **Стек:** Go + SvelteKit + Python (FastAPI) + PostgreSQL + Redis

---

## 📊 Общая Архитектура

Система построена на **Clean Architecture** с разделением на 4 слоя:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              PRESENTATION LAYER                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐ │
│  │  HTTP API    │  │  WebSocket   │  │  Static      │  │  E2E Tests         │ │
│  │  (chi)       │  │  (optional)  │  │  (frontend)  │  │  (Cucumber)        │ │
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

---

## 🎯 Компоненты Системы

### 1. Backend (Go)

**Расположение:** `backend/`

#### 1.1 Entry Points (`cmd/`)

| Команда | Файл | Назначение | Порт |
|---------|------|------------|------|
| `server` | `cmd/server/main.go` | HTTP API сервер | 8080 |
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

**Файлы:**
- `entity.go` — Aggregate root с бизнес-логикой
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

**Файлы:**
- `entity.go` — Entity с методом `UpdateWeight()`
- `value_objects.go` — LinkType, Weight, Metadata
- `repository.go` — Repository interface

##### Graph Domain (`domain/graph/`)

**Алгоритмы обхода графа:**
- `bfs.go` — Breadth-First Search
- `neighbor_loader.go` — Загрузка соседей узла
- `keyword_matcher.go` — Сопоставление по ключевым словам
- `traversal_service.go` — Сервис обхода с интеграцией репозиториев
- `normalizer.go` — Нормализация весов связей
- `aggregation.go` — Агрегация результатов обхода

**Тесты:**
- `traversal_test.go` — Unit tests
- `traversal_integration_test.go` — Integration tests

#### 1.3 Application Layer (`internal/application/`)

##### Graph Application (`application/graph/`)

| Файл | Назначение |
|------|------------|
| `composite_loader.go` | Композитный загрузчик (ключевые слова + эмбеддинги) |
| `embedding_loader.go` | Загрузка по векторной близости |
| `neighbor_loader.go` | Загрузка соседей через связи |
| `composite_loader_test.go` | Tests |

##### Recommendation Application (`application/recommendation/`)

| Файл | Назначение |
|------|------------|
| `refresh_service.go` | Обновление рекомендаций |
| `affected_notes.go` | Определение затронутых заметок |
| `*_test.go` | Unit tests |

##### Common (`application/common/`)

- `task_queue.go` — Abstraction over task queue

##### Queries (`application/queries/graph/`)

- `get_suggestions.go` — Query handler для получения рекомендаций

#### 1.4 Infrastructure Layer (`internal/infrastructure/`)

##### Database (`infrastructure/db/`)

**PostgreSQL Repositories (`db/postgres/`):**

| Файл | Сущность | CRUD | Специфика |
|------|----------|------|-----------|
| `note_repo.go` | Note | ✅ | Full-text search, pagination |
| `link_repo.go` | Link | ✅ | Cascade delete by source |
| `embedding_repo.go` | Embedding | ✅ | pgvector similarity search |
| `tag_repo.go` | Tag | ✅ | Many-to-many with notes |
| `user_repo.go` | User | ✅ | Auth data |
| `recommendation_repo.go` | Recommendation | ✅ | Suggestions storage |

**Models (`db/postgres/*_model.go`):**
- `note_model.go` — GORM model для Note
- `note_embedding_model.go` — Векторные эмбеддинги (pgvector)
- `note_keyword_model.go` — Извлечённые ключевые слова
- `link_model.go` — GORM model для Link
- `tag_model.go`, `note_tag_model.go` — Tagging system
- `recommendation_model.go` — Предвычисленные рекомендации

**Migrations:**
- `migrations.go` — Migration runner
- `../../migrations/` — 27 SQL файлов миграций

##### NLP Client (`infrastructure/nlp/`)

- `client.go` — HTTP client для NLP Service
- `client_test.go` — Tests with mocks

**Endpoints:**
- `POST /extract_keywords` — YAKE keyword extraction
- `POST /embed` — SentenceTransformers embeddings

##### Queue (`infrastructure/queue/`)

**Asynq (Redis-based task queue):**

| Файл | Назначение |
|------|------------|
| `asynq_client.go` | Client для enqueue задач |
| `worker.go` | Worker процессор |
| `tasks.go` | Task definitions |
| `tasks/recommendation.go` | Recommendation refresh task |

**Task Types:**
- `recommendation:refresh` — Обновление рекомендаций для ноты

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
```

---

### 2. Frontend (SvelteKit)

**Расположение:** `frontend/`

#### 2.1 Routes (`src/routes/`)

| Route | Файл | Назначение |
|-------|------|------------|
| `/` | `+page.svelte` | Главная страница со списком заметок |
| `/graph` | `graph/+page.svelte` | 2D интерактивный граф (D3.js) |
| `/graph/3d` | `graph/3d/+page.svelte` | 3D граф (Three.js) |
| `/graph/3d/:id` | `graph/3d/[id]/+page.svelte` | 3D граф с фокусом на ноте |
| `/graph/:id` | `graph/[id]/+page.svelte` | 2D граф с фокусом |
| `/notes/:id` | `notes/[id]/+page.svelte` | Просмотр заметки |
| `/notes/:id/edit` | `notes/[id]/edit/+page.svelte` | Редактирование |
| `/notes/new` | `notes/new/+page.svelte` | Создание заметки |
| `/search` | `search/+page.svelte` | Полнотекстовый поиск |

#### 2.2 Components (`src/lib/components/`)

**Core Components (46 total):**

| Компонент | Технология | Назначение |
|-----------|------------|------------|
| `GraphCanvas.svelte` | D3.js | 2D force-directed graph |
| `Graph3D.svelte` | Three.js | 3D celestial visualization |
| `LazyGraph3D.svelte` | dynamic import | Ленивая загрузка 3D |
| `SmartGraph.svelte` | D3+Svelte | Адаптивный граф (2D/3D) |
| `NoteCard.svelte` | Svelte | Карточка заметки |
| `NoteEditor.svelte` | Svelte | WYSIWYG редактор |
| `NoteSidePanel.svelte` | Svelte | Боковая панель деталей |
| `CreateNoteModal.svelte` | Svelte | Модал создания |
| `EditNoteModal.svelte` | Svelte | Модал редактирования |
| `SearchBar.svelte` | Svelte | Поисковая строка |
| `Sidebar.svelte` | Svelte | Навигация |
| `FloatingControls.svelte` | Svelte | Плавающие кнопки управления |
| `BackButton.svelte` | Svelte | Навигация назад |
| `ConfirmModal.svelte` | Svelte | Подтверждение действий |
| `Modal.svelte` | Svelte | Базовый модал |
| `Button.svelte` | Svelte | UI Button |
| `TypeSelector.svelte` | Svelte | Выбор типа ноды (star/planet/moon) |

#### 2.3 API Client (`src/lib/api/`)

| Файл | Назначение |
|------|------------|
| `client.ts` | ky instance конфигурация |
| `notes.ts` | Notes API (CRUD + search + suggestions) |
| `links.ts` | Links API |
| `graph.ts` | Graph data API |

#### 2.4 3D Engine (`src/lib/three/`)

| Файл | Назначение |
|------|------------|
| `scene.ts` | Three.js scene setup |
| `camera.ts` | Camera controls |
| `renderer.ts` | WebGL renderer |
| `graph3d.ts` | 3D graph visualization logic |
| `celestial.ts` | Celestial body rendering (stars, planets) |
| `controls.ts` | OrbitControls wrapper |
| `animation.ts` | Animation loop |
| `types.ts` | TypeScript types |

#### 2.5 State Management (`src/lib/stores/`)

- `notes.ts` — Svelte store для нот
- `graph.ts` — Store состояния графа
- `ui.ts` — UI state (модалы, выбор)

---

### 3. NLP Service (Python)

**Расположение:** `nlp-service/`

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

| Функция | Библиотека | Назначение |
|---------|------------|------------|
| `extract_keywords()` | YAKE | Извлечение ключевых слов (RU/EN) |
| `embedding_model.encode()` | sentence-transformers | Векторизация текста |

**Model:** `all-MiniLM-L6-v2` (384 dimensions)

#### 3.4 Tests (`tests/`)

- `test_api.py` — FastAPI endpoint tests
- `test_nlp_utils.py` — NLP function tests

---

### 4. Infrastructure Services

#### 4.1 PostgreSQL (pgvector)

**Docker:** `pgvector/pgvector:pg16`

**База:** `knowledge_base`
**User:** `kb_user`

**Extensions:**
- `pgvector` — Векторные операции
- `pg_trgm` — Trigram search

**Таблицы:**
```sql
notes          — Заметки
links          — Связи между заметками
note_embeddings — Векторы (384 dim)
keywords       — Ключевые слова
tags           — Теги
note_tags      — Many-to-many
recommendations — Предвычисленные рекомендации
users          — Пользователи
```

#### 4.2 Redis

**Назначение:**
- Task queue backend (Asynq)
- Cache layer (опционально)

#### 4.3 Docker Compose

**Services:**
1. `postgres` — pgvector (порт 5432)
2. `redis` — Redis 7 (порт 6379)
3. `nlp` — Python service (порт 5000)
4. `backend` — Go API (порт 8080)
5. `worker` — Background worker
6. `frontend` — SvelteKit (порт 3000)

---

## 🔄 Data Flow

### Создание заметки с рекомендациями

```
1. User создаёт заметку (Frontend)
   ↓ POST /notes
2. HTTP Handler принимает запрос
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

### Запрос рекомендаций

```
GET /notes/:id/suggestions
   ↓
Query Handler: GetSuggestions
   ├─ Check precomputed recommendations
   ├─ If stale/empty → trigger refresh
   └─ Return top-N suggestions
```

### Поиск по графу

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

| Тип | Локация | Фреймворк |
|-----|---------|-----------|
| Unit | `*_test.go` (рядом с кодом) | Go testing |
| Integration | `*_integration_test.go` | Go testing + testcontainers |
| E2E | `tests/features/` | Cucumber + Playwright |

### Frontend Tests

| Тип | Локация | Фреймворк |
|-----|---------|-----------|
| Unit | `*.test.ts` | Vitest |
| Component | `*.spec.ts` | Testing Library |
| E2E | `tests/` | Playwright |

### NLP Tests

| Тип | Локация | Фреймворк |
|-----|---------|-----------|
| Unit | `tests/test_nlp_utils.py` | pytest |
| API | `tests/test_api.py` | pytest + FastAPI TestClient |

---

## 📦 Dependencies

### Backend (go.mod)

```
github.com/go-chi/chi/v5        # HTTP router
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
# Или:
make dev
```

### Production Considerations

- **Database:** Connection pooling (pgx pool)
- **Cache:** Redis cluster для high availability
- **Queue:** Horizontal scaling workers
- **NLP:** GPU instances для embeddings
- **Frontend:** CDN для static assets

---

## 🔐 Security

- **CORS:** Configured for frontend origin
- **SQL Injection:** GORM + parameterized queries
- **XSS:** Svelte auto-escaping
- **Input Validation:** Validator на всех слоях

---

## 📚 Дополнительная Документация

- `API_ERRORS.md` — Ошибки API и коды
- `ARCHITECTURE_ROADMAP.md` — Планы развития
- `WEIGHTS_CALCULATION.md` — Логика расчёта весов связей
- `docs/architecture/c4/` — C4 Model диаграммы
- `docs/architecture/decisions/` — ADR (Architecture Decision Records)

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

| Компонент | Файлы | Сложность |
|-----------|-------|-----------|
| Domain | 19 | Низкая (бизнес-логика) |
| Application | 10 | Средняя (оркестрация) |
| Infrastructure | 38 | Высокая (технические детали) |
| Interfaces | 16 | Средняя (HTTP) |
| Frontend | 46 | Средняя (UI) |
| NLP | 4 | Низкая (модели) |

---

*Generated with ❤️ by Cascade*
