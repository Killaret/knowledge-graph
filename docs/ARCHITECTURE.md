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

### Redis и кэш приватного графа [см. ADR 014](architecture/decisions/014-event-driven-cache-invalidation.md)
- Redis используется для кэширования приватного графа пользователя рядом с уже существующими токенами, сессиями и данными очередей.
- Маршрут `/api/v1/me/graph/cached` возвращает мгновенный граф из Redis.
- Маршрут `/api/v1/me/graph/fresh` вычисляет актуальный граф и может возвращать `delta` для инкрементальных обновлений UI.
- Фронтенд гарантирует мгновенное отображение через кеш и плавное обновление Canvas без полной перерисовки.

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

##### Achievement Domain (`domain/achievement/`) [см. ADR 015](architecture/decisions/015-galactic-lexicon-and-achievements.md)

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

**Файлы:**
- `entity.go` — Achievement and UserAchievement entities
- `engine.go` — Condition evaluation logic
- `repository.go` — Repository interface

##### Draft Domain (`domain/draft/`) [см. ADR 011](architecture/decisions/011-drafts-autosave-mongodb.md)

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

**Файлы:**
- `entity.go` — Draft entity with TTL logic
- `repository.go` — Repository interface (MongoDB implementation)
- `service.go` — Autosave coordination service

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

##### Keyword Similarity Architecture

**Расположение кода:**
- `backend/internal/application/recommendation/keyword_similarity.go` — Стратегии сходства [см. ADR 016](architecture/decisions/016-keyword-similarity-strategies.md)
- `backend/internal/application/recommendation/keyword_matcher_impl.go` — Реализация matcher
- `backend/internal/domain/graph/keyword_matcher.go` — Интерфейс в domain слое

**Конфигурация** (`knowledge-graph.config.json`):
```json
{
  "backend": {
    "recommendation": {
      "keyword_similarity_method": "jaccard",  // jaccard, overlap, tversky, weighted_jaccard, cosine
      "keyword_tversky_alpha": 0.5,             // Параметр alpha для Tversky
      "keyword_tversky_beta": 0.5,              // Параметр beta для Tversky
      "gamma": 0.2                             // Вес keyword компонента (включает функцию при > 0)
    }
  }
}
```

**Интеграция** (`backend/cmd/worker/main.go` → `TraversalService`):
```go
// 1. Создание стратегии из конфигурации
keywordSimilarity, err := recommendation.NewKeywordSimilarity(
    cfg.RecommendationKeywordSimilarityMethod,
    cfg.RecommendationKeywordTverskyAlpha,
    cfg.RecommendationKeywordTverskyBeta,
)

// 2. Создание matcher с репозиторием ключевых слов
keywordMatcher := recommendation.NewKeywordMatcherImpl(keywordRepo, keywordSimilarity)

// 3. Настройка TraversalService
traversalSvc := graphDomain.NewTraversalServiceWithWeights(...)

// 4. Установка matcher если gamma > 0
if cfg.RecommendationGamma > 0 {
    traversalSvc.SetKeywordMatcher(keywordMatcher)
}
```

**Доступные стратегии:**
| Стратегия | Описание | Требует веса |
|-----------|----------|--------------|
| `jaccard` | Классический коэффициент Жаккара: \|A ∩ B\| / \|A ∪ B\| | Нет |
| `overlap` | Коэффициент перекрытия: \|A ∩ B\| / min(\|A\|, \|B\|) | Нет |
| `tversky` | Индекс Тверски с параметрами alpha/beta: \|A ∩ B\| / (\|A ∩ B\| + α\|A\\B\| + β\|B\\A\|) | Нет |
| `weighted_jaccard` | Взвешенный Жаккард: sum(min(w1, w2)) / sum(max(w1, w2)) | Да |
| `cosine` | Косинусное сходство векторов весов | Да |

**Поток данных:**
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
| `keyword_similarity.go` | Стратегии сходства ключевых слов (Jaccard, Overlap, Tversky, Weighted Jaccard, Cosine) [см. ADR 016](architecture/decisions/016-keyword-similarity-strategies.md) |
| `keyword_matcher_impl.go` | Реализация KeywordMatcher с использованием KeywordSimilarity |
| `*_test.go` | Unit tests |

##### Achievement Application (`application/achievement/`) [см. ADR 015](architecture/decisions/015-galactic-lexicon-and-achievements.md)

| Файл | Назначение |
|------|------------|
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
| `achievement_repo.go` | Achievement | ✅ | User achievement tracking |
| `user_settings_model.go` | UserSettings | ✅ | User preferences (galactic_mode, etc.) |

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
- `backup:cloud` — Загрузка бэкапа в облачное хранилище

##### Cloud (`infrastructure/cloud/`)

