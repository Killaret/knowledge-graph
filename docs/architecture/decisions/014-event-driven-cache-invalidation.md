# ADR 014: Event-Driven Cache Invalidation

## Status
Accepted

## Context
Knowledge Graph system has multiple layers of caching to improve performance:
- In-memory caching in backend services
- Redis caching for graph data
- Browser caching for static assets
- CDN caching for API responses

As the system evolved with the addition of Graph Service (ADR-013), maintaining cache consistency across services became critical. When data changes (notes, links, user settings), cached data must be invalidated to prevent serving stale content.

### Current State Analysis
The system has:
- Main backend with in-memory cache
- Graph Service with Redis cache for graph data
- Frontend with browser cache
- Multiple services need to stay synchronized

Current challenges:
- **TTL-based invalidation**: Cache entries expire after fixed time, but stale data served before expiration
- **Manual cache clearing**: Ad-hoc cache clearing when issues detected
- **No coordination**: Services don't know when data changes in other services
- **Race conditions**: Multiple services may cache different versions of same data

## Problem Statement
How do we ensure cache consistency across multiple services when underlying data changes, without excessive polling or complex coordination logic?

## Decision Drivers
- **Consistency**: All services should see consistent data
- **Performance**: Cache invalidation should not impact write performance
- **Scalability**: Solution must work with multiple service instances
- **Reliability**: Cache invalidation events should not be lost
- **Simplicity**: Implementation should be maintainable
- **Latency**: Invalidations should propagate quickly

## Considered Options

### Option 1: Time-To-Live (TTL) Only
Rely solely on cache expiration times, no explicit invalidation.

**Pros:**
- ✅ Simplest implementation
- ✅ No coordination between services
- ✅ No network calls for invalidation

**Cons:**
- ❌ Stale data served until TTL expires
- ❌ Need to balance between freshness and performance
- ❌ Cannot invalidate immediately on critical changes
- ❌ Wasted cache resources (stale data stored)
- ❌ Poor user experience when data changes

### Option 2: Polling
Services periodically poll database for changes and invalidate cache accordingly.

**Pros:**
- ✅ Simple to implement
- ✅ Works with any database
- ✅ No additional infrastructure

**Cons:**
- ❌ High load on database (continuous polling)
- ❌ Stale data between polls
- ❌ Difficult to choose appropriate polling interval
- ❌ Doesn't scale well with many services
- ❌ Wasted resources (polling even when no changes)

### Option 3: Outbox Pattern
Write events to an outbox table in the same transaction as data changes, separate process reads and publishes events.

**Pros:**
- ✅ Guaranteed event delivery (transactional)
- ✅ Exactly-once semantics
- ✅ Events persist if publisher fails
- ✅ Can replay events if needed
- ✅ Strong consistency

**Cons:**
- ❌ Additional database write overhead
- ❌ Need separate outbox cleanup process
- ❌ More complex infrastructure
- ❌ Potential table bloat if cleanup fails
- ❌ Additional latency (outbox read + publish)
- ❌ Need to handle duplicate processing

### Option 4: Redis Pub/Sub
Publish cache invalidation events to Redis Pub/Sub channels, services subscribe to relevant channels.

**Pros:**
- ✅ Low latency (real-time propagation)
- ✅ Minimal overhead (lightweight messages)
- ✅ Already using Redis for caching
- ✅ Scales well with multiple subscribers
- ✅ Simple implementation
- ✅ No additional database writes

