# Cursor Rule: knowledge-graph-backend-go

Go 1.25 backend. Module: `knowledge-graph` (see `backend/go.mod`).
Entry points: `backend/cmd/server/main.go` (HTTP), `backend/cmd/worker/main.go` (asynq).

---

## Clean Architecture Layer Rules

```
backend/internal/
  domain/          ← Entities, Value Objects, Repository interfaces, domain errors
  application/     ← Use-cases, services, task queue interface, cache
  infrastructure/  ← GORM repos, Redis, MongoDB, external HTTP clients
  interfaces/api/  ← Gin handlers, DTOs, middleware
```

| Layer | May import | Must NOT import |
|-------|-----------|-----------------|
| `domain` | stdlib only | `gorm`, `redis`, `gin`, `asynq` |
| `application` | `domain`, stdlib | `gorm`, `gin`, HTTP libs |
| `infrastructure` | `domain`, `application`, `gorm`, `redis` | `gin` |
| `interfaces/api` | `application`, `domain`, `gin` | `gorm` directly |

---

## DDD Patterns

### Entity — unexported fields, constructor validates invariants
```go
// backend/internal/domain/note/entity.go
type Note struct {
    id        uuid.UUID
    title     Title       // Value Object
    content   Content     // Value Object
    creatorID *uuid.UUID
    createdAt time.Time
    updatedAt time.Time
}

func NewNoteWithCreator(title Title, content Content, noteType string,
    metadata Metadata, creatorID uuid.UUID) *Note { ... }

// ReconstructNote is called ONLY by the repository layer
func ReconstructNoteWithCreator(id uuid.UUID, ...) *Note { ... }
```

### Value Object — immutable, validated on construction
```go
// backend/internal/domain/note/value_objects.go
type Title struct{ value string }

func NewTitle(value string) (Title, error) {
    trimmed := strings.TrimSpace(value)
    if len(trimmed) == 0 { return Title{}, errors.New("title cannot be empty") }
    if len(trimmed) > 200 { return Title{}, errors.New("title too long (max 200 characters)") }
    return Title{value: trimmed}, nil
}
func (t Title) String() string { return t.value }
```

### Repository Interface — domain layer, returns domain entities
```go
// backend/internal/domain/note/repository.go
var ErrNoteNotFound = errors.New("note not found")

type Repository interface {
    Save(ctx context.Context, note *Note) error
    FindByID(ctx context.Context, id uuid.UUID) (*Note, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]*Note, int64, error)
    Search(ctx context.Context, query string, limit, offset int) ([]*Note, int64, error)
    FindAll(ctx context.Context) ([]*Note, error)
    FindAllPaginated(ctx context.Context, limit, offset int) ([]*Note, int64, error)
}
```

---

## Repository Implementation (GORM)

```go
// backend/internal/infrastructure/db/postgres/note_repo.go
type NoteRepository struct {
    db    *gorm.DB
    redis *redis.Client  // injected, may be nil
}

func NewNoteRepository(db *gorm.DB, redis *redis.Client) *NoteRepository {
    return &NoteRepository{db: db, redis: redis}
}

func (r *NoteRepository) Save(ctx context.Context, n *note.Note) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var existing NoteModel
        err := tx.Where("id = ?", n.ID()).First(&existing).Error
        if errors.Is(err, gorm.ErrRecordNotFound) {
            model, err := toGormNote(n)
            if err != nil { return err }
            return tx.Create(&model).Error
        }
        if err != nil { return err }
        model, err := toGormNote(n)
        if err != nil { return err }
        return tx.Model(&existing).Updates(model).Error
    })
}

// toDomainNote converts GORM model → domain entity (called ONLY inside repo)
func toDomainNote(m *NoteModel) (*note.Note, error) { ... }
```

---

## go-redis/v9 API (CRITICAL — v8 names differ)

```go
// ✅ CORRECT — go-redis/v9 pool options
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr:            cfg.RedisURL,
    ConnMaxIdleTime: time.Minute,      // v9 name
    ConnMaxLifetime: 5 * time.Minute,  // v9 name
    PoolSize:        10,
})

// Check connection
ctx := context.Background()
if err := rdb.Ping(ctx).Err(); err != nil { ... }

// Set with TTL
rdb.Set(ctx, "graph:"+userID, data, 5*time.Minute)

// Get — check for redis.Nil explicitly
val, err := rdb.Get(ctx, key).Result()
if err == redis.Nil { /* cache miss */ }

// ❌ WRONG — these are go-redis/v8 names, will not compile in v9
// MaxConnAge:  time.Minute   ← v8 only
// IdleTimeout: time.Minute   ← v8 only
```

