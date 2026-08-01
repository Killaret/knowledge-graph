# Cursor Rule: knowledge-graph-data

Data layer: PostgreSQL + pgvector, MongoDB, Redis. GORM models, migrations,
connection pooling, key naming, and backup conventions.

## Manual Found → Automated Covered

If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue. Choose the test level by severity and scope:

- **unit** — pure logic, validators, or utilities (e.g. `errorMessage.ts`, email validation).
- **integration** — handlers, repositories, or routes (e.g. `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
- **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g. public graph, achievements, SSE fallback).

The test should fail before the fix (where safe) and pass after the fix. If the defect depends on manual data or config setup, fix the seed or config script — not only the instructions.

---

## GORM Model Conventions

```go
// backend/internal/infrastructure/db/postgres/note_model.go
type NoteModel struct {
    ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
    Title        string         `gorm:"not null"`
    Content      string         `gorm:"type:text"`
    Type         string         `gorm:"type:varchar(50);default:'star'"`
    Metadata     datatypes.JSON `gorm:"type:jsonb"`
    SearchVector string         `gorm:"column:search_vector;type:tsvector;->"` // read-only trigger
    CreatorID    *uuid.UUID     `gorm:"type:uuid;index"`
    Creator      *UserModel     `gorm:"foreignKey:CreatorID"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    *time.Time     `gorm:"index"`  // soft delete
}

func (NoteModel) TableName() string { return "notes" }
```

Rules:
- `TableName()` must always be explicit — never rely on GORM plural inference.
- Use `*time.Time` for nullable `DeletedAt` (soft delete) — NOT `gorm.Model`.
- UUIDs use `gen_random_uuid()` — require `pgcrypto` extension in migrations.
- `datatypes.JSON` for JSONB columns (`gorm.io/datatypes`).
- `tsvector` columns are read-only (`->`) — updated by PostgreSQL triggers.

---

## PostgreSQL Migration Pattern

Raw SQL migrations in `backend/migrations/`, named `{version}_{description}.up.sql`:

```sql
-- backend/migrations/001_create_notes.up.sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "vector";  -- pgvector

CREATE TABLE notes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title        TEXT NOT NULL,
    content      TEXT,
    type         VARCHAR(50) NOT NULL DEFAULT 'star',
    metadata     JSONB,
    search_vector TSVECTOR,
    creator_id   UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX notes_creator_idx ON notes(creator_id);
CREATE INDEX notes_search_idx  ON notes USING GIN(search_vector);

-- Trigger to maintain search_vector from title + content
CREATE OR REPLACE FUNCTION notes_search_vector_update() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('russian', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('simple',  coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('russian', coalesce(NEW.content, '')), 'B');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notes_search_vector_trigger
  BEFORE INSERT OR UPDATE ON notes
  FOR EACH ROW EXECUTE FUNCTION notes_search_vector_update();
```

Migration runner: `backend/internal/infrastructure/db/postgres/migrations.go`
(`RunMigrations(db, migrationsDir)`). Called on `server` startup.

---

## pgvector Column and Index

```go
// backend/internal/infrastructure/db/postgres/note_embedding_model.go
type NoteEmbeddingModel struct {
    NoteID    uuid.UUID        `gorm:"type:uuid;primaryKey"`
    Embedding pgvector.Vector  `gorm:"type:vector(384)"`  // 384 dims for all-MiniLM-L6-v2
    UpdatedAt time.Time
}
func (NoteEmbeddingModel) TableName() string { return "note_embeddings" }
```

```sql
-- Migration: create IVFFlat index for ANN search
CREATE INDEX note_embeddings_ivfflat_idx
    ON note_embeddings
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- Usage: cosine distance operator <=>
SELECT note_id, (1 - (embedding <=> $1::vector)) / 2.0 AS similarity
FROM note_embeddings
ORDER BY embedding <=> $1::vector
LIMIT 10;
```

See full implementation: `backend/internal/infrastructure/db/postgres/embedding_repo.go`.

---

## MongoDB Usage

MongoDB (`mongo:7`, port 27017) stores:
- **Audit logs** — `audit_logs` collection (append-only, no GORM)
- **Drafts** — `drafts` collection (user's unsaved note drafts)

```go
// Pattern: explicit collection names
db.Collection("audit_logs").InsertOne(ctx, bson.M{
    "user_id":    userID.String(),
    "action":     "note.created",
    "resource_id": noteID.String(),
    "timestamp":  time.Now(),
})

// Drafts service: backend/internal/application/draft/service.go
db.Collection("drafts").FindOne(ctx, bson.M{
    "user_id": userID.String(),
    "note_id": noteID.String(),
})
```

Use `go.mongodb.org/mongo-driver v1.17.9` (direct import).
Never use MongoDB for relational data — that belongs in PostgreSQL.

---

## Redis Key Naming Conventions

All keys are namespaced by prefix:

```
notes:all                          ← full notes list (5m TTL)
graph:{userID}                     ← user graph cache (5m TTL)
auth:blacklist:{sha256(token)}     ← JWT blacklist (TTL = remaining token life)
auth:refresh:{sha256(token)}       ← refresh token store (TTL = 7 days)
auth:pkce:{state}                  ← OAuth PKCE (TTL = 10 minutes)
auth:password_reset:{sha256(tok)}  ← password reset (TTL = 15 minutes)
auth:perm:{uid}:{resource}:{action} ← permission cache (short TTL)
```

Rules:
- **No PII in key names** (no email, no plain token — use SHA-256 hash).
- All keys set with explicit TTL — never `Set(key, val, 0)` (no expiry).
- Use `Scan` iterator for bulk deletion, never `KEYS *` in production.

---

## Database Connection Pooling

```go
// backend/internal/infrastructure/db/db.go
sqlDB.SetMaxOpenConns(25)                 // max concurrent DB connections
sqlDB.SetMaxIdleConns(5)                  // keep alive for reuse
sqlDB.SetConnMaxLifetime(5 * time.Minute) // recycle old connections
sqlDB.SetConnMaxIdleTime(1 * time.Minute) // close idle connections faster
```

Redis pool:
```go
redis.NewClient(&redis.Options{
    Addr:            cfg.RedisURL,
    PoolSize:        10,
    ConnMaxIdleTime: time.Minute,      // v9 API — NOT IdleTimeout
    ConnMaxLifetime: 5 * time.Minute,  // v9 API — NOT MaxConnAge
})
```

---

## REDIS_FLUSH_ON_STARTUP Flag

```yaml
# docker-compose.yml
environment:
  REDIS_FLUSH_ON_STARTUP: "false"  # NEVER true in personal/production
```

When `true`, the backend flushes all Redis keys on startup. Only use this
during development to reset caches. Never enable in the personal stack where
real user sessions are stored.

---

## Transaction Patterns with GORM

```go
// Use Transaction() for operations that must be atomic
return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&noteModel).Error; err != nil {
        return err  // auto-rollback on non-nil return
    }
    if err := tx.Create(&embeddingModel).Error; err != nil {
        return err
    }
    return nil  // auto-commit
})

// For read-only — use WithContext directly (no transaction overhead)
r.db.WithContext(ctx).Where("creator_id = ?", userID).Find(&models)
```

---

## Backup Patterns

Personal stack runs `backup_scheduler` service (defined in
`docker-compose.personal.yml`). It dumps PostgreSQL and uploads to Yandex.Disk:

```yaml
backup_scheduler:
  image: postgres:16-alpine
  environment:
    BACKUP_YANDEX_TOKEN: ${BACKUP_YANDEX_TOKEN}
    BACKUP_YANDEX_FOLDER: ${BACKUP_YANDEX_FOLDER:-/KnowledgeGraphBackups}
    BACKUP_CLOUD_ENABLED: ${BACKUP_CLOUD_ENABLED:-true}
```

Go Yandex.Disk client: `backend/internal/infrastructure/cloud/yandex_disk.go`.
Backup trigger endpoint: `POST /api/v1/backup/cloud`.

---

## Anti-Patterns

```go
// ❌ No explicit TableName — GORM guesses plural "note_models"
type NoteModel struct { ... }  // missing TableName()

// ❌ Storing domain entity in Redis (unexported fields serialize as {})
json.Marshal(domainNote)  // returns "{}" — store NoteModel instead

// ❌ Key without TTL — leaks memory forever
r.redis.Set(ctx, "notes:all", data, 0)  // 0 = no expiry

// ❌ KEYS * in production (blocks Redis event loop)
r.redis.Do(ctx, "KEYS", "*")

// ❌ Using GORM AutoMigrate instead of raw SQL migrations
db.AutoMigrate(&NoteModel{})  // loses tsvector triggers, custom indexes, etc.

// ❌ Redis v8 option names with v9 client
redis.NewClient(&redis.Options{
    MaxConnAge:  5 * time.Minute,   // v8 — compile error in v9
    IdleTimeout: time.Minute,       // v8 — compile error in v9
})
```