**Cons:**
- ❌ At-least-once delivery (possible duplicates)
- ❌ No persistence (messages lost if all subscribers down)
- ❌ No backpressure (publisher doesn't know if subscriber processed)
- ❌ Need to handle reconnection logic
- ❌ Not suitable for critical events requiring exactly-once

### Option 5: Message Queue (RabbitMQ/Kafka)
Use a dedicated message queue for cache invalidation events.

**Pros:**
- ✅ Persistent messages (survive restarts)
- ✅ Exactly-once or at-least-once semantics
- ✅ Backpressure support
- ✅ Dead letter queues for failed messages
- ✅ Scales to very high throughput
- ✅ Can replay events

**Cons:**
- ❌ Additional infrastructure to manage
- ❌ Higher latency than in-memory solutions
- ❌ More complex setup and maintenance
- ❌ Overkill for cache invalidation (not critical events)
- ❌ Additional cost if using managed service
- ❌ Learning curve for team

## Decision
**Chosen Approach: Option 4 - Redis Pub/Sub**

Redis Pub/Sub provides the best balance of simplicity, performance, and reliability for cache invalidation use case. Since cache invalidation is not a critical operation (worst case: stale data for short period), the at-least-once semantics and lack of persistence are acceptable trade-offs.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Main Backend (Go)                       │
│  ┌─────────────┐  ┌─────────────┐                          │
│  │ Note CRUD   │  │ Publisher    │                          │
│  └──────┬──────┘  └──────┬──────┘                          │
│         │                │                                  │
│         │ PostgreSQL     │ Publish to Redis                 │
│         │                │ channel: cache:invalidate       │
└─────────┼────────────────┼──────────────────────────────────┘
          │                │
          │                │
┌─────────▼────────────────▼──────────────────────────────────┐
│                      Redis                                  │
│  ┌─────────────┐  ┌─────────────┐                           │
│  │ Cache Store │  │ Pub/Sub     │                           │
│  └─────────────┘  └──────┬──────┘                           │
│                          │                                  │
│                          │ Subscribe                         │
└──────────────────────────┼──────────────────────────────────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
┌────────▼────────┐ ┌──────▼────────┐ ┌───▼──────────────┐
│  Graph Service  │ │  Frontend     │ │  Other Services  │
│  ┌───────────┐  │ │  (SSE/WebSocket)│ │                  │
│  │ Subscriber│  │ │  ┌─────────┐  │ │  ┌─────────────┐  │
│  └─────┬─────┘  │ │  │Subscriber│ │ │  │ Subscriber  │  │
│        │        │ │  └────┬────┘  │ │  └──────┬──────┘  │
│        │        │ │       │        │ │         │          │
│        │        │ │       │ Push   │ │         │          │
│        │ Invalidate    │ updates │ | Invalidate    │
│        │ Cache   │ │ to client│ │ Cache          │
└────────┼─────────┘ └────────┼──────┘ └────────┼─────────┘
         │                    │                    │
         ▼                    ▼                    ▼
    Clear Cache         Update UI             Clear Cache
```

### Implementation Details

#### 1. Event Schema
```json
{
  "type": "cache_invalidation",
  "timestamp": "2024-05-26T12:00:00Z",
  "entity_type": "note|link|user|graph",
  "entity_id": "uuid",
  "tenant_id": "uuid",
  "operation": "create|update|delete",
  "affected_paths": ["/api/v1/notes/123", "/graph/note/123"]
}
```

#### 2. Channel Naming Convention
- `cache:invalidate:all` - Global cache invalidation
- `cache:invalidate:note:{noteId}` - Specific note
- `cache:invalidate:user:{userId}` - User-specific data
- `cache:invalidate:graph:{noteId}` - Graph data for note

#### 3. Publisher Implementation (Go)
```go
type CacheInvalidationPublisher struct {
    redisClient *redis.Client
}

func (p *CacheInvalidationPublisher) PublishInvalidation(ctx context.Context, event CacheInvalidationEvent) error {
    channel := fmt.Sprintf("cache:invalidate:%s:%s", event.EntityType, event.EntityID)
    message, _ := json.Marshal(event)
    return p.redisClient.Publish(ctx, channel, message).Err()
}
```

#### 4. Subscriber Implementation (Go)
```go
type CacheInvalidationSubscriber struct {
    redisClient *redis.Client
    cache       Cache
    patterns    []string // ["cache:invalidate:*"]
}

func (s *CacheInvalidationSubscriber) Subscribe(ctx context.Context) error {
    pubsub := s.redisClient.PSubscribe(ctx, s.patterns...)
    defer pubsub.Close()

    for {
        msg, err := pubsub.ReceiveMessage(ctx)
        if err != nil {
            return err
        }

        var event CacheInvalidationEvent
        if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
            continue
        }

        s.handleInvalidation(ctx, event)
    }
}

