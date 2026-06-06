# knowledge-graph-devops

**Version:** 1.0  
**Purpose:** Infrastructure and deployment automation  
**Status:** Active  
**Priority:** 🟡 High

---

## Overview

`knowledge-graph-devops` specializes in infrastructure automation, deployment, monitoring, and operational excellence.

**Key Areas:**
- Docker Compose optimization
- Kubernetes deployment
- CI/CD pipeline configuration
- Monitoring setup (Prometheus, Grafana)
- Logging configuration (ELK, Loki)
- Backup automation
- Environment management (dev, staging, prod)
- Scaling strategies
- Infrastructure as Code

---

## Infrastructure Patterns

### 1. Docker Optimization

#### Multi-Stage Builds

**Target:** Smaller, secure Docker images

**Techniques:**
```dockerfile
# Backend - Multi-stage build
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/server /server
COPY --from=builder /app/config /config

USER nobody
EXPOSE 8080

CMD ["/server"]
```

```dockerfile
# Frontend - Multi-stage build
FROM node:20-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Nginx runtime
FROM nginx:alpine

COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

#### Image Size Optimization

```dockerfile
# ❌ BAD: Large image
FROM ubuntu:latest
RUN apt-get update && apt-get install -y golang-go
COPY . .
RUN go build -o server

# ✅ GOOD: Optimized image
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -ldflags="-w -s" -o server

FROM scratch
COPY --from=builder /app/server /server
EXPOSE 8080
CMD ["/server"]
# Result: 600MB → 15MB
```

#### Docker Compose Optimization

```yaml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - DATABASE_URL=postgres://postgres:password@postgres:5432/kb
      - REDIS_URL=redis://redis:6379
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 128M
    restart: unless-stopped
    networks:
      - kb-network

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: knowledge_base
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    deploy:
      resources:
        limits:
          memory: 1G
    networks:
      - kb-network

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    deploy:
      resources:
        limits:
          memory: 256M
    networks:
      - kb-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - backend
      - frontend
    networks:
      - kb-network

volumes:
  postgres_data:
  redis_data:

networks:
  kb-network:
    driver: bridge
```

---

### 2. Kubernetes Deployment

#### Deployment Manifests

```yaml
# backend-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  labels:
    app: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: your-registry/knowledge-graph-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        - name: REDIS_URL
          value: "redis://redis-service:6379"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        resources:
          requests:
            memory: "128Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  selector:
    app: backend
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

```yaml
# postgres-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:16-alpine
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        - name: POSTGRES_DB
          value: knowledge_base
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "2000m"
  volumeClaimTemplates:
  - metadata:
      name: postgres-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
spec:
  clusterIP: None
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

#### Horizontal Pod Autoscaler

```yaml
# backend-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
```

---

### 3. CI/CD Pipelines

#### GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [backend, frontend, nlp-service]
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go (backend)
        if: matrix.service == 'backend'
        uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      
      - name: Setup Node (frontend)
        if: matrix.service == 'frontend'
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: Setup Python (nlp-service)
        if: matrix.service == 'nlp-service'
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Backend tests
        if: matrix.service == 'backend'
        run: |
          cd backend
          go test -race -cover ./...
      
      - name: Frontend tests
        if: matrix.service == 'frontend'
        run: |
          cd frontend
          npm ci
          npm run test:unit
          npm run lint
      
      - name: NLP tests
        if: matrix.service == 'nlp-service'
        run: |
          cd nlp-service
          pip install -r requirements.txt
          pytest tests/ -v

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'
      
      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

  build:
    needs: [test, security]
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKER_USERNAME }}
          password: ${{ secrets.DOCKER_PASSWORD }}
      
      - name: Build and push backend
        uses: docker/build-push-action@v4
        with:
          context: ./backend
          push: true
          tags: your-registry/knowledge-graph-backend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
      
      - name: Build and push frontend
        uses: docker/build-push-action@v4
        with:
          context: ./frontend
          push: true
          tags: your-registry/knowledge-graph-frontend:${{ github.sha }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deploy environment'
        required: true
        default: 'staging'
        type: choice
        options:
        - staging
        - production

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: ${{ github.event.inputs.environment }}
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup kubectl
        uses: azure/setup-kubectl@v3
        with:
          version: 'latest'
      
      - name: Configure kubectl
        run: |
          echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > kubeconfig
          export KUBECONFIG=kubeconfig
      
      - name: Deploy to Kubernetes
        run: |
          export KUBECONFIG=kubeconfig
          
          # Update image tags
          kubectl set image deployment/backend \
            backend=your-registry/knowledge-graph-backend:${{ github.sha }}
          
          kubectl set image deployment/frontend \
            frontend=your-registry/knowledge-graph-frontend:${{ github.sha }}
          
          # Rollout restart
          kubectl rollout restart deployment/backend
          kubectl rollout restart deployment/frontend
          
          # Wait for rollout
          kubectl rollout status deployment/backend --timeout=300s
          kubectl rollout status deployment/frontend --timeout=300s
      
      - name: Run database migrations
        run: |
          export KUBECONFIG=kubeconfig
          kubectl run migration-job --rm -it --image=your-registry/knowledge-graph-backend:${{ github.sha }} \
            -- /server migrate
      
      - name: Smoke tests
        run: |
          # Wait for deployment to stabilize
          sleep 30
          
          # Run smoke tests
          curl -f https://${{ github.event.inputs.environment }}.yourdomain.com/health || exit 1
```

