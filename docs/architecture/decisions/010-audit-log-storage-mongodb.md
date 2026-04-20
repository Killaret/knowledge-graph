# ADR 010: Audit Log Storage in MongoDB

## Status
Accepted

## Context
Audit logging generates high-volume, write-heavy traffic:
- Every API request generates 1-2 audit events
- 1000 active users × 100 requests/day = 100,000 events/day
- Events must be retained for compliance (90 days minimum)
- Query patterns: time-range scans, tenant-filtered, user-filtered

PostgreSQL is optimized for transactional workloads, not append-only log storage.

## Problem
How do we store audit logs that:
1. Handles high write throughput without impacting transactional DB
2. Supports flexible schema (different events have different metadata)
3. Allows efficient time-range queries for investigations
4. Auto-expires old data (compliance retention limits)
5. Remains cost-effective at scale

## Decision
**MongoDB Collection with TTL Index**

We choose MongoDB over PostgreSQL for audit logs because:
- Schema flexibility: Events have varying metadata shapes
- Write performance: Optimized for high-volume inserts
- TTL indexes: Automatic data expiration
- Separation: Audit load isolated from transactional workload

### Implementation

#### 1. MongoDB Schema Design
```javascript
// Collection: audit_logs
{
  "_id": ObjectId("..."),           // Auto-generated
  "timestamp": ISODate("2024-01-15T10:30:00Z"),
  "tenant_id": UUID("tenant-uuid"),
  "user_id": UUID("user-uuid"),
  
  // Action classification
  "action": "note:create",            // {resource}:{operation}
  "resource": {
    "type": "note",
    "id": UUID("note-uuid")
  },
  
  // HTTP context
  "http": {
    "method": "POST",
    "path": "/api/v1/notes",
    "status_code": 201,
    "duration_ms": 145,
    "ip_address": "192.168.1.100",
    "user_agent": "Mozilla/5.0..."
  },
  
  // Flexible metadata (varies by action)
  "metadata": {
    "request_size": 2048,
    "response_size": 512,
    "query_params": {"include": "links"},
    // Action-specific fields
    "note_title_preview": "Meeting Notes...",
    "link_count": 3
  },
  
  // Compliance tagging
  "sensitivity": "normal",            // normal, high, gdpr
  "retention_class": "standard"       // standard, extended, permanent
}
```

#### 2. Go Model
```go
package audit

import (
    "time"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditLog struct {
    ID            primitive.ObjectID `bson:"_id,omitempty"`
    Timestamp     time.Time          `bson:"timestamp"`
    TenantID      uuid.UUID          `bson:"tenant_id"`
    UserID        uuid.UUID          `bson:"user_id"`
    Action        string             `bson:"action"`
    Resource      ResourceRef        `bson:"resource"`
    HTTP          HTTPContext        `bson:"http"`
    Metadata      map[string]any     `bson:"metadata"`
    Sensitivity   string             `bson:"sensitivity"`
}

type ResourceRef struct {
    Type string    `bson:"type"`
    ID   uuid.UUID `bson:"id"`
}

type HTTPContext struct {
    Method       string `bson:"method"`
    Path         string `bson:"path"`
    StatusCode   int    `bson:"status_code"`
    DurationMs   int64  `bson:"duration_ms"`
    IPAddress    string `bson:"ip_address"`
    UserAgent    string `bson:"user_agent"`
}
```

#### 3. TTL Index Configuration
```javascript
// Auto-delete after 90 days
db.audit_logs.createIndex(
  { "timestamp": 1 },
  { expireAfterSeconds: 7776000 }  // 90 days
);

// Compound indexes for queries
db.audit_logs.createIndex({ "tenant_id": 1, "timestamp": -1 });
db.audit_logs.createIndex({ "user_id": 1, "timestamp": -1 });
db.audit_logs.createIndex({ "action": 1, "timestamp": -1 });
db.audit_logs.createIndex({ "http.status_code": 1 });

// For investigation queries
db.audit_logs.createIndex({ 
  "tenant_id": 1, 
  "action": 1, 
  "timestamp": -1 
});
```

#### 4. Repository Implementation
```go
type MongoAuditRepository struct {
    collection *mongo.Collection
}

func (r *MongoAuditRepository) Insert(ctx context.Context, log AuditLog) error {
    _, err := r.collection.InsertOne(ctx, log)
    return err
}

func (r *MongoAuditRepository) InsertBatch(ctx context.Context, logs []AuditLog) error {
    docs := make([]interface{}, len(logs))
    for i, log := range logs {
        docs[i] = log
    }
    
    _, err := r.collection.InsertMany(ctx, docs)
    return err
}

// Query by tenant and time range (common compliance query)
func (r *MongoAuditRepository) FindByTenant(
    ctx context.Context, 
    tenantID uuid.UUID, 
    from, to time.Time,
    limit int,
) ([]AuditLog, error) {
    filter := bson.M{
        "tenant_id": tenantID,
        "timestamp": bson.M{
            "$gte": from,
            "$lte": to,
        },
    }
    
    opts := options.Find().
        SetLimit(int64(limit)).
        SetSort(bson.D{{"timestamp", -1}})
    
    cursor, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var logs []AuditLog
    err = cursor.All(ctx, &logs)
    return logs, err
}

// Investigation query: Find all actions by user
func (r *MongoAuditRepository) FindByUser(
    ctx context.Context,
    tenantID, userID uuid.UUID,
    from, to time.Time,
) ([]AuditLog, error) {
    filter := bson.M{
        "tenant_id": tenantID,
        "user_id":   userID,
        "timestamp": bson.M{
            "$gte": from,
            "$lte": to,
        },
    }
    
    cursor, err := r.collection.Find(ctx, filter)
    // ...
}
```

#### 5. Storage Calculation
```
Assumptions:
- 1000 active tenants
- 1000 events per tenant per day
- Average event size: 500 bytes
- 90-day retention

Daily volume: 1M events × 500 bytes = 500 MB/day
90-day storage: 500 MB × 90 = 45 GB
Growth: ~15 TB/year with current retention

With TTL (auto-delete): Storage stabilizes at ~45 GB
```

## Consequences

### Positive
- ✅ Schema flexibility: Different events have different metadata without schema migration
- ✅ Write throughput: MongoDB handles 10K+ inserts/second
- ✅ Query performance: Indexes support investigation patterns
- ✅ Auto-cleanup: TTL removes expired data without batch jobs
- ✅ Isolation: Audit load doesn't affect transactional PostgreSQL

### Negative
- ⚠️ Secondary database: Additional operational complexity (backup, monitoring)
- ⚠️ Eventual consistency: No ACID transactions across Postgres-Mongo
- ⚠️ Query limitations: No JOINs with transactional data
- ⚠️ Backup complexity: Separate backup strategy needed

### Why Not PostgreSQL?
| Concern | PostgreSQL | MongoDB |
|---------|------------|---------|
| Schema changes | Migration required | Flexible documents |
| Write throughput | Moderate | High |
| Auto-expiration | Cron job + DELETE | Native TTL index |
| Storage efficiency | Row overhead | Binary JSON efficient |
| Operational load | Shared with app data | Isolated |

### Resilience
See ADR 009 for circuit breaker and fallback implementation for MongoDB.

## Compliance
- **Retention**: TTL ensures 90-day compliance limit
- **Tamper resistance**: Separate MongoDB user with insert-only permissions
- **Queryable**: Indexes support SOC 2 auditor investigations

## References
- MongoDB TTL Indexes: https://docs.mongodb.com/manual/core/index-ttl/
- Schema Design Patterns: https://www.mongodb.com/blog/post/building-with-patterns-a-summary