**Yandex.Disk Backup Service:**

| Файл | Назначение |
|------|------------|
| `yandex_backup.go` | YandexBackupService для работы с Яндекс.Диск через WebDAV |

**Методы:**
- `UploadBackup()` — Загрузка бэкапа с retry логикой
- `DownloadBackup()` — Скачивание бэкапа
- `ListBackups()` — Список бэкапов в облаке
- `DeleteBackup()` — Удаление бэкапа

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

| Компонент | Технология | Назначение | Статус |
|-----------|------------|------------|--------|
| `GraphCanvas.svelte` | D3.js | 2D force-directed graph | Active |
| `Graph3D.svelte` | Three.js | 3D celestial visualization [см. ADR 017](architecture/decisions/017-color-palette-redesign.md) | **Frozen** |
| `LazyGraph3D.svelte` | dynamic import | Ленивая загрузка 3D | **Frozen** |
| `SmartGraph.svelte` | D3+Svelte | Адаптивный граф (2D/3D) | 2D only |
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
| `ToastNotification.svelte` | Svelte | Ephemeral notifications with galactic mode support |
| `ApiErrorDisplay.svelte` | Svelte | Error display with lexicon integration |
| `ShareModal.svelte` | Svelte | Share modal with lexicon integration |

#### 2.3 API Client (`src/lib/api/`)

| Файл | Назначение |
|------|------------|
| `client.ts` | ky instance конфигурация |
| `notes.ts` | Notes API (CRUD + search + suggestions) |
| `links.ts` | Links API |
| `graph.ts` | Graph data API |
| `achievements.ts` | Achievements API |

#### 2.4 Utilities (`src/lib/utils/`)

| Файл | Назначение |
|------|------------|
| `galactic-lexicon.ts` | Galactic Lexicon - themed messaging system |
| `galactic-lexicon.test.ts` | Unit tests for Galactic Lexicon |

**Galactic Lexicon:** [см. ADR 015](architecture/decisions/015-galactic-lexicon-and-achievements.md)
- Supports two modes: `standard` (technical) and `galactic` (space-themed metaphors)
- Categories: success, error, info, warning, achievement
- Locales: Russian (ru) and English (en)
- User-controlled via `galactic_mode` setting in user_settings table
- Integrated into all UI components (modals, toasts, error displays)

#### 2.5 Stores (`src/lib/stores/`)

| Файл | Назначение |
|------|------------|
| `auth.svelte.js` | Authentication state |
| `lexicon-settings.ts` | Lexicon locale and mode settings |
| `achievements.ts` | Achievement polling and notification state |

#### 2.6 3D Engine (`src/lib/three/`) - **FROZEN for v1.0**

> **🚫 FROZEN FEATURE:** 3D graph functionality has been temporarily frozen for version 1.0 to improve stability and reduce maintenance overhead. See CHANGELOG.md for details.

| Файл | Назначение | Статус |
|------|------------|--------|
| `scene.ts` | Three.js scene setup | Frozen |
| `camera.ts` | Camera controls | Frozen |
| `renderer.ts` | WebGL renderer | Frozen |
| `graph3d.ts` | 3D graph visualization logic | Frozen |
| `celestial.ts` | Celestial body rendering (stars, planets) [см. ADR 017](architecture/decisions/017-color-palette-redesign.md) | Frozen |
| `controls.ts` | OrbitControls wrapper | Frozen |
| `animation.ts` | Animation loop | Frozen |
| `types.ts` | TypeScript types | Frozen |

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

**Services (dev stack):**
1. `postgres` — pgvector (порт 5432)
2. `redis` — Redis 7 (порт 6379)
3. `nlp` — Python service (порт 5000)
4. `backend` — Go API (порт 8080)
5. `worker` — Background worker
6. `frontend` — SvelteKit (порт 3000)

**Services (personal stack):**
1. `postgres_personal` — pgvector (порт 5433)
2. `redis_personal` — Redis 7 (порт 6380)
3. `mongo_personal` — MongoDB 7 (порт 27018) [см. ADR 011](architecture/decisions/011-drafts-autosave-mongodb.md) — черновики
4. `nlp` — Python service (порт 5001)
5. `graph-service-personal` — Graph service (порт 9092) [см. ADR 013](architecture/decisions/013-graph-service-isolation.md)
6. `backend_personal` — Go API (порт 8081)
7. `worker_personal` — Background worker
8. `nginx_personal` — Reverse proxy (порты 8082, 8083)
9. `frontend_personal` — SvelteKit (порт 3001)
10. `backup_scheduler` — Automatic backup service

#### 4.4 Backup Service

**Назначение:** Автоматическое резервное копирование базы данных PostgreSQL с поддержкой локального хранения и облачного бэкапа на Яндекс.Диск.

