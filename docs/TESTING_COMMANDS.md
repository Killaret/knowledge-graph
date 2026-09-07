## Testing Commands & Procedures

**Comprehensive testing commands for AI agents and developers:**

### Frontend Testing
```bash
# Unit tests (Vitest)
cd frontend && npm run test:unit

# E2E tests (Playwright)
cd frontend && npx playwright test

# Visual regression tests
cd frontend && npx playwright test --grep="@visual"

# BDD tests (Cucumber)
cd frontend && npm run test:bdd

# Real-auth Playwright / BDD (requires SKIP_AUTH=false test stack)
cd frontend && npm run test:realauth
cd frontend && npm run test:bdd:realauth

# If the frontend is on a custom port (e.g. FRONTEND_PORT=50070 due to Hyper-V):
# PowerShell
cd frontend; $env:SKIP_AUTH='false'; $env:FRONTEND_URL='http://127.0.0.1:50070'; $env:BACKEND_URL='http://127.0.0.1:18083'; node scripts/run-bdd.cjs
# Bash
SKIP_AUTH=false FRONTEND_URL=http://127.0.0.1:50070 BACKEND_URL=http://127.0.0.1:18083 node scripts/run-bdd.cjs

# Skip-auth Playwright / BDD (default test stack)
cd frontend && npm run test:skipauth
cd frontend && npm run test:bdd:skipauth

# Build verification
cd frontend && npm run build
```

### Backend Testing
```bash
# Unit tests
cd backend && go test ./...

# Integration tests
cd backend && go test -tags=integration ./...

# Race detection (requires CGO_ENABLED=1)
cd backend && CGO_ENABLED=1 go test -race ./...

# Build verification
cd backend && go build ./cmd/server
```

### Unified Test Entry Point
```bash
# All test layers (unit → integration → e2e → bdd)
.\scripts\testing\test.ps1 all          # Windows
./scripts/testing/test.sh all            # Linux/Mac

# Individual targets
.\scripts\testing\test.ps1 unit         # backend + frontend unit
.\scripts\testing\test.ps1 integration  # backend integration tests
.\scripts\testing\test.ps1 e2e          # Playwright E2E
.\scripts\testing\test.ps1 bdd          # Cucumber BDD
.\scripts\testing\test.ps1 coverage     # backend + frontend coverage
.\scripts\testing\test.ps1 clean        # cleanup temporary artifacts
```

### NLP Service Testing
```bash
# Unit tests
cd nlp-service && pytest tests/ -v

# Health check
curl http://localhost:5000/health

# API tests
curl -X POST http://localhost:5000/extract_keywords -H "Content-Type: application/json" -d '{"text":"test","top_n":3}'
curl -X POST http://localhost:5000/embed -H "Content-Type: application/json" -d '{"text":"test"}'
```

### Test Stack Management
```bash
# Full test cycle (isolated model - stops dev/personal stacks)
.\scripts\testing\run-full-test-cycle.ps1      # Windows
./scripts/testing/run-full-test-cycle.sh       # Linux/Mac

# Start test stack (skip-auth by default)
.\scripts\testing\start-test.ps1              # Windows
./scripts/testing/start-test.sh               # Linux/Mac

# Start test stack with real authentication
# The test frontend is built with VITE_API_URL=/api so browser API calls are
# proxied through SvelteKit to the backend and graph service.
# All services (backend, graph-service, worker, frontend) must see SKIP_AUTH=false.
$env:SKIP_AUTH="false"; .\scripts\testing\start-test.ps1   # Windows PowerShell
SKIP_AUTH=false ./scripts/testing/start-test.sh           # Linux/Mac

# If the default frontend port 3002 is blocked (e.g. inside a Windows Hyper-V
# excluded range), override it with FRONTEND_PORT:
$env:FRONTEND_PORT="50070"; docker compose -f docker-compose.test.yml up -d --build   # Windows
FRONTEND_PORT=50070 docker compose -f docker-compose.test.yml up -d --build          # Linux/Mac

# Seed test data
# IMPORTANT: for real-auth testing the test stack must be running with SKIP_AUTH=false,
# otherwise the seeded notes are owned by the anonymous skip-auth user and real-auth
# graph requests will return an empty graph.
$env:SKIP_AUTH="false"; .\scripts\testing\seed-test-data.ps1          # Windows
SKIP_AUTH=false ./scripts/testing/seed-test-data.sh                   # Linux/Mac

# Stop and destroy test stack
.\scripts\testing\stop-test.ps1               # Windows
./scripts\testing\stop-test.sh                # Linux/Mac

# Manual test stack management
docker compose -f docker-compose.test.yml up -d --build
docker compose -f docker-compose.test.yml down -v
```

### Stack Health Checks
```bash
# Check all stacks (default)
.\scripts\ci\check-stacks-health.ps1              # Windows
./scripts/ci/check-stacks-health.sh               # Linux/Mac

# Check specific stack
.\scripts\ci\check-stacks-health.ps1 -Stack dev   # Windows
.\scripts\ci\check-stacks-health.ps1 -Stack personal
.\scripts\ci\check-stacks-health.ps1 -Stack test
./scripts/ci/check-stacks-health.sh --stack dev    # Linux/Mac
./scripts/ci/check-stacks-health.sh --stack personal
./scripts/ci/check-stacks-health.sh --stack test
```

