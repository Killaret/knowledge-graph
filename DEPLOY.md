# Knowledge Graph — Deployment Guide

> Practical guide for starting the **dev**, **personal**, and **test** Docker Compose stacks on a new machine.  
> For production, Kubernetes, and CI/CD details see [`docs/DEPLOYMENT_EN.md`](docs/DEPLOYMENT_EN.md).  
> Testing — [`docs/TESTING.md`](docs/TESTING.md), backups — [`docs/BACKUP.md`](docs/BACKUP.md).

---

## Table of Contents

- [What you need](#what-you-need)
- [Quick Start (TL;DR)](#quick-start-tldr)
- [Clone and branch](#clone-and-branch)
- [Setting up `.env`](#setting-up-env)
- [Databases: users, passwords, connection](#databases-users-passwords-connection)
- [Services: what to configure](#services-what-to-configure)
- [NLP model and `huggingface_cache`](#nlp-model-and-huggingface_cache)
- [The three stacks](#the-three-stacks)
- [Ports and URLs after start](#ports-and-urls-after-start)
- [Verification](#verification)
- [Creating the first user and administrator](#creating-the-first-user-and-administrator)
- [Advanced configuration: OAuth, SMTP, resources](#advanced-configuration-oauth-smtp-resources)
- [Full annotated `.env` template](#full-annotated-env-template)
- [Full deployment checklist](#full-deployment-checklist)
- [Troubleshooting by service](#troubleshooting-by-service)
- [Working inside the Docker network](#working-inside-the-docker-network)
- [Updating code](#updating-code)
- [Moving Personal data to another machine](#moving-personal-data-to-another-machine)
- [Do not touch](#do-not-touch)
- [Related documents](#related-documents)

---

## What you need

- **Windows 10/11** with **WSL2** and **Docker Desktop** (WSL2 integration enabled).
- **Git**.
- **RAM**: 8 GB minimum, 16 GB recommended (NLP model + PostgreSQL + Redis + Mongo).
- **Disk**: 30–50 GB free.
- **Internet** for the first start (to download Docker images and the NLP model), or a pre-copied `huggingface_cache`.

> For Linux/Mac the commands are the same; replace Windows `C:\` paths and PowerShell snippets with `bash` equivalents.

---

## Quick Start (TL;DR)

```powershell
git clone https://github.com/Killaret/knowledge-graph.git
cd knowledge-graph
cp .env.example .env
# open .env and replace JWT_SECRET, POSTGRES_PASSWORD, PERSONAL_POSTGRES_PASSWORD

# Download the NLP model once
docker compose -f docker-compose.personal.yml run --rm -e HF_HUB_OFFLINE=0 nlp-personal python - <<'PY'
from sentence_transformers import SentenceTransformer
SentenceTransformer('paraphrase-multilingual-MiniLM-L12-v2')
PY

# Start Personal (your private data)
docker compose -f docker-compose.personal.yml up -d --build
```

---

## Clone and branch

```powershell
git clone https://github.com/Killaret/knowledge-graph.git
cd knowledge-graph
```

The branch with the latest fixes is `ai-agents`:

```powershell
git checkout ai-agents
git pull origin ai-agents
```

---

## Setting up `.env`

```powershell
cp .env.example .env
```

`.env` **is not committed** (it is listed in `.gitignore`). Open it and fill the required fields:

| Variable | Value | Why it matters |
|----------|-------|----------------|
| `JWT_SECRET` | Long random string, at least 32 chars | Signs access/refresh tokens. If empty, auth fails. |
| `POSTGRES_PASSWORD` | Strong password | Dev database password. |
| `PERSONAL_POSTGRES_PASSWORD` | Strong password | Personal database password. |
| `TEST_POSTGRES_PASSWORD` | Password | For the isolated test stack. |
| `MONGO_URL` | `mongodb://kg-mongo-personal:27017` | MongoDB connection. App reads `MONGO_URL`, not `MONGODB_URL`. |
| `MONGO_DATABASE` | `knowledge_graph` | MongoDB database name. |

If you plan cloud backups, set `BACKUP_YANDEX_TOKEN` as an **environment variable**, not inside `.env` — see [`docs/BACKUP.md`](docs/BACKUP.md).

---

## Databases: users, passwords, connection

### PostgreSQL

When PostgreSQL is started by Docker Compose, **you do not need to create a user** — the PostgreSQL image creates a superuser and database from environment variables:

```env
PERSONAL_POSTGRES_USER=personal
PERSONAL_POSTGRES_PASSWORD=change_me_personal
PERSONAL_POSTGRES_DB=knowledge_personal
```

Under the hood (Docker init):

```sql
CREATE USER personal WITH PASSWORD 'change_me_personal' SUPERUSER;
CREATE DATABASE knowledge_personal OWNER personal;
```

The backend connects via `DATABASE_URL` (dev) or `PERSONAL_DATABASE_URL` (personal), or through a compose template:

```env
PERSONAL_DATABASE_URL=postgresql://personal:change_me_personal@postgres_personal:5432/knowledge_personal?sslmode=disable
```

#### If PostgreSQL already exists (not Docker)

1. Create the database and user:

```sql
-- connect as superuser (e.g. psql -U postgres)
CREATE USER knowledge WITH PASSWORD 'your_password';
CREATE DATABASE knowledge_personal OWNER knowledge;
CREATE DATABASE knowledge_base OWNER knowledge;
```

2. Verify access:

```powershell
psql -h 127.0.0.1 -U knowledge -d knowledge_personal -c "SELECT 1;"
```

3. Set in `.env`:

```env
PERSONAL_DATABASE_URL=postgresql://knowledge:your_password@127.0.0.1:5432/knowledge_personal?sslmode=disable
DATABASE_URL=postgresql://knowledge:your_password@127.0.0.1:5432/knowledge_base?sslmode=disable
```

4. Make sure `docker-compose.personal.yml` does not start `postgres_personal`, or the port `5433` will be taken by a different instance.

### MongoDB

In the default Docker Compose, MongoDB starts **without authentication**. You do not need a user if:

- Mongo is started by the same `docker compose` command;
- it is not exposed externally (ports bind only to `127.0.0.1` or not at all).

Environment variables read by the backend:

```env
MONGO_URL=mongodb://kg-mongo-personal:27017
MONGO_DATABASE=knowledge_graph
```

> **Important:** `.env.example` uses `MONGODB_URL` and `MONGODB_DATABASE`, but the backend expects `MONGO_URL` and `MONGO_DATABASE`. For external/secured Mongo use `MONGO_URL` and `MONGO_DATABASE`.

#### If MongoDB is external or with authentication

1. Create the user in the database:

```javascript
// connect via mongosh
use knowledge_graph;
db.createUser({
  user: "kg_user",
  pwd: "your_password",
  roles: [
    { role: "readWrite", db: "knowledge_graph" }
  ]
});
```

2. Verify access:

```powershell
mongosh "mongodb://kg_user:your_password@127.0.0.1:27017/knowledge_graph?authSource=knowledge_graph" --eval "db.getName()"
```

3. Set in `.env`:

```env
MONGO_URL=mongodb://kg_user:your_password@127.0.0.1:27017/knowledge_graph?authSource=knowledge_graph
MONGO_DATABASE=knowledge_graph
```

4. Disable the built-in Mongo service in `docker-compose.personal.yml`.

### Redis

Redis in Docker Compose starts **without a password**. Backend and graph-service connect via `REDIS_URL`:

```env
REDIS_URL=redis:6379
PERSONAL_REDIS_URL=redis_personal:6379
```

> For Personal, use `PERSONAL_REDIS_URL`. If the variable is empty, graph-service falls back to `redis:6379`, so it is better to set it explicitly.

#### Redis with password (external instance)

1. In `redis.conf` enable:

```
requirepass your_password
```

2. In `.env`:

```env
PERSONAL_REDIS_URL=redis://:your_password@127.0.0.1:6379/0
REDIS_URL=redis://:your_password@127.0.0.1:6379/0
```

3. Verify:

```powershell
redis-cli -h 127.0.0.1 -a "your_password" ping
```

4. Disable the `redis` service in `docker-compose.personal.yml`.

### Summary

| Database | Inside Docker Compose | External instance |
|---|---|---|
| **PostgreSQL** | Created automatically from `.env` | Manually create user + DB, set `DATABASE_URL` / `PERSONAL_DATABASE_URL` |
| **MongoDB** | No auth, no user needed | Manually create, set `MONGO_URL` and `MONGO_DATABASE` (not `MONGODB_URL`!) |
| **Redis** | No auth by default | Set `requirepass`, use `redis://:password@host:6379/0` |

---

## Services: what to configure

### Backend and Worker

Backend and worker are the same Go binary with different `CMD`:

- `kg-backend-personal` — HTTP API (`./server`).
- `kg-worker-personal` — background processing (`./worker`).

Required environment from `.env`:

```env
# Databases
DATABASE_URL=postgresql://...        # for dev
PERSONAL_DATABASE_URL=postgresql://... # for personal
REDIS_URL=redis://...
PERSONAL_REDIS_URL=redis://...
MONGO_URL=mongodb://...
MONGO_DATABASE=knowledge_graph

# Security
JWT_SECRET=your_secret

# Graph service
GRAPH_SERVICE_URL=http://graph-service-personal:9091
GRAPH_SERVICE_INTERNAL_TOKEN=        # if configured, see below

# NLP
NLP_SERVICE_URL=http://nlp-personal:5000
NLP_MODEL_NAME=paraphrase-multilingual-MiniLM-L12-v2

# Optional
SKIP_AUTH=false
BACKUP_ENABLED=true
```

**Verify backend:**

```powershell
curl http://127.0.0.1:18085/health
```

**Verify worker:**

The worker does not expose a port. Check logs:

```powershell
docker logs -f kg-worker-personal
```

Expected: `worker started`, `asynq: ready`, no `connection refused` errors.

### Graph Service

Graph Service is a separate Go microservice. It reads links from PostgreSQL and caches layouts in Redis. For JWT-protected requests, it must use the same `JWT_SECRET` as the backend.

Required variables:

```env
JWT_SECRET=your_secret
POSTGRES_URL=postgresql://personal:change_me_personal@postgres_personal:5432/knowledge_personal?sslmode=disable
REDIS_URL=redis://redis_personal:6379/0
NLP_MODEL_NAME=paraphrase-multilingual-MiniLM-L12-v2
```

> In practice, compose sets `POSTGRES_URL` from `PERSONAL_DATABASE_URL` and `REDIS_URL` from `PERSONAL_REDIS_URL`.

**Verify:**

```powershell
curl http://127.0.0.1:9092/health
```

**If note detail does not show links**, clear the cache:

```powershell
docker exec -i kg-redis-personal redis-cli --scan --pattern "graph-service:*" | ForEach-Object { docker exec -i kg-redis-personal redis-cli del $_ }
```

#### Graph service internal token

`GRAPH_SERVICE_INTERNAL_TOKEN` is an optional shared secret between backend and graph-service. If set:

- Backend calls graph-service with the `X-Internal-Auth` header.
- Graph-service validates it and trusts `X-User-Id`.
- It is not needed in the frontend `.env`, because the browser does not call graph-service directly.

For first start you can leave it empty. In production, set a 32+ random string.

### Frontend

The frontend is built inside Docker and served through nginx. Vite variables are baked at build time:

```env
VITE_API_URL=/api
VITE_GRAPH_SERVICE_URL=/graph-service
VITE_API_TARGET=http://backend_personal:8080
GRAPH_SERVICE_URL=http://graph-service-personal:9091
GRAPH_SERVICE_INTERNAL_TOKEN=        # for SSR only
```

> Usually these do not need editing in `.env`. Change them only if you run the frontend locally (`npm run dev`) or if the backend URL changes.

**Verify:**

```powershell
curl -s http://127.0.0.1:18084 | head
```

Expected: HTML with `__data`.

### nginx

nginx is the single gateway. Its config is mounted from `nginx.personal.conf`. Edit it only if:

- you move the stack to different ports;
- you add new `location` blocks or security headers.

- Port `18082` — API / backend.
- Port `18084` — frontend.

**Verify:**

```powershell
curl http://127.0.0.1:18082/health
curl http://127.0.0.1:18084
```

### Backup scheduler

Check whether the `backup_scheduler` service is enabled in `docker-compose.personal.yml`. If it is, it backs up Postgres on a cron schedule.

Basic variables:

```env
BACKUP_ENABLED=true
BACKUP_LOCAL_PATH=./backups
BACKUP_SCHEDULE=0 23 * * 0
BACKUP_RETENTION_DAYS=14
```

For cloud backup to Yandex:

```powershell
$env:BACKUP_YANDEX_TOKEN = "your_oauth_token"
docker compose -f docker-compose.personal.yml up -d backup_scheduler
```

Manual local backup:

```powershell
.\scripts\devops\backup-personal.ps1 -Mode daily
```

---

## NLP model and `huggingface_cache`

The NLP service runs **offline** (`HF_HUB_OFFLINE=1`) and reads the model from the bind-mount `./huggingface_cache:/root/.cache/huggingface`. On first start the folder is empty.

### Option A — with internet

Download the model through the same container:

```powershell
docker compose -f docker-compose.personal.yml run --rm -e HF_HUB_OFFLINE=0 nlp-personal python - <<'PY'
from sentence_transformers import SentenceTransformer
SentenceTransformer('paraphrase-multilingual-MiniLM-L12-v2')
PY
```

After that, `./huggingface_cache` will contain `models--sentence-transformers--paraphrase-multilingual-MiniLM-L12-v2`.

### Option B — without internet

Copy the `huggingface_cache` folder from an existing machine into the repository root.

### Option C — download locally (not Docker)

```powershell
pip install sentence-transformers
python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('paraphrase-multilingual-MiniLM-L12-v2')"
```

Then find the HuggingFace cache on your machine (`%USERPROFILE%\.cache\huggingface` on Windows) and copy it to `./huggingface_cache`.

---

## The three stacks

### Personal

The personal stack. Data lives in named volumes `pgdata_personal`, `redisdata_personal`, `mongodbdata_personal`. **Do not delete them** without a backup.

```powershell
docker compose -f docker-compose.personal.yml up -d --build
```

The `--build` flag is needed when code changes. The first run is slow because it builds all images. Later you can use:

```powershell
docker compose -f docker-compose.personal.yml up -d
```

Check status:

```powershell
docker ps --format "table {{.Names}}\t{{.Status}}"
```

### Dev

The common dev stack. Its data lives in volumes `postgres_data`, `redis_data`, `mongodb_data`.

```powershell
docker compose up -d --build
```

UI — `http://127.0.0.1:18081`, API — `http://127.0.0.1:18080`.

Dev and Personal can run at the same time (different ports), but on weaker hardware it is better to start them one by one.

### Test

Isolated stack for E2E/BDD/Playwright. Stop Dev and Personal first to avoid port conflicts.

```powershell
# Make sure other stacks are down
docker compose down
docker compose -f docker-compose.personal.yml down

# Start and seed
.\scripts\testing\start-test.ps1
.\scripts\testing\seed-test-data.ps1 -NoteCount 20 -LinkCount 10 -Seed 42 -PublicPercent 50
```

Playwright:

```powershell
cd frontend
$env:FRONTEND_URL = "http://127.0.0.1:3002"
$env:BACKEND_URL = "http://127.0.0.1:18083"
npx playwright test --project=chromium-skip-auth
```

Then:

```powershell
.\scripts\testing\stop-test.ps1
```

---

## Ports and URLs after start

| Stack | UI / Frontend | API gateway (nginx) | Backend direct | Graph service | PostgreSQL | Redis | MongoDB | NLP |
|-------|---------------|---------------------|----------------|---------------|------------|-------|---------|-----|
| **Personal** | http://127.0.0.1:18084 | http://127.0.0.1:18082 | http://127.0.0.1:18085 | http://127.0.0.1:9092 | 5433 | 16380 | 27018 | 5001 |
| **Dev** | http://127.0.0.1:18081 | http://127.0.0.1:18080 | http://127.0.0.1:9000 | http://127.0.0.1:9091 | 15432 | 16379 | 27017 | 5001 |
| **Test** | http://127.0.0.1:3002 | http://127.0.0.1:18086 | http://127.0.0.1:18083 | http://127.0.0.1:19091 | 15434 | 16381 | 27019 | 15002 |

The browser always goes through nginx (for Personal — `18084`). The backend direct port is only for debugging.

---

## Verification

```powershell
# Container list and status
docker ps --format "table {{.Names}}\t{{.Status}}"

# Health of backend / graph / nlp
curl http://127.0.0.1:18085/health
curl http://127.0.0.1:9092/health
curl http://127.0.0.1:5001/health

# Logs of a specific service
docker logs -f kg-backend-personal
docker logs -f kg-graph-service-personal
docker logs -f kg-nlp-personal

# Compose status
docker compose -f docker-compose.personal.yml ps
```

All services should be `Up` and `(healthy)`.

---

## Creating the first user and administrator

### Through the frontend

1. Open `http://127.0.0.1:18084`.
2. Go to `/register`, enter email, name, and password.
3. Log in through `/login`.
4. Create your first note. After a few seconds the NLP model will compute embeddings and recommendations appear.

### Through the API (headless / automation)

```powershell
$body = @{
    email = "admin@example.com"
    name = "Admin"
    password = "StrongPass123!"
} | ConvertTo-Json

Invoke-RestMethod -Method POST -Uri http://127.0.0.1:18082/api/v1/auth/register -ContentType "application/json" -Body $body
```

Then log in:

```powershell
$login = @{
    email = "admin@example.com"
    password = "StrongPass123!"
} | ConvertTo-Json

$response = Invoke-RestMethod -Method POST -Uri http://127.0.0.1:18082/api/v1/auth/login -ContentType "application/json" -Body $login -SessionVariable s
```

### Assign the admin role

Registration defaults to the `user` role. Migrations create three roles: `admin`, `user`, `guest`.

1. Find the role ID:

```sql
SELECT id, name FROM user_roles;
```

2. Find the user ID:

```sql
SELECT id, email FROM users;
```

3. Update the role:

```sql
UPDATE users SET role_id = (SELECT id FROM user_roles WHERE name = 'admin') WHERE email = 'admin@example.com';
```

> Run through `psql` inside the container:
> ```powershell
> docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "UPDATE users SET role_id = (SELECT id FROM user_roles WHERE name = 'admin') WHERE email = 'admin@example.com';"
> ```

### First user for the test stack

For tests, use `cmd/seed`:

```powershell
$env:APP_ENV = "test"
$env:SEED_TEST_USER_PASSWORD = "TestPassword123!"
docker compose -f docker-compose.test.yml run --rm seed
```

This creates user `testuser` with password `TestPassword123!`.

---

## Advanced configuration: OAuth, SMTP, resources

### Yandex OAuth

If you want login through Yandex:

1. Create an app in [Yandex OAuth](https://oauth.yandex.ru/).
2. In `.env`:

```env
YANDEX_CLIENT_ID=your_client_id
YANDEX_CLIENT_SECRET=your_client_secret
PKCE_ENABLED=true
```

3. Rebuild backend: `docker compose -f docker-compose.personal.yml up -d --build backend_personal`.
4. Yandex redirect URL: `http://127.0.0.1:18084/auth/yandex/callback`.

### SMTP for password reset

For production or a Personal stack exposed externally:

```env
SMTP_HOST=smtp.yandex.ru
SMTP_PORT=587
SMTP_USER=your@yandex.ru
SMTP_PASSWORD=app_password
SMTP_FROM=your@yandex.ru
```

> Gmail requires an "App Password"; for testing you can leave it empty — password reset will not work.

### Docker Desktop resources

Open **Settings → Resources → WSL integration**. Recommended:

- **Memory**: at least 6 GB, preferably 8–12 GB.
- **Swap**: 1–2 GB.
- **Disk image location**: on SSD.
- **WSL integration**: enabled for the distro from which you run Docker.

If the stack crashes with OOM, increase the limits or reduce `RECOMMENDATION_TOP_N` and `GRAPH_MAX_LIMIT` in `knowledge-graph.config.json`.

### CI/CD and GitHub Actions

The project is checked in `.github/workflows/main.yml` (on push to `main`) and `_core-checks.yml` (reusable). For CI to work:

1. The GitHub repository must have Actions enabled.
2. Secrets in `Settings → Secrets → Actions`:

| Secret | Purpose |
|---|---|
| `JWT_SECRET` | For test backend and migrations. Any 32+ character string. |
| `GRAPH_SERVICE_INTERNAL_TOKEN` | For E2E tests (if configured). |

3. **Branches**:
   - `main` — production, CI/CD, release.
   - `ai-agents` — Devin/Windsurf.
   - `feature/...` — small fixes.

4. **What CI checks**:
   - `go test ./...` in backend.
   - `npm run test:unit`, `npm run lint` in frontend.
   - `golangci-lint`.
   - Migration check against `pgvector/pgvector:pg16`.
   - Playwright and Cucumber against an isolated PostgreSQL/Redis/Mongo.

5. **Local pre-push check**:

```powershell
cd backend  && go test ./...
cd frontend && npm run lint
cd ..       && .\scripts\ci\check-stacks-identity.ps1
```

### Production: HTTPS and SSL

The local `nginx.personal.conf` uses HTTP. For production with HTTPS:

1. Get a certificate. Example with Let's Encrypt (on the host, not inside Docker):

```bash
sudo certbot certonly --standalone -d example.com
```

2. Create `nginx.production.conf` based on `nginx.personal.conf` and add:

```nginx
server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;
    ssl_protocols TLSv1.3;
    ssl_prefer_server_ciphers on;

    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header Referrer-Policy strict-origin-when-cross-origin;

    # location /api/... — copy from nginx.personal.conf
}

server {
    listen 80;
    server_name example.com;
    return 301 https://$host$request_uri;
}
```

3. Set `CORS_ALLOWED_ORIGINS=https://example.com` in `.env`.
4. Rebuild the nginx container with the production config.

> HTTPS is not required for local use. If you only need it inside your home network, a self-signed certificate works but the browser will warn.

---

## Full annotated `.env` template

> The single source of truth for environment variables is `.env.example` in the repository root.  
> Below is the full template with English comments from `.env.example`. Replace `change_me_*`, `your_*`, and empty values with your own. For the Personal stack, the key blocks are: **Database Configuration**, **Personal Development Environment**, **Redis**, **MongoDB**, **NLP Service**, **Authentication**, **OAuth**, **SMTP**, **Cloud Backup**.

<details>
<summary>Expand full `.env`</summary>

```env
# Knowledge Graph Environment Configuration
# Copy this file to .env and fill in your actual values
# DO NOT commit .env files with real credentials!

# Database Configuration
POSTGRES_USER=kb_user
POSTGRES_PASSWORD=change_me_in_production
POSTGRES_DB=knowledge_base
DATABASE_URL=postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}?sslmode=disable

# Personal Development Environment
PERSONAL_POSTGRES_USER=personal
PERSONAL_POSTGRES_PASSWORD=change_me_personal
PERSONAL_POSTGRES_DB=knowledge_personal

# Test Stack (isolated E2E/BDD testing via docker-compose.test.yml)
TEST_POSTGRES_USER=kb_user
TEST_POSTGRES_PASSWORD=kb_password
TEST_POSTGRES_DB=knowledge_test
#APP_ENV=development      # development | test | production
#SKIP_AUTH=true           # ONLY allowed when APP_ENV=test, NEVER for production
#SEED_TEST_USER_PASSWORD=TestPassword123!  # Used by cmd/seed to create the test user when APP_ENV=test
#JWT_SECRET=              # REQUIRED — set a strong random secret for the test stack

# Redis Configuration
REDIS_URL=redis:6379
PERSONAL_REDIS_URL=redis_personal:6379

# NLP Service
NLP_SERVICE_URL=http://nlp:5000
PERSONAL_NLP_SERVICE_URL=http://nlp-personal:5000

# API Configuration
API_PORT=8080
PERSONAL_API_PORT=8081

# CORS Configuration
# Comma-separated list of allowed origins (e.g., http://localhost:3000,https://example.com)
# If not set, defaults to localhost origins for development
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:5173,http://localhost:18080,http://localhost:18081,http://localhost:18082,http://localhost:18084,http://127.0.0.1:3000,http://127.0.0.1:3001,http://127.0.0.1:5173,http://127.0.0.1:18080,http://127.0.0.1:18081,http://127.0.0.1:18082,http://127.0.0.1:18084
# Allowed HTTP methods (default: GET,POST,PUT,DELETE,OPTIONS)
#CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
# Allowed headers (default: Content-Type,Authorization)
#CORS_ALLOWED_HEADERS=Content-Type,Authorization
# Preflight cache duration in seconds (default: 86400 = 24 hours)
#CORS_MAX_AGE=86400

# Frontend
VITE_API_URL=http://localhost:8080
PERSONAL_VITE_API_URL=http://localhost:8081

# =============================================================================
# Optional Configuration Overrides (env > json > default)
# These variables override values from knowledge-graph.config.json
# Uncomment and modify only if you need to change defaults
# =============================================================================

# Server Configuration
#SERVER_PORT=8080
#SERVER_RATE_LIMIT_ENABLED=false
#SERVER_RATE_LIMIT_REQUESTS=1000
#SERVER_RATE_LIMIT_WINDOW_SECONDS=60

# Database Retry Configuration
#DATABASE_RETRY_MAX_ATTEMPTS=3
#DATABASE_RETRY_DELAY_SECONDS=5
#MIGRATIONS_FAIL_ON_ERROR=false

# PostgreSQL Connection Pool (Backend)
#POSTGRES_MAX_OPEN_CONNS=25
#POSTGRES_MAX_IDLE_CONNS=5
#POSTGRES_CONN_MAX_LIFETIME_MINUTES=5
#POSTGRES_CONN_MAX_IDLE_TIME_MINUTES=1

# PostgreSQL Connection Pool (Graph Service)
#GRAPH_POSTGRES_MAX_CONNS=10
#GRAPH_POSTGRES_MIN_CONNS=2
#GRAPH_POSTGRES_MAX_CONN_LIFETIME_MINUTES=5
#GRAPH_POSTGRES_MAX_CONN_IDLE_TIME_MINUTES=1
#GRAPH_POSTGRES_HEALTH_CHECK_PERIOD_MINUTES=1

# Graph Service Integration
#GRAPH_SERVICE_URL=http://graph-service:9091
#GRAPH_SERVICE_INTERNAL_TOKEN=     # REQUIRED for backend→graph-service private calls
#GRAPH_SERVICE_TRUST_USER_HEADER=false  # Enable only on an internal network whose public proxy strips X-User-Id

# Redis Connection Pool (Both Services)
#REDIS_POOL_SIZE=10
#REDIS_MIN_IDLE_CONNS=3
#REDIS_MAX_CONN_AGE_MINUTES=5
#REDIS_POOL_TIMEOUT_SECONDS=4
#REDIS_IDLE_TIMEOUT_MINUTES=5

# Search Configuration
#SEARCH_FALLBACK_TO_ILIKE=true

# Graph API Limits
#GRAPH_DEFAULT_LIMIT=100
#GRAPH_MAX_LIMIT=1000
#GRAPH_LINK_DEFAULT_LIMIT=500
#GRAPH_LINK_MAX_LIMIT=5000
#GRAPH_LOAD_DEPTH=2

# Pagination
#PAGINATION_DEFAULT_LIMIT=100
#PAGINATION_MAX_LIMIT=300

# Recommendation Algorithm Parameters
#RECOMMENDATION_ALPHA=0.5
#RECOMMENDATION_BETA=0.5
#RECOMMENDATION_GAMMA=0.2
#RECOMMENDATION_DEPTH=3
#RECOMMENDATION_DECAY=0.5
#RECOMMENDATION_TOP_N=50
#RECOMMENDATION_CACHE_TTL_SECONDS=300
#RECOMMENDATION_TASK_DELAY_SECONDS=5
#RECOMMENDATION_BATCH_RATE_LIMIT=10
#RECOMMENDATION_FALLBACK_ENABLED=true
#RECOMMENDATION_FALLBACK_TTL_SECONDS=3600
#RECOMMENDATION_FALLBACK_SEMANTIC_ENABLED=true
#RECOMMENDATION_KEYWORD_ENABLED=true
#RECOMMENDATION_KEYWORD_SIMILARITY_METHOD=jaccard
#RECOMMENDATION_KEYWORD_TVERSKY_ALPHA=0.5
#RECOMMENDATION_KEYWORD_TVERSKY_BETA=0.5

# BFS Algorithm Settings
#BFS_AGGREGATION=max
#BFS_NORMALIZE=true

# Embedding Configuration
#EMBEDDING_SIMILARITY_LIMIT=30

# Asynq Queue Configuration
#ASYNQ_CONCURRENCY=10
#ASYNQ_QUEUE_DEFAULT=1
#ASYNQ_QUEUE_MAX_LEN=10000

# Authentication Configuration (REQUIRED)
#JWT_SECRET=
#JWT_ACCESS_TTL_SECONDS=900
#JWT_REFRESH_TTL_SECONDS=604800
#ARGON2_TIME=3
#ARGON2_MEMORY=65536
#ARGON2_THREADS=4
#API_KEY_ENABLED=true
#STATIC_API_KEY=kg-admin-2024-secret-key
#SKIP_AUTH=true  # ONLY allowed when APP_ENV=test, NEVER for production

# OAuth Configuration (Yandex)
#YANDEX_CLIENT_ID=
#YANDEX_CLIENT_SECRET=
#PKCE_ENABLED=true
#PKCE_CODE_CHALLENGE_LENGTH=128
# When PKCE is enabled, the OAuth flow uses the S256 method (RFC 7636).

# SMTP Configuration (Password Reset)
#SMTP_HOST=
#SMTP_PORT=587
#SMTP_USER=
#SMTP_PASSWORD=
#SMTP_FROM=noreply@example.com
#PASSWORD_RESET_TTL_SECONDS=900

# Password Policy
#PASSWORD_POLICY_MIN_LENGTH=10
#PASSWORD_POLICY_REQUIRE_UPPER=true
#PASSWORD_POLICY_REQUIRE_LOWER=true
#PASSWORD_POLICY_REQUIRE_DIGIT=true
#PASSWORD_POLICY_REQUIRE_SPECIAL=true

# =============================================================================
# MongoDB Configuration
# =============================================================================
# Backend reads MONGO_URL and MONGO_DATABASE (not MONGODB_*).
#MONGO_URL=mongodb://localhost:27017
#MONGO_DATABASE=knowledge_graph

# =============================================================================
# Graph Service Configuration
# =============================================================================
#GRAPH_SERVICE_GRPC_PORT=9090
#GRAPH_SERVICE_HTTP_PORT=9091
#GRAPH_SERVICE_FULL_LIMIT=1000
#GRAPH_SERVICE_DEFAULT_DEPTH=2
#GRAPH_SERVICE_EVENT_CHANNEL=graph:events

# Graph Service Cache Configuration
#GRAPH_SERVICE_CACHE_NOTE_LAYOUT_TTL_SECONDS=300
#GRAPH_SERVICE_CACHE_FULL_LAYOUT_TTL_SECONDS=300
#GRAPH_SERVICE_CACHE_DELTA_TTL_SECONDS=60

# Graph Service Layout Configuration
#GRAPH_SERVICE_LAYOUT_2D_RADIUS=100.0
#GRAPH_SERVICE_LAYOUT_3D_RADIUS=120.0
#GRAPH_SERVICE_LAYOUT_3D_Z_STEP=5.0
#GRAPH_SERVICE_LAYOUT_DEFAULT_NODE_SIZE=1.0

# Graph Service Streaming Configuration
#GRAPH_SERVICE_STREAM_CHUNK_SIZE=100

# Graph Service Event Processing
#GRAPH_SERVICE_EVENT_TRACKING_TTL_HOURS=24
#GRAPH_SERVICE_UNPROCESSED_EVENT_CHECK_INTERVAL_MINUTES=5

# =============================================================================
# NLP Service Configuration
# =============================================================================
NLP_MODEL_NAME=paraphrase-multilingual-MiniLM-L12-v2
#NLP_MAX_TEXT_LENGTH=10000

# =============================================================================
# Frontend Configuration
# =============================================================================
# Frontend Test Configuration
#FRONTEND_TEST_DEBOUNCE_TIMEOUT_MS=300
#FRONTEND_TEST_MAX_RETRY_COUNT=3
#FRONTEND_TEST_MOCK_GOTO_DELAY_MS=0

# Frontend Graph Configuration
#FRONTEND_GRAPH_2D_MAX_NODES=500
#FRONTEND_GRAPH_2D_SHADOWS_THRESHOLD=100
#FRONTEND_GRAPH_3D_MAX_NODES=500

# Frontend API Configuration
#FRONTEND_API_DEFAULT_LIMIT=100
#FRONTEND_API_LINK_LIMIT=0

# Frontend Achievements Configuration
#FRONTEND_ACHIEVEMENTS_POLL_INTERVAL_MS=7000

# =============================================================================
# Cloud Backup Configuration (Optional)
# Configure cloud backups (Yandex.Disk)
# =============================================================================
#BACKUP_CLOUD_ENABLED=false
#BACKUP_CLOUD_PROVIDER=yandex  # Options: yandex
#BACKUP_LOCAL_PATH=./backups
#BACKUP_SCHEDULE=0 23 * * 0
#BACKUP_RETENTION_DAYS=14
#BACKUP_DRAFT_TTL_HOURS=168

# Yandex.Disk configuration (OAuth token)
# CRITICAL: Set this in production to enable cloud backups
# SECURITY: Do NOT store the actual token in this file!
# Set BACKUP_YANDEX_TOKEN as an environment variable instead:
#
# Windows PowerShell:
#   $env:BACKUP_YANDEX_TOKEN = "your_oauth_token"
#   docker-compose -f docker-compose.personal.yml up -d backup_scheduler
#
# Linux/Mac:
#   export BACKUP_YANDEX_TOKEN="your_oauth_token"
#   docker-compose -f docker-compose.personal.yml up -d backup_scheduler
#
#BACKUP_YANDEX_TOKEN=your_oauth_token  # <-- NOT RECOMMENDED
#BACKUP_YANDEX_FOLDER=/KnowledgeGraphBackups
#BACKUP_YANDEX_MAX_BACKUPS=10

# =============================================================================
# CI/CD Configuration
# =============================================================================
#CI_INTEGRATION_TEST_MIGRATE_ALL=true
#CI_INTEGRATION_TEST_TRUNCATE_LIST=notes,links,embeddings,recommendations


```

</details>

---

## Full deployment checklist

Print/copy and check off:

### 1. Environment

- [ ] Docker Desktop + WSL2 installed and running.
- [ ] Repository cloned: `git clone ... && cd knowledge-graph && git checkout ai-agents`.
- [ ] `.env` created from `.env.example`.
- [ ] `JWT_SECRET` filled (32+ chars).
- [ ] PostgreSQL passwords (`POSTGRES_PASSWORD`, `PERSONAL_POSTGRES_PASSWORD`) filled.
- [ ] `MONGO_URL` and `MONGO_DATABASE` filled (if external Mongo).
- [ ] `REDIS_URL` / `PERSONAL_REDIS_URL` filled (if external Redis).
- [ ] `CORS_ALLOWED_ORIGINS` includes `http://127.0.0.1:18084` and `http://localhost:18084`.

### 2. NLP model

- [ ] `huggingface_cache` folder is present.
- [ ] Model `paraphrase-multilingual-MiniLM-L12-v2` downloaded (folder `models--sentence-transformers--...` inside).

### 3. Start

- [ ] `docker compose -f docker-compose.personal.yml up -d --build`.
- [ ] All containers `Up (healthy)`:
  - `kg-postgres-personal`
  - `kg-redis-personal`
  - `kg-mongo-personal`
  - `kg-nlp-personal`
  - `kg-graph-service-personal`
  - `kg-backend-personal`
  - `kg-worker-personal`
  - `kg-frontend-personal`
  - `kg-nginx-personal`
- [ ] `curl http://127.0.0.1:18085/health` → OK.
- [ ] `curl http://127.0.0.1:9092/health` → OK.
- [ ] `curl http://127.0.0.1:5001/health` → OK.
- [ ] `curl http://127.0.0.1:18084` → HTML.

### 4. First user

- [ ] Opened `http://127.0.0.1:18084`.
- [ ] Registration succeeded, user appears in `users` table.
- [ ] Created the first note.
- [ ] After 5–30 seconds, recommendations appeared in the note.

### 5. Backup (optional)

- [ ] `BACKUP_ENABLED=true`.
- [ ] `./backups` folder exists.
- [ ] Manual test backup: `.\scripts\devops\backup-personal.ps1`.

---

## Troubleshooting by service

### `vitest is not recognized` / frontend does not build locally

The frontend is built inside Docker. A local `npm install` is not required. For tests inside the container:

```powershell
docker exec -it kg-frontend-personal sh
npm run build
```

### NLP does not start: `Model not found`

- Check `huggingface_cache/models--sentence-transformers--...`.
- See logs: `docker logs -f kg-nlp-personal`.
- If the folder is empty, download the model again as described in [NLP model](#nlp-model-and-huggingface_cache).

### Frontend says `Could not load knowledge-graph.config.json`

- Make sure `knowledge-graph.config.json` is in the repository root.
- Rebuild the image: `docker compose -f docker-compose.personal.yml up -d --build frontend-personal`.

### Note detail does not show links

Graph service caches layouts in Redis. If you see nothing after an update:

```powershell
docker exec -i kg-redis-personal redis-cli --scan --pattern "graph-service:*" | ForEach-Object { docker exec -i kg-redis-personal redis-cli del $_ }
```

### E2E fails with `ERR_CONNECTION_REFUSED http://localhost:5173`

Playwright takes `baseURL` from `FRONTEND_URL`. On Windows `localhost` resolves to `::1`, so use `127.0.0.1`:

```powershell
$env:FRONTEND_URL = "http://127.0.0.1:3002"
```

### E2E picks an old `auth` from the wrong path

Run from `frontend/`. Relative paths `tests/setup/.auth/` resolve from the process working directory.

### Backend does not start: `migrations failed`

- Check `DATABASE_URL` / `PERSONAL_DATABASE_URL`.
- Make sure Postgres responds:

```powershell
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "SELECT 1;"
```

- If the schema is very old, you may drop the dev database, but **NEVER** Personal data.

### Graph service crashes on start

- Check `POSTGRES_URL` and `REDIS_URL` inside the container:

```powershell
docker logs -f kg-graph-service-personal
```

- Common error: `JWT_SECRET` is empty → `unauthorized` on all private endpoints.

### Worker is idle

- Check logs: `docker logs -f kg-worker-personal`.
- Make sure Redis sees the queue:

```powershell
docker exec -i kg-redis-personal redis-cli llen asynq:{default}
```

If the queue grows but the worker does not pick tasks, check `ASYNQ_CONCURRENCY` and `REDIS_URL`.

### PostgreSQL: `password authentication failed`

- Are `.env` and `docker-compose.personal.yml` using different passwords? They must match.
- If you changed the password in `.env` but the `pgdata_personal` volume was created with the old one, change the password via psql or delete the volume (with a backup) and recreate.

### MongoDB: `connection refused`

- Make sure `MONGO_URL` points to the right host. Inside Docker — `mongodb://kg-mongo-personal:27017`, outside — `mongodb://127.0.0.1:27018`.
- Check Mongo is up: `docker ps | grep mongo`.

### Backup: `BACKUP_YANDEX_TOKEN not set`

- Either set the token in the environment or disable cloud:

```env
BACKUP_CLOUD_ENABLED=false
```

- Check the `backups` folder is accessible to the container:

```powershell
docker exec -i kg-backup-scheduler ls -la /backups
```

### Everything starts, but the frontend is blank/white

- Frontend logs: `docker logs -f kg-frontend-personal`.
- Make sure `knowledge-graph.config.json` is mounted:

```powershell
docker exec -i kg-frontend-personal cat /app/knowledge-graph.config.json | head
```

- Rebuild frontend: `docker compose -f docker-compose.personal.yml up -d --build frontend-personal`.

---

## Working inside the Docker network

Docker Compose creates a dedicated bridge network. Services inside it talk to each other by name: `backend_personal`, `postgres_personal`, `redis_personal`, `mongo_personal`, `graph-service-personal`, `nlp-personal`.

### Find the network name

```powershell
docker network ls
docker network inspect knowledge-graph_default
```

> The network is usually `knowledge-graph_default` (for Personal it may be `knowledge-graph_personal_default` if the project name is set via `docker compose --project-name`).

### Enter a running container

```powershell
# PostgreSQL shell
docker exec -it kg-postgres-personal bash

# Redis shell
docker exec -it kg-redis-personal sh

# Backend shell
docker exec -it kg-backend-personal sh

# Frontend shell
docker exec -it kg-frontend-personal sh
```

### Run a PostgreSQL query

```powershell
docker compose -f docker-compose.personal.yml exec -T postgres_personal psql -U personal -d knowledge_personal -c "SELECT id, email, role_id FROM users;"
```

Or with `docker exec`:

```powershell
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -c "SELECT * FROM notes LIMIT 5;"
```

### Redis

```powershell
docker compose -f docker-compose.personal.yml exec -T redis_personal redis-cli ping
docker compose -f docker-compose.personal.yml exec -T redis_personal redis-cli --scan --pattern "graph-service:*"
```

### MongoDB

```powershell
docker compose -f docker-compose.personal.yml exec -T mongo_personal mongosh knowledge_graph --eval "db.users.countDocuments()"
```

If `mongosh` is not available in the container, open a shell:

```powershell
docker exec -it kg-mongo-personal mongosh
use knowledge_graph
db.users.find().limit(5)
```

### Run a query from the host through the published port

If the clients are installed on the host, you can connect through forwarded ports:

```powershell
psql -h 127.0.0.1 -p 5433 -U personal -d knowledge_personal -c "SELECT 1;"
redis-cli -h 127.0.0.1 -p 16380 ping
mongosh "mongodb://127.0.0.1:27018/knowledge_graph"
```

> PostgreSQL will prompt for a password. For automation: `set PGPASSWORD=change_me_personal` (Windows) or `export PGPASSWORD=...`.

### Start a temporary container inside the network

```powershell
docker run --rm --network knowledge-graph_default -it nicolaka/netshoot
```

Inside it:

```bash
ping backend_personal
dig postgres_personal
curl http://backend_personal:8080/health
curl http://graph-service-personal:9091/health
```

### Start a debug container with clients

```powershell
docker run --rm --network knowledge-graph_default -it alpine sh
# inside:
apk add --no-cache postgresql-client redis mongodb-tools
psql -h postgres_personal -U personal -d knowledge_personal -c "SELECT 1;"
redis-cli -h redis_personal ping
mongosh mongodb://mongo_personal:27017/knowledge_graph --eval "db.users.countDocuments()"
```

### Copy files into or out of a container

```powershell
# host → container
docker cp ./my-script.sql kg-postgres-personal:/tmp/

# run the script inside the container
docker exec -i kg-postgres-personal psql -U personal -d knowledge_personal -f /tmp/my-script.sql

# container → host
docker cp kg-postgres-personal:/tmp/result.csv .\result.csv
```

### Reach backend / graph service from inside the network

```bash
docker run --rm --network knowledge-graph_default curlimages/curl http://backend_personal:8080/health
docker run --rm --network knowledge-graph_default curlimages/curl http://graph-service-personal:9091/health
```

---

## Updating code

```powershell
git pull origin ai-agents

# Rebuild and restart Personal
docker compose -f docker-compose.personal.yml down
docker compose -f docker-compose.personal.yml up -d --build
```

Data in the named volumes is preserved.

---

## Moving Personal data to another machine

### SQL backup (recommended)

```powershell
.\scripts\devops\backup-personal.ps1 -Mode daily
```

A file `backups/backup-personal-daily-<timestamp>.sql.gz` is created. Copy it to the new machine and restore:

```powershell
docker cp backup-personal-daily-....sql.gz kg-postgres-personal:/tmp/
docker exec -i kg-postgres-personal gunzip -c /tmp/backup-personal-daily-....sql.gz | psql -U personal -d knowledge_personal
```

### Docker volume dump

```powershell
# Stop the stack
docker compose -f docker-compose.personal.yml down

# Backup the volumes
docker run --rm -v pgdata_personal:/data -v "${PWD}:/backup" alpine tar cvf /backup/pgdata_personal.tar /data
docker run --rm -v redisdata_personal:/data -v "${PWD}:/backup" alpine tar cvf /backup/redisdata_personal.tar /data
docker run --rm -v mongodbdata_personal:/data -v "${PWD}:/backup" alpine tar cvf /backup/mongodbdata_personal.tar /data
```

On the new machine, restore:

```powershell
docker run --rm -v pgdata_personal:/data -v "${PWD}:/backup" alpine sh -c "cd / && tar xvf /backup/pgdata_personal.tar"
# same for redis and mongo
```

---

## Do not touch

- **Do not delete** Personal named volumes `pgdata_personal`, `redisdata_personal`, `mongodbdata_personal` without a backup.
- **Do not commit** `.env` and `huggingface_cache`.
- **Do not run** E2E/BDD against the Personal stack: use the isolated test stack for that.
- **Do not edit** `docker-compose.personal.yml` unless you are sure about ports and volumes — first read [`docs/DOCKER.md`](docs/DOCKER.md).

---

## Related documents

- [`docs/DEPLOYMENT_EN.md`](docs/DEPLOYMENT_EN.md) — production and Kubernetes.
- [`docs/TESTING.md`](docs/TESTING.md) — test stack, regression, Playwright.
- [`docs/BACKUP.md`](docs/BACKUP.md) — Personal backups.
- [`docs/CONFIGURATION_EN.md`](docs/CONFIGURATION_EN.md) — environment variables and runtime config.
- [`docs/DOCKER.md`](docs/DOCKER.md) — port and volume map.
- [`docs/GRAPH_SERVICE_AUTH.md`](docs/GRAPH_SERVICE_AUTH.md) — graph service authentication.
