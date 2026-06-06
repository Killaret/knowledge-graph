# knowledge-graph-performance

**Version:** 1.0  
**Purpose:** Performance optimization across the stack  
**Status:** Active  
**Priority:** 🔴 Critical (Highest ROI)

---

## Overview

`knowledge-graph-performance` specializes in identifying and resolving performance bottlenecks across the entire Knowledge Graph application stack.

**Key Areas:**
- Frontend performance (bundle size, rendering, lazy loading)
- Backend performance (API response time, database queries)
- 3D visualization optimization (Three.js, GraphCanvas)
- Database optimization (indexes, query plans, connection pooling)
- Caching strategies (Redis, in-memory)
- Network optimization (compression, CDN, API response size)

---

## Performance Patterns & Best Practices

### 1. Frontend Performance

#### Bundle Size Optimization

**Target:** < 500KB gzipped for main bundle

**Techniques:**
```typescript
// Lazy loading routes
const Graph3DPage = lazy(() => import('./routes/graph/3d/+page.svelte'));

// Code splitting
import({ /* webpackChunkName: "graph3d" */ './Graph3D.svelte' });

// Tree shaking - avoid side effects
export { GraphCanvas } from './GraphCanvas.svelte'; // Not: import ... from
```

**Metrics to Track:**
```bash
# Analyze bundle
npm run build
npx source-map-explorer 'build/static/js/*.js'

# Check bundle size
ls -lh build/static/js/*.js
```

#### Rendering Optimization

**Target:** 60 FPS for graph visualization

**Techniques:**
```typescript
// Virtual scrolling for large lists
{#each visibleNodes as node}
  <NodeCard {node} />
{/each}

// RequestAnimationFrame for animations
function animate() {
  requestAnimationFrame(animate);
  updateGraph();
}

// Debounce expensive operations
const debouncedSearch = debounce((query) => {
  searchNodes(query);
}, 300);

// Memoize calculations
const memoizedCalculation = memoize((data) => {
  // expensive computation
});
```

**Three.js Specific:**
```typescript
// Level of Detail (LOD)
const lod = new THREE.LOD();
lod.addLevel(detailHigh, 0);
lod.addLevel(detailMedium, 100);
lod.addLevel(detailLow, 500);

// InstancedMesh for repeated objects
const instancedMesh = new THREE.InstancedMesh(geometry, material, count);

// Frustum culling
object.frustumCulled = true; // Enable automatic culling

// BufferGeometry optimizations
geometry.setDrawRange(0, visibleCount);
```

#### Memory Management

**Target:** < 100MB memory usage

**Techniques:**
```typescript
// Cleanup on unmount
onMount(() => {
  const canvas = createCanvas();
  return () => {
    canvas.dispose(); // Cleanup WebGL resources
  };
});

// Avoid memory leaks
let cache = new Map();
setInterval(() => {
  if (cache.size > 1000) {
    // Evict old entries
    const keys = Array.from(cache.keys()).slice(0, 100);
    keys.forEach(k => cache.delete(k));
  }
}, 60000);

// WeakMap for caches
const cache = new WeakMap(); // Auto GC when objects collected
```

---

### 2. Backend Performance

#### API Response Time

**Target:** < 100ms for most endpoints, < 500ms for complex queries

**Techniques:**
```go
// Middleware for timing
func TimingMiddleware(next gin.HandlerFunc) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        next(c)
        duration := time.Since(start)
        log.Printf("[%s] %s %v", c.Request.Method, c.Request.URL, duration)
    }
}

// Context timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Concurrent processing
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(item Item) {
        defer wg.Done()
        process(item)
    }(item)
}
wg.Wait()
```

#### Database Optimization

**Target:** < 50ms for queries, < 100ms for complex joins

**Techniques:**
```sql
-- Indexes for frequent queries
CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_notes_created_at ON notes(created_at DESC);
CREATE INDEX idx_links_source_target ON links(source_note_id, target_note_id);

-- Composite indexes
CREATE INDEX idx_notes_user_type ON notes(user_id, type);

-- Explain query plan
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM notes WHERE user_id = '...';

-- Use pgvector indexes
CREATE INDEX ON embeddings USING ivfflat (embedding vector_cosine_ops) 
WITH (lists = 100);
```

