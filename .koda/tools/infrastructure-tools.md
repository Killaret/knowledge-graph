# Инструменты Infrastructure Агента

**Версия:** 1.0  
**Назначение:** Инфраструктура, Docker, Kubernetes, мониторинг, бэкапы

---

## 🎯 Основные задачи

1. Управление Docker контейнерами
2. Kubernetes деплоймент
3. CI/CD пайплайны
4. Мониторинг (Prometheus, Grafana)
5. Логирование (ELK, Loki)
6. Бэкапы и восстановление
7. Масштабирование
8. Security hardening

---

## 🐳 Docker

### Multi-Stage Builds

#### Backend
```dockerfile
# backend/Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o server \
    ./cmd/server

# Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/server /server
COPY --from=builder /app/config /config
USER nobody
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --spider -q http://localhost:8080/health
CMD ["/server"]
```

#### Frontend
```dockerfile
# frontend/Dockerfile
# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production
FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Docker Compose

```yaml
# docker-compose.yml
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
      - DATABASE_URL=postgres://postgres:${DB_PASSWORD}@postgres:5432/kb
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./logs:/app/logs
    networks:
      - kb-network
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 128M
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:80"
    depends_on:
      - backend
    networks:
      - kb-network
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: knowledge_base
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    networks:
      - kb-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    networks:
      - kb-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    ports:
      - "9090:9090"
    networks:
      - kb-network
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    ports:
      - "3001:3000"
    networks:
      - kb-network
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:

networks:
  kb-network:
    driver: bridge
```

---

## ☸️ Kubernetes

### Deployment

```yaml
# k8s/backend-deployment.yaml
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
        imagePullPolicy: Always
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
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
      restartPolicy: Always
```

### Service

```yaml
# k8s/backend-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: backend-service
  labels:
    app: backend
spec:
  selector:
    app: backend
  ports:
  - port: 8080
    targetPort: 8080
    protocol: TCP
  type: ClusterIP
```

### Horizontal Pod Autoscaler

```yaml
# k8s/backend-hpa.yaml
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
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 0
      policies:
      - type: Percent
        value: 100
        periodSeconds: 15
      - type: Pods
        value: 4
        periodSeconds: 15
      selectPolicy: Max
```

### Ingress

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kb-ingress
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
spec:
  tls:
  - hosts:
    - api.knowledge-graph.example.com
    - app.knowledge-graph.example.com
    secretName: kb-tls-secret
  rules:
  - host: api.knowledge-graph.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: backend-service
            port:
              number: 8080
  - host: app.knowledge-graph.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-service
            port:
              number: 80
```

---

## 📊 Мониторинг

### Prometheus Configuration

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

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
    
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
    - role: pod
    relabel_configs:
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
      action: keep
      regex: true
    - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
      action: replace
      target_label: __metrics_path__
      regex: (.+)
```

### Alerting Rules

```yaml
# monitoring/alerting-rules.yml
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
          description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes"
          
      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API latency detected"
          description: "p95 latency is {{ $value }}s"
          
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
          summary: "High memory usage on {{ $labels.container }}"
          
      - alert: DatabaseConnectionPoolExhausted
        expr: pg_stat_database_numbackends / pg_settings_max_connections > 0.9
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Database connection pool nearly exhausted"
```

---

## 💾 Бэкапы

### Database Backup Script

```bash
#!/bin/bash
# scripts/backup-database.sh

set -e

BACKUP_DIR="/backups/postgres"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30
DB_NAME="knowledge_base"

# Create backup directory if not exists
mkdir -p "${BACKUP_DIR}"

# Create backup
echo "Creating backup: backup_${TIMESTAMP}.dump"
docker-compose exec -T postgres pg_dump \
    -U postgres \
    -Fc \
    ${DB_NAME} \
    | gzip > "${BACKUP_DIR}/backup_${TIMESTAMP}.dump.gz"

