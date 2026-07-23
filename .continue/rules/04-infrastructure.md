---
name: Infrastructure Rules
alwaysApply: false
globs: ["docker-compose*.yml", "Dockerfile*", "nginx/**", "nginx*.conf", "scripts/**"]
description: Docker multi-stage builds, docker-compose patterns, nginx gateway, health checks, volumes
---

# Infrastructure Rules

## Docker Compose Stack

The project uses two compose files:
- `docker-compose.yml` — Development stack (local development)
- `docker-compose.personal.yml` — Personal deployment (production-like, with nginx)

### Services

| Service    | Image                    | Port  | Purpose                     |
|------------|--------------------------|-------|-----------------------------|
| postgres   | pgvector/pgvector:pg16   | 15432 | Primary DB with vector ext  |
| redis      | redis:7-alpine           | 6379  | Cache, pub/sub, asynq queue |
| mongo      | mongo:7                  | 27017 | Draft storage               |
| nlp        | ./nlp-service (build)    | 5000  | Embedding & keyword service |
| backend    | ./backend (build)        | 8080  | Go API server               |
| frontend   | ./frontend (build)       | 3000  | SvelteKit SSR              |
| nginx      | nginx:alpine             | 80    | Reverse proxy gateway       |

## Health Checks — MANDATORY for all services

Every service MUST have a health check:

```yaml
postgres:
  healthcheck:
    test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-kb_user} -d ${POSTGRES_DB:-knowledge_base}"]
    interval: 10s
    timeout: 5s
    retries: 5

redis:
  healthcheck:
    test: ["CMD", "redis-cli", "ping"]
    interval: 10s
    timeout: 5s
    retries: 5

mongo:
  healthcheck:
    test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
    interval: 10s
    timeout: 5s
    retries: 5

nlp:
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:5000/health"]
    interval: 30s
    timeout: 10s
    retries: 30
    start_period: 600s  # Model loading takes time

backend:
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
    interval: 10s
    timeout: 5s
    retries: 5
```

## Docker Multi-Stage Builds

### Go Backend

```dockerfile
# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server

# Stage 2: Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl
COPY --from=builder /server /server
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=5s CMD curl -f http://localhost:8080/health || exit 1
ENTRYPOINT ["/server"]
```

### Frontend (SvelteKit)

```dockerfile
# Stage 1: Build
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Stage 2: Runtime
FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json ./
COPY --from=builder /app/node_modules ./node_modules
EXPOSE 3000
CMD ["node", "build"]
```

## Volume Naming Conventions

```yaml
volumes:
  postgres_data:      # PostgreSQL persistent data
  mongodb_data:       # MongoDB persistent data
  redis_data:         # Redis persistence (if AOF enabled)
  # Host-mounted volumes:
  ./huggingface_cache:/root/.cache/huggingface   # Pre-cached HF models
  ./init-db:/docker-entrypoint-initdb.d          # DB init scripts
  ./logs:/app/logs                               # Application logs
  ./backups:/app/backups                         # Backup files
```

## Nginx Gateway Configuration

### Development (nginx.conf)

```nginx
upstream backend {
    server backend:8080;
}
upstream frontend {
    server frontend:3000;
}

server {
    listen 80;

    # API routes → backend
    location /api/ {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # WebSocket support
    location /ws {
        proxy_pass http://backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # Everything else → frontend
    location / {
        proxy_pass http://frontend;
        proxy_set_header Host $host;
    }
}
```

### Personal Stack (nginx.personal.conf)

Adds SSL/TLS, rate limiting, and custom domain configuration.

## Depends-On with Health Conditions

```yaml
backend:
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy
    mongo:
      condition: service_healthy

frontend:
  depends_on:
    backend:
      condition: service_healthy
```

## Resource Limits

```yaml
deploy:
  resources:
    limits:
      memory: 512M    # PostgreSQL
    # memory: 256M    # Redis
    # memory: 2G      # NLP (model in memory)
```

## Environment Variables Pattern

```yaml
backend:
  env_file:
    - .env
  environment:
    - DATABASE_URL=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}
    - REDIS_URL=redis://redis:6379
    - MONGO_URL=mongodb://mongo:27017
    - NLP_SERVICE_URL=http://nlp:5000
    - JWT_SECRET=${JWT_SECRET}
```

## Anti-Patterns

```yaml
# ❌ Bad — no health check
backend:
  build: ./backend
  ports:
    - "9000:8080"

# ✅ Good — health check + depends_on conditions
backend:
  build: ./backend
  ports:
    - "9000:8080"
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
    interval: 10s
    timeout: 5s
    retries: 5
  depends_on:
    postgres:
      condition: service_healthy
```

```yaml
# ❌ Bad — secrets in docker-compose.yml
environment:
  - JWT_SECRET=my-super-secret-key

# ✅ Good — reference from .env
env_file:
  - .env
```

```dockerfile
# ❌ Bad — single stage (huge image)
FROM golang:1.25
COPY . .
RUN go build -o /server ./cmd/server
CMD ["/server"]

# ✅ Good — multi-stage (minimal image)
FROM golang:1.25-alpine AS builder
# ... build ...
FROM alpine:3.19
COPY --from=builder /server /server
```