```go
// Connection pooling
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)

// Query optimization with GORM
db.Preload("Links").Find(&notes) // Eager loading
db.Select("id, title").Find(&notes) // Only needed fields

// Batch inserts
db.CreateInBatches(&notes, 100) // 100 records per batch
```

#### Caching Strategies

**Target:** 90% cache hit rate

**Techniques:**
```go
// Redis cache with TTL
func (s *Service) GetGraph(ctx context.Context, userID string) (*Graph, error) {
    cacheKey := fmt.Sprintf("graph:%s", userID)
    
    // Try cache first
    cached, err := s.redis.Get(ctx, cacheKey).Bytes()
    if err == nil {
        return deserialize(cached), nil
    }
    
    // Cache miss - fetch from DB
    graph := s.db.GetGraph(ctx, userID)
    
    // Store in cache (5 minutes TTL)
    s.redis.SetEx(ctx, cacheKey, serialize(graph), 5*time.Minute)
    
    return graph, nil
}

// Cache-aside pattern
func cacheWithTTL(key string, ttl time.Duration, fetch func() ([]byte, error)) ([]byte, error) {
    cached, err := redis.Get(key).Bytes()
    if err == nil {
        return cached, nil
    }
    
    data, err := fetch()
    if err != nil {
        return nil, err
    }
    
    redis.SetEx(key, data, ttl)
    return data, nil
}

// Write-through cache
func saveWithCache(key string, data []byte, ttl time.Duration) error {
    // Write to DB first
    err := db.Save(data)
    if err != nil {
        return err
    }
    
    // Then update cache
    return redis.SetEx(key, data, ttl)
}
```

**Cache Invalidation:**
```go
// Invalidate on write
func updateNote(note *Note) error {
    // Update DB
    err := db.Update(note)
    if err != nil {
        return err
    }
    
    // Invalidate related caches
    redis.Del(ctx, fmt.Sprintf("graph:%s", note.UserID))
    redis.Del(ctx, fmt.Sprintf("note:%s", note.ID))
    redis.Del(ctx, fmt.Sprintf("recommendations:%s", note.ID))
    
    return nil
}

// Cache versioning
cacheKey := fmt.Sprintf("graph:%s:v%d", userID, currentVersion)
```

---

### 3. 3D Visualization Performance

#### GraphCanvas Optimization

**Target:** 60 FPS with 500+ nodes

**Techniques:**
```typescript
// Progressive loading
async function loadGraph() {
  // Load first 100 nodes immediately
  const initialNodes = await fetchNodes({ limit: 100 });
  renderGraph(initialNodes);
  
  // Load rest in background
  const allNodes = await fetchNodes({ limit: 1000 });
  updateGraph(allNodes);
}

// LOD for nodes
function getNodeDetail(distance: number) {
  if (distance < 100) return 'high';
  if (distance < 500) return 'medium';
  return 'low';
}

// Instanced rendering for nodes
const instancedNodes = new THREE.InstancedMesh(
  sphereGeometry,
  material,
  nodeCount
);

// Only render visible nodes
const visibleNodes = cameraFrustum.intersectsObjects(allNodes);
render(visibleNodes);

// WebWorker for layout calculations
const worker = new Worker('graph-layout.worker.ts');
worker.postMessage({ nodes, links });
worker.onmessage = (e) => updatePositions(e.data);
```

#### Three.js Specific Optimizations

```typescript
// Use BufferGeometry
const geometry = new THREE.BufferGeometry();
geometry.setAttribute('position', new THREE.Float32BufferAttribute(coords, 3));

// Merge geometries
const mergedGeometry = mergeGeometries([geo1, geo2, geo3]);

// Reuse materials
const sharedMaterial = new THREE.MeshStandardMaterial({
  color: 0xffcc00,
  metalness: 0.5,
  roughness: 0.5
});

// Disable shadows for distant objects
if (distance > 500) {
  mesh.castShadow = false;
  mesh.receiveShadow = false;
}

// Use CSS2DRenderer for labels (cheaper than 3D text)
const labelRenderer = new CSS2DRenderer();
labelRenderer.setSize(window.innerWidth, window.innerHeight);
document.body.appendChild(labelRenderer.domElement);
```

---

### 4. Network Optimization

#### API Response Optimization

**Target:** < 100KB response size for most endpoints

