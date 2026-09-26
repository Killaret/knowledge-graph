# Knowledge Graph Deployment Guide

> **Version:** 1.0  
> **Date:** April 2026  
> **Status:** Production Ready

🌐 **Languages**: [English](DEPLOYMENT_EN.md)

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
| **MongoDB** | 7+ | 0.5 core | 512 MB | 5 GB |
| **Redis** | 7+ | 0.25 core | 256 MB | 1 GB |
| **NLP** | Python 3.11+ | 1 core | 2 GB | 2 GB |

### Recommended (Production)

| Component | CPU | RAM | Disk | Notes |
|-----------|-----|-----|------|-------|
| **Backend** | 2 cores | 2 GB | 1 GB | Scalable |
| **Frontend** | 1 core | 512 MB | 100 MB | Static files |
| **PostgreSQL** | 4 cores | 4 GB | 100 GB | SSD required |
| **MongoDB** | 2 cores | 2 GB | 20 GB | For drafts |
| **Redis** | 1 core | 1 GB | 5 GB | Persistence enabled |
| **NLP** | 4 cores | 8 GB | 5 GB | GPU optional |

### GPU Requirements (for NLP)

NLP service works on CPU, but GPU significantly accelerates embedding generation:

- **Minimum**: CPU-only (slow for large texts)
- **Recommended**: NVIDIA GPU with CUDA 11.8+
- **Models**: `paraphrase-multilingual-MiniLM-L12-v2` (384 dim), configurable via `NLP_MODEL_NAME`, uses ~500MB VRAM

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
DATABASE_URL=postgresql://kb_user:kb_password@postgres:5432/knowledge_base?sslmode=disable

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

Migrations apply automatically at `server` startup (`postgres.RunMigrations`, `./migrations` inside the image) — there is **no `migrate` CLI** in the backend container. Check `docker-compose logs backend` for the migration report.

```bash
# Seed test data — only on the test stack (APP_ENV=test):
# use scripts/testing/seed-test-data.ps1 instead; the image binary is
# ./test-seed and it refuses to run outside APP_ENV=test.
```

### 5. Health Verification

```bash
# Backend /health checks dependencies (Postgres, Redis, Mongo, NLP)
curl http://localhost:9000/health   # direct backend port
curl http://localhost:18080/health  # nginx gateway health (static "OK")
# There is no /db-check endpoint — backend /health already reports each dependency.

# NLP service
curl http://localhost:5000/health
# → {"status":"ok"}

# Frontend
curl http://localhost:5173
# → HTML page
```

### 6. Application Access

- **Frontend**: http://localhost:5173 (dev server) or http://localhost:18081 (nginx)
- **Backend API**: http://localhost:18080
- **API Docs** (if configured): http://localhost:18080/swagger

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
| MongoDB | 27017 | **27018** | kg-mongo-personal |
| Redis | 6379 | **6380** | kg-redis-personal |
| Backend | 9000 | **18085** | kg-backend-personal |
| Frontend | 5173 / 18081 | **3001 / 18084** | kg-frontend-personal |
| NLP | 5000 | **5001** | kg-nlp-personal |

### Access Points

- **Personal Frontend**: http://localhost:3001 or http://localhost:18084
- **Personal API**: http://localhost:18082 (nginx) or http://localhost:18085 (backend direct)

### Data Isolation

Personal instance uses completely separate volumes:
- `pgdata_personal` - PostgreSQL data
- `mongodbdata_personal` - MongoDB data (drafts)
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
      - "18086:8080"
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
      - POSTGRES_USER=${DB_USER:-kb_user}
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_DB=${DB_NAME:-knowledge_base}
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

  mongo:
    image: mongo:7
    environment:
      - MONGO_INITDB_ROOT_USERNAME=${MONGO_USER:-mongo_user}
      - MONGO_INITDB_ROOT_PASSWORD=${MONGO_PASSWORD}
    volumes:
      - mongo_data:/data/db
    ports:
      - "27017:27017"
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G

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
      - NLP_MODEL_NAME=${NLP_MODEL_NAME:-paraphrase-multilingual-MiniLM-L12-v2}
      - HF_HOME=/root/.cache/huggingface
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
  mongo_data:
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

### Option B: Pre-built images from Docker Hub

The repository publishes ready-made images on every green push to `main`. Use
`docker-compose.deploy.yml` for a faster start without building:

```bash
cp .env.example .env
# set JWT_SECRET and other required variables

docker compose -f docker-compose.deploy.yml up -d --wait
```

The compose file uses `KG_IMAGE_TAG` to pick the image version:

| Tag | Meaning |
|-----|---------|
| `main` | Latest green build from `main` (default). Moves on every successful CI run. |
| `YYYY-MM-DD-<short-sha>` | Frozen release from that commit. Pin to this for a stable, auditable deployment. |

To pin, set the environment variable before starting:

```bash
export KG_IMAGE_TAG=2026-09-15-aab2c76
docker compose -f docker-compose.deploy.yml up -d --wait
```

The `main` tag is convenient; a dated tag is a frozen contract. The running
server's `GET /openapi.yaml` must match `backend/openAPI.yaml` from the same
commit — that is the contract-identity check.

---

## Kubernetes (K8s)

> **Target architecture — not implemented.** The repository contains no `k8s/`
> manifests today; the tree and commands below describe the intended layout for a
> future deployment, not files you can `kubectl apply` now. Docker Compose is the
> only supported runtime at this stage.

### Manifest Structure (planned)

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
docker-compose exec postgres pg_dump -U kb_user knowledge_base > backup_$(date +%Y%m%d).sql

# 2. Pull new images
docker-compose pull

# 3. Recreate containers — pending *.up.sql migrations apply
#    automatically when the server starts (RunMigrations)
docker-compose up -d

# 4. Verification
curl http://localhost:9000/health          # backend + dependencies
curl http://localhost:18080/health         # nginx gateway
docker-compose ps                        # all services healthy
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
docker-compose exec postgres psql -U kb_user -d knowledge_base < backup_20250415.sql

# Rollback to previous version
git checkout v1.0.0
docker-compose up --build -d
```

### Migration Rollback

There is no `migrate` CLI inside the backend image — migrations only run forward
at server startup. Rolling back means restoring the database from the backup
taken before the update (step above). The `*.down.sql` files in
`backend/migrations/` exist for a future CLI and can be applied manually with a
local `golang-migrate` install if a table-level revert is unavoidable.

---

## Monitoring

### Health Checks

```bash
# There is no scripts/health-check.sh — check each component:

# Backend: verifies Postgres, Redis, Mongo and NLP in one call
curl http://localhost:9000/health

# Nginx gateway (static OK) and NLP service
curl http://localhost:18080/health
curl http://localhost:5000/health

# Redis
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

> **Not implemented.** There is no `docker-compose.monitoring.yml` and no
> `/metrics` endpoint in the current code — Prometheus/Grafana are a future
> enhancement, not an optional extra you can start today.

---

## Troubleshooting

### Problem: Backend Won't Start

```bash
# Check logs
docker-compose logs backend

# Common causes:
# 1. No database connection — check credentials and postgres health
docker-compose exec postgres pg_isready -U kb_user

# 2. Failed migrations — the server logs "Failed to run migrations" at
#    startup and continues; fix the reported error and restart:
docker-compose restart backend
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
curl http://localhost:18086/health
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
  curl -s http://localhost:18086/health | jq -r '.status'
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
