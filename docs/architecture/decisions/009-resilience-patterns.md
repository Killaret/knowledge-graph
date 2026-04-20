# ADR 009: Resilience Patterns

## Status
Accepted

## Context
Multi-tenant SaaS requires resilience against:
- External service failures (AI APIs, email providers)
- Database connection exhaustion
- Cache unavailability (Redis outages)
- Cascading failures from dependency degradation

Without resilience patterns, single points of failure can cause total system outage.

## Problem
How do we implement resilience that:
1. Prevents cascade failures when dependencies degrade
2. Provides graceful degradation (reduced functionality vs total outage)
3. Automatically recovers when dependencies heal
4. Alerts operators to persistent issues
5. Maintains data integrity during failures

## Decision
**Circuit Breaker + Redis Queues + Fallback Strategies**

### Implementation

#### 1. Circuit Breaker Configuration (sony/gobreaker)
```go
package infrastructure

import (
    "github.com/sony/gobreaker"
    "time"
)

type CircuitBreakerConfig struct {
    Name              string
    MaxRequests       uint32        // Max requests in half-open state
    Interval          time.Duration // Statistical window
    Timeout           time.Duration // Request timeout
    FailureThreshold  uint32        // Errors to trigger open
    SuccessThreshold  uint32        // Successes to close from half-open
    OpenDuration      time.Duration // How long to stay open
}

func NewRedisCircuitBreaker() *gobreaker.CircuitBreaker {
    settings := gobreaker.Settings{
        Name:        "redis",
        MaxRequests: 3,  // Allow 3 test requests in half-open
        Interval:    30 * time.Second,
        Timeout:     5 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 5 && failureRatio >= 0.6
        },
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            log.Printf("Circuit breaker %s: %s -> %s", name, from, to)
            // Alert on state change
            metrics.RecordCircuitBreakerState(name, to)
        },
    }
    return gobreaker.NewCircuitBreaker(settings)
}

func NewMongoCircuitBreaker() *gobreaker.CircuitBreaker {
    settings := gobreaker.Settings{
        Name:        "mongodb",
        MaxRequests: 2,
        Interval:    30 * time.Second,
        Timeout:     10 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            return counts.ConsecutiveFailures > 5
        },
    }
    return gobreaker.NewCircuitBreaker(settings)
}
```

#### 2. Redis Client with Circuit Breaker
```go
type ResilientRedisClient struct {
    client    *redis.Client
    breaker   *gobreaker.CircuitBreaker
    fallback  LocalCache  // In-memory fallback
}

func (r *ResilientRedisClient) Get(ctx context.Context, key string) (string, error) {
    result, err := r.breaker.Execute(func() (interface{}, error) {
        return r.client.Get(ctx, key).Result()
    })
    
    if err != nil {
        // Circuit open or Redis error - use fallback
        if err == gobreaker.ErrOpenState {
            log.Warn("Redis circuit open, using local cache")
            metrics.Increment("redis.circuit_open")
        }
        return r.fallback.Get(key)
    }
    
    return result.(string), nil
}

func (r *ResilientRedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    _, err := r.breaker.Execute(func() (interface{}, error) {
        return r.client.Set(ctx, key, value, ttl).Result()
    })
    
    if err != nil {
        // Queue for later sync when Redis recovers
        return r.fallback.Set(key, value, ttl)
    }
    
    return nil
}
```

#### 3. MongoDB Client with Circuit Breaker
```go
type ResilientMongoClient struct {
    client    *mongo.Client
    breaker   *gobreaker.CircuitBreaker
    fallback  *sql.DB  // Postgres fallback for critical data
}

func (m *ResilientMongoClient) InsertAuditLog(ctx context.Context, log AuditLog) error {
    _, err := m.breaker.Execute(func() (interface{}, error) {
        return m.client.Database("audit").Collection("logs").InsertOne(ctx, log)
    })
    
    if err != nil {
        // Fallback: write to local file for later replay
        return m.writeToFallbackFile(log)
    }
    
    return nil
}

func (m *ResilientMongoClient) writeToFallbackFile(log AuditLog) error {
    data, _ := json.Marshal(log)
    line := fmt.Sprintf("%s\n", string(data))
    
    f, err := os.OpenFile("/var/log/mongodb-fallback/audit.log",
        os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    
    _, err = f.WriteString(line)
    return err
}
```