**Techniques:**
```go
// Compression middleware
func CompressionMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Content-Encoding", "gzip")
        gzipWriter := gzip.NewWriter(c.Writer)
        defer gzipWriter.Close()
        c.Writer = gzipWriter
        c.Next()
    }
}

// Response caching
c.Header("Cache-Control", "public, max-age=300") // 5 minutes

// Pagination
func GetNotes(c *gin.Context) {
    limit := min(c.DefaultQuery("limit", "100"), 1000)
    offset := c.DefaultQuery("offset", "0")
    
    notes, total := db.GetNotes(limit, offset)
    
    c.JSON(200, gin.H{
        "notes": notes,
        "total": total,
        "limit": limit,
        "offset": offset,
    })
}

// Field selection
func GetNote(c *gin.Context) {
    fields := c.QueryArray("fields") // ?fields=id,title,content
    note := db.GetNoteWithFields(id, fields)
    c.JSON(200, note)
}
```

#### WebSocket Optimization

```go
// Throttle updates
var throttle = time.NewTicker(100 * time.Millisecond)
defer throttle.Stop()

for {
    select {
    case <-throttle.C:
        broadcast(update)
    case update := <-updates:
        // Buffer updates
    }
}

// Binary messages
websocket.SetCompressionLevel(9) // Maximum compression
```

---

## Performance Checklist

### Frontend

- [ ] Bundle size < 500KB gzipped
- [ ] Lighthouse Performance score > 90
- [ ] 60 FPS during interactions
- [ ] Memory usage < 100MB
- [ ] First Contentful Paint < 1.5s
- [ ] Time to Interactive < 3.5s

### Backend

- [ ] API response time < 100ms (p95)
- [ ] Database query time < 50ms (p95)
- [ ] Cache hit rate > 90%
- [ ] Connection pool utilization < 80%
- [ ] Error rate < 0.1%

### 3D Visualization

- [ ] 60 FPS with 500+ nodes
- [ ] Initial render < 2s
- [ ] Memory usage < 200MB
- [ ] No frame drops during pan/zoom
- [ ] Progressive loading implemented

---

## Performance Testing

### Frontend Testing

```bash
# Lighthouse audit
npm run build
npx serve build
npx lighthouse http://localhost:3000 --view

# Bundle analysis
npm run build
npx source-map-explorer 'build/static/js/*.js'

# Web Vitals
npm install web-vitals
# Add to main.ts
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';
getCLS(console.log);
getFID(console.log);
getFCP(console.log);
getLCP(console.log);
getTTFB(console.log);
```

### Backend Testing

```bash
# Load testing with hey
hey -n 1000 -c 100 http://localhost:8080/api/v1/notes

# Benchmark with vegeta
echo "GET http://localhost:8080/api/v1/notes" | vegeta attack -duration=30s | vegeta report
vegeta report -type=report > report.txt

# Database query profiling
EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) 
SELECT * FROM notes WHERE user_id = 'uuid';
```

### 3D Performance Testing

```typescript
// FPS monitoring
class FPSMonitor {
  private frames = 0;
  private lastTime = performance.now();
  
  update() {
    this.frames++;
    const now = performance.now();
    if (now - this.lastTime >= 1000) {
      const fps = Math.round(this.frames * 1000 / (now - this.lastTime));
      console.log(`FPS: ${fps}`);
      this.frames = 0;
      this.lastTime = now;
    }
  }
}

// Memory monitoring
const memory = performance.memory;
console.log(`Used JS: ${memory.usedJSHeapSize / 1024 / 1024}MB`);
```

---

## Common Performance Issues & Solutions

### Issue 1: Slow Graph Loading

**Symptom:** Initial load > 5s for 500+ nodes

**Diagnosis:**
```sql
-- Check query time
EXPLAIN ANALYZE SELECT * FROM notes WHERE user_id = '...';

-- Check missing indexes
SELECT * FROM pg_stat_user_tables WHERE seq_scan > 0;
```

**Solution:**
```sql
-- Add indexes
CREATE INDEX idx_notes_user_id ON notes(user_id);
CREATE INDEX idx_links_user ON links(user_id);

-- Add pagination
LIMIT 100 OFFSET 0
```

```typescript
// Frontend: Progressive loading
const batch1 = await fetch('/api/v1/me/graph?limit=100');
renderGraph(batch1);
const batch2 = await fetch('/api/v1/me/graph?limit=100&offset=100');
updateGraph(batch2);
```

---

### Issue 2: Low FPS in 3D Graph

