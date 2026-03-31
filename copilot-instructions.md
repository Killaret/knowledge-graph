# Copilot Instructions for Knowledge Graph

## Overview
**Knowledge Graph** is a domain-driven design (DDD) backend system written in Go with PostgreSQL + Redis, implementing a note-taking system with graph-based relationships and semantic recommendations using embeddings (pgvector).

---

## Build, Test, and Lint

### Backend (Go)

#### Build
```bash
# Build binary
cd backend
go build -o server ./cmd/server

# Run server
./server

# Docker build
docker-compose up --build
```

#### Testing
```bash
# Unit tests only
cd backend
go test ./...

# Unit tests with verbose output
go test -v ./...

# Single test
go test -v -run TestCreateNote ./...

# Integration tests (requires running PostgreSQL)
go test -tags=integration ./...

# Coverage report
go test -cover ./... -coverprofile=coverage.out
```

**Test Files**: Colocated with source code (e.g., `entity_test.go`, `*_handler_test.go`). Tests use table-driven patterns and in-memory mocks.

#### Environment Setup
Create `.env` in project root (or use `.env` in backend/):
```
DATABASE_URL=postgresql://kb_user:kb_password@localhost:5432/knowledge_base?sslmode=disable
REDIS_URL=redis://localhost:6379
SERVER_PORT=8080
```

#### Run Services
```bash
# Start PostgreSQL 16 + Redis (using Docker Compose)
docker-compose up postgres redis

# Or use Docker Compose for full stack
docker-compose up --build
```

**Database Credentials**:
- User: `kb_user`
- Password: `kb_password`
- Database: `knowledge_base` (prod), `knowledge_base_test` (tests)

**No built-in linter configured** — recommended setup:
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run ./...
```

---

## High-Level Architecture

### System Design: Domain-Driven Design (DDD)

The backend is organized into **four onion layers**:

```
┌─────────────────────────────────────┐
│     Interfaces (HTTP Handlers)      │  API layer: DTOs, request/response
├─────────────────────────────────────┤
│    Application (Use Cases)          │  (Future) CQRS commands/queries
├─────────────────────────────────────┤
│     Domain (Business Logic)         │  Entities, Value Objects, Repositories
├─────────────────────────────────────┤
│  Infrastructure (Adapters)          │  Database, config, external services
└─────────────────────────────────────┘
```

### Directory Structure

```
backend/
├── cmd/server/main.go              # Entry point, dependency injection
├── internal/
│   ├── domain/                     # Pure business logic, NO dependencies
│   │   ├── note/                   # Note aggregate root
│   │   │   ├── entity.go           # Note entity (private fields)
│   │   │   ├── value_objects.go    # Title, Content, Metadata (immutable)
│   │   │   └── repository.go       # Repository interface (NO implementation)
│   │   └── link/                   # Link aggregate root (separate from Note)
│   │       ├── entity.go
│   │       ├── value_objects.go
│   │       └── repository.go
│   ├── application/                # (Future) Use cases, services
│   ├── infrastructure/             # Adapters: DB, Redis, config
│   │   ├── db/
│   │   │   ├── db.go               # DB initialization
│   │   │   └── postgres/           # GORM implementations
│   │   │       ├── note_repo.go    # NoteRepository impl
│   │   │       ├── link_repo.go
│   │   │       └── models.go       # GORM models
│   ├── interfaces/                 # HTTP API layer
│   │   └── api/
│   │       ├── notehandler/        # Note HTTP handlers
│   │       ├── linkhandler/        # Link HTTP handlers
│   │       └── dto/                # Request/response DTOs
│   └── shared/                     # Cross-cutting concerns (future)
└── migrations/                     # Database migrations
```

### Key Architectural Decisions

1. **Aggregates**: `Note` and `Link` are separate aggregates (not nested). Link has its own repository for independent persistence.
   - **Why**: Notes can have thousands of links; loading all links per note query is inefficient.

2. **Immutable Value Objects**: `Title`, `Content`, `LinkType` are immutable, validated at construction.
   - **Why**: Prevents invalid states, ensures consistency.

3. **Repository Pattern**: Interface defined in Domain, implementation in Infrastructure.
   - **Dependencies flow inward**: Domain → Infrastructure (via interface), never outward.

4. **Dependency Injection**: Constructor-based, wired in `main.go`.
   - **Why**: Explicit, testable, no service locator.

5. **No CQRS yet**: Command/Query buses are in design but not yet implemented. All operations are through repositories directly.

---

## Key Conventions

### Naming

| Element | Pattern | Example |
|---------|---------|---------|
| **Packages** | lowercase, domain-focused | `note`, `link`, `postgres`, `notehandler` |
| **Types** | PascalCase, noun | `Note`, `Title`, `Repository`, `Handler` |
| **Functions** | camelCase, verb-first | `NewNote()`, `UpdateTitle()`, `FindByID()` |
| **Private fields** | lowercase | `id`, `title`, `sourceNoteID`, `createdAt` |
| **Factory functions** | `New<Type>` or `Reconstruct<Type>` | `NewNote()`, `ReconstructNote()` (DB reconstruction) |
| **Tests** | `*_test.go`, `Test<Function>` | `note_handler_test.go`, `TestCreateNote()` |
| **Mocks** | `mock_<interface>.go` | `mock_repo.go`, `mockNoteRepository` |

### Pattern: Entities with Private Fields (Encapsulation)

```go
// Domain layer: backend/internal/domain/note/entity.go
type Note struct {
    id        uuid.UUID
    title     Title              // Value Object
    content   Content            // Value Object
    createdAt time.Time
    updatedAt time.Time
}

