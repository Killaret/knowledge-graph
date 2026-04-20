# ADR 011: Drafts Autosave in MongoDB

## Status
Accepted

## Context
Knowledge Graph users expect autosave functionality:
- Frequent saves while typing (every 30 seconds)
- Drafts persist across sessions
- Minimal latency (save should not block typing)
- Eventually sync to PostgreSQL on explicit "Publish"

Traditional database writes are too slow and create too much transaction load for this pattern.

## Problem
How do we implement autosave drafts that:
1. Handles frequent writes (10+ saves per editing session)
2. Provides sub-100ms save latency
3. Persists drafts across browser sessions
4. Syncs to main storage only on explicit user action
5. Doesn't pollute the main notes table with transient data

## Decision
**MongoDB Drafts Collection with Eventual Sync**

Drafts stored in MongoDB for speed, synced to PostgreSQL only on publish.

### Architecture

```
User Types ──▶ Frontend ──▶ Go API ──▶ MongoDB (drafts)
                                          │
                                          │ (explicit publish)
                                          ▼
                                    PostgreSQL (notes)
                                          │
                                          ▼
                                    Redis (sync queue)
```

### Implementation

#### 1. Draft Schema (MongoDB)
```javascript
// Collection: drafts
{
  "_id": ObjectId("..."),
  "note_id": UUID("..."),        // Nullable for new notes
  "user_id": UUID("..."),
  "tenant_id": UUID("..."),
  
  // Draft content
  "title": "Draft: Meeting Notes",
  "content": "# Meeting Notes\n\n- Action item 1\n- Action item 2",
  "version": 5,                  // Incremental version number
  
  // Timestamps
  "created_at": ISODate("2024-01-15T09:00:00Z"),
  "last_saved_at": ISODate("2024-01-15T09:45:00Z"),
  
  // Session tracking
  "session_id": "uuid-session-123",
  "client_timestamp": ISODate("2024-01-15T09:45:00Z"),
  
  // Metadata
  "word_count": 150,
  "char_count": 890,
  "is_new_note": false
}
```

#### 2. Go Models
```go
package drafts

import (
    "time"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type Draft struct {
    ID             primitive.ObjectID `bson:"_id,omitempty"`
    NoteID         *uuid.UUID         `bson:"note_id,omitempty"`  // nil = new note
    UserID         uuid.UUID          `bson:"user_id"`
    TenantID       uuid.UUID          `bson:"tenant_id"`
    Title          string             `bson:"title"`
    Content        string             `bson:"content"`
    Version        int                `bson:"version"`
    CreatedAt      time.Time          `bson:"created_at"`
    LastSavedAt    time.Time          `bson:"last_saved_at"`
    SessionID      string             `bson:"session_id"`
    ClientTimestamp time.Time         `bson:"client_timestamp"`
    IsNewNote      bool               `bson:"is_new_note"`
}

// Conflict detection
type DraftConflict struct {
    HasConflict     bool
    ServerVersion   int
    ClientVersion   int
    ServerContent   string
    ClientContent   string
}
```

#### 3. Repository Implementation
```go
type MongoDraftRepository struct {
    collection *mongo.Collection
}

func (r *MongoDraftRepository) Save(ctx context.Context, draft Draft) error {
    draft.LastSavedAt = time.Now().UTC()
    
    filter := bson.M{
        "note_id": draft.NoteID,
        "user_id": draft.UserID,
    }
    
    // Upsert: Update if exists, insert if new
    opts := options.Update().SetUpsert(true)
    update := bson.M{
        "$set": draft,
        "$inc": bson.M{"version": 1},
    }
    
    _, err := r.collection.UpdateOne(ctx, filter, update, opts)
    return err
}

func (r *MongoDraftRepository) GetByNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*Draft, error) {
    filter := bson.M{
        "note_id": noteID,
        "user_id": userID,
    }
    
    var draft Draft
    err := r.collection.FindOne(ctx, filter).Decode(&draft)
    if err == mongo.ErrNoDocuments {
        return nil, nil
    }
    return &draft, err
}

func (r *MongoDraftRepository) GetByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]Draft, error) {
    filter := bson.M{
        "tenant_id": tenantID,
        "user_id":   userID,
    }
    
    opts := options.Find().SetSort(bson.D{{"last_saved_at", -1}})
    cursor, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    
    var drafts []Draft
    err = cursor.All(ctx, &drafts)
    return drafts, err
}

func (r *MongoDraftRepository) Delete(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error {
    filter := bson.M{
        "note_id": noteID,
        "user_id": userID,
    }
    _, err := r.collection.DeleteOne(ctx, filter)
    return err
}
```

