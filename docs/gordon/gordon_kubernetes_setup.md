# Knowledge Graph — Kubernetes & Monitoring Setup

**Версия:** 1.0 | **Для:** Production Deployment on K8s

---

## Часть 1: Kubernetes Manifests

### 1.1 Backend Deployment

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kg-backend-config
  namespace: knowledge-graph
data:
  DATABASE_CPU_CORES: "8"
  DATABASE_MAX_OPEN_CONNS: "50"
  DATABASE_MAX_IDLE_CONNS: "25"
  DATABASE_CONN_MAX_LIFETIME: "10m"
  DATABASE_QUERY_TIMEOUT: "30s"
  NLP_HTTP_TIMEOUT: "10s"
  NLP_CIRCUIT_BREAKER_THRESHOLD: "5"
  NLP_CIRCUIT_BREAKER_TIMEOUT: "30s"
  REDIS_URL: "redis-master.kg-system.svc.cluster.local:6379"
  NLP_SERVICE_URL: "http://kg-nlp.kg-system.svc.cluster.local:5000"
  APP_ENV: "production"

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kg-backend
  namespace: knowledge-graph
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
  selector:
    matchLabels:
      app: kg-backend
  template:
    metadata:
      labels:
        app: kg-backend
    spec:
      initContainers:
      - name: migrate
        image: kg-backend:1.0.0
        command: ["/app/migrate", "up"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: kg-secrets
              key: database-url
      
      containers:
      - name: backend
        image: kg-backend:1.0.0
        ports:
        - containerPort: 8080
          name: http
        
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi
        
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          failureThreshold: 3
        
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
          failureThreshold: 2
        
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: kg-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: kg-secrets
              key: jwt-secret
        envFrom:
        - configMapRef:
            name: kg-backend-config
      
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - kg-backend
              topologyKey: kubernetes.io/hostname

---
apiVersion: v1
kind: Service
metadata:
  name: kg-backend
  namespace: knowledge-graph
spec:
  type: ClusterIP
  selector:
    app: kg-backend
  ports:
  - port: 8080
    targetPort: http
    protocol: TCP
```

### 1.2 NLP Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kg-nlp
  namespace: knowledge-graph
spec:
  replicas: 2
  selector:
    matchLabels:
      app: kg-nlp
  template:
    metadata:
      labels:
        app: kg-nlp
    spec:
      containers:
      - name: nlp
        image: kg-nlp:1.0.0
        ports:
        - containerPort: 5000
        
        resources:
          requests:
            cpu: 2000m
            memory: 2Gi
          limits:
            cpu: 4000m
            memory: 4Gi
        
        livenessProbe:
          httpGet:
            path: /health
            port: 5000
          initialDelaySeconds: 180
          periodSeconds: 15
          failureThreshold: 3
        
        readinessProbe:
          httpGet:
            path: /ready
            port: 5000
          initialDelaySeconds: 60
          periodSeconds: 10
          failureThreshold: 2
        
        startupProbe:
          httpGet:
            path: /health
            port: 5000
          initialDelaySeconds: 0
          periodSeconds: 10
          failureThreshold: 18
        
        env:
        - name: HF_HOME
          value: /cache/huggingface
        - name: WORKERS
          value: "4"
        
        volumeMounts:
        - name: model-cache
          mountPath: /cache/huggingface
      
      volumes:
      - name: model-cache
        persistentVolumeClaim:
          claimName: kg-nlp-cache-pvc

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: kg-nlp-cache-pvc
  namespace: knowledge-graph
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 5Gi
```

### 1.3 Redis Deployment

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  namespace: knowledge-graph
data:
  redis.conf: |
    maxmemory 2gb
    maxmemory-policy allkeys-lru
    appendonly yes
    save 900 1

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kg-redis
  namespace: knowledge-graph
spec:
  replicas: 1
  selector:
    matchLabels:
      app: kg-redis
  template:
    metadata:
      labels:
        app: kg-redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command:
          - redis-server
          - /usr/local/etc/redis/redis.conf
        ports:
        - containerPort: 6379
        
        resources:
          requests:
            cpu: 500m
            memory: 2Gi
          limits:
            cpu: 1000m
            memory: 3Gi
        
        livenessProbe:
          tcpSocket:
            port: 6379
          initialDelaySeconds: 30
          periodSeconds: 10
        
        volumeMounts:
        - name: config
          mountPath: /usr/local/etc/redis
        - name: data
          mountPath: /data
      
      volumes:
      - name: config
        configMap:
          name: redis-config
      - name: data
        persistentVolumeClaim:
          claimName: kg-redis-pvc

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: kg-redis-pvc
  namespace: knowledge-graph
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 10Gi
```

---

## Часть 2: Prometheus Alerting Rules

**prometheus-rules.yaml:**

```yaml
groups:
- name: knowledge-graph-alerts
  interval: 15s
  rules:
  
  # Backend
  - alert: BackendErrorRate
    expr: rate(backend_requests_total{status=~"5.."}[5m]) > 0.01
    for: 5m
    annotations:
      summary: "Backend error rate > 1%"
  
  - alert: BackendHighLatency
    expr: histogram_quantile(0.95, backend_request_duration_seconds) > 1
    for: 5m
    annotations:
      summary: "Backend p95 latency > 1s"
  
  # Database
  - alert: DBPoolExhausted
    expr: db_pool_utilization_percent > 90
    for: 2m
    annotations:
      summary: "Database pool utilization > 90%"
  
  - alert: PostgreSQLDown
    expr: up{job="postgresql"} == 0
    for: 1m
    annotations:
      summary: "PostgreSQL is down"
  
  # NLP
  - alert: NLPDown
    expr: up{job="kg-nlp"} == 0
    for: 1m
    annotations:
      summary: "NLP service is down"
  
  - alert: NLPHighLatency
    expr: histogram_quantile(0.95, nlp_request_duration_seconds) > 5
    for: 5m
    annotations:
      summary: "NLP p95 latency > 5s"
  
  - alert: NLPCircuitBreakerOpen
    expr: nlp_circuit_breaker_state{state="open"} == 1
    for: 1m
    annotations:
      summary: "NLP circuit breaker is open"
  
  # Redis
  - alert: RedisDown
    expr: up{job="redis"} == 0
    for: 1m
    annotations:
      summary: "Redis is down"
  
  - alert: RedisHighMemory
    expr: redis_memory_used_bytes / redis_memory_max_bytes > 0.9
    for: 5m
    annotations:
      summary: "Redis memory > 90%"
  
  # Queue
  - alert: QueueDepthHigh
    expr: redis_list_length{queue="default"} > 1000
    for: 10m
    annotations:
      summary: "Job queue depth > 1000"
```

---

## Часть 3: Deployment Script

**deploy.sh:**

```bash
#!/bin/bash
set -e

NAMESPACE="knowledge-graph"
VERSION="${1:-1.0.0}"

echo "🚀 Deploying Knowledge Graph v$VERSION..."

# Create namespace
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Apply manifests
echo "📦 Applying backend..."
kubectl apply -f k8s/backend-deployment.yaml

echo "🤖 Applying NLP..."
kubectl apply -f k8s/nlp-deployment.yaml

echo "💾 Applying Redis..."
kubectl apply -f k8s/redis-deployment.yaml

# Wait for rollout
echo "⏳ Waiting for backend rollout..."
kubectl rollout status deployment/kg-backend -n $NAMESPACE --timeout=5m

echo "⏳ Waiting for NLP rollout..."
kubectl rollout status deployment/kg-nlp -n $NAMESPACE --timeout=10m

echo "✅ Deployment complete!"
kubectl get pods -n $NAMESPACE
kubectl get svc -n $NAMESPACE
```

---

## Summary

**K8s Components:**
- ✅ Backend: 3 replicas, 500m/512Mi → 1000m/1Gi
- ✅ NLP: 2 replicas, proper probes (180s startup)
- ✅ Redis: Persistent, 2GB memory, LRU eviction
- ✅ Monitoring: Prometheus rules + alerts
- ✅ Auto-migration: Init containers
- ✅ Pod anti-affinity: Spread across nodes
- ✅ Resource limits: Prevent OOM/CPU throttle
