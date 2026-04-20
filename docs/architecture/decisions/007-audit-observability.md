# ADR 007: Audit & Observability

## Status
Accepted

## Context
Multi-tenant SaaS requires comprehensive audit trails for:
- Security investigations (who accessed what data)
- Compliance reporting (SOC 2, GDPR)
- Debugging production issues
- Understanding user behavior

Audit events must capture context (who, what, when, from where) without impacting user-facing performance.

## Problem
How do we implement audit logging that:
1. Captures complete request context (user, tenant, IP, user-agent)
2. Does not block HTTP responses
3. Handles high write volume (every API call generates events)
4. Supports long-term retention with efficient querying
5. Maintains tamper-evident records

## Decision
**Async Audit Pipeline with Context Extraction**

### Architecture

```
┌─────────────┐    ┌──────────────┐    ┌──────────────┐    ┌─────────────┐
│   HTTP      │───▶│  Middleware  │───▶│  Redis       │───▶│  Worker     │
│   Request   │    │  (Extract    │    │  Queue       │    │  (Persist   │
│             │    │   Context)  │    │  (audit:log) │    │   to Mongo) │
└─────────────┘    └──────────────┘    └──────────────┘    └─────────────┘
       │
       ▼
┌──────────────┐
│  Response    │  ◀── Does not wait for audit write
│  (Fast)      │
└──────────────┘
```

### Implementation

#### 1. Context Extraction Middleware
```go
type AuditContext struct {
    Timestamp    time.Time
    TenantID     uuid.UUID
    UserID       uuid.UUID
    Action       string          // "note:create", "note:read", etc.
    Resource     string          // "note:uuid-123"
    Method       string          // HTTP method
    Path         string          // API endpoint
    StatusCode   int             // HTTP response code
    IPAddress    string          // Client IP
    UserAgent    string          // Browser/client info
    DurationMs   int64           // Request duration
    Metadata     map[string]any  // Flexible additional data
}

func AuditMiddleware(queue AuditQueue) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Wrap response writer to capture status code
            rw := &responseRecorder{ResponseWriter: w, statusCode: 200}
            
            // Extract context before handling
            claims := JWTFromContext(r.Context())
            
            next.ServeHTTP(rw, r)
            
            // Build audit event
            event := AuditContext{
                Timestamp:  time.Now().UTC(),
                TenantID:   uuid.MustParse(claims.TenantID),
                UserID:     uuid.MustParse(claims.Sub),
                Action:     extractAction(r),
                Resource:   extractResource(r),
                Method:     r.Method,
                Path:       r.URL.Path,
                StatusCode: rw.statusCode,
                IPAddress:  extractIP(r),
                UserAgent:  r.UserAgent(),
                DurationMs: time.Since(start).Milliseconds(),
                Metadata: map[string]any{
                    "query_params": r.URL.Query(),
                    "request_size": r.ContentLength,
                },
            }
            
            // Async enqueue - never blocks response
            go func() {
                if err := queue.Enqueue(r.Context(), event); err != nil {
                    log.Error("audit enqueue failed", err)
                }
            }()
        })
    }
}
```

#### 2. Redis Queue Implementation
```go
type RedisAuditQueue struct {
    client *redis.Client
    key    string // "audit:events"
}

func (q *RedisAuditQueue) Enqueue(ctx context.Context, event AuditContext) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    // Push to Redis list (LPUSH for queue)
    return q.client.LPush(ctx, q.key, data).Err()
}

func (q *RedisAuditQueue) DequeueBatch(ctx context.Context, batchSize int) ([]AuditContext, error) {
    // RPOP multiple events for batch insert
    pipe := q.client.Pipeline()
    cmds := make([]*redis.StringCmd, batchSize)
    
    for i := 0; i < batchSize; i++ {
        cmds[i] = pipe.RPop(ctx, q.key)
    }
    
    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, err
    }
    
    var events []AuditContext
    for _, cmd := range cmds {
        data, err := cmd.Result()
        if err == redis.Nil {
            break
        }
        if err != nil {
            continue
        }
        
        var event AuditContext
        if err := json.Unmarshal([]byte(data), &event); err != nil {
            continue
        }
        events = append(events, event)
    }
    
    return events, nil
}
```

#### 3. Background Worker
```go
type AuditWorker struct {
    queue     AuditQueue
    mongo     *mongo.Collection  // audit_logs collection
    batchSize int
    interval  time.Duration
}

func (w *AuditWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            if err := w.processBatch(ctx); err != nil {
                log.Error("audit batch processing failed", err)
            }
        }
    }
}

func (w *AuditWorker) processBatch(ctx context.Context) error {
    events, err := w.queue.DequeueBatch(ctx, w.batchSize)
    if err != nil {
        return err
    }
    
    if len(events) == 0 {
        return nil
    }
    
    // Convert to MongoDB documents
    docs := make([]interface{}, len(events))
    for i, e := range events {
        docs[i] = bson.M{
            "timestamp":    e.Timestamp,
            "tenant_id":    e.TenantID,
            "user_id":      e.UserID,
            "action":       e.Action,
            "resource":     e.Resource,
            "method":       e.Method,
            "path":         e.Path,
            "status_code":  e.StatusCode,
            "ip_address":   e.IPAddress,
            "user_agent":   e.UserAgent,
            "duration_ms":  e.DurationMs,
            "metadata":     e.Metadata,
        }
    }
    
    _, err = w.mongo.InsertMany(ctx, docs)
    return err
}
```

#### 4. Fallback on Redis Failure
```go
func (q *RedisAuditQueue) EnqueueWithFallback(ctx context.Context, event AuditContext) error {
    // Try Redis first
    if err := q.Enqueue(ctx, event); err == nil {
        return nil
    }
    
    // Fallback: write to local file
    data, _ := json.Marshal(event)
    line := fmt.Sprintf("%s\n", string(data))
    
    f, err := os.OpenFile("/var/log/audit/audit.fallback.log", 
        os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    
    _, err = f.WriteString(line)
    return err
}
```

## Consequences

### Positive
- ✅ Non-blocking: HTTP response time unaffected by audit logging
- ✅ High throughput: Redis handles burst writes, worker batches inserts
- ✅ Context rich: Every event has tenant, user, IP, timing
- ✅ Flexible metadata: Additional fields don't require schema changes
- ✅ Fallback resilience: Local file logging if Redis unavailable

### Negative
- ⚠️ Eventual consistency: Audit logs may lag behind real-time
- ⚠️ Potential loss: Redis crash before worker processing loses events
- ⚠️ Storage growth: High-volume events require TTL/cleanup strategy
- ⚠️ Complexity: Additional infrastructure component to monitor

### Risk Mitigations
| Risk | Mitigation |
|------|------------|
| Redis data loss | Worker runs frequently (5s interval); AOF persistence enabled |
| MongoDB unavailability | Circuit breaker + local file fallback |
| Event flooding | Rate limiting per tenant; max queue size with drop policy |
| Tampering | Append-only MongoDB; separate admin credentials |

## Compliance
- **SOC 2 CC7.2**: System monitoring with audit trails
- **GDPR Article 30**: Records of processing activities
- **Immutable storage**: MongoDB collection with restricted delete permissions

## References
- See ADR 010 for MongoDB audit log storage details
- OWASP Logging Cheat Sheet