**Symptom:** FPS < 30 during pan/zoom

**Diagnosis:**
- Check render time (Chrome DevTools Performance tab)
- Check GPU usage
- Check draw calls

**Solution:**
```typescript
// Use InstancedMesh
const instancedMesh = new THREE.InstancedMesh(geometry, material, nodeCount);

// Reduce geometry complexity
const geometry = new THREE.SphereGeometry(1, 16, 16); // Lower segments

// Implement LOD
const lod = new THREE.LOD();
lod.addLevel(highDetail, 0);
lod.addLevel(mediumDetail, 200);
lod.addLevel(lowDetail, 500);

// Disable shadows for distant objects
if (distance > 300) {
  mesh.castShadow = false;
}
```

---

### Issue 3: High Memory Usage

**Symptom:** Memory > 200MB and growing

**Diagnosis:**
```typescript
// Check for memory leaks
window.addEventListener('beforeunload', () => {
  console.log('Memory:', performance.memory?.usedJSHeapSize);
});

// Use Chrome DevTools Memory tab
// Take heap snapshots and compare
```

**Solution:**
```typescript
// Cleanup event listeners
onDestroy(() => {
  element.removeEventListener('click', handler);
});

// Dispose Three.js resources
mesh.geometry.dispose();
mesh.material.dispose();

// Clear caches periodically
setInterval(() => {
  cache.clearOlderThan(1 * 60 * 60 * 1000); // 1 hour
}, 10 * 60 * 1000); // Every 10 minutes
```

---

## Performance Monitoring

### Metrics to Track

```go
// Prometheus metrics
var (
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
        },
        []string{"method", "endpoint"},
    )
    
    graphRenderDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "graph_render_duration_seconds",
            Help:    "Graph rendering duration",
        },
    )
    
    cacheHitRatio = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name:    "cache_hit_ratio",
            Help:    "Cache hit ratio",
        },
    )
)
```

```typescript
// Frontend Web Vitals
import { getCLS, getFID, getFCP, getLCP, getTTFB } from 'web-vitals';

getCLS(console.log);   // Cumulative Layout Shift
getFID(console.log);   // First Input Delay
getFCP(console.log);   // First Contentful Paint
getLCP(console.log);   // Largest Contentful Paint
getTTFB(console.log);  // Time to First Byte
```

### Alerting Thresholds

```yaml
# Prometheus alerting rules
groups:
  - name: performance
    rules:
      - alert: HighAPIResponseTime
        expr: histogram_quantile(0.95, http_request_duration_seconds) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "API response time > 500ms"
          
      - alert: LowFPS
        expr: graph_fps < 30
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Graph FPS < 30"
```

---

## Performance Budgets

### Frontend Budgets

```javascript
// vite.config.js
export default {
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'three': ['three'],
          'svelte': ['svelte'],
        }
      }
    }
  }
};

// Bundle budget
// Main bundle: < 200KB
// Three.js chunk: < 150KB
// Other chunks: < 100KB each
```

### Backend Budgets

```yaml
# API response time budgets
endpoints:
  GET /api/v1/notes:
    budget: 50ms
    p95: 100ms
    p99: 200ms
    
  GET /api/v1/me/graph:
    budget: 100ms
    p95: 200ms
    p99: 500ms
    
  POST /api/v1/notes:
    budget: 100ms
    p95: 200ms
    p99: 300ms
```

### 3D Budgets

```typescript
const PERFORMANCE_BUDGETS = {
  fps: {
    target: 60,
    min: 30,
    critical: 15
  },
  memory: {
    target: 100, // MB
    max: 200,    // MB
    critical: 300 // MB
  },
  renderTime: {
    target: 8, // ms (60 FPS)
    max: 33,   // ms (30 FPS)
    critical: 67 // ms (15 FPS)
  }
};
```

---

## References

- [Web Vitals](https://web.dev/vitals/)
- [Three.js Performance](https://threejs.org/docs/#manual/en/introduction/Creating-a-scene)
- [Redis Caching Patterns](https://redis.io/topics/cache)
- [PostgreSQL Query Optimization](https://www.postgresql.org/docs/current/performance-tips.html)
- [Go Performance Best Practices](https://github.com/golang/go/wiki/CodeReviewComments)

---

**Last Updated:** 2026-05-22  
**Maintainer:** knowledge-graph-docs-maintenance  
**Version:** 1.0