---

### 4. Monitoring & Observability

#### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'backend'
    static_configs:
      - targets: ['backend-service:8080']
    metrics_path: /metrics
    scrape_interval: 10s
    
  - job_name: 'frontend'
    static_configs:
      - targets: ['frontend-service:3000']
    metrics_path: /metrics
    
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
    
  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

#### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Knowledge Graph Overview",
    "panels": [
      {
        "title": "API Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "API Latency (p95)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "{{endpoint}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m]) * 100",
            "legendFormat": "Error %"
          }
        ]
      },
      {
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "postgres_connections_active",
            "legendFormat": "Active"
          },
          {
            "expr": "postgres_connections_idle",
            "legendFormat": "Idle"
          }
        ]
      }
    ]
  }
}
```

#### Alerting Rules

```yaml
# alerting-rules.yml
groups:
  - name: knowledge-graph-alerts
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | printf \"%.2f\" }}%"
          
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API latency detected"
          description: "95th percentile latency is {{ $value }}s"
          
      - alert: PodCrashLooping
        expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Pod {{ $labels.pod }} is crash looping"
          
      - alert: HighMemoryUsage
        expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage on {{ $labels.pod }}"
```

---

### 5. Logging Configuration

#### Centralized Logging with Loki

```yaml
# loki-config.yaml
auth_enabled: false

server:
  http_listen_port: 3100

ingester:
  wal:
    enabled: true
    dir: /loki/wal
  lifecycler:
    ring:
      kvstore:
        store: inmemory
      replication_factor: 1

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

storage_config:
  boltdb_shipper:
    active_index_directory: /loki/index
    cache_location: /loki/cache
  filesystem:
    directory: /loki/chunks

compactor:
  working_directory: /loki/compactor
```

#### Application Logging

```go
// Structured logging
type Logger struct {
    logger *zap.SugaredLogger
}

func NewLogger(env string) *Logger {
    config := zap.NewProductionConfig()
    if env == "development" {
        config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
    }
    
    zapLogger, _ := config.Build()
    return &Logger{logger: zapLogger.Sugar()}
}

