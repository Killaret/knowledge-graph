# Comprehensive Regression Test Plan

**Version:** 1.0  
**Last Updated:** 2026-07-20  
**Status:** Active

## Overview

This document provides a comprehensive regression testing plan for Knowledge Graph, covering all layers of the application stack, infrastructure, and deployment configurations.

## Test Environments

| Stack | Docker Compose | Frontend Port | Backend Port | Database | Purpose |
|-------|----------------|----------------|---------------|----------|---------|
| Dev | docker-compose.yml | 5173 (3001 dev container) | 18080 (9000 backend) | knowledge_base | Development |
| Personal | docker-compose.personal.yml | 3001 | 18082 (18085 backend) | knowledge_personal | Personal use |
| Test | docker-compose.test.yml | 3002 (3000) | 18083 (8080 backend) | knowledge_test | Isolated testing |

## Prerequisites

### Required Tools
- Docker Desktop (running)
- Docker Compose v2+
- Go 1.25+
- Node 20+
- Python 3.11+
- PowerShell (Windows) or Bash (Linux/Mac)

### Required Scripts
- `scripts/ci/check-stacks-health.ps1/.sh`
- `scripts/ci/check-stacks-identity.ps1/.sh`
- `scripts/testing/start-test.ps1/.sh`
- `scripts/testing/seed-test-data.ps1/.sh`
- `scripts/testing/stop-test.ps1/.sh`
- `scripts/testing/run-full-test-cycle.ps1/.sh`
- `scripts/testing/test.ps1/.sh` (unified test entry point)
- `scripts/cleanup/cleanup-test-artifacts.py/.sh` (temporary artifact cleanup)

### Environment Isolation Notes

- **Only the test stack may run during E2E/BDD/regression.** Stop dev and personal stacks before the cycle to prevent Docker instability, resource exhaustion, and port/network conflicts.
- On Windows, Playwright/Node resolves `localhost` to `::1` while Docker binds published ports to `127.0.0.1`. Use `http://127.0.0.1:3002` and `http://127.0.0.1:18083`, or rebuild the test frontend with `VITE_API_URL=http://127.0.0.1:18083`.
- `.env` must contain `JWT_SECRET` and DB passwords matching the existing `postgres_data` / `pgdata_personal` volumes (`POSTGRES_PASSWORD`, `PERSONAL_POSTGRES_PASSWORD`).

---

## PART 0: Stacks Identity Check

### 0.1 Compare Docker Compose Files
**Files:** docker-compose.yml, docker-compose.personal.yml, docker-compose.test.yml

**Checks:**
- [ ] Service names are consistent (backend, frontend, nlp, postgres, redis, mongo)
- [ ] Image versions are identical across stacks
- [ ] Build contexts are identical
- [ ] Port mappings are different (no conflicts)
- [ ] Volume names are different (dev: postgres_data, personal: pgdata_personal, test: test_postgres_data)
- [ ] Environment variables point to correct service names

### 0.1.5 Verify Healthchecks in Dockerfiles
**Files:** backend/Dockerfile, frontend/Dockerfile, nlp-service/Dockerfile