### Regression Testing
```bash
# Full regression cycle (isolated model - 25 steps)
.\scripts\testing\run-full-test-cycle.ps1      # Windows
./scripts/testing/run-full-test-cycle.sh       # Linux/Mac

# Stacks identity check
.\scripts\ci\check-stacks-identity.ps1    # Windows
./scripts/ci/check-stacks-identity.sh     # Linux/Mac

# Cleanup test artifacts
python .\scripts\cleanup\cleanup-test-artifacts.py    # Windows
python ./scripts/cleanup/cleanup-test-artifacts.py     # Linux/Mac

# Individual regression steps
# Step 0: Capture dev stack state snapshot
# Step 1-2: Stop dev and personal stacks
# Step 3: Check stacks identity
# Step 4: Start test stack
# Step 5: Seed test data
# Step 6-17: Run tests and verifications (unit, integration, API, manual)
# Step 18: Documentation verification
# Step 19: Stop test stack
# Step 20: Cleanup temporary files
# Step 21-22: Restore dev and personal stacks
# Step 23: Compare states and verify identity + health
# Step 24: Auto-commit if all checks pass
```

### Database Verification
```bash
# PostgreSQL (test stack)
docker exec kg-test-postgres psql -U kb_user -d knowledge_test -c "SELECT extname FROM pg_extension WHERE extname = 'vector';"
docker exec kg-test-postgres psql -U kb_user -d knowledge_test -c "SELECT COUNT(*) FROM note_embeddings;"

# Redis (test stack)
docker exec kg-test-redis redis-cli PING
docker exec kg-test-redis redis-cli KEYS "*"

# MongoDB (test stack)
docker exec kg-test-mongo mongosh --eval "db.adminCommand('ping')"
docker exec kg-test-mongo mongosh --eval "db.getCollectionNames()"
```

### Health Checks
```bash
# Dev stack
curl http://localhost:18080/health           # Nginx gateway
curl http://localhost:9000/health           # Backend
curl http://localhost:9091/health           # Graph service
curl http://localhost:18080/api/v1/notes     # Notes API

# Personal stack
curl http://localhost:18085/health           # Personal backend
curl http://localhost:18082/health           # Personal API gateway
curl http://localhost:18084/health           # Personal frontend gateway
curl http://localhost:8092/health           # Personal graph service

# Test stack
curl http://localhost:18083/health           # Test backend
curl http://localhost:3002                  # Test frontend
curl http://localhost:15002/health          # Test NLP service
curl http://localhost:19091/health          # Test graph service (HTTP)
```

### Testing Best Practices
- **ALWAYS** use isolated test stack for E2E and BDD testing
- **NEVER** run E2E/BDD tests against dev or personal stacks
- **ALWAYS** verify stacks identity before regression testing
- **ALWAYS** destroy test stack with `down -v` after testing
- **ALWAYS** use i18n keys or `data-testid` selectors in frontend tests to avoid brittle locale-specific text
- **ALWAYS** run unit tests before integration tests
- **ALWAYS** verify health endpoints before API testing
- **ALWAYS** use the full test cycle script for regression testing (isolated model)
- **ALWAYS** check dev stack state before/after testing for data leakage
- **ALWAYS** verify dev/personal identity after testing
- **ALWAYS** stop dev/personal stacks before E2E/BDD; running all stacks together causes Docker instability and Windows `localhost` → `::1` Playwright failures
- **ALWAYS** use `http://127.0.0.1:3002` / `http://127.0.0.1:18083` on Windows, or rebuild the test frontend with `VITE_API_URL=http://127.0.0.1:18083`
- **ALWAYS** keep `.env` aligned with existing `postgres_data` / `pgdata_personal` volume passwords (`POSTGRES_PASSWORD`, `PERSONAL_POSTGRES_PASSWORD`) and set `JWT_SECRET`
- **For real-auth tests** start the test stack with `SKIP_AUTH=false`; otherwise `/api/v1/users/me` returns 500 and `/api/v1/auth/refresh` returns 400
- **If Windows Hyper-V blocks port 3002**, set `FRONTEND_PORT` (e.g. `50070`) and matching `FRONTEND_URL`
- **Playwright is configured with `workers: 2`** to keep real-auth manual-checklist tests stable on Windows; do not override above 2 workers for real-auth runs

### Windows E2E/BDD URL workaround

```powershell
$env:FRONTEND_URL = "http://127.0.0.1:3002"   # or the FRONTEND_PORT you used, e.g. 50070
$env:BACKEND_URL = "http://127.0.0.1:18083"
npx playwright test --project=chromium-skip-auth
```

### Real-auth Playwright run

```powershell
# Start stack first with SKIP_AUTH=false and (if needed) FRONTEND_PORT=50070
$env:FRONTEND_URL = "http://127.0.0.1:50070"
$env:BACKEND_URL = "http://127.0.0.1:18083"
$env:SKIP_AUTH = "false"
npx playwright test --project=chromium-real-auth
```

### Rebuild test frontend for IPv4

```bash
docker compose -f docker-compose.test.yml build --build-arg VITE_API_URL=http://127.0.0.1:18083 frontend-test
```

---