#### 4. Async Queue Resilience
```go
type ResilientQueue struct {
    redis     *ResilientRedisClient
    fallback  chan QueuedTask  // In-memory channel
    maxMemory int              // Max in-memory items
}

func (q *ResilientQueue) Enqueue(ctx context.Context, task QueuedTask) error {
    // Try Redis first
    if err := q.redis.Enqueue(ctx, task); err == nil {
        return nil
    }
    
    // Circuit open - use in-memory fallback
    select {
    case q.fallback <- task:
        log.Warn("Queued to memory fallback", "task", task.ID)
        return nil
    default:
        // Memory buffer full - drop task (metrics alert)
        metrics.Increment("queue.drop")
        return ErrQueueFull
    }
}

func (q *ResilientQueue) StartDrainWorker(ctx context.Context) {
    // Continuously try to drain memory queue to Redis
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case task := <-q.fallback:
            // Try to push to Redis
            if err := q.redis.Enqueue(ctx, task); err != nil {
                // Put it back, try later
                go func() {
                    time.Sleep(1 * time.Second)
                    q.fallback <- task
                }()
            }
        case <-ticker.C:
            // Check Redis health, metrics
        }
    }
}
```

#### 5. Worker Retry with Backoff
```go
type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
}

func ExecuteWithRetry(ctx context.Context, task QueuedTask, handler TaskHandler, config RetryConfig) error {
    var lastErr error
    
    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        err := handler.Handle(ctx, task)
        if err == nil {
            return nil // Success
        }
        
        lastErr = err
        
        // Check if error is retryable
        if !isRetryable(err) {
            return err // Permanent failure
        }
        
        // Exponential backoff with jitter
        delay := min(config.BaseDelay*time.Duration(1<<attempt), config.MaxDelay)
        jitter := time.Duration(rand.Int63n(int64(delay / 2)))
        time.Sleep(delay + jitter)
    }
    
    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func isRetryable(err error) bool {
    // Network errors, timeouts, rate limits are retryable
    var netErr net.Error
    if errors.As(err, &netErr) {
        return netErr.Temporary() || netErr.Timeout()
    }
    
    // Mongo/Redis connection errors
    if mongo.IsNetworkError(err) || redis.IsConnError(err) {
        return true
    }
    
    return false
}
```

## Consequences

### Positive
- ✅ Cascade prevention: Circuit breaker stops hammering failed services
- ✅ Graceful degradation: Fallbacks maintain partial functionality
- ✅ Auto-recovery: Systems heal when dependencies restore
- ✅ Observability: Circuit state changes emit metrics/alerts
- ✅ Data safety: Fallback files prevent audit log loss

### Negative
- ⚠️ Complexity: More code paths to test and maintain
- ⚠️ Stale data: Fallback cache may serve outdated information
- ⚠️ Memory pressure: In-memory queues can OOM under sustained failure
- ⚠️ Eventually consistent: Fallback writes sync later, creating temporary inconsistency

### Thresholds
| Component | Open Threshold | Half-Open Requests | Recovery |
|-----------|----------------|-------------------|------------|
| Redis | 5 errors / 30s | 3 test requests | 5 consecutive successes |
| MongoDB | 5 consecutive errors | 2 test requests | 3 consecutive successes |
| External API | 60% failure / 60s | 1 test request | 2 consecutive successes |

## Compliance
- Resilience aligns with SOC 2 availability requirements
- Circuit breaker state logged for incident analysis
- Fallback data retention documented for compliance

## References
- Nygard, M. "Release It!" (Circuit Breaker pattern)
- sony/gobreaker library documentation
- AWS Well-Architected: Reliability Pillar
