# Docker Deployment Guide

## Overview

Knowledge Graph uses Docker Compose for containerization with microservices architecture.

## Services

### Core Services

| Service | Container | Port | Description |
|---------|-----------|------|-------------|
| Frontend | kg-frontend | 5173 | SvelteKit production build (adapter-node) |
| Backend | kg-backend | 9000 (127.0.0.1) | Go API server (Gin + GORM) |
| Graph Service | kg-graph-service | 9091 | gRPC layout service (Go 1.24) |
| Nginx | kg-nginx | 8080, 8081 | API gateway & reverse proxy |
| Worker | kg-worker | - | Background worker for async tasks |

### Data Services

| Service | Container | Port | Description |
|---------|-----------|------|-------------|
| PostgreSQL | kg-postgres | 15432 | pgvector for semantic search |
| Redis | kg-redis | 6379 | Cache & job queue |
| MongoDB | kg-mongo | 27017 | Drafts storage |
| NLP | kg-nlp | 5000 | Embeddings service (Python/FastAPI) |

## Quick Start

### Dev Stack

```bash
# Start all services
docker-compose up -d postgres redis backend graph-service frontend nginx worker nlp

# Start essential services only
docker-compose up -d postgres redis backend graph-service frontend nginx

# Check service health
docker-compose ps
docker logs kg-backend
docker logs kg-nginx
```

### Personal Stack

```bash
# Start personal instance (different ports)
docker-compose -f docker-compose.personal.yml up -d

# Personal ports:
# Backend: 8085
# API Gateway: 8082
# Graph Service: 9092
```

## Architecture

### Proxy Configuration

**Nginx API Gateway (port 8080):**

```
Frontend (5173) → Nginx (8080) → Backend (8080)
                          → Graph Service (9091)
```

**Nginx Configuration:**

```nginx
location /api/ {
    proxy_pass http://backend:8080;
}

location /graph-service/ {
    rewrite ^/graph-service(/.*)$ $1 break;
    proxy_pass http://graph-service:9091;
}
```

**Frontend Configuration:**

- Dev mode: Vite proxy in `vite.config.ts`
- Production: Nginx proxy (adapter-node + hooks.server.ts)

### Service Dependencies

```
postgres, redis (healthy) ──► backend (healthy)
postgres, redis (healthy) ──► graph-service (healthy)
backend (healthy) ──► nginx (healthy)
graph-service (healthy) ──► nginx (healthy)
backend, graph-service (healthy) ──► frontend (healthy)
```

## Environment Variables

### Backend (.env)

```bash
DATABASE_URL=postgresql://kb_user:kb_password@postgres:5432/knowledge_base?sslmode=disable
REDIS_URL=redis:6379
NLP_SERVICE_URL=http://nlp:5000
GRAPH_SERVICE_URL=http://graph-service:9091
SKIP_AUTH=false
```

### Graph Service (.env)

```bash
POSTGRES_URL=postgresql://kb_user:kb_password@postgres:5432/knowledge_base?sslmode=disable
REDIS_URL=redis:6379
EVENT_CHANNEL=graph:events
GRAPH_FULL_LIMIT=1000
```

### Frontend (.env)

```bash
VITE_API_URL=http://localhost:8080
VITE_GRAPH_SERVICE_URL=http://localhost:8080/graph-service
ARGOS_TOKEN=argos_94zzm1fanz4uk559g2tmsqok8x6ls4p6q8
```

## Health Checks

### Individual Services

```bash
# Backend
curl http://localhost:9000/health

# Graph Service
curl http://localhost:9091/health

# Nginx Gateway
curl http://localhost:8080/health
```

### Through Nginx Proxy

```bash
# Backend API
curl http://localhost:8080/api/v1/notes

# Graph Service API
curl http://localhost:8080/graph-service/api/v1/graph/full

# Full graph data
curl http://localhost:8080/api/v1/graph/all
```

## Troubleshooting

### Container Won't Start

```bash
# Check container logs
docker logs kg-backend
docker logs kg-graph-service
docker logs kg-nginx

# Restart specific service
docker-compose restart backend
docker-compose restart graph-service
```

### Port Conflicts

```bash
# Check what's using the port
netstat -ano | findstr :8080
netstat -ano | findstr :5173

# Kill process on port (Windows)
taskkill /PID <pid> /F
```

### Nginx Proxy Issues

```bash
# Check nginx logs
docker logs kg-nginx --tail 50

# Test nginx config
docker exec kg-nginx nginx -t

# Reload nginx
docker-compose restart nginx
```

### Database Connection Issues

```bash
# Check postgres health
docker exec kg-postgres pg_isready -U kb_user

# Connect to postgres
docker exec -it kg-postgres psql -U kb_user -d knowledge_base

# Check redis health
docker exec kg-redis redis-cli ping
```

## Stopping Services

```bash
# Stop all services
docker-compose down

# Stop specific service
docker-compose stop backend
docker-compose stop frontend

# Remove volumes (⚠️ deletes data)
docker-compose down -v
```

## Personal Stack

The personal stack uses different ports to avoid conflicts:

| Service | Dev Stack | Personal Stack |
|---------|-----------|---------------|
| Backend | 9000 | 8085 |
| Nginx API | 8080 | 8082 |
| Graph Service | 9091 | 9092 |
| Frontend | 5173 | (uses nginx:8081) |
| PostgreSQL | 15432 | 5433 |
| Redis | 6379 | 6380 |
| MongoDB | 27017 | 27018 |

## Monitoring

### Container Resource Usage

```bash
# Check resource usage
docker stats

# Check specific container
docker stats kg-backend
docker stats kg-graph-service
```

### Logs

```bash
# Follow logs
docker-compose logs -f backend
docker-compose logs -f graph-service

# Get last 100 lines
docker-compose logs --tail=100 backend
```

## Building Images

```bash
# Build all images
docker-compose build

# Build specific service
docker-compose build backend
docker-compose build frontend
docker-compose build graph-service

# Build without cache
docker-compose build --no-cache
```

## Production Deployment

### Security Considerations

- Use strong passwords in production
- Change default ports in production
- Use HTTPS/TLS in production
- Configure proper CORS origins via environment variables
- Enable rate limiting in production

### CORS Configuration

**Environment Variables:**
- `CORS_ALLOWED_ORIGINS` - Comma-separated list of allowed origins (e.g., `https://example.com,https://app.example.com`)
- `CORS_ALLOWED_METHODS` - Allowed HTTP methods (default: `GET,POST,PUT,DELETE,OPTIONS`)
- `CORS_ALLOWED_HEADERS` - Allowed headers (default: `Content-Type,Authorization`)
- `CORS_MAX_AGE` - Preflight cache duration in seconds (default: `86400`)

**Example docker-compose.yml:**
```yaml
backend:
  environment:
    - CORS_ALLOWED_ORIGINS=https://example.com,https://app.example.com
    - CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
    - CORS_ALLOWED_HEADERS=Content-Type,Authorization
    - CORS_MAX_AGE=86400
```

**Important:** Never use `*` as CORS origin in production. Always specify exact origins.

### Scaling

```bash
# Scale backend (workers)
docker-compose up -d --scale backend=3

# Scale graph service
docker-compose up -d --scale graph-service=2
```