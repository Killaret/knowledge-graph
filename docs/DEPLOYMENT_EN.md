# Knowledge Graph Deployment Guide

> **Version:** 1.0  
> **Date:** April 2026  
> **Status:** Production Ready

🌐 **Languages**: [English](DEPLOYMENT_EN.md) | [Русский](DEPLOYMENT.md)

---

## 📋 Table of Contents

1. [Requirements](#requirements)
2. [Local Deployment](#local-deployment)
3. [Production Deployment](#production-deployment)
4. [Kubernetes (K8s)](#kubernetes-k8s)
5. [Updates](#updates)
6. [Rollback](#rollback)
7. [Monitoring](#monitoring)

---

## Requirements

### Minimum (MVP)

| Component | Version | CPU | RAM | Disk |
|-----------|---------|-----|-----|------|
| **Backend** | Go 1.21+ | 0.5 core | 512 MB | 100 MB |
| **Frontend** | Node 20+ | 0.5 core | 256 MB | 50 MB |
| **PostgreSQL** | 16+ | 1 core | 1 GB | 10 GB |
| **Redis** | 7+ | 0.25 core | 256 MB | 1 GB |
| **NLP** | Python 3.11+ | 1 core | 2 GB | 2 GB |

### Recommended (Production)

| Component | CPU | RAM | Disk | Notes |
|-----------|-----|-----|------|-------|
| **Backend** | 2 cores | 2 GB | 1 GB | Scalable |
| **Frontend** | 1 core | 512 MB | 100 MB | Static files |
| **PostgreSQL** | 4 cores | 4 GB | 100 GB | SSD required |
| **Redis** | 1 core | 1 GB | 5 GB | Persistence enabled |
| **NLP** | 4 cores | 8 GB | 5 GB | GPU optional |

### GPU Requirements (for NLP)

NLP service works on CPU, but GPU significantly accelerates embedding generation:

- **Minimum**: CPU-only (slow for large texts)
- **Recommended**: NVIDIA GPU with CUDA 11.8+
- **Models**: all-MiniLM-L6-v2 (384 dim), uses ~500MB VRAM

---

## Local Deployment

### 1. Environment Preparation

```bash
# Clone repository
git clone https://github.com/your-org/knowledge-graph.git
cd knowledge-graph

# Check Docker and Docker Compose
docker --version  # 24.0+
docker-compose --version  # 2.20+

# Create .env file
cp .env.example .env
```

### 2. Configuration (.env)

```env
# === Required ===
DATABASE_URL=postgresql://kg_user:kg_password@postgres:5432/knowledge_graph?sslmode=disable

# === Optional ===
SERVER_PORT=8080
REDIS_URL=redis:6379
NLP_SERVICE_URL=http://nlp:5000

# === Recommendations (keep defaults for start) ===
RECOMMENDATION_ALPHA=0.5
RECOMMENDATION_BETA=0.5
RECOMMENDATION_DEPTH=3
RECOMMENDATION_DECAY=0.5
RECOMMENDATION_CACHE_TTL_SECONDS=300
EMBEDDING_SIMILARITY_LIMIT=30

# === Visualization ===
GRAPH_LOAD_DEPTH=2
```

### 3. Launch

```bash
# Build and start all services
docker-compose up --build -d

# Check status
docker-compose ps

# Logs
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f worker
docker-compose logs -f nlp
```

### 4. Database Initialization

```bash
# Apply migrations
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" up

# Seed test data (optional)
docker-compose exec backend ./seed
```

### 5. Health Verification

```bash
# Health check backend
curl http://localhost:8080/health
# → {"status":"ok"}

# Health check DB
curl http://localhost:8080/db-check
# → {"status":"db ok"}

# NLP service
curl http://localhost:5000/health
# → {"status":"ok"}

# Frontend
curl http://localhost:5173
# → HTML page
```

### 6. Application Access

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **API Docs** (if configured): http://localhost:8080/swagger

---

## Personal Instance (Parallel Development)

Run a separate personal Knowledge Graph instance alongside the development stack without conflicts.

### Quick Start

**Windows:**
```powershell
.\start-personal.ps1
```

**Linux/Mac:**
```bash
chmod +x start-personal.sh
./start-personal.sh
```

### Manual Launch

```bash
# Build and start personal services
docker compose -f docker-compose.personal.yml up -d --build

# View logs
docker compose -f docker-compose.personal.yml logs -f

# Stop services
docker compose -f docker-compose.personal.yml stop

# Remove completely
docker compose -f docker-compose.personal.yml down
```

### Service Mapping

| Service | Dev Port | Personal Port | Container Name |
|---------|----------|---------------|----------------|
| PostgreSQL | 5432 | **5433** | kg-postgres-personal |
| Redis | 6379 | **6380** | kg-redis-personal |
| Backend | 8080 | **8081** | kg-backend-personal |
| Frontend | 3000 | **3001** | kg-frontend-personal |
| NLP | 5000 | **5001** | kg-nlp-personal |

### Access Points

- **Personal Frontend**: http://localhost:3001
- **Personal API**: http://localhost:8081

### Data Isolation

Personal instance uses completely separate volumes:
- `pgdata_personal` - PostgreSQL data
- `redisdata_personal` - Redis cache

Your personal notes and dev data never overlap.

### For Users (Non-Developers)

If you just want to use Knowledge Graph for your notes without developing:

1. **Only use the personal instance** — ignore the dev stack entirely
2. **Single command to start:**
   ```powershell
   .\start-personal.ps1  # Windows
   ```
   ```bash
   ./start-personal.sh   # Linux/Mac
   ```
3. **Open browser:** http://localhost:3001
4. **Create your first note** — click "+" button in the sidebar

No need to touch `docker-compose.yml` or port 3000 — that's for developers.

### Choosing Between Ports 3000 and 3001

| Scenario | Use Port | Command |
|----------|----------|---------|
| **I want to add features/fix bugs** | 3000 | `docker compose up -d` |
| **I want to use it for my notes** | 3001 | `.\start-personal.ps1` |
| **Testing experimental changes** | 3000 | Dev stack (data may break) |
| **Daily journaling/work notes** | 3001 | Personal stack (stable) |

**Key rule:** Port 3000 is for code changes. Port 3001 is for actual usage.

### Initial Setup After Launch

After starting the personal instance for the first time:

1. **Wait for NLP service** (first launch takes 2-5 minutes to download model):
   ```powershell
   docker compose -f docker-compose.personal.yml logs -f nlp
   # Wait for "Application startup complete" message
   ```

2. **Open the app:** http://localhost:3001

3. **Create your first note:**
   - Click **"+ New Note"** in the left sidebar
   - Write anything — the graph will build automatically

4. **Verify it's working:**
   - Type some text with related concepts
   - Save the note (Ctrl+S or click Save)
   - Check the graph view — nodes should appear

5. **(Optional) Import from Obsidian:**
   - Go to **Settings → Import**
   - Select your Obsidian vault folder
   - Click **Import**

---

## Production Deployment

### Option A: Docker Compose on Server

#### 1. Server Preparation

```bash
# Ubuntu 22.04 LTS
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin -y
```

#### 2. Production Configuration

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  backend:
    build: 
      context: ./backend
      dockerfile: Dockerfile
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - NLP_SERVICE_URL=${NLP_SERVICE_URL}
      - RECOMMENDATION_ALPHA=${RECOMMENDATION_ALPHA:-0.5}
      - RECOMMENDATION_BETA=${RECOMMENDATION_BETA:-0.5}
      - SERVER_PORT=8080
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    # Comprehensive health check returns status of all dependencies:
    # {"status": "healthy", "database": {"status": "healthy"}, "redis": {"status": "healthy"}, "nlp": {"status": "healthy"}}

  worker:
    build:
      context: ./backend
      dockerfile: Dockerfile.worker
    environment:
      - DATABASE_URL=${DATABASE_URL}
      - REDIS_URL=${REDIS_URL}
      - NLP_SERVICE_URL=${NLP_SERVICE_URL}
    depends_on:
      - postgres
      - redis
      - nlp
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 1G

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
      args:
        - PUBLIC_API_URL=/api
    ports:
      - "80:80"
      - "443:443"
    depends_on:
      - backend
    restart: unless-stopped

  postgres:
    image: pgvector/pgvector:pg16
    environment:
      - POSTGRES_USER=${DB_USER:-kg_user}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=${DB_NAME:-knowledge_graph}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init-db:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 4G

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes --maxmemory 1gb --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    restart: unless-stopped

  nlp:
    build:
      context: ./nlp-service
      dockerfile: Dockerfile
    environment:
      - MODEL_NAME=all-MiniLM-L6-v2
      - CACHE_DIR=/app/cache
    volumes:
      - nlp_cache:/app/cache
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  nlp_cache:
```

#### 3. SSL/TLS with Let's Encrypt

```bash
# Install certbot
sudo apt install certbot -y

# Get certificate
sudo certbot certonly --standalone -d your-domain.com

# Configure nginx (in Docker or on host)
```

#### 4. Launch

```bash
# Production launch
docker-compose -f docker-compose.prod.yml up -d

# Scale workers
docker-compose -f docker-compose.prod.yml up -d --scale worker=3
```

---

## Kubernetes (K8s)

### Manifest Structure

```
k8s/
├── namespace.yaml
├── configmap.yaml          # Non-sensitive settings
├── secret.yaml             # Sensitive data (base64)
├── postgres/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── pvc.yaml
├── redis/
│   ├── deployment.yaml
│   └── service.yaml
├── backend/
│   ├── deployment.yaml
│   └── service.yaml
├── worker/
│   └── deployment.yaml
├── frontend/
│   ├── deployment.yaml
│   └── service.yaml
└── ingress.yaml            # SSL ingress
```

### Quick Start with kubectl

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Config and Secrets
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml

# Database
kubectl apply -f k8s/postgres/

# Cache
kubectl apply -f k8s/redis/

# Application
kubectl apply -f k8s/backend/
kubectl apply -f k8s/worker/
kubectl apply -f k8s/frontend/

# Ingress (requires ingress-nginx)
kubectl apply -f k8s/ingress.yaml
```

### Scaling

```bash
# Scale backend
kubectl scale deployment backend --replicas=3

# Scale workers
kubectl scale deployment worker --replicas=5

# HPA (Horizontal Pod Autoscaler)
kubectl autoscale deployment backend --min=2 --max=10 --cpu-percent=70
```

---

## Updates

### Docker Compose Update

```bash
# 1. Backup database
docker-compose exec postgres pg_dump -U kg_user knowledge_graph > backup_$(date +%Y%m%d).sql

# 2. Pull new images
docker-compose pull

# 3. Apply migrations
docker-compose run --rm backend migrate -path /app/migrations -database "$DATABASE_URL" up

# 4. Restart with zero-downtime (if configured)
docker-compose up -d

# 5. Verification
./scripts/health-check.sh
```

### Rolling Update in Kubernetes

```bash
# Update image
kubectl set image deployment/backend backend=kg-backend:v1.1.0

# Track status
kubectl rollout status deployment/backend

# Rollback if issues
kubectl rollout undo deployment/backend
```

---

## Rollback

### Quick Rollback (Docker Compose)

```bash
# Restore from backup
docker-compose exec postgres psql -U kg_user -d knowledge_graph < backup_20250415.sql

# Rollback to previous version
git checkout v1.0.0
docker-compose up --build -d
```

### Migration Rollback

```bash
# Rollback N versions
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" down 3

# Rollback all
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" down
```

---

## Monitoring

### Health Checks

```bash
# Automatic check
./scripts/health-check.sh

# Manual check of all components
curl http://localhost:8080/health
curl http://localhost:8080/db-check
curl http://localhost:5000/health
docker-compose exec redis redis-cli ping
```

### Logs

```bash
# All logs
docker-compose logs --tail=100 -f

# Specific service
docker-compose logs -f backend

# JSON format for analysis
docker-compose logs backend --format json
```

### Metrics (optional)

```bash
# Prometheus + Grafana
docker-compose -f docker-compose.monitoring.yml up -d

# Access Grafana
curl http://localhost:3000  # admin/admin
```

---

## Troubleshooting

### Problem: Backend Won't Start

```bash
# Check logs
docker-compose logs backend

# Common causes:
# 1. No database connection
docker-compose exec backend nc -zv postgres 5432

# 2. Unapplied migrations
docker-compose exec backend migrate -path /app/migrations -database "$DATABASE_URL" up
```

### Problem: NLP Service is Slow

```bash
# Check resources
docker stats kg-nlp

# Model loading takes time on first start
# Check model cache
docker-compose exec nlp ls -la /app/cache/
```

### Problem: Worker Not Processing Tasks

```bash
# Check queue
docker-compose exec redis redis-cli LLEN asynq:{default}

# Restart worker
docker-compose restart worker
```

---

## Production Deployment Checklist

- [ ] Created `.env` file with production values
- [ ] Set strong passwords (DB, Redis)
- [ ] Configured SSL/TLS certificate
- [ ] Database backup configured (cron + pg_dump)
- [ ] Health check monitoring configured
- [ ] Logs sent to centralized storage
- [ ] Backend/worker scaling configured
- [ ] Resource limits/requests configured
- [ ] Readiness/Liveness probes configured (K8s)
- [ ] PDB (Pod Disruption Budget) configured (K8s)

---

## Health Check Monitoring

### Comprehensive Health Endpoint

The `/health` endpoint provides detailed status of all service dependencies:

```bash
curl http://localhost:8080/health
```

**Healthy Response (200):**
```json
{
  "status": "healthy",
  "timestamp": "2026-04-25T18:30:00Z",
  "version": "1.0.0",
  "database": {"status": "healthy"},
  "redis": {"status": "healthy"},
  "nlp": {"status": "healthy"}
}
```

**Unhealthy Response (503):**
```json
{
  "status": "unhealthy",
  "database": {"status": "unhealthy", "error": "connection refused"},
  "redis": {"status": "healthy"},
  "nlp": {"status": "unhealthy", "error": "timeout"}
}
```

### Individual Health Checks

| Endpoint | Description | Status Codes |
|----------|-------------|--------------|
| `GET /health` | Comprehensive health check | 200 (healthy), 503 (unhealthy) |
| `GET /db-check` | Database only | 200 (ok), 500 (error) |
| `GET /health` (NLP) | NLP service health | 200 (healthy), 503 (unhealthy) |

### Monitoring Integration

**Prometheus-style monitoring:**
```bash
# Check every 30s
while true; do
  curl -s http://localhost:8080/health | jq -r '.status'
  sleep 30
done
```

**Alerting rules:**
- 2 consecutive 503 responses → Alert
- Database unhealthy → Critical alert
- Redis unhealthy → Warning (cache degraded)
- NLP unhealthy → Warning (embeddings delayed)

---

## Useful Links

- [Docker Compose Docs](https://docs.docker.com/compose/)
- [Kubernetes Docs](https://kubernetes.io/docs/)
- [pgvector](https://github.com/pgvector/pgvector)
- [Asynq](https://github.com/hibiken/asynq)