// Factory: creates new Note
func NewNote(title Title, content Content) *Note {
    return &Note{
        id:        uuid.New(),
        title:     title,
        content:   content,
        createdAt: time.Now(),
        updatedAt: time.Now(),
    }
}

// Business logic: updates title with validation
func (n *Note) UpdateTitle(newTitle Title) error {
    if newTitle.String() == "" {
        return errors.New("cannot update with empty title")
    }
    n.title = newTitle
    n.updatedAt = time.Now()
    return nil
}

// Getters expose ID, Title, Content (read-only)
func (n *Note) ID() uuid.UUID { return n.id }
```

### Pattern: Value Objects (Immutable, Validated)

```go
// Domain layer: backend/internal/domain/note/value_objects.go
type Title struct {
    value string
}

// Constructor validates and returns error
func NewTitle(input string) (Title, error) {
    trimmed := strings.TrimSpace(input)
    if len(trimmed) == 0 {
        return Title{}, errors.New("title cannot be empty")
    }
    if len(trimmed) > 200 {
        return Title{}, errors.New("title exceeds 200 characters")
    }
    return Title{value: trimmed}, nil
}

// String() for serialization
func (t Title) String() string { return t.value }

// Comparison by value
func (t Title) Equals(other Title) bool { return t.value == other.value }
```

### Pattern: Repository (Domain Interface, Infrastructure Implementation)

```go
// Domain layer: backend/internal/domain/note/repository.go
type Repository interface {
    Save(ctx context.Context, note *Note) error
    FindByID(ctx context.Context, id uuid.UUID) (*Note, error)
    Delete(ctx context.Context, id uuid.UUID) error
}

// Infrastructure layer: backend/internal/infrastructure/db/postgres/note_repo.go
type NoteRepository struct {
    db *gorm.DB
}

func (r *NoteRepository) Save(ctx context.Context, n *note.Note) error {
    model := &NoteModel{
        ID:        n.ID(),
        Title:     n.Title().String(),
        // ... map other fields
    }
    return r.db.WithContext(ctx).Save(model).Error
}

func (r *NoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
    var model NoteModel
    if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, nil  // Not found
        }
        return nil, err
    }
    return note.ReconstructNote(model.ID, model.Title, model.Content, model.CreatedAt, model.UpdatedAt), nil
}
```

### Pattern: HTTP Handlers (Interfaces Layer)

```go
// backend/internal/interfaces/api/notehandler/note_handler.go
type Handler struct {
    repo domain.NoteRepository  // Depends on domain interface
}

func New(repo domain.NoteRepository) *Handler {
    return &Handler{repo: repo}
}