**Компоненты:**

**Скрипты бэкапа:**
- `scripts/utility/backup-personal.sh` — Bash скрипт для Linux/Mac
- `scripts/utility/backup-personal.ps1` — PowerShell скрипт для Windows

**Go-сервис:**
- `backend/internal/infrastructure/cloud/yandex_backup.go` — YandexBackupService для работы с Яндекс.Диск через WebDAV API

**Asynq задача:**
- `TypeBackupToCloud` — Асинхронная задача для загрузки бэкапов в облако

**Docker сервис:**
- `backup_scheduler` — Автоматический запуск бэкапов каждые 24 часа (в docker-compose.personal.yml)

**Функциональность:**
1. **Локальный бэкап:**
   - pg_dump базы PostgreSQL
   - Сжатие gzip
   - Хранение в `./backups/`
   - Автоматическая очистка старых бэкапов (по умолчанию 7 дней)

2. **Облачный бэкап (Яндекс.Диск):**
   - Загрузка через WebDAV API
   - OAuth аутентификация
   - Хранение в `/KnowledgeGraphBackups/`
   - Автоматическая очистка (max_backups, по умолчанию 10)
   - Retry логика (3 попытки)

3. **Конфигурация:**
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

**Методы YandexBackupService:**
- `UploadBackup(ctx, localPath, remoteKey)` — Загрузка бэкапа с retry логикой
- `DownloadBackup(ctx, remoteKey, localPath)` — Скачивание бэкапа
- `ListBackups(ctx, prefix)` — Список бэкапов в облаке
- `DeleteBackup(ctx, remoteKey)` — Удаление бэкапа
- `ensureFolder(ctx, folderURL)` — Создание папки на Яндекс.Диске
- `cleanupOldBackups(ctx)` — Очистка старых бэкапов

**Переменные окружения:**
- `BACKUP_CLOUD_ENABLED` — Включить облачный бэкап
- `BACKUP_YANDEX_TOKEN` — OAuth токен Яндекс.Диска
- `BACKUP_YANDEX_FOLDER` — Папка на Яндекс.Диске
- `BACKUP_DIR` — Локальная папка для бэкапов
- `CLEANUP_OLD_BACKUPS` — Очистка старых бэкапов

**Подробнее:** [`docs/BACKUP.md`](docs/BACKUP.md)

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

### Ключевые ADR (Architecture Decision Records)

- **[ADR 011: Drafts Autosave in MongoDB](architecture/decisions/011-drafts-autosave-mongodb.md)** — Автосохранение черновиков в MongoDB с eventual sync в PostgreSQL
- **[ADR 013: Graph Service Isolation](architecture/decisions/013-graph-service-isolation.md)** — Выделение Graph Service в отдельный сервис с gRPC и прямым доступом к БД
- **[ADR 014: Event-Driven Cache Invalidation](architecture/decisions/014-event-driven-cache-invalidation.md)** — Использование Redis Pub/Sub для инвалидации кэша
- **[ADR 015: Galactic Lexicon and Achievements](architecture/decisions/015-galactic-lexicon-and-achievements.md)** — Единый галактический лексикон, i18n, SSE для уведомлений о достижениях
- **[ADR 016: Keyword Similarity Strategies](architecture/decisions/016-keyword-similarity-strategies.md)** — Паттерн стратегий для метрик схожести ключевых слов с поддержкой весов
- **[ADR 017: Color Palette Redesign](architecture/decisions/017-color-palette-redesign.md)** — Чёрно-фиолетово-красная палитра для космической темы

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

## 📈 Graph Service

### Overview

Graph Service is an independent microservice responsible for computing 2D/3D graph layouts for the Knowledge Graph frontend. It provides high-performance graph visualization with caching, incremental updates, and event-driven invalidation [см. ADR 014](architecture/decisions/014-event-driven-cache-invalidation.md).

### Architecture

The Graph Service consists of:

- **API Layer**: gRPC server (port 9090) and HTTP fallback (port 9091) [см. ADR 013](architecture/decisions/013-graph-service-isolation.md)
- **Layout Engine**: 2D circular and 3D spiral layout algorithms with delta computation
- **Cache Layer**: Redis-backed caching with configurable TTL
- **Data Layer**: Direct PostgreSQL read access (notes, links, embeddings)
- **Event Subscriber**: Redis Pub/Sub for cache invalidation with acknowledgment tracking

### API Contracts

#### gRPC API (Primary) [см. ADR 013](architecture/decisions/013-graph-service-isolation.md)

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

Graph Service reads directly from PostgreSQL as the single source of truth, eliminating the need for an Outbox pattern while maintaining consistency through event-driven cache invalidation [см. ADR 014](architecture/decisions/014-event-driven-cache-invalidation.md).

---