func (l *Logger) Info(msg string, fields ...interface{}) {
    l.logger.Infow(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
    l.logger.Errorw(msg, fields...)
}

// Usage in handlers
logger.Info("Request processed",
    "method", c.Request.Method,
    "path", c.Request.URL.Path,
    "duration", duration,
    "user_id", userID)
```

---

### 6. Backup Automation

#### Database Backup Script

```bash
#!/bin/bash
# backup-database.sh

BACKUP_DIR="/backups/postgres"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=7

# Create backup
docker-compose exec -T postgres pg_dump \
    -U postgres \
    -Fc \
    knowledge_base \
    > "${BACKUP_DIR}/backup_${TIMESTAMP}.dump"

# Compress
gzip "${BACKUP_DIR}/backup_${TIMESTAMP}.dump"

# Cleanup old backups
find "${BACKUP_DIR}" -name "backup_*.dump.gz" -mtime +${RETENTION_DAYS} -delete

# Upload to S3 (optional)
aws s3 cp "${BACKUP_DIR}/backup_${TIMESTAMP}.dump.gz" \
    s3://your-bucket/backups/postgres/

echo "Backup completed: backup_${TIMESTAMP}.dump.gz"
```

#### Restore Script

```bash
#!/bin/bash
# restore-database.sh

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: ./restore-database.sh <backup-file>"
    exit 1
fi

# Decompress
gunzip "$BACKUP_FILE"

# Restore
docker-compose exec -T postgres pg_restore \
    -U postgres \
    -d knowledge_base \
    "${BACKUP_FILE%.gz}"

echo "Restore completed"
```

#### Cron Job for Daily Backups

```bash
# Add to crontab
0 2 * * * /scripts/backup-database.sh >> /var/log/backup.log 2>&1
# Runs daily at 2 AM
```

---

## Environment Management

### Development Environment

```yaml
# docker-compose.dev.yml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      target: development
    volumes:
      - ./backend:/app
    environment:
      - ENV=development
      - DEBUG=true
    ports:
      - "8080:8080"
      - "40000:40000" # Delve debugger

  frontend:
    build:
      context: ./frontend
      target: development
    volumes:
      - ./frontend:/app
      - /app/node_modules
    ports:
      - "3000:3000"

  postgres:
    image: postgres:16-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: devpassword

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### Production Environment

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  backend:
    image: your-registry/knowledge-graph-backend:latest
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1'
          memory: 512M
    environment:
      - ENV=production
    networks:
      - kb-network

  frontend:
    image: your-registry/knowledge-graph-frontend:latest
    deploy:
      replicas: 2
    networks:
      - kb-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.prod.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - backend
      - frontend
```

---

## Scaling Strategies

### Horizontal Scaling

```yaml
# Kubernetes HPA
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Database Scaling

```sql
-- Read replicas
-- Primary for writes
-- Replicas for reads

-- Application-level routing
func (db *DB) GetReaderDB() *sql.DB {
    return db.readReplicas[rand.Intn(len(db.readReplicas))]
}

func (db *DB) GetWriterDB() *sql.DB {
    return db.primary
}
```

---

## DevOps Checklist

### Deployment

- [ ] Docker images optimized (< 100MB for Go, < 200MB for Node)
- [ ] Multi-stage builds implemented
- [ ] Health checks configured
- [ ] Readiness probes configured
- [ ] Resource limits set
- [ ] Secrets managed securely
- [ ] Rolling update strategy configured
- [ ] Rollback procedure tested

### Monitoring

- [ ] Prometheus metrics exposed
- [ ] Grafana dashboards created
- [ ] Alerts configured
- [ ] Log aggregation setup
- [ ] Distributed tracing enabled
- [ ] Uptime monitoring configured

### CI/CD

- [ ] Automated tests in pipeline
- [ ] Security scanning enabled
- [ ] Automated deployments
- [ ] Environment promotion (dev → staging → prod)
- [ ] Manual approval for production
- [ ] Rollback automation

### Backup & Recovery

- [ ] Daily automated backups
- [ ] Backup encryption enabled
- [ ] Off-site backup storage
- [ ] Restore procedure tested
- [ ] RPO/RTO defined

---

## References

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Docker Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- [GitHub Actions](https://docs.github.com/en/actions)

---

**Last Updated:** 2026-05-22  
**Maintainer:** knowledge-graph-docs-maintenance  
**Version:** 1.0
