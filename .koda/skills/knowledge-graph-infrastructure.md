# knowledge-graph-infrastructure

**Version:** 1.0  
**Purpose:** Infrastructure management, deployment, monitoring, scaling  
**Status:** Active  
**Priority:** 🟢 High

---

## Overview

`knowledge-graph-infrastructure` specializes in infrastructure automation, deployment pipelines, monitoring, and operational excellence.

**Key Areas:**
- Docker & Kubernetes
- CI/CD pipelines
- Monitoring (Prometheus, Grafana)
- Logging (ELK, Loki)
- Database management
- Backup & recovery
- Scaling strategies
- Security hardening

---

## Infrastructure Patterns

### 1. Docker Multi-Stage Builds

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

### 2. Docker Compose

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

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      POSTGRES_DB: knowledge_base
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

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

volumes:
  postgres_data:
  redis_data:
```

### 3. Kubernetes Deployment

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

### 4. Horizontal Pod Autoscaler

```yaml
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
```

---

## CI/CD Pipelines

### GitHub Actions

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
        service: [backend, frontend]
    
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
      
      - name: Security scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

  build:
    needs: test
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
```

---

## Monitoring & Observability

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'backend'
    static_configs:
      - targets: ['backend-service:8080']
    metrics_path: /metrics
    
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
    
  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
```

### Alerting Rules

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
          
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API latency detected"
          
      - alert: PodCrashLooping
        expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
        for: 10m
        labels:
          severity: critical
```

---

## Backup & Recovery

### Database Backup

```bash
#!/bin/bash
# backup-database.sh

BACKUP_DIR="/backups/postgres"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=7

docker-compose exec -T postgres pg_dump \
    -U postgres \
    -Fc \
    knowledge_base \
    > "${BACKUP_DIR}/backup_${TIMESTAMP}.dump"

gzip "${BACKUP_DIR}/backup_${TIMESTAMP}.dump"

find "${BACKUP_DIR}" -name "backup_*.dump.gz" -mtime +${RETENTION_DAYS} -delete

echo "Backup completed: backup_${TIMESTAMP}.dump.gz"
```

### Restore Script

```bash
#!/bin/bash
# restore-database.sh

BACKUP_FILE=$1
gunzip "$BACKUP_FILE"

docker-compose exec -T postgres pg_restore \
    -U postgres \
    -d knowledge_base \
    "${BACKUP_FILE%.gz}"

echo "Restore completed"
```

---

## Commands

### Docker Operations
```bash
# Check health
docker compose ps

# Restart service
docker compose restart backend

# View logs
docker compose logs -f backend --tail=100

# Clean up
docker system prune -f
```

### Kubernetes Operations
```bash
# Deploy
kubectl apply -f k8s/deployment.yaml

# Check status
kubectl get pods -l app=backend

# View logs
kubectl logs -f deployment/backend

# Rollback
kubectl rollout undo deployment/backend

# Scale
kubectl scale deployment/backend --replicas=5
```

### Monitoring
```bash
# Prometheus metrics
curl http://localhost:9090/api/v1/query?query=http_requests_total

# Grafana dashboards
open http://localhost:3000
```

---

## Best Practices

### Security
- Use non-root users in containers
- Scan images with Trivy
- Use secrets management (Vault, K8s Secrets)
- Enable network policies
- Regular dependency updates

### Reliability
- Health checks on all services
- Graceful shutdown handling
- Connection pooling
- Circuit breakers
- Retry with exponential backoff

### Performance
- Resource limits on all containers
- Horizontal scaling with HPA
- Database connection pooling
- Redis caching
- CDN for static assets

---

**Tools:** `devops-tools.md`  
**Uptime Target:** > 99.9%  
**Deployment Success Rate:** > 98%