func (s *CacheInvalidationSubscriber) handleInvalidation(ctx context.Context, event CacheInvalidationEvent) {
    // Deduplicate: track recently processed events
    if s.isRecentlyProcessed(event) {
        return
    }

    // Invalidate cache
    s.cache.Delete(ctx, event.EntityID)
}
```

#### 5. Confirmation of Processing
Since Redis Pub/Sub doesn't provide delivery confirmation, we implement:
- **Deduplication**: Track processed event IDs in memory (short TTL)
- **Idempotent handlers**: Cache invalidation is idempotent (safe to process multiple times)
- **Health checks**: Monitor subscriber lag via metrics

#### 6. Reconnection Logic
```go
func (s *CacheInvalidationSubscriber) Start(ctx context.Context) {
    for {
        err := s.Subscribe(ctx)
        if err != nil {
            log.Printf("Subscription failed: %v", err)
            time.Sleep(5 * time.Second) // Backoff
            continue
        }
    }
}
```

### Failure Scenarios

#### Scenario 1: Subscriber Down
- **Problem**: Subscriber misses invalidation events while down
- **Mitigation**: Cache TTL acts as safety net (stale data expires eventually)
- **Monitoring**: Alert if subscriber is down for extended period

#### Scenario 2: Redis Restart
- **Problem**: All Pub/Sub state lost during restart
- **Mitigation**: Subscribers automatically reconnect on restart
- **Monitoring**: Alert on Redis restarts

#### Scenario 3: Duplicate Events
- **Problem**: Network issues may cause duplicate delivery
- **Mitigation**: Idempotent handlers + deduplication
- **Impact**: Minimal (cache invalidation is idempotent)

#### Scenario 4: Publisher Failure
- **Problem**: Publisher fails to send invalidation event
- **Mitigation**: Non-critical (TTL provides eventual consistency)
- **Monitoring**: Track publish success rate

## Consequences

### Positive Consequences
- ✅ **Real-time cache invalidation**: Cache cleared immediately on data changes
- ✅ **Low overhead**: Minimal performance impact on publishers
- ✅ **Scalable**: Works with multiple service instances
- ✅ **Simple**: Easy to implement and maintain
- ✅ **No additional infrastructure**: Uses existing Redis
- ✅ **Flexible**: Can add new subscribers without changing publishers

### Negative Consequences
- ❌ **At-least-once delivery**: Possible duplicate invalidations (mitigated by idempotency)
- ❌ **No persistence**: Events lost if all subscribers down (mitigated by TTL)
- ❌ **No backpressure**: Publishers don't know if subscribers are overwhelmed
- ❌ **Reconnection complexity**: Need robust reconnection logic
- ❌ **Monitoring complexity**: Need to track subscription health

### Mitigation Strategies
- **Duplicate events**: Idempotent cache invalidation + in-memory deduplication
- **Event loss**: Cache TTL as safety net + monitoring for subscriber downtime
- **No backpressure**: Monitor subscriber lag, alert if processing is slow
- **Reconnection**: Exponential backoff + health checks
- **Monitoring**: Prometheus metrics for publish/subscribe rates, subscriber lag

## When to Reconsider
- If cache invalidation becomes critical (requiring exactly-once semantics)
- If event loss becomes unacceptable (need persistence)
- If system scales to point where Redis Pub/Sub becomes bottleneck
- If need complex event routing/filtering beyond simple pub/sub

## Alternatives for Future
- **Outbox pattern**: If need guaranteed delivery
- **Message queue**: If need backpressure and complex routing
- **Change Data Capture (CDC)**: If need database-level change notifications

## References
- [ADR 013: Graph Service Isolation](./013-graph-service-isolation.md)
- [Redis Pub/Sub Documentation](https://redis.io/docs/manual/pubsub/)
- [Cache Invalidation Patterns](https://docs.microsoft.com/en-us/azure/architecture/patterns/cache-aside)
