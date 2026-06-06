# Инструменты Performance Агента

## 🎯 Основные задачи

1. Анализ производительности
2. Оптимизация кода
3. Load testing
4. Profiling
5. Memory management

---

## 📊 Анализ производительности

### 1. Backend Profiling

#### CPU Profiling
```powershell
# Включить profiling
cd backend
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/profile?seconds=30

# Анализ snapshot
go tool pprof -http=:8082 ./profiles/cpu.prof
```

#### Memory Profiling
```powershell
# Собрать memory profile
curl -o memory.prof http://localhost:8080/debug/pprof/heap

# Анализ
go tool pprof -http=:8083 ./memory.prof

# Выявить leaks
go tool pprof -alloc_space -http=:8084 ./memory.prof
```

#### Trace Analysis
```powershell
# Собрать trace
go tool trace -tracefile=trace.out http://localhost:8080/debug/pprof/trace?seconds=30

# Просмотреть
go tool trace trace.out
```

### 2. Frontend Performance

#### Lighthouse Audit
```powershell
# Запуск lighthouse
npm run test:lighthouse

# Конкретные метрики
lhci autorun --collect.numberOfRuns=3
```

#### Web Vitals
```javascript
// Отслеживание Core Web Vitals
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

getCLS(console.log);
getFID(console.log);
getFCP(console.log);
getLCP(console.log);
getTTFB(console.log);
```

#### Bundle Analysis
```powershell
# Анализ bundle size
npm run build -- --stats
webpack-bundle-analyzer build/stats.json

# Treeshaking check
npm run build -- --stats-json
```

---

## 🧪 Load Testing

### 1. API Load Tests

#### Wrk Benchmark
```powershell
# Simple benchmark
wrk -t12 -c400 -d30s http://localhost:8080/api/v1/graph/full

# С параметрами
wrk -t8 -c200 -d60s -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/graph/note/1?depth=2

# POST request
wrk -t12 -c400 -d30s -H "Content-Type: application/json" \
  -s post.lua http://localhost:8080/api/v1/auth/login
```

#### post.lua script
```lua
wrk.method = "POST"
wrk.body   = '{"email":"test@test.com","password":"password"}'
wrk.headers["Content-Type"] = "application/json"
```

#### k6 Load Testing
```javascript
// load-test.js
import http from 'k6';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 100 },  // Ramp-up
    { duration: '1m', target: 100 },   // Stay at 100
    { duration: '30s', target: 0 },    // Ramp-down
  ],
};

export default function () {
  let res = http.get('http://localhost:8080/api/v1/graph/full');
  check(res, { 'status is 200': (r) => r.status === 200 });
  sleep(1);
}
```

```powershell
# Запуск k6
k6 run load-test.js
k6 run --out influxdb=http://localhost:8086/k6 load-test.js
```

### 2. Database Performance

#### PostgreSQL Query Analysis
```sql
-- Включить query logging
SET log_min_duration_statement = 1000;

-- ANALYZE запрос
EXPLAIN ANALYZE SELECT * FROM notes WHERE id = '123';

-- Найти медленные запросы
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;
```

#### Connection Pool Testing
```powershell
# Проверка пула соединений
docker compose exec postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

-- Настройка pool size
ALTER SYSTEM SET max_connections = 200;
SELECT pg_reload_conf();
```

---

## 🔍 Оптимизация

### 1. Backend Optimizations

#### Query Optimization
```go
// ❌ N+1 query problem
for _, note := range notes {
    links := db.Where("note_id = ?", note.ID).Find(&Links{})
}

// ✅ Eager loading
var notes []Note
db.Preload("Links").Find(&notes)

// ✅ Batch query
db.Where("id IN ?", noteIDs).Find(&links)
```

#### Caching Strategy
```go
// Redis cache with TTL
func (s *GraphService) GetGraph(noteID string, depth int) (*Graph, error) {
    cacheKey := fmt.Sprintf("graph:%s:%d", noteID, depth)
    
    // Try cache first
    cached, err := s.redis.Get(cacheKey).Result()
    if err == nil {
        return deserializeGraph(cached)
    }
    
    // Fetch from DB
    graph, err := s.repo.GetGraph(noteID, depth)
    if err != nil {
        return nil, err
    }
    
    // Cache for 5 minutes
    s.redis.SetEx(cacheKey, serializeGraph(graph), 5*time.Minute)
    
    return graph, nil
}
```