**Checks:**
- [ ] Backend Dockerfile contains HEALTHCHECK directive
- [ ] Frontend Dockerfile contains HEALTHCHECK directive
- [ ] NLP Dockerfile contains HEALTHCHECK directive
- [ ] Healthcheck endpoints are accessible (http://localhost:9000/health, http://localhost:8000/health)
- [ ] Healthcheck commands use appropriate intervals and timeouts

**Expected Result:** STACKS_IDENTICAL

### 0.2 Check Service Versions
**From Dockerfiles:**
- [ ] Go version: golang:1.25-alpine (backend/Dockerfile)
- [ ] Node version: node:20-alpine (frontend/Dockerfile)
- [ ] Python version: python:3.11-slim (nlp-service/Dockerfile)

**Expected Result:** All versions consistent across stacks

### 0.3 Compare Configuration Files
**Files:** knowledge-graph.config.json (mounted in all stacks)

**Checks:**
- [ ] backend.recommendation.* settings are identical
- [ ] backend.graph.* settings are identical
- [ ] frontend.graph.* settings are identical
- [ ] auth.* settings are identical
- [ ] backup.* settings are identical
- [ ] No debug flags enabled in test stack

**Expected Result:** Configurations identical except stack-specific values

### 0.4 Check Database Structure
**Tables:** notes, links, tags, note_tags, note_embeddings, note_keywords, note_recommendations, users, user_settings, achievements, user_achievements, api_keys, note_shares, share_links, tasks, task_history

**Checks:**
- [ ] Table names are identical across stacks
- [ ] Indexes are identical
- [ ] Migrations are applied (schema_migrations table)
- [ ] Migration versions match

**Expected Result:** Database structures identical

### 0.5 Check Health and API
**Dev Stack:**
- [ ] curl http://localhost:18080/health → 200
- [ ] curl http://localhost:18080/api/v1/notes?limit=1 → JSON

**Personal Stack:**
- [ ] curl http://localhost:18082/health → 200
- [ ] curl http://localhost:18082/api/v1/notes?limit=1 → JSON

**Expected Result:** All stacks healthy and accessible

---

## PART 0.5: Docker Build Verification

### 0.5.1 Build Dev Stack
```bash
docker compose build --no-cache
```
**Expected:** Exit code 0, all images built successfully

### 0.5.2 Check Image Sizes
```bash
docker images | grep knowledge-graph
```
**Expected:**
- backend: < 20 MB (alpine)
- frontend: < 50 MB (nginx/node-alpine)
- nlp: < 500 MB (ML model)

### 0.5.3 Check Healthchecks
**Checks:**
- [ ] backend/Dockerfile contains HEALTHCHECK
- [ ] frontend/Dockerfile contains HEALTHCHECK
- [ ] nlp-service/Dockerfile contains HEALTHCHECK

### 0.5.4 Check for Secrets in Images
```bash
docker run --rm kg-backend env | grep -E 'PASSWORD|SECRET|TOKEN'
docker history kg-backend --no-trunc | grep -E 'PASSWORD|SECRET|TOKEN'
```
**Expected:** No secrets in images

### 0.5.5 Build Personal Stack
```bash
docker compose -f docker-compose.personal.yml build --no-cache
```
**Expected:** Exit code 0

### 0.5.6 Build Test Stack
```bash
docker compose -f docker-compose.test.yml build --no-cache
```
**Expected:** Exit code 0

---

## PART 0.6: Environment Configuration Check

### 0.6.1 Check .env.example
**Checks:**
- [ ] .env.example exists
- [ ] Contains all required variables (DATABASE_URL, JWT_SECRET)
- [ ] No actual secrets in .env.example

### 0.6.2 Check Required Environment Variables
**Checks:**
- [ ] DATABASE_URL documented
- [ ] JWT_SECRET documented
- [ ] REDIS_URL documented
- [ ] NLP_SERVICE_URL documented

---

## PART 1: Dev/Personal Stacks Health

### 1.1 Run check-stacks-health
```bash
./scripts/ci/check-stacks-health.sh  # Linux/Mac
./scripts/ci/check-stacks-health.ps1  # Windows
```

**Checks:**
- [ ] Dev containers running (9 containers)
- [ ] Dev health endpoint: http://localhost:18080/health → 200
- [ ] Dev API: http://localhost:18080/api/v1/notes?limit=1 → JSON
- [ ] Personal containers running (9 containers)
- [ ] Personal health endpoint: http://localhost:18082/health → 200
- [ ] Personal API: http://localhost:18082/api/v1/notes?limit=1 → JSON
- [ ] Dev frontend: http://localhost:18081 → loads
- [ ] Personal frontend: http://localhost:18084 → loads

**Expected Result:** All stacks healthy

---

## PART 2: Test Stack Startup

### 2.1 Start Test Stack
```bash
./scripts/testing/start-test.sh  # Linux/Mac
./scripts/testing/start-test.ps1  # Windows
```

**Checks:**
- [ ] Previous test stack destroyed
- [ ] Test stack built successfully
- [ ] All containers healthy
- [ ] Test stack URLs displayed

### 2.2 Seed Test Data
```bash
./scripts/testing/seed-test-data.sh  # Linux/Mac
./scripts/testing/seed-test-data.ps1  # Windows
```

**Checks:**
- [ ] Test user registered (testuser / TestPassword123!)
- [ ] 5 test notes created (star, planet, comet, galaxy, asteroid)
- [ ] 2 test links created

### 2.3 Verify Test Stack Health
**Checks:**
- [ ] Test health: http://localhost:18083/health → 200
- [ ] Test API: http://localhost:18083/api/v1/notes?limit=1 → JSON (5 notes)
- [ ] Test frontend: http://localhost:3002 → loads

**Expected Result:** Test stack ready for testing

---

## PART 3: NLP Service Tests

### 3.1 Run NLP Unit Tests
```bash
cd nlp-service
pytest -v 2>&1 | tee ../logs/test-outputs/test-nlp.log
```

**Expected:** 46 tests collected (all non-skipped pass; skips allowed due to model/cache)

### 3.2 Check NLP Health
```bash
curl http://localhost:5000/health
```
**Expected:** {"status":"ok","model_loaded":true}

### 3.3 Test Keyword Extraction
```bash
curl -X POST http://localhost:5000/extract_keywords \
  -H "Content-Type: application/json" \
  -d '{"text":"Knowledge Graph testing","top_n":5}'
```
**Expected:** JSON with keywords

### 3.4 Test Embedding
```bash
curl -X POST http://localhost:5000/embed \
  -H "Content-Type: application/json" \
  -d '{"text":"test embedding"}'
```
**Expected:** JSON with embedding vector

---

## PART 3.5: Dependencies and Vulnerabilities

### 3.5.1 Check Backend Dependencies
```bash
cd backend
go mod tidy
go list -m all
```
**Expected:** No errors, dependencies resolved

### 3.5.2 Check Frontend Dependencies
```bash
cd frontend
npm audit
```
**Expected:** No high/critical vulnerabilities

### 3.5.3 Check NLP Dependencies
```bash
cd nlp-service
pip list
pip-audit  # if available
```
**Expected:** No high/critical vulnerabilities

---

## PART 4: Backend Tests

### 4.1 Run Backend Unit Tests
```bash
cd backend
go test -race ./... -count=1 2>&1 | tee ../logs/test-outputs/test-backend-unit.log
```

**Expected:** All backend unit tests pass (no failures).

### 4.2 Run Backend Integration Tests
```bash
cd backend
go test -tags=integration ./... -count=1 -p=1 2>&1 | tee ../logs/test-outputs/test-backend-integration.log
```

**Expected:** All integration tests pass in Linux/WSL Docker environment. On Windows with rootless Docker, testcontainers may not start; use WSL2 backend or CI runner.

---

## PART 5: Backend API Verification

### 5.1 Authentication Endpoints
**Base URL:** http://localhost:18083

**Checks:**
- [ ] POST /api/v1/auth/register → 201
- [ ] POST /api/v1/auth/login → 200 + token
- [ ] POST /api/v1/auth/refresh → 200
- [ ] POST /api/v1/auth/logout → 204

### 5.2 Notes Endpoints
**Checks:**
- [ ] GET /api/v1/notes → 200
- [ ] POST /api/v1/notes → 201
- [ ] GET /api/v1/notes/{id} → 200
- [ ] PUT /api/v1/notes/{id} → 200
- [ ] DELETE /api/v1/notes/{id} → 204
- [ ] GET /api/v1/notes/{id}/suggestions → 200
- [ ] GET /api/v1/notes/search?q=test → 200

### 5.3 Links Endpoints
**Checks:**
- [ ] GET /api/v1/links → 200
- [ ] POST /api/v1/links → 201
- [ ] GET /api/v1/links/{id} → 200
- [ ] DELETE /api/v1/links/{id} → 204

### 5.4 Graph Service Endpoints
**Checks:**
- [ ] GET /graph-service/api/v1/graph/full → 200
- [ ] GET /graph-service/api/v1/graph/{noteId}?depth=2 → 200
- [ ] GET /graph-service/api/v1/graph/{noteId}/links → 200

### 5.5 Additional Endpoints
**Checks:**
- [ ] Tags CRUD → 200/201/204
- [ ] Users/me → 200
- [ ] Users/me/settings → 200
- [ ] Achievements → 200
- [ ] Users/me/achievements → 200
- [ ] Export → 200
- [ ] Import → 201

**Record:** Response code and time for each endpoint

---

## PART 5.5: Security Verification

### 5.5.1 Check CORS Headers
```bash
curl -I http://localhost:18083/api/v1/notes
```
**Expected:** CORS headers present

**Environment Variables:**
- `CORS_ALLOWED_ORIGINS` - Comma-separated list of allowed origins (e.g., `http://localhost:3000,http://localhost:3001,http://localhost:3002`)
- `CORS_ALLOWED_METHODS` - Allowed HTTP methods (default: `GET,POST,PUT,DELETE,OPTIONS`)
- `CORS_ALLOWED_HEADERS` - Allowed headers (default: `Content-Type,Authorization`)
- `CORS_MAX_AGE` - Preflight cache duration in seconds (default: `86400`)

**Configuration Locations:**
- Backend: `backend/cmd/server/middleware.go` (reads from environment)
- Docker Compose: Set in `docker-compose.yml`, `docker-compose.personal.yml`, `docker-compose.test.yml`
- Dev/Personal: Localhost origins whitelisted
- Test: Test stack ports whitelisted
- Production: Must configure production origins via environment variables

### 5.5.2 Check Rate Limiting
**Test:** Send rapid requests to write endpoint
**Expected:** Rate limiting active (429 after threshold)

### 5.5.3 Check JWT Validation
**Test:** Access protected endpoint without token
**Expected:** 401 Unauthorized

---

## PART 6: Asynchronous Tasks (Asynq)

### 6.1 Check Worker Status
```bash
docker logs kg-test-worker --tail=20 | grep "Worker started"
```
**Expected:** Worker started successfully

### 6.2 Create Note and Check Task
**Steps:**
1. Create a note via API
2. Check Redis for recommendation task: `KEYS *recommendation*`
**Expected:** Task appears in Redis

### 6.3 Check Worker Logs
```bash
docker logs kg-test-worker --tail=20
```
**Expected:** Task processed successfully

### 6.4 Check Recommendations in DB
```sql
SELECT COUNT(*) FROM note_recommendations;
```
**Expected:** Recommendations saved

---

## PART 6.5: Logs and Monitoring

### 6.5.1 Check Backend Logs
```bash
docker logs kg-test-backend --tail=50
```
**Expected:** No ERROR level logs

### 6.5.2 Check Frontend Logs
```bash
docker logs kg-test-frontend --tail=50
```
**Expected:** No ERROR level logs

### 6.5.3 Check NLP Logs
```bash
docker logs kg-test-nlp --tail=50
```
**Expected:** No ERROR level logs

---

## PART 7: PGVECTOR Verification

### 7.1 Check pgvector Extension
```sql
SELECT * FROM pg_extension WHERE extname = 'vector';
```
**Expected:** Extension installed

### 7.2 Check Embeddings Count
```sql
SELECT COUNT(*) FROM note_embeddings;
```
**Expected:** Embeddings exist

### 7.3 Check Cosine Similarity
```sql
SELECT id, title, 
       1 - (embedding <=> '[0,0,0]') as similarity
FROM notes 
ORDER BY similarity DESC 
LIMIT 5;
```
**Expected:** Query executes successfully

---

## PART 8: Redis and MongoDB

### 8.1 Redis Verification
**Checks:**
- [ ] Redis PING → PONG
- [ ] Check keys: `KEYS setting:*`
- [ ] Check keys: `KEYS refresh:blacklist:*`
- [ ] No expired keys blocking operations

### 8.2 MongoDB Verification
**Checks:**
- [ ] MongoDB ping successful
- [ ] List collections: `show collections`
- [ ] Drafts collection exists
- [ ] No orphaned data

---

## PART 8.3: Backup Verification

### 8.3.1 Check Backup Scheduler
```bash
docker logs kg-backup-scheduler --tail=20
```
**Expected:** Scheduler running (personal stack only)

### 8.3.2 Check Backup Directory
```bash
ls -la backups/
```
**Expected:** Backup directory exists (personal stack only)

---

## PART 9: Frontend Tests

### 9.1 Run Frontend Unit Tests
```bash
cd frontend
npm run test:unit 2>&1 | tee ../logs/test-outputs/test-frontend-unit.log
```

**Expected:** All frontend unit tests pass (0 failures, skipped tests allowed). UI uses Russian (`ru`) locale by default; tests should rely on i18n keys or `data-testid` selectors.

### 9.2 Run Frontend E2E Tests (two-phase)

Phase 1 — `SKIP_AUTH=true` test stack (skip-auth tests):
```bash
cd frontend
FRONTEND_URL=http://localhost:3002 BACKEND_URL=http://localhost:18083 npm run test:skipauth
```

**Expected:** All non-`@auth-real` E2E tests pass (skips allowed).

Phase 2 — `SKIP_AUTH=false` test stack (real auth tests):
```bash
# Rebuild the test stack with SKIP_AUTH=false first, then:
cd frontend
FRONTEND_URL=http://localhost:3002 BACKEND_URL=http://localhost:18083 npm run test:realauth
```

**Expected:** All `@auth-real` tests pass.

### 9.3 Run Frontend Visual Tests
```bash
cd frontend
npx playwright test --project=visual 2>&1 | tee ../logs/test-outputs/test-frontend-visual.log
```

**Expected:** Visual tests pass

### 9.4 Run Frontend BDD Tests

SKIP_AUTH mode:
```bash
cd frontend
FRONTEND_URL=http://localhost:3002 BACKEND_URL=http://localhost:18083 npm run test:bdd:skipauth
```

**Expected:** All scenarios pass against `SKIP_AUTH=true` stack.

Real-auth BDD is not yet implemented; skip or add login-based step definitions when needed.

---

## PART 10: Public Graph Verification

### 10.1 Create Public Notes and Links
**Steps:**
1. Create 3 public notes (star, planet, comet)
2. Create 2 links (A→B reference 1.0, B→C dependency 0.8)

### 10.2 Check Public API
```bash
curl http://localhost:18083/graph-service/api/v1/graph/full?limit=100
```
**Expected:** JSON with 3 nodes and 2 links

### 10.3 Check Public Frontend
**URL:** http://localhost:3002/graph
**Expected:** 3 nodes and 2 links visible

### 10.4 Check Private Note Filtering
**Steps:**
1. Create private note (not published)
2. Create link to private note
3. Check public graph
**Expected:** Private note and link not visible

---

## PART 11: CI/CD Verification

### 11.1 Check Workflow Files
**Files:**
- [ ] .github/workflows/ci.yml exists
- [ ] .github/workflows/main.yml exists (ci-full.yml equivalent)
- [ ] .github/workflows/frontend-tests.yml exists

### 11.2 Check Workflow Configurations
**Checks:**
- [ ] Go version: 1.25
- [ ] Node version: 20
- [ ] Database env vars: kb_user/knowledge_base (not kg_user/knowledge_graph)
- [ ] `backend-integration-tests` job exists in `.github/workflows/main.yml` and `.github/workflows/ci.yml` and runs `go test -tags=integration ./...`

### 11.3 Check Secrets
**Checks:**
- [ ] ARGOS_TOKEN configured (for visual tests)
- [ ] CODECOV_TOKEN configured (for coverage)
- [ ] DOCKER_USERNAME/DOCKER_PASSWORD configured (if deployment enabled)

---

## PART 11.4: Documentation Verification

### 11.4.1 Check OpenAPI Spec
```bash
curl http://localhost:18083/openapi.yaml
```
**Expected:** Valid OpenAPI 3.0 spec

### 11.4.2 Check README
**Checks:**
- [ ] README.md exists
- [ ] Installation instructions accurate
- [ ] Quick start guide accurate

### 11.4.3 Verify Architecture Documentation
**Files:**
- `docs/AGENTS.md`
- `.windsurfrules`
- `backend/internal/auth/README.md`

**Checks:**
- [ ] Architecture boundaries (domain/application/infrastructure/interface) are documented
- [ ] New repository or cache ports are added to the relevant docs
- [ ] Outdated references to concrete `*gorm.DB`, `*redis.Client`, or `*postgres.*` types are removed or updated
- [ ] If code structure changed, docs are updated before merge

---

## PART 12: Final Report and Cleanup

### 12.1 Update docs/archive/FINAL_TEST_REPORT.md
**Sections:**
- Stacks identity status
- Docker build verification results
- NLP test results
- Backend test results
- API verification results
- Security verification results
- Asynchronous tasks status
- PGVECTOR status
- Redis/MongoDB status
- Frontend test results
- Public graph verification results
- CI/CD verification results
- Documentation verification results
- Overall verdict: READY / NOT READY

### 12.2 Destroy Test Stack
```bash
./scripts/testing/stop-test.sh  # Linux/Mac
./scripts/testing/stop-test.ps1  # Windows
```

**Checks:**
- [ ] Test containers stopped
- [ ] Test volumes removed
- [ ] No data leakage to dev/personal stacks

### 12.3 Cleanup Temporary Files and Generated Artifacts
**Run the cleanup helper:**
```bash
python ./scripts/cleanup/cleanup-test-artifacts.py   # Linux/Mac
python .\scripts\cleanup\cleanup-test-artifacts.py  # Windows
```

**Target files and directories:**
- `backend/coverage.out`
- `backend/*.cov`
- `backend/*.tmp`
- `backend/*.log`
- `frontend/coverage`
- `logs/test-outputs/*.log` (keep only latest run)
- `node_modules/.cache` / build artifacts (if generated)

**Checks:**
- [ ] No `coverage.out` or `*.cov` files left in the backend directory
- [ ] No temporary `.tmp` or `.log` files are committed
- [ ] `git status` shows only intended source/doc changes
- [ ] Docker volumes and test databases are removed (PART 12.2)

### 12.4 Verify Dev/Personal Stacks Unaffected
**Checks:**
- [ ] Dev stack still healthy
- [ ] Personal stack still healthy
- [ ] No test containers running
- [ ] No test volumes remaining

### 12.5 Final Verdict
**If all checks pass:**
- System is READY for manual testing
- System is READY for production deployment

**If any checks fail:**
- Document failures in docs/archive/FINAL_TEST_REPORT.md
- Fix failures before proceeding
- Re-run regression cycle

---

## Test Execution Order

1. PART 0: Stacks Identity Check
2. PART 0.5: Docker Build Verification
3. PART 0.6: Environment Configuration Check
4. PART 1: Dev/Personal Stacks Health
5. PART 2: Test Stack Startup
6. PART 3: NLP Service Tests
7. PART 3.5: Dependencies and Vulnerabilities
8. PART 4: Backend Tests
9. PART 5: Backend API Verification
10. PART 5.5: Security Verification
11. PART 6: Asynchronous Tasks
12. PART 6.5: Logs and Monitoring
13. PART 7: PGVECTOR Verification
14. PART 8: Redis and MongoDB
15. PART 8.3: Backup Verification
16. PART 9: Frontend Tests
17. PART 10: Public Graph Verification
18. PART 11: CI/CD Verification
19. PART 11.4: Documentation Verification
20. PART 12: Final Report and Cleanup

---

## Exit Criteria

The regression test is considered PASS when:
- All stacks are identical (PART 0)
- All builds succeed (PART 0.5)
- All unit tests pass (PART 3, 4, 9)
- All critical API endpoints work (PART 5)
- No security vulnerabilities (PART 3.5, 5.5)
- All infrastructure components healthy (PART 6, 7, 8)
- Frontend tests pass (PART 9)
- Public graph works (PART 10)
- CI/CD configured correctly (PART 11)

The regression test is considered FAIL when:
- Any stack identity mismatch (PART 0)
- Any build failure (PART 0.5)
- Any critical test failure (PART 3, 4, 9)
- Any API endpoint failure (PART 5)
- Any security vulnerability (PART 3.5, 5.5)
- Any infrastructure failure (PART 6, 7, 8)
- Data leakage between stacks (PART 12)

---

## Frequency

- **Full Regression:** Before each production deployment
- **Quick Regression:** Before each major feature release
- **Smoke Regression:** After each minor feature release
- **Identity Check:** Before each manual testing session
