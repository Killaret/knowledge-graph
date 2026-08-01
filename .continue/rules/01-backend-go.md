---
name: Backend Go Rules
alwaysApply: false
globs: ["backend/**/*.go", "go.mod"]
description: Go backend patterns - Clean Architecture, DDD, GORM, Redis, JWT, dependency injection
---

# Backend Go Rules

## Manual Found → Automated Covered

If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue. Choose the test level by severity and scope:

- **unit** — pure logic, validators, or utilities (e.g. `errorMessage.ts`, email validation).
- **integration** — handlers, repositories, or routes (e.g. `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
- **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g. public graph, achievements, SSE fallback).

The test should fail before the fix (where safe) and pass after the fix. If the defect depends on manual data or config setup, fix the seed or config script — not only the instructions.

## Clean Architecture Layers

```
backend/internal/
├── domain/          # Pure business logic — NO external deps
│   ├── note/        # Note entity, value objects, repository interface
│   ├── link/        # Link entity, value objects
│   └── graph/       # Graph traversal, BFS, recommendations
├── application/     # Use cases — orchestrate domain
├── infrastructure/  # Implementations (GORM repos, Redis, NLP client, asynq)
│   ├── db/postgres/ # PostgreSQL repositories
│   ├── mongo/       # MongoDB draft storage
│   ├── queue/       # Asynq task queue
│   └── nlp/         # NLP service HTTP client
└── interfaces/      # HTTP layer (Gin handlers, middleware)
    ├── http/        # Handlers, routes
    └── middleware/  # Auth, CORS, rate limiting
```

## Domain Entity Pattern

```go
// Domain entity — private fields, factory function, getters
type Note struct {
    id        uuid.UUID
    title     Title       // Value Object
    content   Content     // Value Object
    type_     string
    metadata  Metadata
    creatorID *uuid.UUID
    createdAt time.Time
    updatedAt time.Time
}

// Factory function — validates and creates
func NewNote(title Title, content Content, noteType string, metadata Metadata) *Note {
    now := time.Now()
    if noteType == "" {
        noteType = "star"
    }
    return &Note{
        id: uuid.New(), title: title, content: content,
        type_: noteType, metadata: metadata,
        createdAt: now, updatedAt: now,
    }
}
```

## Repository Interface (Domain Layer)

```go
// Defined in domain/ — implementation in infrastructure/
type Repository interface {
    Save(ctx context.Context, note *Note) error
    FindByID(ctx context.Context, id uuid.UUID) (*Note, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]*Note, int64, error)
    Search(ctx context.Context, query string, limit, offset int) ([]*Note, int64, error)
}
```

## Infrastructure Implementation (Dependency Injection)

```go
type NoteRepository struct {
    db    *gorm.DB
    redis *redis.Client
}

func NewNoteRepository(db *gorm.DB, redis *redis.Client) *NoteRepository {
    return &NoteRepository{db: db, redis: redis}
}

func (r *NoteRepository) Save(ctx context.Context, n *note.Note) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        model, err := toGormNote(n)
        if err != nil {
            return err
        }
        return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&model).Error
    })
}
```

## Redis Caching Pattern (go-redis/v9)

```go
import "github.com/redis/go-redis/v9"

func (r *NoteRepository) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("note:%s", id.String())
    cached, err := r.redis.Get(ctx, cacheKey).Bytes()
    if err == nil {
        return unmarshalNote(cached)
    }
    // Fallback to DB
    var model NoteModel
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        return nil, note.ErrNoteNotFound
    }
    return toDomainNote(&model), nil
}
```

## JWT Handling

```go
// Middleware extracts user ID from JWT claims
func (m *AuthMiddleware) Handle() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractBearerToken(c.GetHeader("Authorization"))
        claims, err := m.jwtService.ValidateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

## Error Handling Pattern

```go
// Domain errors defined in domain layer
var (
    ErrNoteNotFound = errors.New("note not found")
    ErrInvalidTitle = errors.New("title must be between 1 and 500 characters")
)

// Handler maps domain errors to HTTP status codes
func (h *NoteHandler) GetByID(c *gin.Context) {
    note, err := h.useCase.GetByID(ctx, id)
    if errors.Is(err, note.ErrNoteNotFound) {
        c.JSON(404, gin.H{"error": "Note not found"})
        return
    }
    if err != nil {
        c.JSON(500, gin.H{"error": "Internal server error"})
        return
    }
    c.JSON(200, toResponse(note))
}
```

## Anti-Patterns

```go
// ❌ Bad — global variable
var db *gorm.DB

// ✅ Good — dependency injection via constructor
type Service struct { repo note.Repository }
func NewService(repo note.Repository) *Service { return &Service{repo: repo} }
```

```go
// ❌ Bad — returning infrastructure model from repository
func (r *NoteRepo) FindByID(id uuid.UUID) (*NoteModel, error) { ... }

// ✅ Good — returning domain entity
func (r *NoteRepo) FindByID(ctx context.Context, id uuid.UUID) (*note.Note, error) { ... }
```

```go
// ❌ Bad — missing context parameter
func (r *NoteRepo) Save(note *Note) error { ... }

// ✅ Good — context flows through all layers
func (r *NoteRepo) Save(ctx context.Context, note *Note) error { ... }
```

```go
// ❌ Bad — business logic in handler
func (h *Handler) Create(c *gin.Context) {
    // DON'T do validation/logic here
    if len(req.Title) > 500 { ... }
}

// ✅ Good — business logic in domain via Value Objects
func NewTitle(raw string) (Title, error) {
    if len(raw) == 0 || len(raw) > 500 {
        return Title{}, ErrInvalidTitle
    }
    return Title{value: raw}, nil
}
```

## Test Patterns

```go
func TestNoteRepository_Save(t *testing.T) {
    tests := []struct {
        name    string
        note    *note.Note
        wantErr bool
    }{
        {
            name:    "valid note saves successfully",
            note:    note.NewNote(validTitle, validContent, "star", nil),
            wantErr: false,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := repo.Save(context.Background(), tt.note)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

## pgvector Usage

```go
// Embedding storage with pgvector
type NoteEmbeddingModel struct {
    NoteID    uuid.UUID `gorm:"primaryKey"`
    Embedding []float32 `gorm:"type:vector(384)"` // all-MiniLM-L6-v2 dimension
}

// Similarity search
func (r *EmbeddingRepo) FindSimilar(ctx context.Context, embedding []float32, limit int) ([]*note.Note, error) {
    r.db.WithContext(ctx).Raw(
        "SELECT note_id FROM note_embeddings ORDER BY embedding <=> ? LIMIT ?",
        pgvector.NewVector(embedding), limit,
    )
}
```