# Verify backup
if [ -s "${BACKUP_DIR}/backup_${TIMESTAMP}.dump.gz" ]; then
    echo "Backup completed successfully"
else
    echo "Backup failed!"
    exit 1
fi

# Calculate backup size
BACKUP_SIZE=$(du -h "${BACKUP_DIR}/backup_${TIMESTAMP}.dump.gz" | cut -f1)
echo "Backup size: ${BACKUP_SIZE}"

# Clean old backups
echo "Cleaning backups older than ${RETENTION_DAYS} days"
find "${BACKUP_DIR}" -name "backup_*.dump.gz" -mtime +${RETENTION_DAYS} -delete

# Upload to S3 (if configured)
if [ ! -z "${AWS_S3_BUCKET}" ]; then
    aws s3 cp "${BACKUP_DIR}/backup_${TIMESTAMP}.dump.gz" \
        s3://${AWS_S3_BUCKET}/backups/
fi

echo "Backup completed: backup_${TIMESTAMP}.dump.gz (${BACKUP_SIZE})"
```

### Restore Script

```bash
#!/bin/bash
# scripts/restore-database.sh

set -e

BACKUP_FILE=$1

if [ -z "${BACKUP_FILE}" ]; then
    echo "Usage: $0 <backup-file>"
    exit 1
fi

if [ ! -f "${BACKUP_FILE}" ]; then
    echo "Backup file not found: ${BACKUP_FILE}"
    exit 1
fi

echo "Starting restore from: ${BACKUP_FILE}"

# Decompress if needed
if [[ ${BACKUP_FILE} == *.gz ]]; then
    gunzip ${BACKUP_FILE}
    BACKUP_FILE=${BACKUP_FILE%.gz}
fi

# Restore
docker-compose exec -T postgres pg_restore \
    -U postgres \
    -d knowledge_base \
    --clean \
    --if-exists \
    "${BACKUP_FILE}"

echo "Restore completed successfully"
```

---

## 🔄 CI/CD

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
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
          cache-from: type=registry,ref=your-registry/knowledge-graph-backend:buildcache
          cache-to: type=registry,ref=your-registry/knowledge-graph-backend:buildcache,mode=max
      
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/backend \
            backend=your-registry/knowledge-graph-backend:${{ github.sha }}
          kubectl rollout status deployment/backend --timeout=300s
      
      - name: Verify deployment
        run: |
          kubectl get pods -l app=backend
          kubectl rollout history deployment/backend
```

---

## 🛠️ Команды

### Docker Operations
```bash
# Проверить статус
docker compose ps

# Перезапустить сервис
docker compose restart backend

# Просмотр логов
docker compose logs -f backend --tail=100

# Остановить все
docker compose down

# Очистка
docker system prune -f
```

### Kubernetes Operations
```bash
# Применить конфиги
kubectl apply -f k8s/

# Проверить статус
kubectl get pods
kubectl get services

# Просмотр логов
kubectl logs -f deployment/backend

# Роллаут
kubectl rollout undo deployment/backend
kubectl rollout status deployment/backend

# Масштабирование
kubectl scale deployment/backend --replicas=5

# Port forwarding
kubectl port-forward service/backend-service 8080:8080
```

---

## 📊 Best Practices

### Security
- Использовать non-root пользователи в контейнерах
- Сканировать образы (Trivy)
- Хранить секреты в Kubernetes Secrets / Vault
- Включать Network Policies
- Регулярно обновлять зависимости

### Reliability
- Health checks на всех сервисах
- Graceful shutdown
- Connection pooling
- Circuit breakers
- Retry с exponential backoff

### Performance
- Resource limits на всех контейнерах
- Horizontal scaling с HPA
- Database connection pooling
- Redis кэширование
- CDN для статических ассетов

---

**Tools:** Этот файл + `devops-tools.md`  
**Uptime Target:** > 99.9%  
**Deployment Success Rate:** > 98%