#### Concurrent Processing
```go
// Parallel processing with goroutines
func fetchAllNodes(nodeIDs []string) ([]Node, error) {
    var wg sync.WaitGroup
    result := make([]Node, len(nodeIDs))
    errChan := make(chan error, len(nodeIDs))
    
    for i, id := range nodeIDs {
        wg.Add(1)
        go func(idx int, nodeID string) {
            defer wg.Done()
            node, err := fetchNode(nodeID)
            if err != nil {
                errChan <- err
                return
            }
            result[idx] = node
        }(i, id)
    }
    
    wg.Wait()
    close(errChan)
    
    // Check errors
    for err := range errChan {
        if err != nil {
            return nil, err
        }
    }
    
    return result, nil
}
```

### 2. Frontend Optimizations

#### Code Splitting
```javascript
// Lazy loading routes
const GraphView = lazy(() => import('./GraphView'));
const NoteEditor = lazy(() => import('./NoteEditor'));

// Route-based splitting
const routes = [
  { path: '/graph', component: GraphView },
  { path: '/notes', component: NotesList }
];
```

#### Memoization
```svelte
<script>
  import { memo } from 'svelte';
  
  let heavyComputation = memo(() => {
    // Expensive calculation
    return expensiveCalculation(input());
  });
</script>
```

#### Virtual Scrolling
```javascript
// For large lists
import { VirtualList } from 'svelte-virtual-list';

<VirtualList
  items={largeDataset}
  height={500}
  itemHeight={50}
  renderItem={(item) => <div>{item.title}</div>}
/>
```

---

## 📈 Метрики производительности

### Backend Metrics

```go
// Prometheus metrics
var (
    httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "Duration of HTTP requests",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "endpoint"})
    
    httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"method", "endpoint", "status"})
    
    activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "active_connections",
        Help: "Number of active connections",
    })
)

// Middleware
func PrometheusMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(
            time.Since(start).Seconds())
    })
}
```

### Frontend Metrics

```javascript
// Performance Observer
const observer = new PerformanceObserver((list) => {
  for (const entry of list.getEntries()) {
    console.log(`${entry.name}: ${entry.startTime.toFixed(2)}ms`);
  }
});

observer.observe({ entryTypes: ['paint', 'largest-contentful-paint'] });

// Custom metrics
window.performance.mark('graph-load-start');
// ... load graph ...
window.performance.mark('graph-load-end');
window.performance.measure('graph-load', 
  'graph-load-start', 'graph-load-end');
```

---

## 🎯 Цели производительности

### API Response Times
- **p50** (median): < 100ms
- **p95**: < 500ms
- **p99**: < 1000ms

### Database Queries
- **Simple SELECT**: < 50ms
- **Complex JOIN**: < 200ms
- **Aggregations**: < 500ms

### Frontend Metrics
- **FCP** (First Contentful Paint): < 1.5s
- **LCP** (Largest Contentful Paint): < 2.5s
- **CLS** (Cumulative Layout Shift): < 0.1
- **FID** (First Input Delay): < 100ms

### System Resources
- **CPU usage**: < 70%
- **Memory usage**: < 80%
- **Disk I/O**: < 70%
- **Network I/O**: < 80%

---

## 🔧 Оптимизационные паттерны

### 1. Lazy Loading
```go
// Lazy initialization
type LazyDB struct {
    mu   sync.RWMutex
    conn *sql.DB
    dsn  string
}

func (l *LazyDB) GetConn() (*sql.DB, error) {
    l.mu.RLock()
    conn := l.conn
    l.mu.RUnlock()
    
    if conn != nil {
        return conn, nil
    }
    
    l.mu.Lock()
    defer l.mu.Unlock()
    
    if l.conn == nil {
        var err error
        l.conn, err = sql.Open("postgres", l.dsn)
        if err != nil {
            return nil, err
        }
    }
    return l.conn, nil
}
```

### 2. Connection Pooling
```go
// Optimize DB pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

### 3. Batch Operations
```go
// Batch inserts
func InsertBatch(notes []Note) error {
    batch := 100
    for i := 0; i < len(notes); i += batch {
        end := i + batch
        if end > len(notes) {
            end = len(notes)
        }
        db.Create(notes[i:end])
    }
    return nil
}
```