// POST /notes
func (h *Handler) Create(c *gin.Context) {
    var req CreateNoteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    title, err := domain.NewTitle(req.Title)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    content, err := domain.NewContent(req.Content)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    note := domain.NewNote(title, content)
    if err := h.repo.Save(c.Request.Context(), note); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save note"})
        return
    }

    c.JSON(http.StatusCreated, toDTO(note))
}
```

### Error Handling Conventions

1. **Value Object Construction**: Return `(Type, error)` for validation failures
   ```go
   title, err := NewTitle(input)  // Returns validation errors
   ```

2. **Entity Methods**: Return `error` for business logic violations
   ```go
   if err := note.UpdateTitle(newTitle); err != nil {
       // Handle business rule violation
   }
   ```

3. **Repository Methods**: Return `error` for persistence failures; `nil` for "not found"
   ```go
   note, err := repo.FindByID(ctx, id)
   if err != nil {
       return err  // DB error
   }
   if note == nil {
       return nil  // Not found (no error, just nil)
   }
   ```

4. **HTTP Handlers**: Map domain errors to HTTP status codes
   ```go
   if err != nil {
       switch err.(type) {
       case *ValidationError:
           c.JSON(http.StatusBadRequest, ...)
       case *NotFoundError:
           c.JSON(http.StatusNotFound, ...)
       default:
           c.JSON(http.StatusInternalServerError, ...)
       }
   }
   ```

### Testing Conventions

**Table-Driven Tests**:
```go
func TestNewTitle(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"empty", "", true},
        {"too long", string(make([]byte, 201)), true},
        {"valid", "Hello World", false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewTitle(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewTitle() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Mocking for Handlers**:
```go
func setupNoteRouter(t *testing.T) (*gin.Engine, *mockNoteRepository) {
    gin.SetMode(gin.TestMode)
    repo := newMockNoteRepository()
    handler := New(repo)
    r := gin.Default()
    r.POST("/notes", handler.Create)
    return r, repo
}

func TestCreateNote(t *testing.T) {
    r, mockRepo := setupNoteRouter(t)
    mockRepo.On("Save").Return(nil)

    req := httptest.NewRequest("POST", "/notes", strings.NewReader(`{"title":"Test","content":"Body"}`))
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    if w.Code != http.StatusCreated {
        t.Errorf("expected 201, got %d", w.Code)
    }
}
```

**Integration Tests** (with build tag):
```go
// +build integration

func TestNoteRepositorySave(t *testing.T) {
    // Requires live PostgreSQL
    db := setupTestDB(t)
    repo := postgres.NewNoteRepository(db)

    title, _ := domain.NewTitle("Integration Test")
    note := domain.NewNote(title, ...)
    err := repo.Save(context.Background(), note)

    if err != nil {
        t.Fatalf("Save failed: %v", err)
    }
}
```

---

## Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Backend | Go | 1.26.1 |
| API Framework | Gin Web Framework | v1.12.0 |
| ORM | GORM | v1.31.1 |
| Database | PostgreSQL + pgvector | 16 + v0.3.0 |
| Cache/Queue | Redis | 7 (Alpine) |
| UUID Generation | google/uuid | v1.6.0 |
| Env Config | joho/godotenv | v1.5.1 |
| Containerization | Docker & Docker Compose | Latest |

---

## Quick Reference: File Organization

When adding new features, follow this pattern:

```
feature_name/
├── domain/
│   └── feature/
│       ├── entity.go           # Core business entity
│       ├── value_objects.go    # Immutable value objects
│       ├── repository.go       # Interface (NO implementation)
│       ├── service.go          # (Future) Use cases
│       └── feature_test.go
├── infrastructure/
│   └── db/postgres/
│       ├── feature_repo.go     # Repository implementation
│       ├── model.go            # GORM model
│       └── feature_repo_test.go
└── interfaces/
    └── api/featurehandler/
        ├── handler.go          # HTTP handlers
        ├── dto.go              # Request/response DTOs
        └── handler_test.go
```

---

## References

- **Architecture**: See `/docs/architecture/README.md` for C4 models, UML diagrams, and ADRs
- **Glossary**: `/docs/architecture/glossary.md`
- **ADRs**: `/docs/architecture/adr.md` (Architectural Decision Records)
- **ATAM Analysis**: `/docs/architecture/atam.md`