#### 4. Handler Layer
```go
type SaveDraftHandler struct {
    draftRepo    DraftRepository
    conflictResolver ConflictResolver
}

type SaveDraftCommand struct {
    NoteID         *uuid.UUID  // nil for new notes
    UserID         uuid.UUID
    TenantID       uuid.UUID
    Title          string
    Content        string
    ClientVersion  int
    SessionID      string
}

func (h *SaveDraftHandler) Execute(ctx context.Context, cmd SaveDraftCommand) (*Draft, *DraftConflict, error) {
    // Check for existing draft
    var existing *Draft
    if cmd.NoteID != nil {
        existing, _ = h.draftRepo.GetByNote(ctx, *cmd.NoteID, cmd.UserID)
    }
    
    // Conflict detection (optimistic locking)
    if existing != nil && cmd.ClientVersion < existing.Version {
        return nil, &DraftConflict{
            HasConflict:   true,
            ServerVersion: existing.Version,
            ClientVersion: cmd.ClientVersion,
        }, ErrDraftConflict
    }
    
    draft := Draft{
        NoteID:         cmd.NoteID,
        UserID:         cmd.UserID,
        TenantID:       cmd.TenantID,
        Title:          cmd.Title,
        Content:        cmd.Content,
        Version:        existing.Version + 1,
        SessionID:      cmd.SessionID,
        ClientTimestamp: time.Now(),
        IsNewNote:      cmd.NoteID == nil,
    }
    
    if existing != nil {
        draft.CreatedAt = existing.CreatedAt
    } else {
        draft.CreatedAt = time.Now()
    }
    
    if err := h.draftRepo.Save(ctx, draft); err != nil {
        return nil, nil, err
    }
    
    return &draft, nil, nil
}
```

#### 5. Publish Sync (Draft → PostgreSQL)
```go
type PublishNoteHandler struct {
    draftRepo    DraftRepository
    noteRepo     domain.NoteRepository
    queue        PublishQueue
}

type PublishCommand struct {
    DraftID  primitive.ObjectID
    UserID   uuid.UUID
    TenantID uuid.UUID
}

func (h *PublishNoteHandler) Execute(ctx context.Context, cmd PublishCommand) (*domain.Note, error) {
    // 1. Load draft from MongoDB
    draft, err := h.draftRepo.GetByID(ctx, cmd.DraftID)
    if err != nil {
        return nil, err
    }
    
    // 2. Verify ownership
    if draft.UserID != cmd.UserID || draft.TenantID != cmd.TenantID {
        return nil, ErrUnauthorized
    }
    
    // 3. Create or update note in PostgreSQL
    var note *domain.Note
    if draft.NoteID != nil {
        // Update existing note
        note, err = h.noteRepo.GetByID(ctx, *draft.NoteID)
        if err != nil {
            return nil, err
        }
        note.UpdateContent(draft.Title, draft.Content)
    } else {
        // Create new note
        note, err = domain.NewNote(cmd.TenantID, cmd.UserID, draft.Title, draft.Content)
        if err != nil {
            return nil, err
        }
    }
    
    // 4. Save to PostgreSQL (transactional)
    if err := h.noteRepo.Save(ctx, note); err != nil {
        return nil, err
    }
    
    // 5. Delete draft from MongoDB
    if draft.NoteID != nil {
        h.draftRepo.Delete(ctx, *draft.NoteID, cmd.UserID)
    } else {
        h.draftRepo.DeleteByID(ctx, draft.ID)
    }
    
    // 6. Queue side effects (embeddings, links)
    h.queue.Enqueue(EmbeddingJob{NoteID: note.ID()})
    
    return note, nil
}
```

#### 6. Cleanup Strategy
```javascript
// Remove drafts older than 30 days (abandoned edits)
db.drafts.createIndex(
  { "last_saved_at": 1 },
  { expireAfterSeconds: 2592000 }  // 30 days
);

// Also clean when note is published (explicit delete)
```

## Consequences

### Positive
- ✅ Low latency: MongoDB writes faster than PostgreSQL for this pattern
- ✅ Reduced DB load: Frequent saves don't hit transactional database
- ✅ Better UX: Drafts persist across sessions without explicit save
- ✅ Isolation: Draft corruption can't affect published notes
- ✅ Version history: Draft version tracking for conflict detection

### Negative
- ⚠️ Data inconsistency window: Draft exists in MongoDB but not Postgres
- ⚠️ Sync complexity: Publish operation must handle both databases
- ⚠️ Orphaned drafts: Abandoned drafts need TTL cleanup
- ⚠️ Conflict resolution: Multi-tab editing requires version tracking

### Trade-offs
| Decision | Rationale |
|----------|-----------|
| MongoDB over Postgres | Write speed and schema flexibility |
| 30-day draft TTL | Balance storage vs user recovery needs |
| Optimistic locking (versions) | Simpler than pessimistic locking for this use case |
| Explicit publish | User intent clear; avoids accidental overwrites |

### Sync Failure Handling
```go
func (h *PublishNoteHandler) ExecuteWithRetry(ctx context.Context, cmd PublishCommand) (*domain.Note, error) {
    // 1. Save to Postgres
    note, err := h.saveToPostgres(ctx, cmd)
    if err != nil {
        return nil, err
    }
    
    // 2. Best-effort draft deletion
    // If this fails, TTL will clean it up eventually
    if err := h.draftRepo.Delete(ctx, note.ID(), cmd.UserID); err != nil {
        log.Warn("draft cleanup failed", "note_id", note.ID(), "error", err)
        // Don't fail the publish
    }
    
    return note, nil
}
```

## Compliance
- Drafts are transient: TTL ensures no long-term retention
- Audit trail: Publish events logged to audit system
- Access control: Same JWT/RBAC checks apply to draft endpoints

## References
- MongoDB Write Concern documentation
- Eventual Consistency patterns (Cassandra/Datastax)