See `backend/internal/auth/redis_store.go` and
`backend/internal/application/cache/graph_cache.go` for production patterns.

---

## asynq Task Queue

```go
// backend/internal/application/common/task_queue.go
type TaskQueue interface {
    Enqueue(ctx context.Context, task *asynq.Task) error
    EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error
    EnqueueComputeEmbedding(ctx context.Context, noteID string) error
}

// Enqueue a task from application layer (never call asynq directly in domain)
if err := q.EnqueueComputeEmbedding(ctx, note.ID().String()); err != nil {
    return fmt.Errorf("enqueue embedding: %w", err)
}

// Worker entrypoint: backend/cmd/worker/main.go
// Handlers live in internal/interfaces/api/ or internal/application/
```

---

## pgvector Usage

```go
// backend/internal/infrastructure/db/postgres/embedding_repo.go
import "github.com/pgvector/pgvector-go"

type NoteEmbeddingModel struct {
    NoteID    uuid.UUID        `gorm:"type:uuid;primaryKey"`
    Embedding pgvector.Vector  `gorm:"type:vector(384)"`
    UpdatedAt time.Time
}

// Upsert embedding
r.db.WithContext(ctx).Clauses(clause.OnConflict{
    Columns:   []clause.Column{{Name: "note_id"}},
    DoUpdates: clause.AssignmentColumns([]string{"embedding", "updated_at"}),
}).Create(&NoteEmbeddingModel{NoteID: noteID, Embedding: vec})

// Cosine similarity search (operator <=> = cosine distance)
r.db.WithContext(ctx).Raw(`
    SELECT e2.note_id,
           (1 - (e1.embedding <=> e2.embedding)) / 2.0 AS similarity
    FROM note_embeddings e1
    JOIN note_embeddings e2 ON e1.note_id != e2.note_id
    WHERE e1.note_id = ?
    ORDER BY similarity DESC
    LIMIT ?
`, noteID, limit).Scan(&results)
```

---

## Gin Handler Structure

```go
// backend/internal/interfaces/api/notehandler/handler.go
type Handler struct {
    service note.Service  // application interface, NOT concrete type
}

func NewHandler(service note.Service) *Handler {
    return &Handler{service: service}
}

// @Summary      Create note
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        body  body      CreateNoteRequest  true  "Note data"
// @Success      201   {object}  NoteResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /api/v1/notes [post]
func (h *Handler) Create(c *gin.Context) {
    var req CreateNoteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }
    // validate with go-playground/validator via ShouldBindJSON tags
    n, err := h.service.Create(c.Request.Context(), req.Title, req.Content)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }
    c.JSON(http.StatusCreated, toNoteResponse(n))
}
```

Route registration: `backend/cmd/server/router.go` (`v1 := r.Group("/api/v1")`).

---

## JWT Middleware

```go
// backend/internal/auth/jwt.go
manager := auth.NewJWTManager(secret, 15*time.Minute, 7*24*time.Hour)
pair, err := manager.GenerateTokenPair(userID, login, role)
claims, err := manager.ValidateToken(tokenString, "access")

// Middleware wired in: backend/cmd/server/router.go
r.Use(middleware.JWTAuth(jwtConfig))
```

Redis-backed blacklisting: `backend/internal/auth/redis_store.go` —
`BlacklistToken`, `IsTokenBlacklisted`, `StoreRefreshToken`.

---

## Error Handling

```go
// ✅ Wrap and propagate
if err := r.db.WithContext(ctx).First(&model).Error; err != nil {
    return nil, fmt.Errorf("NoteRepo.FindByID: %w", err)
}

// ✅ Sentinel errors for business logic
if errors.Is(err, note.ErrNoteNotFound) {
    c.JSON(http.StatusNotFound, ErrorResponse{Error: "note not found"})
    return
}

// ❌ Never panic in handler or service code
// panic("something went wrong")  — use gin's RecoveryMiddleware instead
```

---

## Anti-Patterns

```go
// ❌ Global database variable
var DB *gorm.DB  // NEVER

// ❌ GORM model leaking out of repository
func (h *Handler) Create(c *gin.Context) {
    var m postgres.NoteModel  // wrong — domain layer only
}

// ❌ Business logic in handler
func (h *Handler) Create(c *gin.Context) {
    if len(req.Title) > 200 { ... }  // wrong — belongs in Value Object / domain
}

// ❌ Importing gorm in domain layer
import "gorm.io/gorm"  // in domain/note/entity.go — NEVER

// ❌ go-redis v8 option names
redis.NewClient(&redis.Options{MaxConnAge: time.Minute})  // v8 only, breaks v9
```
