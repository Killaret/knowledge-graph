# Cursor Rule: knowledge-graph-infrastructure

Docker, Compose, nginx, volumes, health checks. Reference files:
`docker-compose.yml`, `docker-compose.personal.yml`, `nginx.conf`,
`backend/Dockerfile`, `frontend/Dockerfile`.

---

## Docker Multi-Stage Build Pattern

### Backend (Go) — `backend/Dockerfile`
```dockerfile
FROM golang:1.25-alpine AS builder
COPY . /app
WORKDIR /app/backend
RUN go mod download
RUN go build -ldflags="-s -w" -o /app/server ./cmd/server
RUN go build -ldflags="-s -w" -o /app/worker ./cmd/worker
RUN go build -ldflags="-s -w" -o /app/cli    ./cmd/cli

FROM alpine:latest AS base
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/

FROM base AS server         # target: server
COPY --from=builder /app/server .
COPY --from=builder /app/backend/migrations ./migrations
COPY --from=builder /app/knowledge-graph.config.json .
EXPOSE 8080
CMD ["./server"]

FROM base AS worker         # target: worker
COPY --from=builder /app/worker .
CMD ["./worker"]

FROM base AS cli            # target: cli
COPY --from=builder /app/cli .
CMD ["./cli"]
```

Select a target in docker-compose:
```yaml
build:
  context: .
  dockerfile: ./backend/Dockerfile
  target: server
```

### Frontend (SvelteKit) — `frontend/Dockerfile`
```dockerfile
FROM node:20-alpine AS builder
ARG VITE_API_URL=http://localhost:8080   # baked into client bundle at build time
ENV VITE_API_URL=${VITE_API_URL}
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/. .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json .
RUN apk add --no-cache curl && npm ci --omit=dev
EXPOSE 3000
CMD ["node", "build"]
```

**Important:** `VITE_*` variables baked at build time cannot be changed at
runtime. Pass them as `ARG` / `ENV` during `docker compose build`.

---

## Health Check Requirements

Every service **must** expose `/health` and declare a `healthcheck` in compose.

| Service | Health check URL | start_period |
|---------|-----------------|-------------|
| `postgres` | `pg_isready` | default |
| `redis` | `redis-cli ping` | default |
| `mongo` | `mongosh db.adminCommand('ping')` | default |
| `nlp` | `curl http://localhost:5000/health` | **600s** (model loads ~15s from cache) |
| `backend` | `curl http://localhost:8080/health` | 10s |
| `graph-service` | `curl http://localhost:9091/health` | 10s |
| `frontend` | `curl http://localhost:3000/` | 10s |
| `nginx` | `curl http://127.0.0.1:8080/health` | 15s |

The NLP service has `retries: 30` and `start_period: 600s` because the model
loads lazily on the first `/health` call — not on container start.

---

## Volume Naming Conventions

```yaml
volumes:
  postgres_data:     # dev stack (docker-compose.yml)
  mongodb_data:      # dev stack
  # redis has no named volume in dev (ephemeral is acceptable)

  pgdata_personal:   # personal stack (docker-compose.personal.yml)
  redisdata_personal:
  mongodbdata_personal:
```

HuggingFace model cache is a **bind mount** (not a named volume):
```yaml
volumes:
  - ./huggingface_cache:/root/.cache/huggingface
```
This allows the cache to be pre-populated on the host without rebuilding images.

---

## Dev Stack vs Personal Stack

| Property | `docker-compose.yml` | `docker-compose.personal.yml` |
|----------|---------------------|-------------------------------|
| Purpose | Development / CI | Persistent personal instance |
| Postgres port | 15432 | 5433 |
| Redis port | 6379 | 6380 |
| Mongo port | 27017 | 27018 |
| NLP port | 5000 | 5001 |
| Backend port | 9000 (127.0.0.1) | 8085 |
| Nginx API port | 8080 | 8082 |
| Nginx frontend port | 8081 | 8083 |
| Backup scheduler | no | **yes** (`backup_scheduler` service) |
| Nginx config | `nginx.conf` | `nginx.personal.conf` |
| DB name | `knowledge_base` | `knowledge_personal` |

Run personal stack: `docker compose -f docker-compose.personal.yml up -d`

---

## Docker Compose Service Dependencies

```yaml
backend:
  depends_on:
    postgres:
      condition: service_healthy   # waits for pg_isready
    redis:
      condition: service_healthy
    nlp:
      condition: service_started   # nlp health is slow; don't block backend

graph-service:
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy

frontend:
  depends_on:
    backend:
      condition: service_healthy
    graph-service:
      condition: service_healthy
    nginx:
      condition: service_healthy
```

---

## Nginx Configuration Patterns

```nginx
# nginx.conf — port 8080 (API + graph-service gateway)
events { worker_connections 1024; }
http {
    resolver 127.0.0.11 valid=30s;  # Docker embedded DNS — REQUIRED

    server {
        listen 8080;

        location /api/ {
            set $backend_host backend:8080;  # variable = runtime DNS resolution
            proxy_pass http://$backend_host;
            proxy_set_header Host             $host;
            proxy_set_header X-Real-IP        $remote_addr;
            proxy_set_header X-Forwarded-For  $proxy_add_x_forwarded_for;
        }

        location /graph-service/ {
            set $graph_service_host graph-service:9091;
            rewrite ^/graph-service(/.*)$ $1 break;  # strip prefix
            proxy_pass http://$graph_service_host;
        }

        location /health {
            return 200 "OK\n";
            add_header Content-Type text/plain;
        }
    }

    server {
        listen 8081;  # frontend proxy
        location / {
            set $frontend_host frontend:3000;
            proxy_pass http://$frontend_host/;
        }
    }
}
```

**Do not use static `upstream` blocks** — they resolve DNS at config load time
and will fail if the container isn't yet running.

---

## Environment Variable Management

- Secrets live in `.env` (not committed — see `.gitignore`).
- `env_file: - .env` loads it into every service.
- Per-service overrides go in `environment:` block.
- Never hard-code secrets in `docker-compose.yml`.

```yaml
# .env (gitignored)
POSTGRES_USER=kb_user
POSTGRES_PASSWORD=very_secret
JWT_SECRET=another_secret
BACKUP_YANDEX_TOKEN=aqXXXXXXXX
```

```yaml
# docker-compose.yml — read from env with fallback
environment:
  DATABASE_URL: ${DATABASE_URL:-postgresql://${POSTGRES_USER:-kb_user}:${POSTGRES_PASSWORD:-kb_password}@postgres:5432/${POSTGRES_DB:-knowledge_base}?sslmode=disable}
```

---

## Anti-Patterns

```dockerfile
# ❌ Single-stage build — ships Go toolchain in production image
FROM golang:1.25
COPY . .
RUN go build -o server .
CMD ["./server"]

# ❌ No health check — depends_on: condition: service_healthy won't work
services:
  myservice:
    image: myimage
    # missing healthcheck:
```

```yaml
# ❌ Static upstream — breaks when service restarts with new IP
upstream backend { server backend:8080; }

# ❌ VITE_ variable as runtime env (not effective after build)
environment:
  VITE_API_URL: http://newhost  # too late — bundle already built
```
