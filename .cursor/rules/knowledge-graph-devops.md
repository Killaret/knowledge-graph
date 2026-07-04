# Cursor Rule: knowledge-graph-devops

Build, deploy, monitor, and rollback procedures for the Knowledge Graph project.
Two stacks: **dev** (`docker-compose.yml`) and **personal** (`docker-compose.personal.yml`).

---

## Build Commands

### Go backend
```bash
# From repo root
cd backend

# Run tests with coverage
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | tail -1   # show total %

# Build binary locally
go build -ldflags="-s -w" -o ./bin/server ./cmd/server
go build -ldflags="-s -w" -o ./bin/worker ./cmd/worker

# Regenerate Swagger docs (requires swag CLI)
swag init -g cmd/server/main.go -o docs/
```

### Frontend
```bash
cd frontend
npm ci
npm run build    # SvelteKit production build → ./build/

# Unit tests
npx vitest run --coverage

# Type check
npx svelte-check --tsconfig ./tsconfig.json
```

### Docker images
```bash
# Build all images (dev stack)
docker compose build

# Build single service
docker compose build backend
docker compose build frontend

# Personal stack
docker compose -f docker-compose.personal.yml build
```

---

## Deployment Commands

### Dev stack
```bash
# Full start (waits for health checks)
docker compose up -d

# Restart single service after rebuild
docker compose up -d --no-deps --build backend

# Stop everything
docker compose down

# Stop and remove volumes (DESTRUCTIVE — wipes postgres data)
docker compose down -v
```

### Personal stack
```bash
docker compose -f docker-compose.personal.yml up -d
docker compose -f docker-compose.personal.yml up -d --no-deps --build backend_personal
docker compose -f docker-compose.personal.yml down
```

### Run database migrations manually
```bash
docker compose exec backend ./server migrate   # or via cli target
# Direct binary (if migrations binary is built):
docker compose run --rm backend ./cli migrate
```

---

## Log Monitoring

```bash
# Follow all services
docker compose logs -f

# Follow single service
docker compose logs -f backend
docker compose logs -f nlp
docker compose logs -f worker

# Last 100 lines
docker compose logs --tail=100 backend

# Personal stack
docker compose -f docker-compose.personal.yml logs -f backend_personal
```

Key log patterns to watch:
- `[GIN]` — HTTP request logs from backend
- `Embedding model loaded via cache` — NLP model ready
- `migration applied:` — database migration success
- `asynq: Failed to process` — task queue error

---

## Health Check Verification

```bash
# Verify all services healthy
docker compose ps

# Manual health checks
curl -f http://localhost:8080/health          # nginx / backend gateway
curl -f http://localhost:9000/health          # backend direct (dev only)
curl -f http://localhost:9091/health          # graph-service direct
curl -f http://localhost:5000/health          # nlp service (triggers model load)

# Personal stack ports
curl -f http://localhost:8082/health          # nginx personal
curl -f http://localhost:9092/health          # graph-service personal
curl -f http://localhost:5001/health          # nlp personal

# Batch health check script
for svc in 8080 9000 9091 5000; do
  echo -n "Port $svc: "
  curl -sf http://localhost:$svc/health && echo OK || echo FAIL
done
```

---

## Rollback Procedures

### Quick rollback (redeploy previous image)
```bash
# Tag before deploying new version
docker tag knowledge-graph-backend:latest knowledge-graph-backend:rollback

# If new deployment fails, restore
docker compose stop backend
docker tag knowledge-graph-backend:rollback knowledge-graph-backend:latest
docker compose up -d --no-deps backend
```

### Database rollback
```bash
# Run down migration (if .down.sql files exist)
docker compose exec backend ./cli migrate-down 1

# Restore from backup (personal stack)
docker compose -f docker-compose.personal.yml exec postgres_personal \
  psql -U personal -d knowledge_personal < /backups/latest.sql
```

---

## Environment-Specific Deployment

### Dev environment checklist
- [ ] `SKIP_AUTH=false` (or `true` only for test runs)
- [ ] `REDIS_FLUSH_ON_STARTUP=false` (avoid wiping cache on restart)
- [ ] `HF_HUB_OFFLINE=1` (use local model cache, no network calls)
- [ ] `.env` file present (copy from `.env.example`)

### Personal environment checklist
- [ ] `BACKUP_YANDEX_TOKEN` set in `.env`
- [ ] `BACKUP_CLOUD_ENABLED=true` in compose environment
- [ ] `PERSONAL_POSTGRES_*` credentials in `.env`
- [ ] `huggingface_cache/` populated before first run

---

## Release Checklist

Before merging to main:
- [ ] `go test ./... -cover` passes with >60% coverage
- [ ] `npx vitest run` passes
- [ ] `docker compose build` succeeds (all stages)
- [ ] `docker compose up -d` + all health checks green
- [ ] `swag init` regenerated if any handler annotations changed
- [ ] No new `.env` secrets committed (run `git diff --name-only | grep .env`)
- [ ] Legacy `/notes` routes not extended (new routes under `/api/v1/`)
- [ ] `knowledge-graph-data.md` updated if new Redis keys or DB tables added

---

## Useful One-Liners

```bash
# Show running containers and their health
docker compose ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"

# Enter backend container
docker compose exec backend sh

# Check Postgres directly
docker compose exec postgres psql -U kb_user -d knowledge_base

# Flush Redis (dev only — never in personal with data)
docker compose exec redis redis-cli FLUSHALL

# View NLP model load time
docker compose logs nlp | grep "model loaded"

# Check asynq worker queue depth (via redis-cli)
docker compose exec redis redis-cli LLEN "asynq:{default}:pending"
```
